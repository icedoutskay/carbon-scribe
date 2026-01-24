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

func TestCreateSystemMetric(t *testing.T) {
	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	v1 := router.Group("/api/v1")

	// Load database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/carbonscribe?sslmode=disable"
	}

	// Initialize database connection
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize real dependencies
	repo := health.NewRepository(db)
	service := health.NewService(repo)
	handler := health.NewHandler(service)
	handler.RegisterRoutes(v1)

	// Prepare payload
	reqBody := health.CreateSystemMetricRequest{
		MetricName: "cpu_usage_test",
		MetricType: "gauge",
		Value:      45.5,
	}
	body, _ := json.Marshal(reqBody)

	// Execute request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/health/metrics", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)

	var response health.SystemMetric
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "cpu_usage_test", response.MetricName)
}
