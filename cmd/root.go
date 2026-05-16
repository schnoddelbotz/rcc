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
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/schnoddelbotz/retrospective-code-coverage/rcc"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "retrospective-code-coverage",
	Short:        "A brief description of your application",
	Long:         ``,
	RunE:         rootCmdRunE,
	SilenceUsage: true,
}

func rootCmdRunE(cmd *cobra.Command, args []string) error {
	workers, _ := cmd.Flags().GetInt("workers")
	outfile, _ := cmd.Flags().GetString("output")
	language, _ := cmd.Flags().GetString("language")
	debug, _ := cmd.Flags().GetBool("debug")
	open, _ := cmd.Flags().GetBool("open")
	includeLanguages, _ := cmd.Flags().GetStringSlice("include-languages")

	jobOpts := rcc.JobOptions{
		RunColoc:      true,
		RunColocTests: true,
		IncludeLangs:  includeLanguages,
	}

	repoPath, err := os.Getwd()
	if err != nil {
		return err
	}
	if len(args) > 0 {
		repoPath = args[0]
	}
	langdata := rcc.GetLanguage(language)
	if langdata == nil {
		return fmt.Errorf("invalid --language; currently supported: Go - or leave blank for generic")
	}

	repo, err := rcc.NewSourceRepository(repoPath)
	if err != nil {
		return err
	}
	commits, err := repo.GetCommits(time.UnixMilli(0), time.Now(), "main") // FIXME branch / all
	if err != nil {
		return err
	}

	log.Printf("Running for %s on %d commits with %d workers, repoPath: %s",
		langdata.Description, len(commits), workers, repoPath)
	runner := rcc.NewRunner(repo, langdata, jobOpts)
	runner.Run(workers, commits)

	title := fmt.Sprintf("%s LoC / Coverage [%s]", filepath.Base(repoPath), langdata.Description)
	err = rcc.NewGnuplotGraph(runner.StatData, title, outfile, langdata.GoclocName, debug).Create()
	if err != nil {
		return err
	}

	log.Printf("Graph successfully written to: %s", outfile)
	rcc.OpenAsNeeded(open, outfile)
	return nil
}

func init() {
	flags := rootCmd.Flags()
	flags.IntP("workers", "w", 5, "Number of workers")
	flags.StringP("output", "o", "rcc-output.png", "Plot/Graph PNG output filename")
	flags.StringP("language", "l", "", "Enables details and coverage for given language")
	flags.BoolP("debug", "d", false, "enable debug output")
	flags.BoolP("open", "O", false, "open graph upon completion")
	flags.StringSliceP("include-languages", "i", []string{}, "Explicitly list languages")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
