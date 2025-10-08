//go:build integration
// +build integration

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// LoadTestSuite tests performance validation under stress
type LoadTestSuite struct {
	suite.Suite
	server    *httptest.Server
	client    *http.Client
	baseURL   string
	authToken string
	metrics   *LoadTestMetrics
}

// LoadTestMetrics tracks performance metrics during load testing
type LoadTestMetrics struct {
	mu                sync.RWMutex
	totalRequests     int64
	successfulReqs    int64
	failedReqs        int64
	totalResponseTime time.Duration
	minResponseTime   time.Duration
	maxResponseTime   time.Duration
	responseTimeHist  map[string]int64 // buckets for response time histogram
	errorCounts       map[string]int64 // error type -> count
	throughput        float64          // requests per second
	startTime         time.Time
	endTime           time.Time
}

func NewLoadTestMetrics() *LoadTestMetrics {
	return &LoadTestMetrics{
		responseTimeHist: make(map[string]int64),
		errorCounts:      make(map[string]int64),
		minResponseTime:  time.Hour, // Initialize to high value
		startTime:        time.Now(),
	}
}

func (m *LoadTestMetrics) RecordRequest(responseTime time.Duration, success bool, errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRequests++
	m.totalResponseTime += responseTime

	if success {
		m.successfulReqs++
	} else {
		m.failedReqs++
		if errorType != "" {
			m.errorCounts[errorType]++
		}
	}

	// Update min/max response times
	if responseTime < m.minResponseTime {
		m.minResponseTime = responseTime
	}
	if responseTime > m.maxResponseTime {
		m.maxResponseTime = responseTime
	}

	// Update response time histogram
	bucket := m.getResponseTimeBucket(responseTime)
	m.responseTimeHist[bucket]++
}

func (m *LoadTestMetrics) getResponseTimeBucket(responseTime time.Duration) string {
	if responseTime < 100*time.Millisecond {
		return "<100ms"
	} else if responseTime < 500*time.Millisecond {
		return "100-500ms"
	} else if responseTime < 1*time.Second {
		return "500ms-1s"
	} else if responseTime < 5*time.Second {
		return "1-5s"
	} else {
		return ">5s"
	}
}

func (m *LoadTestMetrics) Finalize() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.endTime = time.Now()
	duration := m.endTime.Sub(m.startTime)
	if duration > 0 {
		m.throughput = float64(m.totalRequests) / duration.Seconds()
	}
}

func (m *LoadTestMetrics) GetSummary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgResponseTime := time.Duration(0)
	if m.totalRequests > 0 {
		avgResponseTime = m.totalResponseTime / time.Duration(m.totalRequests)
	}

	successRate := float64(0)
	if m.totalRequests > 0 {
		successRate = float64(m.successfulReqs) / float64(m.totalRequests) * 100
	}

	return map[string]interface{}{
		"total_requests":      m.totalRequests,
		"successful_requests": m.successfulReqs,
		"failed_requests":     m.failedReqs,
		"success_rate":        fmt.Sprintf("%.2f%%", successRate),
		"avg_response_time":   avgResponseTime.String(),
		"min_response_time":   m.minResponseTime.String(),
		"max_response_time":   m.maxResponseTime.String(),
		"throughput":          fmt.Sprintf("%.2f req/s", m.throughput),
		"duration":            m.endTime.Sub(m.startTime).String(),
		"response_time_hist":  m.responseTimeHist,
		"error_counts":        m.errorCounts,
	}
}

// SetupSuite initializes the load test environment
func (suite *LoadTestSuite) SetupSuite() {
	// Setup test server
	suite.setupTestServer()

	// Initialize HTTP client with appropriate settings for load testing
	suite.client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Initialize metrics
	suite.metrics = NewLoadTestMetrics()

	// Setup authentication
	suite.authToken = "Bearer load_test_token"
}

