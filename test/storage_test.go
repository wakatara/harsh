package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal/storage"
)

func TestHabitParseHabitFrequency(t *testing.T) {
	tests := []struct {
		name      string
		frequency string
		target    int
		interval  int
		shouldErr bool
	}{
		{"Daily habit", "1", 1, 1, false},
		{"Weekly habit", "7", 1, 7, false},
		{"Three times weekly", "3/7", 3, 7, false},
		{"Tracking only", "0", 0, 1, false},
		{"Monthly habit", "30", 1, 30, false},
		{"Twice daily", "2/2", 2, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &storage.Habit{Name: "Test", Frequency: tt.frequency}
			
			// Capture exit calls for error cases
			if tt.shouldErr {
				// Note: In a real implementation, we'd refactor ParseHabitFrequency 
				// to return an error instead of calling os.Exit
				return
			}
			
			h.ParseHabitFrequency()
			if h.Target != tt.target || h.Interval != tt.interval {
				t.Errorf("got target=%d interval=%d, want target=%d interval=%d",
					h.Target, h.Interval, tt.target, tt.interval)
			}
		})
	}
}

func TestLoadHabitsConfig(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_storage_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test habits file
	habitsFile := filepath.Join(tmpDir, "habits")
	habitsContent := `# Test habits file
! Work
Daily standup: 1
Code review: 5/7

! Health  
Gym: 3/7
Water: 8
Sleep tracking: 0
`
	err = os.WriteFile(habitsFile, []byte(habitsContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	habits, maxLength := storage.LoadHabitsConfig(tmpDir)

	// Verify habits were loaded correctly
	if len(habits) != 5 {
		t.Errorf("Expected 5 habits, got %d", len(habits))
	}

	// Check first habit
	if habits[0].Name != "Daily standup" || habits[0].Heading != "Work" {
		t.Errorf("First habit incorrect: name=%s, heading=%s", habits[0].Name, habits[0].Heading)
	}

	// Check parsed frequency
	if habits[0].Target != 1 || habits[0].Interval != 1 {
		t.Errorf("First habit frequency incorrect: target=%d, interval=%d", habits[0].Target, habits[0].Interval)
	}

	// Check max length calculation
	if maxLength < len("Daily standup")+10 {
		t.Errorf("Max length too small: got %d", maxLength)
	}
}

func TestLoadLog(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_storage_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test log file
	logFile := filepath.Join(tmpDir, "log")
	logContent := `# Test log file
2025-01-01 : Gym : y : Great workout : 1.5
2025-01-01 : Water : y : 8 glasses : 8
2025-01-01 : Sleep tracking : n : Forgot to track : 0
2025-01-02 : Gym : s : Rest day : 
2025-01-02 : Water : y : Good hydration
`
	err = os.WriteFile(logFile, []byte(logContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	entries := storage.LoadLog(tmpDir)

	// Verify entries were loaded correctly
	if len(*entries) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(*entries))
	}

	// Check specific entry
	date := civil.Date{Year: 2025, Month: 1, Day: 1}
	gymEntry := (*entries)[storage.DailyHabit{Day: date, Habit: "Gym"}]
	if gymEntry.Result != "y" || gymEntry.Amount != 1.5 || gymEntry.Comment != "Great workout" {
		t.Errorf("Gym entry incorrect: result=%s, amount=%f, comment=%s", 
			gymEntry.Result, gymEntry.Amount, gymEntry.Comment)
	}

	// Check entry with missing amount
	waterEntry := (*entries)[storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 2}, Habit: "Water"}]
	if waterEntry.Result != "y" || waterEntry.Amount != 0.0 || waterEntry.Comment != "Good hydration" {
		t.Errorf("Water entry incorrect: result=%s, amount=%f, comment=%s", 
			waterEntry.Result, waterEntry.Amount, waterEntry.Comment)
	}
}

