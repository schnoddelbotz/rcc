package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testing against our git $self ... with this early (2nd) commit
var firstCommit = "a12c2436fda1a9d9d09af41ebe739a9b030a7e0e"
var knownSHA = "cd36f87e39a9989096f210946a1ee1bc81a1d953"
var knownSHAFailsUnitTest = "b785f264"

func TestGetCommits(t *testing.T) {
	r, err := NewSourceRepository(".")
	require.Nil(t, err)

	commits, err := r.GetCommits(time.Time{}, time.Now(), 3, "main")

	require.Nil(t, err)
	require.Contains(t, commits, firstCommit)

}

func TestLocalClone(t *testing.T) {
	r, err := NewSourceRepository(".")
	require.Nil(t, err)
	testClonePath := t.TempDir()

	_, err = r.LocalClone(testClonePath, knownSHA)
	require.Nil(t, err)

	// verify clone
	clone, err := NewSourceRepository(testClonePath)
	require.Nil(t, err)
	commits, err := clone.GetCommits(time.Time{}, time.Now(), 1, "main")
	require.Nil(t, err)
	require.Contains(t, commits, knownSHA)
}

func TestLocalCloneFails(t *testing.T) {
	r, err := NewSourceRepository(".")
	require.Nil(t, err)
	_, err = r.LocalClone("/dev/null", "123")
	assert.ErrorContains(t, err, "copy failed")

	// // tmp := t.TempDir()
	src := &SourceRepository{}
	// require.Nil(t, err)
	_, err = src.LocalClone(t.TempDir(), "123")
	require.ErrorContains(t, err, "reference not found")
}
