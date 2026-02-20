package test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal/storage"
	"github.com/wakatara/harsh/internal/ui"
)

// jsonLog mirrors the JSON output structure for unmarshaling in tests
type jsonLog struct {
	Date   string `json:"date"`
	Scores struct {
		Today     float64 `json:"today"`
		Yesterday float64 `json:"yesterday"`
	} `json:"scores"`
	Habits []jsonHabit `json:"habits"`
}

type jsonHabit struct {
	Name              string  `json:"name"`
	Heading           string  `json:"heading"`
	Frequency         string  `json:"frequency"`
	Target            int     `json:"target"`
	Interval          int     `json:"interval"`
	LoggedToday       bool    `json:"logged_today"`
	Result            *string `json:"result"`
	StreakStatus       string  `json:"streak_status"`
	DaysUntilBreak    *int    `json:"days_until_break"`
	LastCompleted     *string `json:"last_completed"`
	CompletedInWindow *int    `json:"completed_in_window,omitempty"`
	Stats             struct {
		DaysTracked int     `json:"days_tracked"`
		Streaks     int     `json:"streaks"`
		Breaks      int     `json:"breaks"`
		Skips       int     `json:"skips"`
		Total       float64 `json:"total"`
	} `json:"stats"`
}

func captureJSONOutput(t *testing.T, fn func() error) []byte {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := fn()
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	return buf.Bytes()
}

func TestShowHabitLogJSON_ValidJSON(t *testing.T) {
	now := civil.DateOf(time.Now())
	habits := []*storage.Habit{
		{Name: "Test1", Heading: "Work", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-10)},
		{Name: "Test2", Heading: "Health", Frequency: "3/7", Target: 3, Interval: 7, FirstRecord: now.AddDays(-10)},
	}

	entries := &storage.Entries{
		storage.DailyHabit{Day: now, Habit: "Test1"}: {Result: "y", Amount: 5.0},
	}

	output := captureJSONOutput(t, func() error {
		return ui.ShowHabitLogJSON(habits, entries, "", false)
	})

	var result jsonLog
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, string(output))
	}

	if result.Date == "" {
		t.Error("Date should not be empty")
	}

	if len(result.Habits) != 2 {
		t.Errorf("Expected 2 habits, got %d", len(result.Habits))
	}
}

func TestShowHabitLogJSON_LoggedToday(t *testing.T) {
	now := civil.DateOf(time.Now())
	habits := []*storage.Habit{
		{Name: "Logged", Heading: "Test", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-10)},
		{Name: "NotLogged", Heading: "Test", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-10)},
	}

	entries := &storage.Entries{
		storage.DailyHabit{Day: now, Habit: "Logged"}: {Result: "y"},
	}

	output := captureJSONOutput(t, func() error {
		return ui.ShowHabitLogJSON(habits, entries, "", false)
	})

	var result jsonLog
	json.Unmarshal(output, &result)

	var logged, notLogged *jsonHabit
	for i := range result.Habits {
		if result.Habits[i].Name == "Logged" {
			logged = &result.Habits[i]
		}
		if result.Habits[i].Name == "NotLogged" {
			notLogged = &result.Habits[i]
		}
	}

	if logged == nil || notLogged == nil {
		t.Fatal("Expected both habits in output")
	}

	if !logged.LoggedToday {
		t.Error("Logged habit should have logged_today=true")
	}
	if logged.Result == nil || *logged.Result != "y" {
		t.Error("Logged habit should have result='y'")
	}

	if notLogged.LoggedToday {
		t.Error("NotLogged habit should have logged_today=false")
	}
	if notLogged.Result != nil {
		t.Error("NotLogged habit should have result=null")
	}
}

func TestShowHabitLogJSON_FragmentFilter(t *testing.T) {
	now := civil.DateOf(time.Now())
	habits := []*storage.Habit{
		{Name: "Morning Run", Heading: "Fitness", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-10)},
		{Name: "Evening Read", Heading: "Learning", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-10)},
	}

	entries := &storage.Entries{}

	output := captureJSONOutput(t, func() error {
		return ui.ShowHabitLogJSON(habits, entries, "morn", false)
	})

	var result jsonLog
	json.Unmarshal(output, &result)

	if len(result.Habits) != 1 {
		t.Errorf("Expected 1 filtered habit, got %d", len(result.Habits))
	}
	if len(result.Habits) > 0 && result.Habits[0].Name != "Morning Run" {
		t.Errorf("Expected 'Morning Run', got '%s'", result.Habits[0].Name)
	}
}

