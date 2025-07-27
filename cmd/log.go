package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wakatara/harsh/internal"
	"github.com/wakatara/harsh/internal/ui"
)

var logCmd = &cobra.Command{
	Use:     "log [habit-fragment]",
	Short:   "Show graph of logged habits",
	Long:    "Shows consistency graph of logged habits. Can filter by habit fragment.",
	Aliases: []string{"l"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		harsh := internal.NewHarsh()
		var habitFragment string
		if len(args) > 0 {
			habitFragment = args[0]
		}

		display := ui.NewDisplay(noColor)
		display.ShowHabitLog(
			harsh.GetHabits(),
			harsh.GetEntries(),
			harsh.GetCountBack(),
			harsh.GetMaxHabitNameLength(),
			habitFragment,
		)
		return nil
	},
}