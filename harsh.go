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
	layoutISO = "2020-02-20"
)

type Habit struct {
	Name  string
	Every int
}

var Habits []Habit

type Entry struct {
	EntryDate time.Time
	HabitName string
	Outcome   string
}

var Entries []Entry

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
						askHabit(habit.Name)
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
						graph[habit.Name] = append(graph[habit.Name], buildGraph(&habit, &entries, from, to))
					}

					for _, habit := range habits {
						fmt.Printf("%25v", habit.Name+"  ")
						fmt.Printf(strings.Join(graph[habit.Name], "") + "\n")
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

func buildGraph(habit *Habit, entries *[]Entry, from time.Time, to time.Time) string {
	var graphDay string
	var dayOutcome string
	var consistency []string

	for d := from; d.After(to) == false; d = d.AddDate(0, 0, 1) {
		for _, entry := range *entries {
			if entry.HabitName == habit.Name && entry.EntryDate == d {
				dayOutcome = entry.Outcome
			}
		}

		if dayOutcome == "y" {
			graphDay = "━"
		} else if dayOutcome == "s" {
			graphDay = "•"
			// } else if satisfied(entry.Habit, entry.EntryDate) {
			// 	graphDay = "─"
		} else if dayOutcome == "n" {
			graphDay = " "
		} else if dayOutcome == "" {
			graphDay = "?"
		}

		consistency = append(consistency, graphDay)
	}

	return strings.Join(consistency, "")
}

// func getEntry(habit *Habit, date time.Time) string {
// 	for _, entry := range Entries {
// 		if entry.HabitName == habit.Name && entry.EntryDate == date {
// 			return entry.Outcome
// 		}
// 	}
// 	return "nil"
// }

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
			h := Habit{Name: result[0], Every: r1}
			Habits = append(Habits, h)
		}
	}
	return Habits
}

func loadLog() []Entry {
	file, err := os.Open("log")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if scanner.Text() != "" {
			result := strings.Split(scanner.Text(), " : ")
			result0, _ := time.Parse(layoutISO, result[0])
			e := Entry{EntryDate: result0, HabitName: result[1], Outcome: result[2]}
			Entries = append(Entries, e)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return Entries
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

func satisfied(habit Habit, entryDate time.Time) bool {
	if habit.Every < 1 {
		return false
	}

	// from := entryDate.AddDate(0, 0, -habit.Every)
	// end := entryDate
	// for d := from; d.After(end) == false; d = d.AddDate(0, 0, 1) {
	// 	if entry_outcome(date) == "y" {
	// 		return true
	// 	}
	// }
	return false

}