func TestShowHabitLogJSON_HideEnded(t *testing.T) {
	now := civil.DateOf(time.Now())
	habits := []*storage.Habit{
		{Name: "Active", Heading: "Test", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-10)},
		{Name: "Retired", Heading: "Test", Frequency: "1", Target: 1, Interval: 1,
			FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
			EndRecord:   civil.Date{Year: 2025, Month: 6, Day: 1}},
	}

	entries := &storage.Entries{}

	// With hideEnded=true
	output := captureJSONOutput(t, func() error {
		return ui.ShowHabitLogJSON(habits, entries, "", true)
	})

	var result jsonLog
	json.Unmarshal(output, &result)

	if len(result.Habits) != 1 {
		t.Errorf("Expected 1 habit with hideEnded, got %d", len(result.Habits))
	}
	if len(result.Habits) > 0 && result.Habits[0].Name != "Active" {
		t.Errorf("Expected 'Active', got '%s'", result.Habits[0].Name)
	}

	// With hideEnded=false, should show both
	output2 := captureJSONOutput(t, func() error {
		return ui.ShowHabitLogJSON(habits, entries, "", false)
	})

	var result2 jsonLog
	json.Unmarshal(output2, &result2)

	if len(result2.Habits) != 2 {
		t.Errorf("Expected 2 habits without hideEnded, got %d", len(result2.Habits))
	}
}

func TestShowHabitLogJSON_CompletedInWindow(t *testing.T) {
	now := civil.DateOf(time.Now())
	habits := []*storage.Habit{
		{Name: "Daily", Heading: "Test", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-10)},
		{Name: "Weekly", Heading: "Test", Frequency: "3/7", Target: 3, Interval: 7, FirstRecord: now.AddDays(-10)},
		{Name: "Tracking", Heading: "Test", Frequency: "0", Target: 0, Interval: 1, FirstRecord: now.AddDays(-10)},
	}

	entries := &storage.Entries{
		storage.DailyHabit{Day: now, Habit: "Weekly"}:            {Result: "y"},
		storage.DailyHabit{Day: now.AddDays(-2), Habit: "Weekly"}: {Result: "y"},
	}

	output := captureJSONOutput(t, func() error {
		return ui.ShowHabitLogJSON(habits, entries, "", false)
	})

	var result jsonLog
	json.Unmarshal(output, &result)

	for _, h := range result.Habits {
		switch h.Name {
		case "Daily":
			if h.CompletedInWindow != nil {
				t.Error("Daily habit should not have completed_in_window")
			}
		case "Weekly":
			if h.CompletedInWindow == nil {
				t.Error("Weekly habit should have completed_in_window")
			} else if *h.CompletedInWindow != 2 {
				t.Errorf("Weekly habit completed_in_window: expected 2, got %d", *h.CompletedInWindow)
			}
		case "Tracking":
			if h.CompletedInWindow != nil {
				t.Error("Tracking habit should not have completed_in_window")
			}
		}
	}
}

func TestShowHabitLogJSON_Scores(t *testing.T) {
	now := civil.DateOf(time.Now())
	habits := []*storage.Habit{
		{Name: "Test1", Heading: "Work", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-10)},
		{Name: "Test2", Heading: "Work", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-10)},
	}

	entries := &storage.Entries{
		storage.DailyHabit{Day: now, Habit: "Test1"}:            {Result: "y"},
		storage.DailyHabit{Day: now, Habit: "Test2"}:            {Result: "y"},
		storage.DailyHabit{Day: now.AddDays(-1), Habit: "Test1"}: {Result: "y"},
		storage.DailyHabit{Day: now.AddDays(-1), Habit: "Test2"}: {Result: "n"},
	}

	output := captureJSONOutput(t, func() error {
		return ui.ShowHabitLogJSON(habits, entries, "", false)
	})

	var result jsonLog
	json.Unmarshal(output, &result)

	if result.Scores.Today != 100.0 {
		t.Errorf("Expected today score 100.0, got %f", result.Scores.Today)
	}

	if result.Scores.Yesterday != 50.0 {
		t.Errorf("Expected yesterday score 50.0, got %f", result.Scores.Yesterday)
	}
}

