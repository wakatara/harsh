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