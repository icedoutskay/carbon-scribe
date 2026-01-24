package health

import (
	"context"

	"gorm.io/gorm"
)

// Repository defines the interface for health metrics data access
type Repository interface {
	// System Metric
	CreateSystemMetric(ctx context.Context, metric *SystemMetric) error
}

// repository implements the Repository interface
type repository struct {
	db *gorm.DB
}

// NewRepository creates a new health metrics repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ========== System metrics ==========

func (r *repository) CreateSystemMetric(ctx context.Context, metric *SystemMetric) error {
	return r.db.WithContext(ctx).Create(metric).Error
}
