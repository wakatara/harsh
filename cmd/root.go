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
	hideEnded   bool
	RootCmd     = &cobra.Command{
		Use:     "harsh",
		Short:   "habit tracking for geeks",
		Long:    "A simple, minimalist CLI for tracking and understanding habits.",
		Version: version, // Use the version from version.go
		Run: func(cmd *cobra.Command, args []string) {
			// Trigger onboarding for first-time users running bare 'harsh'
			getHarsh()
			// Then show help
			cmd.Help()
		},
	}
)

var harsh *internal.Harsh

// getHarsh returns the global harsh instance, initializing it lazily if needed.
// This allows commands like 'version' to run without triggering onboarding.
func getHarsh() *internal.Harsh {
	if harsh == nil {
		harsh = internal.NewHarsh()
	}
	return harsh
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&colorOption, "color", "C", "auto", `manage colors in output, "always", "never" or "auto" (defaults to auto)`)
	RootCmd.PersistentFlags().BoolVarP(&hideEnded, "hide-ended", "H", false, "Hide habits that have an end date")
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
				stdoutNotPiped := (fi.Mode() & os.ModeCharDevice) != 0
				color.Enable = stdoutNotPiped;
			}
		default:
			fmt.Fprintf(os.Stderr, `invalid color option "%s". should be "never", "always" or "auto"`+"\n", colorOption)
			os.Exit(1)
		}
	})
	// Note: harsh instance is now lazily initialized via getHarsh()
	// This allows 'harsh version' to work without triggering onboarding
}

func colorCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return []cobra.Completion{"always", "never", "auto"}, cobra.ShellCompDirectiveNoFileComp
}

// Execute runs the root command
func Execute() error {
	return RootCmd.Execute()
}
