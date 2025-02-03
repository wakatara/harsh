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
	"golang.org/x/term"
)

var configDir string

type Days int

type Habit struct {
	Heading     string
	Name        string
	Frequency   string
	Target      int
	Interval    int
	FirstRecord civil.Date
}

// Outcome is the explicit recorded result of a habit
// on a day (y, n, or s) and an optional amount and comment
type Outcome struct {
	Result  string
	Amount  float64
	Comment string
}

// DailyHabit combines Day and Habit with an Outcome to yield Entries
type DailyHabit struct {
	Day   civil.Date
	Habit string
}

// HabitStats holds total stats for a Habit in the file
type HabitStats struct {
	DaysTracked int
	Total       float64
	Streaks     int
	Breaks      int
	Skips       int
}

// Entries maps DailyHabit{ISO date + habit}: Outcome and log format
type Entries map[DailyHabit]Outcome

type Harsh struct {
	Habits             []*Habit
	MaxHabitNameLength int
	CountBack          int
	Entries            *Entries
}

func main() {
	app := &cli.App{
		Name:        "Harsh",
		Usage:       "habit tracking for geeks",
		Description: "A simple, minimalist CLI for tracking and understanding habits.",
		Version:     "0.10.7",
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
				Action: func(c *cli.Context) error {
					harsh := newHarsh()
					habit_fragment := c.Args().First()

					harsh.askHabits(habit_fragment)
					return nil
				},
			},
			{
				Name:    "todo",
				Aliases: []string{"t"},
				Usage:   "Shows undone habits for today.",
				Action: func(_ *cli.Context) error {
					harsh := newHarsh()
					now := civil.DateOf(time.Now())
					undone := harsh.getTodos(now, 0)

					heading := ""
					if len(undone) == 0 {
						fmt.Println("All todos logged up to today.")
					} else {
						for date, todos := range undone {
							color.Bold.Println(date + ":")
							for _, habit := range harsh.Habits {
								for _, todo := range todos {
									if heading != habit.Heading && habit.Heading == todo {
										color.Bold.Printf("\n" + habit.Heading + "\n")
										heading = habit.Heading
									}
									if habit.Name == todo {
										fmt.Printf("%*v", harsh.MaxHabitNameLength, todo+"\n")
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
				Action: func(c *cli.Context) error {
					harsh := newHarsh()
					habit_fragment := c.Args().First()

					// Checks for any fragment argument sent along only only asks for it, otherwise all
					habits := []*Habit{}
					if len(strings.TrimSpace(habit_fragment)) > 0 {
						for _, habit := range harsh.Habits {
							if strings.Contains(strings.ToLower(habit.Name), strings.ToLower(habit_fragment)) {
								habits = append(habits, habit)
							}
						}
					} else {
						habits = harsh.Habits
					}

					now := civil.DateOf(time.Now())
					to := now
					from := to.AddDays(-harsh.CountBack)
					consistency := map[string][]string{}
					undone := harsh.getTodos(to, 0)

					sparkline, calline := harsh.buildSpark(from, to)
					fmt.Printf("%*v", harsh.MaxHabitNameLength, "")
					fmt.Print(strings.Join(sparkline, ""))
					fmt.Printf("\n")
					fmt.Printf("%*v", harsh.MaxHabitNameLength, "")
					fmt.Print(strings.Join(calline, ""))
					fmt.Printf("\n")

					heading := ""
					for _, habit := range habits {
						consistency[habit.Name] = append(consistency[habit.Name], harsh.buildGraph(habit, false))
						if heading != habit.Heading {
							color.Bold.Printf(habit.Heading + "\n")
							heading = habit.Heading
						}
						fmt.Printf("%*v", harsh.MaxHabitNameLength, habit.Name+"  ")
						fmt.Print(strings.Join(consistency[habit.Name], ""))
						fmt.Printf("\n")
					}

					undone_num := strconv.Itoa(len(undone[to.String()]))

					scoring := fmt.Sprintf("%.1f", harsh.score(now.AddDays(-1)))
					fmt.Printf("\n" + "Yesterday's Score: ")
					fmt.Printf("%9v", scoring)
					fmt.Printf("%%\n")
					if undone_num == "0" {
						fmt.Printf("All todos logged for today.")
					} else {
						fmt.Printf("Today's unlogged todos: ")
						fmt.Printf("%2v", undone_num)
					}
					fmt.Printf("\n")

					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name:    "stats",
						Aliases: []string{"s"},
						Usage:   "Shows habit stats for entire log file",
						Action: func(c *cli.Context) error {
							harsh := newHarsh()

							// to := civil.DateOf(time.Now())
							// from := to.AddDays(-(365 * 5))
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
								stats[habit.Name] = harsh.buildStats(habit)
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
								fmt.Printf(" days")
								if stats[habit.Name].Total == 0 {
									fmt.Printf("%4v", "")
									fmt.Printf("      ")
									fmt.Printf("%5v", "")
									fmt.Printf("     \n")
								} else {
									fmt.Printf("%4v", "")
									color.FgBlue.Printf("Total ")
									color.FgBlue.Printf("%5v", (stats[habit.Name].Total))
									color.FgBlue.Printf("     \n")
								}
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
	now := civil.DateOf(time.Now())
	to := now
	from := to.AddDays(-365 * 5)
	entries.firstRecords(from, to, habits)
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	countBack := 100
	if width < 100+maxHabitNameLength {
		// Remove 2 for the habitName and graph padding
		countBack = width - maxHabitNameLength - 2
	}

	return &Harsh{habits, maxHabitNameLength, countBack, entries}
}

// Ask function prompts
func (h *Harsh) askHabits(check string) {
	now := civil.DateOf(time.Now())
	to := now
	from := to.AddDays(-h.CountBack - 40)

	// Goes back 8 days to check unresolved entries
	checkBackDays := 10
	// If log file is empty, we onboard the user
	// For onboarding, we ask how many days to start tracking from
	if len(*h.Entries) == 0 {
		checkBackDays = onboard()
		for _, habit := range h.Habits {
			habit.FirstRecord = to.AddDays(-checkBackDays)
		}
	}

	// Checks for any fragment argument sent along only only asks for it, otherwise all
	habits := []*Habit{}
	if len(strings.TrimSpace(check)) > 0 {
		for _, habit := range h.Habits {
			if strings.Contains(strings.ToLower(habit.Name), strings.ToLower(check)) {
				habits = append(habits, habit)
			}
		}
	} else {
		habits = h.Habits
	}

	dayHabits := h.getTodos(to, checkBackDays)

	for dt := from; !dt.After(to); dt = dt.AddDays(1) {
		if dayhabit, ok := dayHabits[dt.String()]; ok {

			color.Bold.Println(dt.String() + ":")

			// Go through habit file ordered habits,
			// Check if in returned todos for day and prompt
			heading := ""
			for _, habit := range habits {
				for _, dh := range dayhabit {
					if habit.Name == dh && (dt.After(habit.FirstRecord) || dt == habit.FirstRecord) {
						if heading != habit.Heading {
							color.Bold.Printf("\n" + habit.Heading + "\n")
							heading = habit.Heading
						}
						for {
							fmt.Printf("%*v", h.MaxHabitNameLength, habit.Name+"  ")
							fmt.Print(h.buildGraph(habit, true))
							fmt.Printf(" [y/n/s/⏎] ")

							reader := bufio.NewReader(os.Stdin)
							habitResultInput, err := reader.ReadString('\n')
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
							}

							// No input
							if len(habitResultInput) == 1 {
								break
							}

							// Sanitize : colons out of string for log files
							habitResultInput = strings.ReplaceAll(habitResultInput, ":", "")

							var result, amount, comment string
							atIndex := strings.Index(habitResultInput, "@")
							hashIndex := strings.Index(habitResultInput, "#")

							if atIndex > 0 && hashIndex > 0 && atIndex < hashIndex {
								parts := strings.SplitN(habitResultInput, "@", 2)
								secondParts := strings.SplitN(parts[1], "#", 2)
								result = strings.TrimSpace(parts[0])
								amount = strings.TrimSpace(secondParts[0])
								comment = strings.TrimSpace(secondParts[1])
							}
							// only has an @ Amount
							if hashIndex == -1 && atIndex > 0 {
								parts := strings.SplitN(habitResultInput, "@", 2)
								result = strings.TrimSpace(parts[0])
								amount = strings.TrimSpace(parts[1])
								comment = ""
							}
							// only has a # comment
							if atIndex == -1 && hashIndex > 0 {
								parts := strings.SplitN(habitResultInput, "#", 2)
								result = strings.TrimSpace(parts[0])
								amount = ""
								comment = strings.TrimSpace(parts[1])
							}
							if atIndex == -1 && hashIndex == -1 {
								result = strings.TrimSpace(habitResultInput)
							}

							if strings.ContainsAny(result, "yns") && len(result) == 1 {
								writeHabitLog(dt, habit.Name, result, comment, amount)
								// Updates the Entries map to get updated buildGraph across days
								famount, _ := strconv.ParseFloat(amount, 64)
								(*h.Entries)[DailyHabit{dt, habit.Name}] = Outcome{Result: result, Amount: famount, Comment: comment}
								break
							}

							color.FgRed.Printf("%*v", h.MaxHabitNameLength+22, "Sorry! Please choose from")
							color.FgRed.Printf(" [y/n/s/⏎] " + "(+ optional @ amounts then # comments)" + "\n")
						}
					}
				}
			}
		}
	}
}

func (e *Entries) firstRecords(from civil.Date, to civil.Date, habits []*Habit) {
	for dt := to; !dt.Before(from); dt = dt.AddDays(-1) {
		for _, habit := range habits {
			if _, ok := (*e)[DailyHabit{Day: dt, Habit: habit.Name}]; ok {
				habit.FirstRecord = dt
			}
		}
	}
}

func (h *Harsh) getTodos(to civil.Date, daysBack int) map[string][]string {
	tasksUndone := map[string][]string{}
	dayHabits := map[string]bool{}
	from := to.AddDays(-daysBack)
	noFirstRecord := civil.Date{0, 0, 0}
	// Put in conditional for onboarding starting at 0 days or normal lookback
	if daysBack == 0 {
		for _, habit := range h.Habits {
			dayHabits[habit.Name] = true
		}
		for habit := range dayHabits {
			tasksUndone[from.String()] = append(tasksUndone[from.String()], habit)
		}
	} else {
		for dt := to; !dt.Before(from); dt = dt.AddDays(-1) {
			// build map of habit array to make deletions cleaner
			// +more efficient than linear search array deletes
			for _, habit := range h.Habits {
				dayHabits[habit.Name] = true
			}

			for _, habit := range h.Habits {

				if _, ok := (*h.Entries)[DailyHabit{Day: dt, Habit: habit.Name}]; ok {
					delete(dayHabits, habit.Name)
				}
				// Edge case for 0 day lookback onboard onboards and does not complete at onboard time
				if habit.FirstRecord == noFirstRecord && dt != to {
					delete(dayHabits, habit.Name)
				}
				// Remove days before the first observed outcome of habit
				if dt.Before(habit.FirstRecord) {
					delete(dayHabits, habit.Name)
				}
			}

			for habit := range dayHabits {
				tasksUndone[dt.String()] = append(tasksUndone[dt.String()], habit)
			}
		}
	}
	return tasksUndone
}

// Consistency graph, sparkline, and scoring functions
func (h *Harsh) buildSpark(from civil.Date, to civil.Date) ([]string, []string) {
	sparkline := []string{}
	calline := []string{}
	sparks := []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	i := 0
	LetterDay := map[string]string{
		"Sunday": " ", "Monday": "M", "Tuesday": " ", "Wednesday": "W",
		"Thursday": " ", "Friday": "F", "Saturday": " ",
	}

	for d := from; !d.After(to); d = d.AddDays(1) {
		dailyScore := h.score(d)
		// divide score into  score to map to sparks slice graphic for sparkline
		if dailyScore == 100 {
			i = 8
		} else {
			i = int(math.Ceil(dailyScore / float64(100/(len(sparks)-1))))
		}
		t, _ := time.Parse(time.RFC3339, d.String()+"T00:00:00Z")
		w := t.Weekday().String()

		calline = append(calline, LetterDay[w])
		sparkline = append(sparkline, sparks[i])
	}

	return sparkline, calline
}

func (h *Harsh) buildGraph(habit *Habit, ask bool) string {
	graphLen := h.CountBack
	var graphDay string
	var consistency strings.Builder

	to := civil.DateOf(time.Now())
	from := to.AddDays(-h.CountBack)
	if ask {
		graphLen = h.CountBack + 12
	}
	consistency.Grow(graphLen)

	for d := from; !d.After(to); d = d.AddDays(1) {
		if outcome, ok := (*h.Entries)[DailyHabit{Day: d, Habit: habit.Name}]; ok {
			switch {
			case outcome.Result == "y":
				graphDay = "━"
			case outcome.Result == "s":
				graphDay = "•"
			// look at cases of "n" being entered but
			// within bounds of the habit every x days
			case satisfied(d, habit, *h.Entries):
				graphDay = "─"
			case skipified(d, habit, *h.Entries):
				graphDay = "·"
			case outcome.Result == "n":
				graphDay = " "
			}
		} else {
			if warning(d, habit, *h.Entries) && (to.DaysSince(d) < 14) {
				// warning: sigils max out at 2 weeks (~90 day habit in formula)
				graphDay = "!"
			} else {
				graphDay = " "
			}
		}
		consistency.WriteString(graphDay)
	}

	return consistency.String()
}

func (h *Harsh) buildStats(habit *Habit) HabitStats {
	var streaks, breaks, skips int
	var total float64
	now := civil.DateOf(time.Now())
	to := now

	for d := habit.FirstRecord; !d.After(to); d = d.AddDays(1) {
		if outcome, ok := (*h.Entries)[DailyHabit{Day: d, Habit: habit.Name}]; ok {
			switch {
			case outcome.Result == "y":
				streaks += 1
			case outcome.Result == "s":
				skips += 1
			// look at cases of "n" being entered but
			// within bounds of the habit every x days
			case satisfied(d, habit, *h.Entries):
				streaks += 1
			case skipified(d, habit, *h.Entries):
				skips += 1
			case outcome.Result == "n":
				breaks += 1
			}
			total += outcome.Amount
		}
	}
	return HabitStats{DaysTracked: int((to.DaysSince(habit.FirstRecord)) + 1), Streaks: streaks, Breaks: breaks, Skips: skips, Total: total}
}

func satisfied(d civil.Date, habit *Habit, entries Entries) bool {
	if habit.Target <= 1 && habit.Interval == 1 {
		return false
	}

	// Look back from date, interval days and see if target count satisfied
	target_counter := 0
	from := d
	to := d.AddDays(-int(habit.Interval))
	for dt := from; !dt.Before(to); dt = dt.AddDays(-1) {
		if v, ok := entries[DailyHabit{Day: dt, Habit: habit.Name}]; ok {
			if v.Result == "y" {
				target_counter++
			}
		}
	}
	return target_counter >= habit.Target
}

func skipified(d civil.Date, habit *Habit, entries Entries) bool {
	if habit.Target <= 1 && habit.Interval == 1 {
		return false
	}

	from := d
	to := d.AddDays(-int(math.Ceil(float64(habit.Interval) / float64(habit.Target))))
	for dt := from; !dt.Before(to); dt = dt.AddDays(-1) {
		if v, ok := entries[DailyHabit{Day: dt, Habit: habit.Name}]; ok {
			if v.Result == "s" {
				return true
			}
		}
	}
	return false
}

func warning(d civil.Date, habit *Habit, entries Entries) bool {
	if habit.Target < 1 {
		return false
	}

	warningDays := int(habit.Interval)/7 + 1
	to := d
	from := d.AddDays(-int(habit.Interval) + warningDays)
	noFirstRecord := civil.Date{0, 0, 0}
	for dt := from; !dt.After(to); dt = dt.AddDays(1) {
		if v, ok := entries[DailyHabit{Day: dt, Habit: habit.Name}]; ok {
			switch v.Result {
			case "y":
				return false
			case "s":
				return false
			}
		}
		// Edge case for 0 day onboard and later completes null entry habits
		if habit.FirstRecord == noFirstRecord {
			return false
		}
		if dt.Before(habit.FirstRecord) {
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
		if habit.Target > 0 && !d.Before(habit.FirstRecord) {
			scorableHabits++
			if outcome, ok := (*h.Entries)[DailyHabit{Day: d, Habit: habit.Name}]; ok {
				switch {
				case outcome.Result == "y":
					scored++
				case outcome.Result == "s":
					skipped++
				// look at cases of n being entered but
				// within bounds of the habit every x days
				case satisfied(d, habit, *h.Entries):
					scored++
				case skipified(d, habit, *h.Entries):
					skipped++
				}
			}
		}
	}

	var score float64
	// Edge case on if there is nothing to score and the scorable vs skipped issue
	if scorableHabits == 0 {
		score = 0.0
	} else {
		score = 100.0 // deal with scorable habits - skipped == 0 causing divide by zero issue
	}
	if scorableHabits-skipped != 0 {
		score = (scored / (scorableHabits - skipped)) * 100
	}
	return score
}

//////////////////////////////////////
// Loading and writing file functions
//////////////////////////////////////

// loadHabitsConfig loads habits in config file ordered slice
func loadHabitsConfig(configDir string) ([]*Habit, int) {
	file, err := os.Open(filepath.Join(configDir, "/habits"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var heading string
	var habits []*Habit
	for scanner.Scan() {
		if len(scanner.Text()) > 0 {
			if scanner.Text()[0] == '!' {
				result := strings.Split(scanner.Text(), "! ")
				heading = result[1]
			} else if scanner.Text()[0] != '#' {
				result := strings.Split(scanner.Text(), ": ")
				h := Habit{Heading: heading, Name: result[0], Frequency: result[1]}
				(&h).parseHabitFrequency()
				habits = append(habits, &h)
			}
		}
	}

	maxHabitNameLength := 0
	for _, habit := range habits {
		if len(habit.Name) > maxHabitNameLength {
			maxHabitNameLength = len(habit.Name)
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
	line_count := 0
	for scanner.Scan() {
		line_count++
		if len(scanner.Text()) > 0 {
			if scanner.Text()[0] != '#' {
				// Discards comments from read record read as result[3]
				result := strings.Split(scanner.Text(), " : ")
				cd, err := civil.ParseDate(result[0])
				if err != nil {
					fmt.Println("Error parsing log date format.")
				}
				switch len(result) {
				case 5:
					if result[4] == "" {
						result[4] = "0"
					}
					amount, err := strconv.ParseFloat(result[4], 64)
					if err != nil {
						fmt.Printf("Error: there is a non-number in your log file at line %d where we expect a number.\n", line_count)
					}
					entries[DailyHabit{Day: cd, Habit: result[1]}] = Outcome{Result: result[2], Comment: result[3], Amount: amount}
				case 4:
					entries[DailyHabit{Day: cd, Habit: result[1]}] = Outcome{Result: result[2], Comment: result[3], Amount: 0.0}
				default:
					entries[DailyHabit{Day: cd, Habit: result[1]}] = Outcome{Result: result[2], Comment: "", Amount: 0.0}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return &entries
}

func (habit *Habit) parseHabitFrequency() {
	freq := strings.Split(habit.Frequency, "/")
	target, err := strconv.Atoi(strings.TrimSpace(freq[0]))
	if err != nil {
		fmt.Println("Error: A frequency in your habit file has a non-integer before the slash.")
		fmt.Println("The problem entry to fix is: " + habit.Name + " : " + habit.Frequency)
		os.Exit(1)
	}

	var interval int
	if len(freq) == 1 {
		if target == 0 {
			interval = 1
		} else {
			interval = target
			target = 1
		}
	} else {
		interval, err = strconv.Atoi(strings.TrimSpace(freq[1]))
		if err != nil || interval == 0 {
			fmt.Println("Error: A frequency in your habit file has a non-integer or zero after the slash.")
			fmt.Println("The problem entry to fix is: " + habit.Name + " : " + habit.Frequency)
			os.Exit(1)
		}
	}
	if target > interval {
		fmt.Println("Error: A frequency in your habit file has a target value greater than the interval period.")
		fmt.Println("The problem entry to fix is: " + habit.Name + " : " + habit.Frequency)
		os.Exit(1)
	}
	habit.Target = target
	habit.Interval = interval
}

// writeHabitLog writes the log entry for a habit to file
func writeHabitLog(d civil.Date, habit string, result string, comment string, amount string) error {
	fileName := filepath.Join(configDir, "/log")
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write([]byte(d.String() + " : " + habit + " : " + result + " : " + comment + " : " + amount + "\n")); err != nil {
		f.Close() // ignore error; Write error takes precedence
		return fmt.Errorf("writing log file: %w", err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	return nil
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
		f.WriteString("# You can also track targets within a set number of days.\n")
		f.WriteString("# For example, Gym 3 times a week would translate to 3/7.\n")
		f.WriteString("# 0 is for tracking a habit. 0 frequency habits will not warn or score.\n")
		f.WriteString("# Examples:\n\n")
		f.WriteString("Gymmed: 3/7\n")
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
