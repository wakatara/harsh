package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal/graph"
	"github.com/wakatara/harsh/internal/storage"
)

type logJSON struct {
	Date   string     `json:"date"`
	Scores scoresJSON `json:"scores"`
	Habits []habitJSON `json:"habits"`
}

type scoresJSON struct {
	Today     float64 `json:"today"`
	Yesterday float64 `json:"yesterday"`
}

type habitJSON struct {
	Name              string    `json:"name"`
	Heading           string    `json:"heading"`
	Frequency         string    `json:"frequency"`
	Target            int       `json:"target"`
	Interval          int       `json:"interval"`
	LoggedToday       bool      `json:"logged_today"`
	Result            *string   `json:"result"`
	StreakStatus       string   `json:"streak_status"`
	DaysUntilBreak    *int     `json:"days_until_break"`
	LastCompleted     *string   `json:"last_completed"`
	CompletedInWindow *int      `json:"completed_in_window,omitempty"`
	Stats             statsJSON `json:"stats"`
}

type statsJSON struct {
	DaysTracked int     `json:"days_tracked"`
	Streaks     int     `json:"streaks"`
	Breaks      int     `json:"breaks"`
	Skips       int     `json:"skips"`
	Total       float64 `json:"total"`
}

// ShowHabitLogJSON outputs habit status as JSON for programmatic consumption
func ShowHabitLogJSON(habits []*storage.Habit, entries *storage.Entries, habitFragment string, hideEnded bool) error {
	now := civil.DateOf(time.Now())

	// Filter habits by fragment if provided
	filteredHabits := habits
	if len(strings.TrimSpace(habitFragment)) > 0 {
		filteredHabits = []*storage.Habit{}
		for _, habit := range habits {
			if strings.Contains(strings.ToLower(habit.Name), strings.ToLower(habitFragment)) {
				filteredHabits = append(filteredHabits, habit)
			}
		}
	}

	// Filter out ended habits if hideEnded is true
	if hideEnded {
		activeHabits := []*storage.Habit{}
		for _, habit := range filteredHabits {
			if !habit.IsEnded() {
				activeHabits = append(activeHabits, habit)
			}
		}
		filteredHabits = activeHabits
	}

	habitItems := make([]habitJSON, 0, len(filteredHabits))
	for _, habit := range filteredHabits {
		item := habitJSON{
			Name:      habit.Name,
			Heading:   habit.Heading,
			Frequency: habit.Frequency,
			Target:    habit.Target,
			Interval:  habit.Interval,
		}

		// Check if logged today
		outcome, loggedToday := (*entries)[storage.DailyHabit{Day: now, Habit: habit.Name}]
		item.LoggedToday = loggedToday
		if loggedToday {
			item.Result = &outcome.Result
		}

		// Streak status and days until break
		daysUntil := graph.DaysUntilStreakBreak(now, habit, *entries)
		inSkipPeriod := graph.IsInSkipPeriod(now, habit, *entries)

		noFirstRecord := civil.Date{Year: 0, Month: 0, Day: 0}
		switch {
		case habit.Target < 1:
			item.StreakStatus = "tracking"
		case habit.FirstRecord == noFirstRecord:
			item.StreakStatus = "unstarted"
		case inSkipPeriod:
			item.StreakStatus = "skipping"
			item.DaysUntilBreak = &daysUntil
		case daysUntil < 0:
			item.StreakStatus = "broken"
		default:
			item.StreakStatus = "active"
			item.DaysUntilBreak = &daysUntil
		}

		// Last completed date
		if lastDate := lastCompleted(now, habit, entries); !lastDate.IsZero() {
			s := lastDate.String()
			item.LastCompleted = &s
		}

		// Completed in window (only for multi-day interval habits)
		if habit.Interval > 1 {
			count := completedInWindow(now, habit, entries)
			item.CompletedInWindow = &count
		}

		// Stats
		stats := BuildStats(habit, entries)
		item.Stats = statsJSON{
			DaysTracked: stats.DaysTracked,
			Streaks:     stats.Streaks,
			Breaks:      stats.Breaks,
			Skips:       stats.Skips,
			Total:       stats.Total,
		}

		habitItems = append(habitItems, item)
	}

	output := logJSON{
		Date: now.String(),
		Scores: scoresJSON{
			Today:     graph.Score(now, habits, entries),
			Yesterday: graph.Score(now.AddDays(-1), habits, entries),
		},
		Habits: habitItems,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// lastCompleted finds the most recent date a habit was completed (y or s)
func lastCompleted(d civil.Date, habit *storage.Habit, entries *storage.Entries) civil.Date {
	noDate := civil.Date{}
	if habit.FirstRecord == noDate {
		return noDate
	}
	for dt := d; !dt.Before(habit.FirstRecord); dt = dt.AddDays(-1) {
		if v, ok := (*entries)[storage.DailyHabit{Day: dt, Habit: habit.Name}]; ok {
			if v.Result == "y" || v.Result == "s" {
				return dt
			}
		}
	}
	return noDate
}

// completedInWindow counts y/s results within the current interval window
func completedInWindow(d civil.Date, habit *storage.Habit, entries *storage.Entries) int {
	count := 0
	from := d.AddDays(-habit.Interval + 1)
	for dt := from; !dt.After(d); dt = dt.AddDays(1) {
		if v, ok := (*entries)[storage.DailyHabit{Day: dt, Habit: habit.Name}]; ok {
			if v.Result == "y" || v.Result == "s" {
				count++
			}
		}
	}
	return count
}
