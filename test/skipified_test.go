package test

import (
	"testing"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal/graph"
	"github.com/wakatara/harsh/internal/storage"
	"github.com/wakatara/harsh/internal/ui"
)

// TestSkipifiedDailyHabitAlwaysFalse verifies daily habits (1/1) never report skipified.
func TestSkipifiedDailyHabitAlwaysFalse(t *testing.T) {
	habit := &storage.Habit{
		Name: "Daily", Target: 1, Interval: 1,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Daily"}: {Result: "s"},
	}
	if graph.Skipified(civil.Date{Year: 2026, Month: 3, Day: 6}, habit, entries) {
		t.Error("Skipified should always return false for daily (1/1) habits")
	}
}

// TestSkipifiedWeeklyHabitGracePeriod tests that a 1/7 skip covers the full 7-day interval.
func TestSkipifiedWeeklyHabitGracePeriod(t *testing.T) {
	habit := &storage.Habit{
		Name: "Weekly", Target: 1, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Weekly"}: {Result: "s"},
	}

	// Days 10 through 16 (7 days from skip) should be skipified
	for day := 10; day <= 16; day++ {
		d := civil.Date{Year: 2026, Month: 3, Day: day}
		if !graph.Skipified(d, habit, entries) {
			t.Errorf("Day %d should be skipified (within 7-day grace of skip on day 10)", day)
		}
	}

	// Day 17 should NOT be skipified (outside interval)
	d := civil.Date{Year: 2026, Month: 3, Day: 17}
	if graph.Skipified(d, habit, entries) {
		t.Error("Day 17 should NOT be skipified (outside 7-day grace period)")
	}
}

// TestSkipifiedMultiTargetInterval_3of7 is the core bug regression test.
// For a 3/7 habit, a skip should cover the full 7-day interval, not just ceil(7/3)=3 days.
func TestSkipifiedMultiTargetInterval_3of7(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 1, Day: 1},
	}
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Dev"}: {Result: "s"},
	}

	// All 7 days from the skip (day 10 through 16) should be skipified
	for day := 10; day <= 16; day++ {
		d := civil.Date{Year: 2026, Month: 3, Day: day}
		if !graph.Skipified(d, habit, entries) {
			t.Errorf("Day %d should be skipified (within 7-day grace of skip on day 10 for 3/7 habit)", day)
		}
	}

	// Day 17 should NOT be skipified
	d := civil.Date{Year: 2026, Month: 3, Day: 17}
	if graph.Skipified(d, habit, entries) {
		t.Error("Day 17 should NOT be skipified (outside 7-day grace for 3/7 habit)")
	}
}

// TestSkipifiedMultiTargetInterval_2of14 tests that a 2/14 habit skip covers 14 days.
func TestSkipifiedMultiTargetInterval_2of14(t *testing.T) {
	habit := &storage.Habit{
		Name: "Biweekly", Target: 2, Interval: 14,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Biweekly"}: {Result: "s"},
	}

	// Day 18 (14 days from skip) should still be skipified
	if !graph.Skipified(civil.Date{Year: 2026, Month: 3, Day: 18}, habit, entries) {
		t.Error("Day 18 should be skipified (within 14-day grace of skip on day 5 for 2/14 habit)")
	}

	// Day 19 (15 days from skip) should NOT be skipified
	if graph.Skipified(civil.Date{Year: 2026, Month: 3, Day: 19}, habit, entries) {
		t.Error("Day 19 should NOT be skipified (outside 14-day grace for 2/14 habit)")
	}
}

// TestSkipifiedExactBoundary tests the exact last day of the grace period.
func TestSkipifiedExactBoundary(t *testing.T) {
	tests := []struct {
		name     string
		target   int
		interval int
	}{
		{"1/7 weekly", 1, 7},
		{"3/7 three-per-week", 3, 7},
		{"2/14 biweekly", 2, 14},
		{"5/7 weekdays", 5, 7},
		{"1/30 monthly", 1, 30},
		{"3/30 tri-monthly", 3, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			habit := &storage.Habit{
				Name: "Test", Target: tt.target, Interval: tt.interval,
				FirstRecord: civil.Date{Year: 2026, Month: 1, Day: 1},
			}
			skipDay := civil.Date{Year: 2026, Month: 3, Day: 1}
			entries := storage.Entries{
				storage.DailyHabit{Day: skipDay, Habit: "Test"}: {Result: "s"},
			}

			// Last day of grace period (skip day + interval - 1 days later)
			lastGraceDay := skipDay.AddDays(tt.interval - 1)
			if !graph.Skipified(lastGraceDay, habit, entries) {
				t.Errorf("Last grace day (%s) should be skipified for %s habit", lastGraceDay, tt.name)
			}

			// One day beyond grace period
			beyondGrace := skipDay.AddDays(tt.interval)
			if graph.Skipified(beyondGrace, habit, entries) {
				t.Errorf("Day beyond grace (%s) should NOT be skipified for %s habit", beyondGrace, tt.name)
			}
		})
	}
}

