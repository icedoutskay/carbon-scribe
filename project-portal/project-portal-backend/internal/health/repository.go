package health

import (
	"context"

	"gorm.io/gorm"
)

// Repository defines the interface for health metrics data access
type Repository interface {
	// System Metric
	CreateSystemMetric(ctx context.Context, metric *SystemMetric) error
	QuerySystemMetrics(ctx context.Context, query MetricQuery) ([]SystemMetric, error)

	// Status
	PingDB(ctx context.Context) error

	// Service
	ListServiceHealthChecks(ctx context.Context) ([]ServiceHealthCheck, error)

	// Check
	CreateServiceHealthCheck(ctx context.Context, check *ServiceHealthCheck) error

	// System Alerts
	GetSystemAlertByID(ctx context.Context, id string) (*SystemAlert, error)
	QuerySystemAlerts(ctx context.Context, query AlertQuery) ([]SystemAlert, error)
	UpdateSystemAlert(ctx context.Context, alert *SystemAlert) error
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

func (r *repository) QuerySystemMetrics(ctx context.Context, query MetricQuery) ([]SystemMetric, error) {
	var metrics []SystemMetric
	db := r.db.WithContext(ctx)

	if query.MetricName != "" {
		db = db.Where("metric_name = ?", query.MetricName)
	}
	if query.MetricType != "" {
		db = db.Where("metric_type = ?", query.MetricType)
	}
	if query.ServiceName != "" {
		db = db.Where("service_name = ?", query.ServiceName)
	}
	if query.Endpoint != "" {
		db = db.Where("endpoint = ?", query.Endpoint)
	}
	if query.InstanceID != "" {
		db = db.Where("instance_id = ?", query.InstanceID)
	}
	if query.Region != "" {
		db = db.Where("region = ?", query.Region)
	}
	if !query.StartTime.IsZero() {
		db = db.Where("time >= ?", query.StartTime)
	}
	if !query.EndTime.IsZero() {
		db = db.Where("time <= ?", query.EndTime)
	}

	err := db.Order("time DESC").Limit(query.Limit).Find(&metrics).Error
	return metrics, err
}

func (r *repository) PingDB(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (r *repository) ListServiceHealthChecks(ctx context.Context) ([]ServiceHealthCheck, error) {
	var checks []ServiceHealthCheck
	err := r.db.Find(&checks).Error
	return checks, err
}

func (r *repository) CreateServiceHealthCheck(ctx context.Context, check *ServiceHealthCheck) error {
	return r.db.Create(check).Error
}

// ========== System alerts ==========

func (r *repository) QuerySystemAlerts(ctx context.Context, query AlertQuery) ([]SystemAlert, error) {
	var alerts []SystemAlert
	db := r.db.WithContext(ctx)

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Severity != "" {
		db = db.Where("alert_severity = ?", query.Severity)
	}
	if query.ServiceName != "" {
		db = db.Where("service_name = ?", query.ServiceName)
	}
	if query.AlertSource != "" {
		db = db.Where("alert_source = ?", query.AlertSource)
	}
	if !query.StartTime.IsZero() {
		db = db.Where("fired_at >= ?", query.StartTime)
	}
	if !query.EndTime.IsZero() {
		db = db.Where("fired_at <= ?", query.EndTime)
	}

	err := db.Order("fired_at DESC").Limit(query.Limit).Find(&alerts).Error
	return alerts, err
}

func (r *repository) UpdateSystemAlert(ctx context.Context, alert *SystemAlert) error {
	return r.db.WithContext(ctx).Save(alert).Error
}

func (r *repository) GetSystemAlertByID(ctx context.Context, id string) (*SystemAlert, error) {
	var alert SystemAlert
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}
