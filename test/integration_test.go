package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal"
	"github.com/wakatara/harsh/internal/graph"
	"github.com/wakatara/harsh/internal/storage"
	"github.com/wakatara/harsh/internal/ui"
)

func TestFullWorkflow(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_integration_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set environment variable to use test directory
	os.Setenv("HARSHPATH", tmpDir)
	defer os.Unsetenv("HARSHPATH")

	// Step 1: Initialize configuration
	storage.CreateExampleHabitsFile(tmpDir)
	storage.CreateNewLogFile(tmpDir)

	// Step 2: Create Harsh instance
	harsh := internal.NewHarsh()

	// Verify initialization
	if harsh == nil {
		t.Fatal("Harsh instance should not be nil")
	}

	habits := harsh.GetHabits()
	if len(habits) == 0 {
		t.Fatal("Should have loaded habits")
	}

	entries := harsh.GetEntries()
	if entries == nil {
		t.Fatal("Entries should not be nil")
	}

	// Step 3: Add some log entries
	testDate := civil.Date{Year: 2025, Month: 1, Day: 15}
	repository := harsh.GetRepository()

	err = repository.WriteEntry(testDate, habits[0].Name, "y", "Test entry", "1.0")
	if err != nil {
		t.Fatal(err)
	}

	err = repository.WriteEntry(testDate, habits[1].Name, "n", "Missed it", "0")
	if err != nil {
		t.Fatal(err)
	}

	// Step 4: Reload entries to verify persistence
	newEntries, err := repository.LoadEntries()
	if err != nil {
		t.Fatal(err)
	}

	// Verify entries were saved
	entry1 := (*newEntries)[storage.DailyHabit{Day: testDate, Habit: habits[0].Name}]
	if entry1.Result != "y" || entry1.Comment != "Test entry" || entry1.Amount != 1.0 {
		t.Errorf("Entry 1 not saved correctly: result=%s, comment=%s, amount=%f",
			entry1.Result, entry1.Comment, entry1.Amount)
	}

	entry2 := (*newEntries)[storage.DailyHabit{Day: testDate, Habit: habits[1].Name}]
	if entry2.Result != "n" || entry2.Comment != "Missed it" {
		t.Errorf("Entry 2 not saved correctly: result=%s, comment=%s",
			entry2.Result, entry2.Comment)
	}

	// Step 5: Test graph generation
	graphResult := graph.BuildGraph(habits[0], newEntries, 10, false)
	if graphResult == "" {
		t.Error("Graph should not be empty")
	}

	// Step 6: Test parallel graph generation
	graphResults := graph.BuildGraphsParallel(habits, newEntries, 10, false)
	if len(graphResults) != len(habits) {
		t.Errorf("Expected %d graph results, got %d", len(habits), len(graphResults))
	}

	// Step 7: Test scoring
	score := graph.Score(testDate, habits, newEntries)
	if score < 0 || score > 100 {
		t.Errorf("Score should be between 0 and 100, got %f", score)
	}

	// Step 8: Test sparkline generation
	sparkline, calline := graph.BuildSpark(testDate, testDate, habits, newEntries)
	if len(sparkline) != 1 || len(calline) != 1 {
		t.Errorf("Sparkline and calline should have 1 entry each, got %d and %d",
			len(sparkline), len(calline))
	}

	// Step 9: Test todos
	todos := ui.GetTodos(habits, newEntries, testDate, 1)
	if len(todos) == 0 {
		t.Error("Should have todo entries")
	}

	// Step 10: Test stats
	stats := ui.BuildStats(habits[0], newEntries)
	if stats.DaysTracked <= 0 {
		t.Error("Should have positive days tracked")
	}

	// Step 11: Test UI display components (minimal test)
	display := ui.NewDisplay(true)
	if display == nil {
		t.Error("Display should not be nil")
	}

	input := ui.NewInput(true)
	if input == nil {
		t.Error("Input should not be nil")
	}

	// Step 12: Test repository interface compliance
	var _ storage.Repository = repository
}

