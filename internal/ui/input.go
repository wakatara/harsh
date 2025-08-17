package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal/graph"
	"github.com/wakatara/harsh/internal/storage"
)

// Input handles user input operations
type Input struct {
	colorManager *ColorManager
}

// NewInput creates a new input handler
func NewInput(noColor bool) *Input {
	return &Input{
		colorManager: NewColorManager(noColor),
	}
}

// Onboard prompts new users for initial setup
func (i *Input) Onboard() int {
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

		i.colorManager.PrintfRed("Sorry! Please choose a valid number (0-7) ")
	}
	return numberOfDays
}

// AskHabits handles the interactive habit asking process
func (i *Input) AskHabits(habits []*storage.Habit, entries *storage.Entries, repository storage.Repository, maxHabitNameLength int, countBack int, check string) {
	now := civil.DateOf(time.Now())
	to := now
	from := to.AddDays(-countBack - 40)

	// Goes back 8 days to check unresolved entries
	checkBackDays := 10
	// If log file is empty, we onboard the user
	// For onboarding, we ask how many days to start tracking from
	if len(*entries) == 0 {
		checkBackDays = i.Onboard()
		for _, habit := range habits {
			habit.FirstRecord = to.AddDays(-checkBackDays)
		}
	}
	// Checks for any fragment argument sent along only only asks for it, otherwise all
	filteredHabits := []*storage.Habit{}
	if len(strings.TrimSpace(check)) > 0 {
		askDate, err := civil.ParseDate(check)
		if err == nil {
			from = askDate
			to = from.AddDays(0)
			filteredHabits = habits
		}
		if check == "yday" || check == "yd" {
			from = to.AddDays(-1)
			to = from.AddDays(0)
			filteredHabits = habits
		} else {
			for _, habit := range habits {
				if strings.Contains(strings.ToLower(habit.Name), strings.ToLower(check)) {
					filteredHabits = append(filteredHabits, habit)
				}
			}
		}
	} else {
		filteredHabits = habits
	}

	dayHabits := GetTodos(habits, entries, to, checkBackDays)

	if len(filteredHabits) == 0 {
		fmt.Println("You have no habits that contain that string")
	} else {
		for dt := from; !dt.After(to); dt = dt.AddDays(1) {
			if dayhabit, ok := dayHabits[dt.String()]; ok {

				day, _ := time.Parse(time.DateOnly, dt.String())
				dayOfWeek := day.Weekday().String()[:3]

				i.colorManager.PrintlnBold(dt.String() + " " + dayOfWeek + ":")

				// Go through habit file ordered habits,
				// Check if in returned todos for day and prompt
				heading := ""
				for _, habit := range filteredHabits {
					for _, dh := range dayhabit {
						if habit.Name == dh && (dt.After(habit.FirstRecord) || dt == habit.FirstRecord) {
							if heading != habit.Heading {
								i.colorManager.PrintfBold("\n%s\n", habit.Heading)
								heading = habit.Heading
							}
							for {
								fmt.Printf("%*v", maxHabitNameLength, habit.Name+"  ")
								fmt.Print(graph.BuildGraph(habit, entries, countBack, true))
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
									repository.WriteEntry(dt, habit.Name, result, comment, amount)
									// Updates the Entries map to get updated buildGraph across days
									famount, _ := strconv.ParseFloat(amount, 64)
									(*entries)[storage.DailyHabit{Day: dt, Habit: habit.Name}] = storage.Outcome{Result: result, Amount: famount, Comment: comment}
									break
								}

								i.colorManager.PrintfRed("%*v", maxHabitNameLength+22, "Sorry! Please choose from")
								i.colorManager.PrintfRed(" [y/n/s/⏎] " + "(+ optional @ amounts then # comments)" + "\n")
							}
						}
					}
				}
			}
		}
	}
}
