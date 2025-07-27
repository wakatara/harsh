package storage

import "cloud.google.com/go/civil"

// Repository defines the interface for data access operations
type Repository interface {
	// Habit operations
	LoadHabits() ([]*Habit, int, error)
	
	// Log operations
	LoadEntries() (*Entries, error)
	WriteEntry(d civil.Date, habit string, result string, comment string, amount string) error
	
	// Configuration
	GetConfigDir() string
	InitializeConfig() error
}

// FileRepository implements Repository using file-based storage
type FileRepository struct {
	configDir string
}

// NewFileRepository creates a new file-based repository
func NewFileRepository() *FileRepository {
	configDir := FindConfigFiles()
	return &FileRepository{configDir: configDir}
}

// LoadHabits loads habits from the config file
func (r *FileRepository) LoadHabits() ([]*Habit, int, error) {
	habits, maxLength := LoadHabitsConfig(r.configDir)
	return habits, maxLength, nil
}

// LoadEntries loads log entries from the log file
func (r *FileRepository) LoadEntries() (*Entries, error) {
	entries := LoadLog(r.configDir)
	return entries, nil
}

// WriteEntry writes a log entry to the log file
func (r *FileRepository) WriteEntry(d civil.Date, habit string, result string, comment string, amount string) error {
	return WriteHabitLog(r.configDir, d, habit, result, comment, amount)
}

// GetConfigDir returns the configuration directory
func (r *FileRepository) GetConfigDir() string {
	return r.configDir
}

// InitializeConfig initializes the configuration if needed
func (r *FileRepository) InitializeConfig() error {
	// This is handled by FindConfigFiles() which calls welcome() if needed
	return nil
}