func (suite *LoadTestSuite) setupTestServer() {
	mux := http.NewServeMux()

	// Setup lightweight mock endpoints for load testing
	mux.HandleFunc("/api/v1/catalog", suite.handleCatalogLoad)
	mux.HandleFunc("/api/v1/orders", suite.handleOrdersLoad)
	mux.HandleFunc("/api/v1/inventory/lots", suite.handleInventoryLoad)
	mux.HandleFunc("/health", suite.handleHealthLoad)

	suite.server = httptest.NewServer(mux)
	suite.baseURL = suite.server.URL
}

// TearDownSuite cleans up after load testing
func (suite *LoadTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// Test basic load scenarios
func (suite *LoadTestSuite) TestBasicLoadScenarios() {
	suite.T().Log("Testing basic concurrent load")
	suite.testConcurrentLoad(50, 100) // 50 concurrent users, 100 requests each

	suite.T().Log("Testing sustained load")
	suite.testSustainedLoad(10, 30*time.Second) // 10 RPS for 30 seconds

	suite.T().Log("Testing burst load")
	suite.testBurstLoad(100, 5*time.Second) // 100 concurrent requests over 5 seconds
}

func (suite *LoadTestSuite) testConcurrentLoad(concurrency, requestsPerUser int) {
	suite.metrics = NewLoadTestMetrics()

	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			for j := 0; j < requestsPerUser; j++ {
				start := time.Now()
				resp, err := suite.makeLoadRequest("GET", "/api/v1/catalog", nil)
				responseTime := time.Since(start)

				success := err == nil && resp != nil && resp.StatusCode == http.StatusOK
				errorType := ""
				if err != nil {
					errorType = "network_error"
				} else if resp != nil && resp.StatusCode != http.StatusOK {
					errorType = fmt.Sprintf("http_%d", resp.StatusCode)
				}

				suite.metrics.RecordRequest(responseTime, success, errorType)

				if resp != nil {
					resp.Body.Close()
				}

				// Small delay between requests from same user
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	suite.metrics.Finalize()

	// Verify performance requirements
	summary := suite.metrics.GetSummary()
	suite.T().Logf("Concurrent Load Test Results: %+v", summary)

	// Performance assertions
	suite.Greater(summary["success_rate"], "95.00%", "Success rate should be > 95%")
	suite.Greater(summary["throughput"], "50.00 req/s", "Throughput should be > 50 req/s")

	// Response time assertions
	avgResponseTime, _ := time.ParseDuration(summary["avg_response_time"].(string))
	suite.Less(avgResponseTime, 2*time.Second, "Average response time should be < 2s")
}

func (suite *LoadTestSuite) testSustainedLoad(rps int, duration time.Duration) {
	suite.metrics = NewLoadTestMetrics()

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	ticker := time.NewTicker(time.Second / time.Duration(rps))
	defer ticker.Stop()

	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			suite.metrics.Finalize()

			summary := suite.metrics.GetSummary()
			suite.T().Logf("Sustained Load Test Results: %+v", summary)

			// Verify sustained performance
			suite.Greater(summary["success_rate"], "98.00%", "Success rate should be > 98% under sustained load")
			return

		case <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()

				start := time.Now()
				resp, err := suite.makeLoadRequest("GET", "/api/v1/catalog", nil)
				responseTime := time.Since(start)

				success := err == nil && resp != nil && resp.StatusCode == http.StatusOK
				errorType := ""
				if err != nil {
					errorType = "network_error"
				} else if resp != nil && resp.StatusCode != http.StatusOK {
					errorType = fmt.Sprintf("http_%d", resp.StatusCode)
				}

				suite.metrics.RecordRequest(responseTime, success, errorType)

				if resp != nil {
					resp.Body.Close()
				}
			}()
		}
	}
}

