package graph

import (
	"math"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"github.com/gookit/color"
	"github.com/wakatara/harsh/internal/storage"
)

// BuildGraph creates a consistency graph for a single habit
func BuildGraph(habit *storage.Habit, entries *storage.Entries, countBack int, ask bool) string {
	graphLen := countBack
	if ask {
		graphLen = max(1, graphLen-12)
	}
	var graphDay string
	var consistency strings.Builder

	to := civil.DateOf(time.Now())
	from := to.AddDays(-graphLen)
	consistency.Grow(graphLen)

	for d := from; !d.After(to); d = d.AddDays(1) {
		// After the end date, show blank (habit has been retired)
		if habit.HasEnded(d) {
			// Show muted end marker on the first day after the end date
			if !habit.EndRecord.IsZero() && d == habit.EndRecord.AddDays(1) {
				graphDay = color.C256(245).Sprint("▏")
			} else {
				graphDay = " "
			}
		} else if outcome, ok := (*entries)[storage.DailyHabit{Day: d, Habit: habit.Name}]; ok {
			switch {
			case outcome.Result == "y":
				graphDay = "━"
			case outcome.Result == "s":
				graphDay = "•"
			// look at cases of "n" being entered but
			// within bounds of the habit every x days
			case Satisfied(d, habit, *entries):
				graphDay = "─"
			case Skipified(d, habit, *entries):
				graphDay = "·"
			case outcome.Result == "n":
				graphDay = " "
			}
		} else {
			if Warning(d, habit, *entries) && (to.DaysSince(d) < 14) {
				// warning: sigils max out at 2 weeks (~90 day habit in formula)
				graphDay = "!"
			} else if d.After(habit.FirstRecord) {
				// For people who miss days but then put in later ones
				graphDay = "◌"
			} else {
				graphDay = " "
			}
		}
		consistency.WriteString(graphDay)
	}

	return consistency.String()
}

// Satisfied checks if a habit target is satisfied within its interval window
func Satisfied(d civil.Date, habit *storage.Habit, entries storage.Entries) bool {
	if habit.Target <= 1 && habit.Interval == 1 {
		return false
	}

	// For date d, check all possible interval-length windows that include d
	// Key insight: Allow future data only if there's supporting data at or before d in the same window

	earliestStart := d.AddDays(-habit.Interval + 1)
	if earliestStart.Before(habit.FirstRecord) {
		earliestStart = habit.FirstRecord
	}

	latestStart := d

	// Try all possible window start positions that would include date d
	for winStart := earliestStart; !winStart.After(latestStart); winStart = winStart.AddDays(1) {
		winEnd := winStart.AddDays(habit.Interval - 1)

		// The window must include the date d we're checking
		if winEnd.Before(d) || winStart.After(d) {
			continue
		}

		// Count successes in the entire window (including potential future data)
		// Skips ("s") count as success - they maintain the streak
		countTotal := 0
		countUpToD := 0

		// Early termination: stop counting once we exceed target
		for dt := winStart; !dt.After(winEnd) && countTotal < habit.Target+1; dt = dt.AddDays(1) {
			if v, ok := entries[storage.DailyHabit{Day: dt, Habit: habit.Name}]; ok && (v.Result == "y" || v.Result == "s") {
				countTotal++
				if !dt.After(d) {
					countUpToD++
				}
				// Early exit optimization for Target=1 special case
				if habit.Target == 1 && countTotal == 2 {
					// But only if we have supporting data up to d
					if countUpToD > 0 {
						return true
					}
				}
			}
		}

		// Check if target is met for this window
		if countTotal >= habit.Target {
			// Allow future data only if there's at least some supporting data up to date d
			// This prevents the original bug where dates before any success were marked satisfied
			if countUpToD > 0 {
				return true
			}
		}
	}

	return false
}

// Skipified checks if a habit has been skipped within its grace period
func Skipified(d civil.Date, habit *storage.Habit, entries storage.Entries) bool {
	if habit.Target <= 1 && habit.Interval == 1 {
		return false
	}

	from := d
	to := d.AddDays(-int(math.Ceil(float64(habit.Interval) / float64(habit.Target))))
	for dt := from; !dt.Before(to); dt = dt.AddDays(-1) {
		if v, ok := entries[storage.DailyHabit{Day: dt, Habit: habit.Name}]; ok {
			if v.Result == "s" {
				return true
			}
		}
	}
	return false
}

