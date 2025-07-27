package storage

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cloud.google.com/go/civil"
)

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

// Entries maps DailyHabit{ISO date + habit}: Outcome and log format
type Entries map[DailyHabit]Outcome

// LoadLog reads entries from log file
func LoadLog(configDir string) *Entries {
	file, err := os.Open(filepath.Join(configDir, "/log"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	entries := Entries{}
	lineCount := 0
	for scanner.Scan() {
		lineCount++
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
						fmt.Printf("Error: there is a non-number in your log file at line %d where we expect a number.\n", lineCount)
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

// WriteHabitLog writes the log entry for a habit to file
func WriteHabitLog(configDir string, d civil.Date, habit string, result string, comment string, amount string) error {
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

// FirstRecords sets the FirstRecord field for habits based on their earliest entries
func (e *Entries) FirstRecords(from civil.Date, to civil.Date, habits []*Habit) {
	for dt := to; !dt.Before(from); dt = dt.AddDays(-1) {
		for _, habit := range habits {
			if _, ok := (*e)[DailyHabit{Day: dt, Habit: habit.Name}]; ok {
				habit.FirstRecord = dt
			}
		}
	}
}