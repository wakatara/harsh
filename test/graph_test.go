package test

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal/graph"
	"github.com/wakatara/harsh/internal/storage"
)

func TestGraphBuildGraph(t *testing.T) {
	// Create test habit and entries
	habit := &storage.Habit{
		Name:        "Test Habit",
		Target:      1,
		Interval:    1,
		FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
	}

	entries := &storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 1}, Habit: "Test Habit"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 2}, Habit: "Test Habit"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 3}, Habit: "Test Habit"}: {Result: "s"},
	}

	// Test graph building
	graphResult := graph.BuildGraph(habit, entries, 7, false)

	// Should return a string
	if graphResult == "" {
		t.Error("Graph should not be empty")
	}

	// Should contain graph characters
	if !strings.ContainsAny(graphResult, "━•─·◌!") {
		t.Error("Graph should contain graph characters")
	}

	// Test ask mode (shorter graph)
	graphAsk := graph.BuildGraph(habit, entries, 7, true)
	if len(graphAsk) >= len(graphResult) {
		t.Error("Ask mode graph should be shorter")
	}
}

func TestGraphBuildGraphsParallel(t *testing.T) {
	// Create test habits
	habits := []*storage.Habit{
		{Name: "Habit1", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}},
		{Name: "Habit2", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}},
		{Name: "Habit3", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}},
	}

	entries := &storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 1}, Habit: "Habit1"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 1}, Habit: "Habit2"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 1}, Habit: "Habit3"}: {Result: "s"},
	}

	// Test parallel graph building
	results := graph.BuildGraphsParallel(habits, entries, 7, false)

	// Should have results for all habits
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// Each habit should have a graph
	for _, habit := range habits {
		if graph, exists := results[habit.Name]; !exists || graph == "" {
			t.Errorf("No graph found for habit %s", habit.Name)
		}
	}

	// Test with more habits than CPU cores to ensure proper worker management
	manyHabits := make([]*storage.Habit, runtime.NumCPU()*2)
	for i := range manyHabits {
		manyHabits[i] = &storage.Habit{
			Name:        fmt.Sprintf("Habit%d", i),
			Target:      1,
			Interval:    1,
			FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
		}
	}

	manyResults := graph.BuildGraphsParallel(manyHabits, entries, 7, false)
	if len(manyResults) != len(manyHabits) {
		t.Errorf("Expected %d results, got %d", len(manyHabits), len(manyResults))
	}
}

