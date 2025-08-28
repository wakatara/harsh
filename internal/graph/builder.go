package graph

import (
	"math"
	"strings"
	"time"

	"cloud.google.com/go/civil"
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
		if outcome, ok := (*entries)[storage.DailyHabit{Day: d, Habit: habit.Name}]; ok {
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
		countTotal := 0
		countUpToD := 0
		
		// Early termination: stop counting once we exceed target
		for dt := winStart; !dt.After(winEnd) && countTotal < habit.Target+1; dt = dt.AddDays(1) {
			if v, ok := entries[storage.DailyHabit{Day: dt, Habit: habit.Name}]; ok && v.Result == "y" {
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