func TestShowHabitLogJSON_Stats(t *testing.T) {
	now := civil.DateOf(time.Now())
	firstRecord := now.AddDays(-5)
	habits := []*storage.Habit{
		{Name: "Test", Heading: "Work", Frequency: "1", Target: 1, Interval: 1, FirstRecord: firstRecord},
	}

	entries := &storage.Entries{
		storage.DailyHabit{Day: firstRecord, Habit: "Test"}:            {Result: "y", Amount: 5.0},
		storage.DailyHabit{Day: firstRecord.AddDays(1), Habit: "Test"}: {Result: "y", Amount: 3.0},
		storage.DailyHabit{Day: firstRecord.AddDays(2), Habit: "Test"}: {Result: "n"},
		storage.DailyHabit{Day: firstRecord.AddDays(3), Habit: "Test"}: {Result: "s"},
		storage.DailyHabit{Day: firstRecord.AddDays(4), Habit: "Test"}: {Result: "y", Amount: 2.0},
	}

	output := captureJSONOutput(t, func() error {
		return ui.ShowHabitLogJSON(habits, entries, "", false)
	})

	var result jsonLog
	json.Unmarshal(output, &result)

	if len(result.Habits) != 1 {
		t.Fatalf("Expected 1 habit, got %d", len(result.Habits))
	}

	stats := result.Habits[0].Stats
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
}

func TestShowHabitLogJSON_StreakStatus(t *testing.T) {
	now := civil.DateOf(time.Now())

	tests := []struct {
		name           string
		habits         []*storage.Habit
		entries        *storage.Entries
		expectedStatus string
		expectDays     bool
	}{
		{
			name: "skipping",
			habits: []*storage.Habit{
				{Name: "Weekly", Heading: "Test", Frequency: "1w", Target: 1, Interval: 7, FirstRecord: now.AddDays(-30)},
			},
			entries: &storage.Entries{
				storage.DailyHabit{Day: now.AddDays(-2), Habit: "Weekly"}: {Result: "s"},
			},
			expectedStatus: "skipping",
			expectDays:     true,
		},
		{
			name: "broken",
			habits: []*storage.Habit{
				{Name: "Daily", Heading: "Test", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-30)},
			},
			entries: &storage.Entries{
				storage.DailyHabit{Day: now.AddDays(-5), Habit: "Daily"}: {Result: "y"},
			},
			expectedStatus: "broken",
			expectDays:     false,
		},
		{
			name: "tracking",
			habits: []*storage.Habit{
				{Name: "Coffee", Heading: "Test", Frequency: "0", Target: 0, Interval: 1, FirstRecord: now.AddDays(-10)},
			},
			entries:        &storage.Entries{},
			expectedStatus: "tracking",
			expectDays:     false,
		},
		{
			name: "active",
			habits: []*storage.Habit{
				{Name: "Daily", Heading: "Test", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-10)},
			},
			entries: &storage.Entries{
				storage.DailyHabit{Day: now, Habit: "Daily"}: {Result: "y"},
			},
			expectedStatus: "active",
			expectDays:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureJSONOutput(t, func() error {
				return ui.ShowHabitLogJSON(tt.habits, tt.entries, "", false)
			})

			var result jsonLog
			json.Unmarshal(output, &result)

			if len(result.Habits) != 1 {
				t.Fatalf("Expected 1 habit, got %d", len(result.Habits))
			}

			h := result.Habits[0]
			if h.StreakStatus != tt.expectedStatus {
				t.Errorf("Expected streak_status=%q, got %q", tt.expectedStatus, h.StreakStatus)
			}

			if tt.expectDays && h.DaysUntilBreak == nil {
				t.Error("Expected days_until_break to have a value")
			}
			if !tt.expectDays && h.DaysUntilBreak != nil {
				t.Errorf("Expected days_until_break=null, got %d", *h.DaysUntilBreak)
			}
		})
	}
}

