package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCliArgs(t *testing.T) {
	os.Args = []string{"foo", "bar", "-v"}

	args := getCliArgs(pflag.ExitOnError)

	assert.Equal(t, []string{"bar"}, args.argv)
	assert.True(t, args.printVersionOnly)
}

func TestGetHelp(t *testing.T) {
	os.Args = []string{"foo", "-h"}
	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)

	_ = getCliArgs(pflag.ContinueOnError)

	assert.Contains(t, buf.String(), "Retrospective Code Coverage (rcc) walks a local git repo")
	assert.Contains(t, buf.String(), "-I, --cover-integration")
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

func TestOpenAsNeeded(t *testing.T) {
	err := OpenAsNeeded(true, "/not/here")
	// just ensures that calling open works - currently unable to catch open failure
	assert.Nil(t, err)
}
