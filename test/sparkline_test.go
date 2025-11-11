package test

import (
	"strings"
	"testing"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal/graph"
	"github.com/wakatara/harsh/internal/storage"
)

// TestMonthBoundaryMarker tests that month boundaries are marked correctly
// Rule: Use right-aligned ⎹ on last day of month UNLESS it's M/W/F, then use left-aligned ⎸ on first day of next month
func TestMonthBoundaryMarker(t *testing.T) {
	tests := []struct {
		name         string
		from         civil.Date
		to           civil.Date
		wantMark     bool   // Whether we expect a month boundary marker
		wantPos      int    // Expected position of the marker in calline
		wantMarker   string // Expected marker character (⎹ or ⎸)
		wantBaseChar string // Expected base character at position (for left markers)
	}{
		{
			name:         "January to February boundary (Jan 31 = Wednesday)",
			from:         civil.Date{Year: 2024, Month: 1, Day: 30},
			to:           civil.Date{Year: 2024, Month: 2, Day: 2},
			wantMark:     true,
			wantPos:      2,   // Feb 1 (index 0 is Jan 30, index 1 is Jan 31, index 2 is Feb 1)
			wantMarker:   "⎸", // Left marker because Jan 31 is Wednesday
			wantBaseChar: " ",  // Feb 1 is Thursday (no letter)
		},
		{
			name:         "February to March boundary (Feb 28, 2023 = Tuesday)",
			from:         civil.Date{Year: 2023, Month: 2, Day: 27},
			to:           civil.Date{Year: 2023, Month: 3, Day: 2},
			wantMark:     true,
			wantPos:      1,   // Feb 28 (index 0 is Feb 27, index 1 is Feb 28)
			wantMarker:   "⎹", // Right marker because Feb 28 is Tuesday
			wantBaseChar: " ",  // Feb 28 is Tuesday (no letter)
		},
		{
			name:         "February to March boundary (Feb 29, 2024 = Thursday)",
			from:         civil.Date{Year: 2024, Month: 2, Day: 28},
			to:           civil.Date{Year: 2024, Month: 3, Day: 2},
			wantMark:     true,
			wantPos:      1,   // Feb 29 (index 0 is Feb 28, index 1 is Feb 29)
			wantMarker:   "⎹", // Right marker because Feb 29 is Thursday
			wantBaseChar: " ",  // Feb 29 is Thursday (no letter)
		},
		{
			name:     "No boundary within same month",
			from:     civil.Date{Year: 2024, Month: 1, Day: 15},
			to:       civil.Date{Year: 2024, Month: 1, Day: 20},
			wantMark: false,
			wantPos:  -1,
		},
		{
			name:         "December to January boundary (Dec 31, 2023 = Sunday)",
			from:         civil.Date{Year: 2023, Month: 12, Day: 30},
			to:           civil.Date{Year: 2024, Month: 1, Day: 2},
			wantMark:     true,
			wantPos:      1,   // Dec 31 (index 0 is Dec 30, index 1 is Dec 31)
			wantMarker:   "⎹", // Right marker because Dec 31 is Sunday
			wantBaseChar: " ",  // Dec 31 is Sunday (no letter)
		},
		{
			name:         "September to October boundary (Sep 30, 2024 = Monday)",
			from:         civil.Date{Year: 2024, Month: 9, Day: 29},
			to:           civil.Date{Year: 2024, Month: 10, Day: 2},
			wantMark:     true,
			wantPos:      2,   // Oct 1 (index 0 is Sep 29, index 1 is Sep 30, index 2 is Oct 1)
			wantMarker:   "⎸", // Left marker because Sep 30 is Monday
			wantBaseChar: " ",  // Oct 1 is Tuesday (no letter)
		},
		{
			name:         "August to September boundary (Aug 31, 2024 = Saturday)",
			from:         civil.Date{Year: 2024, Month: 8, Day: 30},
			to:           civil.Date{Year: 2024, Month: 9, Day: 2},
			wantMark:     true,
			wantPos:      1,   // Aug 31 (index 0 is Aug 30, index 1 is Aug 31)
			wantMarker:   "⎹", // Right marker because Aug 31 is Saturday
			wantBaseChar: " ",  // Aug 31 is Saturday (no letter)
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
				if !strings.Contains(calline[tt.wantPos], tt.wantMarker) {
					t.Errorf("Expected month boundary marker %q at position %d, got: %q", tt.wantMarker, tt.wantPos, calline[tt.wantPos])
				}
				// Verify the expected base character is also present
				if !strings.Contains(calline[tt.wantPos], tt.wantBaseChar) {
					t.Errorf("Expected base character %q at position %d, got: %q", tt.wantBaseChar, tt.wantPos, calline[tt.wantPos])
				}
			} else {
				// Check that no marker exists
				for i, char := range calline {
					if strings.Contains(char, "⎹") || strings.Contains(char, "⎸") {
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
	// Jan 31 (Wednesday) = "W"
	// Feb 1 (Thursday) = "⎸ " (left-aligned marker + space, because Jan 31 is Wednesday)
	// Feb 2 (Friday) = "F"

	expected := []string{"M", " ", "W", "⎸ ", "F"}
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

			// Verify the marker appears exactly once on the last day of February or first of March
			markerCount := 0
			markerPos := -1
			for i, char := range calline {
				if strings.Contains(char, "⎹") || strings.Contains(char, "⎸") {
					markerCount++
					markerPos = i
				}
			}

			if markerCount != 1 {
				t.Errorf("Expected exactly 1 month boundary marker, got %d", markerCount)
			}

			// The marker position depends on whether the last day of Feb is M/W/F
			// Feb 29, 2024 = Thursday → right marker on Feb 29 (index 2)
			// Feb 28, 2023 = Tuesday → right marker on Feb 28 (index 1)
			var expectedPos int
			if tt.wantDays == 5 {
				expectedPos = 2 // Leap year: Feb 29 (Thursday) gets right marker
			} else {
				expectedPos = 1 // Non-leap year: Feb 28 (Tuesday) gets right marker
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

	// Count markers (both left and right)
	markerCount := 0
	for _, char := range calline {
		if strings.Contains(char, "⎹") || strings.Contains(char, "⎸") {
			markerCount++
		}
	}

	// Should have 2 markers: Jan→Feb and Feb→Mar
	if markerCount != 2 {
		t.Errorf("Expected 2 month boundary markers, got %d", markerCount)
	}
}

// Test2025Boundaries tests the specific month boundaries that occur in 2025
func Test2025Boundaries(t *testing.T) {
	tests := []struct {
		name         string
		from         civil.Date
		to           civil.Date
		wantMarker   string
		wantPos      int
		description  string
	}{
		{
			name:         "Aug 31 to Sep 1, 2025 (Sunday to Monday)",
			from:         civil.Date{Year: 2025, Month: 8, Day: 30},
			to:           civil.Date{Year: 2025, Month: 9, Day: 2},
			wantMarker:   "⎹", // Right marker on Aug 31 (Sunday)
			wantPos:      1,   // Aug 31 position
			description:  "Aug 31 is Sunday, should get right marker",
		},
		{
			name:         "Sep 30 to Oct 1, 2025 (Tuesday to Wednesday)",
			from:         civil.Date{Year: 2025, Month: 9, Day: 29},
			to:           civil.Date{Year: 2025, Month: 10, Day: 2},
			wantMarker:   "⎹", // Right marker on Sep 30 (Tuesday)
			wantPos:      1,   // Sep 30 position
			description:  "Sep 30 is Tuesday, should get right marker",
		},
		{
			name:         "Oct 31 to Nov 1, 2025 (Friday to Saturday)",
			from:         civil.Date{Year: 2025, Month: 10, Day: 30},
			to:           civil.Date{Year: 2025, Month: 11, Day: 2},
			wantMarker:   "⎸", // Left marker on Nov 1 (Saturday)
			wantPos:      2,   // Nov 1 position (index 0=Oct 30, 1=Oct 31, 2=Nov 1)
			description:  "Oct 31 is Friday, should get left marker on Nov 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			habits := []*storage.Habit{}
			entries := &storage.Entries{}

			_, calline := graph.BuildSpark(tt.from, tt.to, habits, entries)

			if !strings.Contains(calline[tt.wantPos], tt.wantMarker) {
				t.Errorf("%s: Expected marker %q at position %d, got: %q",
					tt.description, tt.wantMarker, tt.wantPos, calline[tt.wantPos])
			}
		})
	}
}