// TestSkipifiedNoSkipEntryReturnsFalse verifies that non-skip entries don't trigger skipified.
func TestSkipifiedNoSkipEntryReturnsFalse(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 11}, Habit: "Dev"}: {Result: "n"},
	}

	d := civil.Date{Year: 2026, Month: 3, Day: 12}
	if graph.Skipified(d, habit, entries) {
		t.Error("Skipified should return false when there's no skip entry in the grace window")
	}
}

// TestSkipifiedEmptyEntriesReturnsFalse verifies no entries means no skipified.
func TestSkipifiedEmptyEntriesReturnsFalse(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}
	entries := storage.Entries{}

	d := civil.Date{Year: 2026, Month: 3, Day: 10}
	if graph.Skipified(d, habit, entries) {
		t.Error("Skipified should return false with empty entries")
	}
}

// TestSkipifiedMultipleSkipsExtendGrace tests that a second skip resets the grace window.
func TestSkipifiedMultipleSkipsExtendGrace(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Dev"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 17}, Habit: "Dev"}: {Result: "s"},
	}

	// Day 23 (6 days after second skip): within grace of skip on 17
	d := civil.Date{Year: 2026, Month: 3, Day: 23}
	if !graph.Skipified(d, habit, entries) {
		t.Error("Day 23 should be skipified (within grace of second skip on day 17)")
	}

	// Day 24 (7 days after second skip): should NOT be skipified
	d = civil.Date{Year: 2026, Month: 3, Day: 24}
	if graph.Skipified(d, habit, entries) {
		t.Error("Day 24 should NOT be skipified (outside grace of second skip)")
	}
}

// --- Interaction tests: skips, skipified, breaks, and restarts ---

// TestSkipThenBreakThenRestart tests the full lifecycle: skip grace expires, break occurs,
// then a new "y" restarts the streak.
func TestSkipThenBreakThenRestart(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}

	// Active streak, then skip, then nothing (break), then restart
	entries := storage.Entries{
		// Active period
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 1}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 2}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 3}, Habit: "Dev"}: {Result: "y"},
		// Skip on day 5
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Dev"}: {Result: "s"},
		// "n" entries through the grace period and beyond
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 6}, Habit: "Dev"}:  {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 7}, Habit: "Dev"}:  {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 8}, Habit: "Dev"}:  {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 9}, Habit: "Dev"}:  {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 11}, Habit: "Dev"}: {Result: "n"},
		// Day 12: beyond grace, this is a break
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 12}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 13}, Habit: "Dev"}: {Result: "n"},
		// Restart on day 15
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 15}, Habit: "Dev"}: {Result: "y"},
	}

	// Day 6-11: within 7-day grace of skip on day 5 → skipified
	for day := 6; day <= 11; day++ {
		d := civil.Date{Year: 2026, Month: 3, Day: day}
		if !graph.Skipified(d, habit, entries) {
			t.Errorf("Day %d should be skipified (within grace of skip on day 5)", day)
		}
	}

	// Day 12: outside grace → NOT skipified (break)
	if graph.Skipified(civil.Date{Year: 2026, Month: 3, Day: 12}, habit, entries) {
		t.Error("Day 12 should NOT be skipified (outside grace, this is a break)")
	}

	// Day 13: also a break
	if graph.Skipified(civil.Date{Year: 2026, Month: 3, Day: 13}, habit, entries) {
		t.Error("Day 13 should NOT be skipified (break territory)")
	}

	// Day 15: restart with "y" — not skipified (it's a "y" so the graph shows done, not skipified)
	// Skipified is only relevant for "n" entries checked via BuildGraph
	if graph.Skipified(civil.Date{Year: 2026, Month: 3, Day: 15}, habit, entries) {
		t.Error("Day 15 with y result should not need skipified check")
	}
}

