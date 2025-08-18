package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal/graph"
	"github.com/wakatara/harsh/internal/storage"
)

// HabitStats holds total stats for a Habit in the file
type HabitStats struct {
	DaysTracked int
	Total       float64
	Streaks     int
	Breaks      int
	Skips       int
}

// Display handles the formatting and output of habit information
type Display struct {
	colorManager *ColorManager
}

// NewDisplay creates a new display handler
func NewDisplay(noColor bool) *Display {
	return &Display{
		colorManager: NewColorManager(noColor),
	}
}

// ShowHabitLog displays the habit log with sparkline and graphs
func (d *Display) ShowHabitLog(habits []*storage.Habit, entries *storage.Entries, countBack int, maxHabitNameLength int, habitFragment string) {
	// Filter habits by fragment if provided
	filteredHabits := []*storage.Habit{}
	if len(strings.TrimSpace(habitFragment)) > 0 {
		for _, habit := range habits {
			if strings.Contains(strings.ToLower(habit.Name), strings.ToLower(habitFragment)) {
				filteredHabits = append(filteredHabits, habit)
			}
		}
	} else {
		filteredHabits = habits
	}

	now := civil.DateOf(time.Now())
	to := now
	from := to.AddDays(-countBack)
	
	// Build sparkline
	sparkline, calline := graph.BuildSpark(from, to, habits, entries)
	fmt.Printf("%*v", maxHabitNameLength, "")
	fmt.Print(strings.Join(sparkline, ""))
	fmt.Printf("\n")
	fmt.Printf("%*v", maxHabitNameLength, "")
	fmt.Print(strings.Join(calline, ""))
	fmt.Printf("\n")

	// Build graphs in parallel
	graphResults := graph.BuildGraphsParallel(filteredHabits, entries, countBack, false)

	heading := ""
	for _, habit := range filteredHabits {
		if heading != habit.Heading {
			d.colorManager.PrintfBold("%s\n", habit.Heading)
			heading = habit.Heading
		}
		fmt.Printf("%*v", maxHabitNameLength, habit.Name+"  ")
		fmt.Print(graphResults[habit.Name])
		fmt.Printf("\n")
	}

	// Show scores and undone count
	undone := GetTodos(habits, entries, now, 7)
	var undoneCount int
	for _, v := range undone {
		undoneCount += len(v)
	}

	yscore := fmt.Sprintf("%.1f", graph.Score(now.AddDays(-1), habits, entries))
	tscore := fmt.Sprintf("%.1f", graph.Score(now, habits, entries))
	fmt.Printf("\n" + "Yesterday's Score: ")
	fmt.Printf("%8v", yscore)
	fmt.Printf("%%\n")
	fmt.Printf("Today's Score: ")
	fmt.Printf("%12v", tscore)
	fmt.Printf("%%\n")
	if undoneCount == 0 {
		fmt.Printf("All habits logged up to today.")
	} else {
		fmt.Printf("Total unlogged habits: ")
		fmt.Printf("%2v", undoneCount)
	}
	fmt.Printf("\n")
}

// ShowHabitStats displays statistics for all habits
func (d *Display) ShowHabitStats(habits []*storage.Habit, entries *storage.Entries, maxHabitNameLength int) {
	heading := ""
	for _, habit := range habits {
		if heading != habit.Heading {
			d.colorManager.PrintfBold("\n%s\n", habit.Heading)
			heading = habit.Heading
		}
		stats := BuildStats(habit, entries)
		fmt.Printf("%*v", maxHabitNameLength, habit.Name+"  ")
		d.colorManager.PrintGreen("Streaks ")
		d.colorManager.PrintfGreen("%4v", strconv.Itoa(stats.Streaks))
		d.colorManager.PrintGreen(" days")
		fmt.Printf("%4v", "")
		d.colorManager.PrintRed("Breaks ")
		d.colorManager.PrintfRed("%4v", strconv.Itoa(stats.Breaks))
		d.colorManager.PrintRed(" days")
		fmt.Printf("%4v", "")
		d.colorManager.PrintYellow("Skips ")
		d.colorManager.PrintfYellow("%4v", strconv.Itoa(stats.Skips))
		d.colorManager.PrintYellow(" days")
		fmt.Printf("%4v", "")
		fmt.Printf("Tracked ")
		fmt.Printf("%4v", strconv.Itoa(stats.DaysTracked))
		fmt.Printf(" days")
		if stats.Total == 0 {
			fmt.Printf("%4v", "")
			fmt.Printf("      ")
			fmt.Printf("%5v", "")
			fmt.Printf("     \n")
		} else {
			fmt.Printf("%4v", "")
			d.colorManager.PrintBlue("Total ")
			d.colorManager.PrintfBlue("%5v", (stats.Total))
			d.colorManager.PrintBlue("     \n")
		}
	}
}

