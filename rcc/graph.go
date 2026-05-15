package rcc

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"github.com/hhatto/gocloc"
)

func GraphGnuplot(graphData *StatData, outfile string) error {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "rcc-gp")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmpdir)
		log.Printf("removed gnuplot tmpdir %s", tmpdir)
	}()
	datafile := filepath.Join(tmpdir, "rcc-gc.dat")

	err = gnuplotWriteData(graphData, datafile)
	if err != nil {
		return err
	}

	/////
	// dat, _ := os.ReadFile(datafile)
	// fmt.Print(string(dat))
	/////

	script := gnuplotCreateScript(graphData, datafile, outfile)

	err = gnuplotExec(script)
	if err != nil {
		return err
	}

	return nil
}

func gnuplotWriteData(sd *StatData, datafile string) error {
	datTpl := `# date                    {{range .Languages}}{{.}} {{end}}coverage coverage_integration
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
		"locCols": func(loc *gocloc.Result) string {
			outline := ""
			// range over all languages in data set. if entry lacks lang, 0-fill gap.
			for _, l := range sd.languages() {
				if lang, exists := loc.Languages[l]; exists {
					outline += fmt.Sprintf("%d ", lang.Code)
				} else {
					outline += "0 "
				}
			}
			// HACK HACK
			outline += "5 5 5 5 5 5"
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
		Languages: sd.languages(),
		Entries:   sd.entries,
	})
	if err != nil {
		panic(err)
	}
	return nil
}

func gnuplotCreateScript(sd *StatData, datafile, outfile string) string {
	scriptTpl := `set title '{{.}}'
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
        plot '{{.DataFile}}' using 1:2 t 'Go, including tests' with linespoints, \
             '{{.DataFile}}' using 1:3 t 'Go, excluding tests' with linespoints, \
             '{{.DataFile}}' using 1:4 t 'HTML' with linespoints, \
             '{{.DataFile}}' using 1:5 t 'JS' with linespoints, \
             '{{.DataFile}}' using 1:6 t 'PS1' with linespoints, \
             '{{.DataFile}}' using 1:7 t 'Coverage' axis x1y2 with linespoints`

	tpl := template.Must(template.New("gnuplot").Parse(scriptTpl))

	type tplData struct {
		DataFile  string
		OutFile   string
		Title     string
		Languages []string
	}

	buf := bytes.NewBuffer([]byte(``))
	err := tpl.Execute(buf, tplData{
		DataFile:  datafile,
		OutFile:   outfile,
		Languages: sd.languages(),
		Title:     "foo",
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