// TestSkipCannotRestartBrokenStreak_Windowed tests that for multi-target habits,
// a skip after a break doesn't magically produce skipified on subsequent days
// unless there was a prior streak to maintain.
func TestSkipCannotRestartBrokenStreak_Windowed(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}

	// Long break, then a skip (no prior streak within interval)
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 1}, Habit: "Dev"}: {Result: "y"},
		// Big gap — streak is broken by the time we get to day 20
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 20}, Habit: "Dev"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 21}, Habit: "Dev"}: {Result: "n"},
	}

	// The skip on day 20 IS found in the grace window, so Skipified returns true
	// (Skipified doesn't check streak state, it just checks for nearby skips)
	// But DaysUntilStreakBreak should show broken because skip can't restart streak
	daysUntil := graph.DaysUntilStreakBreak(civil.Date{Year: 2026, Month: 3, Day: 21}, habit, entries)
	if daysUntil >= 0 {
		t.Errorf("DaysUntilStreakBreak should be negative (broken) after gap, got %d", daysUntil)
	}
}

// TestSkipifiedInteractionWithSatisfied tests that a skip always resets the
// display to skipified, even when genuine completions exist in the window.
// The skip is the most recent action, so it takes precedence.
func TestSkipifiedInteractionWithSatisfied(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}

	// 3 genuine completions THEN a skip: skip resets display
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 1}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 3}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 6}, Habit: "Dev"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 7}, Habit: "Dev"}: {Result: "n"},
	}

	d := civil.Date{Year: 2026, Month: 3, Day: 7}

	// SatisfiedByCompletions: 3 y's in window -> true
	if !graph.SatisfiedByCompletions(d, habit, entries) {
		t.Error("SatisfiedByCompletions should be true (3 y's in window)")
	}

	// Skipified is also true (skip on day 6)
	if !graph.Skipified(d, habit, entries) {
		t.Error("Skipified should also be true (skip on day 6)")
	}

	// IsInSkipPeriod: skip on day 6 is more recent than y on day 5
	if !graph.IsInSkipPeriod(d, habit, entries) {
		t.Error("IsInSkipPeriod should be true (s on day 6 is most recent)")
	}

	// Display: skipified wins because skip is more recent than any completion
	if s := displayStatus(d, habit, entries); s != "skipified" {
		t.Errorf("Display should be 'skipified' (skip resets display), got '%s'", s)
	}
}

// TestSkipifiedInteractionWithSatisfied_YAfterSkip tests that when a y appears
// AFTER a skip, satisfied display resumes for subsequent n's.
func TestSkipifiedInteractionWithSatisfied_YAfterSkip(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}

	// y y y s y n — the y AFTER the skip restores satisfied
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 1}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 3}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 6}, Habit: "Dev"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 7}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 8}, Habit: "Dev"}: {Result: "n"},
	}

	d := civil.Date{Year: 2026, Month: 3, Day: 8}

	// IsInSkipPeriod: y on day 7 is more recent than s on day 6 -> false
	if graph.IsInSkipPeriod(d, habit, entries) {
		t.Error("IsInSkipPeriod should be false (y on day 7 is more recent)")
	}

	// Display: satisfied (y on day 7 is most recent, 3+ completions in window)
	if s := displayStatus(d, habit, entries); s != "satisfied" {
		t.Errorf("Display should be 'satisfied' (y after skip restores), got '%s'", s)
	}
}

// TestSkipifiedWithWarning tests that Warning fires correctly after skip grace expires.
func TestSkipifiedWithWarningAfterGraceExpires(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}

	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 1}, Habit: "Dev"}: {Result: "s"},
	}

	// Day 7 (last day of grace): skipified
	if !graph.Skipified(civil.Date{Year: 2026, Month: 3, Day: 7}, habit, entries) {
		t.Error("Day 7 should be skipified (last day of grace)")
	}

	// Day 8 (grace expired): should trigger Warning since no completions in window
	if !graph.Warning(civil.Date{Year: 2026, Month: 3, Day: 8}, habit, entries) {
		t.Error("Day 8 should be Warning (grace expired, no completions)")
	}
}