// Warning checks if a habit should show a warning indicator
func Warning(d civil.Date, habit *storage.Habit, entries storage.Entries) bool {
	if habit.Target < 1 {
		return false
	}

	warningDays := int(habit.Interval)/7 + 1
	to := d
	from := d.AddDays(-int(habit.Interval) + warningDays)
	noFirstRecord := civil.Date{Year: 0, Month: 0, Day: 0}
	for dt := from; !dt.After(to); dt = dt.AddDays(1) {
		if v, ok := entries[storage.DailyHabit{Day: dt, Habit: habit.Name}]; ok {
			switch v.Result {
			case "y":
				return false
			case "s":
				return false
			}
		}
		// Edge case for 0 day onboard and later completes null entry habits
		if habit.FirstRecord == noFirstRecord {
			return false
		}
		if dt.Before(habit.FirstRecord) {
			return false
		}
	}
	return true
}

// DaysUntilStreakBreak calculates how many days until a habit's streak will break
// Returns -1 if the habit doesn't have a streak or is a tracking-only habit (target 0)
func DaysUntilStreakBreak(d civil.Date, habit *storage.Habit, entries storage.Entries) int {
	// Tracking-only habits (target 0) don't have streaks
	if habit.Target < 1 {
		return -1
	}

	noFirstRecord := civil.Date{Year: 0, Month: 0, Day: 0}
	// Edge case for habits not yet started
	if habit.FirstRecord == noFirstRecord || d.Before(habit.FirstRecord) {
		return -1
	}

	// For habits with Target=1 (e.g., 1/1, 1/7, 1/90), use the simpler direct approach
	// Only use windowing for multi-target habits (e.g., 3/7, 2/14)
	if habit.Target == 1 {
		return daysUntilStreakBreakSimple(d, habit, entries)
	}

	// For interval habits with Target > 1 (e.g., 3/7), use sliding window approach
	return daysUntilStreakBreakWindowed(d, habit, entries)
}

// daysUntilStreakBreakSimple handles Target=1 habits (1/1, 1/7, etc.)
func daysUntilStreakBreakSimple(d civil.Date, habit *storage.Habit, entries storage.Entries) int {
	// Look back to find the last success
	maxLookback := max(habit.Interval*3, 365)
	lookbackStart := d.AddDays(-maxLookback)
	if lookbackStart.Before(habit.FirstRecord) {
		lookbackStart = habit.FirstRecord
	}

	lastSuccessDate := civil.Date{Year: 0, Month: 0, Day: 0}
	lastSuccessResult := ""
	for dt := d; !dt.Before(lookbackStart); dt = dt.AddDays(-1) {
		if v, ok := entries[storage.DailyHabit{Day: dt, Habit: habit.Name}]; ok {
			if v.Result == "y" || v.Result == "s" {
				lastSuccessDate = dt
				lastSuccessResult = v.Result
				break
			}
		}
	}

	// If no success found, streak is broken
	if lastSuccessDate.Year == 0 {
		return -999
	}

	// A "y" (completion) always starts or maintains a streak.
	// A "s" (skip) only maintains an existing streak — it cannot restart
	// a broken one. Verify the streak was intact when the skip was logged.
	if lastSuccessResult == "s" {
		priorLookbackStart := lastSuccessDate.AddDays(-habit.Interval * 2)
		if priorLookbackStart.Before(habit.FirstRecord) {
			priorLookbackStart = habit.FirstRecord
		}

		priorSuccessDate := civil.Date{Year: 0, Month: 0, Day: 0}
		for dt := lastSuccessDate.AddDays(-1); !dt.Before(priorLookbackStart); dt = dt.AddDays(-1) {
			if v, ok := entries[storage.DailyHabit{Day: dt, Habit: habit.Name}]; ok {
				if v.Result == "y" || v.Result == "s" {
					priorSuccessDate = dt
					break
				}
			}
		}

		// If there's a prior success and the gap exceeds interval, the skip came after a break.
		// No prior success means the skip is the first entry — no streak existed to break.
		if priorSuccessDate.Year != 0 && lastSuccessDate.DaysSince(priorSuccessDate) > habit.Interval {
			return -999
		}
	}

	// For Target=1 habits: last success + interval = break date
	streakBreakDate := lastSuccessDate.AddDays(habit.Interval)
	return streakBreakDate.DaysSince(d)
}

