package cmd

import (
	"fmt"
	"os"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/wakatara/harsh/internal"
)

var (
	colorOption string
	RootCmd     = &cobra.Command{
		Use:     "harsh",
		Short:   "habit tracking for geeks",
		Long:    "A simple, minimalist CLI for tracking and understanding habits.",
		Version: version, // Use the version from version.go
	}
)

var harsh *internal.Harsh

func init() {
	RootCmd.PersistentFlags().StringVarP(&colorOption, "color", "C", "auto", `manage colors in output, "always", "never" or "auto" (defaults to auto)`)
	RootCmd.RegisterFlagCompletionFunc("color", colorCompletionFunc)
	RootCmd.AddCommand(askCmd)
	RootCmd.AddCommand(todoCmd)
	RootCmd.AddCommand(logCmd)
	RootCmd.AddCommand(versionCmd)

	// Add stats as subcommand of log
	logCmd.AddCommand(statsCmd)

	// Set color disable based on color arg, or bas
	cobra.OnInitialize(func() {
		switch colorOption {
		case "never":
			color.Enable = false
		case "always":
			color.Enable = true
		case "auto":
			if color.Enable {
				fi, _ := os.Stdout.Stat()
				stdoutNotPiped := (fi.Mode() & os.ModeCharDevice) == 0
				color.Enable = stdoutNotPiped
			}
		default:
			fmt.Fprintf(os.Stderr, `invalid color option "%s". should be "never", "always" or "auto"`+"\n", colorOption)
			os.Exit(1)
		}
	})
	// initialize the global harsh instance (for context aware completion)
	harsh = internal.NewHarsh()
}

func colorCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return []cobra.Completion{"always", "never", "auto"}, cobra.ShellCompDirectiveNoFileComp
}

// Execute runs the root command
func Execute() error {
	return RootCmd.Execute()
}
