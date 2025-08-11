package main

import (
	"os"
	"path/filepath"
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
	tests := []struct {
		name    string
		d       civil.Date
		habit   Habit
		entries Entries
		want    bool
	}{
		{
			name:  "Target = 1, Interval = 1 (should always fail)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 24},
			habit: Habit{Name: "Daily Walk", Target: 1, Interval: 1},
			entries: Entries{
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 24}, Habit: "Daily Walk"}: {Result: "y"},
			},
			want: false,
		},
		{
			name:  "Target = 1, Interval = 7 (meets target - valid streak)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 15},
			habit: Habit{Name: "Habit", Target: 1, Interval: 7},
			entries: Entries{
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 14}, Habit: "Habit"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 21}, Habit: "Habit"}: {Result: "y"},
			},
			want: true, // Habit satisfied in the last 7 days (14 → 21)
		},
		{
			name:  "Target = 1, Interval = 7 (streak is broken, does not meet target)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 23},
			habit: Habit{Name: "Habit", Target: 1, Interval: 7},
			entries: Entries{
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 16}, Habit: "Habit"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 24}, Habit: "Habit"}: {Result: "y"},
			},
			want: true, // Streak was broken (March 16) before the last valid "y" (March 24)
		},
		{
			name:  "Target = 1, Interval = 7 (no streak at all)",
			d:     civil.Date{Year: 2025, Month: 2, Day: 21},
			habit: Habit{Name: "Habit", Target: 1, Interval: 7},
			entries: Entries{
				DailyHabit{Day: civil.Date{Year: 2025, Month: 2, Day: 9}, Habit: "Habit"}:  {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 2, Day: 17}, Habit: "Habit"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 2, Day: 25}, Habit: "Habit"}: {Result: "y"},
			},
			want: true, // No "y" in the last 7 days before Feb 21, and previous streak is broken
		},

		{
			name:  "Target = 2, Interval = 7 (meets target)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 26},
			habit: Habit{Name: "Bike 10k", Target: 2, Interval: 7},
			entries: Entries{
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 25}, Habit: "Bike 10k"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 28}, Habit: "Bike 10k"}: {Result: "y"},
			},
			want: true,
		},
		{
			name:  "Target = 2, Interval = 7 (does not meet target)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 27},
			habit: Habit{Name: "Bike 10k", Target: 2, Interval: 7},
			entries: Entries{
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 23}, Habit: "Bike 10k"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 30}, Habit: "Bike 10k"}: {Result: "y"},
			},
			want: false,
		},
		{
			name:  "Target = 4, Interval = 7 (does not meet target)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 24},
			habit: Habit{Name: "Run 5k", Target: 4, Interval: 7},
			entries: Entries{
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 20}, Habit: "Run 5k"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 22}, Habit: "Run 5k"}: {Result: "y"},
			},
			want: false,
		},
		{
			name:  "Target = 7, Interval = 10 (meets target)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 24},
			habit: Habit{Name: "Swim", Target: 7, Interval: 10},
			entries: Entries{
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Swim"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 16}, Habit: "Swim"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 17}, Habit: "Swim"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 18}, Habit: "Swim"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 19}, Habit: "Swim"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 20}, Habit: "Swim"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 23}, Habit: "Swim"}: {Result: "y"},
			},
			want: true,
		},
		{
			name:  "Target = 10, Interval = 14 (meets target)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 24},
			habit: Habit{Name: "Yoga", Target: 10, Interval: 14},
			entries: Entries{
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 11}, Habit: "Yoga"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 12}, Habit: "Yoga"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 13}, Habit: "Yoga"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 14}, Habit: "Yoga"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Yoga"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 16}, Habit: "Yoga"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 17}, Habit: "Yoga"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 18}, Habit: "Yoga"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 19}, Habit: "Yoga"}: {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 22}, Habit: "Yoga"}: {Result: "y"},
			},
			want: true,
		},
		{
			name:  "Target = 3, Interval = 28 (does not meet target)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 24},
			habit: Habit{Name: "Strength Training", Target: 3, Interval: 28},
			entries: Entries{
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 5}, Habit: "Strength Training"}:  {Result: "y"},
				DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Strength Training"}: {Result: "y"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := satisfied(tt.d, &tt.habit, tt.entries)
			if got != tt.want {
				t.Errorf("satisfied() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScore(t *testing.T) {
	h := &Harsh{
		Habits: []*Habit{
			{Name: "Test1", Target: 1, Interval: 1},
			{Name: "Test2", Target: 1, Interval: 1},
			{Name: "Test3", Target: 1, Interval: 1},
			{Name: "Test4", Target: 1, Interval: 1},
		},
		Entries: &Entries{},
	}

	today := civil.DateOf(time.Now())
	(*h.Entries)[DailyHabit{Day: today, Habit: "Test1"}] = Outcome{Result: "y"}
	(*h.Entries)[DailyHabit{Day: today, Habit: "Test2"}] = Outcome{Result: "y"}
	(*h.Entries)[DailyHabit{Day: today, Habit: "Test3"}] = Outcome{Result: "n"}
	(*h.Entries)[DailyHabit{Day: today, Habit: "Test4"}] = Outcome{Result: "y"}

	score := h.score(today)
	if score != 75.0 {
		t.Errorf("Expected score 75.0, got %f", score)
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
	tom := today.AddDays(1)
	(*entries)[DailyHabit{Day: today, Habit: "Test"}] = Outcome{Result: "y"}
	(*entries)[DailyHabit{Day: tom, Habit: "Test"}] = Outcome{Result: "y"}

	graph := h.buildGraph(habit, false)
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

func TestNewHabitIntegration(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_test_integration")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original config dir and restore it after test
	originalConfigDir := configDir
	configDir = tmpDir
	defer func() { configDir = originalConfigDir }()

	// Create initial habits file
	habitsFile := filepath.Join(tmpDir, "habits")
	err = os.WriteFile(habitsFile, []byte("# Daily habits\nExisting habit: 1\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create empty log file
	logFile := filepath.Join(tmpDir, "log")
	err = os.WriteFile(logFile, []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Load initial configuration
	harsh := newHarsh()

	// Verify initial state
	if len(harsh.Habits) != 1 {
		t.Errorf("Expected 1 habit, got %d", len(harsh.Habits))
	}

	// Add new habit to habits file
	f, err := os.OpenFile(habitsFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString("\nNew habit: 1\n")
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Reload configuration
	harsh = newHarsh()

	// Verify new habit was loaded
	if len(harsh.Habits) != 2 {
		t.Errorf("Expected 2 habits, got %d", len(harsh.Habits))
	}

	foundNewHabit := false
	for _, h := range harsh.Habits {
		if h.Name == "New habit" {
			foundNewHabit = true
			break
		}
	}
	if !foundNewHabit {
		t.Error("New habit was not found in loaded habits")
	}

	// Test that new habit appears in todos
	today := civil.DateOf(time.Now())
	todos := harsh.getTodos(today, 0)

	foundInTodos := false
	for _, todo := range todos[today.String()] {
		if todo == "New habit" {
			foundInTodos = true
			break
		}
	}
	if !foundInTodos {
		t.Error("New habit was not found in todos")
	}

	// Test that new habit would be included in ask
	// We can't directly test the CLI interaction, but we can verify the habit
	// would be included in the list of habits to ask about
	undone := harsh.getTodos(today, 0)

	foundInUndone := false
	for _, h := range undone[today.String()] {
		if h == "New habit" {
			foundInUndone = true
			break
		}
	}
	if !foundInUndone {
		t.Error("New habit was not found in undone habits (would not be asked about)")
	}
}
