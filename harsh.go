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

type Habit struct {
	name  string
	every int
}

var Habits []Habit

type DailyHabit struct {
	day   string
	habit string
}

type Entries map[DailyHabit]string

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
						fmt.Printf("│" + "\n")
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
		if outcome, ok := entries[DailyHabit{day: d.Format(ISO), habit: habit.name}]; ok {
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
			if warning(d, habit, entries) {
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
	if habit.every <= 1 {
		return false
	}

	from := d
	to := d.AddDate(0, 0, -habit.every)
	for dt := from; dt.Before(to) == false; dt = dt.AddDate(0, 0, -1) {
		if entries[DailyHabit{day: dt.Format(ISO), habit: habit.name}] == "y" {
			return true
		}
	}
	return false
}

func skipified(d time.Time, habit *Habit, entries Entries) bool {
	if habit.every <= 1 {
		return false
	}

	from := d
	to := d.AddDate(0, 0, -habit.every)
	for dt := from; dt.Before(to) == false; dt = dt.AddDate(0, 0, -1) {
		if entries[DailyHabit{day: dt.Format(ISO), habit: habit.name}] == "s" {
			return true
		}
	}
	return false
}

func warning(d time.Time, habit *Habit, entries Entries) bool {
	if habit.every < 1 {
		return false
	}

	warningDays := int(math.Floor(float64(habit.every/7))) + 1
	to := d
	from := d.AddDate(0, 0, -habit.every+warningDays)
	for dt := from; dt.After(to) == false; dt = dt.AddDate(0, 0, 1) {
		if entries[DailyHabit{day: dt.Format(ISO), habit: habit.name}] == "y" {
			return false
		}
		if entries[DailyHabit{day: dt.Format(ISO), habit: habit.name}] == "s" {
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
		if habit.every > 0 {
			scorableHabits++
			if outcome, ok := entries[DailyHabit{day: d.Format(ISO), habit: habit.name}]; ok {

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
	// habits := loadHabitsConfig()
	to := time.Now()
	from := to.AddDate(0, 0, -60)

	// Goes back 8 days in case of unresolved entries
	// then iterates through unresolved todos
	dayHabits := getTodos(to, 8, *entries)

	for dt := from; dt.After(to) == false; dt = dt.AddDate(0, 0, 1) {
		if dayhabit, ok := dayHabits[dt.Format(ISO)]; ok {
			fmt.Println(dt.Format(ISO) + ":")
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
								writeHabitLog(habit.name, habitResult)
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
	// returns a map of date => habitName
	tasksUndone := map[string][]Habit{}
	habits := loadHabitsConfig()
	dayHabits := map[Habit]bool{}

	from := to.AddDate(0, 0, -daysBack)

	for dt := from; dt.After(to) == false; dt = dt.AddDate(0, 0, 1) {
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

func writeHabitLog(habit string, result string) {
	date := (time.Now()).Format("2006-01-02")
	f, err := os.OpenFile("log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(date + " : " + habit + " : " + result + "\n")); err != nil {
		f.Close() // ignore error; Write error takes precedence
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	// date, _ := time.Parse(ISO, rightNow) // for when parsing passed dates
}
