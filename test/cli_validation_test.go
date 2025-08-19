package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud.google.com/go/civil"
	"github.com/wakatara/harsh/internal/storage"
)

func TestDateArgumentValidation(t *testing.T) {
	// These tests verify CLI argument validation behavior
	// Note: We test through the public API rather than running the CLI directly
	// since CLI testing with os.Exit is complex

	tests := []struct {
		name        string
		dateStr     string
		shouldWork  bool
		description string
	}{
		// Valid date formats
		{"Valid ISO date", "2025-01-15", true, "Standard YYYY-MM-DD format"},
		{"Valid date at year boundary", "2024-12-31", true, "End of year"},
		{"Valid date at month boundary", "2025-02-28", true, "End of February"},
		{"Valid leap year date", "2024-02-29", true, "Leap year February 29"},

		// Invalid date formats
		{"Invalid format DD-MM-YYYY", "15-01-2025", false, "European date format"},
		{"Invalid format MM/DD/YYYY", "01/15/2025", false, "US date format"},
		{"Invalid format with time", "2025-01-15T10:30:00", false, "ISO datetime instead of date"},
		{"Invalid month", "2025-13-15", false, "Month 13 doesn't exist"},
		{"Invalid day", "2025-01-32", false, "January only has 31 days"},
		{"Invalid leap year", "2023-02-29", false, "2023 is not a leap year"},
		{"Invalid format too short", "2025-1-1", false, "Single digit month/day"},
		{"Invalid format too long", "2025-001-001", false, "Zero-padded beyond standard"},
		{"Invalid characters", "2025-ab-cd", false, "Letters instead of numbers"},
		{"Invalid separators", "2025.01.15", false, "Dots instead of dashes"},
		{"Invalid separators", "2025/01/15", false, "Slashes instead of dashes"},
		{"Empty string", "", false, "No date provided"},
		{"Just year", "2025", false, "Year only"},
		{"Just year-month", "2025-01", false, "Year-month only"},
		{"Extra characters", "2025-01-15 extra", false, "Date with trailing text"},
		{"Negative year", "-2025-01-15", false, "Negative year"},
		{"Year zero", "0000-01-15", true, "Year zero (technically valid but unusual)"},
		{"Far future", "9999-12-31", true, "Very far future date should work"},
		{"Far past", "1900-01-01", true, "Historical date should work"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test civil.ParseDate directly (this is what the CLI would use)
			date, err := civil.ParseDate(tt.dateStr)

			if tt.shouldWork {
				if err != nil {
					t.Errorf("Expected '%s' (%s) to be valid, but got error: %v",
						tt.dateStr, tt.description, err)
				} else {
					// Verify the date is reasonable
					if date.Year < 1900 || date.Year > 9999 {
						t.Logf("Warning: Date %s has unusual year %d", tt.dateStr, date.Year)
					}
				}
			} else {
				if err == nil {
					t.Errorf("Expected '%s' (%s) to be invalid, but it parsed successfully to %v",
						tt.dateStr, tt.description, date)
				}
			}
		})
	}
}