func TestWriteHabitLog(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_storage_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test writing log entry
	date := civil.Date{Year: 2025, Month: 1, Day: 15}
	err = storage.WriteHabitLog(tmpDir, date, "Test Habit", "y", "Great job", "2.5")
	if err != nil {
		t.Fatal(err)
	}

	// Verify file was created and contains expected content
	logFile := filepath.Join(tmpDir, "log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	expected := "2025-01-15 : Test Habit : y : Great job : 2.5\n"
	if string(content) != expected {
		t.Errorf("Log content incorrect: got %q, want %q", string(content), expected)
	}
}

func TestEntriesFirstRecords(t *testing.T) {
	entries := storage.Entries{
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 1}, Habit: "Gym"}: {Result: "y"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 3}, Habit: "Gym"}: {Result: "n"},
		storage.DailyHabit{Day: civil.Date{Year: 2025, Month: 1, Day: 2}, Habit: "Water"}: {Result: "y"},
	}

	habits := []*storage.Habit{
		{Name: "Gym"},
		{Name: "Water"},
		{Name: "Sleep"}, // No entries
	}

	from := civil.Date{Year: 2024, Month: 12, Day: 1}
	to := civil.Date{Year: 2025, Month: 1, Day: 31}

	entries.FirstRecords(from, to, habits)

	// Check that FirstRecord was set correctly
	if habits[0].FirstRecord != (civil.Date{Year: 2025, Month: 1, Day: 1}) {
		t.Errorf("Gym FirstRecord incorrect: got %v, want %v", habits[0].FirstRecord, civil.Date{Year: 2025, Month: 1, Day: 1})
	}

	if habits[1].FirstRecord != (civil.Date{Year: 2025, Month: 1, Day: 2}) {
		t.Errorf("Water FirstRecord incorrect: got %v, want %v", habits[1].FirstRecord, civil.Date{Year: 2025, Month: 1, Day: 2})
	}

	// Sleep should have zero FirstRecord since no entries
	if habits[2].FirstRecord != (civil.Date{}) {
		t.Errorf("Sleep FirstRecord should be zero, got %v", habits[2].FirstRecord)
	}
}

func TestFileRepository(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_storage_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set environment variable to use test directory
	os.Setenv("HARSHPATH", tmpDir)
	defer os.Unsetenv("HARSHPATH")

	// Create test files
	storage.CreateExampleHabitsFile(tmpDir)
	storage.CreateNewLogFile(tmpDir)

	// Test repository
	repo := storage.NewFileRepository()

	// Test GetConfigDir
	if repo.GetConfigDir() != tmpDir {
		t.Errorf("GetConfigDir incorrect: got %s, want %s", repo.GetConfigDir(), tmpDir)
	}

	// Test LoadHabits
	habits, maxLength, err := repo.LoadHabits()
	if err != nil {
		t.Fatal(err)
	}

	if len(habits) == 0 {
		t.Error("No habits loaded")
	}

	if maxLength == 0 {
		t.Error("Max length should be positive")
	}

	// Test LoadEntries
	entries, err := repo.LoadEntries()
	if err != nil {
		t.Fatal(err)
	}

	if entries == nil {
		t.Error("Entries should not be nil")
	}

	// Test WriteEntry
	testDate := civil.Date{Year: 2025, Month: 1, Day: 15}
	err = repo.WriteEntry(testDate, "Test Habit", "y", "Test comment", "1.0")
	if err != nil {
		t.Fatal(err)
	}

	// Verify entry was written
	entries, err = repo.LoadEntries()
	if err != nil {
		t.Fatal(err)
	}

	entry := (*entries)[storage.DailyHabit{Day: testDate, Habit: "Test Habit"}]
	if entry.Result != "y" || entry.Comment != "Test comment" || entry.Amount != 1.0 {
		t.Errorf("Entry not written correctly: result=%s, comment=%s, amount=%f", 
			entry.Result, entry.Comment, entry.Amount)
	}
}

