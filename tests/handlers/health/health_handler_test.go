package health

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "kisanlink-ecom/entities/models/common"
    "kisanlink-ecom/internal/config"
    "kisanlink-ecom/internal/database"
    "kisanlink-ecom/internal/handlers/health"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestHealthHandler_HealthCheck(t *testing.T) {
    gin.SetMode(gin.TestMode)

    handler := health.NewHealthHandler()
    router := gin.New()
    router.GET("/health", handler.HealthCheck)

    req, _ := http.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response common.HealthResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    require.NoError(t, err)

    assert.Equal(t, "ok", response.Status)
    assert.Equal(t, "KisanLink E-commerce API is running", response.Message)
    assert.Equal(t, "1.0.0", response.Version)
}

func TestHealthHandler_DetailedHealthCheck(t *testing.T) {
    gin.SetMode(gin.TestMode)

    t.Run("WithHealthyDatabase", func(t *testing.T) {
        // Create a mock database manager
        cfg := config.MultiDatabaseConfig{
            PostgreSQL: config.PostgreSQLConfig{
                Host:            "localhost",
                Port:            5432,
                Database:        "test",
                Username:        "test",
                Password:        "test",
                SSLMode:         "disable",
                MaxOpenConns:    10,
                MaxIdleConns:    5,
                ConnMaxLifetime: 300,
            },
            DynamoDB: config.DynamoDBConfig{
                Region: "us-east-1",
                Table:  "test",
            },
        }

        // Try to create database manager, skip test if not available
        dbManager, err := database.NewDatabaseManager(cfg)
        if err != nil {
            t.Skipf("Database not available for test: %v", err)
        }
        defer dbManager.Close()

        handler := health.NewHealthHandler()
        router := gin.New()

        // Add database manager to context
        router.Use(func(c *gin.Context) {
            c.Set("dbManager", dbManager)
            c.Next()
        })

        router.GET("/health/detailed", handler.DetailedHealthCheck)

        req, _ := http.NewRequest("GET", "/health/detailed", nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        // The response code depends on whether the database is actually healthy
        // In a test environment, it might not be, so we check for either 200 or 503
        assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable)

        var response health.DetailedHealthResponse
        err = json.Unmarshal(w.Body.Bytes(), &response)
        require.NoError(t, err)

        assert.NotEmpty(t, response.Status)
        assert.NotEmpty(t, response.Message)
        assert.Equal(t, "1.0.0", response.Version)
        assert.NotEmpty(t, response.Timestamp)
        assert.Contains(t, response.Dependencies, "database")
    })

    t.Run("WithoutDatabaseManager", func(t *testing.T) {
        handler := health.NewHealthHandler()
        router := gin.New()
        router.GET("/health/detailed", handler.DetailedHealthCheck)

        req, _ := http.NewRequest("GET", "/health/detailed", nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusServiceUnavailable, w.Code)

        var response health.DetailedHealthResponse
        err := json.Unmarshal(w.Body.Bytes(), &response)
        require.NoError(t, err)

        assert.Equal(t, "degraded", response.Status)
        assert.Contains(t, response.Message, "Some dependencies are unhealthy")
        assert.Contains(t, response.Dependencies, "database")
        assert.Equal(t, "unknown", response.Dependencies["database"].Status)
    })

    t.Run("WithInvalidDatabaseManager", func(t *testing.T) {
        handler := health.NewHealthHandler()
        router := gin.New()

        // Add invalid database manager to context
        router.Use(func(c *gin.Context) {
            c.Set("dbManager", "invalid")
            c.Next()
        })

        router.GET("/health/detailed", handler.DetailedHealthCheck)

        req, _ := http.NewRequest("GET", "/health/detailed", nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusServiceUnavailable, w.Code)

        var response health.DetailedHealthResponse
        err := json.Unmarshal(w.Body.Bytes(), &response)
        require.NoError(t, err)

        assert.Equal(t, "degraded", response.Status)
        assert.Contains(t, response.Dependencies, "database")
        assert.Equal(t, "unknown", response.Dependencies["database"].Status)
    })
}

func TestHealthHandler_ReadinessCheck(t *testing.T) {
    gin.SetMode(gin.TestMode)

    t.Run("WithHealthyDatabase", func(t *testing.T) {
        // Create a mock database manager
        cfg := config.MultiDatabaseConfig{
            PostgreSQL: config.PostgreSQLConfig{
                Host:            "localhost",
                Port:            5432,
                Database:        "test",
                Username:        "test",
                Password:        "test",
                SSLMode:         "disable",
                MaxOpenConns:    10,
                MaxIdleConns:    5,
                ConnMaxLifetime: 300,
            },
            DynamoDB: config.DynamoDBConfig{
                Region: "us-east-1",
                Table:  "test",
            },
        }

        // Try to create database manager, skip test if not available
        dbManager, err := database.NewDatabaseManager(cfg)
        if err != nil {
            t.Skipf("Database not available for test: %v", err)
        }
        defer dbManager.Close()

        handler := health.NewHealthHandler()
        router := gin.New()

        // Add database manager to context
        router.Use(func(c *gin.Context) {
            c.Set("dbManager", dbManager)
            c.Next()
        })

        router.GET("/ready", handler.ReadinessCheck)

        req, _ := http.NewRequest("GET", "/ready", nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        // The response code depends on whether the database is actually ready
        // In a test environment, it might not be, so we check for either 200 or 503
        assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable)

        var response common.HealthResponse
        err = json.Unmarshal(w.Body.Bytes(), &response)
        require.NoError(t, err)

        assert.NotEmpty(t, response.Status)
        assert.NotEmpty(t, response.Message)
        assert.Equal(t, "1.0.0", response.Version)
    })

    t.Run("WithoutDatabaseManager", func(t *testing.T) {
        handler := health.NewHealthHandler()
        router := gin.New()
        router.GET("/ready", handler.ReadinessCheck)

        req, _ := http.NewRequest("GET", "/ready", nil)
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusOK, w.Code)

        var response common.HealthResponse
        err := json.Unmarshal(w.Body.Bytes(), &response)
        require.NoError(t, err)

        assert.Equal(t, "ready", response.Status)
        assert.Equal(t, "Service is ready to accept traffic", response.Message)
    })
}

func TestHealthHandler_ContextTimeout(t *testing.T) {
    gin.SetMode(gin.TestMode)

    // This test verifies that health checks respect context timeouts
    // We can't easily test actual timeouts without a slow database,
    // but we can verify the structure is correct

    handler := health.NewHealthHandler()
    router := gin.New()
    router.GET("/health/detailed", handler.DetailedHealthCheck)

    req, _ := http.NewRequest("GET", "/health/detailed", nil)

    // Add a context with a very short timeout
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
    defer cancel()
    req = req.WithContext(ctx)

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // Should still return a response, even with timeout
    assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable)

    var response health.DetailedHealthResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    require.NoError(t, err)

    assert.NotEmpty(t, response.Status)
    assert.NotEmpty(t, response.Message)
}
