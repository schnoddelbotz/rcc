package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// testing against our git $self ... with this early (2nd) commit
var knownSHA = "cd36f87e39a9989096f210946a1ee1bc81a1d953"

func TestGetCommits(t *testing.T) {
	r, err := NewSourceRepository(".")
	require.Nil(t, err)

	commits, err := r.GetCommits(time.Time{}, time.Now(), 1, "main")

	require.Nil(t, err)
	require.Contains(t, commits, knownSHA)

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
