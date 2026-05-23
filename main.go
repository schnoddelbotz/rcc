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
	"time"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "retrospective-code-coverage",
	Short:        "Retrospectively build LoC and Coverage stats and display as graph",
	Long:         `First argument can be local git repo to analyze, defaults to current workdir`,
	RunE:         rootCmdRunE,
	SilenceUsage: true,
}

func main() {
	log.SetFlags(0)

	flags := rootCmd.Flags()
	flags.IntP("workers", "w", 5, "Number of workers")
	flags.StringP("output", "o", "rcc-output.png", "Plot/Graph PNG output filename")
	flags.StringP("tmp", "t", os.TempDir(), "Temp directory path to use for history clones")
	flags.StringP("language", "l", "", "Enables details and coverage for given language")
	flags.BoolP("debug", "d", false, "Enable debug output")
	flags.BoolP("open", "O", false, "Open graph upon completion")
	flags.BoolP("skip-autodetect", "s", false, "Disable language auto detection")
	flags.StringSliceP("include-languages", "i", []string{}, "Explicitly list languages")
	// only if -l ...:
	flags.BoolP("no-cover-unit", "U", false, "Run unit tests (for given --language)")
	flags.BoolP("cover-integration", "I", false, "Run integration tests (for given --language)")
	flags.BoolP("no-cover-duration", "D", false, "Do not include duration for coverage runs in graph")
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func rootCmdRunE(cmd *cobra.Command, args []string) error {
	workers, _ := cmd.Flags().GetInt("workers")
	outfile, _ := cmd.Flags().GetString("output")
	language, _ := cmd.Flags().GetString("language")
	tmpPath, _ := cmd.Flags().GetString("tmp")
	debug, _ := cmd.Flags().GetBool("debug")
	open, _ := cmd.Flags().GetBool("open")
	noCoverU, _ := cmd.Flags().GetBool("no-cover-unit")
	CoverI, _ := cmd.Flags().GetBool("cover-integration")
	noCoverD, _ := cmd.Flags().GetBool("no-cover-duration")
	includeLanguages, _ := cmd.Flags().GetStringSlice("include-languages")
	skipAutoDetect, _ := cmd.Flags().GetBool("skip-autodetect")

	jobOpts := JobOptions{
		RunColoc:      true,
		RunColocTests: true,
		IncludeLangs:  includeLanguages,
		TmpPath:       tmpPath,
		debug:         debug,
	}

	repoPath, err := os.Getwd()
	if err != nil {
		return err
	}
	if len(args) > 0 {
		repoPath = args[0]
	}

	langdata := GetLanguage(language, skipAutoDetect, repoPath)
	if langdata == nil {
		return fmt.Errorf("invalid --language; currently supported: Go - or leave blank for generic")
	}
	titleParts := ""
	if langdata.UnitTestCmd != "" && !noCoverU {
		log.Println("Coverage (unit tests): enabled")
		jobOpts.runCoverUnit = true
		titleParts = " + Coverage (Unit-Tests)"
	}
	if langdata.IntegrationTestCmd != "" && CoverI {
		log.Println("Coverage (integration tests): enabled")
		jobOpts.runCoverIntegration = true
		titleParts = " + Coverage (Integration-Tests)"
	}
	if (jobOpts.runCoverUnit || jobOpts.runCoverIntegration) && !noCoverD {
		log.Println("Graphing of coverage test duration: enabled")
		jobOpts.includeDuration = true
		titleParts += " + Test Duration"
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
		langdata.Description, len(commits), workers, repoPath)
	runner := NewRunner(repo, langdata, jobOpts)
	runner.Run(workers, commits)

	title := fmt.Sprintf("%s LoC %s [%s]", filepath.Base(repoPath), titleParts, langdata.Description)
	err = NewGnuplotGraph(runner.StatData, title, outfile, langdata.GoclocName, jobOpts).Create()
	if err != nil {
		return err
	}

	log.Printf("Graph successfully written to: %s", outfile)
	OpenAsNeeded(open, outfile)
	return nil
}
