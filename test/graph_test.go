package test

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

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
			entries:  storage.Entries{},
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
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Test1"}:    {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Test2"}:    {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Test3"}:    {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 15}, Habit: "Test4"}:    {Result: "y"},
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