func TestGraphSatisfied(t *testing.T) {
	tests := []struct {
		name     string
		date     civil.Date
		habit    *storage.Habit
		entries  storage.Entries
		expected bool
	}{
		{
			name: "Daily habit should always return false",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Daily",
				Target:      1,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Daily"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Daily"}: {Result: "n"},
			},
			expected: false,
		},
		{
			name: "Weekly habit with target met should return true",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Weekly",
				Target:      1,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Weekly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Weekly"}: {Result: "n"},
			},
			expected: true,
		},
		{
			name: "Weekly habit with target not met should return false",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Weekly",
				Target:      2,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Weekly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Weekly"}: {Result: "n"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graph.Satisfied(tt.date, tt.habit, tt.entries)
			if result != tt.expected {
				t.Errorf("Satisfied() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGraphSkipified(t *testing.T) {
	tests := []struct {
		name     string
		date     civil.Date
		habit    *storage.Habit
		entries  storage.Entries
		expected bool
	}{
		{
			name: "Daily habit should always return false",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Daily",
				Target:      1,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Daily"}: {Result: "s"},
			},
			expected: false,
		},
		{
			name: "Weekly habit with recent skip should return true",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Weekly",
				Target:      1,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Weekly"}: {Result: "s"},
			},
			expected: true,
		},
		{
			name: "Weekly habit without recent skip should return false",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Weekly",
				Target:      1,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 1}, Habit: "Weekly"}: {Result: "s"},
			},
			expected: false, // The skip is too old, outside the grace period
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graph.Skipified(tt.date, tt.habit, tt.entries)
			if result != tt.expected {
				t.Errorf("Skipified() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGraphWarning(t *testing.T) {
	tests := []struct {
		name     string
		date     civil.Date
		habit    *storage.Habit
		entries  storage.Entries
		expected bool
	}{
		{
			name: "Habit with recent activity should not warn",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Test",
				Target:      1,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Test"}: {Result: "y"},
			},
			expected: false,
		},
		{
			name: "Habit without recent activity should warn",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Test",
				Target:      1,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 1}, Habit: "Test"}: {Result: "n"},
			},
			expected: true,
		},
		{
			name: "Tracking habit should not warn",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Tracking",
				Target:      0,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graph.Warning(tt.date, tt.habit, tt.entries)
			if result != tt.expected {
				t.Errorf("Warning() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGraphScore(t *testing.T) {
	habits := []*storage.Habit{
		{Name: "Test1", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}},
		{Name: "Test2", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}},
		{Name: "Test3", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}},
		{Name: "Test4", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}},
		{Name: "Tracking", Target: 0, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}}, // Should not affect score
	}

	entries := &storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Test1"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Test2"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Test3"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Test4"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Tracking"}: {Result: "y"},
	}

	score := graph.Score(civil.Date{Year: 2025, Month: 1, Day: 15}, habits, entries)
	expected := 75.0 // 3 out of 4 scoring habits completed
	if score != expected {
		t.Errorf("Score() = %f, want %f", score, expected)
	}

	// Test with skipped habits
	(*entries)[storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Test3"}] = storage.Outcome{Result: "s"}
	score = graph.Score(civil.Date{Year: 2025, Month: 1, Day: 15}, habits, entries)
	expected = 100.0 // 3 out of 3 non-skipped habits completed
	if score != expected {
		t.Errorf("Score() with skip = %f, want %f", score, expected)
	}

	// Test with no scorable habits
	noScoreHabits := []*storage.Habit{
		{Name: "Track1", Target: 0, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}},
		{Name: "Track2", Target: 0, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}},
	}
	score = graph.Score(civil.Date{Year: 2025, Month: 1, Day: 15}, noScoreHabits, entries)
	expected = 0.0
	if score != expected {
		t.Errorf("Score() with no scorable habits = %f, want %f", score, expected)
	}
}

func TestGraphBuildSpark(t *testing.T) {
	habits := []*storage.Habit{
		{Name: "Test1", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}},
		{Name: "Test2", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}},
	}

	entries := &storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 1}, Habit: "Test1"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 1}, Habit: "Test2"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 2}, Habit: "Test1"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 2}, Habit: "Test2"}: {Result: "n"},
	}

	from := civil.Date{Year: 2025, Month: 1, Day: 1}
	to := civil.Date{Year: 2025, Month: 1, Day: 2}

	sparkline, calline := graph.BuildSpark(from, to, habits, entries)

	// Should have 2 entries (2 days)
	if len(sparkline) != 2 {
		t.Errorf("Expected 2 sparkline entries, got %d", len(sparkline))
	}

	if len(calline) != 2 {
		t.Errorf("Expected 2 calline entries, got %d", len(calline))
	}

	// First day should be full score (100%)
	if sparkline[0] != "█" {
		t.Errorf("First day should be full spark, got %s", sparkline[0])
	}

	// Second day should be empty (0%)
	if sparkline[1] != " " {
		t.Errorf("Second day should be empty spark, got %s", sparkline[1])
	}

	// Calendar line should contain day letters
	validDayLetters := "MTWF "
	for _, letter := range calline {
		if !strings.Contains(validDayLetters, letter) {
			t.Errorf("Invalid calendar letter: %s", letter)
		}
	}
}

