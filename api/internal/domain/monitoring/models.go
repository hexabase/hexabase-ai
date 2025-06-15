package monitoring

import (
	"time"
)

// WorkspaceMetrics represents aggregated metrics for a workspace
type WorkspaceMetrics struct {
	WorkspaceID   string                 `json:"workspace_id"`
	Period        string                 `json:"period"`
	CPUUsage      *ResourceMetric        `json:"cpu_usage"`
	MemoryUsage   *ResourceMetric        `json:"memory_usage"`
	PodCount      *CountMetric           `json:"pod_count"`
	NetworkIO     *IOMetric              `json:"network_io"`
	DiskIO        *IOMetric              `json:"disk_io"`
	Timestamps    []time.Time            `json:"timestamps"`
	CustomMetrics map[string]interface{} `json:"custom_metrics,omitempty" gorm:"type:jsonb"`
}

// ResourceMetric represents CPU or Memory usage metrics
type ResourceMetric struct {
	Current     float64   `json:"current"`
	Average     float64   `json:"average"`
	Peak        float64   `json:"peak"`
	Percentile95 float64   `json:"percentile_95"`
	History     []float64 `json:"history"`
	Unit        string    `json:"unit"`
}

// CountMetric represents countable metrics like pods
type CountMetric struct {
	Current int   `json:"current"`
	Average float64 `json:"average"`
	Peak    int   `json:"peak"`
	History []int `json:"history"`
}

// IOMetric represents input/output metrics
type IOMetric struct {
	ReadRate  float64   `json:"read_rate"`
	WriteRate float64   `json:"write_rate"`
	History   []float64 `json:"history"`
	Unit      string    `json:"unit"`
}

// ClusterHealth represents the health status of a vCluster
type ClusterHealth struct {
	WorkspaceID string                    `json:"workspace_id"`
	Healthy     bool                      `json:"healthy"`
	Components  map[string]ComponentHealth `json:"components"`
	LastChecked time.Time                 `json:"last_checked"`
}

// ComponentHealth represents health of a single component
type ComponentHealth struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // healthy, degraded, unhealthy
	Message string `json:"message,omitempty"`
}

// ResourceUsage represents current resource usage
type ResourceUsage struct {
	WorkspaceID string             `json:"workspace_id"`
	CPU         ResourceUsageDetail `json:"cpu"`
	Memory      ResourceUsageDetail `json:"memory"`
	Storage     ResourceUsageDetail `json:"storage"`
	Pods        ResourceUsageDetail `json:"pods"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// ResourceUsageDetail contains usage details for a specific resource
type ResourceUsageDetail struct {
	Used      float64 `json:"used"`
	Limit     float64 `json:"limit"`
	Requested float64 `json:"requested"`
	Unit      string  `json:"unit"`
}

// Alert represents a monitoring alert
type Alert struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"` // critical, warning, info
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Resource    string    `json:"resource,omitempty"`
	Threshold   float64   `json:"threshold,omitempty"`
	Value       float64   `json:"value,omitempty"`
	Status      string    `json:"status"` // active, resolved, acknowledged
	CreatedAt   time.Time `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// MetricDataPoint represents a single metric measurement
type MetricDataPoint struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	MetricName  string    `json:"metric_name"`
	Value       float64   `json:"value"`
	Labels      map[string]string `json:"labels"`
	Timestamp   time.Time `json:"timestamp"`
}

// QueryOptions represents options for querying metrics
type QueryOptions struct {
	Period    string    // 1h, 6h, 1d, 7d, 30d
	StartTime time.Time
	EndTime   time.Time
	Step      time.Duration
}