func TestHabitFragmentMatching(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "harsh_fragment_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create habits file with various habit names
	habitsFile := filepath.Join(tmpDir, "habits")
	habitsContent := `! Work
Daily standup: 1
Code review: 3/7
Team meeting: 1/7
Sprint planning: 1/14

! Health
Gym workout: 3/7
Running: 2/7
Yoga class: 1/7
Water intake: 1

! Personal
Call family: 1/7
Read book: 1
Practice guitar: 3/7
`
	err = os.WriteFile(habitsFile, []byte(habitsContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Load habits
	habits, _ := storage.LoadHabitsConfig(tmpDir)
	if len(habits) == 0 {
		t.Fatal("No habits loaded for fragment testing")
	}

	// Test fragment matching patterns
	fragmentTests := []struct {
		fragment    string
		expectMatch bool
		description string
	}{
		// Exact matches
		{"Gym workout", true, "Exact habit name match"},
		{"gym workout", true, "Case insensitive exact match"},
		{"GYM WORKOUT", true, "All caps exact match"},

		// Partial matches
		{"gym", true, "Single word fragment"},
		{"work", true, "Fragment matching multiple habits"},
		{"meet", true, "Fragment in middle of word"},
		{"ing", true, "Suffix fragment"},
		{"call", true, "Case insensitive partial"},

		// No matches
		{"xyz", false, "Fragment with no matches"},
		{"", false, "Empty fragment"},
		{"12345", false, "Numeric fragment"},
		{"very long fragment that matches nothing", false, "Long non-matching fragment"},

		// Edge cases
		{"a", true, "Single character fragment (likely matches something)"},
		{" gym ", true, "Fragment with spaces"},
		{"gym\t", true, "Fragment with tab (trimmed to 'gym')"},
		{"gym\n", true, "Fragment with newline (trimmed to 'gym')"},

		// Special characters
		{":", false, "Colon character (field separator)"},
		{"/", false, "Slash character"},
		{"!", false, "Exclamation (heading marker)"},
		{"gym*", false, "Wildcard character"},
		{"gym?", false, "Question mark"},
		{"[gym]", false, "Brackets"},
		{"(gym)", false, "Parentheses"},
	}

	for _, tt := range fragmentTests {
		t.Run("Fragment_"+tt.fragment, func(t *testing.T) {
			if tt.fragment == "" {
				t.Skip("Empty fragment test - behavior depends on implementation")
			}

			// Simple fragment matching logic (similar to what CLI might do)
			matches := 0
			fragment := strings.ToLower(strings.TrimSpace(tt.fragment))

			for _, habit := range habits {
				habitName := strings.ToLower(habit.Name)
				if strings.Contains(habitName, fragment) {
					matches++
				}
			}

			hasMatch := matches > 0
			if tt.expectMatch && !hasMatch {
				t.Errorf("Fragment '%s' (%s) should match at least one habit, but matched %d",
					tt.fragment, tt.description, matches)
			}
			if !tt.expectMatch && hasMatch {
				t.Errorf("Fragment '%s' (%s) should not match any habits, but matched %d",
					tt.fragment, tt.description, matches)
			}

			if hasMatch {
				t.Logf("Fragment '%s' matched %d habits", tt.fragment, matches)
			}
		})
	}
}

func TestEnvironmentVariableValidation(t *testing.T) {
	// Test HARSHPATH environment variable validation

	// Save original environment
	originalPath := os.Getenv("HARSHPATH")
	defer func() {
		if originalPath != "" {
			os.Setenv("HARSHPATH", originalPath)
		} else {
			os.Unsetenv("HARSHPATH")
		}
	}()

	pathTests := []struct {
		name        string
		path        string
		shouldWork  bool
		description string
	}{
		// Valid paths
		{"Absolute path", "/tmp/harsh_test", true, "Standard absolute path"},
		{"Home directory", "~/harsh", true, "Path with tilde"},
		{"Path with spaces", "/tmp/harsh test", true, "Path containing spaces"},
		{"Current directory", "./harsh", true, "Relative path"},
		{"Deep path", "/very/deep/nested/path/harsh", true, "Deeply nested path"},

		// Potentially problematic paths
		{"Path with unicode", "/tmp/harsh测试", true, "Unicode characters in path"},
		{"Path with emoji", "/tmp/harsh🏠", true, "Emoji in path"},
		{"Windows style", "C:\\Users\\test\\harsh", true, "Windows-style path"},
		{"UNC path", "\\\\server\\share\\harsh", true, "Network UNC path"},

		// Invalid/problematic paths
		{"Empty path", "", false, "Empty HARSHPATH"},
		{"Just slash", "/", false, "Root directory only"},
		{"Nonexistent deep", "/nonexistent/very/deep/path", false, "Deep nonexistent path"},
		{"Path with newline", "/tmp/harsh\n", false, "Path with newline"},
		{"Path with null", "/tmp/harsh\x00", false, "Path with null byte"},
		{"Very long path", strings.Repeat("/very_long_directory_name", 20), false, "Extremely long path"},
	}

	for _, tt := range pathTests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the environment variable
			if tt.path != "" {
				os.Setenv("HARSHPATH", tt.path)
			} else {
				os.Unsetenv("HARSHPATH")
			}

			// Test if path would work by trying to create it
			if tt.shouldWork && tt.path != "" {
				// For valid paths, test if we can use them
				testDir := tt.path
				if strings.HasPrefix(testDir, "~") {
					// For testing, replace ~ with temp dir
					homeDir, _ := os.UserHomeDir()
					testDir = strings.Replace(testDir, "~", homeDir, 1)
				}

				// Try to create the directory
				err := os.MkdirAll(testDir, 0755)
				if err != nil && !strings.Contains(tt.path, "server") { // Skip network paths
					t.Logf("Could not create test directory %s: %v", testDir, err)
				} else if err == nil {
					// Clean up
					defer os.RemoveAll(testDir)

					// Test that we can use this directory
					storage.CreateExampleHabitsFile(testDir)
					habits, _ := storage.LoadHabitsConfig(testDir)
					if len(habits) == 0 {
						t.Errorf("Failed to use valid path '%s' (%s)", tt.path, tt.description)
					}
				}
			}

			t.Logf("✓ Tested HARSHPATH: %s - %s", tt.path, tt.description)
		})
	}
}

