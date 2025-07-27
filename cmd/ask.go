package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wakatara/harsh/internal"
	"github.com/wakatara/harsh/internal/ui"
)

var askCmd = &cobra.Command{
	Use:     "ask [habit-fragment|date|yday]",
	Short:   "Ask and record your undone habits",
	Long:    "Asks and records your undone habits. Can filter by habit fragment, specific date (YYYY-MM-DD), or 'yday' for yesterday.",
	Aliases: []string{"a"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		harsh := internal.NewHarsh()
		var habitFragment string
		if len(args) > 0 {
			habitFragment = args[0]
		}
		
		input := ui.NewInput(noColor)
		input.AskHabits(
			harsh.GetHabits(), 
			harsh.GetEntries(), 
			harsh.GetRepository(), 
			harsh.GetMaxHabitNameLength(), 
			harsh.GetCountBack(), 
			habitFragment,
		)
		return nil
	},
}

