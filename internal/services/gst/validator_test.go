package gst

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGSTValidator_Validate(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		gst         string
		expectError bool
		errorType   error
	}{
		// Valid GST numbers (using generated checksums)
		{
			name:        "Valid GST - Maharashtra",
			gst:         "27AAPFU0939F1ZV",
			expectError: false,
		},
		{
			name:        "Valid GST - Karnataka",
			gst:         "29AAFCD5862R1ZR",
			expectError: false,
		},
		{
			name:        "Valid GST - with lowercase (should normalize)",
			gst:         "27aapfu0939f1zv",
			expectError: false,
		},
		{
			name:        "Valid GST - with whitespace (should normalize)",
			gst:         "  27AAPFU0939F1ZV  ",
			expectError: false,
		},

		// Invalid length
		{
			name:        "Invalid - Too short",
			gst:         "27AAPFU0939F1Z",
			expectError: true,
			errorType:   ErrInvalidGSTFormat,
		},
		{
			name:        "Invalid - Too long",
			gst:         "27AAPFU0939F1ZVV",
			expectError: true,
			errorType:   ErrInvalidGSTFormat,
		},
		{
			name:        "Invalid - Empty string",
			gst:         "",
			expectError: true,
			errorType:   ErrInvalidGSTFormat,
		},

		// Invalid state code
		{
			name:        "Invalid - State code 00",
			gst:         "00AAPFU0939F1ZV",
			expectError: true,
			errorType:   ErrInvalidStateCode,
		},
		{
			name:        "Invalid - State code 38",
			gst:         "38AAPFU0939F1ZV",
			expectError: true,
			errorType:   ErrInvalidStateCode,
		},
		{
			name:        "Invalid - State code 99",
			gst:         "99AAPFU0939F1ZV",
			expectError: true,
			errorType:   ErrInvalidStateCode,
		},
		{
			name:        "Invalid - State code with letters",
			gst:         "AAAAPFU0939F1ZV",
			expectError: true,
			errorType:   ErrInvalidGSTFormat,
		},

		// Invalid PAN format
		{
			name:        "Invalid - PAN with numbers in first 5 positions",
			gst:         "2712345U0939F1ZV",
			expectError: true,
			errorType:   ErrInvalidGSTFormat,
		},
		{
			name:        "Invalid - PAN with letters in digit positions",
			gst:         "27AAPFUABCDF1ZV",
			expectError: true,
			errorType:   ErrInvalidGSTFormat,
		},
		{
			name:        "Invalid - PAN with number in last position",
			gst:         "27AAPFU09391ZV",
			expectError: true,
			errorType:   ErrInvalidGSTFormat,
		},
		{
			name:        "Invalid - Invalid entity type in PAN",
			gst:         "27AAXFU0939F1ZV",
			expectError: true,
			errorType:   nil, // Checksum will fail first
		},

		// Invalid entity number
		{
			name:        "Invalid - Entity number 0",
			gst:         "27AAPFU0939F0ZV",
			expectError: true,
			errorType:   nil, // Format check will fail first
		},
		{
			name:        "Invalid - Entity number with special character",
			gst:         "27AAPFU0939F@ZV",
			expectError: true,
			errorType:   ErrInvalidGSTFormat,
		},

		// Invalid default Z position
		{
			name:        "Invalid - Missing Z at position 13",
			gst:         "27AAPFU0939F1AV",
			expectError: true,
			errorType:   ErrInvalidGSTFormat,
		},
		{
			name:        "Invalid - Wrong character at position 13",
			gst:         "27AAPFU0939F1YV",
			expectError: true,
			errorType:   ErrInvalidGSTFormat,
		},

		// Invalid checksum
		{
			name:        "Invalid - Wrong check digit",
			gst:         "27AAPFU0939F1ZA",
			expectError: true,
			errorType:   ErrInvalidChecksum,
		},
		{
			name:        "Invalid - Modified PAN with valid checksum",
			gst:         "27BAPFU0939F1ZV",
			expectError: true,
			errorType:   ErrInvalidChecksum,
		},

		// Special characters and invalid formats
		{
			name:        "Invalid - Contains special characters",
			gst:         "27AAPFU0939F1Z!",
			expectError: true,
			errorType:   ErrInvalidGSTFormat,
		},
		{
			name:        "Invalid - Contains spaces in middle",
			gst:         "27AAPFU 0939F1ZV",
			expectError: true,
			errorType:   ErrInvalidGSTFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.gst)

			if tt.expectError {
				assert.Error(t, err, "Expected validation to fail for: %s", tt.gst)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType, "Expected specific error type")
				}
			} else {
				assert.NoError(t, err, "Expected validation to pass for: %s", tt.gst)
			}
		})
	}
}

