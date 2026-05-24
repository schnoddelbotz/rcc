package main

import (
	"fmt"
	"path/filepath"
)

type Graph struct {
	statData     *StatData
	outfile      string
	language     string
	title        string
	jobOptions   JobOptions
	embedJSON    bool // for HTML output, embed or link JSON data?
	embedChartJS bool
}

func NewGraph(statData *StatData, title, outfile, language string, jobopts JobOptions, embedJSON, embedChartJS bool) *Graph {
	return &Graph{
		statData:     statData,
		outfile:      outfile,
		language:     language,
		title:        title,
		jobOptions:   jobopts,
		embedJSON:    embedJSON,
		embedChartJS: embedChartJS,
	}
}

func (g *Graph) Create() error {
	switch filepath.Ext(g.outfile) {
	case ".png":
		return g.CreateGnuplot()
	case ".json":
		return g.CreateChartJS(false, false, true)
	case ".html":
		return g.CreateChartJS(g.embedJSON, g.embedChartJS, false)
	default:
		return fmt.Errorf("unsupported output file '%s' format; use .png, .html or .json extension", g.outfile)
	}
}
