package health

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// SystemMetric represents a time-series metric for system health monitoring
// This model is designed to work with TimescaleDB hypertables.
type SystemMetric struct {
	Time       time.Time `gorm:"type:timestamptz;not null;primaryKey" json:"time"`
	MetricName string    `gorm:"type:varchar(255);not null;index:idx_system_metrics_name_time,priority:1" json:"metric_name"`
	MetricType string    `gorm:"type:varchar(50);not null" json:"metric_type"` // gauge, counter, histogram, summary

	// Labels/dimensions for the metric
	ServiceName    string `gorm:"type:varchar(100);index:idx_system_metrics_service,priority:1" json:"service_name,omitempty"`
	Endpoint       string `gorm:"type:varchar(500)" json:"endpoint,omitempty"`
	HTTPMethod     string `gorm:"type:varchar(10)" json:"http_method,omitempty"`
	HTTPStatusCode int    `gorm:"type:integer" json:"http_status_code,omitempty"`
	InstanceID     string `gorm:"type:varchar(100)" json:"instance_id,omitempty"`
	Region         string `gorm:"type:varchar(50)" json:"region,omitempty"`

	// Metric values
	Value        float64         `gorm:"type:double precision;not null" json:"value"`
	Count        int             `gorm:"type:integer" json:"count,omitempty"`
	BucketBounds pq.Float64Array `gorm:"type:double precision[]" json:"bucket_bounds,omitempty"`

	// Additional metadata
	Labels   datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"labels,omitempty"`
	Metadata datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata,omitempty"`
}

// TableName specifies the table name for GORM
func (SystemMetric) TableName() string {
	return "system_metrics"
}

// ========== Request/Response Types ==========

// CreateSystemMetricRequest represents the request to create a system metric
type CreateSystemMetricRequest struct {
	MetricName  string `json:"metric_name" binding:"required"`
	MetricType  string `json:"metric_type" binding:"required"` // gauge, counter, histogram, summary
	ServiceName string `json:"service_name,omitempty"`

	// Labels/dimensions for the metric
	Endpoint       string `json:"endpoint,omitempty"`
	HTTPMethod     string `json:"http_method,omitempty"`
	HTTPStatusCode int    `json:"http_status_code,omitempty"`
	InstanceID     string `json:"instance_id,omitempty"`
	Region         string `json:"region,omitempty"`

	// Metric values
	Value        float64   `json:"value" binding:"required"`
	Count        int       `json:"count,omitempty"`
	BucketBounds []float64 `json:"bucket_bounds,omitempty"`

	// Additional metadata
	Labels   map[string]any `json:"labels,omitempty" swaggertype:"object"`
	Metadata map[string]any `json:"metadata,omitempty" swaggertype:"object"`
}
