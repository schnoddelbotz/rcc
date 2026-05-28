package main

import (
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
		runColoc:         true,
		runColocTests:    true,
		includeLanguages: []string{"Go"},
	})
	runner.Run(5, []string{knownSHA})

	require.NotNil(t, runner.StatData)
	assert.Equal(t, 1, len(runner.StatData.entries))
	require.NotNil(t, runner.StatData.entries[0].Loc)
	require.NotNil(t, runner.StatData.entries[0].Loc.Languages)
	require.Equal(t, 2, len(runner.StatData.entries[0].Loc.Languages))
	require.Contains(t, runner.StatData.entries[0].Loc.Languages, "Go")
	require.Contains(t, runner.StatData.entries[0].Loc.Languages, "GoExcludingTests")
	assert.Equal(t, int32(123), runner.StatData.entries[0].Loc.Languages["Go"].Code)
	assert.Equal(t, int32(123), runner.StatData.entries[0].Loc.Languages["GoExcludingTests"].Code)
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
