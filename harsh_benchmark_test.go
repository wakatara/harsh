package main

import (
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/civil"
)

// BenchmarkSequentialGraphBuilding benchmarks original sequential approach
func BenchmarkSequentialGraphBuilding(b *testing.B) {
	harsh := createTestHarsh()
	habits := createManyTestHabits(10)

	for b.Loop() {
		consistency := map[string][]string{}
		for _, habit := range habits {
			consistency[habit.Name] = append(consistency[habit.Name], harsh.buildGraph(habit, false))
		}
	}
}

// Benchmark ParallelGraphBuilding -- benchmarks new parallel approach
func BenchmarkParallelGraphBuilding(b *testing.B) {
	harsh := createTestHarsh()
	habits := createManyTestHabits(10)

	for b.Loop() {
		_ = harsh.buildGraphsParallel(habits, false)
	}
}

// createTestHarsh creates a test Harsh instance with sample data
func createTestHarsh() *Harsh {
	entries := make(Entries)

	// Create some sample entries for the last 100 days
	now := civil.DateOf(time.Now())
	for i := range 100 {
		date := now.AddDays(-i)
		entries[DailyHabit{Day: date, Habit: "Test Habit"}] = Outcome{Result: "y", Amount: 1.0, Comment: ""}
	}

	habits := []*Habit{
		{Name: "Test Habit", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-100)},
	}

	return &Harsh{
		Habits:             habits,
		MaxHabitNameLength: 20,
		CountBack:          100,
		Entries:            &entries,
	}
}

// createManyTestHabits creates multiple test habits for benchmarking
func createManyTestHabits(count int) []*Habit {
	habits := make([]*Habit, count)
	now := civil.DateOf(time.Now())

	for i := range count {
		habits[i] = &Habit{
			Name:        fmt.Sprintf("Habit %d", i+1),
			Frequency:   "1",
			Target:      1,
			Interval:    1,
			FirstRecord: now.AddDays(-100),
		}
	}

	return habits
}
