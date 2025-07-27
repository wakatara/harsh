package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wakatara/harsh/internal"
	"github.com/wakatara/harsh/internal/ui"
)

var statsCmd = &cobra.Command{
	Use:     "stats",
	Short:   "Show habit stats for entire log file",
	Long:    "Shows statistics for all habits including streaks, breaks, skips, and totals.",
	Aliases: []string{"s"},
	RunE: func(cmd *cobra.Command, args []string) error {
		harsh := internal.NewHarsh()
		
		display := ui.NewDisplay(noColor)
		display.ShowHabitStats(
			harsh.GetHabits(),
			harsh.GetEntries(),
			harsh.GetMaxHabitNameLength(),
		)
		return nil
	},
}