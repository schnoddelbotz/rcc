package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/hhatto/gocloc"
	"github.com/schnoddelbotz/rcc/resources"
)

func (g *Graph) CreateGnuplot() error {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "rcc-gp")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmpdir)
		// log.Printf("removed gnuplot tmpdir %s", tmpdir)
	}()
	datafile := filepath.Join(tmpdir, "rcc-gc.dat")

	err = g.gnuplotWriteData(datafile)
	if err != nil {
		return err
	}

	script := g.gnuplotCreateScript(datafile)

	if g.jobOptions.debug {
		dat, _ := os.ReadFile(datafile)
		fmt.Println(string(dat))
		fmt.Println(script)
	}

	err = gnuplotExec(script)
	if err != nil {
		return err
	}

	return nil
}

func (g *Graph) gnuplotWriteData(datafile string) error {
	datTpl := `# date                    {{locHeaderLangs}} coverage_unit coverage_integration coverage_unit_duration coverage_integration_duration
        {{- range .Entries }}
{{formatDate .Date}} {{locCols .Loc}}{{covCols .}}
        {{- end }}
`
	type tplData struct {
		Entries   []StatDataEntry
		Languages []string
	}
	var funcMap = template.FuncMap{
		"formatDate": func(timeStamp time.Time) string {
			return timeStamp.Format(time.RFC3339Nano)
		},
		"locHeaderLangs": func() string {
			// + COVERAGE FIXME // DURATION
			return strings.Join(g.statData.languages(), " ")
		},
		"covCols": func(sd *StatDataEntry) string {
			// g.jobOptions.includeDuration affects duration display for both unit and integration tests
			outline := ""
			if g.jobOptions.runCoverUnit {
				outline += fmt.Sprintf("%2f ", sd.CoverageUnit)
				if g.jobOptions.includeDuration {
					outline += fmt.Sprintf("%2f ", sd.UnitDuration.Seconds())
				}
			}
			if g.jobOptions.runCoverIntegration {
				outline += fmt.Sprintf("%2f ", sd.CoverageIntegration)
				if g.jobOptions.includeDuration {
					outline += fmt.Sprintf("%2f ", sd.IntegrationDuration.Seconds())
				}
			}
			return outline
		},
		"locCols": func(loc *gocloc.Result) string {
			var outline strings.Builder
			// range over all languages in data set. if entry lacks lang, 0-fill gap.
			for _, l := range g.statData.languages() {
				if l == g.language {
					if lang, exists := loc.Languages[l]; exists {
						fmt.Fprintf(&outline, "%d ", lang.Code)
					} else {
						outline.WriteString("0 ")
					}
					continue
				}
				if l == g.language+"ExcludingTests" {
					if lang, exists := loc.Languages[l]; exists {
						fmt.Fprintf(&outline, "%d ", lang.Code)
					} else {
						outline.WriteString("0 ")
					}
					continue
				}
				if lang, exists := loc.Languages[l]; exists {
					fmt.Fprintf(&outline, "%d ", lang.Code)
				} else {
					outline.WriteString("0 ")
				}
			}
			return outline.String()
		},
	}
	tpl := template.Must(template.New("gnuplot").Funcs(funcMap).Parse(datTpl))

	fh, err := os.Create(datafile)
	if err != nil {
		return err
	}
	defer func() { _ = fh.Close() }()

	err = tpl.Execute(fh, tplData{
		Languages: g.statData.languages(),
		Entries:   g.statData.entries,
	})
	if err != nil {
		panic(err)
	}
	return nil
}

func (g *Graph) gnuplotCreateScript(datafile string) string {
	var funcMap = template.FuncMap{
		"title":   func() string { return g.title },
		"outfile": func() string { return g.outfile },
		"y2Label": func() string {
			label := ""
			if g.jobOptions.runCoverIntegration || g.jobOptions.runCoverUnit {
				label = "Coverage (%)"
				if g.jobOptions.includeDuration {
					// fixme: third axis required if values > 100 ... or switch to minutes?
					label += " + Test Duration (s)"
				}
			}
			return label
		},
		"y2RangeIfAny": func() string {
			if g.jobOptions.runCoverIntegration || g.jobOptions.runCoverUnit {
				return "\n        set y2range [0:100]"
			}
			return ""
		},
		"plotArgs": func() string {
			var res strings.Builder
			plotArgFmt := `'%s' using 1:%d t '%s' with linespoints`
			colNames := g.statData.languages()
			for column, lang := range colNames {
				fmt.Fprintf(&res, plotArgFmt, datafile, column+2, lang)
				res.WriteString(", \\\n")
			}
			covCol := len(colNames) + 2
			if g.jobOptions.runCoverUnit {
				plotArgFmt = `'%s' using 1:%d t '%s' axis x1y2 with linespoints`
				fmt.Fprintf(&res, plotArgFmt, datafile, covCol, "UnitTestCoverage")
				res.WriteString(", \\\n")
				covCol++
				if g.jobOptions.includeDuration {
					fmt.Fprintf(&res, plotArgFmt, datafile, len(colNames)+3, "UnitTestDuration")
					res.WriteString(", \\\n")
					covCol++
				}
			}
			if g.jobOptions.runCoverIntegration {
				plotArgFmt = `'%s' using 1:%d t '%s' axis x1y2 with linespoints`
				fmt.Fprintf(&res, plotArgFmt, datafile, covCol, "IntegrationTestCoverage")
				res.WriteString(", \\\n")
				covCol++
				if g.jobOptions.includeDuration {
					fmt.Fprintf(&res, plotArgFmt, datafile, covCol, "IntegrationTestDuration")
					res.WriteString(", \\\n")
					covCol++
				}
			}
			return res.String()
		},
	}

	tpl := template.Must(template.New("gnuplot").Funcs(funcMap).Parse(resources.GnuplotScriptTemplate))

	buf := bytes.NewBuffer([]byte(``))
	err := tpl.Execute(buf, nil)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func gnuplotExec(script string) error {
	cmd := exec.Command("gnuplot")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer func() { _ = stdin.Close() }()
		_, err := io.WriteString(stdin, script)
		if err != nil {
			panic(err)
		}
	}()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gnuplot failed: %w", err)
	}
	return nil
}
