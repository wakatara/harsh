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
		{
			name: "Completion after gap exceeding interval starts new streak (1/8 habit)",
			date: civil.Date{Year: 2025, Month: 1, Day: 22},
			habit: &storage.Habit{
				Name:        "Review",
				Target:      1,
				Interval:    8,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Success on day 9, then gap of 10 days (exceeds interval of 8)
				// New success on day 19 should start a fresh streak
				// Today is day 22, break date is 19+8=27, so 5 days until break
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 9}, Habit: "Review"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 10}, Habit: "Review"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 11}, Habit: "Review"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 12}, Habit: "Review"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 13}, Habit: "Review"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Review"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Review"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 16}, Habit: "Review"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 17}, Habit: "Review"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 18}, Habit: "Review"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 19}, Habit: "Review"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 20}, Habit: "Review"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 21}, Habit: "Review"}: {Result: "n"},
			},
			expected: 5, // New streak from day 19, breaks on 27, today is 22 = 5 days
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
			name: "Interval 3/7 - six successes, target-th most recent anchors break",
			date: civil.Date{Year: 2026, Month: 5, Day: 22},
			habit: &storage.Habit{
				Name:        "Dev",
				Target:      3,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2026, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Trailing window [5/16, 5/22] has six successes
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 16}, Habit: "Dev"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 17}, Habit: "Dev"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 18}, Habit: "Dev"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 19}, Habit: "Dev"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 20}, Habit: "Dev"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 21}, Habit: "Dev"}: {Result: "y"},
			},
			expected: 4, // 3rd most recent success is 5/19; breaks on 5/19+7=5/26; today 5/22 -> 4 days
		},
		{
			name: "Interval 2/15 - eight successes, target-th most recent anchors break",
			date: civil.Date{Year: 2026, Month: 5, Day: 22},
			habit: &storage.Habit{
				Name:        "Meet",
				Target:      2,
				Interval:    15,
				FirstRecord: civil.Date{Year: 2026, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Trailing window [5/8, 5/22] has eight successes
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 8}, Habit: "Meet"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 9}, Habit: "Meet"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 10}, Habit: "Meet"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 12}, Habit: "Meet"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 17}, Habit: "Meet"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 19}, Habit: "Meet"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 20}, Habit: "Meet"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 21}, Habit: "Meet"}: {Result: "y"},
			},
			expected: 13, // 2nd most recent success is 5/20; breaks on 5/20+15=6/4; today 5/22 -> 13 days
		},
		{
			name: "Interval 3/7 - skips count toward target-th most recent",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name:        "Gym",
				Target:      3,
				Interval:    7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 10}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 12}, Habit: "Gym"}: {Result: "s"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 13}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Gym"}: {Result: "y"},
			},
			expected: 5, // 3rd most recent (incl. skip) is day 13; breaks on 13+7=20; today 15 -> 5 days
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

