package internal

import (
	"os"
	"time"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal/storage"
	"golang.org/x/term"
)

// Harsh is the main application struct containing all habit data and configuration
type Harsh struct {
	Repository         storage.Repository
	Habits             []*storage.Habit
	MaxHabitNameLength int
	CountBack          int
	Log               *storage.Log
}

// NewHarsh creates a new Harsh instance with loaded configuration and data
func NewHarsh() *Harsh {
	repository := storage.NewFileRepository()
	habits, maxHabitNameLength, _ := repository.LoadHabits()
	log, _ := repository.LoadEntries()

	now := civil.DateOf(time.Now())
	to := now
	from := to.AddDays(-365 * 5)
	log.Entries.FirstRecords(from, to, habits)

	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Default width for testing or when terminal size cannot be determined
		width = 120
	}
	countBack := max(1, min(width-maxHabitNameLength-2, 100))

	return &Harsh{
		Repository:         repository,
		Habits:             habits,
		MaxHabitNameLength: maxHabitNameLength,
		CountBack:          countBack,
		Log:            log,
	}
}

// GetRepository returns the repository instance
func (h *Harsh) GetRepository() storage.Repository {
	return h.Repository
}

// GetHabits returns the habits slice
func (h *Harsh) GetHabits() []*storage.Habit {
	return h.Habits
}

// GetLog returns the entries map
func (h *Harsh) GetLog() *storage.Log {
	return h.Log
}

// GetMaxHabitNameLength returns the maximum habit name length for formatting
func (h *Harsh) GetMaxHabitNameLength() int {
	return h.MaxHabitNameLength
}

// GetCountBack returns the count back value for graph length
func (h *Harsh) GetCountBack() int {
	return h.CountBack
}
