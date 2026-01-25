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
	GetServicesHealth(ctx context.Context) ([]ServiceHealthInfo, error)
	CreateServiceHealthCheck(ctx context.Context, req CreateServiceHealthCheckRequest) (*ServiceHealthCheck, error)
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

func (s *service) GetServicesHealth(ctx context.Context) ([]ServiceHealthInfo, error) {
	checks, err := s.repo.ListServiceHealthChecks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list service health checks: %w", err)
	}

	var servicesHealth []ServiceHealthInfo
	for _, check := range checks {
		status := "healthy"
		if check.ConsecutiveFailures >= check.AlertThresholdFailures {
			status = "unhealthy"
		} else if check.ConsecutiveFailures > 0 {
			status = "degraded"
		}

		servicesHealth = append(servicesHealth, ServiceHealthInfo{
			ServiceName:            check.ServiceName,
			Status:                 status,
			CheckType:              check.CheckType,
			LastCheck:              check.LastCheckTime,
			Failures:               check.ConsecutiveFailures,
			IntervalSeconds:        check.IntervalSeconds,
			TimeoutSeconds:         check.TimeoutSeconds,
			AlertThresholdFailures: check.AlertThresholdFailures,
			AlertSeverity:          check.AlertSeverity,
		})
	}

	return servicesHealth, nil
}

func (s *service) CreateServiceHealthCheck(ctx context.Context, req CreateServiceHealthCheckRequest) (*ServiceHealthCheck, error) {
	checkConfigJSON, err := json.Marshal(req.CheckConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize service health check: %w", err)
	}

	check := &ServiceHealthCheck{
		ServiceName:            req.ServiceName,
		CheckType:              req.CheckType,
		CheckConfig:            datatypes.JSON(checkConfigJSON),
		IntervalSeconds:        req.IntervalSeconds,
		TimeoutSeconds:         req.TimeoutSeconds,
		AlertOnFailure:         req.AlertOnFailure,
		AlertThresholdFailures: req.AlertThresholdFailures,
		AlertSeverity:          req.AlertSeverity,
		IsEnabled:              req.IsEnabled,
	}

	// Set defaults
	if check.IntervalSeconds == 0 {
		check.IntervalSeconds = 60
	}
	if check.TimeoutSeconds == 0 {
		check.TimeoutSeconds = 10
	}
	if check.AlertThresholdFailures == 0 {
		check.AlertThresholdFailures = 3
	}
	if check.AlertSeverity == "" {
		check.AlertSeverity = "critical"
	}

	if err := s.repo.CreateServiceHealthCheck(ctx, check); err != nil {
		return nil, fmt.Errorf("failed to create health check: %w", err)
	}

	return check, nil
}