// TestSkipifiedRealWorldDevScenario replicates the user's actual Dev habit data around
// the March 10 and March 17 skips to ensure correct behavior.
func TestSkipifiedRealWorldDevScenario(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2025, Month: 12, Day: 31},
	}

	entries := storage.Entries{
		// Recent history leading into the skips
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 9}, Habit: "Dev"}:  {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Dev"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 11}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 12}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 13}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 14}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 15}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 16}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 17}, Habit: "Dev"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 18}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 19}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 20}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 21}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 22}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 23}, Habit: "Dev"}: {Result: "n"},
	}

	// Skip on day 10 → skipified through day 16
	for day := 11; day <= 16; day++ {
		d := civil.Date{Year: 2026, Month: 3, Day: day}
		if !graph.Skipified(d, habit, entries) {
			t.Errorf("Day %d should be skipified (within 7-day grace of skip on March 10)", day)
		}
	}

	// Skip on day 17 → skipified through day 23
	for day := 18; day <= 23; day++ {
		d := civil.Date{Year: 2026, Month: 3, Day: day}
		if !graph.Skipified(d, habit, entries) {
			t.Errorf("Day %d should be skipified (within 7-day grace of skip on March 17)", day)
		}
	}

	// Day 24 should NOT be skipified (outside grace of both skips)
	if graph.Skipified(civil.Date{Year: 2026, Month: 3, Day: 24}, habit, entries) {
		t.Error("Day 24 should NOT be skipified (outside all skip grace periods)")
	}
}

// TestSkipifiedTrackingOnlyHabit verifies tracking-only habits (target 0) are never skipified.
func TestSkipifiedTrackingOnlyHabit(t *testing.T) {
	habit := &storage.Habit{
		Name: "Track", Target: 0, Interval: 1,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Track"}: {Result: "s"},
	}

	// Target <= 1 && Interval == 1 → early return false
	if graph.Skipified(civil.Date{Year: 2026, Month: 3, Day: 6}, habit, entries) {
		t.Error("Tracking-only habits should never be skipified")
	}
}

// TestSkipifiedOnSkipDayItself verifies the skip day itself counts as within grace.
func TestSkipifiedOnSkipDayItself(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Dev"}: {Result: "s"},
	}

	// The skip day itself should be skipified (BuildGraph shows "•" for "s" entries
	// before reaching Skipified, but the function itself should return true)
	if !graph.Skipified(civil.Date{Year: 2026, Month: 3, Day: 10}, habit, entries) {
		t.Error("Skip day itself should return true from Skipified")
	}
}

// TestSkipifiedBeforeSkipDayFalse verifies days before the skip are not skipified.
func TestSkipifiedBeforeSkipDayFalse(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Dev"}: {Result: "s"},
	}

	// Day 9 (before the skip) should NOT be skipified
	if graph.Skipified(civil.Date{Year: 2026, Month: 3, Day: 9}, habit, entries) {
		t.Error("Day before skip should NOT be skipified")
	}
}

// --- Display priority tests ---
// These test the ACTUAL rendering behavior, simulating the BuildGraph switch.
// BuildGraph checks: SatisfiedByCompletions (y only) -> Skipified -> break.
// This ensures skips show as skipified (not satisfied), but genuine completions
// still show as satisfied even when a skip is nearby.

// displayStatus simulates BuildGraph's switch logic for "n" entries.
// Matches the actual display: SatisfiedByCompletions AND NOT IsInSkipPeriod,
// then Skipified, then break.
func displayStatus(d civil.Date, habit *storage.Habit, entries storage.Entries) string {
	switch {
	case graph.SatisfiedByCompletions(d, habit, entries) && !graph.IsInSkipPeriod(d, habit, entries):
		return "satisfied"
	case graph.Skipified(d, habit, entries):
		return "skipified"
	default:
		return "break"
	}
}

// TestDisplayPriority_SkipOnlyTarget_1of9 tests that for a 1/9 habit,
// after a skip, "n" days show skipified not satisfied. The y on day 9
// genuinely satisfies through day 17 (9-day window), but days 18+ where
// only the skip keeps the window alive should show skipified.
func TestDisplayPriority_SkipOnlyTarget_1of9(t *testing.T) {
	habit := &storage.Habit{
		Name: "Call", Target: 1, Interval: 9,
		FirstRecord: civil.Date{Year: 2026, Month: 1, Day: 1},
	}

	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 9}, Habit: "Call"}:  {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 11}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 12}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 13}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 14}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 15}, Habit: "Call"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 16}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 17}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 18}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 19}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 20}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 21}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 22}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 23}, Habit: "Call"}: {Result: "n"},
	}

	// Days 10-14: y on day 9 genuinely satisfies, no skip yet
	for day := 10; day <= 14; day++ {
		d := civil.Date{Year: 2026, Month: 3, Day: day}
		status := displayStatus(d, habit, entries)
		if status != "satisfied" {
			t.Errorf("Day %d: should be 'satisfied' (y on day 9 in window, before skip), got '%s'", day, status)
		}
	}

	// Days 16-23: skip on day 15 resets display to skipified.
	// Even though y on day 9 is still in the window for days 16-17,
	// the skip intervened so skipified takes over until a new y appears.
	for day := 16; day <= 23; day++ {
		d := civil.Date{Year: 2026, Month: 3, Day: day}
		status := displayStatus(d, habit, entries)
		if status != "skipified" {
			t.Errorf("Day %d: should be 'skipified' (skip on 15 resets display), got '%s'", day, status)
		}
	}

	// Day 24: outside skip grace (9 days from skip on 15)
	if graph.Skipified(civil.Date{Year: 2026, Month: 3, Day: 24}, habit, entries) {
		t.Error("Day 24 should NOT be skipified (outside 9-day grace)")
	}
}

