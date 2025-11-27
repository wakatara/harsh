package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestOnboardingNotTriggeredForVersionAndHelp verifies that version and help commands
// do not trigger onboarding when config files don't exist.
// This is important for automation scenarios where users just want to check the version.
func TestOnboardingNotTriggeredForVersionAndHelp(t *testing.T) {
	// Build the test binary
	buildCmd := exec.Command("go", "build", "-o", "harsh-test-onboarding", ".")
	buildCmd.Dir = ".."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}

	// Clean up the test binary after all tests
	defer func() {
		cleanCmd := exec.Command("rm", "harsh-test-onboarding")
		cleanCmd.Dir = ".."
		_ = cleanCmd.Run()
	}()

	tests := []struct {
		name              string
		args              []string
		shouldOnboard     bool
		expectedInOutput  []string
		notExpectedOutput []string
		description       string
	}{
		{
			name:              "harsh --version should not onboard",
			args:              []string{"--version"},
			shouldOnboard:     false,
			expectedInOutput:  []string{"harsh version"},
			notExpectedOutput: []string{"Welcome to harsh!", "Created"},
			description:       "Version flag should display version without onboarding",
		},
		{
			name:              "harsh version should not onboard",
			args:              []string{"version"},
			shouldOnboard:     false,
			expectedInOutput:  []string{"harsh version", "go version"},
			notExpectedOutput: []string{"Welcome to harsh!", "Created"},
			description:       "Version subcommand should display version without onboarding",
		},
		{
			name:              "harsh --help should not onboard",
			args:              []string{"--help"},
			shouldOnboard:     false,
			expectedInOutput:  []string{"Usage:", "Available Commands:"},
			notExpectedOutput: []string{"Welcome to harsh!", "Created"},
			description:       "Help flag should display help without onboarding",
		},
		{
			name:              "harsh help should not onboard",
			args:              []string{"help"},
			shouldOnboard:     false,
			expectedInOutput:  []string{"Usage:", "Available Commands:"},
			notExpectedOutput: []string{"Welcome to harsh!", "Created"},
			description:       "Help subcommand should display help without onboarding",
		},
		{
			name:              "harsh -h should not onboard",
			args:              []string{"-h"},
			shouldOnboard:     false,
			expectedInOutput:  []string{"Usage:", "Available Commands:"},
			notExpectedOutput: []string{"Welcome to harsh!", "Created"},
			description:       "Short help flag should display help without onboarding",
		},
		{
			name:              "bare harsh should onboard",
			args:              []string{},
			shouldOnboard:     true,
			expectedInOutput:  []string{"Welcome to harsh!", "Created"},
			notExpectedOutput: []string{},
			description:       "Running harsh without arguments should trigger onboarding for new users",
		},
		{
			name:              "harsh todo should onboard",
			args:              []string{"todo"},
			shouldOnboard:     true,
			expectedInOutput:  []string{"Welcome to harsh!", "Created"},
			notExpectedOutput: []string{},
			description:       "Todo command should trigger onboarding for new users",
		},
		{
			name:              "harsh log should onboard",
			args:              []string{"log"},
			shouldOnboard:     true,
			expectedInOutput:  []string{"Welcome to harsh!", "Created"},
			notExpectedOutput: []string{},
			description:       "Log command should trigger onboarding for new users",
		},
		{
			name:              "harsh ask should onboard",
			args:              []string{"ask"},
			shouldOnboard:     true,
			expectedInOutput:  []string{"Welcome to harsh!", "Created"},
			notExpectedOutput: []string{},
			description:       "Ask command should trigger onboarding for new users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh temporary directory for each test
			tmpDir, err := os.MkdirTemp("", "harsh_onboard_test_*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Use a non-existent subdirectory as HARSHPATH to ensure fresh state
			harshPath := filepath.Join(tmpDir, "config")

			// Run the command with the fresh HARSHPATH
			cmd := exec.Command("./harsh-test-onboarding", tt.args...)
			cmd.Dir = ".."
			cmd.Env = append(os.Environ(), "HARSHPATH="+harshPath)
			output, _ := cmd.CombinedOutput()
			// Note: We ignore the error because onboarding exits with 0 and
			// some commands may exit with non-zero if habits don't exist after onboarding

			outputStr := string(output)

			// Check for expected content
			for _, expected := range tt.expectedInOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected '%s' in output for %s\nGot: %s", expected, tt.description, outputStr)
				}
			}

			// Check for content that should NOT be present
			for _, notExpected := range tt.notExpectedOutput {
				if strings.Contains(outputStr, notExpected) {
					t.Errorf("Did not expect '%s' in output for %s\nGot: %s", notExpected, tt.description, outputStr)
				}
			}

			// Verify onboarding behavior by checking if config files were created
			habitsFile := filepath.Join(harshPath, "habits")
			logFile := filepath.Join(harshPath, "log")

			habitsExists := fileExists(habitsFile)
			logExists := fileExists(logFile)
			filesCreated := habitsExists && logExists

			if tt.shouldOnboard && !filesCreated {
				t.Errorf("Expected config files to be created for %s, but they weren't", tt.description)
			}
			if !tt.shouldOnboard && filesCreated {
				t.Errorf("Did not expect config files to be created for %s, but they were", tt.description)
			}
		})
	}
}

