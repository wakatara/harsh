package test

import (
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal"
	"github.com/wakatara/harsh/internal/graph"
	"github.com/wakatara/harsh/internal/storage"
)

// BenchmarkSequentialGraphBuilding benchmarks original sequential approach
func BenchmarkSequentialGraphBuilding(b *testing.B) {
	harsh := createTestHarsh()
	habits := createManyTestHabits(10)

	for b.Loop() {
		consistency := map[string][]string{}
		for _, habit := range habits {
			consistency[habit.Name] = append(consistency[habit.Name], graph.BuildGraph(habit, harsh.GetEntries(), harsh.GetCountBack(), false))
		}
	}
}

// Benchmark ParallelGraphBuilding -- benchmarks new parallel approach
func BenchmarkParallelGraphBuilding(b *testing.B) {
	harsh := createTestHarsh()
	habits := createManyTestHabits(10)

	for b.Loop() {
		_ = graph.BuildGraphsParallel(habits, harsh.GetEntries(), harsh.GetCountBack(), false)
	}
}

// createTestHarsh creates a test Harsh instance with sample data
func createTestHarsh() *internal.Harsh {
	entries := make(storage.Entries)

	// Create some sample entries for the last 100 days
	now := civil.DateOf(time.Now())
	for i := range 100 {
		date := now.AddDays(-i)
		entries[storage.DailyHabit{Day: date, Habit: "Test Habit"}] = storage.Outcome{Result: "y", Amount: 1.0, Comment: ""}
	}

	habits := []*storage.Habit{
		{Name: "Test Habit", Frequency: "1", Target: 1, Interval: 1, FirstRecord: now.AddDays(-100)},
	}

	return &internal.Harsh{
		Habits:             habits,
		MaxHabitNameLength: 20,
		CountBack:          100,
		Entries:            &entries,
	}
}

// createManyTestHabits creates multiple test habits for benchmarking
func createManyTestHabits(count int) []*storage.Habit {
	habits := make([]*storage.Habit, count)
	now := civil.DateOf(time.Now())

	for i := range count {
		habits[i] = &storage.Habit{
			Name:        fmt.Sprintf("Habit %d", i+1),
			Frequency:   "1",
			Target:      1,
			Interval:    1,
			FirstRecord: now.AddDays(-100),
		}
	}

	return habits
}
