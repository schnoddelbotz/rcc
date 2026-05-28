package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSorting(t *testing.T) {
	testdata := StatData{
		entries: []StatDataEntry{
			// should get sorted by time - sha indicates desired position for test
			{sha: "c", Date: time.Now().Add(2 * time.Hour)},
			{sha: "a", Date: time.Now()},
			{sha: "b", Date: time.Now().Add(1 * time.Hour)},
			{sha: "d", Date: time.Now().Add(3 * time.Hour)},
		},
	}

	testdata.sort()

	assert.Equal(t, "a", testdata.entries[0].sha)
	assert.Equal(t, "b", testdata.entries[1].sha)
	assert.Equal(t, "c", testdata.entries[2].sha)
	assert.Equal(t, "d", testdata.entries[3].sha)
}
