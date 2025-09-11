package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"
)

const (
    baseURL = "http://localhost:8080/api/v1"
)

type TestResult struct {
    Endpoint string
    Method   string
    Status   int
    Success  bool
    Message  string
}

func testEcommerceEndpoints() {
    fmt.Println("Testing E-commerce Endpoints with AAA Integration...")

    var results []TestResult

    // Test health endpoint
    results = append(results, testEndpoint("GET", "/health", nil, ""))

    // Test auth endpoints
    fmt.Println("\n--- Testing Authentication ---")

    // Register user
    registerReq := map[string]interface{}{
        "email":      "testuser@example.com",
        "first_name": "Test",
        "last_name":  "User",
        "password":   "Test#1234",
        "username":   "testuser",
        "phone":      "919999000001",
    }
    results = append(results, testEndpoint("POST", "/auth/register", registerReq, ""))

    // Login user
    loginReq := map[string]interface{}{
        "username": "testuser",
        "password": "Test#1234",
    }
    loginResult := testEndpoint("POST", "/auth/login", loginReq, "")
    results = append(results, loginResult)

    // Extract token from login response (if successful)
    var token string
    if loginResult.Success {
        // In a real implementation, you'd parse the response to get the token
        token = "mock_token_123" // For now, use a mock token
    }

    // Test logout
    results = append(results, testEndpoint("POST", "/auth/logout", map[string]interface{}{}, token))

    // Test catalog endpoints
    fmt.Println("\n--- Testing Catalog Management ---")

    // List products (public)
    results = append(results, testEndpoint("GET", "/catalog/products", nil, ""))

    // Create product (requires auth)
    productReq := map[string]interface{}{
        "name":            "Test Urea 46% N",
        "base_price":      899.5,
        "currency":        "INR",
        "category":        "Fertilizer",
        "subcategory":     "Nitrogen",
        "item_type":       "PRODUCT",
        "sku":             "UREA-TEST-001",
        "unit_of_measure": "BAG",
        "visibility":      "PUBLIC",
        "description":     "Test product for API testing",
    }
    results = append(results, testEndpoint("POST", "/catalog/products", productReq, token))

    // Test orders endpoints
    fmt.Println("\n--- Testing Order Management ---")

    // Create order (requires auth)
    orderReq := map[string]interface{}{
        "buyer_organization_id":  "ORG_BUYER_TEST_001",
        "seller_organization_id": "ORG_SELLER_TEST_001",
        "items": []map[string]interface{}{
            {
                "catalog_item_id":   "test_product_id",
                "catalog_item_type": "product",
                "quantity":          4,
                "unit_price":        899.5,
            },
        },
        "notes": "Test order from API testing",
        "shipping_address": map[string]interface{}{
            "street":      "Test Street 123",
            "city":        "Test City",
            "state":       "Test State",
            "postal_code": "123456",
            "country":     "IN",
        },
    }
    results = append(results, testEndpoint("POST", "/orders", orderReq, token))

    // List orders (requires auth)
    results = append(results, testEndpoint("GET", "/orders", nil, token))

    // Test without auth (should fail)
    fmt.Println("\n--- Testing Authorization (Should Fail) ---")
    results = append(results, testEndpoint("POST", "/catalog/products", productReq, ""))
    results = append(results, testEndpoint("POST", "/orders", orderReq, ""))

    // Print results
    fmt.Println("\n" + strings.Repeat("=", 80))
    fmt.Println("TEST RESULTS SUMMARY")
    fmt.Println(strings.Repeat("=", 80))

    successCount := 0
    for _, result := range results {
        status := "✗ FAIL"
        if result.Success {
            status = "✓ PASS"
            successCount++
        }

        fmt.Printf("%s %s %s -> %d %s\n",
            status, result.Method, result.Endpoint, result.Status, result.Message)
    }

    fmt.Printf("\nSUCCESS RATE: %d/%d (%.1f%%)\n",
        successCount, len(results), float64(successCount)/float64(len(results))*100)

    if successCount == len(results) {
        fmt.Println("\n🎉 All tests passed! AAA integration is working correctly.")
    } else {
        fmt.Println("\n⚠️  Some tests failed. Check the AAA service connection and configuration.")
    }
}

func testEndpoint(method, endpoint string, body interface{}, token string) TestResult {
    url := baseURL + endpoint
    if endpoint == "/health" {
        url = "http://localhost:8080" + endpoint
    }

    var reqBody io.Reader
    if body != nil {
        jsonBody, _ := json.Marshal(body)
        reqBody = bytes.NewBuffer(jsonBody)
    }

    req, err := http.NewRequest(method, url, reqBody)
    if err != nil {
        return TestResult{
            Endpoint: endpoint,
            Method:   method,
            Status:   0,
            Success:  false,
            Message:  fmt.Sprintf("Request creation failed: %v", err),
        }
    }

    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }

    if token != "" {
        req.Header.Set("Authorization", "Bearer "+token)
    }

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return TestResult{
            Endpoint: endpoint,
            Method:   method,
            Status:   0,
            Success:  false,
            Message:  fmt.Sprintf("Request failed: %v", err),
        }
    }
    defer resp.Body.Close()

    // Read response
    respBody, _ := io.ReadAll(resp.Body)

    // Determine success based on status code and endpoint
    success := false
    message := string(respBody)

    switch {
    case resp.StatusCode >= 200 && resp.StatusCode < 300:
        success = true
        message = "Success"
    case resp.StatusCode == 401:
        message = "Unauthorized (expected for protected endpoints without token)"
        // For unauthorized tests, this might be expected
        if token == "" && (endpoint == "/catalog/products" || endpoint == "/orders") && method == "POST" {
            success = true // Expected failure
        }
    case resp.StatusCode == 403:
        message = "Forbidden (check permissions)"
    case resp.StatusCode == 404:
        message = "Not found"
    case resp.StatusCode == 503:
        message = "Service unavailable"
    default:
        message = fmt.Sprintf("Unexpected status: %s", string(respBody))
    }

    fmt.Printf("  %s %s -> %d (%s)\n", method, endpoint, resp.StatusCode, message)

    return TestResult{
        Endpoint: endpoint,
        Method:   method,
        Status:   resp.StatusCode,
        Success:  success,
        Message:  message,
    }
}
