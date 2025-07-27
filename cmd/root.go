package cmd

import (
	"github.com/gookit/color"
	"github.com/spf13/cobra"
)

var (
	noColor bool
	RootCmd = &cobra.Command{
		Use:     "harsh",
		Short:   "habit tracking for geeks",
		Long:    "A simple, minimalist CLI for tracking and understanding habits.",
		Version: "0.11.0",
	}
)

func init() {
	RootCmd.PersistentFlags().BoolVarP(&noColor, "no-color", "n", false, "no colors in output")
	RootCmd.AddCommand(askCmd)
	RootCmd.AddCommand(todoCmd)
	RootCmd.AddCommand(logCmd)

	// Add stats as subcommand of log
	logCmd.AddCommand(statsCmd)

	// Set color disable based on flag
	cobra.OnInitialize(func() {
		if noColor {
			color.Disable()
		}
	})
}

// Execute runs the root command
func Execute() error {
	return RootCmd.Execute()
}