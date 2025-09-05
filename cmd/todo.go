package cmd

import (
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/wakatara/harsh/internal/ui"
)

var todoCmd = &cobra.Command{
	Use:     "todo",
	Short:   "Show undone habits for today",
	Long:    "Shows undone habits for today and recent days.",
	Aliases: []string{"t"},
	RunE: func(cmd *cobra.Command, args []string) error {
		display := ui.NewDisplay(!color.Enable)
		display.ShowTodos(
			harsh.GetHabits(),
			harsh.GetEntries(),
			harsh.GetMaxHabitNameLength(),
		)
		return nil
	},
}