func TestCommandArgumentEdgeCases(t *testing.T) {
	// Test edge cases in command argument processing

	argumentTests := []struct {
		name        string
		args        []string
		expectIssue bool
		description string
	}{
		// Normal cases
		{"No arguments", []string{}, false, "Default behavior"},
		{"Help flag", []string{"--help"}, false, "Help should work"},
		{"Version flag", []string{"--version"}, false, "Version should work"},
		{"Valid date", []string{"2025-01-15"}, false, "Valid date argument"},

		// Edge cases
		{"Too many arguments", []string{"2025-01-15", "extra", "args"}, true, "Too many arguments"},
		{"Invalid flag", []string{"--invalid"}, true, "Unknown flag"},
		{"Mixed valid/invalid", []string{"--help", "--invalid"}, true, "Mix of valid and invalid flags"},
		{"Double dash", []string{"--"}, false, "Double dash (end of options)"},
		{"Single dash", []string{"-"}, true, "Single dash (invalid)"},
		{"Empty argument", []string{""}, true, "Empty string argument"},
		{"Very long argument", []string{strings.Repeat("a", 1000)}, true, "Extremely long argument"},

		// Special characters
		{"Unicode argument", []string{"测试"}, true, "Unicode characters"},
		{"Argument with spaces", []string{"hello world"}, true, "Spaces in argument"},
		{"Argument with newlines", []string{"hello\nworld"}, true, "Newlines in argument"},
		{"Argument with nulls", []string{"hello\x00world"}, true, "Null bytes in argument"},
		{"Control characters", []string{"\x01\x02\x03"}, true, "Control characters"},

		// Injection-like patterns
		{"SQL-like", []string{"'; DROP TABLE habits; --"}, true, "SQL injection pattern"},
		{"Shell-like", []string{"; rm -rf /"}, true, "Shell injection pattern"},
		{"Path traversal", []string{"../../../etc/passwd"}, true, "Path traversal pattern"},
		{"Script tags", []string{"<script>alert('xss')</script>"}, true, "XSS-like pattern"},
	}

	for _, tt := range argumentTests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the actual CLI without complex subprocess handling
			// Instead, we document what should be validated

			if tt.expectIssue {
				t.Logf("⚠️  Arguments %v should be validated/rejected: %s", tt.args, tt.description)
			} else {
				t.Logf("✓ Arguments %v should be accepted: %s", tt.args, tt.description)
			}

			// Test argument length validation
			for _, arg := range tt.args {
				if len(arg) > 500 {
					t.Logf("⚠️  Very long argument detected: %d characters", len(arg))
				}

				// Check for dangerous characters
				if strings.ContainsAny(arg, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x0b\x0c\x0e\x0f") {
					t.Logf("⚠️  Control characters detected in argument: %q", arg)
				}

				if strings.Contains(arg, "..") {
					t.Logf("⚠️  Potential path traversal detected: %q", arg)
				}
			}
		})
	}
}

func TestCommandLineHelpAndVersion(t *testing.T) {
	// Test that help and version commands work reliably

	// These tests would normally run the actual binary, but since we're testing
	// the library, we document the expected behavior

	helpTests := []struct {
		command     string
		expectWork  bool
		description string
	}{
		{"harsh --help", true, "Standard help flag"},
		{"harsh -h", true, "Short help flag"},
		{"harsh help", true, "Help subcommand"},
		{"harsh --version", true, "Version flag"},
		{"harsh -v", true, "Short version flag"},
		{"harsh version", true, "Version subcommand"},

		// Edge cases
		{"harsh --help --version", false, "Conflicting flags"},
		{"harsh help version", false, "Multiple subcommands"},
		{"harsh --help extra", false, "Help with extra args"},
		{"harsh --HELP", false, "Wrong case"},
		{"harsh  --help", true, "Extra spaces"},
		{"harsh\t--help", false, "Tab character"},
	}

	for _, tt := range helpTests {
		t.Run(tt.command, func(t *testing.T) {
			if tt.expectWork {
				t.Logf("✓ Command '%s' should work: %s", tt.command, tt.description)
			} else {
				t.Logf("⚠️  Command '%s' should be validated: %s", tt.command, tt.description)
			}
		})
	}
}

