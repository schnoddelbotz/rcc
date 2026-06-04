package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var goTestOutput = `Leading output ...
github.com/schnoddelbotz/rcc/statdata.go:26:		sort				0.0%
github.com/schnoddelbotz/rcc/statdata.go:32:		languages			100.0%
total:							(statements)			30.1%
... Trailing output.
`
var pytestOutput = `--------------------------------------------
TOTAL                      166      20   100%
=========================== short test summary info ============================`

func TestExtractGo(t *testing.T) {
	pattern := languagesMap["Go"].CoverageRegex

	val := extractCoverage(goTestOutput, pattern)

	assert.Equal(t, float32(30.1), val)
}

func TestExtractPytest(t *testing.T) {
	pattern := languagesMap["Python"].CoverageRegex

	val := extractCoverage(pytestOutput, pattern)

	assert.Equal(t, float32(100), val)
}

func TestRunnerConstructorNilArgs(t *testing.T) {
	r := NewRunner(nil, nil, JobOptions{})
	assert.NotNil(t, r)
}

func TestRunnerAgainstSelf(t *testing.T) {
	repo, err := NewSourceRepository(".")
	lang := GetLanguage("Go", true, "/ignored/should/use/userLang")
	require.Nil(t, err)

	runner := NewRunner(repo, lang, JobOptions{
		runColoc:            true,
		runColocTests:       true,
		runCoverUnit:        true,
		runCoverIntegration: true,
		includeLanguages:    []string{"Go"},
		debug:               true,
	})
	stats, err := runner.Run(context.Background(), 5, []string{knownSHA, knownSHAFailsUnitTest})

	require.Nil(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, 1, len(stats.entries))
	require.NotNil(t, stats.entries[0].Loc)
	require.NotNil(t, stats.entries[0].Loc.Languages)
	require.Equal(t, 2, len(stats.entries[0].Loc.Languages))
	require.Contains(t, stats.entries[0].Loc.Languages, "Go")
	require.Contains(t, stats.entries[0].Loc.Languages, "GoExcludingTests")
	assert.Equal(t, int32(123), stats.entries[0].Loc.Languages["Go"].Code)
	assert.Equal(t, int32(123), stats.entries[0].Loc.Languages["GoExcludingTests"].Code)
}

func TestRunTestCmd(t *testing.T) {
	result := runTestCmd(t.TempDir(), "echo foobar 99.5%", `\d+.\d+`, true)

	require.Nil(t, result.Err)
	assert.Equal(t, float32(99.5), result.Coverage)
}

func TestRunTestError(t *testing.T) {
	result := runTestCmd(t.TempDir(), "asdasdasd dasdasda", `\d+.\d+`, true)

	require.Error(t, result.Err)
}

func TestGetJobOptions(t *testing.T) {
	cliArgs := CliArgs{
		coverUnitCmd:        "testme.sh",
		coverIntegrationCmd: "testi.sh",
		customCoverageRegex: `\d+`,
		doCoverIntegration:  true,
	}
	langdata := GetLanguage("Go", true, "/tmp")
	require.NotNil(t, langdata)

	opts := getJobOptions(cliArgs, langdata)

	assert.Equal(t, "testme.sh", langdata.UnitTestCmd)
	assert.Equal(t, "testi.sh", langdata.IntegrationTestCmd)
	assert.Equal(t, `\d+`, langdata.CoverageRegex)
	assert.Equal(t, true, opts.runCoverUnit)
	assert.Equal(t, true, opts.runCoverIntegration)
	assert.Contains(t, opts.titleParts, "Coverage (Unit-Tests)")
	assert.Contains(t, opts.titleParts, "Coverage (Integration-Tests)")
}