// TestDaysUntilStreakBreakSurplus exhaustively covers the case the original tests
// missed: when a windowed habit has MORE successes than target. The buggy
// implementation anchored the streak-break date on the earliest success in any
// satisfying window, which only coincides with the correct answer when
// successes == target. The fix anchors on the target-th most recent success
// in the trailing window -- earlier successes age out first, but the streak
// survives them as long as the target-th most recent stays in range.
func TestDaysUntilStreakBreakSurplus(t *testing.T) {
	tests := []struct {
		name     string
		date     civil.Date
		habit    *storage.Habit
		entries  storage.Entries
		expected int
	}{
		{
			// User-reported bug: Dev 3/7 with 6 wins in trailing window.
			// 3rd most recent success is day 19; breaks on 19+7=26; today 22 -> 4
			name: "3/7 with 6 successes back-to-back",
			date: civil.Date{Year: 2026, Month: 5, Day: 22},
			habit: &storage.Habit{
				Name: "Dev", Target: 3, Interval: 7,
				FirstRecord: civil.Date{Year: 2026, Month: 1, Day: 1},
			},
			entries: makeYEntries("Dev", civil.Date{Year: 2026, Month: 5, Day: 16}, 6),
			expected: 4,
		},
		{
			// User-reported bug: Meet 2/15 with 8 wins.
			// 2nd most recent success is 5/20; breaks on 5/20+15=6/4; today 5/22 -> 13
			name: "2/15 with 8 successes scattered",
			date: civil.Date{Year: 2026, Month: 5, Day: 22},
			habit: &storage.Habit{
				Name: "Meet", Target: 2, Interval: 15,
				FirstRecord: civil.Date{Year: 2026, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 8}, Habit: "Meet"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 9}, Habit: "Meet"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 10}, Habit: "Meet"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 12}, Habit: "Meet"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 17}, Habit: "Meet"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 19}, Habit: "Meet"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 20}, Habit: "Meet"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 5, Day: 21}, Habit: "Meet"}: {Result: "y"},
			},
			expected: 13,
		},
		{
			// All 7 days in window. 3rd most recent is d-2; breaks d-2+7=d+5.
			name: "3/7 with full 7/7 successes",
			date: civil.Date{Year: 2025, Month: 1, Day: 20},
			habit: &storage.Habit{
				Name: "Daily-ish", Target: 3, Interval: 7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: makeYEntries("Daily-ish", civil.Date{Year: 2025, Month: 1, Day: 14}, 7),
			expected: 5,
		},
		{
			// 4 wins for 3/7. 3rd most recent is d-3; breaks d-3+7=d+4.
			name: "3/7 with 4 successes",
			date: civil.Date{Year: 2025, Month: 1, Day: 20},
			habit: &storage.Habit{
				Name: "Gym", Target: 3, Interval: 7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 17}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 19}, Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 20}, Habit: "Gym"}: {Result: "y"},
			},
			expected: 4, // 3rd most recent in window is day 17; breaks 17+7=24; 24-20=4.
		},
		{
			// 5/14 with 7 wins clustered late. 5th most recent ends up far back.
			name: "5/14 with 7 successes",
			date: civil.Date{Year: 2025, Month: 3, Day: 15},
			habit: &storage.Habit{
				Name: "Workout", Target: 5, Interval: 14,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Trailing window [3/2, 3/15]. Successes on 3/4, 3/6, 3/8, 3/10, 3/12, 3/14, 3/15
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 4}, Habit: "Workout"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 6}, Habit: "Workout"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 8}, Habit: "Workout"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 10}, Habit: "Workout"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 12}, Habit: "Workout"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 14}, Habit: "Workout"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Workout"}: {Result: "y"},
			},
			// 5th most recent: 15, 14, 12, 10, 8 -> day 8. Breaks 8+14=22. 22-15=7.
			expected: 7,
		},
		{
			// Skips count toward Nth-most-recent same as completions.
			name: "3/7 mixed y and s with surplus",
			date: civil.Date{Year: 2025, Month: 1, Day: 20},
			habit: &storage.Habit{
				Name: "Read", Target: 3, Interval: 7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Read"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 16}, Habit: "Read"}: {Result: "s"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 18}, Habit: "Read"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 19}, Habit: "Read"}: {Result: "s"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 20}, Habit: "Read"}: {Result: "y"},
			},
			// 3rd most recent (counting skips): 20, 19, 18 -> day 18. Breaks 18+7=25. 25-20=5.
			expected: 5,
		},
		{
			// "n" entries among successes shouldn't shift the count.
			name: "3/7 with surplus successes interspersed with n",
			date: civil.Date{Year: 2025, Month: 1, Day: 20},
			habit: &storage.Habit{
				Name: "Yoga", Target: 3, Interval: 7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 14}, Habit: "Yoga"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Yoga"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 16}, Habit: "Yoga"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 17}, Habit: "Yoga"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 18}, Habit: "Yoga"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 19}, Habit: "Yoga"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 20}, Habit: "Yoga"}: {Result: "y"},
			},
			// 3rd most recent: 20, 19, 18 -> day 18. Breaks 25. 25-20=5.
			expected: 5,
		},
		{
			// Surplus where some successes are outside trailing window -- they
			// should not anchor the break date.
			name: "3/7 with successes outside trailing window ignored",
			date: civil.Date{Year: 2025, Month: 2, Day: 15},
			habit: &storage.Habit{
				Name: "Code", Target: 3, Interval: 7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Old wins outside trailing window [2/9, 2/15]
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 5}, Habit: "Code"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 30}, Habit: "Code"}: {Result: "y"},
				// In-window wins
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 2, Day: 11}, Habit: "Code"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 2, Day: 13}, Habit: "Code"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 2, Day: 14}, Habit: "Code"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 2, Day: 15}, Habit: "Code"}: {Result: "y"},
			},
			// 3rd most recent in window: 15, 14, 13 -> day 13. Breaks 13+7=20. 20-15=5.
			expected: 5,
		},
		{
			// Large interval, many surplus successes. Verifies the algorithm
			// scales beyond weekly habits.
			name: "5/30 with 12 successes",
			date: civil.Date{Year: 2025, Month: 6, Day: 30},
			habit: &storage.Habit{
				Name: "Monthly", Target: 5, Interval: 30,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				// Trailing window [6/1, 6/30]
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 2}, Habit: "Monthly"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 5}, Habit: "Monthly"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 8}, Habit: "Monthly"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 11}, Habit: "Monthly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 14}, Habit: "Monthly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 17}, Habit: "Monthly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 20}, Habit: "Monthly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 23}, Habit: "Monthly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 25}, Habit: "Monthly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 27}, Habit: "Monthly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 29}, Habit: "Monthly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 30}, Habit: "Monthly"}: {Result: "y"},
			},
			// 5th most recent: 30, 29, 27, 25, 23 -> day 23. Breaks 23+30=Jul 23. Jul 23 - Jun 30 = 23 days.
			expected: 23,
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