// TestOnboardingCreatesValidFiles verifies that onboarding creates properly formatted files
func TestOnboardingCreatesValidFiles(t *testing.T) {
	// Build the test binary
	buildCmd := exec.Command("go", "build", "-o", "harsh-test-onboarding-files", ".")
	buildCmd.Dir = ".."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}

	defer func() {
		cleanCmd := exec.Command("rm", "harsh-test-onboarding-files")
		cleanCmd.Dir = ".."
		_ = cleanCmd.Run()
	}()

	// Create a fresh temporary directory
	tmpDir, err := os.MkdirTemp("", "harsh_onboard_files_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	harshPath := filepath.Join(tmpDir, "config")

	// Run bare harsh to trigger onboarding
	cmd := exec.Command("./harsh-test-onboarding-files")
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(), "HARSHPATH="+harshPath)
	_, _ = cmd.CombinedOutput()

	// Verify habits file was created with expected content
	habitsFile := filepath.Join(harshPath, "habits")
	if !fileExists(habitsFile) {
		t.Fatal("habits file was not created")
	}

	habitsContent, err := os.ReadFile(habitsFile)
	if err != nil {
		t.Fatalf("Failed to read habits file: %v", err)
	}

	// Check for expected example habits
	expectedHabits := []string{
		"Gymmed:",
		"Bed by midnight:",
		"Cleaned House:",
		"Called Mom:",
	}
	for _, habit := range expectedHabits {
		if !strings.Contains(string(habitsContent), habit) {
			t.Errorf("Expected example habit '%s' in habits file", habit)
		}
	}

	// Verify log file was created (should be empty)
	logFile := filepath.Join(harshPath, "log")
	if !fileExists(logFile) {
		t.Fatal("log file was not created")
	}

	logContent, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(logContent) != 0 {
		t.Errorf("Expected empty log file, got %d bytes", len(logContent))
	}
}

// TestVersionOutputFormat verifies version command output format regardless of onboarding state
func TestVersionOutputFormat(t *testing.T) {
	// Build the test binary with a known version
	buildCmd := exec.Command(
		"go", "build",
		"-ldflags", "-X github.com/wakatara/harsh/cmd.version=1.2.3-test",
		"-o", "harsh-test-version-format",
	)
	buildCmd.Dir = ".."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}

	defer func() {
		cleanCmd := exec.Command("rm", "harsh-test-version-format")
		cleanCmd.Dir = ".."
		_ = cleanCmd.Run()
	}()

	// Create a fresh temporary directory (no config)
	tmpDir, err := os.MkdirTemp("", "harsh_version_format_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	harshPath := filepath.Join(tmpDir, "config")

	tests := []struct {
		name string
		args []string
	}{
		{"--version flag", []string{"--version"}},
		{"version subcommand", []string{"version"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./harsh-test-version-format", tt.args...)
			cmd.Dir = ".."
			cmd.Env = append(os.Environ(), "HARSHPATH="+harshPath)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Version command failed: %v\nOutput: %s", err, output)
			}

			outputStr := string(output)

			// Verify version is shown
			if !strings.Contains(outputStr, "1.2.3-test") {
				t.Errorf("Expected version '1.2.3-test' in output, got: %s", outputStr)
			}

			// Verify no onboarding message
			if strings.Contains(outputStr, "Welcome") {
				t.Errorf("Version command should not show welcome message, got: %s", outputStr)
			}
		})
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
