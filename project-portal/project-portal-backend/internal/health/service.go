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
}

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