// TestDaysUntilStreakBreakSimpleSurplus locks down the Target=1 path under
// realistic conditions (multiple successes in the lookback). Even though the
// Simple path is mathematically immune to the windowed bug (last == 1st most
// recent), we explicitly cover surplus scenarios so future refactors can't
// regress the property.
func TestDaysUntilStreakBreakSimpleSurplus(t *testing.T) {
	tests := []struct {
		name     string
		date     civil.Date
		habit    *storage.Habit
		entries  storage.Entries
		expected int
	}{
		{
			// 1/1 habit with multiple recent successes -- the *most recent*
			// is what matters, not the earliest.
			name: "1/1 with multiple successes, only last matters",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name: "Daily", Target: 1, Interval: 1,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: makeYEntries("Daily", civil.Date{Year: 2025, Month: 1, Day: 10}, 6),
			// Most recent: 1/15. Breaks 1/15+1=1/16. 1 day.
			expected: 1,
		},
		{
			// 1/7 (weekly) with multiple wins in window -- only latest anchors.
			name: "1/7 with multiple wins, last anchors",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name: "Weekly", Target: 1, Interval: 7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 10}, Habit: "Weekly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 12}, Habit: "Weekly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 13}, Habit: "Weekly"}: {Result: "y"},
			},
			// Latest is 1/13. Breaks 1/13+7=1/20. 1/20-1/15=5 days.
			expected: 5,
		},
		{
			// 1/30 with many wins -- still only latest matters.
			name: "1/30 with multiple wins, last anchors",
			date: civil.Date{Year: 2025, Month: 6, Day: 15},
			habit: &storage.Habit{
				Name: "Monthly", Target: 1, Interval: 30,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 5, Day: 25}, Habit: "Monthly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 1}, Habit: "Monthly"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 6, Day: 10}, Habit: "Monthly"}: {Result: "y"},
			},
			// Latest is 6/10. Breaks 6/10+30=7/10. 7/10-6/15=25 days.
			expected: 25,
		},
		{
			// Skip is most recent, but valid prior success exists within interval.
			// Skip maintains the streak.
			name: "1/7 with prior y then later s within interval",
			date: civil.Date{Year: 2025, Month: 1, Day: 15},
			habit: &storage.Habit{
				Name: "Weekly", Target: 1, Interval: 7,
				FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 10}, Habit: "Weekly"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 13}, Habit: "Weekly"}: {Result: "s"},
			},
			// Latest "success" (y or s) is 1/13. Breaks 1/13+7=1/20. 5 days.
			expected: 5,
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

