package graph

import (
	"runtime"
	"sync"

	"github.com/wakatara/harsh/internal/storage"
)

// HabitGraphResult holds the result of building a graph for a single habit
type HabitGraphResult struct {
	HabitName string
	Graph     string
	Error     error
}

// BuildGraphsParallel builds graphs for multiple habits concurrently
func BuildGraphsParallel(habits []*storage.Habit, entries *storage.Entries, countBack int, ask bool) map[string]string {
	// Determine optimal number of workers
	numWorkers := min(len(habits), runtime.NumCPU())

	// Create channels for work distribution
	habitChan := make(chan *storage.Habit, len(habits))
	resultChan := make(chan HabitGraphResult, len(habits))

	var wg sync.WaitGroup

	// Start worker goroutines
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for habit := range habitChan {
				graph := BuildGraph(habit, entries, countBack, ask)
				resultChan <- HabitGraphResult{
					HabitName: habit.Name,
					Graph:     graph,
					Error:     nil,
				}
			}
		}()
	}

	// Send habits to workers
	go func() {
		for _, habit := range habits {
			habitChan <- habit
		}
		close(habitChan)
	}()

	// Close result channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make(map[string]string, len(habits))
	for result := range resultChan {
		if result.Error == nil {
			results[result.HabitName] = result.Graph
		}
	}

	return results
}