// ShowTodos displays undone habits for today and recent days
func (d *Display) ShowTodos(habits []*storage.Habit, entries *storage.Entries, maxHabitNameLength int) {
	now := civil.DateOf(time.Now())
	undone := GetTodos(habits, entries, now, 8)

	heading := ""
	if len(undone) == 0 {
		fmt.Println("All todos logged up to today.")
	} else {
		for date, todos := range undone {
			t, _ := time.Parse(time.DateOnly, date)
			dayOfWeek := t.Weekday().String()[:3]
			d.colorManager.PrintlnBold(date + " " + dayOfWeek + ":")
			for _, habit := range habits {
				for _, todo := range todos {
					if heading != habit.Heading && habit.Heading == todo {
						d.colorManager.PrintfBold("\n%s\n", habit.Heading)
						heading = habit.Heading
					}
					if habit.Name == todo {
						fmt.Printf("%*v", maxHabitNameLength, todo+"\n")
					}
				}
			}
		}
	}
}

// GetTodos returns a map of date strings to habit names that are undone
func GetTodos(habits []*storage.Habit, entries *storage.Entries, to civil.Date, daysBack int) map[string][]string {
	tasksUndone := map[string][]string{}
	dayHabits := map[string]bool{}
	from := to.AddDays(-daysBack)
	noFirstRecord := civil.Date{Year: 0, Month: 0, Day: 0}
	// Put in conditional for onboarding starting at 0 days or normal lookback
	if daysBack == 0 {
		for _, habit := range habits {
			dayHabits[habit.Name] = true
		}
		for habit := range dayHabits {
			tasksUndone[from.String()] = append(tasksUndone[from.String()], habit)
		}
	} else {
		for dt := to; !dt.Before(from); dt = dt.AddDays(-1) {
			// build map of habit array to make deletions cleaner
			// +more efficient than linear search array deletes
			for _, habit := range habits {
				dayHabits[habit.Name] = true
			}

			for _, habit := range habits {
				// if habit's target is once, remove from todos if is done in past <interval> days
				if habit.Target <= 1 {
					for days := range habit.Interval {
						if _, ok := (*entries)[storage.DailyHabit{Day: dt.AddDays(-days), Habit: habit.Name}]; ok {
							delete(dayHabits, habit.Name);
							break;
						}
					}
				} else {
					if _, ok := (*entries)[storage.DailyHabit{Day: dt, Habit: habit.Name}]; ok {
						delete(dayHabits, habit.Name)
					}
				}

				// Edge case for 0 day lookback onboard onboards and does not complete at onboard time
				if habit.FirstRecord == noFirstRecord && dt != to {
					delete(dayHabits, habit.Name)
				}
				// Remove days before the first observed outcome of habit
				if dt.Before(habit.FirstRecord) {
					delete(dayHabits, habit.Name)
				}
			}

			for habit := range dayHabits {
				tasksUndone[dt.String()] = append(tasksUndone[dt.String()], habit)
			}
		}
	}
	return tasksUndone
}

// BuildStats calculates statistics for a habit
func BuildStats(habit *storage.Habit, entries *storage.Entries) HabitStats {
	var streaks, breaks, skips int
	var total float64
	now := civil.DateOf(time.Now())
	to := now

	for d := habit.FirstRecord; !d.After(to); d = d.AddDays(1) {
		if outcome, ok := (*entries)[storage.DailyHabit{Day: d, Habit: habit.Name}]; ok {
			switch {
			case outcome.Result == "y":
				streaks += 1
			case outcome.Result == "s":
				skips += 1
			// look at cases of "n" being entered but streak within
			// bounds of a sliding window of the habit every x days
			case graph.Satisfied(d, habit, *entries):
				streaks += 1
			case graph.Skipified(d, habit, *entries):
				skips += 1
			case outcome.Result == "n":
				breaks += 1
			}
			total += outcome.Amount
		}
	}
	return HabitStats{DaysTracked: int((to.DaysSince(habit.FirstRecord)) + 1), Streaks: streaks, Breaks: breaks, Skips: skips, Total: total}
}
