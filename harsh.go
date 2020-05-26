package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
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

var configDir string

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
		Name:        "Harsh",
		Usage:       "habit tracking for geeks",
		Description: "A simple, minimalist CLI for tracking and understanding habits.",
		Version:     "0.8.1",
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
					config := findConfigFiles()
					habits := loadHabitsConfig(config)
					entries := loadLog(config)
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
				Usage:   "Shows log graph of habits",
				Action: func(c *cli.Context) error {
					config := findConfigFiles()
					habits := loadHabitsConfig(config)
					entries := loadLog(config)

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
	config := findConfigFiles()
	habits := loadHabitsConfig(config)
	entries := loadLog(config)
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

							color.FgRed.Printf("%86v", "Sorry! You must choose from")
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
	config := findConfigFiles()
	habits := loadHabitsConfig(config)
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

		for habit := range dayHabits {
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
			// look at cases of "n" being entered but
			// within bounds of the habit every x days
			case satisfied(d, habit, entries):
				graphDay = "─"
			case skipified(d, habit, entries):
				graphDay = "·"
			case outcome == "n":
				graphDay = " "
			}
		} else {
			// warning sigils max out at 2 weeks (~90 day habit in formula)
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
func loadHabitsConfig(configDir string) []Habit {

	file, err := os.Open(filepath.Join(configDir, "/habits"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if len(scanner.Text()) > 0 && scanner.Text()[0] != '#' {
			result := strings.Split(scanner.Text(), ": ")
			r1, _ := strconv.Atoi(result[1])
			h := Habit{Name: result[0], Frequency: Day(r1)}
			Habits = append(Habits, h)
		}
	}
	return Habits
}

// loadLog reads entries from log file
func loadLog(configDir string) *Entries {
	file, err := os.Open(filepath.Join(configDir, "/log"))
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
	fileName := filepath.Join(configDir, "/log")
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

// findConfigFile checks os relevant habits and log file exist, returns path
// If they do not exist, calls writeNewHabits and writeNewLog
func findConfigFiles() string {
	if runtime.GOOS == "windows" {
		configDir = "AppData"
	} else {
		configDir = filepath.Join(os.Getenv("HOME"), ".config/harsh")
	}

	if _, err := os.Stat(filepath.Join(configDir, "habits")); err == nil {
	} else {
		welcome(configDir)
	}

	return configDir
}

// welcome onboards a new user and creates example habits and log files
func welcome(configDir string) {
	createExampleHabitsFile(configDir)
	createNewLogFile(configDir)
	fmt.Println("Welcome to harsh!\n")
	fmt.Println("Created " + filepath.Join(configDir, "/habits") + "   This file lists your habits.")
	fmt.Println("Created " + filepath.Join(configDir, "/log") + "      This file is your habit log.")

	fmt.Println("\nNo habits of your own yet?")
	fmt.Println("Open your habits file @ " + filepath.Join(configDir, "/habits"))
	fmt.Println("with a text editor (nano, vim, VS Code, Atom, emacs) and modify and save the habits list.")
	fmt.Println("Then:\n")
	fmt.Println("Run       harsh ask     to start tracking")
	fmt.Println("Running   harsh todo    will show you undone habits for today.")
	fmt.Println("Running   harsh log     will show you a consistency graph of your efforts.")
	fmt.Println("                        (the graph gets way cooler looking over time.")
	fmt.Println("\nHappy tracking! I genuinely hope this helps you achieve your goals. Bonne chance!\n")
	os.Exit(0)
}

// createExampleHabitsFile writes a fresh Habits file for people to follow
func createExampleHabitsFile(configDir string) {
	fileName := filepath.Join(configDir, "/habits")
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			os.Mkdir(configDir, 0755)
		}
		f, _ := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		f.WriteString("# This is your habits file.\n")
		f.WriteString("# It tells harsh what to track and how frequently.\n")
		f.WriteString("# 1 means daily, 7 means weekly, 14 every two weeks.\n")
		f.WriteString("# 0 is for tracking a habit. 0 frequency habits will not warn or score.\n")
		f.WriteString("# Examples:\n\n")
		f.WriteString("Gymmed: 2\n")
		f.WriteString("Bed by midnight: 1\n")
		f.WriteString("Cleaned House: 7\n")
		f.WriteString("Called Mom: 7\n")
		f.WriteString("Tracked Finances: 15\n")
		f.WriteString("New Skill: 90\n")
		f.WriteString("Too much coffee: 0\n")
		f.WriteString("Used harsh: 0\n")
		f.Close()
	}
}

// createNewLogFile writes an empty log file for people to start tracking into
func createNewLogFile(configDir string) {
	fileName := filepath.Join(configDir, "/log")
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			os.Mkdir(configDir, 0755)
		}
		os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0644)
	}
}
