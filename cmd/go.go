package cmd

import (
	"log"
	"os"
	"time"

	"github.com/schnoddelbotz/retrospective-code-coverage/rcc"
	"github.com/spf13/cobra"
)

// build coverage stats for a go module / app
// default to languages: Go, HTML, JS
// graph: automatically drop all-zero languages

// goCmd represents the go command
var goCmd = &cobra.Command{
	Use:   "go",
	Short: "Build loc/coverage/duration graph for a Go project",
	RunE:  goCmdRunE,
}

func goCmdRunE(cmd *cobra.Command, args []string) error {
	workers, _ := cmd.Flags().GetInt("workers")

	repoPath, err := os.Getwd()
	if err != nil {
		return err
	}
	if len(args) > 0 {
		repoPath = args[0]
	}
	log.Printf("go called, repoPath: %s", repoPath)
	repo, err := rcc.NewSourceRepository(repoPath)
	if err != nil {
		return err
	}
	commits, err := repo.GetCommits(time.UnixMilli(0), time.Now(), "main") // FIXME branch / all
	if err != nil {
		return err
	}

	runner := rcc.NewRunner(repo)
	runner.Run(workers, commits)

	return nil
}

func init() {
	runCmd.AddCommand(goCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// goCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// goCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
