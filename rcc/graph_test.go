package rcc

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hhatto/gocloc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGraph(t *testing.T) {
	testEntry := StatDataEntry{
		sha:  "abc123",
		Date: time.Now(),
		Loc: &gocloc.Result{
			Languages: map[string]*gocloc.Language{
				"Go": {
					Code:     100,
					Comments: 20,
				},
				"GoTests": {
					Code:     50,
					Comments: 5,
				},
				"HTML": {
					Code:     5,
					Comments: 3,
				},
			},
		},
		Coverage: 75.5,
		Duration: time.Second,
	}

	statData := &StatData{
		entries: []StatDataEntry{testEntry},
	}

	tmpdir := t.TempDir()
	outfile := filepath.Join(tmpdir, "test_graph.png")

	graph := NewGnuplotGraph(statData, "Test Graph", outfile, "Go", false)
	err := graph.Create()
	require.NoError(t, err, "Create() should not return an error")

	_, err = os.Stat(outfile)
	require.NoError(t, err, "Output file should exist")
	pngFile, err := os.ReadFile(outfile)
	require.NoError(t, err, "Should be able to read the PNG file")
	expectedMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	require.True(t, len(pngFile) >= len(expectedMagic), "PNG file should be at least 8 bytes")
	assert.Equal(t, expectedMagic, pngFile[:len(expectedMagic)], "File should have valid PNG magic bytes")
}
