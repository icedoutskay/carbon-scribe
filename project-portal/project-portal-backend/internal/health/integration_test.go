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
