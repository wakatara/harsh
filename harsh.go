package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
)

const (
	layoutISO = "2006-01-02"
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
				Value: "0.4",
				Usage: "Version of the Harsh app",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "ask",
				Aliases: []string{"a"},
				Usage:   "Asks and logs your undone habits",
				Action: func(c *cli.Context) error {
					habits := loadHabitsConfig()
					for _, habit := range habits {
						askHabit(habit.name)
					}

					return nil
				},
			},
			{
				Name:    "todo",
				Aliases: []string{"a"},
				Usage:   "Shows undone habits for today.",
				Action: func(c *cli.Context) error {
					// habits := loadHabitsConfig()
					entries := loadLog()
					to := time.Now()
					undone := getTodos(to, 0, *entries)

					for date, habits := range undone {
						fmt.Println(date + ":")
						for _, habit := range habits {
							fmt.Println("     " + habit)
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
		if outcome, ok := entries[DailyHabit{day: d.Format(layoutISO), habit: habit.name}]; ok {
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
		if entries[DailyHabit{day: dt.Format(layoutISO), habit: habit.name}] == "y" {
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
		if entries[DailyHabit{day: dt.Format(layoutISO), habit: habit.name}] == "s" {
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
		if entries[DailyHabit{day: dt.Format(layoutISO), habit: habit.name}] == "y" {
			return false
		}
		if entries[DailyHabit{day: dt.Format(layoutISO), habit: habit.name}] == "s" {
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
			if outcome, ok := entries[DailyHabit{day: d.Format(layoutISO), habit: habit.name}]; ok {

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

func askHabit(habit string) {
	validate := func(input string) error {
		err := !(strings.ContainsAny(input, "yns") || input == "")
		if err != false {
			return errors.New("Must be [y/n/s/⏎]")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    habit + " [y/n/s/⏎]",
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	if result != "" {
		writeHabitLog(habit, result)
	}

}

func getTodos(to time.Time, daysBack int, entries Entries) map[string][]string {
	// returns a map of date => habitName
	tasksUndone := map[string][]string{}
	dayHabits := []Habit{}
	habits := loadHabitsConfig()
	for _, habit := range habits {
		if habit.every > 0 {
			dayHabits = append(dayHabits, habit)
		}
	}

	from := to.AddDate(0, 0, -daysBack)
	for dt := from; dt.After(to) == false; dt = dt.AddDate(0, 0, 1) {
		dh := dayHabits
		for _, habit := range habits {
			if _, ok := entries[DailyHabit{day: dt.Format(layoutISO), habit: habit.name}]; ok {
				// finds and removes found keys from copy of habits array so returned in order
				for i, h := range dh {
					if h.name == habit.name {
						dh = append(dh[:i], dh[i+1:]...)
						break
					}
				}
			}
		}
		for _, dayHabit := range dayHabits {
			tasksUndone[dt.Format(layoutISO)] = append(tasksUndone[dt.Format(layoutISO)], dayHabit.name)
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
	// date, _ := time.Parse(layoutISO, rightNow) // for when parsing passed dates
}
