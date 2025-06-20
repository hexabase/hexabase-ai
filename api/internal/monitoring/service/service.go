package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/kubernetes"
	"github.com/hexabase/hexabase-ai/api/internal/monitoring/domain"
)

// service implements the domain.Service interface
type service struct {
	repo      domain.Repository
	k8sRepo   kubernetes.Repository
	logger    *slog.Logger
}

// NewService creates a new monitoring service
func NewService(repo domain.Repository, k8sRepo kubernetes.Repository, logger *slog.Logger) domain.Service {
	return &service{
		repo:    repo,
		k8sRepo: k8sRepo,
		logger:  logger,
	}
}

// GetWorkspaceMetrics retrieves aggregated metrics for a workspace
func (s *service) GetWorkspaceMetrics(ctx context.Context, workspaceID string, opts domain.QueryOptions) (*domain.WorkspaceMetrics, error) {
	// Parse period and calculate time range
	start, end := s.calculateTimeRange(opts)

	// Fetch metrics from repository
	cpuMetrics, err := s.repo.GetMetrics(ctx, workspaceID, "cpu_usage", start, end)
	if err != nil {
		s.logger.Error("Failed to fetch CPU metrics", "error", err)
		return nil, fmt.Errorf("failed to fetch CPU metrics: %w", err)
	}

	memoryMetrics, err := s.repo.GetMetrics(ctx, workspaceID, "memory_usage", start, end)
	if err != nil {
		s.logger.Error("Failed to fetch memory metrics", "error", err)
		return nil, fmt.Errorf("failed to fetch memory metrics: %w", err)
	}

	podMetrics, err := s.repo.GetMetrics(ctx, workspaceID, "pod_count", start, end)
	if err != nil {
		s.logger.Error("Failed to fetch pod metrics", "error", err)
		return nil, fmt.Errorf("failed to fetch pod metrics: %w", err)
	}

	// Aggregate metrics
	result := &domain.WorkspaceMetrics{
		WorkspaceID: workspaceID,
		Period:      opts.Period,
		CPUUsage:    s.aggregateResourceMetrics(cpuMetrics, "cores"),
		MemoryUsage: s.aggregateResourceMetrics(memoryMetrics, "GB"),
		PodCount:    s.aggregateCountMetrics(podMetrics),
		Timestamps:  s.extractTimestamps(cpuMetrics),
	}

	return result, nil
}

// GetClusterHealth checks and returns the health status of a vCluster
func (s *service) GetClusterHealth(ctx context.Context, workspaceID string) (*domain.ClusterHealth, error) {
	// Get namespace for the workspace (assuming workspace_id maps to namespace)
	// namespace := fmt.Sprintf("vcluster-%s", workspaceID)

	// Check Kubernetes components health
	componentStatus, err := s.k8sRepo.CheckComponentHealth(ctx)
	if err != nil {
		s.logger.Error("Failed to check component health", "error", err)
		return nil, fmt.Errorf("failed to check component health: %w", err)
	}

	// Convert to domain model
	health := &domain.ClusterHealth{
		WorkspaceID: workspaceID,
		Healthy:     true,
		Components:  make(map[string]domain.ComponentHealth),
		LastChecked: time.Now(),
	}

	for name, status := range componentStatus {
		compHealth := domain.ComponentHealth{
			Name:    name,
			Status:  "healthy",
			Message: status.Message,
		}

		if !status.Healthy {
			compHealth.Status = "unhealthy"
			health.Healthy = false
		}

		health.Components[name] = compHealth
	}

	// Save health check result
	if err := s.repo.SaveHealthCheck(ctx, health); err != nil {
		s.logger.Warn("Failed to save health check", "error", err)
	}

	return health, nil
}