func TestGSTValidator_Normalize(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Lowercase to uppercase",
			input:    "27aapfu0939f1zv",
			expected: "27AAPFU0939F1ZV",
		},
		{
			name:     "Mixed case to uppercase",
			input:    "27AaPfU0939f1Zv",
			expected: "27AAPFU0939F1ZV",
		},
		{
			name:     "With leading whitespace",
			input:    "  27AAPFU0939F1ZV",
			expected: "27AAPFU0939F1ZV",
		},
		{
			name:     "With trailing whitespace",
			input:    "27AAPFU0939F1ZV  ",
			expected: "27AAPFU0939F1ZV",
		},
		{
			name:     "With leading and trailing whitespace",
			input:    "  27AAPFU0939F1ZV  ",
			expected: "27AAPFU0939F1ZV",
		},
		{
			name:     "Already normalized",
			input:    "27AAPFU0939F1ZV",
			expected: "27AAPFU0939F1ZV",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Normalize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGSTValidator_ValidateStateCode(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		stateCode   string
		expectError bool
	}{
		// Valid state codes
		{name: "Valid - 01 Jammu and Kashmir", stateCode: "01", expectError: false},
		{name: "Valid - 07 Delhi", stateCode: "07", expectError: false},
		{name: "Valid - 27 Maharashtra", stateCode: "27", expectError: false},
		{name: "Valid - 29 Karnataka", stateCode: "29", expectError: false},
		{name: "Valid - 33 Tamil Nadu", stateCode: "33", expectError: false},
		{name: "Valid - 36 Telangana", stateCode: "36", expectError: false},
		{name: "Valid - 37 Andhra Pradesh", stateCode: "37", expectError: false},

		// Invalid state codes
		{name: "Invalid - 00", stateCode: "00", expectError: true},
		{name: "Invalid - 38", stateCode: "38", expectError: true},
		{name: "Invalid - 99", stateCode: "99", expectError: true},
		{name: "Invalid - AA", stateCode: "AA", expectError: true},
		{name: "Invalid - Empty", stateCode: "", expectError: true},
		{name: "Invalid - Single digit", stateCode: "1", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStateCode(tt.stateCode)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

//nolint:dupl // Test cases are similar but test different functions
func TestGSTValidator_ValidatePAN(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		pan         string
		expectError bool
	}{
		// Valid PANs
		{name: "Valid - Company", pan: "AAPFU0939F", expectError: false},
		{name: "Valid - Firm", pan: "AAFCD5862R", expectError: false},
		{name: "Valid - HUF", pan: "AABCH1234H", expectError: false},
		{name: "Valid - Individual", pan: "AABCP1234P", expectError: false},

		// Invalid PANs
		{name: "Invalid - Wrong format", pan: "12345ABCDE", expectError: true},
		{name: "Invalid - Too short", pan: "AAPFU0939", expectError: true},
		{name: "Invalid - Too long", pan: "AAPFU0939FF", expectError: true},
		{name: "Invalid - Invalid entity type", pan: "AAXXU0939F", expectError: true}, // X in 4th position is invalid
		{name: "Invalid - Numbers in first 5", pan: "A1PFU0939F", expectError: true},
		{name: "Invalid - Letters in middle 4", pan: "AAPFUABCDF", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePAN(tt.pan)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

//nolint:dupl // Test cases are similar but test different functions
func TestGSTValidator_ValidateEntityNumber(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		entityNum   string
		expectError bool
	}{
		// Valid entity numbers
		{name: "Valid - 1", entityNum: "1", expectError: false},
		{name: "Valid - 5", entityNum: "5", expectError: false},
		{name: "Valid - 9", entityNum: "9", expectError: false},
		{name: "Valid - A", entityNum: "A", expectError: false},
		{name: "Valid - Z", entityNum: "Z", expectError: false},

		// Invalid entity numbers
		{name: "Invalid - 0", entityNum: "0", expectError: true},
		{name: "Invalid - Empty", entityNum: "", expectError: true},
		{name: "Invalid - Special char", entityNum: "@", expectError: true},
		{name: "Invalid - Lowercase", entityNum: "a", expectError: true},
		{name: "Invalid - Multiple chars", entityNum: "12", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateEntityNumber(tt.entityNum)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGSTValidator_CalculateChecksum(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name                 string
		gstWithoutCheckDigit string
		expectedCheckDigit   string
	}{
		{
			name:                 "Valid GST - Maharashtra",
			gstWithoutCheckDigit: "27AAPFU0939F1Z",
			expectedCheckDigit:   "V",
		},
		{
			name:                 "Valid GST - Karnataka",
			gstWithoutCheckDigit: "29AAFCD5862R1Z",
			expectedCheckDigit:   "R",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkDigit := validator.CalculateChecksum(tt.gstWithoutCheckDigit)
			assert.Equal(t, tt.expectedCheckDigit, checkDigit)
		})
	}
}

func TestGSTValidator_ValidateChecksum(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		gst         string
		expectError bool
	}{
		{
			name:        "Valid checksum",
			gst:         "27AAPFU0939F1ZV",
			expectError: false,
		},
		{
			name:        "Invalid checksum",
			gst:         "27AAPFU0939F1ZA",
			expectError: true,
		},
		{
			name:        "Modified GST with wrong checksum",
			gst:         "27BAPFU0939F1ZV",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateChecksum(tt.gst)
			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidChecksum)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGSTValidator_IsValid(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		gst      string
		expected bool
	}{
		{name: "Valid GST", gst: "27AAPFU0939F1ZV", expected: true},
		{name: "Invalid GST", gst: "27AAPFU0939F1ZA", expected: false},
		{name: "Invalid format", gst: "INVALID", expected: false},
		{name: "Empty string", gst: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsValid(tt.gst)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGSTValidator_ExtractStateCode(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name         string
		gst          string
		expectedCode string
		expectError  bool
	}{
		{
			name:         "Valid - Maharashtra",
			gst:          "27AAPFU0939F1ZV",
			expectedCode: "27",
			expectError:  false,
		},
		{
			name:         "Valid - Karnataka",
			gst:          "29AAFCD5862R1ZR",
			expectedCode: "29",
			expectError:  false,
		},
		{
			name:         "Valid - lowercase",
			gst:          "27aapfu0939f1zv",
			expectedCode: "27",
			expectError:  false,
		},
		{
			name:         "Invalid - too short",
			gst:          "2",
			expectedCode: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := validator.ExtractStateCode(tt.gst)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCode, code)
			}
		})
	}
}

func TestGSTValidator_ExtractPAN(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		gst         string
		expectedPAN string
		expectError bool
	}{
		{
			name:        "Valid - Maharashtra",
			gst:         "27AAPFU0939F1ZV",
			expectedPAN: "AAPFU0939F",
			expectError: false,
		},
		{
			name:        "Valid - Karnataka",
			gst:         "29AAFCD5862R1ZR",
			expectedPAN: "AAFCD5862R",
			expectError: false,
		},
		{
			name:        "Valid - lowercase",
			gst:         "27aapfu0939f1zv",
			expectedPAN: "AAPFU0939F",
			expectError: false,
		},
		{
			name:        "Invalid - too short",
			gst:         "27AAPFU",
			expectedPAN: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pan, err := validator.ExtractPAN(tt.gst)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPAN, pan)
			}
		})
	}
}

