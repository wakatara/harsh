package cmd

import (
	"strings"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/wakatara/harsh/internal/ui"
)

var askCmd = &cobra.Command{
	Use:     "ask [habit-fragment|date|yday]",
	Short:   "Ask and record your undone habits",
	Long:    "Asks and records your undone habits. Can filter by habit fragment, specific date (YYYY-MM-DD), or 'yday' for yesterday.",
	ValidArgsFunction: askCmdValidArgs,
	Aliases: []string{"a"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var habitFragment string
		if len(args) > 0 {
			habitFragment = args[0]
		}
		
		input := ui.NewInput(!color.Enable)
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

func askCmdValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	out := []cobra.Completion{"yesterday","yday","yd", "w", "week", "last-week"};
	for _, habit := range harsh.GetHabits() {
		if strings.Contains(habit.Name, toComplete) {
			out = append(out, habit.Name)
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}
