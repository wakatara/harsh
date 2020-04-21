package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
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
				Value: "0.2",
				Usage: "Version of the Harsh app",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "ask",
				Aliases: []string{"a"},
				Usage:   "Asks about and logs your undone habits",
				Action: func(c *cli.Context) error {
					habits := loadHabitsConfig()
					for _, habit := range habits {
						askHabit(habit.name)
					}

					return nil
				},
			},
			{
				Name:    "log",
				Aliases: []string{"l"},
				Usage:   "Shows a consistency graph of your habits",
				Action: func(c *cli.Context) error {
					habits := loadHabitsConfig()
					entries := loadLog()

					to := time.Now()
					from := to.AddDate(0, 0, -100)
					graph := map[string][]string{}

					for _, habit := range habits {
						graph[habit.name] = append(graph[habit.name], buildGraph(&habit, *entries, from, to))
						fmt.Printf("%25v", habit.name+"  ")
						fmt.Printf(strings.Join(graph[habit.name], ""))
						fmt.Printf("|" + "\n")

					}

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
			case satisfied(d, habit, entries):
				graphDay = "─"
			case skipified(d, habit, entries):
				graphDay = "·"
			case outcome == "n":
				graphDay = " "
			}
		} else {
			graphDay = " "
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

// Loading of files section

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

// func satisfied(habit Habit, entryDate time.Time) bool {
// 	if habit.Every < 1 {
// 		return false
// 	}

// 	// from := entryDate.AddDate(0, 0, -habit.Every)
// 	// end := entryDate
// 	// for d := from; d.After(end) == false; d = d.AddDate(0, 0, 1) {
// 	// 	if entry_outcome(date) == "y" {
// 	// 		return true
// 	// 	}
// 	// }
// 	return false

// }
