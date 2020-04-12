package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
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
					// habitsLog := readHabitsLog()

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

func loadHabitsConfig() []Habit {

	// file, _ := os.Open("/Users/daryl/.config/harsh/habits")
	file, _ := os.Open("./habits")
	defer file.Close()
	decoder := yaml.NewDecoder(file)
	configuration := []map[string]int{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	for i, _ := range configuration {
		for n, e := range configuration[i] {
			h := Habit{}
			h.Name = n
			h.Every = e
			Habits = append(Habits, h)
		}
	}
	return Habits
}

// func readHabitsLog() []Entry {
// 	return 1
// }

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
	// Put in writing of the line to log file here.

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