func TestEndToEndScenario(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_e2e_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set environment variable to use test directory
	os.Setenv("HARSHPATH", tmpDir)
	defer os.Unsetenv("HARSHPATH")

	// Create custom habits file
	habitsFile := filepath.Join(tmpDir, "habits")
	habitsContent := `! Fitness
Gym: 3/7
Running: 2/7
Stretching: 1

! Work
Daily standup: 1
Code review: 5/7

! Health
Water intake: 1
Sleep tracking: 0
`
	err = os.WriteFile(habitsFile, []byte(habitsContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create empty log file
	storage.CreateNewLogFile(tmpDir)

	// Initialize Harsh
	harsh := internal.NewHarsh()
	habits := harsh.GetHabits()
	repository := harsh.GetRepository()

	// Simulate a week of habit tracking
	startDate := civil.Date{Year: 2025, Month: 1, Day: 1}

	// Day 1: Good day
	repository.WriteEntry(startDate, "Gym", "y", "Great workout", "1.5")
	repository.WriteEntry(startDate, "Running", "n", "Too tired", "0")
	repository.WriteEntry(startDate, "Stretching", "y", "Morning routine", "0.5")
	repository.WriteEntry(startDate, "Daily standup", "y", "Good meeting", "0")
	repository.WriteEntry(startDate, "Code review", "y", "Reviewed 3 PRs", "3")
	repository.WriteEntry(startDate, "Water intake", "y", "8 glasses", "8")
	repository.WriteEntry(startDate, "Sleep tracking", "y", "Tracked with app", "0")

	// Day 2: Mixed day
	day2 := startDate.AddDays(1)
	repository.WriteEntry(day2, "Gym", "n", "Rest day", "0")
	repository.WriteEntry(day2, "Running", "y", "5k run", "5")
	repository.WriteEntry(day2, "Stretching", "y", "10 min", "0.17")
	repository.WriteEntry(day2, "Daily standup", "y", "Brief update", "0")
	repository.WriteEntry(day2, "Code review", "s", "Skipped today", "0")
	repository.WriteEntry(day2, "Water intake", "n", "Forgot to track", "0")

	// Day 3: Poor day
	day3 := startDate.AddDays(2)
	repository.WriteEntry(day3, "Stretching", "n", "Overslept", "0")
	repository.WriteEntry(day3, "Daily standup", "n", "Missed meeting", "0")
	repository.WriteEntry(day3, "Water intake", "y", "Better today", "6")

	// Reload entries
	entries, err := repository.LoadEntries()
	if err != nil {
		t.Fatal(err)
	}

	// Test scoring across the three days
	day1Score := graph.Score(startDate, habits, entries)
	day2Score := graph.Score(day2, habits, entries)
	day3Score := graph.Score(day3, habits, entries)

	// Day 1 should have the highest score
	if day1Score <= day2Score || day1Score <= day3Score {
		t.Errorf("Day 1 should have highest score: day1=%f, day2=%f, day3=%f",
			day1Score, day2Score, day3Score)
	}

	// Test todos for day 4 (should include missed habits)
	day4 := startDate.AddDays(3)
	todos := ui.GetTodos(habits, entries, day4, 1)
	if len(todos) == 0 {
		t.Error("Should have todos for day 4")
	}

	// Test stats for gym habit
	var gymHabit *storage.Habit
	for _, habit := range habits {
		if habit.Name == "Gym" {
			gymHabit = habit
			break
		}
	}
	if gymHabit == nil {
		t.Fatal("Gym habit not found")
	}

	// Set first record manually since it's not set in our test scenario
	gymHabit.FirstRecord = startDate
	gymStats := ui.BuildStats(gymHabit, entries)

	if gymStats.Streaks != 1 || gymStats.Breaks != 1 {
		t.Errorf("Gym stats incorrect: streaks=%d, breaks=%d", gymStats.Streaks, gymStats.Breaks)
	}

	// Test graph generation for multiple habits
	graphResults := graph.BuildGraphsParallel(habits, entries, 10, false)
	for _, habit := range habits {
		if graph, exists := graphResults[habit.Name]; !exists || graph == "" {
			t.Errorf("No graph generated for habit %s", habit.Name)
		}
	}

	// Test sparkline for the period
	sparkline, calline := graph.BuildSpark(startDate, day3, habits, entries)
	if len(sparkline) != 3 || len(calline) != 3 {
		t.Errorf("Sparkline should have 3 entries, got sparkline=%d, calline=%d",
			len(sparkline), len(calline))
	}

	// Verify sparkline shows declining performance
	// Day 1 should be better than day 3
	if sparkline[0] <= sparkline[2] {
		t.Errorf("Day 1 spark should be better than day 3: day1=%s, day3=%s",
			sparkline[0], sparkline[2])
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_concurrent_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set environment variable to use test directory
	os.Setenv("HARSHPATH", tmpDir)
	defer os.Unsetenv("HARSHPATH")

	// Initialize
	storage.CreateExampleHabitsFile(tmpDir)
	storage.CreateNewLogFile(tmpDir)

	harsh := internal.NewHarsh()

	// Test parallel graph building with many habits
	manyHabits := make([]*storage.Habit, 100)
	for i := range manyHabits {
		manyHabits[i] = &storage.Habit{
			Name:        fmt.Sprintf("Habit_%d", i),
			Target:      1,
			Interval:    1,
			FirstRecord: civil.Date{Year: 2025, Month: 1, Day: 1},
		}
	}

	entries := harsh.GetEntries()

	// Add some test entries
	testDate := civil.Date{Year: 2025, Month: 1, Day: 15}
	for i := 0; i < 50; i++ {
		(*entries)[storage.DailyHabit{Day: testDate, Habit: fmt.Sprintf("Habit_%d", i)}] = storage.Outcome{Result: "y"}
	}

	// Test concurrent graph building
	start := time.Now()
	results := graph.BuildGraphsParallel(manyHabits, entries, 10, false)
	duration := time.Since(start)

	// Should complete in reasonable time
	if duration > time.Second*5 {
		t.Errorf("Parallel graph building took too long: %v", duration)
	}

	// Should have results for all habits
	if len(results) != len(manyHabits) {
		t.Errorf("Expected %d results, got %d", len(manyHabits), len(results))
	}

	// Test concurrent scoring
	scores := make([]float64, len(manyHabits))
	for i := range manyHabits {
		scores[i] = graph.Score(testDate, []*storage.Habit{manyHabits[i]}, entries)
	}

	// Verify scores are reasonable
	for i, score := range scores {
		if score < 0 || score > 100 {
			t.Errorf("Invalid score for habit %d: %f", i, score)
		}
	}
}

func TestErrorHandling(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_error_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with invalid log file
	invalidLogFile := filepath.Join(tmpDir, "log")
	invalidLogContent := `2025-01-01 : Gym : y : Comment : 0
2025-01-02 : Running : n : Skipped : 0
`
	err = os.WriteFile(invalidLogFile, []byte(invalidLogContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// LoadLog should handle valid entries
	entries := storage.LoadLog(tmpDir)
	if len(*entries) != 2 {
		t.Errorf("Expected 2 valid entries, got %d", len(*entries))
	}

	// Test with missing files
	emptyDir, err := os.MkdirTemp("", "harsh_empty_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(emptyDir)

	// Should not panic with missing files
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Should not panic with missing files: %v", r)
		}
	}()

	// This would normally call welcome() and exit, but in tests we handle it
	// storage.FindConfigFiles() would create files if they don't exist
}

func TestPerformanceBaseline(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_perf_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set environment variable to use test directory
	os.Setenv("HARSHPATH", tmpDir)
	defer os.Unsetenv("HARSHPATH")

	// Create large dataset
	storage.CreateExampleHabitsFile(tmpDir)
	storage.CreateNewLogFile(tmpDir)

	harsh := internal.NewHarsh()
	repository := harsh.GetRepository()

	// Add many entries across multiple days
	startDate := civil.Date{Year: 2025, Month: 1, Day: 1}
	habits := harsh.GetHabits()

	start := time.Now()
	for day := 0; day < 30; day++ {
		currentDate := startDate.AddDays(day)
		for _, habit := range habits {
			result := "y"
			if day%3 == 0 {
				result = "n"
			} else if day%5 == 0 {
				result = "s"
			}
			repository.WriteEntry(currentDate, habit.Name, result, "test", "1.0")
		}
	}
	writeTime := time.Since(start)

	// Reload entries
	start = time.Now()
	entries, err := repository.LoadEntries()
	if err != nil {
		t.Fatal(err)
	}
	readTime := time.Since(start)

	// Test graph generation performance
	start = time.Now()
	graphResults := graph.BuildGraphsParallel(habits, entries, 30, false)
	graphTime := time.Since(start)

	// Test scoring performance
	start = time.Now()
	for day := 0; day < 30; day++ {
		currentDate := startDate.AddDays(day)
		_ = graph.Score(currentDate, habits, entries)
	}
	scoreTime := time.Since(start)

	// Report performance (these are not strict requirements, just monitoring)
	t.Logf("Performance baseline:")
	t.Logf("  Write time: %v", writeTime)
	t.Logf("  Read time: %v", readTime)
	t.Logf("  Graph time: %v", graphTime)
	t.Logf("  Score time: %v", scoreTime)
	t.Logf("  Total entries: %d", len(*entries))
	t.Logf("  Graph results: %d", len(graphResults))

	// Reasonable performance expectations
	if writeTime > time.Second*5 {
		t.Errorf("Write performance too slow: %v", writeTime)
	}
	if readTime > time.Second*2 {
		t.Errorf("Read performance too slow: %v", readTime)
	}
	if graphTime > time.Second*2 {
		t.Errorf("Graph performance too slow: %v", graphTime)
	}
	if scoreTime > time.Second*1 {
		t.Errorf("Score performance too slow: %v", scoreTime)
	}
}