func TestShowHabitLogJSON_MultiDayIntervalStreak(t *testing.T) {
	now := civil.DateOf(time.Now())

	tests := []struct {
		name               string
		entries            *storage.Entries
		expectedStatus     string
		expectedWindow     int
		expectDaysNotNil   bool
	}{
		{
			name: "partially completed, target not met so broken",
			entries: &storage.Entries{
				storage.DailyHabit{Day: now, Habit: "Gym"}:            {Result: "y"},
				storage.DailyHabit{Day: now.AddDays(-2), Habit: "Gym"}: {Result: "y"},
			},
			expectedStatus:   "broken",
			expectedWindow:   2,
			expectDaysNotNil: false,
		},
		{
			name: "target fully met in window",
			entries: &storage.Entries{
				storage.DailyHabit{Day: now, Habit: "Gym"}:            {Result: "y"},
				storage.DailyHabit{Day: now.AddDays(-2), Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: now.AddDays(-4), Habit: "Gym"}: {Result: "y"},
			},
			expectedStatus:   "active",
			expectedWindow:   3,
			expectDaysNotNil: true,
		},
		{
			name: "no completions, streak broken",
			entries: &storage.Entries{
				storage.DailyHabit{Day: now.AddDays(-10), Habit: "Gym"}: {Result: "y"},
			},
			expectedStatus:   "broken",
			expectedWindow:   0,
			expectDaysNotNil: false,
		},
		{
			name: "skip in window counts toward completion",
			entries: &storage.Entries{
				storage.DailyHabit{Day: now, Habit: "Gym"}:            {Result: "y"},
				storage.DailyHabit{Day: now.AddDays(-1), Habit: "Gym"}: {Result: "s"},
				storage.DailyHabit{Day: now.AddDays(-3), Habit: "Gym"}: {Result: "y"},
			},
			expectedStatus:   "active",
			expectedWindow:   3,
			expectDaysNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			habits := []*storage.Habit{
				{Name: "Gym", Heading: "Fitness", Frequency: "3/7", Target: 3, Interval: 7, FirstRecord: now.AddDays(-30)},
			}

			output := captureJSONOutput(t, func() error {
				return ui.ShowHabitLogJSON(habits, tt.entries, "", false)
			})

			var result jsonLog
			json.Unmarshal(output, &result)

			h := result.Habits[0]

			if h.StreakStatus != tt.expectedStatus {
				t.Errorf("Expected streak_status=%q, got %q", tt.expectedStatus, h.StreakStatus)
			}

			if h.CompletedInWindow == nil {
				t.Fatal("Expected completed_in_window for 3/7 habit")
			}
			if *h.CompletedInWindow != tt.expectedWindow {
				t.Errorf("Expected completed_in_window=%d, got %d", tt.expectedWindow, *h.CompletedInWindow)
			}

			if tt.expectDaysNotNil && h.DaysUntilBreak == nil {
				t.Error("Expected days_until_break to have a value")
			}
			if !tt.expectDaysNotNil && h.DaysUntilBreak != nil {
				t.Errorf("Expected days_until_break=null, got %d", *h.DaysUntilBreak)
			}

			// Verify target is included so agent can compute remaining
			if h.Target != 3 {
				t.Errorf("Expected target=3, got %d", h.Target)
			}
		})
	}
}

