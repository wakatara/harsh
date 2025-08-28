package test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"cloud.google.com/go/civil"
	"github.com/spf13/cobra"
	"github.com/wakatara/harsh/cmd"
	"github.com/wakatara/harsh/internal"
	"github.com/wakatara/harsh/internal/graph"
	"github.com/wakatara/harsh/internal/storage"
	"github.com/wakatara/harsh/internal/ui"
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
			h := &storage.Habit{Name: "Test", Frequency: tt.freq}
			h.ParseHabitFrequency()
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
		habit   storage.Habit
		entries storage.Entries
		want    bool
	}{
		{
			name:  "Target = 1, Interval = 1 (should always fail)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 24},
			habit: storage.Habit{Name: "Daily Walk", Target: 1, Interval: 1},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 24}, Habit: "Daily Walk"}: {Result: "y"},
			},
			want: false,
		},
		{
			name:  "Target = 1, Interval = 7 (meets target - valid streak)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 15},
			habit: storage.Habit{Name: "Habit", Target: 1, Interval: 7},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 14}, Habit: "Habit"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 21}, Habit: "Habit"}: {Result: "y"},
			},
			want: true, // Habit satisfied in the last 7 days (14 → 21)
		},
		{
			name:  "Target = 1, Interval = 7 (streak is broken, does not meet target)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 23},
			habit: storage.Habit{Name: "Habit", Target: 1, Interval: 7},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 16}, Habit: "Habit"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 24}, Habit: "Habit"}: {Result: "y"},
			},
			want: false, // On March 23, looking back 7 days (March 17-23), there are no "y" entries
		},
		{
			name:  "Target = 1, Interval = 7 (no streak at all)",
			d:     civil.Date{Year: 2025, Month: 2, Day: 21},
			habit: storage.Habit{Name: "Habit", Target: 1, Interval: 7},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 2, Day: 9}, Habit: "Habit"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 2, Day: 17}, Habit: "Habit"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 2, Day: 25}, Habit: "Habit"}: {Result: "y"},
			},
			want: true, // No "y" in the last 7 days before Feb 21, and previous streak is broken
		},

		{
			name:  "Target = 2, Interval = 7 (meets target with gap filling)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 26},
			habit: storage.Habit{Name: "Bike 10k", Target: 2, Interval: 7},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 25}, Habit: "Bike 10k"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 28}, Habit: "Bike 10k"}: {Result: "y"},
			},
			want: true, // March 26 should be satisfied: window March 25-31 contains 2 successes with supporting data
		},
		{
			name:  "Target = 2, Interval = 7 (does not meet target)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 27},
			habit: storage.Habit{Name: "Bike 10k", Target: 2, Interval: 7},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 23}, Habit: "Bike 10k"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 30}, Habit: "Bike 10k"}: {Result: "y"},
			},
			want: false,
		},
		{
			name:  "Target = 4, Interval = 7 (does not meet target)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 24},
			habit: storage.Habit{Name: "Run 5k", Target: 4, Interval: 7},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 20}, Habit: "Run 5k"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 22}, Habit: "Run 5k"}: {Result: "y"},
			},
			want: false,
		},
		{
			name:  "Target = 7, Interval = 10 (meets target)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 24},
			habit: storage.Habit{Name: "Swim", Target: 7, Interval: 10},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Swim"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 16}, Habit: "Swim"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 17}, Habit: "Swim"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 18}, Habit: "Swim"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 19}, Habit: "Swim"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 20}, Habit: "Swim"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 23}, Habit: "Swim"}: {Result: "y"},
			},
			want: true,
		},
		{
			name:  "Target = 10, Interval = 14 (meets target)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 24},
			habit: storage.Habit{Name: "Yoga", Target: 10, Interval: 14},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 11}, Habit: "Yoga"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 12}, Habit: "Yoga"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 13}, Habit: "Yoga"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 14}, Habit: "Yoga"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Yoga"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 16}, Habit: "Yoga"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 17}, Habit: "Yoga"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 18}, Habit: "Yoga"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 19}, Habit: "Yoga"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 22}, Habit: "Yoga"}: {Result: "y"},
			},
			want: true,
		},
		{
			name:  "Target = 3, Interval = 28 (does not meet target)",
			d:     civil.Date{Year: 2025, Month: 3, Day: 24},
			habit: storage.Habit{Name: "Strength Training", Target: 3, Interval: 28},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 5}, Habit: "Strength Training"}:  {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Strength Training"}: {Result: "y"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := graph.Satisfied(tt.d, &tt.habit, tt.entries)
			if got != tt.want {
				t.Errorf("graph.Satisfied() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSatisfiedGapScenarios tests scenarios where gaps should be filled by satisfied markers
func TestSatisfiedGapScenarios(t *testing.T) {
	tests := []struct {
		name     string
		checkingDate civil.Date
		habit    storage.Habit
		entries  storage.Entries
		want     bool
		explanation string
	}{
		{
			name:     "Astro real-world scenario: 2/7 habit with 4 consecutive successes",
			checkingDate: civil.Date{Year: 2025, Month: 8, Day: 25}, // Checking Aug 25
			habit:    storage.Habit{Name: "Astro", Target: 2, Interval: 7, FirstRecord: civil.Date{Year: 2025, Month: 8, Day: 20}},
			entries: storage.Entries{
				// Real data from the user's log
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 24}, Habit: "Astro"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 25}, Habit: "Astro"}: {Result: "y"}, 
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 26}, Habit: "Astro"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 27}, Habit: "Astro"}: {Result: "y"},
			},
			want: true, // Should be satisfied - window contains 4 successes, well above the 2 target
			explanation: "Aug 25 with 4 consecutive successes should definitely be satisfied for a 2/7 habit",
		},
		{
			name:     "Astro scenario: Days between successes should be satisfied",
			checkingDate: civil.Date{Year: 2025, Month: 8, Day: 25}, // Day between successes
			habit:    storage.Habit{Name: "Astro", Target: 2, Interval: 7, FirstRecord: civil.Date{Year: 2025, Month: 8, Day: 20}},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 21}, Habit: "Astro"}: {Result: "y"}, // Past success for support
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 24}, Habit: "Astro"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 25}, Habit: "Astro"}: {Result: "n"}, // Checking this
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 26}, Habit: "Astro"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 27}, Habit: "Astro"}: {Result: "y"},
			},
			want: true, // Window Aug 21-27 contains 4 successes, satisfies 2/7 requirement
			explanation: "Days between multiple successes should be satisfied when target is met in window",
		},
		{
			name:     "2/7 habit: Middle day with successes on both sides should be satisfied",
			checkingDate: civil.Date{Year: 2025, Month: 8, Day: 25}, // Checking middle day
			habit:    storage.Habit{Name: "Astro", Target: 2, Interval: 7, FirstRecord: civil.Date{Year: 2025, Month: 8, Day: 20}},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 24}, Habit: "Astro"}: {Result: "y"}, // Before
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 25}, Habit: "Astro"}: {Result: "n"}, // Checking this day
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 26}, Habit: "Astro"}: {Result: "y"}, // After
			},
			want: true, // Should be satisfied because window Aug 24-30 contains 2 successes (Aug 24, Aug 26)
			explanation: "Aug 25 should be satisfied because the 7-day window Aug 24-30 contains 2 successes",
		},
		{
			name:     "2/7 habit: Gap day between two pairs of successes",
			checkingDate: civil.Date{Year: 2025, Month: 8, Day: 23},
			habit:    storage.Habit{Name: "Astro", Target: 2, Interval: 7, FirstRecord: civil.Date{Year: 2025, Month: 8, Day: 20}},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 21}, Habit: "Astro"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 22}, Habit: "Astro"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 23}, Habit: "Astro"}: {Result: "n"}, // Checking
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 25}, Habit: "Astro"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 26}, Habit: "Astro"}: {Result: "y"},
			},
			want: true, // Window Aug 21-27 contains 4 successes
			explanation: "Aug 23 should be satisfied because window Aug 21-27 contains successes on Aug 21,22,25,26",
		},
		{
			name:     "3/7 habit: Gap with insufficient surrounding successes",
			checkingDate: civil.Date{Year: 2025, Month: 8, Day: 25},
			habit:    storage.Habit{Name: "Exercise", Target: 3, Interval: 7, FirstRecord: civil.Date{Year: 2025, Month: 8, Day: 20}},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 24}, Habit: "Exercise"}: {Result: "y"}, // 1 success
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 25}, Habit: "Exercise"}: {Result: "n"}, // Checking
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 26}, Habit: "Exercise"}: {Result: "y"}, // 1 success
				// Only 2 successes in any 7-day window, but need 3
			},
			want: false, // Not satisfied - only 2 successes available, need 3
			explanation: "Aug 25 should NOT be satisfied because no 7-day window contains 3+ successes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := graph.Satisfied(tt.checkingDate, &tt.habit, tt.entries)
			if got != tt.want {
				t.Errorf("graph.Satisfied() = %v, want %v\nExplanation: %s", got, tt.want, tt.explanation)
			}
		})
	}
}

