package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hhatto/gocloc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDate, _ = time.Parse("2006-01-02", "2026-05-26")
var testEntry = StatDataEntry{
	sha:  "abc123456",
	Date: testDate,
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
	CoverageIntegration: 75.5,
	UnitDuration:        time.Second,
}
var testGraph = &Graph{
	jobOptions: JobOptions{
		runCoverUnit:        true,
		includeDuration:     true,
		runCoverIntegration: true,
	},
	statData: &StatData{entries: []StatDataEntry{testEntry}},
}

func TestCreateGnuplotPNG(t *testing.T) {
	testEntry2 := testEntry
	testEntry2.sha = "cdef7890"
	testEntry2.Date = testEntry.Date.Add(30 * time.Hour)
	statData := &StatData{entries: []StatDataEntry{testEntry, testEntry2}}
	tmpdir := t.TempDir()
	outfile := filepath.Join(tmpdir, "test_graph.png")

	graph := NewGraph(statData, "Test Graph", outfile, "Go", JobOptions{}, false, false)
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

func TestCreateChartjsHTML(t *testing.T) {
	testEntry2 := testEntry
	testEntry2.sha = "cdef7890"
	testEntry2.Date = testEntry.Date.Add(30 * time.Hour)
	statData := &StatData{entries: []StatDataEntry{testEntry, testEntry2}}
	tmpdir := t.TempDir()
	outfile := filepath.Join(tmpdir, "test_graph.html")

	graph := NewGraph(statData, "Test Graph", outfile, "Go", JobOptions{}, true, true)
	err := graph.Create()
	require.NoError(t, err, "Create() should not return an error")

	_, err = os.Stat(outfile)
	require.NoError(t, err, "Output file should exist")
	htmlFile, err := os.ReadFile(outfile)
	require.NoError(t, err, "Should be able to read the HTML file")
	assert.Contains(t, string(htmlFile), "https://www.chartjs.org", "File should contain embedded chartjs library")
}

func TestCreateChartjsJSON(t *testing.T) {
	statData := &StatData{entries: []StatDataEntry{testEntry}}
	tmpdir := t.TempDir()
	outfile := filepath.Join(tmpdir, "test_graph.json")

	graph := NewGraph(statData, "Test Graph", outfile, "Go", JobOptions{}, false, false)
	err := graph.Create()
	require.NoError(t, err, "Create() should not return an error")

	_, err = os.Stat(outfile)
	require.NoError(t, err, "Output file should exist")
	jsonFile, err := os.ReadFile(outfile)
	require.NoError(t, err, "Should be able to read the JSON file")
	assert.Contains(t, string(jsonFile), "{\"labels\":[\"2026-05-26 abc12345\"],\"datasets\":[{\"label\":\"Go\",", "File should contain expected JSON structure")
}

func TestChartjsCreateJSON(t *testing.T) {
	data := testGraph.createJSON()

	// testEntry has 3 columns + coverage U + coverage I + duration U + duration I = 7
	assert.Equal(t, 7, len(data.Datasets))
}

func TestChartjsWriteJSONErrors(t *testing.T) {
	testGraph.outfile = "/dev/forbidden"

	data := testGraph.createJSON()
	err := testGraph.writeJSONToFile(data)

	assert.Contains(t, err.Error(), "operation not permitted")
}

func TestChartjsWriteHTMLErrors(t *testing.T) {
	testGraph.outfile = "/dev/forbidden"

	data := testGraph.createJSON()
	err := testGraph.writeChartHTML(data, false, false)

	assert.Contains(t, err.Error(), "operation not permitted")
}

func TestGnuplotCreateScript(t *testing.T) {
	script := testGraph.gnuplotCreateScript("/dev/null")

	assert.Contains(t, script, "set y2range [0:100]")
	assert.Contains(t, script, "'/dev/null' using 1:2 t 'Go' with linespoints")
	assert.Contains(t, script, "'/dev/null' using 1:5 t 'UnitTestCoverage' axis x1y2 with linespoints")
}
