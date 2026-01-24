package health

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"
)

// Service defines the interface for health business logic
type Service interface {
	CreateSystemMetric(ctx context.Context, req CreateSystemMetricRequest) (*SystemMetric, error)
	GetSystemMetrics(ctx context.Context, query MetricQuery) ([]SystemMetric, error)
	GetStatus(ctx context.Context) (SystemStatusResponse, error)
	GetDetailedStatus(ctx context.Context) (DetailedStatusResponse, error)
}

const defaultServiceName = "carbon-scribe-project-portal"
const defaultVersion = "1.0.0"

var startTime = time.Now()

// service implements the Service interface
type service struct {
	repo Repository
}

// NewService creates a new health service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// ========== Health Metrics Definitions ==========

func (s *service) CreateSystemMetric(ctx context.Context, req CreateSystemMetricRequest) (*SystemMetric, error) {
	systemMetric := &SystemMetric{
		Time:           time.Now(),
		MetricName:     req.MetricName,
		MetricType:     req.MetricType,
		ServiceName:    req.ServiceName,
		Endpoint:       req.Endpoint,
		HTTPMethod:     req.HTTPMethod,
		HTTPStatusCode: req.HTTPStatusCode,
		InstanceID:     req.InstanceID,
		Region:         req.Region,
		Value:          req.Value,
		Count:          req.Count,
		BucketBounds:   req.BucketBounds,
	}

	// Simple conversion for JSON fields if provided
	if req.Labels != nil {
		labelsJSON, _ := json.Marshal(req.Labels)
		systemMetric.Labels = datatypes.JSON(labelsJSON)
	}
	if req.Metadata != nil {
		metadataJSON, _ := json.Marshal(req.Metadata)
		systemMetric.Metadata = datatypes.JSON(metadataJSON)
	}

	if err := s.repo.CreateSystemMetric(ctx, systemMetric); err != nil {
		return nil, fmt.Errorf("failed to create metric: %w", err)
	}

	return systemMetric, nil
}

func (s *service) GetSystemMetrics(ctx context.Context, query MetricQuery) ([]SystemMetric, error) {
	return s.repo.QuerySystemMetrics(ctx, query)
}

func (s *service) GetStatus(ctx context.Context) (SystemStatusResponse, error) {
	status := "healthy"
	if err := s.repo.PingDB(ctx); err != nil {
		status = "unhealthy"
	}

	return SystemStatusResponse{
		Status:    status,
		Service:   defaultServiceName,
		Timestamp: time.Now(),
		Version:   defaultVersion,
	}, nil
}

func (s *service) GetDetailedStatus(ctx context.Context) (DetailedStatusResponse, error) {
	dbStatus := "up"
	dbError := ""
	start := time.Now()
	if err := s.repo.PingDB(ctx); err != nil {
		dbStatus = "down"
		dbError = err.Error()
	}
	dbLatency := time.Since(start).Milliseconds()

	overallStatus := "healthy"
	if dbStatus == "down" {
		overallStatus = "unhealthy"
	}

	uptime := time.Since(startTime).String()

	return DetailedStatusResponse{
		Status:    overallStatus,
		Service:   defaultServiceName,
		Timestamp: time.Now(),
		Version:   defaultVersion,
		Uptime:    uptime,
		Components: map[string]ComponentStatus{
			"database": {
				Status:        dbStatus,
				Details:       dbError,
				LatencyMs:     dbLatency,
				LastCheckTime: time.Now(),
			},
		},
	}, nil
}
