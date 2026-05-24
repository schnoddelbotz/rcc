package main

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"os"
)

//go:embed graph_resources/chart.js
var chartJS string

//go:embed graph_resources/chart.html
var chartHTML string

type chartjsData struct {
	Labels   []string         `json:"labels"`
	Datasets []chartjsDataset `json:"datasets"`
}
type chartjsDataset struct {
	Label           string    `json:"label"`
	Data            []float32 `json:"data"`
	borderColor     string    //: Utils.CHART_COLORS.red,
	backgroundColor string    //: Utils.transparentize(Utils.CHART_COLORS.red, 0.5),
	YAxisID         string    `json:"yAxisID"` // 'y' / 'y1'
}

func (g *Graph) CreateChartJS(embedJSON, embedChartJS, onlyJSON bool) error {
	if onlyJSON {
		return g.writeJSONToFile(g.createJSON())
	}
	return g.writeChartHTML(g.createJSON(), embedJSON, embedChartJS)
}

func (g *Graph) createJSON() chartjsData {
	labels := []string{}
	for _, commit := range g.statData.entries {
		labels = append(labels, commit.Date.String())
	}
	datasets := []chartjsDataset{}
	for _, lang := range g.statData.languages() {
		ldata := []float32{}
		for _, commit := range g.statData.entries {
			if e, exists := commit.Loc.Languages[lang]; exists {
				ldata = append(ldata, float32(e.Code))
			} else {
				ldata = append(ldata, 0)
			}
		}
		datasets = append(datasets, chartjsDataset{
			Label:   lang,
			Data:    ldata,
			YAxisID: "y",
		})
	}
	if g.jobOptions.runCoverUnit {
		ldata := []float32{}
		for _, commit := range g.statData.entries {
			ldata = append(ldata, commit.CoverageUnit)
		}
		datasets = append(datasets, chartjsDataset{
			Label:   "CoverageUnitTests",
			Data:    ldata,
			YAxisID: "y1",
		})
		if g.jobOptions.includeDuration {
			ddata := []float32{}
			for _, commit := range g.statData.entries {
				ddata = append(ddata, float32(commit.UnitDuration.Seconds()))
			}
			datasets = append(datasets, chartjsDataset{
				Label:   "CoverageUnitDuration",
				Data:    ddata,
				YAxisID: "y1",
			})
		}
	}
	// + integration
	return chartjsData{
		Labels:   labels,
		Datasets: datasets,
	}
}

func (g *Graph) writeJSONToFile(data chartjsData) error {
	out, err := os.Create(g.outfile)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	enc := json.NewEncoder(out)
	return enc.Encode(data)
}

func (g *Graph) writeChartHTML(data chartjsData, embedJSON, embedChartJS bool) error {
	chartjson, _ := json.Marshal(data)
	var funcMap = template.FuncMap{
		"title":           func() string { return g.title },
		"embedJSON":       func() bool { return embedJSON },
		"embedChartJS":    func() bool { return embedChartJS },
		"chartJS":         func() template.JS { return template.JS(chartJS) },
		"chartData":       func() template.JS { return template.JS(chartjson) },
		"includeCoverage": func() bool { return g.jobOptions.runCoverUnit || g.jobOptions.runCoverIntegration },
	}
	tpl := template.Must(template.New("chartjs").Funcs(funcMap).Parse(chartHTML))

	out, err := os.Create(g.outfile)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	return tpl.Execute(out, nil)
}