func TestShowHabitLogJSON_SlidingWindowEdgeCases(t *testing.T) {
	now := civil.DateOf(time.Now())

	tests := []struct {
		name               string
		target             int
		interval           int
		frequency          string
		entries            *storage.Entries
		expectedStatus     string
		expectedWindow     int
		expectDaysNotNil   bool
		minDaysUntilBreak  int // only checked if expectDaysNotNil
	}{
		{
			// 3 completions exactly at the start of the trailing window
			// day -6, -5, -4: all within [now-6, now] window
			name:      "completions at window start, about to age out",
			target:    3,
			interval:  7,
			frequency: "3/7",
			entries: &storage.Entries{
				storage.DailyHabit{Day: now.AddDays(-6), Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: now.AddDays(-5), Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: now.AddDays(-4), Habit: "Gym"}: {Result: "y"},
			},
			expectedStatus:    "active",
			expectedWindow:    3,
			expectDaysNotNil:  true,
			minDaysUntilBreak: 1, // earliest (day -6) ages out at day +1
		},
		{
			// Completion on day -7 is OUTSIDE the trailing window [now-6, now]
			// so completedInWindow should NOT count it
			name:      "completion just outside window boundary",
			target:    3,
			interval:  7,
			frequency: "3/7",
			entries: &storage.Entries{
				storage.DailyHabit{Day: now.AddDays(-7), Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: now.AddDays(-3), Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: now, Habit: "Gym"}:             {Result: "y"},
			},
			// Window [now-6, now] has 2 (day -3 and now). day -7 is outside.
			// But Satisfied() checks window [now-7, now-1] which has day -7 and day -3 = 2. Not 3.
			// No window has 3 completions. Broken.
			expectedStatus:   "broken",
			expectedWindow:   2,
			expectDaysNotNil: false,
		},
		{
			// All completions at the end of window (recent)
			// Should have maximum days_until_break
			name:      "completions at window end, maximum buffer",
			target:    3,
			interval:  7,
			frequency: "3/7",
			entries: &storage.Entries{
				storage.DailyHabit{Day: now, Habit: "Gym"}:            {Result: "y"},
				storage.DailyHabit{Day: now.AddDays(-1), Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: now.AddDays(-2), Habit: "Gym"}: {Result: "y"},
			},
			expectedStatus:    "active",
			expectedWindow:    3,
			expectDaysNotNil:  true,
			minDaysUntilBreak: 5, // earliest (day -2) ages out at day +5
		},
		{
			// 2/7 habit: exactly 2 completions, satisfied
			name:      "2/7 habit with target met",
			target:    2,
			interval:  7,
			frequency: "2/7",
			entries: &storage.Entries{
				storage.DailyHabit{Day: now, Habit: "Gym"}:            {Result: "y"},
				storage.DailyHabit{Day: now.AddDays(-3), Habit: "Gym"}: {Result: "y"},
			},
			expectedStatus:    "active",
			expectedWindow:    2,
			expectDaysNotNil:  true,
			minDaysUntilBreak: 4, // earliest (day -3) ages out at day +4
		},
		{
			// Completions spread so one is about to age out today
			// day -6: y (ages out when window shifts to [now-5, now+1])
			// But from today's perspective, [now-6, now] has it
			name:      "completion aging out tomorrow",
			target:    3,
			interval:  7,
			frequency: "3/7",
			entries: &storage.Entries{
				storage.DailyHabit{Day: now.AddDays(-6), Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: now.AddDays(-3), Habit: "Gym"}: {Result: "y"},
				storage.DailyHabit{Day: now, Habit: "Gym"}:             {Result: "y"},
			},
			expectedStatus:    "active",
			expectedWindow:    3,
			expectDaysNotNil:  true,
			minDaysUntilBreak: 1, // day -6 ages out at day +1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			habits := []*storage.Habit{
				{Name: "Gym", Heading: "Fitness", Frequency: tt.frequency, Target: tt.target, Interval: tt.interval, FirstRecord: now.AddDays(-30)},
			}

			output := captureJSONOutput(t, func() error {
				return ui.ShowHabitLogJSON(habits, tt.entries, "", false)
			})

			var result jsonLog
			json.Unmarshal(output, &result)

			h := result.Habits[0]

			if h.StreakStatus != tt.expectedStatus {
				t.Errorf("Expected streak_status=%q, got %q", tt.expectedStatus, h.StreakStatus)
			}

			if h.CompletedInWindow == nil {
				t.Fatal("Expected completed_in_window for multi-day habit")
			}
			if *h.CompletedInWindow != tt.expectedWindow {
				t.Errorf("Expected completed_in_window=%d, got %d", tt.expectedWindow, *h.CompletedInWindow)
			}

			if tt.expectDaysNotNil {
				if h.DaysUntilBreak == nil {
					t.Fatal("Expected days_until_break to have a value")
				}
				if *h.DaysUntilBreak < tt.minDaysUntilBreak {
					t.Errorf("Expected days_until_break >= %d, got %d", tt.minDaysUntilBreak, *h.DaysUntilBreak)
				}
			} else {
				if h.DaysUntilBreak != nil {
					t.Errorf("Expected days_until_break=null, got %d", *h.DaysUntilBreak)
				}
			}
		})
	}
}