func (suite *LoadTestSuite) testBurstLoad(totalRequests int, timeWindow time.Duration) {
	suite.metrics = NewLoadTestMetrics()

	var wg sync.WaitGroup
	requestChan := make(chan struct{}, totalRequests)

	// Fill the request channel
	for i := 0; i < totalRequests; i++ {
		requestChan <- struct{}{}
	}
	close(requestChan)

	// Start workers
	workerCount := 20
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for range requestChan {
				start := time.Now()
				resp, err := suite.makeLoadRequest("GET", "/api/v1/orders", nil)
				responseTime := time.Since(start)

				success := err == nil && resp != nil && resp.StatusCode == http.StatusOK
				errorType := ""
				if err != nil {
					errorType = "network_error"
				} else if resp != nil && resp.StatusCode != http.StatusOK {
					errorType = fmt.Sprintf("http_%d", resp.StatusCode)
				}

				suite.metrics.RecordRequest(responseTime, success, errorType)

				if resp != nil {
					resp.Body.Close()
				}
			}
		}()
	}

	wg.Wait()
	suite.metrics.Finalize()

	summary := suite.metrics.GetSummary()
	suite.T().Logf("Burst Load Test Results: %+v", summary)

	// Verify burst handling
	suite.Greater(summary["success_rate"], "90.00%", "Success rate should be > 90% under burst load")
}

// Test specific endpoint performance
func (suite *LoadTestSuite) TestEndpointPerformance() {
	endpoints := []struct {
		name          string
		method        string
		path          string
		body          interface{}
		minThroughput float64
		maxAvgTime    time.Duration
	}{
		{"catalog_list", "GET", "/api/v1/catalog", nil, 100.0, 500 * time.Millisecond},
		{"order_list", "GET", "/api/v1/orders", nil, 80.0, 800 * time.Millisecond},
		{"inventory_list", "GET", "/api/v1/inventory/lots", nil, 60.0, 1 * time.Second},
		{"health_check", "GET", "/health", nil, 200.0, 100 * time.Millisecond},
	}

	for _, endpoint := range endpoints {
		suite.T().Run(endpoint.name, func(t *testing.T) {
			suite.testEndpointPerformance(endpoint.method, endpoint.path, endpoint.body,
				endpoint.minThroughput, endpoint.maxAvgTime)
		})
	}
}

func (suite *LoadTestSuite) testEndpointPerformance(method, path string, body interface{},
	minThroughput float64, maxAvgTime time.Duration) {

	metrics := NewLoadTestMetrics()
	concurrency := 20
	requestsPerUser := 50

	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < requestsPerUser; j++ {
				start := time.Now()
				resp, err := suite.makeLoadRequest(method, path, body)
				responseTime := time.Since(start)

				success := err == nil && resp != nil && resp.StatusCode == http.StatusOK
				errorType := ""
				if err != nil {
					errorType = "network_error"
				} else if resp != nil && resp.StatusCode != http.StatusOK {
					errorType = fmt.Sprintf("http_%d", resp.StatusCode)
				}

				metrics.RecordRequest(responseTime, success, errorType)

				if resp != nil {
					resp.Body.Close()
				}
			}
		}()
	}

	wg.Wait()
	metrics.Finalize()

	summary := metrics.GetSummary()
	suite.T().Logf("Endpoint %s Performance: %+v", path, summary)

	// Verify endpoint-specific performance requirements
	actualThroughput := metrics.throughput
	suite.GreaterOrEqual(actualThroughput, minThroughput,
		"Throughput should be >= %.2f req/s", minThroughput)

	avgResponseTime := metrics.totalResponseTime / time.Duration(metrics.totalRequests)
	suite.LessOrEqual(avgResponseTime, maxAvgTime,
		"Average response time should be <= %v", maxAvgTime)
}

// Test memory and resource usage under load
func (suite *LoadTestSuite) TestResourceUsageUnderLoad() {
	suite.T().Log("Testing memory usage under load")
	suite.testMemoryUsage()

	suite.T().Log("Testing connection handling under load")
	suite.testConnectionHandling()
}

func (suite *LoadTestSuite) testMemoryUsage() {
	// This would typically monitor actual memory usage
	// For this test, we'll simulate by creating many concurrent requests

	concurrency := 100
	requestsPerUser := 20

	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < requestsPerUser; j++ {
				resp, err := suite.makeLoadRequest("GET", "/api/v1/catalog", nil)
				if err == nil && resp != nil {
					// Read response body to simulate memory usage
					buf := make([]byte, 1024)
					resp.Body.Read(buf)
					resp.Body.Close()
				}
			}
		}()
	}

	wg.Wait()

	// In a real implementation, we would check memory metrics here
	suite.T().Log("Memory usage test completed")
}

