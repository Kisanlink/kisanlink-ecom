package gst

import (
	"context"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create collaborators table
	err = db.Exec(`
		CREATE TABLE collaborators (
			id TEXT PRIMARY KEY,
			tax_id TEXT,
			deleted_at TIMESTAMP
		)
	`).Error
	require.NoError(t, err)

	return db
}

func TestNewService(t *testing.T) {
	db := setupTestDB(t)
	logger := logrus.New()

	service := NewService(db, logger)

	assert.NotNil(t, service)
	assert.NotNil(t, service.db)
	assert.NotNil(t, service.validator)
	assert.NotNil(t, service.logger)
}

func TestGSTService_ValidateAndNormalize(t *testing.T) {
	db := setupTestDB(t)
	logger := logrus.New()
	service := NewService(db, logger)

	tests := []struct {
		name        string
		gst         string
		expectedGST string
		expectError bool
	}{
		{
			name:        "Valid GST - uppercase",
			gst:         "27AAPFU0939F1ZV",
			expectedGST: "27AAPFU0939F1ZV",
			expectError: false,
		},
		{
			name:        "Valid GST - lowercase",
			gst:         "27aapfu0939f1zv",
			expectedGST: "27AAPFU0939F1ZV",
			expectError: false,
		},
		{
			name:        "Valid GST - with whitespace",
			gst:         "  27AAPFU0939F1ZV  ",
			expectedGST: "27AAPFU0939F1ZV",
			expectError: false,
		},
		{
			name:        "Invalid GST - wrong checksum",
			gst:         "27AAPFU0939F1ZA",
			expectedGST: "",
			expectError: true,
		},
		{
			name:        "Invalid GST - wrong format",
			gst:         "INVALID",
			expectedGST: "",
			expectError: true,
		},
		{
			name:        "Invalid GST - empty",
			gst:         "",
			expectedGST: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, err := service.ValidateAndNormalize(tt.gst)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, normalized)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedGST, normalized)
			}
		})
	}
}

func TestGSTService_CheckGSTExists(t *testing.T) {
	db := setupTestDB(t)
	logger := logrus.New()
	service := NewService(db, logger)

	ctx := context.Background()

	// Insert a test GST
	err := db.Exec(`
		INSERT INTO collaborators (id, tax_id, deleted_at)
		VALUES ('test-id-1', '27AAPFU0939F1ZV', NULL)
	`).Error
	require.NoError(t, err)

	// Insert a soft-deleted GST
	err = db.Exec(`
		INSERT INTO collaborators (id, tax_id, deleted_at)
		VALUES ('test-id-2', '29AAFCD5862R1ZR', datetime('now'))
	`).Error
	require.NoError(t, err)

	tests := []struct {
		name        string
		gst         string
		expected    bool
		expectError bool
	}{
		{
			name:        "Existing GST - exact match",
			gst:         "27AAPFU0939F1ZV",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Existing GST - lowercase",
			gst:         "27aapfu0939f1zv",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Existing GST - with whitespace",
			gst:         "  27AAPFU0939F1ZV  ",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Non-existing GST",
			gst:         "07AABCU9603R1ZP",
			expected:    false,
			expectError: false,
		},
		{
			name:        "Soft-deleted GST - should not exist",
			gst:         "29AAFCD5862R1ZR",
			expected:    false,
			expectError: false,
		},
		{
			name:        "Invalid GST",
			gst:         "INVALID",
			expected:    false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := service.CheckGSTExists(ctx, tt.gst)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, exists)
			}
		})
	}
}

func TestGSTService_GetInfo(t *testing.T) {
	db := setupTestDB(t)
	logger := logrus.New()
	service := NewService(db, logger)

	tests := []struct {
		name           string
		gst            string
		expectValid    bool
		expectedState  string
		expectedPAN    string
		expectedEntity string
	}{
		{
			name:           "Valid GST - Company",
			gst:            "27AAPFU0939F1ZV",
			expectValid:    true,
			expectedState:  "27",
			expectedPAN:    "AAPFU0939F",
			expectedEntity: "Firm",
		},
		{
			name:           "Valid GST - Company",
			gst:            "29AAFCD5862R1ZR",
			expectValid:    true,
			expectedState:  "29",
			expectedPAN:    "AAFCD5862R",
			expectedEntity: "Company", // C at position 4
		},
		{
			name:           "Valid GST - lowercase",
			gst:            "27aapfu0939f1zv",
			expectValid:    true,
			expectedState:  "27",
			expectedPAN:    "AAPFU0939F",
			expectedEntity: "Firm",
		},
		{
			name:        "Invalid GST",
			gst:         "INVALID",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := service.GetInfo(tt.gst)

			if tt.expectValid {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				assert.True(t, info.IsValid)
				assert.Equal(t, tt.expectedState, info.StateCode)
				assert.Equal(t, tt.expectedPAN, info.PAN)
				assert.NotEmpty(t, info.StateName)
				assert.Equal(t, tt.expectedEntity, info.EntityType)
			} else {
				assert.Error(t, err)
				assert.NotNil(t, info)
				assert.False(t, info.IsValid)
			}
		})
	}
}

func TestGSTService_GetInfo_EntityTypes(t *testing.T) {
	db := setupTestDB(t)
	logger := logrus.New()
	service := NewService(db, logger)

	// Test different entity types based on 4th character in PAN
	tests := []struct {
		panChar      byte
		expectedType string
	}{
		{'C', "Company"},
		{'F', "Firm"},
		{'H', "Hindu Undivided Family"},
		{'A', "Association of Persons"},
		{'P', "Individual"},
		{'L', "Local Authority"},
		{'J', "Artificial Juridical Person"},
		{'T', "Trust"},
		{'B', "Body of Individuals"},
		{'G', "Government"},
		{'S', "Others"},
	}

	for _, tt := range tests {
		t.Run("EntityType_"+string(tt.panChar), func(t *testing.T) {
			// Create a GST with the specific entity type character
			// PAN format: LLLLL NNNN L where 4th character (index 3) is entity type
			gst := fmt.Sprintf("27AAA%cU0939F1Z", tt.panChar)

			// Calculate checksum
			validator := NewValidator()
			checkDigit := validator.CalculateChecksum(gst)
			fullGST := gst + checkDigit

			// Get GST info
			info, err := service.GetInfo(fullGST)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, info.EntityType)
		})
	}
}

// Benchmark tests
func BenchmarkService_ValidateAndNormalize(b *testing.B) {
	db := setupTestDB(&testing.T{})
	logger := logrus.New()
	service := NewService(db, logger)
	gst := "27AAPFU0939F1ZV"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ValidateAndNormalize(gst)
	}
}

func BenchmarkService_CheckGSTExists(b *testing.B) {
	db := setupTestDB(&testing.T{})
	logger := logrus.New()
	service := NewService(db, logger)
	ctx := context.Background()
	gst := "27AAPFU0939F1ZV"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CheckGSTExists(ctx, gst)
	}
}
