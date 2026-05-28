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
	"runtime/debug"
	"time"

	"github.com/spf13/pflag"
)

type CliArgs struct {
	doCoverIntegration bool
	noCoverDuration    bool
	noCoverUnit        bool
	noEmbedChartJS     bool
	noEmbedJSON        bool
	logTimestamps      bool
	open               bool
	printDebug         bool
	skipAutoDetect     bool
	printVersionOnly   bool

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
	log.SetFlags(0)
	if err := runRCC(getCliArgs(pflag.ExitOnError)); err != nil {
		log.Fatalf("ERROR: %s", err)
	}
}

func getCliArgs(errorHandling pflag.ErrorHandling) CliArgs {
	opt := CliArgs{}
	f := pflag.NewFlagSet("rcc", errorHandling)
	f.BoolVarP(&opt.doCoverIntegration, "cover-integration", "I", false, "Run integration tests")
	f.BoolVarP(&opt.noCoverDuration, "no-cover-duration", "D", false, "Don't graph coverage duration")
	f.BoolVarP(&opt.noCoverUnit, "no-cover-unit", "U", false, "Don't run unit tests")
	f.BoolVarP(&opt.noEmbedChartJS, "html-no-embed-chartjs", "J", false, "Don't embed ChartJS into generated .html")
	f.BoolVarP(&opt.noEmbedJSON, "html-no-embed-json", "j", false, "Don't embed JSON into generated .html")
	f.BoolVarP(&opt.open, "open", "O", false, "Open --output file upon completion")
	f.BoolVarP(&opt.logTimestamps, "timestamps", "T", false, "Timestamp console log output")
	f.BoolVarP(&opt.printDebug, "debug", "d", false, "Enable debug output")
	f.BoolVarP(&opt.printVersionOnly, "version", "v", false, "Print rcc version and exit")
	f.BoolVarP(&opt.skipAutoDetect, "skip-autodetect", "s", false, "Disable language auto detection")
	f.IntVarP(&opt.workers, "workers", "w", 5, "Number of workers")
	f.StringSliceVarP(&opt.includeLanguages, "include-languages", "i", []string{}, "Explicitly list languages for LoC")
	f.StringVarP(&opt.coverIntegrationCmd, "cover-integration-cmd", "X", "", "Command for running integration tests")
	f.StringVarP(&opt.coverUnitCmd, "cover-unit-cmd", "C", "", "Command for running unit tests")
	f.StringVarP(&opt.customCoverageRegex, "cover-regex", "R", "", "Regex for coverage value extraction")
	f.StringVarP(&opt.language, "language", "l", "", "Set project language (default autodetect)")
	f.StringVarP(&opt.outfile, "output", "o", "rcc-output.html", "Output file")
	f.StringVarP(&opt.tmpPath, "tmp", "t", os.TempDir(), "Temporary work directory")
	f.Usage = func() {
		log.Println("Retrospective Code Coverage (rcc) walks a local git repo's history  and collects")
		log.Println("lines of code (LoC) and test coverage statistics data of each commit on its way.")
		log.Println("Collected data can be plotted as HTML (default) or PNG  and be exported as JSON.")
		log.Println("Without an  REPOSITORY  argument, rcc will analyze the current directory's repo.")
		log.Println("For more details, see https://github.com/schnoddelbotz/rcc")
		log.Println()
		log.Println("Usage:")
		log.Println("  rcc [flags] [REPOSITORY]")
		log.Println()
		log.Println("Flags:")
		log.Println(f.FlagUsages())
	}
	_ = f.Parse(os.Args[1:])
	opt.argv = f.Args()
	return opt
}

func runRCC(args CliArgs) error {
	log.Print(version())
	if args.printVersionOnly {
		return nil
	}
	if args.logTimestamps {
		log.SetFlags(log.LstdFlags)
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
	return OpenAsNeeded(args.open, args.outfile)
}

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "rcc VersionUnknown"
	}
	return fmt.Sprintf("rcc %s (%s)", info.Main.Version, info.GoVersion)
}