// TestDisplayPriority_SkipOnlyTarget_1of8 mirrors Review habit (1/8).
func TestDisplayPriority_SkipOnlyTarget_1of8(t *testing.T) {
	habit := &storage.Habit{
		Name: "Review", Target: 1, Interval: 8,
		FirstRecord: civil.Date{Year: 2026, Month: 1, Day: 1},
	}

	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 8}, Habit: "Review"}:  {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 9}, Habit: "Review"}:  {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 15}, Habit: "Review"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 16}, Habit: "Review"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 20}, Habit: "Review"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 22}, Habit: "Review"}: {Result: "n"},
	}

	// Day 9: y on day 8 is in the 8-day window -> satisfied
	if s := displayStatus(civil.Date{Year: 2026, Month: 3, Day: 9}, habit, entries); s != "satisfied" {
		t.Errorf("Day 9: should be 'satisfied' (y on day 8 in window), got '%s'", s)
	}

	// Day 16: y on day 8 is 8 days ago, outside window. Skip on 15 -> skipified
	if s := displayStatus(civil.Date{Year: 2026, Month: 3, Day: 16}, habit, entries); s != "skipified" {
		t.Errorf("Day 16: should be 'skipified', got '%s'", s)
	}

	// Day 22 (7 days after skip): still within interval-1=7 days of grace
	if s := displayStatus(civil.Date{Year: 2026, Month: 3, Day: 22}, habit, entries); s != "skipified" {
		t.Errorf("Day 22: should be 'skipified' (within 8-day grace of skip on 15), got '%s'", s)
	}

	// Day 23 (8 days after skip): outside grace
	if graph.Skipified(civil.Date{Year: 2026, Month: 3, Day: 23}, habit, entries) {
		t.Error("Day 23 should NOT be Skipified (outside 8-day grace)")
	}
}

// TestDisplayPriority_SatisfiedStillWorksWithoutSkip confirms that when there's
// no skip in the window, SatisfiedByCompletions renders correctly.
func TestDisplayPriority_SatisfiedStillWorksWithoutSkip(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}

	// 3 completions in the window, no skips anywhere
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 1}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 3}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 6}, Habit: "Dev"}: {Result: "n"},
	}

	d := civil.Date{Year: 2026, Month: 3, Day: 6}
	if s := displayStatus(d, habit, entries); s != "satisfied" {
		t.Errorf("Display should be 'satisfied' (3 y's in window, no skip), got '%s'", s)
	}
}

// TestDisplayPriority_SkipThenRestart_1of7 tests the critical scenario:
// s n y n n -- the "n" after the "y" restart should show satisfied, not skipified,
// because the y genuinely satisfies the window.
func TestDisplayPriority_SkipThenRestart_1of7(t *testing.T) {
	habit := &storage.Habit{
		Name: "Weekly", Target: 1, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}

	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 1}, Habit: "Weekly"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 2}, Habit: "Weekly"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 3}, Habit: "Weekly"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 4}, Habit: "Weekly"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Weekly"}: {Result: "n"},
	}

	// Day 2: no y in window (only s on day 1), SatisfiedByCompletions=false, Skipified=true
	if s := displayStatus(civil.Date{Year: 2026, Month: 3, Day: 2}, habit, entries); s != "skipified" {
		t.Errorf("Day 2: should be 'skipified' (no y, skip grace), got '%s'", s)
	}

	// Day 4: y on day 3 satisfies the window genuinely
	if s := displayStatus(civil.Date{Year: 2026, Month: 3, Day: 4}, habit, entries); s != "satisfied" {
		t.Errorf("Day 4: should be 'satisfied' (y on day 3 in window), got '%s'", s)
	}

	// Day 5: still satisfied by y on day 3 (within 7-day window)
	if s := displayStatus(civil.Date{Year: 2026, Month: 3, Day: 5}, habit, entries); s != "satisfied" {
		t.Errorf("Day 5: should be 'satisfied' (y on day 3 in window), got '%s'", s)
	}
}