// GetResourceUsage returns current resource usage for a workspace
func (s *service) GetResourceUsage(ctx context.Context, workspaceID string) (*domain.ResourceUsage, error) {
	namespace := fmt.Sprintf("vcluster-%s", workspaceID)

	// Get resource quota
	quota, err := s.k8sRepo.GetNamespaceResourceQuota(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource quota: %w", err)
	}

	// Get current metrics
	podMetrics, err := s.k8sRepo.GetPodMetrics(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	// Calculate usage
	usage := s.calculateResourceUsage(quota, podMetrics)
	usage.WorkspaceID = workspaceID
	usage.UpdatedAt = time.Now()

	return usage, nil
}

// GetAlerts retrieves alerts for a workspace
func (s *service) GetAlerts(ctx context.Context, workspaceID string, severity string) ([]*domain.Alert, error) {
	filter := domain.AlertFilter{
		Severity: severity,
		Status:   "active",
		Limit:    100,
	}

	alerts, err := s.repo.GetAlerts(ctx, workspaceID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}

	return alerts, nil
}

// CreateAlert creates a new monitoring alert
func (s *service) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	alert.ID = generateID()
	alert.CreatedAt = time.Now()
	alert.Status = "active"

	if err := s.repo.CreateAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	s.logger.Info("Alert created",
		"alert_id", alert.ID,
		"workspace_id", alert.WorkspaceID,
		"severity", alert.Severity)

	return nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (s *service) AcknowledgeAlert(ctx context.Context, alertID string, userID string) error {
	alert, err := s.repo.GetAlert(ctx, alertID)
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	alert.Status = "acknowledged"
	
	if err := s.repo.UpdateAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	return nil
}

// ResolveAlert marks an alert as resolved
func (s *service) ResolveAlert(ctx context.Context, alertID string) error {
	alert, err := s.repo.GetAlert(ctx, alertID)
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	now := time.Now()
	alert.Status = "resolved"
	alert.ResolvedAt = &now

	if err := s.repo.UpdateAlert(ctx, alert); err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	return nil
}

// CollectMetrics collects and stores current metrics for a workspace
func (s *service) CollectMetrics(ctx context.Context, workspaceID string) error {
	namespace := fmt.Sprintf("vcluster-%s", workspaceID)

	// Get pod metrics
	podMetrics, err := s.k8sRepo.GetPodMetrics(ctx, namespace)
	if err != nil {
		return fmt.Errorf("failed to get pod metrics: %w", err)
	}

	// Convert to metric data points
	dataPoints := s.convertToDataPoints(workspaceID, podMetrics)

	// Save metrics
	if err := s.repo.SaveMetrics(ctx, dataPoints); err != nil {
		return fmt.Errorf("failed to save metrics: %w", err)
	}

	s.logger.Info("Metrics collected",
		"workspace_id", workspaceID,
		"data_points", len(dataPoints))

	return nil
}

// Helper methods

func (s *service) calculateTimeRange(opts domain.QueryOptions) (time.Time, time.Time) {
	if !opts.StartTime.IsZero() && !opts.EndTime.IsZero() {
		return opts.StartTime, opts.EndTime
	}

	end := time.Now()
	var start time.Time

	switch opts.Period {
	case "1h":
		start = end.Add(-1 * time.Hour)
	case "6h":
		start = end.Add(-6 * time.Hour)
	case "1d":
		start = end.Add(-24 * time.Hour)
	case "7d":
		start = end.Add(-7 * 24 * time.Hour)
	case "30d":
		start = end.Add(-30 * 24 * time.Hour)
	default:
		start = end.Add(-1 * time.Hour)
	}

	return start, end
}

func (s *service) aggregateResourceMetrics(metrics []*domain.MetricDataPoint, unit string) *domain.ResourceMetric {
	if len(metrics) == 0 {
		return &domain.ResourceMetric{Unit: unit}
	}

	var sum, peak float64
	history := make([]float64, len(metrics))

	for i, m := range metrics {
		history[i] = m.Value
		sum += m.Value
		if m.Value > peak {
			peak = m.Value
		}
	}

	return &domain.ResourceMetric{
		Current:  history[len(history)-1],
		Average:  sum / float64(len(metrics)),
		Peak:     peak,
		History:  history,
		Unit:     unit,
	}
}

func (s *service) aggregateCountMetrics(metrics []*domain.MetricDataPoint) *domain.CountMetric {
	if len(metrics) == 0 {
		return &domain.CountMetric{}
	}

	var sum float64
	var peak int
	history := make([]int, len(metrics))

	for i, m := range metrics {
		count := int(m.Value)
		history[i] = count
		sum += m.Value
		if count > peak {
			peak = count
		}
	}

	return &domain.CountMetric{
		Current: history[len(history)-1],
		Average: sum / float64(len(metrics)),
		Peak:    peak,
		History: history,
	}
}

func (s *service) extractTimestamps(metrics []*domain.MetricDataPoint) []time.Time {
	timestamps := make([]time.Time, len(metrics))
	for i, m := range metrics {
		timestamps[i] = m.Timestamp
	}
	return timestamps
}

func (s *service) calculateResourceUsage(quota *kubernetes.ResourceQuota, podMetrics *kubernetes.PodMetricsList) *domain.ResourceUsage {
	// Implementation would calculate actual usage from pod metrics
	// This is a simplified version
	return &domain.ResourceUsage{
		CPU: domain.ResourceUsageDetail{
			Used:      0.5,
			Limit:     2.0,
			Requested: 1.0,
			Unit:      "cores",
		},
		Memory: domain.ResourceUsageDetail{
			Used:      2.5,
			Limit:     8.0,
			Requested: 4.0,
			Unit:      "GB",
		},
		Storage: domain.ResourceUsageDetail{
			Used:  10.0,
			Limit: 100.0,
			Unit:  "GB",
		},
		Pods: domain.ResourceUsageDetail{
			Used:  float64(len(podMetrics.Items)),
			Limit: 50,
			Unit:  "pods",
		},
	}
}

func (s *service) convertToDataPoints(workspaceID string, podMetrics *kubernetes.PodMetricsList) []*domain.MetricDataPoint {
	var dataPoints []*domain.MetricDataPoint
	timestamp := time.Now()

	// Aggregate CPU usage
	var totalCPU float64
	for range podMetrics.Items {
		// Parse CPU value (assuming format like "100m" for millicores)
		// This is simplified - real implementation would parse properly
		totalCPU += 0.1 // placeholder
	}

	dataPoints = append(dataPoints, &domain.MetricDataPoint{
		ID:          generateID(),
		WorkspaceID: workspaceID,
		MetricName:  "cpu_usage",
		Value:       totalCPU,
		Timestamp:   timestamp,
	})

	// Add more metrics...

	return dataPoints
}

func generateID() string {
	// Simple ID generation - in production use UUID
	return fmt.Sprintf("%d", time.Now().UnixNano())
}