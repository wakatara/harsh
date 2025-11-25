package graph

import (
	"math"
	"time"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal/storage"
)

// BuildSpark creates sparkline and calendar line for visualization
func BuildSpark(from civil.Date, to civil.Date, habits []*storage.Habit, entries *storage.Entries) ([]string, []string) {
	sparkline := []string{}
	calline := []string{}
	sparks := []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	i := 0
	LetterDay := map[string]string{
		"Sunday": " ", "Monday": "M", "Tuesday": " ", "Wednesday": "W",
		"Thursday": " ", "Friday": "F", "Saturday": " ",
	}

	var prevMonth time.Month
	var prevWeekday string
	isFirstDay := true

	for d := from; !d.After(to); d = d.AddDays(1) {
		dailyScore := Score(d, habits, entries)
		// divide score into score to map to sparks slice graphic for sparkline
		if dailyScore == 100 {
			i = 8
		} else {
			i = int(math.Ceil(dailyScore / float64(100/(len(sparks)-1))))
		}
		t, _ := time.Parse(time.DateOnly, d.String())
		w := t.Weekday().String()
		currentMonth := t.Month()

		// Check if this is a month boundary (1st of the month, but not the very first day)
		isMonthBoundary := !isFirstDay && d.Day == 1 && currentMonth != prevMonth

		// Determine which marker to use based on the previous day's weekday
		if isMonthBoundary && len(calline) > 0 {
			// If previous day was Monday, Wednesday, or Friday, use left-aligned marker on current day
			if prevWeekday == "Monday" || prevWeekday == "Wednesday" || prevWeekday == "Friday" {
				// Left marker with current day's letter (e.g., "⎸ " or "⎸M")
				calline = append(calline, "⎸"+LetterDay[w])
			} else {
				// Right-aligned marker on previous day, preserving the day letter
				calline[len(calline)-1] = calline[len(calline)-1] + "⎹"
				calline = append(calline, LetterDay[w])
			}
		} else {
			// Normal day - just add the letter
			calline = append(calline, LetterDay[w])
		}

		sparkline = append(sparkline, sparks[i])

		prevMonth = currentMonth
		prevWeekday = w
		isFirstDay = false
	}

	return sparkline, calline
}

// Score calculates the daily score for a given date
func Score(d civil.Date, habits []*storage.Habit, entries *storage.Entries) float64 {
	scored := 0.0
	skipped := 0.0
	scorableHabits := 0.0

	for _, habit := range habits {
		if habit.Target > 0 && !d.Before(habit.FirstRecord) {
			scorableHabits++
			if outcome, ok := (*entries)[storage.DailyHabit{Day: d, Habit: habit.Name}]; ok {
				switch {
				case outcome.Result == "y":
					scored++
				case outcome.Result == "s":
					skipped++
				// look at cases of n being entered but
				// within bounds of the habit every x days
				case Satisfied(d, habit, *entries):
					scored++
				case Skipified(d, habit, *entries):
					skipped++
				}
			}
		}
	}

	var score float64
	// Edge case on if there is nothing to score and the scorable vs skipped issue
	if scorableHabits == 0 {
		score = 0.0
	} else {
		score = 100.0 // deal with scorable habits - skipped == 0 causing divide by zero issue
	}
	if scorableHabits-skipped != 0 {
		score = (scored / (scorableHabits - skipped)) * 100
	}
	return score
}
