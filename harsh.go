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
	// "gopkg.in/yaml.v2"
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
	EntryDate string
	HabitName string
	Outcome   string
}

var Entries []Entry

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "version, v",
				Value: "0.1",
				Usage: "Version of the Harsh app",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "ask",
				Aliases: []string{"a"},
				Usage:   "Asks you about your undone habits",
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
				Usage:   "Shows you a nice graph of your habits",
				Action: func(c *cli.Context) error {
					habits := loadHabitsConfig()
					entries := loadLog()

					consistencyGraph := map[string][]string{}
					for _, entry := range entries {
						consistencyGraph[entry.HabitName] = append(consistencyGraph[entry.HabitName], graphBuilder(entry.Outcome))
					}

					for _, habit := range habits {
						fmt.Printf("%25v", habit.Name+"  ")
						fmt.Printf(strings.Join(consistencyGraph[habit.Name], "") + "\n")
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

func graphBuilder(outcome string) string {
	var graphDay string
	switch outcome {
	case "y":
		graphDay = "━"
	case "s":
		graphDay = "•"
	case "n":
		graphDay = " "
	default:
		graphDay = "?"
	}
	return graphDay
}

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
			e := Entry{EntryDate: result[0], HabitName: result[1], Outcome: result[2]}
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
