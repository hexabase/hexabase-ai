package monitoring

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/kubernetes"
	"github.com/hexabase/hexabase-ai/api/internal/domain/monitoring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock monitoring repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) SaveMetrics(ctx context.Context, metrics []*monitoring.MetricDataPoint) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func (m *mockRepository) GetMetrics(ctx context.Context, workspaceID string, metricName string, start, end time.Time) ([]*monitoring.MetricDataPoint, error) {
	args := m.Called(ctx, workspaceID, metricName, start, end)
	if args.Get(0) != nil {
		return args.Get(0).([]*monitoring.MetricDataPoint), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) GetLatestMetrics(ctx context.Context, workspaceID string, metricNames []string) (map[string]*monitoring.MetricDataPoint, error) {
	args := m.Called(ctx, workspaceID, metricNames)
	if args.Get(0) != nil {
		return args.Get(0).(map[string]*monitoring.MetricDataPoint), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) DeleteOldMetrics(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *mockRepository) CreateAlert(ctx context.Context, alert *monitoring.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *mockRepository) GetAlert(ctx context.Context, alertID string) (*monitoring.Alert, error) {
	args := m.Called(ctx, alertID)
	if args.Get(0) != nil {
		return args.Get(0).(*monitoring.Alert), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) GetAlerts(ctx context.Context, workspaceID string, filter monitoring.AlertFilter) ([]*monitoring.Alert, error) {
	args := m.Called(ctx, workspaceID, filter)
	if args.Get(0) != nil {
		return args.Get(0).([]*monitoring.Alert), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) UpdateAlert(ctx context.Context, alert *monitoring.Alert) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *mockRepository) DeleteAlert(ctx context.Context, alertID string) error {
	args := m.Called(ctx, alertID)
	return args.Error(0)
}

func (m *mockRepository) SaveHealthCheck(ctx context.Context, health *monitoring.ClusterHealth) error {
	args := m.Called(ctx, health)
	return args.Error(0)
}

func (m *mockRepository) GetLatestHealthCheck(ctx context.Context, workspaceID string) (*monitoring.ClusterHealth, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) != nil {
		return args.Get(0).(*monitoring.ClusterHealth), args.Error(1)
	}
	return nil, args.Error(1)
}

// Mock kubernetes repository
type mockK8sRepository struct {
	mock.Mock
}

func (m *mockK8sRepository) GetNamespace(ctx context.Context, name string) (*kubernetes.Namespace, error) {
	args := m.Called(ctx, name)
	if args.Get(0) != nil {
		return args.Get(0).(*kubernetes.Namespace), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockK8sRepository) CreateNamespace(ctx context.Context, namespace *kubernetes.Namespace) error {
	args := m.Called(ctx, namespace)
	return args.Error(0)
}

func (m *mockK8sRepository) UpdateNamespace(ctx context.Context, namespace *kubernetes.Namespace) error {
	args := m.Called(ctx, namespace)
	return args.Error(0)
}

func (m *mockK8sRepository) DeleteNamespace(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *mockK8sRepository) ListNamespaces(ctx context.Context, labelSelector string) ([]*kubernetes.Namespace, error) {
	args := m.Called(ctx, labelSelector)
	if args.Get(0) != nil {
		return args.Get(0).([]*kubernetes.Namespace), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockK8sRepository) GetResourceQuota(ctx context.Context, namespace, name string) (*kubernetes.ResourceQuota, error) {
	args := m.Called(ctx, namespace, name)
	if args.Get(0) != nil {
		return args.Get(0).(*kubernetes.ResourceQuota), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockK8sRepository) GetNamespaceResourceQuota(ctx context.Context, namespace string) (*kubernetes.ResourceQuota, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) != nil {
		return args.Get(0).(*kubernetes.ResourceQuota), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockK8sRepository) CreateResourceQuota(ctx context.Context, namespace string, quota *kubernetes.ResourceQuota) error {
	args := m.Called(ctx, namespace, quota)
	return args.Error(0)
}

func (m *mockK8sRepository) UpdateResourceQuota(ctx context.Context, namespace string, quota *kubernetes.ResourceQuota) error {
	args := m.Called(ctx, namespace, quota)
	return args.Error(0)
}

func (m *mockK8sRepository) DeleteResourceQuota(ctx context.Context, namespace, name string) error {
	args := m.Called(ctx, namespace, name)
	return args.Error(0)
}

func (m *mockK8sRepository) GetPodMetrics(ctx context.Context, namespace string) (*kubernetes.PodMetricsList, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) != nil {
		return args.Get(0).(*kubernetes.PodMetricsList), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockK8sRepository) CheckComponentHealth(ctx context.Context) (map[string]kubernetes.ComponentStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).(map[string]kubernetes.ComponentStatus), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestService_GetWorkspaceMetrics(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockK8sRepo := new(mockK8sRepository)
	svc := NewService(mockRepo, mockK8sRepo, slog.Default())

	t.Run("successful metrics retrieval", func(t *testing.T) {
		workspaceID := "ws-123"
		opts := monitoring.QueryOptions{
			Period: "1h",
		}

		// Mock CPU metrics
		cpuMetrics := []*monitoring.MetricDataPoint{
			{
				ID:          uuid.New().String(),
				WorkspaceID: workspaceID,
				MetricName:  "cpu_usage",
				Value:       45.5,
				Timestamp:   time.Now().Add(-30 * time.Minute),
			},
			{
				ID:          uuid.New().String(),
				WorkspaceID: workspaceID,
				MetricName:  "cpu_usage",
				Value:       50.0,
				Timestamp:   time.Now(),
			},
		}

		// Mock memory metrics
		memoryMetrics := []*monitoring.MetricDataPoint{
			{
				ID:          uuid.New().String(),
				WorkspaceID: workspaceID,
				MetricName:  "memory_usage",
				Value:       60.0,
				Timestamp:   time.Now().Add(-30 * time.Minute),
			},
			{
				ID:          uuid.New().String(),
				WorkspaceID: workspaceID,
				MetricName:  "memory_usage",
				Value:       65.0,
				Timestamp:   time.Now(),
			},
		}

		// Mock pod count metrics
		podMetrics := []*monitoring.MetricDataPoint{
			{
				ID:          uuid.New().String(),
				WorkspaceID: workspaceID,
				MetricName:  "pod_count",
				Value:       15.0,
				Timestamp:   time.Now(),
			},
		}

		mockRepo.On("GetMetrics", ctx, workspaceID, "cpu_usage", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return(cpuMetrics, nil)
		mockRepo.On("GetMetrics", ctx, workspaceID, "memory_usage", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return(memoryMetrics, nil)
		mockRepo.On("GetMetrics", ctx, workspaceID, "pod_count", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return(podMetrics, nil)

		metrics, err := svc.GetWorkspaceMetrics(ctx, workspaceID, opts)
		assert.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, workspaceID, metrics.WorkspaceID)
		assert.Equal(t, 50.0, metrics.CPUUsage.Current)
		assert.Equal(t, 65.0, metrics.MemoryUsage.Current)
		assert.Equal(t, 15, metrics.PodCount.Current)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		workspaceID := "ws-error"
		opts := monitoring.QueryOptions{Period: "1h"}

		mockRepo.On("GetMetrics", ctx, workspaceID, "cpu_usage", mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
			Return(nil, errors.New("database error"))

		metrics, err := svc.GetWorkspaceMetrics(ctx, workspaceID, opts)
		assert.Error(t, err)
		assert.Nil(t, metrics)

		mockRepo.AssertExpectations(t)
	})
}

func TestService_GetClusterHealth(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockK8sRepo := new(mockK8sRepository)
	svc := NewService(mockRepo, mockK8sRepo, slog.Default())

	t.Run("healthy cluster", func(t *testing.T) {
		workspaceID := "ws-healthy"

		componentStatus := map[string]kubernetes.ComponentStatus{
			"api-server": {
				Name:    "api-server",
				Healthy: true,
				Message: "Running",
			},
			"etcd": {
				Name:    "etcd",
				Healthy: true,
				Message: "Running",
			},
			"scheduler": {
				Name:    "scheduler",
				Healthy: true,
				Message: "Running",
			},
		}

		mockK8sRepo.On("CheckComponentHealth", ctx).Return(componentStatus, nil)
		mockRepo.On("SaveHealthCheck", ctx, mock.AnythingOfType("*monitoring.ClusterHealth")).Return(nil)

		health, err := svc.GetClusterHealth(ctx, workspaceID)
		assert.NoError(t, err)
		assert.NotNil(t, health)
		assert.True(t, health.Healthy)
		assert.Len(t, health.Components, 3)

		mockK8sRepo.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unhealthy cluster", func(t *testing.T) {
		workspaceID := "ws-unhealthy"

		componentStatus := map[string]kubernetes.ComponentStatus{
			"api-server": {
				Name:    "api-server",
				Healthy: true,
				Message: "Running",
			},
			"etcd": {
				Name:    "etcd",
				Healthy: false,
				Message: "Connection failed",
			},
		}

		mockK8sRepo.On("CheckComponentHealth", ctx).Return(componentStatus, nil)
		mockRepo.On("SaveHealthCheck", ctx, mock.AnythingOfType("*monitoring.ClusterHealth")).Return(nil)

		health, err := svc.GetClusterHealth(ctx, workspaceID)
		assert.NoError(t, err)
		assert.NotNil(t, health)
		assert.False(t, health.Healthy)
		assert.Equal(t, "unhealthy", health.Components["etcd"].Status)

		mockK8sRepo.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_GetResourceUsage(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockK8sRepo := new(mockK8sRepository)
	svc := NewService(mockRepo, mockK8sRepo, slog.Default())

	t.Run("get current resource usage", func(t *testing.T) {
		workspaceID := "ws-usage"
		namespace := "vcluster-ws-usage"

		quota := &kubernetes.ResourceQuota{
			Name: "resource-quota",
			Hard: kubernetes.ResourceList{
				"cpu":    "2000m",
				"memory": "8Gi",
			},
			Used: kubernetes.ResourceList{
				"cpu":    "500m",
				"memory": "2Gi",
			},
		}

		podMetrics := &kubernetes.PodMetricsList{
			Items: []kubernetes.PodMetrics{
				{Name: "pod-1"},
				{Name: "pod-2"},
			},
		}

		mockK8sRepo.On("GetNamespaceResourceQuota", ctx, namespace).Return(quota, nil)
		mockK8sRepo.On("GetPodMetrics", ctx, namespace).Return(podMetrics, nil)

		usage, err := svc.GetResourceUsage(ctx, workspaceID)
		assert.NoError(t, err)
		assert.NotNil(t, usage)
		assert.Equal(t, workspaceID, usage.WorkspaceID)
		assert.Equal(t, 2.0, usage.Pods.Used)

		mockK8sRepo.AssertExpectations(t)
	})
}

func TestService_GetAlerts(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockK8sRepo := new(mockK8sRepository)
	svc := NewService(mockRepo, mockK8sRepo, slog.Default())

	t.Run("get active alerts", func(t *testing.T) {
		workspaceID := "ws-alerts"
		severity := "warning"

		expectedAlerts := []*monitoring.Alert{
			{
				ID:          "alert-1",
				WorkspaceID: workspaceID,
				Type:        "cpu",
				Severity:    "warning",
				Title:       "High CPU Usage",
				Status:      "active",
			},
			{
				ID:          "alert-2",
				WorkspaceID: workspaceID,
				Type:        "memory",
				Severity:    "warning",
				Title:       "High Memory Usage",
				Status:      "active",
			},
		}

		expectedFilter := monitoring.AlertFilter{
			Severity: severity,
			Status:   "active",
			Limit:    100,
		}

		mockRepo.On("GetAlerts", ctx, workspaceID, expectedFilter).Return(expectedAlerts, nil)

		alerts, err := svc.GetAlerts(ctx, workspaceID, severity)
		assert.NoError(t, err)
		assert.Len(t, alerts, 2)

		mockRepo.AssertExpectations(t)
	})
}

func TestService_CreateAlert(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockK8sRepo := new(mockK8sRepository)
	svc := NewService(mockRepo, mockK8sRepo, slog.Default())

	t.Run("create custom alert", func(t *testing.T) {
		alert := &monitoring.Alert{
			WorkspaceID: "ws-create",
			Type:        "custom",
			Severity:    "warning",
			Title:       "Custom Alert",
			Description: "User-defined alert rule",
			Threshold:   100.0,
		}

		mockRepo.On("CreateAlert", ctx, mock.AnythingOfType("*monitoring.Alert")).Return(nil)

		err := svc.CreateAlert(ctx, alert)
		assert.NoError(t, err)

		// Verify ID was generated
		createCall := mockRepo.Calls[0]
		createdAlert := createCall.Arguments[1].(*monitoring.Alert)
		assert.NotEmpty(t, createdAlert.ID)
		assert.Equal(t, "active", createdAlert.Status)
		assert.False(t, createdAlert.CreatedAt.IsZero())

		mockRepo.AssertExpectations(t)
	})
}

func TestService_AcknowledgeAlert(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockK8sRepo := new(mockK8sRepository)
	svc := NewService(mockRepo, mockK8sRepo, slog.Default())

	t.Run("acknowledge alert", func(t *testing.T) {
		alertID := "alert-ack"
		userID := "user-123"

		existingAlert := &monitoring.Alert{
			ID:          alertID,
			WorkspaceID: "ws-ack",
			Status:      "active",
		}

		mockRepo.On("GetAlert", ctx, alertID).Return(existingAlert, nil)
		mockRepo.On("UpdateAlert", ctx, mock.AnythingOfType("*monitoring.Alert")).Return(nil)

		err := svc.AcknowledgeAlert(ctx, alertID, userID)
		assert.NoError(t, err)

		// Verify status was updated
		updateCall := mockRepo.Calls[1]
		updatedAlert := updateCall.Arguments[1].(*monitoring.Alert)
		assert.Equal(t, "acknowledged", updatedAlert.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("alert not found", func(t *testing.T) {
		alertID := "non-existent"
		userID := "user-456"

		mockRepo.On("GetAlert", ctx, alertID).Return(nil, errors.New("not found"))

		err := svc.AcknowledgeAlert(ctx, alertID, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		mockRepo.AssertExpectations(t)
	})
}

func TestService_ResolveAlert(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockK8sRepo := new(mockK8sRepository)
	svc := NewService(mockRepo, mockK8sRepo, slog.Default())

	t.Run("resolve alert", func(t *testing.T) {
		alertID := "alert-resolve"

		existingAlert := &monitoring.Alert{
			ID:          alertID,
			WorkspaceID: "ws-resolve",
			Status:      "active",
		}

		mockRepo.On("GetAlert", ctx, alertID).Return(existingAlert, nil)
		mockRepo.On("UpdateAlert", ctx, mock.AnythingOfType("*monitoring.Alert")).Return(nil)

		err := svc.ResolveAlert(ctx, alertID)
		assert.NoError(t, err)

		// Verify status and timestamp were updated
		updateCall := mockRepo.Calls[1]
		updatedAlert := updateCall.Arguments[1].(*monitoring.Alert)
		assert.Equal(t, "resolved", updatedAlert.Status)
		assert.NotNil(t, updatedAlert.ResolvedAt)

		mockRepo.AssertExpectations(t)
	})
}

func TestService_CollectMetrics(t *testing.T) {
	ctx := context.Background()
	
	mockRepo := new(mockRepository)
	mockK8sRepo := new(mockK8sRepository)
	svc := NewService(mockRepo, mockK8sRepo, slog.Default())

	t.Run("collect and save metrics", func(t *testing.T) {
		workspaceID := "ws-collect"
		namespace := "vcluster-ws-collect"

		podMetrics := &kubernetes.PodMetricsList{
			Items: []kubernetes.PodMetrics{
				{Name: "pod-1"},
				{Name: "pod-2"},
				{Name: "pod-3"},
			},
		}

		mockK8sRepo.On("GetPodMetrics", ctx, namespace).Return(podMetrics, nil)
		mockRepo.On("SaveMetrics", ctx, mock.AnythingOfType("[]*monitoring.MetricDataPoint")).Return(nil)

		err := svc.CollectMetrics(ctx, workspaceID)
		assert.NoError(t, err)

		// Verify metrics were saved
		saveCall := mockRepo.Calls[0]
		dataPoints := saveCall.Arguments[1].([]*monitoring.MetricDataPoint)
		assert.NotEmpty(t, dataPoints)

		mockK8sRepo.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}