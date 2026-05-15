package rcc

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/hhatto/gocloc"
)

type GnuplotGraph struct {
	graphData *StatData
	outfile   string
	language  string
	debug     bool
	title     string
}

func NewGnuplotGraph(graphData *StatData, title, outfile, language string, debug bool) *GnuplotGraph {
	return &GnuplotGraph{
		graphData: graphData,
		outfile:   outfile,
		language:  language,
		debug:     debug,
		title:     title,
	}
}

func (g *GnuplotGraph) Create() error {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "rcc-gp")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmpdir)
		log.Printf("removed gnuplot tmpdir %s", tmpdir)
	}()
	datafile := filepath.Join(tmpdir, "rcc-gc.dat")

	err = g.gnuplotWriteData(datafile)
	if err != nil {
		return err
	}

	script := g.gnuplotCreateScript(datafile)

	if g.debug {
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

func getColNames(lang string, languages []string) []string {
	res := []string{}
	for _, l := range languages {
		if l == lang {
			res = append(res, l+"Code")
			res = append(res, l+"Comments")
			continue
		}
		res = append(res, l)
	}
	return res
}

func (g *GnuplotGraph) gnuplotWriteData(datafile string) error {
	datTpl := `# date                    {{locHeaderLangs}} coverage coverage_integration
        {{- range .Entries }}
{{formatDate .Date}} {{locCols .Loc}}
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
			return strings.Join(getColNames(g.language, g.graphData.languages()), " ")
		},
		"locCols": func(loc *gocloc.Result) string {
			outline := ""
			// range over all languages in data set. if entry lacks lang, 0-fill gap.
			for _, l := range g.graphData.languages() {
				if l == g.language {
					if lang, exists := loc.Languages[l]; exists {
						outline += fmt.Sprintf("%d %d ", lang.Code, lang.Comments)
					} else {
						outline += "0 0 "
					}
					continue
				}
				if l == g.language+"Tests" {
					if lang, exists := loc.Languages[l]; exists {
						outline += fmt.Sprintf("%d ", lang.Code)
					} else {
						outline += "0 "
					}
					continue
				}
				if lang, exists := loc.Languages[l]; exists {
					outline += fmt.Sprintf("%d ", lang.Code)
				} else {
					outline += "0 "
				}
			}
			// HACK HACK
			//outline += "5 5 5 5 5 5"
			return outline
		},
	}
	tpl := template.Must(template.New("gnuplot").Funcs(funcMap).Parse(datTpl))

	fh, err := os.Create(datafile)
	if err != nil {
		return err
	}
	defer fh.Close()

	err = tpl.Execute(fh, tplData{
		Languages: g.graphData.languages(),
		Entries:   g.graphData.entries,
	})
	if err != nil {
		panic(err)
	}
	return nil
}

func (g *GnuplotGraph) gnuplotCreateScript(datafile string) string {
	scriptTpl := `set title '{{.Title}}'
        set xlabel 'Date'
        set timefmt "%Y-%m-%dT%H:%M:%S+02:00"
        set xdata time
        set ytics 500 nomirror
        set ylabel 'LoC'
        set y2tics 10 nomirror
        set y2label 'Coverage'
        set y2range [0:100]
        set term pngcairo
        set terminal png size 1024,768
        set output "{{.OutFile}}"
        plot {{plotArgs}}
`

	var funcMap = template.FuncMap{
		"plotArgs": func() string {
			res := ""
			plotArgFmt := `'%s' using 1:%d t '%s' with linespoints`
			colNames := getColNames(g.language, g.graphData.languages())
			for column, lang := range colNames {
				res += fmt.Sprintf(plotArgFmt, datafile, column+2, lang)
				if column+1 < len(colNames) {
					res += ", \\\n"
				}
			}
			return res
		},
	}

	tpl := template.Must(template.New("gnuplot").Funcs(funcMap).Parse(scriptTpl))

	type tplData struct {
		// DataFile  string
		OutFile string
		Title   string
		// Languages []string
	}

	buf := bytes.NewBuffer([]byte(``))
	err := tpl.Execute(buf, tplData{
		// DataFile:  datafile,
		OutFile: g.outfile,
		// Languages: g.graphData.languages(),
		Title: g.title,
	})
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
		defer stdin.Close()
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