// daysUntilStreakBreakWindowed handles interval habits (e.g., 3/7) using sliding windows
func daysUntilStreakBreakWindowed(d civil.Date, habit *storage.Habit, entries storage.Entries) int {
	// First, check if today is satisfied
	todaySatisfied := Satisfied(d, habit, entries)

	if !todaySatisfied {
		// Check if it was satisfied yesterday to determine if streak broke today or earlier
		yesterday := d.AddDays(-1)
		if !yesterday.Before(habit.FirstRecord) && Satisfied(yesterday, habit, entries) {
			// Was satisfied yesterday but not today - breaks today
			return 0
		}
		// Already broken before today
		return -999
	}

	// Today is satisfied. Find the earliest success in any window that keeps today satisfied
	// This is the critical success - when it ages out, the streak breaks
	earliestCriticalSuccess := civil.Date{Year: 0, Month: 0, Day: 0}

	// Check all windows that could contain today and have target successes
	earliestStart := d.AddDays(-habit.Interval + 1)
	if earliestStart.Before(habit.FirstRecord) {
		earliestStart = habit.FirstRecord
	}

	for winStart := earliestStart; !winStart.After(d); winStart = winStart.AddDays(1) {
		winEnd := winStart.AddDays(habit.Interval - 1)

		// Skip if window doesn't contain today
		if winEnd.Before(d) {
			continue
		}

		// Count successes in this window (only up to today)
		successCount := 0
		var firstSuccessInWindow civil.Date

		for dt := winStart; !dt.After(d) && !dt.After(winEnd); dt = dt.AddDays(1) {
			if v, ok := entries[storage.DailyHabit{Day: dt, Habit: habit.Name}]; ok && (v.Result == "y" || v.Result == "s") {
				successCount++
				if firstSuccessInWindow.Year == 0 {
					firstSuccessInWindow = dt
				}
			}
		}

		// If this window satisfies the target, track its earliest success
		if successCount >= habit.Target && firstSuccessInWindow.Year != 0 {
			if earliestCriticalSuccess.Year == 0 || firstSuccessInWindow.Before(earliestCriticalSuccess) {
				earliestCriticalSuccess = firstSuccessInWindow
			}
		}
	}

	// If no critical success found (shouldn't happen if Satisfied returned true)
	if earliestCriticalSuccess.Year == 0 {
		return -999
	}

	// Streak breaks when the earliest critical success ages out
	streakBreakDate := earliestCriticalSuccess.AddDays(habit.Interval)
	return streakBreakDate.DaysSince(d)
}

// IsInSkipPeriod checks if a habit's most recent entry (within the interval) was a skip
// This is used to show a distinct indicator for habits in a skip grace period
func IsInSkipPeriod(d civil.Date, habit *storage.Habit, entries storage.Entries) bool {
	// Tracking-only habits don't have skip periods
	if habit.Target < 1 {
		return false
	}

	noFirstRecord := civil.Date{Year: 0, Month: 0, Day: 0}
	if habit.FirstRecord == noFirstRecord || d.Before(habit.FirstRecord) {
		return false
	}

	// Look back within the interval to find the most recent entry
	maxLookback := max(habit.Interval*2, 14)
	lookbackStart := d.AddDays(-maxLookback)
	if lookbackStart.Before(habit.FirstRecord) {
		lookbackStart = habit.FirstRecord
	}

	// Find the most recent "y" or "s" entry
	for dt := d; !dt.Before(lookbackStart); dt = dt.AddDays(-1) {
		if v, ok := entries[storage.DailyHabit{Day: dt, Habit: habit.Name}]; ok {
			if v.Result == "s" {
				return true
			}
			if v.Result == "y" {
				return false
			}
			// "n" doesn't count as maintaining streak, keep looking
		}
	}

	return false
}
