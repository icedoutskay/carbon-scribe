package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for the health module
type Handler struct {
	service Service
}

// NewHandler creates a new health handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all report routes with the Gin router
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	reports := router.Group("/health")
	{
		reports.POST("/metrics", h.CreateSystemMetric)
		reports.GET("/metrics", h.GetSystemMetrics)
		reports.GET("/status", h.GetSystemStatus)
		reports.GET("/status/detailed", h.GetDetailedStatus)
	}
}

// ========== Health metrics ==========

// CreateSystemMetric creates a new health metric
// @Summary Create a new health metric
// @Description Create a new health metric with custom configuration
// @Tags health
// @Accept json
// @Produce json
// @Param request body CreateSystemMetricRequest true "System Metric configuration"
// @Success 201 {object} SystemMetric
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/health/metrics [post]
func (h *Handler) CreateSystemMetric(c *gin.Context) {
	var req CreateSystemMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metric, err := h.service.CreateSystemMetric(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, metric)
}

// GetSystemMetrics queries system metrics
// @Summary Query system metrics
// @Description Query system metrics with filtering support
// @Tags health
// @Accept json
// @Produce json
// @Param metric_name query string false "Metric name"
// @Param metric_type query string false "Metric type"
// @Param service_name query string false "Service name"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Param limit query int false "Limit"
// @Success 200 {array} SystemMetric
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/health/metrics [get]
func (h *Handler) GetSystemMetrics(c *gin.Context) {
	var query MetricQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metrics, err := h.service.GetSystemMetrics(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetSystemStatus returns basic system status
// @Summary Get basic system status
// @Description Get basic system status (healthy/degraded/unhealthy)
// @Tags health
// @Produce json
// @Success 200 {object} SystemStatusResponse
// @Router /api/v1/health/status [get]
func (h *Handler) GetSystemStatus(c *gin.Context) {
	status, err := h.service.GetStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetDetailedStatus returns detailed system status
// @Summary Get detailed system status
// @Description Get detailed system status with component details
// @Tags health
// @Produce json
// @Success 200 {object} DetailedStatusResponse
// @Router /api/v1/health/status/detailed [get]
func (h *Handler) GetDetailedStatus(c *gin.Context) {
	status, err := h.service.GetDetailedStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}
