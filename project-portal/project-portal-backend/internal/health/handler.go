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
