package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCliArgs(t *testing.T) {
	os.Args = []string{"foo", "bar", "-v"}

	args := getCliArgs()

	assert.Equal(t, []string{"bar"}, args.argv)
	assert.True(t, args.printVersionOnly)
}

func TestGetVersionOnly(t *testing.T) {
	err := runRCC(CliArgs{printVersionOnly: true})
	assert.Nil(t, err)
}

func TestRunOnSelf(t *testing.T) {
	outdir := t.TempDir()
	outfile := filepath.Join(outdir, "test.html")
	err := runRCC(CliArgs{outfile: outfile, noCoverUnit: true, workers: 4})
	require.Nil(t, err)
	assert.FileExists(t, outfile)
}

func TestRunInvalidLanguageErrors(t *testing.T) {
	err := runRCC(CliArgs{language: "kwak"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --language")
}

func TestRunInvalidRepoErrors(t *testing.T) {
	err := runRCC(CliArgs{argv: []string{"/not/here"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "repository does not exist")
}
