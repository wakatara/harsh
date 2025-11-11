package test

import (
	"strings"
	"testing"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal/graph"
	"github.com/wakatara/harsh/internal/storage"
)

// TestMonthBoundaryMarker tests that month boundaries are marked correctly with ▕ on the last day of the previous month
func TestMonthBoundaryMarker(t *testing.T) {
	tests := []struct {
		name     string
		from     civil.Date
		to       civil.Date
		wantMark bool // Whether we expect a month boundary marker
		wantPos  int   // Expected position of the marker in calline
	}{
		{
			name:     "January to February boundary",
			from:     civil.Date{Year: 2024, Month: 1, Day: 30},
			to:       civil.Date{Year: 2024, Month: 2, Day: 2},
			wantMark: true,
			wantPos:  1, // After Jan 31 (index 0 is Jan 30, index 1 is Jan 31)
		},
		{
			name:     "February to March boundary (non-leap year)",
			from:     civil.Date{Year: 2023, Month: 2, Day: 27},
			to:       civil.Date{Year: 2023, Month: 3, Day: 2},
			wantMark: true,
			wantPos:  1, // After Feb 28 (index 0 is Feb 27, index 1 is Feb 28)
		},
		{
			name:     "February to March boundary (leap year)",
			from:     civil.Date{Year: 2024, Month: 2, Day: 28},
			to:       civil.Date{Year: 2024, Month: 3, Day: 2},
			wantMark: true,
			wantPos:  1, // After Feb 29 (index 0 is Feb 28, index 1 is Feb 29)
		},
		{
			name:     "No boundary within same month",
			from:     civil.Date{Year: 2024, Month: 1, Day: 15},
			to:       civil.Date{Year: 2024, Month: 1, Day: 20},
			wantMark: false,
			wantPos:  -1,
		},
		{
			name:     "December to January boundary (year change)",
			from:     civil.Date{Year: 2023, Month: 12, Day: 30},
			to:       civil.Date{Year: 2024, Month: 1, Day: 2},
			wantMark: true,
			wantPos:  1, // After Dec 31 (index 0 is Dec 30, index 1 is Dec 31)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal test data
			habits := []*storage.Habit{}
			entries := &storage.Entries{}

			_, calline := graph.BuildSpark(tt.from, tt.to, habits, entries)

			if tt.wantMark {
				// Check if the marker exists at the expected position
				if tt.wantPos >= len(calline) {
					t.Fatalf("Expected position %d is out of range (calline length: %d)", tt.wantPos, len(calline))
				}
				if !strings.Contains(calline[tt.wantPos], "▕") {
					t.Errorf("Expected month boundary marker ▕ at position %d, got: %q", tt.wantPos, calline[tt.wantPos])
				}
			} else {
				// Check that no marker exists
				for i, char := range calline {
					if strings.Contains(char, "▕") {
						t.Errorf("Unexpected month boundary marker at position %d: %q", i, char)
					}
				}
			}
		})
	}
}

// TestMonthBoundaryDoesNotShiftMWF tests that adding boundary markers doesn't shift M W F positions
func TestMonthBoundaryDoesNotShiftMWF(t *testing.T) {
	// Test a date range that includes a month boundary
	// Jan 29 (Mon), Jan 30 (Tue), Jan 31 (Wed), Feb 1 (Thu), Feb 2 (Fri)
	from := civil.Date{Year: 2024, Month: 1, Day: 29}
	to := civil.Date{Year: 2024, Month: 2, Day: 2}

	habits := []*storage.Habit{}
	entries := &storage.Entries{}

	_, calline := graph.BuildSpark(from, to, habits, entries)

	// Expected: 5 elements total (one per day)
	if len(calline) != 5 {
		t.Fatalf("Expected 5 elements in calline, got %d", len(calline))
	}

	// Jan 29 (Monday) = "M"
	// Jan 30 (Tuesday) = " "
	// Jan 31 (Wednesday) = "W" + "▕" (boundary marker)
	// Feb 1 (Thursday) = " "
	// Feb 2 (Friday) = "F"

	expected := []string{"M", " ", "W▕", " ", "F"}
	for i, want := range expected {
		if calline[i] != want {
			t.Errorf("Position %d: expected %q, got %q", i, want, calline[i])
		}
	}
}

// TestMonthBoundaryLeapYear tests that leap year boundaries are handled correctly
func TestMonthBoundaryLeapYear(t *testing.T) {
	tests := []struct {
		name     string
		from     civil.Date
		to       civil.Date
		wantDays int
	}{
		{
			name:     "Leap year Feb has 29 days",
			from:     civil.Date{Year: 2024, Month: 2, Day: 27},
			to:       civil.Date{Year: 2024, Month: 3, Day: 2},
			wantDays: 5, // Feb 27, 28, 29, Mar 1, Mar 2
		},
		{
			name:     "Non-leap year Feb has 28 days",
			from:     civil.Date{Year: 2023, Month: 2, Day: 27},
			to:       civil.Date{Year: 2023, Month: 3, Day: 2},
			wantDays: 4, // Feb 27, 28, Mar 1, Mar 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			habits := []*storage.Habit{}
			entries := &storage.Entries{}

			_, calline := graph.BuildSpark(tt.from, tt.to, habits, entries)

			if len(calline) != tt.wantDays {
				t.Errorf("Expected %d days, got %d", tt.wantDays, len(calline))
			}

			// Verify the marker appears exactly once on the last day of February
			markerCount := 0
			markerPos := -1
			for i, char := range calline {
				if strings.Contains(char, "▕") {
					markerCount++
					markerPos = i
				}
			}

			if markerCount != 1 {
				t.Errorf("Expected exactly 1 month boundary marker, got %d", markerCount)
			}

			// The marker should be on the day before the end (last Feb day)
			// For leap year: index 2 (Feb 29)
			// For non-leap: index 1 (Feb 28)
			var expectedPos int
			if tt.wantDays == 5 {
				expectedPos = 2 // Leap year
			} else {
				expectedPos = 1 // Non-leap year
			}

			if markerPos != expectedPos {
				t.Errorf("Expected marker at position %d, got %d", expectedPos, markerPos)
			}
		})
	}
}

// TestMultipleMonthBoundaries tests that multiple month boundaries work correctly
func TestMultipleMonthBoundaries(t *testing.T) {
	// Span three months: Jan 30 to Mar 2
	from := civil.Date{Year: 2024, Month: 1, Day: 30}
	to := civil.Date{Year: 2024, Month: 3, Day: 2}

	habits := []*storage.Habit{}
	entries := &storage.Entries{}

	_, calline := graph.BuildSpark(from, to, habits, entries)

	// Count markers
	markerCount := 0
	for _, char := range calline {
		if strings.Contains(char, "▕") {
			markerCount++
		}
	}

	// Should have 2 markers: Jan→Feb and Feb→Mar
	if markerCount != 2 {
		t.Errorf("Expected 2 month boundary markers, got %d", markerCount)
	}
}