// TestSatisfiedFutureDataBug tests the specific bug where future data incorrectly affected past date calculations
func TestSatisfiedFutureDataBug(t *testing.T) {
	tests := []struct {
		name     string
		checkingDate civil.Date
		habit    storage.Habit
		entries  storage.Entries
		want     bool
		explanation string
	}{
		{
			name:     "1/3 habit: Past date should not be satisfied by future success",
			checkingDate: civil.Date{Year: 2025, Month: 8, Day: 26}, // Checking Aug 26
			habit:    storage.Habit{Name: "Write", Target: 1, Interval: 3, FirstRecord: civil.Date{Year: 2025, Month: 8, Day: 25}},
			entries: storage.Entries{
				// Aug 25: n, Aug 26: n, Aug 27: n, Aug 28: y
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 25}, Habit: "Write"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 26}, Habit: "Write"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 27}, Habit: "Write"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 28}, Habit: "Write"}: {Result: "y"}, // Future success
			},
			want: false, // Should be false - Aug 26 cannot be satisfied by Aug 28's success
			explanation: "When checking Aug 26, sliding windows should only consider Aug 24-26, not future Aug 28",
		},
		{
			name:     "1/3 habit: Past date should not be satisfied by future success (Aug 27)",
			checkingDate: civil.Date{Year: 2025, Month: 8, Day: 27}, // Checking Aug 27
			habit:    storage.Habit{Name: "Write", Target: 1, Interval: 3, FirstRecord: civil.Date{Year: 2025, Month: 8, Day: 25}},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 25}, Habit: "Write"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 26}, Habit: "Write"}: {Result: "n"}, 
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 27}, Habit: "Write"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 28}, Habit: "Write"}: {Result: "y"}, // Future success
			},
			want: false, // Should be false - Aug 27 cannot be satisfied by Aug 28's success
			explanation: "When checking Aug 27, sliding windows should only consider Aug 25-27, not future Aug 28",
		},
		{
			name:     "1/3 habit: Current date with success should be satisfied",
			checkingDate: civil.Date{Year: 2025, Month: 8, Day: 28}, // Checking Aug 28 (current)
			habit:    storage.Habit{Name: "Write", Target: 1, Interval: 3, FirstRecord: civil.Date{Year: 2025, Month: 8, Day: 25}},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 25}, Habit: "Write"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 26}, Habit: "Write"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 27}, Habit: "Write"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 28}, Habit: "Write"}: {Result: "y"},
			},
			want: true, // Should be true - Aug 28 window (Aug 26-28) contains Aug 28 success
			explanation: "When checking Aug 28, sliding windows Aug 26-28 contains the Aug 28 success",
		},
		{
			name:     "1/3 habit: Past success within window should satisfy",
			checkingDate: civil.Date{Year: 2025, Month: 8, Day: 27},
			habit:    storage.Habit{Name: "Write", Target: 1, Interval: 3, FirstRecord: civil.Date{Year: 2025, Month: 8, Day: 25}},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 25}, Habit: "Write"}: {Result: "y"}, // Past success
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 26}, Habit: "Write"}: {Result: "n"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 27}, Habit: "Write"}: {Result: "n"},
			},
			want: true, // Should be true - Aug 27 window (Aug 25-27) contains Aug 25 success
			explanation: "When checking Aug 27, sliding windows Aug 25-27 contains the Aug 25 success",
		},
		{
			name:     "2/7 habit: Date before any success should not be satisfied by future successes",
			checkingDate: civil.Date{Year: 2025, Month: 8, Day: 24}, // Before first success
			habit:    storage.Habit{Name: "Exercise", Target: 2, Interval: 7, FirstRecord: civil.Date{Year: 2025, Month: 8, Day: 20}},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 25}, Habit: "Exercise"}: {Result: "y"}, // Future from perspective of Aug 24
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 8, Day: 26}, Habit: "Exercise"}: {Result: "y"}, // Future from perspective of Aug 24
			},
			want: false, // Should be false - Aug 24 has no supporting data up to that date
			explanation: "When checking Aug 24, there are no successes up to Aug 24, so future successes shouldn't satisfy it",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := graph.Satisfied(tt.checkingDate, &tt.habit, tt.entries)
			if got != tt.want {
				t.Errorf("graph.Satisfied() = %v, want %v\nExplanation: %s", got, tt.want, tt.explanation)
				
				// Provide detailed debugging info
				start := tt.checkingDate.AddDays(-tt.habit.Interval + 1)
				if start.Before(tt.habit.FirstRecord) {
					start = tt.habit.FirstRecord
				}
				t.Errorf("Debug: Checking date=%s, habit=%s (target=%d, interval=%d)", 
					tt.checkingDate.String(), tt.habit.Name, tt.habit.Target, tt.habit.Interval)
				t.Errorf("Debug: Valid window should be %s to %s", start.String(), tt.checkingDate.String())
				
				// Show what entries exist
				for dh, outcome := range tt.entries {
					if dh.Habit == tt.habit.Name {
						t.Errorf("Debug: Entry on %s: %s", dh.Day.String(), outcome.Result)
					}
				}
			}
		})
	}
}

