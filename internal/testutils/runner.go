package testutils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TestRunner provides programmatic test execution capabilities
type TestRunner struct {
	CoverageThreshold float64
	TestTimeout       time.Duration
	Verbose           bool
	RaceDetection     bool
}

// NewTestRunner creates a new test runner with default configuration
func NewTestRunner() *TestRunner {
	return &TestRunner{
		CoverageThreshold: 70.0,
		TestTimeout:       30 * time.Second,
		Verbose:           true,
		RaceDetection:     true,
	}
}

// RunTests executes all tests in the project
func (tr *TestRunner) RunTests() error {
	args := []string{"test"}

	if tr.Verbose {
		args = append(args, "-v")
	}

	if tr.RaceDetection {
		args = append(args, "-race")
	}

	args = append(args, "./...")

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// RunTestsWithCoverage executes tests with coverage reporting
func (tr *TestRunner) RunTestsWithCoverage() error {
	// Run tests with coverage
	args := []string{"test", "-v", "-race", "-coverprofile=coverage.out", "-covermode=atomic", "./..."}

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	// Generate HTML coverage report
	htmlCmd := exec.Command("go", "tool", "cover", "-html=coverage.out", "-o=coverage.html")
	if err := htmlCmd.Run(); err != nil {
		return fmt.Errorf("failed to generate HTML coverage: %w", err)
	}

	// Generate function coverage report
	funcCmd := exec.Command("go", "tool", "cover", "-func=coverage.out")
	funcOutput, err := funcCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to generate function coverage: %w", err)
	}

	// Write function coverage to file
	if err := os.WriteFile("coverage-func.txt", funcOutput, 0644); err != nil {
		return fmt.Errorf("failed to write function coverage: %w", err)
	}

	return nil
}

// AnalyzeCoverage analyzes the coverage report and returns coverage information
func (tr *TestRunner) AnalyzeCoverage() (*CoverageInfo, error) {
	// Check if coverage file exists
	if _, err := os.Stat("coverage.out"); os.IsNotExist(err) {
		return nil, fmt.Errorf("coverage.out not found, run tests with coverage first")
	}

	// Get function coverage
	cmd := exec.Command("go", "tool", "cover", "-func=coverage.out")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze coverage: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var totalCoverage float64
	var coveredLines, totalLines int

	for _, line := range lines {
		if strings.Contains(line, "total:") {
			// Parse total coverage percentage
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				fmt.Sscanf(parts[2], "%f%%", &totalCoverage)
			}
		}
	}

	// Parse package coverage
	packages := make(map[string]float64)
	for _, line := range lines {
		if strings.Contains(line, "%") && !strings.Contains(line, "total:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				var pkgCoverage float64
				if _, err := fmt.Sscanf(parts[1], "%f%%", &pkgCoverage); err == nil {
					pkgName := strings.TrimSuffix(parts[0], ".go")
					packages[pkgName] = pkgCoverage
				}
			}
		}
	}

	return &CoverageInfo{
		TotalCoverage:   totalCoverage,
		CoveredLines:    coveredLines,
		TotalLines:      totalLines,
		PackageCoverage: packages,
		MeetsThreshold:  totalCoverage >= tr.CoverageThreshold,
		Threshold:       tr.CoverageThreshold,
	}, nil
}

// CoverageInfo holds coverage analysis results
type CoverageInfo struct {
	TotalCoverage   float64
	CoveredLines    int
	TotalLines      int
	PackageCoverage map[string]float64
	MeetsThreshold  bool
	Threshold       float64
}

// PrintCoverageReport prints a formatted coverage report
func (ci *CoverageInfo) PrintCoverageReport() {
	fmt.Printf("📊 Coverage Report\n")
	fmt.Printf("==================\n")
	fmt.Printf("Total Coverage: %.2f%%\n", ci.TotalCoverage)
	fmt.Printf("Threshold: %.2f%%\n", ci.Threshold)

	if ci.MeetsThreshold {
		fmt.Printf("✅ Coverage threshold met\n")
	} else {
		fmt.Printf("❌ Coverage threshold not met\n")
	}

	fmt.Printf("\nPackage Coverage:\n")
	for pkg, coverage := range ci.PackageCoverage {
		status := "✅"
		if coverage < 50 {
			status = "❌"
		} else if coverage < 70 {
			status = "⚠️"
		}
		fmt.Printf("  %s %s: %.2f%%\n", status, pkg, coverage)
	}

	fmt.Printf("\nReports generated:\n")
	fmt.Printf("  - coverage.out (raw data)\n")
	fmt.Printf("  - coverage.html (HTML report)\n")
	fmt.Printf("  - coverage-func.txt (function report)\n")
}

// RunSpecificTest runs a specific test or test pattern
func (tr *TestRunner) RunSpecificTest(testPattern string) error {
	args := []string{"test", "-v"}

	if tr.RaceDetection {
		args = append(args, "-race")
	}

	args = append(args, "-run", testPattern, "./...")

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// RunPackageTests runs tests for a specific package
func (tr *TestRunner) RunPackageTests(packagePath string) error {
	args := []string{"test", "-v"}

	if tr.RaceDetection {
		args = append(args, "-race")
	}

	args = append(args, packagePath)

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// CleanupCoverageFiles removes coverage report files
func (tr *TestRunner) CleanupCoverageFiles() error {
	files := []string{"coverage.out", "coverage.html", "coverage-func.txt"}

	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", file, err)
		}
	}

	return nil
}

// ValidateTestEnvironment checks if the test environment is properly set up
func (tr *TestRunner) ValidateTestEnvironment() error {
	// Check if go is available
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go command not found: %w", err)
	}

	// Check if go.mod exists
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found, not in a Go module")
	}

	// Try to build the project
	buildCmd := exec.Command("go", "build", "./...")
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("project build failed: %w", err)
	}

	// Try to run a basic test
	testCmd := exec.Command("go", "test", "-v", "-run", "TestHealthCheck", "./internal/handlers")
	if err := testCmd.Run(); err != nil {
		return fmt.Errorf("basic test failed: %w", err)
	}

	return nil
}

// GetTestFiles returns a list of all test files in the project
func (tr *TestRunner) GetTestFiles() ([]string, error) {
	var testFiles []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, "_test.go") {
			testFiles = append(testFiles, path)
		}

		return nil
	})

	return testFiles, err
}

// PrintTestSummary prints a summary of available tests
func (tr *TestRunner) PrintTestSummary() error {
	testFiles, err := tr.GetTestFiles()
	if err != nil {
		return fmt.Errorf("failed to get test files: %w", err)
	}

	fmt.Printf("📋 Test Summary\n")
	fmt.Printf("===============\n")
	fmt.Printf("Found %d test files:\n", len(testFiles))

	for _, file := range testFiles {
		fmt.Printf("  - %s\n", file)
	}

	return nil
}
