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
	// DateFormat is an ISO8601 date
	DateFormat = "2006-01-02"
)

type Day int

type Habit struct {
	Name      string
	Frequency Day
}

var Habits []Habit

// Outcome is explicitly recorded outcome of habit
// on a day and restricted to y,s, or n
type Outcome string

type DailyHabit struct {
	Day   string
	Habit string
}

// Entries maps {ISO date + habit}: Outcome and log format
type Entries map[DailyHabit]Outcome

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "version, v",
				Value: "0.7.0",
				Usage: "Version of the Harsh app",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "ask",
				Aliases: []string{"a"},
				Usage:   "Asks and records your undone habits",
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
					habits := loadHabitsConfig()
					entries := loadLog()
					to := time.Now()
					undone := getTodos(to, 0, *entries)

					for date, todos := range undone {
						fmt.Println(date + ":")
						for _, habit := range habits {
							for _, todo := range todos {
								if habit.Name == todo.Name && todo.Frequency > 0 {
									fmt.Println("     " + todo.Name)
								}
							}
						}
					}

					return nil
				},
			},
			{
				Name:    "log",
				Aliases: []string{"l"},
				Usage:   "Shows graph of habits",
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
						consistency[habit.Name] = append(consistency[habit.Name], buildGraph(&habit, *entries, from, to))
						fmt.Printf("%25v", habit.Name+"  ")
						fmt.Printf(strings.Join(consistency[habit.Name], ""))
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
		if dayhabit, ok := dayHabits[dt.Format(DateFormat)]; ok {
			fmt.Println(dt.Format(DateFormat) + ":")
			// Go through habit file ordered habits,
			// Check if in returned todos for day and prompt
			for _, habit := range habits {
				for _, dh := range dayhabit {
					if habit.Name == dh.Name {
						for {
							fmt.Printf("%25v", habit.Name+"  ")
							fmt.Printf(buildGraph(&habit, *entries, from, to))
							fmt.Printf(" [y/n/s/⏎] ")

							reader := bufio.NewReader(os.Stdin)
							habitResult, err := reader.ReadString('\n')
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
							}

							habitResult = strings.TrimSuffix(habitResult, "\n")
							if strings.ContainsAny(habitResult, "yns") {
								writeHabitLog(dt, habit.Name, habitResult)
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
			if _, ok := entries[DailyHabit{Day: dt.Format(DateFormat), Habit: habit.Name}]; ok {
				delete(dayHabits, habit)
			}
		}

		for habit, _ := range dayHabits {
			tasksUndone[dt.Format(DateFormat)] = append(tasksUndone[dt.Format(DateFormat)], habit)
		}
	}
	return tasksUndone
}

// Consistency graph, sparkline, and scoring functions
func buildSpark(habits []Habit, entries Entries, from time.Time, to time.Time) []string {

	sparkline := []string{}
	sparks := []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	i := 0

	for d := from; d.After(to) == false; d = d.AddDate(0, 0, 1) {
		score := score(d, habits, entries)
		if score == 100 {
			i = 8
		} else {
			i = int(math.Ceil(score / float64(100/(len(sparks)-1))))
		}
		sparkline = append(sparkline, sparks[i])
	}

	return sparkline
}

func buildGraph(habit *Habit, entries Entries, from time.Time, to time.Time) string {
	var graphDay string
	var consistency []string
	for d := from; d.After(to) == false; d = d.AddDate(0, 0, 1) {
		if outcome, ok := entries[DailyHabit{Day: d.Format(DateFormat), Habit: habit.Name}]; ok {
			switch {
			case outcome == "y":
				graphDay = "━"
			case outcome == "s":
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

func satisfied(d time.Time, habit *Habit, entries Entries) bool {
	if habit.Frequency <= 1 {
		return false
	}

	from := d
	to := d.AddDate(0, 0, -int(habit.Frequency))
	for dt := from; dt.Before(to) == false; dt = dt.AddDate(0, 0, -1) {
		if entries[DailyHabit{Day: dt.Format(DateFormat), Habit: habit.Name}] == "y" {
			return true
		}
	}
	return false
}

func skipified(d time.Time, habit *Habit, entries Entries) bool {
	if habit.Frequency <= 1 {
		return false
	}

	from := d
	to := d.AddDate(0, 0, -int(habit.Frequency))
	for dt := from; dt.Before(to) == false; dt = dt.AddDate(0, 0, -1) {
		if entries[DailyHabit{Day: dt.Format(DateFormat), Habit: habit.Name}] == "s" {
			return true
		}
	}
	return false
}

func warning(d time.Time, habit *Habit, entries Entries) bool {
	if habit.Frequency < 1 {
		return false
	}

	warningDays := int(math.Floor(float64(habit.Frequency/7))) + 1
	to := d
	from := d.AddDate(0, 0, -int(habit.Frequency)+warningDays)
	for dt := from; dt.After(to) == false; dt = dt.AddDate(0, 0, 1) {
		if entries[DailyHabit{Day: dt.Format(DateFormat), Habit: habit.Name}] == "y" {
			return false
		}
		if entries[DailyHabit{Day: dt.Format(DateFormat), Habit: habit.Name}] == "s" {
			return false
		}
	}
	return true
}

func score(d time.Time, habits []Habit, entries Entries) float64 {
	scored := 0.0
	skipped := 0.0
	scorableHabits := 0.0

	for _, habit := range habits {
		if habit.Frequency > 0 {
			scorableHabits++
			if outcome, ok := entries[DailyHabit{Day: d.Format(DateFormat), Habit: habit.Name}]; ok {

				switch {
				case outcome == "y":
					scored++
				case outcome == "s":
					skipped++
				// look at cases of n being entered but
				// within bounds of the habit every x days
				case satisfied(d, &habit, entries):
					scored++
				case skipified(d, &habit, entries):
					skipped++
				}
			}
		}
	}
	score := (scored / (scorableHabits - skipped)) * 100
	return score
}

//////////////////////////////////////
// Loading and writing file functions
//////////////////////////////////////

// loadHabitsConfig loads habits in config file ordered slice
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
			h := Habit{Name: result[0], Frequency: Day(r1)}
			Habits = append(Habits, h)
		}
	}
	return Habits
}

// loadLog reads entries from log file
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
			entries[DailyHabit{Day: result[0], Habit: result[1]}] = Outcome(result[2])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return &entries
}

// writeHabitLog writes the log entry for a habit to file
func writeHabitLog(d time.Time, habit string, result string) {
	f, err := os.OpenFile("log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(d.Format(DateFormat) + " : " + habit + " : " + result + "\n")); err != nil {
		f.Close() // ignore error; Write error takes precedence
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
