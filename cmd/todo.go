package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wakatara/harsh/internal"
	"github.com/wakatara/harsh/internal/ui"
)

var todoCmd = &cobra.Command{
	Use:     "todo",
	Short:   "Show undone habits for today",
	Long:    "Shows undone habits for today and recent days.",
	Aliases: []string{"t"},
	RunE: func(cmd *cobra.Command, args []string) error {
		harsh := internal.NewHarsh()
		
		display := ui.NewDisplay(noColor)
		display.ShowTodos(
			harsh.GetHabits(),
			harsh.GetEntries(),
			harsh.GetMaxHabitNameLength(),
		)
		return nil
	},
}