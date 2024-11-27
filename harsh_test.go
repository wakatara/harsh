package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/civil"
)

func TestHabitParsing(t *testing.T) {
	tests := []struct {
		name     string
		freq     string
		target   int
		interval int
	}{
		{"Daily habit", "1", 1, 1},
		{"Weekly habit", "7", 1, 7},
		{"Three times weekly", "3/7", 3, 7},
		{"Tracking only", "0", 0, 1},
		{"Monthly habit", "30", 1, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Habit{Name: "Test", Frequency: tt.freq}
			h.parseHabitFrequency()
			if h.Target != tt.target || h.Interval != tt.interval {
				t.Errorf("got target=%d interval=%d, want target=%d interval=%d",
					h.Target, h.Interval, tt.target, tt.interval)
			}
		})
	}
}

func TestSatisfied(t *testing.T) {
	entries := Entries{}
	today := civil.DateOf(time.Now())
	yesterday := today.AddDays(-1)
	week_ago := today.AddDays(-7)

	// Test case for 3/7 habit
	habit := &Habit{
		Name:     "Test",
		Target:   3,
		Interval: 7,
	}

	// Add some test entries
	entries[DailyHabit{Day: today, Habit: "Test"}] = Outcome{Result: "y"}
	entries[DailyHabit{Day: yesterday, Habit: "Test"}] = Outcome{Result: "y"}
	entries[DailyHabit{Day: week_ago, Habit: "Test"}] = Outcome{Result: "y"}

	if !satisfied(today, habit, entries) {
		t.Error("Expected habit to be satisfied with 3 completions")
	}
}

func TestScore(t *testing.T) {
	h := &Harsh{
		Habits: []*Habit{
			{Name: "Test1", Target: 1, Interval: 1},
			{Name: "Test2", Target: 1, Interval: 1},
		},
		Entries: &Entries{},
	}

	today := civil.DateOf(time.Now())
	(*h.Entries)[DailyHabit{Day: today, Habit: "Test1"}] = Outcome{Result: "y"}
	(*h.Entries)[DailyHabit{Day: today, Habit: "Test2"}] = Outcome{Result: "y"}

	score := h.score(today)
	if score != 100.0 {
		t.Errorf("Expected score 100.0, got %f", score)
	}
}

func TestConfigFileCreation(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test habits file creation
	createExampleHabitsFile(tmpDir)
	if _, err := os.Stat(filepath.Join(tmpDir, "habits")); os.IsNotExist(err) {
		t.Error("Habits file was not created")
	}

	// Test log file creation
	createNewLogFile(tmpDir)
	if _, err := os.Stat(filepath.Join(tmpDir, "log")); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestBuildGraph(t *testing.T) {
	entries := &Entries{}
	h := &Harsh{
		CountBack: 7,
		Entries:   entries,
	}

	habit := &Habit{
		Name:     "Test",
		Target:   1,
		Interval: 1,
	}

	today := civil.DateOf(time.Now())
	(*entries)[DailyHabit{Day: today, Habit: "Test"}] = Outcome{Result: "y"}
	(*entries)[DailyHabit{Day: today, Habit: "Test"}] = Outcome{Result: "y"}

	graph := h.buildGraph(habit, false)
	graph = strings.TrimSpace(graph)
	length := len(graph)
	// Due to using Builder.Grow for string efficiency, this is 10, not 8
	if length != 10 {
		t.Errorf("Expected graph length 10, got %d", length)
	}
}

func TestWarning(t *testing.T) {
	entries := Entries{}
	today := civil.DateOf(time.Now())

	habit := &Habit{
		Name:        "Test",
		Target:      1,
		Interval:    7,
		FirstRecord: today.AddDays(-10),
	}

	if !warning(today, habit, entries) {
		t.Error("Expected warning for habit with no entries")
	}

	entries[DailyHabit{Day: today, Habit: "Test"}] = Outcome{Result: "y"}
	if warning(today, habit, entries) {
		t.Error("Expected no warning after completing habit")
	}
}