func TestScore(t *testing.T) {
	h := &internal.Harsh{
		Habits: []*storage.Habit{
			{Name: "Test1", Target: 1, Interval: 1},
			{Name: "Test2", Target: 1, Interval: 1},
			{Name: "Test3", Target: 1, Interval: 1},
			{Name: "Test4", Target: 1, Interval: 1},
		},
		Log: &storage.Log { Entries: storage.Entries{}, Header: storage.DefaultHeader },
	}

	today := civil.DateOf(time.Now())
	h.Log.Entries[storage.DailyHabit{Day: today, Habit: "Test1"}] = storage.Outcome{Result: "y"}
	h.Log.Entries[storage.DailyHabit{Day: today, Habit: "Test2"}] = storage.Outcome{Result: "y"}
	h.Log.Entries[storage.DailyHabit{Day: today, Habit: "Test3"}] = storage.Outcome{Result: "n"}
	h.Log.Entries[storage.DailyHabit{Day: today, Habit: "Test4"}] = storage.Outcome{Result: "y"}

	score := graph.Score(today, h.GetHabits(), &h.GetLog().Entries)
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
	storage.CreateExampleHabitsFile(tmpDir)
	if _, err := os.Stat(filepath.Join(tmpDir, "habits")); os.IsNotExist(err) {
		t.Error("Habits file was not created")
	}

	// Test log file creation
	storage.CreateNewLogFile(tmpDir)
	if _, err := os.Stat(filepath.Join(tmpDir, "log")); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestBuildGraph(t *testing.T) {
	log := &storage.Log {Entries: storage.Entries{}}
	h := &internal.Harsh{
		CountBack: 7,
		Log:   log,
	}

	habit := &storage.Habit{
		Name:        "Test",
		Target:      1,
		Interval:    1,
		FirstRecord: civil.DateOf(time.Now()).AddDays(-10),
	}

	today := civil.DateOf(time.Now())
	log.Entries[storage.DailyHabit{Day: today, Habit: "Test"}] = storage.Outcome{Result: "y"}
	log.Entries[storage.DailyHabit{Day: today.AddDays(-1), Habit: "Test"}] = storage.Outcome{Result: "y"}

	graphResult := graph.BuildGraph(habit, &h.GetLog().Entries, h.GetCountBack(), false)
	length := utf8.RuneCountInString(graphResult)
	// Calculate the actual expected length based on the buildGraph logic
	to := civil.DateOf(time.Now())
	from := to.AddDays(-h.CountBack)
	expectedLength := 0
	for d := from; !d.After(to); d = d.AddDays(1) {
		expectedLength++
	}
	if length != expectedLength {
		t.Errorf("Expected graph length %d, got %d. CountBack=%d, from=%s, to=%s, graph=%q", expectedLength, length, h.GetCountBack(), from.String(), to.String(), graphResult)
	}
}

func TestWarning(t *testing.T) {
	entries := storage.Entries{}
	today := civil.DateOf(time.Now())

	habit := &storage.Habit{
		Name:        "Test",
		Target:      1,
		Interval:    7,
		FirstRecord: today.AddDays(-10),
	}

	if !graph.Warning(today, habit, entries) {
		t.Error("Expected warning for habit with no entries")
	}

	entries[storage.DailyHabit{Day: today, Habit: "Test"}] = storage.Outcome{Result: "y"}
	if graph.Warning(today, habit, entries) {
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
	originalConfigDir := tmpDir
	tmpDir = tmpDir
	defer func() { tmpDir = originalConfigDir }()

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

	// Load initial configuration using component functions directly
	// to avoid terminal size issues in tests
	habits, maxHabitNameLength := storage.LoadHabitsConfig(tmpDir)
	log := storage.LoadLog(tmpDir)
	now := civil.DateOf(time.Now())
	to := now
	from := to.AddDays(-365 * 5)
	log.Entries.FirstRecords(from, to, habits)

	harsh := &internal.Harsh{
		Habits:             habits,
		MaxHabitNameLength: maxHabitNameLength,
		CountBack:          100, // Set a default for testing
		Log:            log,
	}

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
	habits, maxHabitNameLength = storage.LoadHabitsConfig(tmpDir)
	log = storage.LoadLog(tmpDir)
	log.Entries.FirstRecords(from, to, habits)

	harsh = &internal.Harsh{
		Habits:             habits,
		MaxHabitNameLength: maxHabitNameLength,
		CountBack:          100,
		Log:            log,
	}

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
	todos := ui.GetTodos(harsh.GetHabits(), &harsh.GetLog().Entries, today, 0)

	foundInTodos := false
	for _, todoList := range todos {
		for _, todo := range todoList {
			if todo == "New habit" {
				foundInTodos = true
				break
			}
		}
	}
	if !foundInTodos {
		t.Error("New habit was not found in todos")
	}
}

// Test the skipified function
func TestSkipified(t *testing.T) {
	tests := []struct {
		name    string
		d       civil.Date
		habit   storage.Habit
		entries storage.Entries
		want    bool
	}{
		{
			name:  "Daily habit should always return false",
			d:     civil.Date{Year: 2025, Month: 3, Day: 15},
			habit: storage.Habit{Name: "Daily", Target: 1, Interval: 1},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 14}, Habit: "Daily"}: {Result: "s"},
			},
			want: false,
		},
		{
			name:  "Weekly habit with skip should return true",
			d:     civil.Date{Year: 2025, Month: 3, Day: 15},
			habit: storage.Habit{Name: "Weekly", Target: 1, Interval: 7},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 14}, Habit: "Weekly"}: {Result: "s"},
			},
			want: true,
		},
		{
			name:  "Weekly habit without skip should return false",
			d:     civil.Date{Year: 2025, Month: 3, Day: 15},
			habit: storage.Habit{Name: "Weekly", Target: 1, Interval: 7},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 14}, Habit: "Weekly"}: {Result: "y"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := graph.Skipified(tt.d, &tt.habit, tt.entries)
			if got != tt.want {
				t.Errorf("graph.Skipified() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test the scoring function with more complex scenarios
func TestScoreComplex(t *testing.T) {
	tests := []struct {
		name     string
		habits   []*storage.Habit
		entries  storage.Entries
		date     civil.Date
		expected float64
	}{
		{
			name: "All habits completed",
			habits: []*storage.Habit{
				{Name: "Test1", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 3, Day: 1}},
				{Name: "Test2", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 3, Day: 1}},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Test1"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Test2"}: {Result: "y"},
			},
			date:     civil.Date{Year: 2025, Month: 3, Day: 15},
			expected: 100.0,
		},
		{
			name: "Half habits completed",
			habits: []*storage.Habit{
				{Name: "Test1", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 3, Day: 1}},
				{Name: "Test2", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 3, Day: 1}},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Test1"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Test2"}: {Result: "n"},
			},
			date:     civil.Date{Year: 2025, Month: 3, Day: 15},
			expected: 50.0,
		},
		{
			name: "One habit skipped should be excluded from score",
			habits: []*storage.Habit{
				{Name: "Test1", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 3, Day: 1}},
				{Name: "Test2", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 3, Day: 1}},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Test1"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Test2"}: {Result: "s"},
			},
			date:     civil.Date{Year: 2025, Month: 3, Day: 15},
			expected: 100.0, // Only Test1 counts, and it's completed
		},
		{
			name: "Tracking habits should not affect score",
			habits: []*storage.Habit{
				{Name: "Test1", Target: 1, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 3, Day: 1}},
				{Name: "Track", Target: 0, Interval: 1, FirstRecord: civil.Date{Year: 2025, Month: 3, Day: 1}},
			},
			entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Test1"}: {Result: "y"},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 15}, Habit: "Track"}: {Result: "y"},
			},
			date:     civil.Date{Year: 2025, Month: 3, Day: 15},
			expected: 100.0, // Only Test1 counts for scoring
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := graph.Score(tt.date, tt.habits, &tt.entries)
			if score != tt.expected {
				t.Errorf("score() = %f, want %f", score, tt.expected)
			}
		})
	}
}

