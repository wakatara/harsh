package cmd

import (
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/wakatara/harsh/internal/ui"
)

var statsCmd = &cobra.Command{
	Use:     "stats",
	Short:   "Show habit stats for entire log file",
	Long:    "Shows statistics for all habits including streaks, breaks, skips, and totals.",
	Aliases: []string{"s"},
	RunE: func(cmd *cobra.Command, args []string) error {
		h := getHarsh()
		display := ui.NewDisplay(!color.Enable)
		display.ShowHabitStats(
			h.GetHabits(),
			h.GetEntries(),
			h.GetMaxHabitNameLength(),
			hideEnded,
		)
		return nil
	},
}
