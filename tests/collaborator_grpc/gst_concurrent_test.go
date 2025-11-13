package collaborator_grpc_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"kisanlink-ecom/internal/services/gst"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Collaborator model for testing (simplified)
type Collaborator struct {
	ID        uint64 `gorm:"primaryKey"`
	TaxID     string `gorm:"column:tax_id;unique;not null"`
	Name      string
	FPOID     uint64 `gorm:"column:fpo_id"`
	DeletedAt *time.Time
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Create collaborators table
	err = db.Exec(`
		CREATE TABLE collaborators (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tax_id TEXT UNIQUE NOT NULL,
			name TEXT,
			fpo_id INTEGER,
			deleted_at TIMESTAMP
		)
	`).Error
	require.NoError(t, err)

	return db
}

func setupTestRedis(t *testing.T) redis.UniversalClient {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use dedicated test database
	})

	ctx := context.Background()
	err := client.FlushDB(ctx).Err()
	require.NoError(t, err)

	return client
}

//nolint:gosec // Test uses small int values, overflow impossible
func TestGSTService_NoDuplicates_Concurrent(t *testing.T) {
	// Skip if Redis not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	redisClient := setupTestRedis(t)
	defer func() { _ = redisClient.Close() }()

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	gstService := gst.NewService(db, redisClient, logger)
	ctx := context.Background()

	// Same GST number used by multiple FPOs concurrently
	gstNumber := "29ABCDE1234F1Z5"
	numGoroutines := 50

	var wg sync.WaitGroup
	successCount := atomic.Int32{}
	reservedBy := atomic.Value{}

	// Spawn concurrent goroutines trying to reserve the same GST
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(fpoID int) {
			defer wg.Done()

			requestID := uuid.New().String()
			exists, err := gstService.CheckAndReserveGST(ctx, gstNumber, uint64(fpoID), requestID)

			if err != nil {
				t.Logf("FPO %d got error: %v", fpoID, err)
				return
			}

			if !exists {
				// Successfully reserved - simulate creating collaborator
				successCount.Add(1)
				reservedBy.Store(fpoID)

				// Simulate database insert
				collaborator := Collaborator{
					TaxID: gstNumber,
					Name:  fmt.Sprintf("Test Collaborator FPO %d", fpoID),
					FPOID: uint64(fpoID),
				}

				if err := db.Create(&collaborator).Error; err != nil {
					t.Errorf("FPO %d failed to create collaborator: %v", fpoID, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify: Only ONE goroutine should have successfully reserved and created
	assert.Equal(t, int32(1), successCount.Load(), "Exactly one FPO should reserve the GST")

	// Verify database: Only one record should exist
	var count int64
	err := db.Model(&Collaborator{}).Where("tax_id = ?", gstNumber).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "Database should have exactly one collaborator with this GST")

	t.Logf("GST reserved by FPO: %v", reservedBy.Load())
}

//nolint:gosec // Test uses small int values, overflow impossible
func TestGSTService_NoDuplicates_HighConcurrency(t *testing.T) {
	// Skip if Redis not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	redisClient := setupTestRedis(t)
	defer func() { _ = redisClient.Close() }()

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	gstService := gst.NewService(db, redisClient, logger)
	ctx := context.Background()

	// Test with 100+ concurrent goroutines
	gstNumber := "24AAAAA0000A1Z5"
	numGoroutines := 100

	var wg sync.WaitGroup
	successCount := atomic.Int32{}

	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(fpoID int) {
			defer wg.Done()

			requestID := uuid.New().String()
			exists, err := gstService.CheckAndReserveGST(ctx, gstNumber, uint64(fpoID), requestID)

			if err == nil && !exists {
				successCount.Add(1)

				// Simulate database insert
				collaborator := Collaborator{
					TaxID: gstNumber,
					Name:  fmt.Sprintf("Collaborator %d", fpoID),
					FPOID: uint64(fpoID),
				}
				db.Create(&collaborator)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Assertions
	assert.Equal(t, int32(1), successCount.Load(), "Only one should succeed")

	var dbCount int64
	db.Model(&Collaborator{}).Where("tax_id = ?", gstNumber).Count(&dbCount)
	assert.Equal(t, int64(1), dbCount, "Only one database record")

	t.Logf("100 concurrent requests completed in %v", duration)
	assert.Less(t, duration, 10*time.Second, "Should complete within reasonable time")
}

//nolint:gosec // Test uses small int values, overflow impossible
func TestGSTService_MultipleGSTs_NoCrossContention(t *testing.T) {
	// Skip if Redis not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	redisClient := setupTestRedis(t)
	defer func() { _ = redisClient.Close() }()

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	gstService := gst.NewService(db, redisClient, logger)
	ctx := context.Background()

	// Test multiple different GST numbers simultaneously
	gstNumbers := []string{
		"29ABCDE1234F1Z5",
		"24BBBBB2345G2Z6",
		"27CCCCC3456H3Z7",
		"33DDDDD4567I4Z8",
		"07EEEEE5678J5Z9",
	}

	var wg sync.WaitGroup
	successCounts := make([]atomic.Int32, len(gstNumbers))

	// Each GST number attempted by 20 concurrent FPOs
	for gstIdx, gstNumber := range gstNumbers {
		for fpoID := 0; fpoID < 20; fpoID++ {
			wg.Add(1)
			go func(idx int, gst string, fpo int) {
				defer wg.Done()

				requestID := uuid.New().String()
				exists, err := gstService.CheckAndReserveGST(ctx, gst, uint64(fpo), requestID)

				if err == nil && !exists {
					successCounts[idx].Add(1)

					collaborator := Collaborator{
						TaxID: gst,
						Name:  fmt.Sprintf("Collab GST%d FPO%d", idx, fpo),
						FPOID: uint64(fpo),
					}
					db.Create(&collaborator)
				}
			}(gstIdx, gstNumber, fpoID)
		}
	}

	wg.Wait()

	// Each GST should have exactly one success
	for idx, gst := range gstNumbers {
		assert.Equal(t, int32(1), successCounts[idx].Load(),
			"GST %s should have exactly one reservation", gst)

		var dbCount int64
		db.Model(&Collaborator{}).Where("tax_id = ?", gst).Count(&dbCount)
		assert.Equal(t, int64(1), dbCount,
			"GST %s should have exactly one database record", gst)
	}

	// Total should be 5 unique collaborators
	var totalCount int64
	db.Model(&Collaborator{}).Count(&totalCount)
	assert.Equal(t, int64(5), totalCount, "Should have 5 total collaborators")
}

//nolint:gosec // Test uses small int values, overflow impossible
func TestGSTService_RaceDetector(t *testing.T) {
	// This test should be run with: go test -race
	// Skip if Redis not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	redisClient := setupTestRedis(t)
	defer func() { _ = redisClient.Close() }()

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	gstService := gst.NewService(db, redisClient, logger)
	ctx := context.Background()

	gstNumber := "29ZZZZZ9999Z9Z9"

	// High contention scenario
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			requestID := uuid.New().String()
			exists, _ := gstService.CheckAndReserveGST(ctx, gstNumber, uint64(id), requestID)

			if !exists {
				// Winner writes to database
				collaborator := Collaborator{
					TaxID: gstNumber,
					Name:  fmt.Sprintf("Winner %d", id),
					FPOID: uint64(id),
				}
				db.Create(&collaborator)
			}
		}(i)
	}

	wg.Wait()

	// Race detector will fail if there are any data races
	var count int64
	db.Model(&Collaborator{}).Where("tax_id = ?", gstNumber).Count(&count)
	assert.Equal(t, int64(1), count)
}

//nolint:gosec // Test uses small int values, overflow impossible
func TestGSTService_LockPerformance(t *testing.T) {
	// Skip if Redis not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	redisClient := setupTestRedis(t)
	defer func() { _ = redisClient.Close() }()

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	gstService := gst.NewService(db, redisClient, logger)
	ctx := context.Background()

	// Measure P50, P95, P99 latencies
	numRequests := 1000
	latencies := make([]time.Duration, numRequests)

	var wg sync.WaitGroup
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Each goroutine uses a unique GST (no contention)
			gstNumber := fmt.Sprintf("29TEST%04d%01dZ%01d", idx, idx%10, idx%36)
			requestID := uuid.New().String()

			start := time.Now()
			_, _ = gstService.CheckAndReserveGST(ctx, gstNumber, uint64(idx), requestID)
			latencies[idx] = time.Since(start)
		}(i)
	}

	wg.Wait()

	// Calculate percentiles (simple sorting approach)
	// In production, use a proper percentile library
	p50 := calculatePercentile(latencies, 50)
	p95 := calculatePercentile(latencies, 95)
	p99 := calculatePercentile(latencies, 99)

	t.Logf("Lock acquisition latencies:")
	t.Logf("  P50: %v", p50)
	t.Logf("  P95: %v", p95)
	t.Logf("  P99: %v", p99)

	// Performance assertions
	assert.Less(t, p99, 500*time.Millisecond, "P99 should be under 500ms")
	assert.Less(t, p95, 200*time.Millisecond, "P95 should be under 200ms")
}

// Helper function to calculate percentile
func calculatePercentile(durations []time.Duration, percentile int) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	// Simple bubble sort for small datasets
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := (len(sorted) * percentile) / 100
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}
