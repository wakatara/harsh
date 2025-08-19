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
	logPath := filepath.Join(configDir, "/log")
	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Check for common cloud storage scenarios
			icloudPath := filepath.Join(configDir, ".log.icloud")
			if _, err := os.Stat(icloudPath); err == nil {
				fmt.Println("Error: Your log file is currently syncing with iCloud.")
				fmt.Println("The file appears as '.log.icloud' while syncing.")
				fmt.Println("Please wait for sync to complete, or disable iCloud for the harsh folder.")
				os.Exit(1)
			}
			
			// Check if config directory exists but log file doesn't
			if _, err := os.Stat(configDir); err == nil {
				fmt.Printf("Error: Log file not found at %s\n", logPath)
				fmt.Println("This might be your first time using harsh.")
				fmt.Println("Run 'harsh' without arguments to initialize your configuration.")
				os.Exit(1)
			}
			
			// Config directory doesn't exist
			fmt.Printf("Error: Configuration directory not found at %s\n", configDir)
			fmt.Println("Run 'harsh' without arguments to initialize your configuration.")
			os.Exit(1)
		}
		
		// For permission errors or other issues, provide context
		if os.IsPermission(err) {
			fmt.Printf("Error: Permission denied accessing log file at %s\n", logPath)
			fmt.Println("Check file permissions or try running with appropriate privileges.")
			os.Exit(1)
		}
		
		// For other errors, use the original behavior but with more context
		fmt.Printf("Error opening log file at %s: %v\n", logPath, err)
		os.Exit(1)
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
				
				// Check for minimum required fields (date, habit, result)
				if len(result) < 3 {
					fmt.Printf("Warning: Skipping malformed log entry at line %d: %s\n", lineCount, scanner.Text())
					fmt.Println("Expected format: YYYY-MM-DD : Habit Name : y/n/s : Comment : Amount")
					continue
				}
				
				cd, err := civil.ParseDate(result[0])
				if err != nil {
					fmt.Printf("Warning: Skipping log entry with invalid date at line %d: %s\n", lineCount, result[0])
					continue
				}
				
				// Validate habit name is not empty
				if strings.TrimSpace(result[1]) == "" {
					fmt.Printf("Warning: Skipping log entry with empty habit name at line %d\n", lineCount)
					continue
				}
				
				// Validate result is y, n, or s
				result[2] = strings.TrimSpace(result[2])
				if result[2] != "y" && result[2] != "n" && result[2] != "s" {
					fmt.Printf("Warning: Skipping log entry with invalid result '%s' at line %d (expected y/n/s)\n", result[2], lineCount)
					continue
				}
				
				switch len(result) {
				case 5:
					if result[4] == "" {
						result[4] = "0"
					}
					amount, err := strconv.ParseFloat(result[4], 64)
					if err != nil {
						fmt.Printf("Warning: Invalid amount '%s' at line %d, using 0.0\n", result[4], lineCount)
						amount = 0.0
					}
					entries[DailyHabit{Day: cd, Habit: result[1]}] = Outcome{Result: result[2], Comment: result[3], Amount: amount}
				case 4:
					entries[DailyHabit{Day: cd, Habit: result[1]}] = Outcome{Result: result[2], Comment: result[3], Amount: 0.0}
				case 3:
					entries[DailyHabit{Day: cd, Habit: result[1]}] = Outcome{Result: result[2], Comment: "", Amount: 0.0}
				default:
					// This shouldn't happen due to the check above, but just in case
					fmt.Printf("Warning: Unexpected number of fields (%d) at line %d\n", len(result), lineCount)
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
		// Provide more specific error messages based on the type of error
		if os.IsNotExist(err) {
			return fmt.Errorf("configuration directory does not exist: %s", configDir)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied writing to log file: %s (check file permissions)", fileName)
		}
		// Check for disk space issues (this is a common cause of write failures)
		return fmt.Errorf("cannot open log file %s: %w (this might be due to insufficient disk space or file system issues)", fileName, err)
	}
	defer f.Close()
	logEntry := d.String()
	for _, item := range []string{habit, result, comment, amount} {
		if item != "" {
			logEntry+=" : " + item
		}
	}
	logEntry += "\n"
	if _, err := f.Write([]byte(logEntry)); err != nil {
		f.Close() // ignore error; Write error takes precedence
		// Check for common write failure causes
		if strings.Contains(err.Error(), "no space left") || strings.Contains(err.Error(), "disk full") {
			return fmt.Errorf("failed to write log entry: disk full or insufficient space")
		}
		return fmt.Errorf("failed to write log entry to %s: %w", fileName, err)
	}
	if err := f.Close(); err != nil {
		// Convert this from log.Fatal to a proper error return
		return fmt.Errorf("failed to close log file %s: %w", fileName, err)
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
