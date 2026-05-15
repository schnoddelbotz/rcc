package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
// Idea:
// build coverage info + produce graph
// global process flags:
// default to --start-date 0
// default to --end-date $today
// default to --no-coverage false (only LoC)
// default to --no-loc (only coverage)
// default to --test-timeout 30s
// global graph flags:
// default to --include-duration false (test duration, in s)
// ^ -> https://stackoverflow.com/questions/63771600/is-there-a-way-to-have-3-different-y-axes-on-one-graph-using-gnuplot`,
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("run called")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	flags := runCmd.PersistentFlags()
	flags.IntP("workers", "w", 5, "Number of workers")
	flags.StringP("output", "o", "rcc-output.png", "Plot/Graph PNG output filename")
	flags.BoolP("debug", "d", false, "enable debug output")
	flags.StringSliceP("include-languages", "i", []string{}, "Explicitly list languages")

}