// TestDaysUntilStreakBreakInvariant is a property test: for any windowed habit
// where Satisfied(today) is true, the days-until-break MUST equal
// (target-th-most-recent-success + interval) - today. Sweeps a grid of
// target/interval/success-pattern combinations to ensure the bug class can't
// reappear under any distribution.
func TestDaysUntilStreakBreakInvariant(t *testing.T) {
	today := civil.Date{Year: 2025, Month: 6, Day: 30}
	firstRecord := civil.Date{Year: 2025, Month: 1, Day: 1}

	type config struct {
		target, interval int
		// successOffsets are days BEFORE today (0 = today, 1 = yesterday, ...).
		// Must contain at least target entries in [0, interval-1] for the
		// invariant to apply (i.e. today is satisfied).
		successOffsets []int
	}

	configs := []config{
		// 3/7 surplus distributions
		{3, 7, []int{0, 1, 2, 3, 4, 5}},
		{3, 7, []int{0, 1, 2}},
		{3, 7, []int{0, 3, 6}},
		{3, 7, []int{0, 1, 6}},
		{3, 7, []int{0, 2, 4, 6}},
		// 2/15 surplus distributions
		{2, 15, []int{0, 1}},
		{2, 15, []int{0, 14}},
		{2, 15, []int{0, 2, 5, 9, 14}},
		// 5/14 surplus
		{5, 14, []int{0, 2, 4, 6, 8, 10, 12}},
		{5, 14, []int{0, 1, 2, 3, 4}},
		// 4/30 surplus
		{4, 30, []int{0, 5, 10, 15, 20, 25, 29}},
	}

	for i, cfg := range configs {
		t.Run(fmt.Sprintf("cfg_%d_%d/%d_n=%d", i, cfg.target, cfg.interval, len(cfg.successOffsets)), func(t *testing.T) {
			habit := &storage.Habit{
				Name: "H", Target: cfg.target, Interval: cfg.interval, FirstRecord: firstRecord,
			}
			entries := storage.Entries{}
			for _, off := range cfg.successOffsets {
				entries[storage.DailyHabit{Day: today.AddDays(-off), Habit: "H"}] = storage.Outcome{Result: "y"}
			}

			// Compute expected: target-th most recent success in trailing window.
			// successOffsets is unordered; the target-th most recent has the
			// target-th smallest offset.
			sorted := append([]int{}, cfg.successOffsets...)
			// Ascending sort -- smallest offset = most recent.
			for a := 0; a < len(sorted); a++ {
				for b := a + 1; b < len(sorted); b++ {
					if sorted[b] < sorted[a] {
						sorted[a], sorted[b] = sorted[b], sorted[a]
					}
				}
			}
			if len(sorted) < cfg.target {
				t.Fatalf("config has fewer successes than target: %d < %d", len(sorted), cfg.target)
			}
			anchorOffset := sorted[cfg.target-1]
			anchorDate := today.AddDays(-anchorOffset)
			expected := anchorDate.AddDays(cfg.interval).DaysSince(today)

			got := graph.DaysUntilStreakBreak(today, habit, entries)
			if got != expected {
				t.Errorf("target=%d interval=%d offsets=%v: got %d, want %d (anchor=%v)",
					cfg.target, cfg.interval, cfg.successOffsets, got, expected, anchorDate)
			}
		})
	}
}

// makeYEntries creates n consecutive daily "y" entries for habitName starting at start.
func makeYEntries(habitName string, start civil.Date, n int) storage.Entries {
	e := storage.Entries{}
	for i := 0; i < n; i++ {
		e[storage.DailyHabit{Day: start.AddDays(i), Habit: habitName}] = storage.Outcome{Result: "y"}
	}
	return e
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