package health_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"carbon-scribe/project-portal/project-portal-backend/internal/health"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	v1 := router.Group("/api/v1")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/carbonscribe?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	repo := health.NewRepository(db)
	service := health.NewService(repo)
	handler := health.NewHandler(service)
	handler.RegisterRoutes(v1)

	// Clean up database tables for isolation
	db.Exec("TRUNCATE TABLE system_metrics, service_health_checks, health_check_results RESTART IDENTITY CASCADE")

	return router, db
}

func TestMetricsResource(t *testing.T) {
	router, _ := setupTestRouter(t)

	// POST /metrics
	reqBody := health.CreateSystemMetricRequest{
		MetricName: "cpu_usage_refactor",
		MetricType: "gauge",
		Value:      55.5,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/health/metrics", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// GET /metrics
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/health/metrics?metric_name=cpu_usage_refactor", nil)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	var getResponse []health.SystemMetric
	json.Unmarshal(w2.Body.Bytes(), &getResponse)
	assert.NotEmpty(t, getResponse)
}

func TestStatusResource(t *testing.T) {
	router, _ := setupTestRouter(t)

	// GET /status
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var statusResponse health.SystemStatusResponse
	json.Unmarshal(w.Body.Bytes(), &statusResponse)
	assert.Equal(t, "healthy", statusResponse.Status)
	assert.Equal(t, "carbon-scribe-project-portal", statusResponse.Service)
	assert.Equal(t, "1.0.0", statusResponse.Version)
	assert.NotZero(t, statusResponse.Timestamp)

	// GET /status/detailed
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/health/status/detailed", nil)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	var detailedResponse health.DetailedStatusResponse
	json.Unmarshal(w2.Body.Bytes(), &detailedResponse)
	assert.Equal(t, "healthy", detailedResponse.Status)
	assert.Equal(t, "carbon-scribe-project-portal", detailedResponse.Service)
	assert.Equal(t, "1.0.0", detailedResponse.Version)
	assert.NotZero(t, detailedResponse.Timestamp)
	assert.NotEmpty(t, detailedResponse.Uptime)
	assert.Contains(t, detailedResponse.Components, "database")
	assert.Equal(t, "up", detailedResponse.Components["database"].Status)
}

func TestServicesResource(t *testing.T) {
	router, db := setupTestRouter(t)

	// Seed a test service health check
	testCheck := health.ServiceHealthCheck{
		ServiceName:            "test-external-api",
		CheckType:              "http",
		CheckConfig:            []byte(`{"url": "http://example.com"}`),
		IntervalSeconds:        30,
		TimeoutSeconds:         5,
		ConsecutiveFailures:    0,
		AlertThresholdFailures: 3,
		IsEnabled:              true,
	}
	db.Create(&testCheck)

	// Execute GET /services
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health/services", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response health.ServiceHealthResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response.Services)

	found := false
	for _, s := range response.Services {
		if s.ServiceName == "test-external-api" {
			found = true
			assert.Equal(t, "healthy", s.Status)
			break
		}
	}
	assert.True(t, found, "test-external-api not found in services health response")
}

func TestChecksResource(t *testing.T) {
	router, _ := setupTestRouter(t)

	// POST /checks
	reqBody := health.CreateServiceHealthCheckRequest{
		ServiceName:            "new-service-check",
		CheckType:              "http",
		CheckConfig:            map[string]any{"url": "http://new-service.com"},
		IntervalSeconds:        45,
		TimeoutSeconds:         15,
		AlertThresholdFailures: 5,
		AlertSeverity:          "warning",
		IsEnabled:              true,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/health/checks", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var createdCheck health.ServiceHealthCheck
	json.Unmarshal(w.Body.Bytes(), &createdCheck)
	assert.NotEmpty(t, createdCheck.ID)
	assert.Equal(t, "new-service-check", createdCheck.ServiceName)
	assert.Equal(t, "http", createdCheck.CheckType)
	assert.Equal(t, 45, createdCheck.IntervalSeconds)
	assert.Equal(t, 15, createdCheck.TimeoutSeconds)
	assert.Equal(t, 5, createdCheck.AlertThresholdFailures)
	assert.Equal(t, "warning", createdCheck.AlertSeverity)

	// Verify with GET /services
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/health/services", nil)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	var response health.ServiceHealthResponse
	json.Unmarshal(w2.Body.Bytes(), &response)

	found := false
	for _, s := range response.Services {
		if s.ServiceName == "new-service-check" {
			found = true
			assert.Equal(t, "healthy", s.Status)
			assert.Equal(t, "http", s.CheckType)
			assert.Equal(t, 0, s.Failures)
			assert.Equal(t, 45, s.IntervalSeconds)
			assert.Equal(t, 15, s.TimeoutSeconds)
			assert.Equal(t, 5, s.AlertThresholdFailures)
			assert.Equal(t, "warning", s.AlertSeverity)
			break
		}
	}
	assert.True(t, found, "new-service-check not found in services health response after creation")
}
