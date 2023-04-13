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

	"cloud.google.com/go/civil"
	"github.com/gookit/color"
	"github.com/urfave/cli/v2"
)

var configDir string

type Days int

type Habit struct {
	Heading   string
	Name      string
	Frequency Days
}

// Outcome is the explicit recorded outcome of habit on a day (y, n, or s)
type Outcome string

// DailyHabit combines Day and Habit with an Outcome to yield Entries
type DailyHabit struct {
	Day   civil.Date
	Habit string
}

// HabitStats holds total stats for a Habit in the file
type HabitStats struct {
	DaysTracked int
	Streaks     int
	Breaks      int
	Skips       int
}

// Entries maps DailyHabit{ISO date + habit}: Outcome and log format
type Entries map[DailyHabit]Outcome

type Harsh struct {
	Habits             []Habit
	MaxHabitNameLength int
	Entries            *Entries
}

func main() {
	app := &cli.App{
		Name:        "Harsh",
		Usage:       "habit tracking for geeks",
		Description: "A simple, minimalist CLI for tracking and understanding habits.",
		Version:     "0.8.26",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "no-color",
				Aliases: []string{"n"},
				Usage:   "no colors in output",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "ask",
				Aliases: []string{"a"},
				Usage:   "Asks and records your undone habits",
				Action: func(_ *cli.Context) error {
					harsh := newHarsh()
					harsh.askHabits()
					return nil
				},
			},
			{
				Name:    "todo",
				Aliases: []string{"t"},
				Usage:   "Shows undone habits for today.",
				Action: func(_ *cli.Context) error {
					harsh := newHarsh()
					to := civil.DateOf(time.Now())
					undone := harsh.getTodos(to, 0)

					heading := ""
					if len(undone) == 0 {
						fmt.Println("All todos logged up to today.")
					} else {
						for date, todos := range undone {
							color.Bold.Println(date + ":")
							for _, habit := range harsh.Habits {
								for _, todo := range todos {
									if heading != habit.Heading && habit.Heading == todo.Heading {
										color.Bold.Printf("\n" + habit.Heading + "\n")
										heading = habit.Heading
									}
									if habit.Name == todo.Name {
										fmt.Printf("%*v", harsh.MaxHabitNameLength, todo.Name+"\n")
									}
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
				Usage:   "Shows graph of logged habits",
				Action: func(_ *cli.Context) error {
					harsh := newHarsh()

					to := civil.DateOf(time.Now())
					from := to.AddDays(-100)
					firstRecords := harsh.firstRecords(from, to)
					consistency := map[string][]string{}

					sparkline := harsh.buildSpark(from, to)
					fmt.Printf("%*v", harsh.MaxHabitNameLength, "")
					fmt.Print(strings.Join(sparkline, ""))
					fmt.Printf("\n")

					heading := ""
					for _, habit := range harsh.Habits {
						consistency[habit.Name] = append(consistency[habit.Name], harsh.buildGraph(&habit, firstRecords[habit], from, to))
						if heading != habit.Heading {
							color.Bold.Printf(habit.Heading + "\n")
							heading = habit.Heading
						}
						fmt.Printf("%*v", harsh.MaxHabitNameLength, habit.Name+"  ")
						fmt.Print(strings.Join(consistency[habit.Name], ""))
						fmt.Printf("\n")
					}

					scoring := fmt.Sprintf("%.1f", harsh.score(civil.DateOf(time.Now()).AddDays(-1)))
					fmt.Printf("\n" + "Yesterday's Score: " + scoring + "%%\n")

					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name:    "stats",
						Aliases: []string{"s"},
						Usage:   "Shows habit stats for entire log file",
						Action: func(c *cli.Context) error {
							harsh := newHarsh()

							to := civil.DateOf(time.Now())
							from := to.AddDays(-(365 * 5))
							firstRecords := harsh.firstRecords(from, to)
							stats := map[string]HabitStats{}

							heading := ""
							for _, habit := range harsh.Habits {
								if c.Bool("no-color") {
									color.Disable()
								}
								if heading != habit.Heading {
									color.Bold.Printf("\n" + habit.Heading + "\n")
									heading = habit.Heading
								}
								stats[habit.Name] = harsh.buildStats(&habit, firstRecords[habit], to)
								fmt.Printf("%*v", harsh.MaxHabitNameLength, habit.Name+"  ")
								color.FgGreen.Printf("Streaks ")
								color.FgGreen.Printf("%4v", strconv.Itoa(stats[habit.Name].Streaks))
								color.FgGreen.Printf(" days")
								fmt.Printf("%4v", "")
								color.FgRed.Printf("Breaks ")
								color.FgRed.Printf("%4v", strconv.Itoa(stats[habit.Name].Breaks))
								color.FgRed.Printf(" days")
								fmt.Printf("%4v", "")
								color.FgYellow.Printf("Skips ")
								color.FgYellow.Printf("%4v", strconv.Itoa(stats[habit.Name].Skips))
								color.FgYellow.Printf(" days")
								fmt.Printf("%4v", "")
								fmt.Printf("Tracked ")
								fmt.Printf("%4v", strconv.Itoa(stats[habit.Name].DaysTracked))
								fmt.Printf(" days\n")
							}
							return nil
						},
					},
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

func newHarsh() *Harsh {
	config := findConfigFiles()
	habits, maxHabitNameLength := loadHabitsConfig(config)
	entries := loadLog(config)

	return &Harsh{habits, maxHabitNameLength, entries}
}

// Ask function prompts
func (h *Harsh) askHabits() {

	to := civil.DateOf(time.Now())
	from := to.AddDays(-60)
	firstRecords := h.firstRecords(from, to)

	// Goes back 8 days to check unresolved entries
	checkBackDays := 8
	// If log file is empty, we onboard the user
	// For onboarding, we ask how many days to start tracking from
	if len(*h.Entries) == 0 {
		checkBackDays = onboard()
		for _, habit := range h.Habits {
			firstRecords[habit] = to.AddDays(-(checkBackDays + 1))
		}
	}

	dayHabits := h.getTodos(to, checkBackDays)

	for dt := from; !dt.After(to); dt = dt.AddDays(1) {
		if dayhabit, ok := dayHabits[dt.String()]; ok {

			color.Bold.Println(dt.String() + ":")

			// Go through habit file ordered habits,
			// Check if in returned todos for day and prompt
			heading := ""
			for _, habit := range h.Habits {
				for _, dh := range dayhabit {
					if habit.Name == dh.Name && dt.After(firstRecords[habit]) {
						if heading != dh.Heading {
							color.Bold.Printf("\n" + habit.Heading + "\n")
							heading = habit.Heading
						}
						for {
							fmt.Printf("%*v", h.MaxHabitNameLength, habit.Name+"  ")
							fmt.Print(h.buildGraph(&habit, firstRecords[habit], from, to))
							fmt.Printf(" [y/n/s/⏎] ")

							reader := bufio.NewReader(os.Stdin)
							habitResultInput, err := reader.ReadString('\n')
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
							}

							habitResultString := strings.Split(habitResultInput, "#")
							habitResult := strings.TrimSpace(habitResultString[0])
							var comment string
							if len(habitResultString) > 1 {
								comment = strings.TrimSpace(habitResultString[1])
							} else {
								comment = ""
							}

							if strings.ContainsAny(habitResult, "yns") && len(habitResult) == 1 {
								writeHabitLog(dt, habit.Name, habitResult, comment)
								// Updates the Entries map to get updated buildGraph across days
								(*h.Entries)[DailyHabit{dt, habit.Name}] = Outcome(habitResult)
								break
							}

							if habitResult == "" {
								break
							}

							color.FgRed.Printf("%*v", h.MaxHabitNameLength+25, "Sorry! Please choose from")
							color.FgRed.Printf(" [y/n/s/⏎] " + "(and an optional #-denoted comment)" + "\n")
						}
					}
				}

			}
		}
	}
}

func (h *Harsh) firstRecords(from civil.Date, to civil.Date) map[Habit]civil.Date {
	firstRecords := map[Habit]civil.Date{}
	for dt := to; !dt.Before(from); dt = dt.AddDays(-1) {
		for _, habit := range h.Habits {
			if _, ok := (*h.Entries)[DailyHabit{Day: dt, Habit: habit.Name}]; ok {
				firstRecords[habit] = dt
			}
		}
	}
	return firstRecords
}

func (h *Harsh) getTodos(to civil.Date, daysBack int) map[string][]Habit {
	tasksUndone := map[string][]Habit{}
	dayHabits := map[Habit]bool{}
	from := to.AddDays(-daysBack)
	firstRecords := h.firstRecords(from, to)
	for dt := to; !dt.Before(from); dt = dt.AddDays(-1) {
		// build map of habit array to make deletions cleaner
		// +more efficient than linear search array deletes
		for _, habit := range h.Habits {
			dayHabits[habit] = true
		}

		for _, habit := range h.Habits {
			if _, ok := (*h.Entries)[DailyHabit{Day: dt, Habit: habit.Name}]; ok {
				delete(dayHabits, habit)
			}
			if dt.Before(firstRecords[habit]) {
				delete(dayHabits, habit)
			}
		}

		for habit := range dayHabits {
			tasksUndone[dt.String()] = append(tasksUndone[dt.String()], habit)
		}
	}
	return tasksUndone
}

// Consistency graph, sparkline, and scoring functions
func (h *Harsh) buildSpark(from civil.Date, to civil.Date) []string {

	sparkline := []string{}
	sparks := []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	i := 0

	for d := from; !d.After(to); d = d.AddDays(1) {
		dailyScore := h.score(d)
		// divide score into  score to map to sparks slice graphic for sparkline
		if dailyScore == 100 {
			i = 8
		} else {
			i = int(math.Ceil(dailyScore / float64(100/(len(sparks)-1))))
		}
		sparkline = append(sparkline, sparks[i])
	}

	return sparkline
}

func (h *Harsh) buildGraph(habit *Habit, firstRecord civil.Date, from civil.Date, to civil.Date) string {
	var graphDay string
	var consistency []string

	for d := from; !d.After(to); d = d.AddDays(1) {
		if outcome, ok := (*h.Entries)[DailyHabit{Day: d, Habit: habit.Name}]; ok {
			switch {
			case outcome == "y":
				graphDay = "━"
			case outcome == "s":
				graphDay = "•"
			// look at cases of "n" being entered but
			// within bounds of the habit every x days
			case satisfied(d, habit, *h.Entries):
				graphDay = "─"
			case skipified(d, habit, *h.Entries):
				graphDay = "·"
			case outcome == "n":
				graphDay = " "
			}
		} else {
			if warning(d, habit, *h.Entries, firstRecord) && (to.DaysSince(d) < 14) {
				// warning: sigils max out at 2 weeks (~90 day habit in formula)
				graphDay = "!"
			} else {
				graphDay = " "
			}
		}
		consistency = append(consistency, graphDay)
	}
	return strings.Join(consistency, "")
}

func (h *Harsh) buildStats(habit *Habit, firstRecord civil.Date, to civil.Date) HabitStats {
	var streaks int
	var breaks int
	var skips int

	for d := firstRecord; !d.After(to); d = d.AddDays(1) {
		if outcome, ok := (*h.Entries)[DailyHabit{Day: d, Habit: habit.Name}]; ok {
			switch {
			case outcome == "y":
				streaks += 1
			case outcome == "s":
				skips += 1
			// look at cases of "n" being entered but
			// within bounds of the habit every x days
			case satisfied(d, habit, *h.Entries):
				streaks += 1
			case skipified(d, habit, *h.Entries):
				skips += 1
			case outcome == "n":
				breaks += 1
			}
		}
	}
	return HabitStats{DaysTracked: int((to.DaysSince(firstRecord)) + 1), Streaks: streaks, Breaks: breaks, Skips: skips}
}

func satisfied(d civil.Date, habit *Habit, entries Entries) bool {
	if habit.Frequency <= 1 {
		return false
	}

	from := d
	to := d.AddDays(-int(habit.Frequency))
	for dt := from; !dt.Before(to); dt = dt.AddDays(-1) {
		if entries[DailyHabit{Day: dt, Habit: habit.Name}] == "y" {
			return true
		}
	}
	return false
}

func skipified(d civil.Date, habit *Habit, entries Entries) bool {
	if habit.Frequency <= 1 {
		return false
	}

	from := d
	to := d.AddDays(-int(habit.Frequency))
	for dt := from; !dt.Before(to); dt = dt.AddDays(-1) {
		if entries[DailyHabit{Day: dt, Habit: habit.Name}] == "s" {
			return true
		}
	}
	return false
}

func warning(d civil.Date, habit *Habit, entries Entries, firstRecord civil.Date) bool {
	if habit.Frequency < 1 {
		return false
	}

	warningDays := int(habit.Frequency)/7 + 1
	to := d
	from := d.AddDays(-int(habit.Frequency) + warningDays)
	for dt := from; !dt.After(to); dt = dt.AddDays(1) {
		if entries[DailyHabit{Day: dt, Habit: habit.Name}] == "y" {
			return false
		}
		if entries[DailyHabit{Day: dt, Habit: habit.Name}] == "s" {
			return false
		}
		if dt.Before(firstRecord) {
			return false
		}
	}
	return true
}

func (h *Harsh) score(d civil.Date) float64 {
	scored := 0.0
	skipped := 0.0
	scorableHabits := 0.0

	for _, habit := range h.Habits {
		if habit.Frequency > 0 {
			scorableHabits++
			if outcome, ok := (*h.Entries)[DailyHabit{Day: d, Habit: habit.Name}]; ok {

				switch {
				case outcome == "y":
					scored++
				case outcome == "s":
					skipped++
				// look at cases of n being entered but
				// within bounds of the habit every x days
				case satisfied(d, &habit, *h.Entries):
					scored++
				case skipified(d, &habit, *h.Entries):
					skipped++
				}
			}
		}
	}

	score := 100.0 // deal with scorable habits - skipped = 0 causing divide by zero issue
	if scorableHabits-skipped != 0 {
		score = (scored / (scorableHabits - skipped)) * 100
	}
	return score
}

//////////////////////////////////////
// Loading and writing file functions
//////////////////////////////////////

// loadHabitsConfig loads habits in config file ordered slice
func loadHabitsConfig(configDir string) ([]Habit, int) {

	file, err := os.Open(filepath.Join(configDir, "/habits"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var heading string
	var habits []Habit
	for scanner.Scan() {
		if len(scanner.Text()) > 0 {
			if scanner.Text()[0] == '!' {
				result := strings.Split(scanner.Text(), "! ")
				heading = result[1]
			} else if scanner.Text()[0] != '#' {
				result := strings.Split(scanner.Text(), ": ")
				r1, _ := strconv.Atoi(result[1])
				h := Habit{Heading: heading, Name: result[0], Frequency: Days(r1)}
				habits = append(habits, h)
			}
		}
	}

	maxHabitNameLength := 0
	for _, h := range habits {
		if len(h.Name) > maxHabitNameLength {
			maxHabitNameLength = len(h.Name)
		}
	}

	return habits, maxHabitNameLength + 10
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
		if len(scanner.Text()) > 0 {
			if scanner.Text()[0] != '#' {
				// Discards comments from read record read as result[3]
				result := strings.Split(scanner.Text(), " : ")
				cd, err := civil.ParseDate(result[0])
				if err != nil {
					fmt.Println("Error parsing log date format.")
				}
				entries[DailyHabit{Day: cd, Habit: result[1]}] = Outcome(result[2])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return &entries
}

// writeHabitLog writes the log entry for a habit to file
func writeHabitLog(d civil.Date, habit string, result string, comment string) {
	fileName := filepath.Join(configDir, "/log")
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if len(comment) > 0 {
		comment = " : " + comment
	}
	if _, err := f.Write([]byte(d.String() + " : " + habit + " : " + result + comment + "\n")); err != nil {
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

	configDir = os.Getenv("HARSHPATH")

	if len(configDir) == 0 {
		if runtime.GOOS == "windows" {
			configDir = filepath.Join(os.Getenv("APPDATA"), "harsh")
		} else {
			configDir = filepath.Join(os.Getenv("HOME"), ".config/harsh")
		}
	}

	if _, err := os.Stat(filepath.Join(configDir, "habits")); err == nil {
	} else {
		welcome(configDir)
	}

	return configDir
}

// welcome a new user and creates example habits and log files
func welcome(configDir string) {
	createExampleHabitsFile(configDir)
	createNewLogFile(configDir)
	fmt.Println("Welcome to harsh!")
	fmt.Println("Created " + filepath.Join(configDir, "/habits") + "   This file lists your habits.")
	fmt.Println("Created " + filepath.Join(configDir, "/log") + "      This file is your habit log.")
	fmt.Println("")
	fmt.Println("No habits of your own yet?")
	fmt.Println("Open your habits file @ " + filepath.Join(configDir, "/habits"))
	fmt.Println("with a text editor (nano, vim, VS Code, Atom, emacs) and modify and save the habits list.")
	fmt.Println("Then:")
	fmt.Println("Run       harsh ask     to start tracking")
	fmt.Println("Running   harsh todo    will show you undone habits for today.")
	fmt.Println("Running   harsh log     will show you a consistency graph of your efforts.")
	fmt.Println("                        (the graph gets way cooler looking over time.")
	fmt.Println("For more depth, you can read https://github.com/wakatara/harsh#usage")
	fmt.Println("")
	fmt.Println("Happy tracking! I genuinely hope this helps you with your goals. Buena suerte!")
	os.Exit(0)
}

// first time ask is used and log empty asks user how far back to track
func onboard() int {
	fmt.Println("Your log file looks empty. Let's setup your tracking.")
	fmt.Println("How many days back shall we start tracking from in days?")
	fmt.Println("harsh will ask you about each habit for every day back.")
	fmt.Println("Starting today would be 0. Choose. (0-7) ")
	var numberOfDays int
	for {
		reader := bufio.NewReader(os.Stdin)
		dayResult, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		dayResult = strings.TrimSpace(dayResult)
		dayNum, err := strconv.Atoi(dayResult)
		if err == nil {
			if dayNum >= 0 && dayNum <= 7 {
				numberOfDays = dayNum
				break
			}
		}

		color.FgRed.Printf("Sorry! Please choose a valid number (0-7) ")
	}
	return numberOfDays
}

// createExampleHabitsFile writes a fresh Habits file for people to follow
func createExampleHabitsFile(configDir string) {
	fileName := filepath.Join(configDir, "/habits")
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			os.MkdirAll(configDir, os.ModePerm)
		}
		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
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
			os.MkdirAll(configDir, os.ModePerm)
		}
		_, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
	}
}
