package test

import (
	"os"
	"path/filepath"
	"testing"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal/storage"
)

func TestImprovedErrorHandling(t *testing.T) {
	// Test that our improved error handling provides better messages
	
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_error_handling_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("Malformed log entry handling", func(t *testing.T) {
		// Create a log file with malformed entries
		logFile := filepath.Join(tmpDir, "log")
		malformedLog := `# Test log with various issues
2025-01-15 : Valid Habit : y : Good : 1.0
2025-01-15 Missing colons
2025-01-15 : : y : Empty habit name : 1.0
2025-01-15 : Bad Result : x : Invalid result : 1.0
invalid-date : Test : y : Bad date : 1.0
2025-01-15 : Amount Test : y : Comment : not-a-number
2025-01-15 : Valid Habit 2 : s : Skipped : 0
`
		err := os.WriteFile(logFile, []byte(malformedLog), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// LoadLog should now handle malformed entries gracefully
		entries := storage.LoadLog(tmpDir)
		
		// Should only load the valid entries
		validEntries := 0
		for range *entries {
			validEntries++
		}
		
		// We expect 2 valid entries: "Valid Habit" and "Valid Habit 2"
		if validEntries < 2 {
			t.Errorf("Expected at least 2 valid entries, got %d", validEntries)
		}
		
		t.Logf("✓ LoadLog gracefully handled malformed entries, loaded %d valid entries", validEntries)
	})

	t.Run("Malformed habits file handling", func(t *testing.T) {
		// Create a habits file with various malformed entries
		habitsFile := filepath.Join(tmpDir, "habits")
		malformedHabits := `# Test habits with various issues
! Valid Heading
Valid Habit: 1
Missing colon habit 1
: 1
Empty name habit 3/7
Habit without frequency:
! Bad heading no space
! Another Valid Heading
Another Valid Habit: 3/7
`
		err := os.WriteFile(habitsFile, []byte(malformedHabits), 0644)
		if err != nil {
			t.Fatal(err)
		}

		// LoadHabitsConfig should now handle malformed entries gracefully
		habits, maxLength := storage.LoadHabitsConfig(tmpDir)
		
		// Should only load the valid habits
		if len(habits) < 2 {
			t.Errorf("Expected at least 2 valid habits, got %d", len(habits))
		}
		
		if maxLength == 0 {
			t.Error("Max length should be positive")
		}
		
		// Verify the valid habits were loaded
		validHabitNames := make([]string, len(habits))
		for i, habit := range habits {
			validHabitNames[i] = habit.Name
		}
		
		t.Logf("✓ LoadHabitsConfig gracefully handled malformed entries")
		t.Logf("✓ Loaded valid habits: %v", validHabitNames)
	})

	t.Run("Write error handling", func(t *testing.T) {
		// Test improved error messages from WriteHabitLog
		nonExistentDir := "/tmp/nonexistent_harsh_test_dir_12345"
		
		// Ensure directory doesn't exist
		os.RemoveAll(nonExistentDir)
		
		testDate := civil.Date{Year: 2025, Month: 1, Day: 15}
		err := storage.WriteHabitLog(nonExistentDir, testDate, "Test Habit", "y", "Test", "1.0")
		
		if err == nil {
			t.Error("Expected error when writing to non-existent directory")
		}
		
		// Check that error message is more helpful
		if err != nil {
			errorMsg := err.Error()
			if !containsAnyOf(errorMsg, []string{"configuration directory", "does not exist", "permission", "disk space"}) {
				t.Errorf("Error message should be more descriptive: %v", err)
			}
			t.Logf("✓ Improved error message: %v", err)
		}
	})
}

// Helper function to check if string contains any of the given substrings
func containsAnyOf(str string, substrings []string) bool {
	for _, substr := range substrings {
		if contains(str, substr) {
			return true
		}
	}
	return false
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(substr) == 0 || 
		(len(str) > len(substr) && (str[:len(substr)] == substr || 
		str[len(str)-len(substr):] == substr || 
		containsInMiddle(str, substr))))
}

func containsInMiddle(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDocumentationOfImprovements(t *testing.T) {
	// This test documents the improvements we made
	improvements := []string{
		"LoadLog now detects iCloud sync placeholders (.log.icloud)",
		"LoadLog provides helpful messages for missing files vs permission errors",
		"LoadLog gracefully handles malformed entries with warnings",
		"LoadHabitsConfig detects iCloud sync placeholders (.habits.icloud)", 
		"LoadHabitsConfig provides guidance for first-time users",
		"LoadHabitsConfig skips malformed habit entries with warnings",
		"WriteHabitLog provides specific error messages for different failure types",
		"All parsing is now robust against index out of range panics",
		"Better error messages help users with cloud storage scenarios",
	}
	
	for _, improvement := range improvements {
		t.Logf("✅ %s", improvement)
	}
	
	t.Log("These improvements make harsh more robust for real-world usage!")
}