func TestShowHabitLogJSON_HabitFields(t *testing.T) {
	now := civil.DateOf(time.Now())
	habits := []*storage.Habit{
		{Name: "Gym", Heading: "Fitness", Frequency: "3/7", Target: 3, Interval: 7, FirstRecord: now.AddDays(-10)},
	}

	entries := &storage.Entries{}

	output := captureJSONOutput(t, func() error {
		return ui.ShowHabitLogJSON(habits, entries, "", false)
	})

	var result jsonLog
	json.Unmarshal(output, &result)

	h := result.Habits[0]
	if h.Name != "Gym" {
		t.Errorf("Expected name 'Gym', got '%s'", h.Name)
	}
	if h.Heading != "Fitness" {
		t.Errorf("Expected heading 'Fitness', got '%s'", h.Heading)
	}
	if h.Frequency != "3/7" {
		t.Errorf("Expected frequency '3/7', got '%s'", h.Frequency)
	}
	if h.Target != 3 {
		t.Errorf("Expected target 3, got %d", h.Target)
	}
	if h.Interval != 7 {
		t.Errorf("Expected interval 7, got %d", h.Interval)
	}
}

func TestShowHabitLogJSON_LastCompleted(t *testing.T) {
	now := civil.DateOf(time.Now())

	tests := []struct {
		name          string
		entries       *storage.Entries
		expectedDate  *string
	}{
		{
			name: "completed today",
			entries: &storage.Entries{
				storage.DailyHabit{Day: now, Habit: "Test"}: {Result: "y"},
			},
			expectedDate: strPtr(now.String()),
		},
		{
			name: "completed 3 days ago",
			entries: &storage.Entries{
				storage.DailyHabit{Day: now.AddDays(-3), Habit: "Test"}: {Result: "y"},
			},
			expectedDate: strPtr(now.AddDays(-3).String()),
		},
		{
			name: "skip counts as completion",
			entries: &storage.Entries{
				storage.DailyHabit{Day: now.AddDays(-1), Habit: "Test"}: {Result: "s"},
				storage.DailyHabit{Day: now.AddDays(-5), Habit: "Test"}: {Result: "y"},
			},
			expectedDate: strPtr(now.AddDays(-1).String()),
		},
		{
			name: "only n entries, no completion",
			entries: &storage.Entries{
				storage.DailyHabit{Day: now, Habit: "Test"}:            {Result: "n"},
				storage.DailyHabit{Day: now.AddDays(-1), Habit: "Test"}: {Result: "n"},
			},
			expectedDate: nil,
		},
		{
			name:         "no entries at all",
			entries:      &storage.Entries{},
			expectedDate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			habits := []*storage.Habit{
				{Name: "Test", Heading: "Work", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-10)},
			}

			output := captureJSONOutput(t, func() error {
				return ui.ShowHabitLogJSON(habits, tt.entries, "", false)
			})

			var result jsonLog
			json.Unmarshal(output, &result)

			h := result.Habits[0]

			if tt.expectedDate == nil {
				if h.LastCompleted != nil {
					t.Errorf("Expected last_completed=null, got %q", *h.LastCompleted)
				}
			} else {
				if h.LastCompleted == nil {
					t.Errorf("Expected last_completed=%q, got null", *tt.expectedDate)
				} else if *h.LastCompleted != *tt.expectedDate {
					t.Errorf("Expected last_completed=%q, got %q", *tt.expectedDate, *h.LastCompleted)
				}
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func TestShowHabitLogJSON_EmptyHabits(t *testing.T) {
	habits := []*storage.Habit{}
	entries := &storage.Entries{}

	output := captureJSONOutput(t, func() error {
		return ui.ShowHabitLogJSON(habits, entries, "", false)
	})

	var result jsonLog
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if result.Habits == nil {
		t.Error("Habits should be empty array, not null")
	}
	if len(result.Habits) != 0 {
		t.Errorf("Expected 0 habits, got %d", len(result.Habits))
	}
}
