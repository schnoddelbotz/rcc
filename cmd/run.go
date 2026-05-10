package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `Idea:
build coverage info + produce graph
global process flags:
default to CWD for project root
default to tmpdir for local repo cloning for test runs
default to 5 parallel test runners
default to --start-date 0
default to --end-date $today
default to --no-coverage false (only LoC)
default to --no-loc (only coverage)
default to --test-timeout 30s
default to --loc-languages auto -> include all languages reported by gocloc
global graph flags:
default to --include-duration false (test duration)
default to --graph-using-gnuplot true
default to --graph-output retro-coverage.png
^ -> https://stackoverflow.com/questions/63771600/is-there-a-way-to-have-3-different-y-axes-on-one-graph-using-gnuplot`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("run called")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