// TestDisplayPriority_SkipThenRestart_3of7 tests the same for multi-target.
// For 3/7: s n y y y n -- the final n has 3 genuine completions in window.
func TestDisplayPriority_SkipThenRestart_3of7(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}

	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 1}, Habit: "Dev"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 2}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 3}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 4}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 6}, Habit: "Dev"}: {Result: "n"},
	}

	// Day 2: only 1 skip, no y's yet -> skipified
	if s := displayStatus(civil.Date{Year: 2026, Month: 3, Day: 2}, habit, entries); s != "skipified" {
		t.Errorf("Day 2: should be 'skipified' (no completions yet), got '%s'", s)
	}

	// Day 6: 3 y's in the window -> genuinely satisfied despite skip on day 1
	if s := displayStatus(civil.Date{Year: 2026, Month: 3, Day: 6}, habit, entries); s != "satisfied" {
		t.Errorf("Day 6: should be 'satisfied' (3 y's in window), got '%s'", s)
	}
}

// TestDisplayPriority_MixedSkipAndCompletions_3of7 tests 2y + 1s for a 3/7 habit.
// Target=3 but only 2 genuine completions -> should show skipified, not satisfied.
func TestDisplayPriority_MixedSkipAndCompletions_3of7(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}

	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 1}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 3}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Dev"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 6}, Habit: "Dev"}: {Result: "n"},
	}

	d := civil.Date{Year: 2026, Month: 3, Day: 6}

	// SatisfiedByCompletions: only 2 y's in window, target=3 -> false
	if graph.SatisfiedByCompletions(d, habit, entries) {
		t.Error("SatisfiedByCompletions should be false (2 y's < target 3)")
	}

	// Satisfied (with skips): 2 y's + 1 s = 3 -> true
	if !graph.Satisfied(d, habit, entries) {
		t.Error("Satisfied (with skips) should be true (2y + 1s = 3)")
	}

	// Display: skipified (target not met by completions, but skip grace active)
	if s := displayStatus(d, habit, entries); s != "skipified" {
		t.Errorf("Day 6: should be 'skipified' (2y + 1s, not genuinely satisfied), got '%s'", s)
	}
}

// TestSatisfiedByCompletions_IgnoresSkips directly tests the new function.
func TestSatisfiedByCompletions_IgnoresSkips(t *testing.T) {
	habit := &storage.Habit{
		Name: "Weekly", Target: 1, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}

	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Weekly"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 7}, Habit: "Weekly"}: {Result: "n"},
	}

	d := civil.Date{Year: 2026, Month: 3, Day: 7}

	// Satisfied counts the skip -> true
	if !graph.Satisfied(d, habit, entries) {
		t.Error("Satisfied should be true (skip counts)")
	}

	// SatisfiedByCompletions ignores the skip -> false
	if graph.SatisfiedByCompletions(d, habit, entries) {
		t.Error("SatisfiedByCompletions should be false (skip ignored, no y in window)")
	}
}

// TestSatisfiedByCompletions_YesEntryStillWorks confirms y entries work normally.
func TestSatisfiedByCompletions_YesEntryStillWorks(t *testing.T) {
	habit := &storage.Habit{
		Name: "Weekly", Target: 1, Interval: 7,
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}

	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Weekly"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 7}, Habit: "Weekly"}: {Result: "n"},
	}

	d := civil.Date{Year: 2026, Month: 3, Day: 7}

	if !graph.SatisfiedByCompletions(d, habit, entries) {
		t.Error("SatisfiedByCompletions should be true (y in window)")
	}
}