func TestGraphDaysUntilStreakBreak(t *testing.T) {
	tests := []struct {
		name     string
		date     civil.Date
		habit    *storage.Habit
		entries  storage.Entries
		expected int
	}{
		{
			name: "Daily habit - last success today",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Daily",
				Target:      1,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Daily"}: {Result: "y"},
			},
			expected: 1, // Streak breaks tomorrow (day 16)
		},
		{
			name: "Daily habit - last success yesterday (breaks today)",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Daily",
				Target:      1,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Daily"}: {Result: "y"},
			},
			expected: 0, // Streak breaks today (day 15) - last chance to maintain it
		},
		{
			name: "Daily habit - last success 2 days ago (already broken)",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Daily",
				Target:      1,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 13}, Habit: "Daily"}: {Result: "y"},
			},
			expected: -1, // Streak broke yesterday (day 14)
		},
		{
			name: "Weekly habit - last success 2 days ago",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Weekly",
				Target:      1,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 13}, Habit: "Weekly"}: {Result: "y"},
			},
			expected: 5, // Last success day 13, breaks on day 13+7=20, today is 15, so 20-15=5
		},
		{
			name: "90-day habit - last success 75 days ago (15 days until break)",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Travel",
				Target:      1,
				Interval:    90,
				FirstRecord: civil.Date{Year: 2024, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Last success was Nov 1, 2024 (breaks on Jan 30, 2025)
				storage.DailyHabit{Day: civil.Date{Year: 2024, Month: 11, Day: 1}, Habit: "Travel"}: {Result: "y"},
			},
			expected: 15, // Breaks on Nov 1 + 90 = Jan 30, so Jan 30 - Jan 15 = 15 days
		},
		{
			name: "90-day habit - last success 95 days ago (streak broken)",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Travel",
				Target:      1,
				Interval:    90,
				FirstRecord: civil.Date{Year: 2024, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Last success was Oct 12, 2024 (95 days before Jan 15, 2025)
				storage.DailyHabit{Day: civil.Date{Year: 2024, Month: 10, Day: 12}, Habit: "Travel"}: {Result: "y"},
			},
			expected: -5, // Breaks on Oct 12 + 90 = Jan 10, which is 5 days before Jan 15
		},
		{
			name: "90-day habit - last success 85 days ago (should be 5 days left)",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Travel",
				Target:      1,
				Interval:    90,
				FirstRecord: civil.Date{Year: 2024, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Last success was Oct 22, 2024 (85 days before Jan 15, 2025)
				storage.DailyHabit{Day: civil.Date{Year: 2024, Month: 10, Day: 22}, Habit: "Travel"}: {Result: "y"},
			},
			expected: 5, // Breaks on Oct 22 + 90 = Jan 20, so Jan 20 - Jan 15 = 5 days
		},
		{
			name: "Tracking habit (target 0) - should return -1",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Tracking",
				Target:      0,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{},
			expected: -1,
		},
		{
			name: "No success found - only 'n' entries (streak broken)",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Daily",
				Target:      1,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 13}, Habit: "Daily"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Daily"}: {Result: "n"},
			},
			expected: -999, // No success found means long broken
		},
		{
			name: "No entries at all - streak broken",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Daily",
				Target:      1,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{},
			expected: -999, // No entries at all means never started or long broken
		},
		{
			name: "Success very old (beyond 2x interval) - streak broken",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Weekly",
				Target:      1,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2024, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Last success was Dec 16, 2024 (30 days before Jan 15, 2025)
				storage.DailyHabit{Day: civil.Date{Year: 2024, Month: 12, Day: 16}, Habit: "Weekly"}: {Result: "y"},
			},
			expected: -23, // Broke on Dec 16 + 7 = Dec 23; Dec 23 to Jan 15 = 23 days
		},
		{
			name: "Skip counts as success",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Weekly",
				Target:      1,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 13}, Habit: "Weekly"}: {Result: "s"},
			},
			expected: 5, // Skip on day 13, breaks on day 20, today is 15, so 5 days
		},
		{
			name: "Skip after break does not restart streak (1/2 habit)",
			date: civil.Date{Year: 2025, Month: 1, Day: 29},
			habit: &storage.Habit{
				Name:        "Fit",
				Target:      1,
				Interval:    2,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Last valid success on day 25
				// Day 26 and 27 are 'n' (breaks, but interval allows 2 days)
				// Day 27 is deadline (25+2=27), no success by then = streak broken
				// Day 28 has skip (after break)
				// Day 29 is today
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 25}, Habit: "Fit"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 26}, Habit: "Fit"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 27}, Habit: "Fit"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 28}, Habit: "Fit"}: {Result: "s"},
			},
			expected: -999, // Skip on day 28 came AFTER break on day 27, should not restart countdown
		},
		// Interval habits tests (3/7, 2/7, etc.)
		{
			name: "Interval 3/7 - three successes, earliest success determines break",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Gym",
				Target:      3,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Successes on days 9, 12, 14
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 9}, Habit: "Gym"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 12}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Gym"}: {Result: "y"},
			},
			expected: 1, // Window 9-15 has 3 successes. Earliest is day 9. Breaks on 9+7=16. Today is 15, so 1 day
		},
		{
			name: "Interval 3/7 - only 2 successes, streak broken",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Gym",
				Target:      3,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 10}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Gym"}: {Result: "y"},
			},
			expected: -999, // Only 2 successes in the window, not satisfied
		},
		{
			name: "Interval 3/7 - critical window with tightly packed successes",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Gym",
				Target:      3,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Days 13, 14, 15 - all consecutive
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 13}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Gym"}: {Result: "y"},
			},
			expected: 5, // Window 13-19 has days 13,14,15. Earliest is 13. Breaks on 13+7=20. Today is 15, so 5 days
		},
		{
			name: "Interval 3/7 - with old success that doesn't matter",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Gym",
				Target:      3,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 5}, Habit: "Gym"}:  {Result: "y"}, // Too old
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 10}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 13}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Gym"}: {Result: "y"},
			},
			expected: 2, // Window 10-16 has days 10,13,15. Earliest is 10. Breaks on 10+7=17. Today is 15, so 2 days
		},
		{
			name: "Interval 2/7 - simpler case",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Call Mom",
				Target:      2,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 11}, Habit: "Call Mom"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Call Mom"}: {Result: "y"},
			},
			expected: 3, // Window 11-17 has days 11,15. Earliest is 11. Breaks on 11+7=18. Today is 15, so 3 days
		},
		{
			name: "Interval 2/7 - only 1 success, streak broken",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Call Mom",
				Target:      2,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Call Mom"}: {Result: "y"},
			},
			expected: -999, // Only 1 success, need 2
		},
		{
			name: "Interval 1/7 (weekly) - should work like simple habit",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Review",
				Target:      1,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 10}, Habit: "Review"}: {Result: "y"},
			},
			expected: 2, // Last success day 10, breaks on 10+7=17, today is 15, so 2 days
		},
		{
			name: "Interval 3/7 - streak broke yesterday",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Gym",
				Target:      3,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Last 3 successes: days 6, 7, 8
				// Window 6-12 satisfied days 6-12
				// Window 7-13 satisfied days 7-13
				// Window 8-14 only has 2 successes (days 7,8), so day 14 not satisfied
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 6}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 7}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 8}, Habit: "Gym"}: {Result: "y"},
			},
			expected: -999, // Broke on day 13 (6+7=13), today is 15, so -2 days (already broken)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graph.DaysUntilStreakBreak(tt.date, tt.habit, tt.entries)
			if result != tt.expected {
				t.Errorf("DaysUntilStreakBreak() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGraphScoreWithEndedHabits(t *testing.T) {
	// Test that ended habits are excluded from scoring after their end date
	habits := []*storage.Habit{
		{Name: "Active", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}},
		{Name: "Ended", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1}, EndRecord: civil.Date{Year: 2025, Month: 1, Day: 10}},
	}

	entries := &storage.Entries{
		// Both habits completed on day 5 (before end date)
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 5}, Habit: "Active"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 5}, Habit: "Ended"}:  {Result: "y"},
		// Only active habit completed on day 15 (after end date)
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Active"}: {Result: "y"},
	}

	// Day 5: Both habits active, both completed - should be 100%
	score := graph.Score(civil.Date{Year: 2025, Month: 1, Day: 5}, habits, entries)
	if score != 100.0 {
		t.Errorf("Score on day 5 (both active, both completed) = %f, want 100.0", score)
	}

	// Day 10: Last day of "Ended" habit - still counts
	(*entries)[storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 10}, Habit: "Active"}] = storage.Outcome{Result: "y"}
	(*entries)[storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 10}, Habit: "Ended"}] = storage.Outcome{Result: "y"}
	score = graph.Score(civil.Date{Year: 2025, Month: 1, Day: 10}, habits, entries)
	if score != 100.0 {
		t.Errorf("Score on day 10 (end date, both completed) = %f, want 100.0", score)
	}

	// Day 11: "Ended" habit should be excluded (after end date)
	(*entries)[storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 11}, Habit: "Active"}] = storage.Outcome{Result: "y"}
	score = graph.Score(civil.Date{Year: 2025, Month: 1, Day: 11}, habits, entries)
	if score != 100.0 {
		t.Errorf("Score on day 11 (only active habit counts) = %f, want 100.0", score)
	}

	// Day 15: Only "Active" habit counts, and it's completed
	score = graph.Score(civil.Date{Year: 2025, Month: 1, Day: 15}, habits, entries)
	if score != 100.0 {
		t.Errorf("Score on day 15 (only active completed) = %f, want 100.0", score)
	}

	// Day 15: If active habit is not completed, score should be 0
	delete(*entries, storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Active"})
	(*entries)[storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Active"}] = storage.Outcome{Result: "n"}
	score = graph.Score(civil.Date{Year: 2025, Month: 1, Day: 15}, habits, entries)
	if score != 0.0 {
		t.Errorf("Score on day 15 (active not completed, ended excluded) = %f, want 0.0", score)
	}
}

func TestGraphBuildGraphWithEndedHabit(t *testing.T) {
	// Test that graph shows blank and end marker after end date
	// Use dates relative to today so the end date is within the visible graph window
	now := civil.DateOf(time.Now())
	endDate := now.AddDays(-5) // End date is 5 days ago
	firstRecord := endDate.AddDays(-10) // Started 10 days before end date

	habit := &storage.Habit{
		Name:        "Ended Habit",
		Target:      1,
		Interval:    1,
		FirstRecord: firstRecord,
		EndRecord:   endDate,
	}

	entries := &storage.Entries{
		storage.DailyHabit{Day: endDate.AddDays(-2), Habit: "Ended Habit"}: {Result: "y"},
		storage.DailyHabit{Day: endDate.AddDays(-1), Habit: "Ended Habit"}: {Result: "y"},
		storage.DailyHabit{Day: endDate, Habit: "Ended Habit"}:             {Result: "y"}, // Last day
	}

	// Build a 20-day graph to ensure we capture the end date and days after
	graphResult := graph.BuildGraph(habit, entries, 20, false)

	// Graph should not be empty
	if graphResult == "" {
		t.Error("Graph should not be empty")
	}

	// Days after EndRecord should show as blank (space) or end marker (▏)
	if !strings.Contains(graphResult, " ") && !strings.Contains(graphResult, "▏") {
		t.Error("Graph should contain blanks or end marker for days after end date")
	}

	// The graph should contain the end marker (▏) on the day after the end date
	if !strings.Contains(graphResult, "▏") {
		t.Error("Graph should contain end marker (▏) on day after end date")
	}
}

func TestGraphIsInSkipPeriod(t *testing.T) {
	tests := []struct {
		name     string
		date     civil.Date
		habit    *storage.Habit
		entries  storage.Entries
		expected bool
	}{
		{
			name: "Daily habit - last entry was skip",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Daily",
				Target:      1,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Daily"}: {Result: "s"},
			},
			expected: true,
		},
		{
			name: "Daily habit - last entry was yes",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Daily",
				Target:      1,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Daily"}: {Result: "y"},
			},
			expected: false,
		},
		{
			name: "Weekly habit - skip within interval",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Weekly",
				Target:      1,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 12}, Habit: "Weekly"}: {Result: "s"},
			},
			expected: true,
		},
		{
			name: "Weekly habit - yes after skip",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Weekly",
				Target:      1,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 10}, Habit: "Weekly"}: {Result: "s"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Weekly"}: {Result: "y"},
			},
			expected: false, // Most recent is "y", not "s"
		},
		{
			name: "Tracking habit - no skip period",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Tracking",
				Target:      0,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Tracking"}: {Result: "s"},
			},
			expected: false, // Tracking habits don't have skip periods
		},
		{
			name: "No entries - not in skip period",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Daily",
				Target:      1,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries:  storage.Entries{},
			expected: false,
		},
		{
			name: "Skip followed by no entry",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Daily",
				Target:      1,
				Interval:    1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 13}, Habit: "Daily"}: {Result: "s"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Daily"}: {Result: "n"},
			},
			expected: true, // Most recent success-type entry is skip on day 13
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graph.IsInSkipPeriod(tt.date, tt.habit, tt.entries)
			if result != tt.expected {
				t.Errorf("IsInSkipPeriod() = %v, want %v", result, tt.expected)
			}
		})
	}
}