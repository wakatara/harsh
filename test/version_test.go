package test

import (
	"os/exec"
	"strings"
	"testing"
)

// TestVersionInjection verifies that the version is properly injected during build
func TestVersionInjection(t *testing.T) {
	tests := []struct {
		name        string
		buildFlags  []string
		expectDev   bool
		description string
	}{
		{
			name:        "Build without ldflags should default to 'dev'",
			buildFlags:  []string{"build", "-o", "harsh-test-noversion"},
			expectDev:   true,
			description: "When built without version injection, should show 'dev'",
		},
		{
			name: "Build with ldflags should inject version",
			buildFlags: []string{
				"build",
				"-ldflags",
				"-X github.com/wakatara/harsh/cmd.version=test-version-1.2.3",
				"-o",
				"harsh-test-withversion",
			},
			expectDev:   false,
			description: "When built with -ldflags, should show injected version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the binary with specified flags
			buildCmd := exec.Command("go", tt.buildFlags...)
			buildCmd.Dir = ".."
			if err := buildCmd.Run(); err != nil {
				t.Fatalf("Failed to build test binary: %v", err)
			}

			// Clean up the test binary after test
			binaryName := tt.buildFlags[len(tt.buildFlags)-1]
			defer func() {
				cleanCmd := exec.Command("rm", binaryName)
				cleanCmd.Dir = ".."
				_ = cleanCmd.Run()
			}()

			// Run version command
			versionCmd := exec.Command("./"+binaryName, "version")
			versionCmd.Dir = ".."
			output, err := versionCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to run version command: %v\nOutput: %s", err, output)
			}

			outputStr := string(output)

			if tt.expectDev {
				if !strings.Contains(outputStr, "dev") {
					t.Errorf("Expected version 'dev' in output, got: %s", outputStr)
				}
			} else {
				if strings.Contains(outputStr, "dev") {
					t.Errorf("Expected custom version (not 'dev') in output, got: %s", outputStr)
				}
				if !strings.Contains(outputStr, "test-version-1.2.3") {
					t.Errorf("Expected 'test-version-1.2.3' in output, got: %s", outputStr)
				}
			}
		})
	}
}

// TestVersionFromGitDescribe verifies git describe integration
func TestVersionFromGitDescribe(t *testing.T) {
	// Get version from git describe
	gitCmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	gitCmd.Dir = ".."
	gitOutput, err := gitCmd.CombinedOutput()
	if err != nil {
		t.Skipf("Skipping git describe test: %v", err)
	}

	gitVersion := strings.TrimSpace(string(gitOutput))
	if gitVersion == "" {
		t.Skip("No git version available, skipping test")
	}

	// Build with git version
	buildCmd := exec.Command(
		"go", "build",
		"-ldflags", "-X github.com/wakatara/harsh/cmd.version="+gitVersion,
		"-o", "harsh-test-git",
	)
	buildCmd.Dir = ".."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build with git version: %v", err)
	}

	// Clean up
	defer func() {
		cleanCmd := exec.Command("rm", "harsh-test-git")
		cleanCmd.Dir = ".."
		_ = cleanCmd.Run()
	}()

	// Run version command
	versionCmd := exec.Command("./harsh-test-git", "version")
	versionCmd.Dir = ".."
	output, err := versionCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run version command: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, gitVersion) {
		t.Errorf("Expected git version '%s' in output, got: %s", gitVersion, outputStr)
	}

	// Verify the version format contains expected patterns
	// Could be: v0.11.5, v0.11.5-dirty, v0.11.5-3-g1234abc, or just commit hash
	if !strings.Contains(outputStr, "harsh version") {
		t.Errorf("Expected 'harsh version' prefix in output, got: %s", outputStr)
	}
}

// TestVersionCommandOutput verifies the version command produces expected format
func TestVersionCommandOutput(t *testing.T) {
	// Build a test binary
	buildCmd := exec.Command(
		"go", "build",
		"-ldflags", "-X github.com/wakatara/harsh/cmd.version=0.99.99-test",
		"-o", "harsh-test-format",
	)
	buildCmd.Dir = ".."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}

	// Clean up
	defer func() {
		cleanCmd := exec.Command("rm", "harsh-test-format")
		cleanCmd.Dir = ".."
		_ = cleanCmd.Run()
	}()

	// Run version command
	versionCmd := exec.Command("./harsh-test-format", "version")
	versionCmd.Dir = ".."
	output, err := versionCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run version command: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)

	// Verify output contains expected components
	expectedComponents := []string{
		"harsh version",
		"0.99.99-test",
		"go version",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(outputStr, component) {
			t.Errorf("Expected '%s' in version output, got: %s", component, outputStr)
		}
	}

	// Verify output format has two lines (version line and go version line)
	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines in version output, got %d: %s", len(lines), outputStr)
	}
}
