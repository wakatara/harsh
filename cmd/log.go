package cmd

import (
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/wakatara/harsh/internal/ui"
)

var logCmd = &cobra.Command{
	Use:     "log [habit-fragment]",
	Short:   "Show graph of logged habits",
	Long:    "Shows consistency graph of logged habits. Can filter by habit fragment.",
	Aliases: []string{"l"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var habitFragment string
		if len(args) > 0 {
			habitFragment = args[0]
		}

		h := getHarsh()

		if jsonOutput {
			return ui.ShowHabitLogJSON(
				h.GetHabits(),
				h.GetEntries(),
				habitFragment,
				hideEnded,
			)
		}

		display := ui.NewDisplay(!color.Enable)
		display.ShowHabitLog(
			h.GetHabits(),
			h.GetEntries(),
			h.GetCountBack(),
			h.GetMaxHabitNameLength(),
			habitFragment,
			hideEnded,
		)
		return nil
	},
}