// Test the buildStats function
func TestBuildStats(t *testing.T) {
	h := &internal.Harsh{
		Log: &storage.Log {
			Header: storage.DefaultHeader,
			Entries: storage.Entries{
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 10}, Habit: "Test"}: {Result: "y", Amount: 5.0},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 11}, Habit: "Test"}: {Result: "y", Amount: 3.0},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 12}, Habit: "Test"}: {Result: "n", Amount: 0.0},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 13}, Habit: "Test"}: {Result: "s", Amount: 0.0},
				storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 3, Day: 14}, Habit: "Test"}: {Result: "y", Amount: 2.0},
			},
		},
	}

	habit := &storage.Habit{
		Name:        "Test",
		Target:      1,
		Interval:    1,
		FirstRecord: civil.Date{Year: 2025, Month: 3, Day: 10},
	}

	stats := ui.BuildStats(habit, &h.GetLog().Entries)

	// Check the stats
	if stats.Streaks != 3 {
		t.Errorf("Expected 3 streaks, got %d", stats.Streaks)
	}
	if stats.Breaks != 1 {
		t.Errorf("Expected 1 break, got %d", stats.Breaks)
	}
	if stats.Skips != 1 {
		t.Errorf("Expected 1 skip, got %d", stats.Skips)
	}
	if stats.Total != 10.0 {
		t.Errorf("Expected total 10.0, got %f", stats.Total)
	}
	// DaysTracked should be from FirstRecord to today
	expectedDays := int(civil.DateOf(time.Now()).DaysSince(habit.FirstRecord)) + 1
	if stats.DaysTracked != expectedDays {
		t.Errorf("Expected %d days tracked, got %d", expectedDays, stats.DaysTracked)
	}
}

