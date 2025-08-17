package storage

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"cloud.google.com/go/civil"
)

// Habit represents a habit with its configuration and tracking information
type Habit struct {
	Heading     string
	Name        string
	Frequency   string
	Target      int
	Interval    int
	FirstRecord civil.Date
}

const DEFAULT_HABITS = 
`# This is your habits file.
# It tells harsh what to track and how frequently.
# 1 means daily, 7 means weekly, 14 every two weeks.
# You can also track targets within a set number of days.
# For example, Gym 3 times a week would translate to 3/7.
# 0 is for tracking a habit. 0 frequency habits will not warn or score.
# Examples:

Gymmed: 3/7
Bed by midnight: 1
Cleaned House: 7
Called Mom: 7
Tracked Finances: 15
New Skill: 90
Too much coffee: 0
Used harsh: 0
`

// ParseHabitFrequency parses the frequency string and sets Target and Interval
func (habit *Habit) ParseHabitFrequency() {
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

// LoadHabitsConfig loads habits in config file ordered slice
func LoadHabitsConfig(configDir string) ([]*Habit, int) {
	habitsPath := filepath.Join(configDir, "/habits")
	file, err := os.Open(habitsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Check for common cloud storage scenarios
			icloudPath := filepath.Join(configDir, ".habits.icloud")
			if _, err := os.Stat(icloudPath); err == nil {
				fmt.Println("Error: Your habits file is currently syncing with iCloud.")
				fmt.Println("The file appears as '.habits.icloud' while syncing.")
				fmt.Println("Please wait for sync to complete, or disable iCloud for the harsh folder.")
				os.Exit(1)
			}
			
			// Check if config directory exists but habits file doesn't
			if _, err := os.Stat(configDir); err == nil {
				fmt.Printf("Error: Habits file not found at %s\n", habitsPath)
				fmt.Println("This might be your first time using harsh.")
				fmt.Println("Run 'harsh' without arguments to create an example habits file.")
				os.Exit(1)
			}
			
			// Config directory doesn't exist
			fmt.Printf("Error: Configuration directory not found at %s\n", configDir)
			fmt.Println("Run 'harsh' without arguments to initialize your configuration.")
			os.Exit(1)
		}
		
		// For permission errors or other issues, provide context
		if os.IsPermission(err) {
			fmt.Printf("Error: Permission denied accessing habits file at %s\n", habitsPath)
			fmt.Println("Check file permissions or try running with appropriate privileges.")
			os.Exit(1)
		}
		
		// For other errors, use the original behavior but with more context
		fmt.Printf("Error opening habits file at %s: %v\n", habitsPath, err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var heading string
	var habits []*Habit
	lineCount := 0
	
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		
		if len(line) > 0 {
			if line[0] == '!' {
				// Parse heading line
				if !strings.Contains(line, "! ") {
					fmt.Printf("Warning: Malformed heading at line %d: %s\n", lineCount, line)
					fmt.Println("Expected format: ! Heading Name")
					continue
				}
				result := strings.Split(line, "! ")
				if len(result) > 1 {
					heading = result[1]
				}
			} else if line[0] != '#' {
				// Parse habit line
				if !strings.Contains(line, ": ") {
					fmt.Printf("Warning: Skipping malformed habit at line %d: %s\n", lineCount, line)
					fmt.Println("Expected format: Habit Name: frequency")
					continue
				}
				
				result := strings.Split(line, ": ")
				if len(result) < 2 {
					fmt.Printf("Warning: Skipping habit with missing frequency at line %d: %s\n", lineCount, line)
					continue
				}
				
				habitName := strings.TrimSpace(result[0])
				frequency := strings.TrimSpace(result[1])
				
				if habitName == "" {
					fmt.Printf("Warning: Skipping habit with empty name at line %d\n", lineCount)
					continue
				}
				
				if frequency == "" {
					fmt.Printf("Warning: Skipping habit '%s' with empty frequency at line %d\n", habitName, lineCount)
					continue
				}
				
				h := Habit{Heading: heading, Name: habitName, Frequency: frequency}
				
				// ParseHabitFrequency may call os.Exit on invalid frequency
				// This is the intended behavior for invalid config
				(&h).ParseHabitFrequency()
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

// FindConfigFiles checks os relevant habits and log file exist, returns path
// If they do not exist, calls CreateExampleHabitsFile and CreateNewLogFile
func FindConfigFiles() string {
	configDir := os.Getenv("HARSHPATH")

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

// CreateExampleHabitsFile writes a fresh Habits file for people to follow
func CreateExampleHabitsFile(configDir string) {
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
		defer f.Close()
		f.WriteString(DEFAULT_HABITS)
	}
}

// CreateNewLogFile writes an empty log file for people to start tracking into
func CreateNewLogFile(configDir string) {
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

// welcome a new user and creates example habits and log files
func welcome(configDir string) {
	CreateExampleHabitsFile(configDir)
	CreateNewLogFile(configDir)
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
