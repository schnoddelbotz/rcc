/*
Copyright © 2026 Jan Hacker

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/spf13/pflag"
)

type CliArgs struct {
	CoverI           bool
	noCoverD         bool
	noCoverU         bool
	noEmbedChartJS   bool
	noEmbedJSON      bool
	open             bool
	printDebug       bool
	skipAutoDetect   bool
	printVersionOnly bool

	workers int

	argv             []string
	includeLanguages []string

	coverIntegrationCmd string
	coverUnitCmd        string
	customCoverageRegex string
	language            string
	outfile             string
	tmpPath             string
}

func main() {
	if err := runRCC(getCliArgs()); err != nil {
		log.Fatal(err)
	}
}

func getCliArgs() CliArgs {
	opt := CliArgs{}
	pflag.BoolVarP(&opt.CoverI, "cover-integration", "I", false, "Run integration tests (for given --language)")
	pflag.BoolVarP(&opt.noCoverD, "no-cover-duration", "D", false, "Do not include duration for coverage runs in graph")
	pflag.BoolVarP(&opt.noCoverU, "no-cover-unit", "U", false, "Do not run unit tests (for given --language)")
	pflag.BoolVarP(&opt.noEmbedChartJS, "html-no-embed-chartjs", "J", false, "Do not embed ChartJS into generated .html, but link it")
	pflag.BoolVarP(&opt.noEmbedJSON, "html-no-embed-json", "j", false, "Do not embed JSON data into generated .html, but link it")
	pflag.BoolVarP(&opt.open, "open", "O", false, "Open graph upon completion")
	pflag.BoolVarP(&opt.printDebug, "debug", "d", false, "Enable debug output")
	pflag.BoolVarP(&opt.printVersionOnly, "version", "v", false, "Print rcc version and exit")
	pflag.BoolVarP(&opt.skipAutoDetect, "skip-autodetect", "s", false, "Disable language auto detection")
	pflag.IntVarP(&opt.workers, "workers", "w", 5, "Number of workers")
	pflag.StringSliceVarP(&opt.includeLanguages, "include-languages", "i", []string{}, "Explicitly list languages for LoC")
	pflag.StringVarP(&opt.coverIntegrationCmd, "cover-integration-cmd", "X", "", "Custom shell command for running integration tests")
	pflag.StringVarP(&opt.coverUnitCmd, "cover-unit-cmd", "C", "", "Custom shell command for running unit tests")
	pflag.StringVarP(&opt.customCoverageRegex, "cover-regex", "R", "", "Custom regex to extract coverage value from test command output")
	pflag.StringVarP(&opt.language, "language", "l", "", "Enables details and coverage for given language")
	pflag.StringVarP(&opt.outfile, "output", "o", "rcc-output.html", "Plot/Graph html/json/png output filename")
	pflag.StringVarP(&opt.tmpPath, "tmp", "t", os.TempDir(), "Temp directory path to use for history clones")
	pflag.Parse()
	opt.argv = pflag.Args()
	return opt
}

func runRCC(args CliArgs) error {
	log.SetFlags(0)
	log.Printf("rcc %s", versionInfo())
	if args.printVersionOnly {
		return nil
	}

	repoPath, err := os.Getwd()
	if err != nil {
		return err
	}
	if len(args.argv) > 0 {
		repoPath = args.argv[0]
	}

	language := GetLanguage(args.language, args.skipAutoDetect, repoPath)
	if language == nil {
		return fmt.Errorf("invalid --language; currently supported: Go,Python,Java - or leave blank for generic")
	}

	repo, err := NewSourceRepository(repoPath)
	if err != nil {
		return err
	}
	commits, err := repo.GetCommits(time.UnixMilli(0), time.Now(), "main") // FIXME branch / all
	if err != nil {
		return err
	}

	log.Printf("Running for %s on %d commits with %d workers, repoPath: %s",
		language.Description, len(commits), args.workers, repoPath)
	jobOpts := getJobOptions(args, language)
	runner := NewRunner(repo, language, jobOpts)
	runner.Run(args.workers, commits)

	title := fmt.Sprintf("%s LoC %s [%s]", filepath.Base(repoPath), jobOpts.titleParts, language.Description)
	err = NewGraph(runner.StatData, title, args.outfile, language.GoclocName, jobOpts, !args.noEmbedJSON, !args.noEmbedChartJS).Create()
	if err != nil {
		return err
	}

	log.Printf("Graph successfully written to: %s", args.outfile)
	OpenAsNeeded(args.open, args.outfile)
	return nil
}

func versionInfo() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "VersionUnknown"
	}
	gitrev := getBuildSetting(info, "vcs.revision")
	if len(gitrev) == 40 {
		gitrev = gitrev[:8]
	}
	return fmt.Sprintf("%s built %s (%s)", gitrev, getBuildSetting(info, "vcs.time"), runtime.Version())
}

func getBuildSetting(info *debug.BuildInfo, name string) string {
	for _, v := range info.Settings {
		if v.Key == name {
			return v.Value
		}
	}
	return "[?" + name + "?]"
}
