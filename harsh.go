package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/urfave/cli/v2"
)

const (
	ISO = "2006-01-02"
)

// Habit is where I started my review.  General  comment on naming right: Habit
// vs. DailyHabit; whats relation?  Is daily habit a tracking of Habit?
// Structs are not objects. I find the naming a bit confusing.  I then
// immediate tried to find where and how these are used.
type Habit struct {
	name  string
	every int
}

type Outcome string

const (
	// Success marks a completed habit based on your goals
	Success Outcome = "y"

	// Skipped means you didn't do it.  you may have had a good reason.  either way: it wasn't done
	Skipped Outcome = "s"

	// Warning is another outcome - this was hidden magic.
	Warning = "w"
)

var Habits []Habit

type DailyHabit struct {
	day   string
	habit string
}

// Entries is overall data model of habits over time.  It maps {ISO date + habit}: outcome
type Entries map[DailyHabit]Outcome

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "version, v",
				Value: "0.5.5",
				Usage: "Version of the Harsh app",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "ask",
				Aliases: []string{"a"},
				Usage:   "Asks and logs your undone habits",
				Action: func(c *cli.Context) error {

					askHabits()

					return nil
				},
			},
			{
				Name:    "todo",
				Aliases: []string{"t"},
				Usage:   "Shows undone habits for today.",
				Action: func(c *cli.Context) error {
					// habits := loadHabitsConfig()
					entries := loadLog()
					to := time.Now()
					undone := getTodos(to, 0, *entries)

					for date, habits := range undone {
						fmt.Println(date + ":")
						for _, habit := range habits {
							if habit.every > 0 {
								fmt.Println("     " + habit.name)
							}
						}
					}

					return nil
				},
			},
			{
				Name:    "log",
				Aliases: []string{"l"},
				Usage:   "Shows graphs of habits",
				Action: func(c *cli.Context) error {
					habits := loadHabitsConfig()
					entries := loadLog()

					to := time.Now()
					from := to.AddDate(0, 0, -100)
					consistency := map[string][]string{}

					sparkline := buildSpark(habits, *entries, from, to)
					fmt.Printf("%25v", "")
					fmt.Printf(strings.Join(sparkline, ""))
					fmt.Printf("\n")

					for _, habit := range habits {
						consistency[habit.name] = append(consistency[habit.name], buildGraph(&habit, *entries, from, to))
						fmt.Printf("%25v", habit.name+"  ")
						fmt.Printf(strings.Join(consistency[habit.name], ""))
						fmt.Printf("\n")
					}

					scoring := fmt.Sprintf("%.1f", score(time.Now().AddDate(0, 0, -1), habits, *entries))
					fmt.Printf("Yesterday's Score: " + scoring + "%%\n")

					return nil
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Consistency graph, sparkline, and scoring functions
func buildSpark(habits []Habit, entries Entries, from time.Time, to time.Time) []string {

	sparkline := []string{}
	sparks := []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	i := 0

	for d := from; d.After(to) == false; d = d.AddDate(0, 0, 1) {
		// You are shadowing the function score wich is confusing and can result in problems
		entryScore := score(d, habits, entries)
		if entryScore == 100 {
			i = 8
		} else {
			i = int(math.Ceil(entryScore / float64(100/(len(sparks)-1))))
		}
		sparkline = append(sparkline, sparks[i])
	}

	return sparkline
}

func buildGraph(habit *Habit, entries Entries, from time.Time, to time.Time) string {
	var graphDay string
	var consistency []string
	for d := from; d.After(to) == false; d = d.AddDate(0, 0, 1) {
		if outcome, ok := entries[DailyHabit{day: d.Format(ISO), habit: habit.name}]; ok {
			switch {
			case outcome == Success:
				graphDay = "━"
			case outcome == Skipped:
				graphDay = "•"
			// look at cases of n being entered but
			// within bounds of the habit every x days
			case satisfied(d, habit, entries):
				graphDay = "─"
			case skipified(d, habit, entries):
				graphDay = "·"
			case outcome == "n":
				graphDay = " "
			}
		} else {
			// warning sigils max out at 2 weeks (~90 day in formula)
			if warning(d, habit, entries) && (to.Sub(d).Hours() < 336) {
				graphDay = "!"
			} else {
				graphDay = " "
			}
		}
		consistency = append(consistency, graphDay)
	}
	return strings.Join(consistency, "")
}

// satisfied does not satisfy me.  why does it have a reference type?
// This needs the following: a Timestamp to check, a Period to check it for, and a key
// (habit name).  It askes over the last "period" (every), did we hit our goal?  This
// looks back in time, but that is an implementation detail.
// we are asking the question 'did we satisfy Habit as of Day?' (OOP).
// or is Habit a Satisfier?  Habit.Satisfied(d, entry)
//
// we can also change this up: by changing the type to Outcome: can we change the question
// to this?  entries.Outcome(d, habit)
func satisfied(d time.Time, habit *Habit, entries Entries) bool {
	if habit.every <= 1 {
		return false
	}

	from := d
	// needs a period with an implied day unit
	to := d.AddDate(0, 0, -habit.every)
	for dt := from; dt.Before(to) == false; dt = dt.AddDate(0, 0, -1) {
		if entries[DailyHabit{day: dt.Format(ISO), habit: habit.name}] == Success {
			return true
		}
	}
	return false
}

// skipified same same.
// This needs the following: a Timestamp to check, a Period to check it for, and a key
// (habit name)
func skipified(d time.Time, habit *Habit, entries Entries) bool {
	if habit.every <= 1 {
		return false
	}

	from := d
	to := d.AddDate(0, 0, -habit.every)
	for dt := from; dt.Before(to) == false; dt = dt.AddDate(0, 0, -1) {
		if entries[DailyHabit{day: dt.Format(ISO), habit: habit.name}] == Skipped {
			return true
		}
	}
	return false
}

// Warning is a hidden Outcome; because in most of the other calls it seems to be hidden
// or treated as a second class or derived output.
func warning(d time.Time, habit *Habit, entries Entries) bool {
	if habit.every < 1 {
		return false
	}

	warningDays := int(math.Floor(float64(habit.every/7))) + 1
	to := d
	from := d.AddDate(0, 0, -habit.every+warningDays)

	// This is confusing for me.  I'm not really sure what we are trying to accomplish for this.
	for dt := from; dt.After(to) == false; dt = dt.AddDate(0, 0, 1) {
		if entries[DailyHabit{day: dt.Format(ISO), habit: habit.name}] == Success {
			return false
		}
		if entries[DailyHabit{day: dt.Format(ISO), habit: habit.name}] == Skipped {
			return false
		}
	}
	return true
}

// Score (first pass review)is a first step to figure out the interface.  that
// it is taking in Habit is a problem
// (second pass) Now i've gone through the satisfied and skipified functions: i'd pull them
// into their own Outcome method.  I think habit.Outcome(d, entries) and we can ditch the 'repeated'
// cases of 1, 0 or 0, 1
func (e Entries) Score(d time.Time, habit Habit) (success int, skipped int) {
	if outcome, ok := e[DailyHabit{day: d.Format(ISO), habit: habit.name}]; ok {
		switch {
		case outcome == Success:
			return 1, 0
		case outcome == Skipped:
			return 0, 1
		// look at cases of n being entered but
		// within bounds of the habit every x days
		case satisfied(d, &habit, e):
			return 1, 0
		case skipified(d, &habit, e):
			return 0, 1
		}
	}
	return 0, 0
}

// score is a helper: but its also a major behavour of a habit.
// What we are doing is scoring a collection of habits on a specific day
// This, in combination with the variable inside the first If: `scorableHabits`
// makes me feel that Habit.Score(day, entry), or entries.Score(habit, day) should exist.
//
// What we do was is our Scorer.  At this point i think its the entries.
// (second pass review) Now this becomes either a 'main' level function (like it is now)
// that takes in interfaces.  The typing for habits is my sticking points on design here
func score(d time.Time, habits []Habit, entries Entries) float64 {

	scored := 0.0
	skipped := 0.0
	scorableHabits := 0.0

	for _, habit := range habits {
		if habit.every > 0 {
			// don't have better names, so going with the convention you set: y, s
			// Also: habit makes no sense for a `Scorer`
			y, s := entries.Score(d, habit)
			scorableHabits++
			skipped += s
			scored += y
		}
	}
	// missed this
	if scorableHabits == skipped {
		return -1.0
	}
	return (scored / (scorableHabits - skipped)) * 100
}

// Loading of habit and log files
func loadHabitsConfig() []Habit {

	// file, _ := os.Open("/Users/daryl/.config/harsh/habits")
	file, err := os.Open("habits")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if scanner.Text() != "" {
			result := strings.Split(scanner.Text(), ": ")
			r1, _ := strconv.Atoi(result[1])
			h := Habit{name: result[0], every: r1}
			Habits = append(Habits, h)
		}
	}
	return Habits
}

func loadLog() *Entries {
	file, err := os.Open("log")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	entries := Entries{}
	for scanner.Scan() {
		if scanner.Text() != "" {
			result := strings.Split(scanner.Text(), " : ")
			entries[DailyHabit{day: result[0], habit: result[1]}] = result[2]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return &entries
}

// Ask function prompts
func askHabits() {
	habits := loadHabitsConfig()
	entries := loadLog()
	to := time.Now()
	from := to.AddDate(0, 0, -60)

	// Goes back 8 days in case of unresolved entries
	// then iterates through unresolved todos
	dayHabits := getTodos(to, 8, *entries)

	for dt := from; dt.After(to) == false; dt = dt.AddDate(0, 0, 1) {
		if dayhabit, ok := dayHabits[dt.Format(ISO)]; ok {
			fmt.Println(dt.Format(ISO) + ":")
			// Go through habit file ordered habits,
			// Check if in returned todos for day and prompt
			for _, habit := range habits {
				for _, dh := range dayhabit {
					if habit.name == dh.name {
						for {
							fmt.Printf("%25v", habit.name+"  ")
							fmt.Printf(buildGraph(&habit, *entries, from, to))
							fmt.Printf(" [y/n/s/⏎] ")

							reader := bufio.NewReader(os.Stdin)
							habitResult, err := reader.ReadString('\n')
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
							}

							habitResult = strings.TrimSuffix(habitResult, "\n")
							if strings.ContainsAny(habitResult, "yns") {
								writeHabitLog(dt, habit.name, habitResult)
								break
							}

							if habitResult == "" {
								break
							}

							color.FgRed.Printf("%87v", "Sorry! You must choose from")
							color.FgRed.Printf(" [y/n/s/⏎] " + "\n")
						}
					}
				}

			}
		}
	}
}

func getTodos(to time.Time, daysBack int, entries Entries) map[string][]Habit {
	tasksUndone := map[string][]Habit{}
	habits := loadHabitsConfig()
	dayHabits := map[Habit]bool{}

	from := to.AddDate(0, 0, -daysBack)

	for dt := from; dt.After(to) == false; dt = dt.AddDate(0, 0, 1) {
		// build map of habit array to make deletions
		// cleaner+more efficient than linear search array deletes
		for _, habit := range habits {
			dayHabits[habit] = true
		}

		for _, habit := range habits {
			if _, ok := entries[DailyHabit{day: dt.Format(ISO), habit: habit.name}]; ok {
				delete(dayHabits, habit)
			}
		}

		for habit, _ := range dayHabits {
			tasksUndone[dt.Format(ISO)] = append(tasksUndone[dt.Format(ISO)], habit)
		}
	}
	return tasksUndone
}

func writeHabitLog(d time.Time, habit string, result string) {
	f, err := os.OpenFile("log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(d.Format(ISO) + " : " + habit + " : " + result + "\n")); err != nil {
		f.Close() // ignore error; Write error takes precedence
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