// Test CLI command execution
func TestCLICommands(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_cli_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original config dir and restore it after test
	originalConfigDir := tmpDir
	tmpDir = tmpDir
	defer func() { tmpDir = originalConfigDir }()

	// Create test habits file
	habitsFile := filepath.Join(tmpDir, "habits")
	err = os.WriteFile(habitsFile, []byte("Test habit: 1\nWeekly habit: 7\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create empty log file
	logFile := filepath.Join(tmpDir, "log")
	err = os.WriteFile(logFile, []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test help command
	t.Run("help command", func(t *testing.T) {
		// Create a new root command for testing to avoid state pollution
		testCmd := &cobra.Command{
			Use:     "harsh",
			Short:   "habit tracking for geeks",
			Long:    "A simple, minimalist CLI for tracking and understanding habits.",
			Version: "0.10.22",
		}
		testCmd.SetArgs([]string{"--help"})
		var buf bytes.Buffer
		testCmd.SetOut(&buf)
		err := testCmd.Execute()
		if err != nil {
			t.Fatalf("Help command failed: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "habit tracking for geeks") && !strings.Contains(output, "A simple, minimalist CLI") {
			t.Errorf("Help output should contain app description. Got: %s", output)
		}
	})

	// Test version command
	t.Run("version command", func(t *testing.T) {
		// Create a new root command for testing to avoid state pollution
		testCmd := &cobra.Command{
			Use:     "harsh",
			Short:   "habit tracking for geeks",
			Long:    "A simple, minimalist CLI for tracking and understanding habits.",
			Version: "0.10.22",
		}
		testCmd.SetArgs([]string{"--version"})
		var buf bytes.Buffer
		testCmd.SetOut(&buf)
		err := testCmd.Execute()
		if err != nil {
			t.Fatalf("Version command failed: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "0.10.22") {
			t.Errorf("Version output should contain version number. Got: %s", output)
		}
	})

	// Reset command args for other tests
	cmd.RootCmd.SetArgs([]string{})
}

// Test parallel graph building
func TestBuildGraphsParallel(t *testing.T) {
	log := &storage.Log {
		Entries: storage.Entries{},
		Header: storage.DefaultHeader,
	}
	h := &internal.Harsh{
		CountBack: 7,
		Log:   log,
	}

	habits := []*storage.Habit{
		{Name: "Test1", Target: 1, Interval: 1, FirstRecord: civil.DateOf(time.Now()).AddDays(-10)},
		{Name: "Test2", Target: 1, Interval: 1, FirstRecord: civil.DateOf(time.Now()).AddDays(-10)},
		{Name: "Test3", Target: 1, Interval: 1, FirstRecord: civil.DateOf(time.Now()).AddDays(-10)},
	}

	today := civil.DateOf(time.Now())
	log.Entries[storage.DailyHabit{Day: today, Habit: "Test1"}] = storage.Outcome{Result: "y"}
	log.Entries[storage.DailyHabit{Day: today, Habit: "Test2"}] = storage.Outcome{Result: "n"}
	log.Entries[storage.DailyHabit{Day: today, Habit: "Test3"}] = storage.Outcome{Result: "s"}

	results := graph.BuildGraphsParallel(habits, &h.GetLog().Entries, h.GetCountBack(), false)

	// Check that all habits have results
	for _, habit := range habits {
		if graph, ok := results[habit.Name]; !ok || graph == "" {
			t.Errorf("Expected graph for habit %s, but got empty result", habit.Name)
		}
	}

	// Check that all graphs have expected length
	to := civil.DateOf(time.Now())
	from := to.AddDays(-h.CountBack)
	expectedLength := 0
	for d := from; !d.After(to); d = d.AddDays(1) {
		expectedLength++
	}
	for habitName, graph := range results {
		if utf8.RuneCountInString(graph) != expectedLength {
			t.Errorf("Graph for %s has length %d, expected %d", habitName, utf8.RuneCountInString(graph), expectedLength)
		}
	}
}