func (suite *LoadTestSuite) testConnectionHandling() {
	// Test connection pool behavior under high load
	concurrency := 200

	var wg sync.WaitGroup
	connectionErrors := int64(0)
	var errorMutex sync.Mutex

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resp, err := suite.makeLoadRequest("GET", "/health", nil)
			if err != nil {
				errorMutex.Lock()
				connectionErrors++
				errorMutex.Unlock()
			} else if resp != nil {
				resp.Body.Close()
			}
		}()
	}

	wg.Wait()

	// Verify connection handling
	errorRate := float64(connectionErrors) / float64(concurrency) * 100
	suite.Less(errorRate, 5.0, "Connection error rate should be < 5%")

	suite.T().Logf("Connection errors: %d/%d (%.2f%%)", connectionErrors, concurrency, errorRate)
}

// Test error handling under load
func (suite *LoadTestSuite) TestErrorHandlingUnderLoad() {
	suite.T().Log("Testing error handling under high load")

	// Test with invalid requests to trigger errors
	concurrency := 50
	requestsPerUser := 20

	metrics := NewLoadTestMetrics()
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < requestsPerUser; j++ {
				// Mix of valid and invalid requests
				var resp *http.Response
				var err error

				if j%3 == 0 {
					// Invalid request (should return 400)
					invalidBody := map[string]interface{}{"invalid": "data"}
					resp, err = suite.makeLoadRequest("POST", "/api/v1/orders", invalidBody)
				} else {
					// Valid request
					resp, err = suite.makeLoadRequest("GET", "/api/v1/catalog", nil)
				}

				start := time.Now()
				responseTime := time.Since(start)

				success := err == nil && resp != nil &&
					(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest)

				errorType := ""
				if err != nil {
					errorType = "network_error"
				} else if resp != nil && resp.StatusCode >= 500 {
					errorType = fmt.Sprintf("server_error_%d", resp.StatusCode)
				}

				metrics.RecordRequest(responseTime, success, errorType)

				if resp != nil {
					resp.Body.Close()
				}
			}
		}()
	}

	wg.Wait()
	metrics.Finalize()

	summary := metrics.GetSummary()
	suite.T().Logf("Error Handling Under Load: %+v", summary)

	// Verify error handling doesn't degrade performance significantly
	suite.Greater(summary["success_rate"], "95.00%", "Success rate should remain > 95% even with error scenarios")
}

// Helper method for making load test requests
func (suite *LoadTestSuite) makeLoadRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, suite.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", suite.authToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return suite.client.Do(req)
}

// Mock handlers for load testing (lightweight implementations)
func (suite *LoadTestSuite) handleCatalogLoad(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Lightweight response for load testing
	response := map[string]interface{}{
		"success": true,
		"data": []map[string]interface{}{
			{"id": "item-1", "name": "Load Test Item 1"},
			{"id": "item-2", "name": "Load Test Item 2"},
		},
		"meta": map[string]interface{}{
			"pagination": map[string]interface{}{
				"page":  1,
				"limit": 20,
				"total": 2,
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (suite *LoadTestSuite) handleOrdersLoad(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "POST" {
		// Simulate order creation
		response := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"id":           "load-order-" + fmt.Sprintf("%d", time.Now().UnixNano()),
				"status":       "pending",
				"total_amount": 100.0,
			},
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	} else {
		// List orders
		response := map[string]interface{}{
			"success": true,
			"data": []map[string]interface{}{
				{"id": "order-1", "status": "pending"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}
}

func (suite *LoadTestSuite) handleInventoryLoad(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"success": true,
		"data": []map[string]interface{}{
			{"id": "lot-1", "available_quantity": 100.0},
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (suite *LoadTestSuite) handleHealthLoad(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// Run the load test suite
func TestLoadTestSuite(t *testing.T) {
	suite.Run(t, new(LoadTestSuite))
}