// TestDisplayPriority_SkipIntervenesPreSkipCompletion_1of9 is the exact
// real-world regression for "Call Mom & Dad" (1/9). A y on 3/9 satisfies
// the window through 3/17, but a skip on 3/15 intervenes. Days 3/16-3/17
// must show skipified, not satisfied, because the skip resets the display.
func TestDisplayPriority_SkipIntervenesPreSkipCompletion_1of9(t *testing.T) {
	habit := &storage.Habit{
		Name: "Call Mom & Dad", Target: 1, Interval: 9,
		FirstRecord: civil.Date{Year: 2025, Month: 12, Day: 31},
	}

	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 9}, Habit: "Call Mom & Dad"}:  {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Call Mom & Dad"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 11}, Habit: "Call Mom & Dad"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 12}, Habit: "Call Mom & Dad"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 13}, Habit: "Call Mom & Dad"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 14}, Habit: "Call Mom & Dad"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 15}, Habit: "Call Mom & Dad"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 16}, Habit: "Call Mom & Dad"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 17}, Habit: "Call Mom & Dad"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 18}, Habit: "Call Mom & Dad"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 19}, Habit: "Call Mom & Dad"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 20}, Habit: "Call Mom & Dad"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 21}, Habit: "Call Mom & Dad"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 22}, Habit: "Call Mom & Dad"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 23}, Habit: "Call Mom & Dad"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 24}, Habit: "Call Mom & Dad"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 25}, Habit: "Call Mom & Dad"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 26}, Habit: "Call Mom & Dad"}: {Result: "n"},
	}

	// Expected graph: ────•········━━─
	// 3/10-3/14: satisfied (y on 3/9 in window, before skip)
	for day := 10; day <= 14; day++ {
		d := civil.Date{Year: 2026, Month: 3, Day: day}
		if s := displayStatus(d, habit, entries); s != "satisfied" {
			t.Errorf("Day %d: should be 'satisfied' (y on 3/9 in window, before skip), got '%s'", day, s)
		}
	}

	// 3/16-3/17: skip on 3/15 intervenes, even though y on 3/9 is still in window
	for day := 16; day <= 17; day++ {
		d := civil.Date{Year: 2026, Month: 3, Day: day}
		if s := displayStatus(d, habit, entries); s != "skipified" {
			t.Errorf("Day %d: should be 'skipified' (skip on 3/15 intervenes), got '%s'", day, s)
		}
	}

	// 3/18-3/23: skip grace continues (y on 3/9 has aged out anyway)
	for day := 18; day <= 23; day++ {
		d := civil.Date{Year: 2026, Month: 3, Day: day}
		if s := displayStatus(d, habit, entries); s != "skipified" {
			t.Errorf("Day %d: should be 'skipified' (skip grace from 3/15), got '%s'", day, s)
		}
	}

	// 3/26: y on 3/24 is the most recent y/s (not the skip), so satisfied
	d := civil.Date{Year: 2026, Month: 3, Day: 26}
	if s := displayStatus(d, habit, entries); s != "satisfied" {
		t.Errorf("Day 26: should be 'satisfied' (y on 3/24-25 in window, after skip period), got '%s'", s)
	}
}

// TestDisplayPriority_SkipThenYThenN_SatisfiedAfterRestart verifies that
// after a skip, once a new y appears, subsequent n's show satisfied again.
// Pattern: y...s n n y n n -> satisfied, skip, skipified, skipified, done, satisfied, satisfied
func TestDisplayPriority_SkipThenYThenN_SatisfiedAfterRestart(t *testing.T) {
	habit := &storage.Habit{
		Name: "Review", Target: 1, Interval: 8,
		FirstRecord: civil.Date{Year: 2026, Month: 1, Day: 1},
	}

	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 1}, Habit: "Review"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Review"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 6}, Habit: "Review"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 7}, Habit: "Review"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 8}, Habit: "Review"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 9}, Habit: "Review"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Review"}: {Result: "n"},
	}

	// Day 6: after skip on 5, in skip period -> skipified
	if s := displayStatus(civil.Date{Year: 2026, Month: 3, Day: 6}, habit, entries); s != "skipified" {
		t.Errorf("Day 6: should be 'skipified' (after skip on 5), got '%s'", s)
	}

	// Day 7: still in skip period -> skipified
	if s := displayStatus(civil.Date{Year: 2026, Month: 3, Day: 7}, habit, entries); s != "skipified" {
		t.Errorf("Day 7: should be 'skipified' (still in skip period), got '%s'", s)
	}

	// Day 9: y on day 8 is the most recent y/s -> not in skip period -> satisfied
	if s := displayStatus(civil.Date{Year: 2026, Month: 3, Day: 9}, habit, entries); s != "satisfied" {
		t.Errorf("Day 9: should be 'satisfied' (y on day 8 restarts satisfied), got '%s'", s)
	}

	// Day 10: y on day 8 still most recent -> satisfied
	if s := displayStatus(civil.Date{Year: 2026, Month: 3, Day: 10}, habit, entries); s != "satisfied" {
		t.Errorf("Day 10: should be 'satisfied' (y on day 8 in window), got '%s'", s)
	}
}