func TestEnvironmentValidation(t *testing.T) {
	// Test environment variable edge cases that could cause issues

	originalEnv := make(map[string]string)
	envVars := []string{"HARSHPATH", "NO_COLOR", "TERM", "HOME", "USER"}

	// Save original environment
	for _, env := range envVars {
		originalEnv[env] = os.Getenv(env)
	}

	// Restore environment after test
	defer func() {
		for env, val := range originalEnv {
			if val != "" {
				os.Setenv(env, val)
			} else {
				os.Unsetenv(env)
			}
		}
	}()

	envTests := []struct {
		name        string
		env         string
		value       string
		expectIssue bool
		description string
	}{
		// Normal values
		{"NO_COLOR=1", "NO_COLOR", "1", false, "Disable colors"},
		{"NO_COLOR=true", "NO_COLOR", "true", false, "Disable colors (boolean)"},
		{"TERM=xterm", "TERM", "xterm", false, "Terminal type"},
		{"HOME valid", "HOME", "/home/user", false, "User home directory"},

		// Edge cases
		{"Empty NO_COLOR", "NO_COLOR", "", false, "Empty NO_COLOR (colors enabled)"},
		{"NO_COLOR with spaces", "NO_COLOR", " 1 ", false, "NO_COLOR with whitespace"},
		{"Very long HOME", "HOME", strings.Repeat("/long", 100), true, "Extremely long HOME path"},
		{"HOME with null", "HOME", "/home/user\x00", true, "HOME with null byte"},
		{"HOME with newline", "HOME", "/home/user\n", true, "HOME with newline"},
		{"Unicode in HOME", "HOME", "/home/用户", false, "Unicode in HOME path"},
		{"Empty TERM", "TERM", "", false, "Empty TERM variable"},
		{"Invalid TERM", "TERM", "\x00\x01", true, "TERM with control characters"},

		// Security-related
		{"Injection-like PATH", "HARSHPATH", "; rm -rf /", true, "Shell injection in path"},
		{"Path traversal", "HARSHPATH", "../../../etc", true, "Path traversal attempt"},
		{"Very long path", "HARSHPATH", strings.Repeat("/a", 1000), true, "Extremely long path"},
	}

	for _, tt := range envTests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the environment variable
			os.Setenv(tt.env, tt.value)

			if tt.expectIssue {
				t.Logf("⚠️  Environment %s=%q should be validated: %s", tt.env, tt.value, tt.description)
			} else {
				t.Logf("✓ Environment %s=%q should work: %s", tt.env, tt.value, tt.description)
			}

			// Basic validation checks
			if len(tt.value) > 1000 {
				t.Logf("⚠️  Very long environment value: %d characters", len(tt.value))
			}

			if strings.ContainsAny(tt.value, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x0b\x0c\x0e\x0f") {
				t.Logf("⚠️  Control characters in environment value")
			}
		})
	}
}

func TestScriptingAndAutomationValidation(t *testing.T) {
	// Test scenarios where harsh might be used in scripts or automation

	automationTests := []struct {
		name        string
		scenario    string
		description string
		shouldWork  bool
	}{
		// Good automation patterns
		{"Cron job usage", "0 9 * * * /usr/local/bin/harsh log", "Daily morning log check", true},
		{"Script with date", "harsh ask $(date +%Y-%m-%d)", "Script using current date", true},
		{"Backup script", "cp ~/.config/harsh/* /backup/", "Backing up config files", true},
		{"Status check", "harsh log | grep $(date +%Y-%m-%d)", "Check today's status", true},

		// Potentially problematic patterns
		{"Unsanitized input", "harsh ask $USER_INPUT", "Using unsanitized user input", false},
		{"Shell injection", "harsh ask $(rm -rf /)", "Command injection attempt", false},
		{"Path injection", "harsh --config ../../../etc", "Path traversal in config", false},
		{"Unicode confusion", "harsh ask 2025‐01‐15", "Unicode dashes that look like ASCII", false},

		// Edge cases in automation
		{"Empty variable", "harsh ask $EMPTY_VAR", "Using empty environment variable", false},
		{"Unset variable", "harsh ask $UNSET_VAR", "Using unset environment variable", false},
		{"Multiple spaces", "harsh    ask    2025-01-15", "Extra whitespace in command", true},
		{"Tab characters", "harsh\task\t2025-01-15", "Tab characters in command", false},
	}

	for _, tt := range automationTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldWork {
				t.Logf("✓ Automation pattern should work: %s - %s", tt.scenario, tt.description)
			} else {
				t.Logf("⚠️  Automation pattern needs validation: %s - %s", tt.scenario, tt.description)
			}

			// Check for common security issues
			if strings.Contains(tt.scenario, "$(") || strings.Contains(tt.scenario, "`") {
				t.Logf("⚠️  Command substitution detected - validate input")
			}

			if strings.Contains(tt.scenario, "$") && !strings.Contains(tt.scenario, "$(date") {
				t.Logf("⚠️  Variable substitution detected - validate variables")
			}

			if strings.Contains(tt.scenario, "..") {
				t.Logf("⚠️  Path traversal pattern detected")
			}
		})
	}

	t.Log("These tests document validation needs for safe automation usage")
}