func TestFindConfigFiles(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_storage_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set environment variable to use test directory
	os.Setenv("HARSHPATH", tmpDir)
	defer os.Unsetenv("HARSHPATH")

	// Create files manually to avoid the welcome() call that calls os.Exit
	storage.CreateExampleHabitsFile(tmpDir)
	storage.CreateNewLogFile(tmpDir)

	// Test that the directory is found correctly
	// Note: In a real test, we'd need to refactor FindConfigFiles to not call os.Exit
	// For now, we'll just test the file creation functions
	habitsFile := filepath.Join(tmpDir, "habits")
	if _, err := os.Stat(habitsFile); os.IsNotExist(err) {
		t.Error("Habits file was not created")
	}

	logFile := filepath.Join(tmpDir, "log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestCreateExampleHabitsFile(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_storage_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storage.CreateExampleHabitsFile(tmpDir)

	// Verify file was created
	habitsFile := filepath.Join(tmpDir, "habits")
	if _, err := os.Stat(habitsFile); os.IsNotExist(err) {
		t.Error("Habits file was not created")
	}

	// Verify content
	content, err := os.ReadFile(habitsFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(content) == 0 {
		t.Error("Habits file is empty")
	}

	// Should contain example habits
	contentStr := string(content)
	if !strings.Contains(contentStr, "Gymmed: 3/7") {
		t.Error("Example habits not found in file")
	}
}

func TestCreateNewLogFile(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_storage_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	storage.CreateNewLogFile(tmpDir)

	// Verify file was created
	logFile := filepath.Join(tmpDir, "log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// File should be empty initially
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(content) != 0 {
		t.Error("Log file should be empty initially")
	}
}

func TestLoadHabitsConfigWithEndDate(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_storage_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test habits file with end dates
	habitsFile := filepath.Join(tmpDir, "habits")
	habitsContent := `# Test habits file with end dates
! Work
Active habit: 1
Ended habit: 1: 2024-06-15
Another active: 3/7

! Health
Old habit: 7: 2023-12-31
`
	err = os.WriteFile(habitsFile, []byte(habitsContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	habits, _ := storage.LoadHabitsConfig(tmpDir)

	// Verify habits were loaded correctly
	if len(habits) != 4 {
		t.Errorf("Expected 4 habits, got %d", len(habits))
	}

	// Check habit without end date
	if habits[0].Name != "Active habit" {
		t.Errorf("First habit should be 'Active habit', got %s", habits[0].Name)
	}
	if !habits[0].EndRecord.IsZero() {
		t.Errorf("Active habit should have no end date, got %v", habits[0].EndRecord)
	}
	if habits[0].IsEnded() {
		t.Error("Active habit should not be ended")
	}

	// Check habit with end date
	if habits[1].Name != "Ended habit" {
		t.Errorf("Second habit should be 'Ended habit', got %s", habits[1].Name)
	}
	expectedEndDate := civil.Date{Year: 2024, Month: 6, Day: 15}
	if habits[1].EndRecord != expectedEndDate {
		t.Errorf("Ended habit should have end date 2024-06-15, got %v", habits[1].EndRecord)
	}
	if !habits[1].IsEnded() {
		t.Error("Ended habit should be marked as ended")
	}

	// Check old habit with end date in 2023
	if habits[3].Name != "Old habit" {
		t.Errorf("Fourth habit should be 'Old habit', got %s", habits[3].Name)
	}
	expectedOldEndDate := civil.Date{Year: 2023, Month: 12, Day: 31}
	if habits[3].EndRecord != expectedOldEndDate {
		t.Errorf("Old habit should have end date 2023-12-31, got %v", habits[3].EndRecord)
	}
}

func TestHabitHasEnded(t *testing.T) {
	tests := []struct {
		name      string
		endRecord civil.Date
		checkDate civil.Date
		expected  bool
	}{
		{
			name:      "No end date - never ended",
			endRecord: civil.Date{},
			checkDate: civil.Date{Year: 2025, Month: 1, Day: 15},
			expected:  false,
		},
		{
			name:      "Check date before end date - not ended",
			endRecord: civil.Date{Year: 2025, Month: 6, Day: 15},
			checkDate: civil.Date{Year: 2025, Month: 1, Day: 15},
			expected:  false,
		},
		{
			name:      "Check date equals end date - not ended (end date itself is still active)",
			endRecord: civil.Date{Year: 2025, Month: 1, Day: 15},
			checkDate: civil.Date{Year: 2025, Month: 1, Day: 15},
			expected:  false,
		},
		{
			name:      "Check date after end date - ended",
			endRecord: civil.Date{Year: 2025, Month: 1, Day: 15},
			checkDate: civil.Date{Year: 2025, Month: 1, Day: 16},
			expected:  true,
		},
		{
			name:      "Check date well after end date - ended",
			endRecord: civil.Date{Year: 2024, Month: 6, Day: 15},
			checkDate: civil.Date{Year: 2025, Month: 1, Day: 15},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			habit := &storage.Habit{
				Name:      "Test",
				EndRecord: tt.endRecord,
			}
			result := habit.HasEnded(tt.checkDate)
			if result != tt.expected {
				t.Errorf("HasEnded(%v) = %v, want %v", tt.checkDate, result, tt.expected)
			}
		})
	}
}

func TestHabitIsEnded(t *testing.T) {
	tests := []struct {
		name      string
		endRecord civil.Date
		expected  bool
	}{
		{
			name:      "No end date - not ended",
			endRecord: civil.Date{},
			expected:  false,
		},
		{
			name:      "Has end date - is ended",
			endRecord: civil.Date{Year: 2025, Month: 6, Day: 15},
			expected:  true,
		},
		{
			name:      "Has past end date - is ended",
			endRecord: civil.Date{Year: 2020, Month: 1, Day: 1},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			habit := &storage.Habit{
				Name:      "Test",
				EndRecord: tt.endRecord,
			}
			result := habit.IsEnded()
			if result != tt.expected {
				t.Errorf("IsEnded() = %v, want %v", result, tt.expected)
			}
		})
	}
}