// --- Stats tests: verify BuildStats classifies days the same as display ---

// TestBuildStats_SkipIntervenesPreSkipCompletion verifies stats count days
// after a skip as skips (not streaks), matching the display classification.
func TestBuildStats_SkipIntervenesPreSkipCompletion(t *testing.T) {
	habit := &storage.Habit{
		Name: "Call", Target: 1, Interval: 9,
		Frequency: "1/9",
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 9},
	}

	// y(9) n(10-14) s(15) n(16-20)
	entries := &storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 9}, Habit: "Call"}:  {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 11}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 12}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 13}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 14}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 15}, Habit: "Call"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 16}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 17}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 18}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 19}, Habit: "Call"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 20}, Habit: "Call"}: {Result: "n"},
	}

	stats := ui.BuildStats(habit, entries)

	// Day 9: y -> streak
	// Days 10-14: SatisfiedByCompletions (y on 9 in window) AND not in skip period -> streak (5)
	// Day 15: s -> skip (1)
	// Days 16-17: SatisfiedByCompletions true (y on 9 still in window) BUT IsInSkipPeriod -> skipified -> skip (2)
	// Days 18-20: y aged out, Skipified true (skip on 15 in grace) -> skip (3)
	// Total: streaks=6 (1y + 5 satisfied), skips=6 (1s + 2 skipified-in-skip-period + 3 skipified), breaks=0

	if stats.Streaks != 6 {
		t.Errorf("Expected 6 streak days (1 y + 5 satisfied before skip), got %d", stats.Streaks)
	}
	if stats.Skips != 6 {
		t.Errorf("Expected 6 skip days (1 s + 5 skipified after skip), got %d", stats.Skips)
	}
	if stats.Breaks != 0 {
		t.Errorf("Expected 0 break days, got %d", stats.Breaks)
	}
}

// TestBuildStats_SkipThenRestart verifies stats correctly transition from
// skip to streak when a new y appears after a skip.
func TestBuildStats_SkipThenRestart(t *testing.T) {
	habit := &storage.Habit{
		Name: "Review", Target: 1, Interval: 8,
		Frequency: "1/8",
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 5},
	}

	// s(5) n(6) n(7) y(8) n(9) n(10)
	entries := &storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Review"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 6}, Habit: "Review"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 7}, Habit: "Review"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 8}, Habit: "Review"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 9}, Habit: "Review"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Review"}: {Result: "n"},
	}

	stats := ui.BuildStats(habit, entries)

	// Day 5: s -> skip
	// Days 6-7: in skip period, skipified -> skip (2)
	// Day 8: y -> streak
	// Days 9-10: satisfied (y on 8 in window), not in skip period -> streak (2)
	// Total: streaks=3, skips=3, breaks=0

	if stats.Streaks != 3 {
		t.Errorf("Expected 3 streak days (1 y + 2 satisfied after restart), got %d", stats.Streaks)
	}
	if stats.Skips != 3 {
		t.Errorf("Expected 3 skip days (1 s + 2 skipified), got %d", stats.Skips)
	}
	if stats.Breaks != 0 {
		t.Errorf("Expected 0 break days, got %d", stats.Breaks)
	}
}

// TestBuildStats_ConsistentWithDisplay verifies that the total of streaks + skips + breaks
// equals the number of logged days (every logged day must be classified).
func TestBuildStats_ConsistentWithDisplay(t *testing.T) {
	habit := &storage.Habit{
		Name: "Dev", Target: 3, Interval: 7,
		Frequency: "3/7",
		FirstRecord: civil.Date{Year: 2026, Month: 3, Day: 1},
	}

	entries := &storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 1}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 2}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 3}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 4}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 5}, Habit: "Dev"}: {Result: "s"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 6}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 7}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 8}, Habit: "Dev"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 9}, Habit: "Dev"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2026, Month: 3, Day: 10}, Habit: "Dev"}: {Result: "n"},
	}

	stats := ui.BuildStats(habit, entries)

	loggedDays := 10 // all 10 days have entries
	total := stats.Streaks + stats.Skips + stats.Breaks
	if total != loggedDays {
		t.Errorf("Stats total (%d streaks + %d skips + %d breaks = %d) should equal logged days (%d)",
			stats.Streaks, stats.Skips, stats.Breaks, total, loggedDays)
	}
}