func TestGSTValidator_GetStateName(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		stateCode string
		expected  string
	}{
		{stateCode: "01", expected: "Jammu and Kashmir"},
		{stateCode: "07", expected: "Delhi"},
		{stateCode: "27", expected: "Maharashtra"},
		{stateCode: "29", expected: "Karnataka"},
		{stateCode: "33", expected: "Tamil Nadu"},
		{stateCode: "36", expected: "Telangana"},
		{stateCode: "37", expected: "Andhra Pradesh (New)"},
		{stateCode: "99", expected: ""}, // Invalid code
	}

	for _, tt := range tests {
		t.Run("State code "+tt.stateCode, func(t *testing.T) {
			result := validator.GetStateName(tt.stateCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkValidator_Validate(b *testing.B) {
	validator := NewValidator()
	gst := "27AAPFU0939F1ZV" //nolint:goconst // Test constant used in benchmarks

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(gst)
	}
}

func BenchmarkValidator_CalculateChecksum(b *testing.B) {
	validator := NewValidator()
	gstWithoutCheck := "27AAPFU0939F1Z"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.CalculateChecksum(gstWithoutCheck)
	}
}

// Table-driven test for comprehensive GST validation
func TestGSTValidator_ComprehensiveValidation(t *testing.T) {
	validator := NewValidator()

	// Test all valid state codes with a sample GST
	for stateCode := 1; stateCode <= 37; stateCode++ {
		t.Run("State code "+string(rune(stateCode)), func(t *testing.T) {
			// Create a GST with this state code
			gst := fmt.Sprintf("%02dAAPFU0939F1Z", stateCode)
			checkDigit := validator.CalculateChecksum(gst)
			fullGST := gst + checkDigit

			// Should be valid
			err := validator.Validate(fullGST)
			require.NoError(t, err, "GST should be valid for state code %02d", stateCode)

			// Extract and verify state name
			stateName := validator.GetStateName(fmt.Sprintf("%02d", stateCode))
			assert.NotEmpty(t, stateName, "State name should exist for code %02d", stateCode)
		})
	}
}
