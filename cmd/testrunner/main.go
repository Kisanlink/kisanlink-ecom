package main

import (
	"flag"
	"fmt"
	"log"

	"kisanlink-ecom/internal/testutils"
)

func main() {
	// Define command line flags
	var (
		coverage     = flag.Bool("coverage", false, "Run tests with coverage reporting")
		analyze      = flag.Bool("analyze", false, "Analyze existing coverage report")
		threshold    = flag.Float64("threshold", 70.0, "Coverage threshold percentage")
		verbose      = flag.Bool("verbose", true, "Verbose output")
		race         = flag.Bool("race", true, "Enable race detection")
		specificTest = flag.String("run", "", "Run specific test pattern")
		packagePath  = flag.String("package", "", "Run tests for specific package")
		validate     = flag.Bool("validate", false, "Validate test environment")
		summary      = flag.Bool("summary", false, "Show test summary")
		cleanup      = flag.Bool("cleanup", false, "Clean up coverage files")
		help         = flag.Bool("help", false, "Show help")
	)

	flag.Parse()

	if *help {
		printHelp()
		return
	}

	// Create test runner
	runner := testutils.NewTestRunner()
	runner.CoverageThreshold = *threshold
	runner.Verbose = *verbose
	runner.RaceDetection = *race

	// Handle different commands
	switch {
	case *validate:
		if err := runner.ValidateTestEnvironment(); err != nil {
			log.Fatalf("❌ Test environment validation failed: %v", err)
		}
		fmt.Println("✅ Test environment validated successfully")

	case *summary:
		if err := runner.PrintTestSummary(); err != nil {
			log.Fatalf("❌ Failed to print test summary: %v", err)
		}

	case *cleanup:
		if err := runner.CleanupCoverageFiles(); err != nil {
			log.Fatalf("❌ Failed to cleanup coverage files: %v", err)
		}
		fmt.Println("✅ Coverage files cleaned up")

	case *analyze:
		coverageInfo, err := runner.AnalyzeCoverage()
		if err != nil {
			log.Fatalf("❌ Failed to analyze coverage: %v", err)
		}
		coverageInfo.PrintCoverageReport()

	case *coverage:
		fmt.Println("🧪 Running tests with coverage...")
		if err := runner.RunTestsWithCoverage(); err != nil {
			log.Fatalf("❌ Tests with coverage failed: %v", err)
		}
		fmt.Println("✅ Tests with coverage completed")

		// Analyze coverage after running tests
	coverageInfo, err := runner.AnalyzeCoverage()
	if err != nil {
		log.Fatalf("❌ Failed to analyze coverage: %v", err)
	} else {
		coverageInfo.PrintCoverageReport()
	}

	case *specificTest != "":
		fmt.Printf("🧪 Running specific test: %s\n", *specificTest)
		if err := runner.RunSpecificTest(*specificTest); err != nil {
			log.Fatalf("❌ Specific test failed: %v", err)
		}
		fmt.Println("✅ Specific test completed")

	case *packagePath != "":
		fmt.Printf("🧪 Running tests for package: %s\n", *packagePath)
		if err := runner.RunPackageTests(*packagePath); err != nil {
			log.Fatalf("❌ Package tests failed: %v", err)
		}
		fmt.Println("✅ Package tests completed")

	default:
		// Default: run all tests
		fmt.Println("🧪 Running all tests...")
		if err := runner.RunTests(); err != nil {
			log.Fatalf("❌ Tests failed: %v", err)
		}
		fmt.Println("✅ All tests completed")
	}
}

func printHelp() {
	fmt.Println("Test Runner - Go-based test execution utility")
	fmt.Println("=============================================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/testrunner/main.go [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -coverage     Run tests with coverage reporting")
	fmt.Println("  -analyze      Analyze existing coverage report")
	fmt.Println("  -threshold    Coverage threshold percentage (default: 70.0)")
	fmt.Println("  -verbose      Verbose output (default: true)")
	fmt.Println("  -race         Enable race detection (default: true)")
	fmt.Println("  -run          Run specific test pattern")
	fmt.Println("  -package      Run tests for specific package")
	fmt.Println("  -validate     Validate test environment")
	fmt.Println("  -summary      Show test summary")
	fmt.Println("  -cleanup      Clean up coverage files")
	fmt.Println("  -help         Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/testrunner/main.go")
	fmt.Println("  go run cmd/testrunner/main.go -coverage")
	fmt.Println("  go run cmd/testrunner/main.go -run TestHealthCheck")
	fmt.Println("  go run cmd/testrunner/main.go -package ./internal/handlers")
	fmt.Println("  go run cmd/testrunner/main.go -analyze")
	fmt.Println("  go run cmd/testrunner/main.go -validate")
}
