package application

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApplicationType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		appType  ApplicationType
		expected bool
	}{
		{
			name:     "valid stateless type",
			appType:  ApplicationTypeStateless,
			expected: true,
		},
		{
			name:     "valid stateful type",
			appType:  ApplicationTypeStateful,
			expected: true,
		},
		{
			name:     "invalid type",
			appType:  ApplicationType("invalid"),
			expected: false,
		},
		{
			name:     "empty type",
			appType:  ApplicationType(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appType.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSourceType_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		sourceType SourceType
		expected   bool
	}{
		{
			name:       "valid image source",
			sourceType: SourceTypeImage,
			expected:   true,
		},
		{
			name:       "valid git source",
			sourceType: SourceTypeGit,
			expected:   true,
		},
		{
			name:       "invalid source",
			sourceType: SourceType("invalid"),
			expected:   false,
		},
		{
			name:       "empty source",
			sourceType: SourceType(""),
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.sourceType.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplicationStatus_CanTransition(t *testing.T) {
	tests := []struct {
		name         string
		currentStatus ApplicationStatus
		targetStatus  ApplicationStatus
		canTransition bool
	}{
		// Pending transitions
		{
			name:          "pending to deploying",
			currentStatus: ApplicationStatusPending,
			targetStatus:  ApplicationStatusDeploying,
			canTransition: true,
		},
		{
			name:          "pending to error",
			currentStatus: ApplicationStatusPending,
			targetStatus:  ApplicationStatusError,
			canTransition: true,
		},
		{
			name:          "pending to running (invalid)",
			currentStatus: ApplicationStatusPending,
			targetStatus:  ApplicationStatusRunning,
			canTransition: false,
		},
		// Deploying transitions
		{
			name:          "deploying to running",
			currentStatus: ApplicationStatusDeploying,
			targetStatus:  ApplicationStatusRunning,
			canTransition: true,
		},
		{
			name:          "deploying to error",
			currentStatus: ApplicationStatusDeploying,
			targetStatus:  ApplicationStatusError,
			canTransition: true,
		},
		{
			name:          "deploying to stopped (invalid)",
			currentStatus: ApplicationStatusDeploying,
			targetStatus:  ApplicationStatusStopped,
			canTransition: false,
		},
		// Running transitions
		{
			name:          "running to updating",
			currentStatus: ApplicationStatusRunning,
			targetStatus:  ApplicationStatusUpdating,
			canTransition: true,
		},
		{
			name:          "running to stopping",
			currentStatus: ApplicationStatusRunning,
			targetStatus:  ApplicationStatusStopping,
			canTransition: true,
		},
		{
			name:          "running to error",
			currentStatus: ApplicationStatusRunning,
			targetStatus:  ApplicationStatusError,
			canTransition: true,
		},
		{
			name:          "running to stopped (invalid)",
			currentStatus: ApplicationStatusRunning,
			targetStatus:  ApplicationStatusStopped,
			canTransition: false,
		},
		// Updating transitions
		{
			name:          "updating to running",
			currentStatus: ApplicationStatusUpdating,
			targetStatus:  ApplicationStatusRunning,
			canTransition: true,
		},
		{
			name:          "updating to error",
			currentStatus: ApplicationStatusUpdating,
			targetStatus:  ApplicationStatusError,
			canTransition: true,
		},
		// Stopping transitions
		{
			name:          "stopping to stopped",
			currentStatus: ApplicationStatusStopping,
			targetStatus:  ApplicationStatusStopped,
			canTransition: true,
		},
		{
			name:          "stopping to error",
			currentStatus: ApplicationStatusStopping,
			targetStatus:  ApplicationStatusError,
			canTransition: true,
		},
		// Stopped transitions
		{
			name:          "stopped to deploying",
			currentStatus: ApplicationStatusStopped,
			targetStatus:  ApplicationStatusDeploying,
			canTransition: true,
		},
		{
			name:          "stopped to deleting",
			currentStatus: ApplicationStatusStopped,
			targetStatus:  ApplicationStatusDeleting,
			canTransition: true,
		},
		{
			name:          "stopped to running (invalid)",
			currentStatus: ApplicationStatusStopped,
			targetStatus:  ApplicationStatusRunning,
			canTransition: false,
		},
		// Error transitions
		{
			name:          "error to deploying",
			currentStatus: ApplicationStatusError,
			targetStatus:  ApplicationStatusDeploying,
			canTransition: true,
		},
		{
			name:          "error to deleting",
			currentStatus: ApplicationStatusError,
			targetStatus:  ApplicationStatusDeleting,
			canTransition: true,
		},
		// Deleting transitions (terminal state)
		{
			name:          "deleting to any state (invalid)",
			currentStatus: ApplicationStatusDeleting,
			targetStatus:  ApplicationStatusRunning,
			canTransition: false,
		},
		// Invalid status
		{
			name:          "invalid status",
			currentStatus: ApplicationStatus("invalid"),
			targetStatus:  ApplicationStatusRunning,
			canTransition: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.currentStatus.CanTransition(tt.targetStatus)
			assert.Equal(t, tt.canTransition, result)
		})
	}
}

func TestApplication_Structure(t *testing.T) {
	now := time.Now()
	app := &Application{
		ID:          "app-123",
		WorkspaceID: "ws-123",
		ProjectID:   "proj-123",
		Name:        "test-app",
		Type:        ApplicationTypeStateless,
		Status:      ApplicationStatusRunning,
		Source: ApplicationSource{
			Type:  SourceTypeImage,
			Image: "nginx:latest",
		},
		Config: ApplicationConfig{
			Replicas: 3,
			Port:     80,
			EnvVars: map[string]string{
				"ENV": "production",
			},
			Resources: ResourceRequests{
				CPURequest:    "100m",
				CPULimit:      "500m",
				MemoryRequest: "128Mi",
				MemoryLimit:   "512Mi",
			},
			NodeSelector: map[string]string{
				"node-type": "dedicated",
			},
		},
		Endpoints: []Endpoint{
			{
				Type: "ingress",
				URL:  "https://test-app.example.com",
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, "app-123", app.ID)
	assert.Equal(t, "ws-123", app.WorkspaceID)
	assert.Equal(t, "proj-123", app.ProjectID)
	assert.Equal(t, "test-app", app.Name)
	assert.Equal(t, ApplicationTypeStateless, app.Type)
	assert.Equal(t, ApplicationStatusRunning, app.Status)
	assert.Equal(t, SourceTypeImage, app.Source.Type)
	assert.Equal(t, "nginx:latest", app.Source.Image)
	assert.Equal(t, 3, app.Config.Replicas)
	assert.Equal(t, 80, app.Config.Port)
	assert.Equal(t, "production", app.Config.EnvVars["ENV"])
	assert.Equal(t, "100m", app.Config.Resources.CPURequest)
	assert.Equal(t, "dedicated", app.Config.NodeSelector["node-type"])
	assert.Len(t, app.Endpoints, 1)
	assert.Equal(t, "https://test-app.example.com", app.Endpoints[0].URL)
}

func TestApplicationSource_GitSource(t *testing.T) {
	source := ApplicationSource{
		Type:      SourceTypeGit,
		GitURL:    "https://github.com/example/app.git",
		GitRef:    "main",
		Buildpack: "nodejs",
	}

	assert.Equal(t, SourceTypeGit, source.Type)
	assert.Equal(t, "https://github.com/example/app.git", source.GitURL)
	assert.Equal(t, "main", source.GitRef)
	assert.Equal(t, "nodejs", source.Buildpack)
	assert.Empty(t, source.Image)
}

func TestStorageConfig(t *testing.T) {
	storage := &StorageConfig{
		Size:         "10Gi",
		StorageClass: "ssd",
		MountPath:    "/data",
	}

	assert.Equal(t, "10Gi", storage.Size)
	assert.Equal(t, "ssd", storage.StorageClass)
	assert.Equal(t, "/data", storage.MountPath)
}

func TestNetworkConfig(t *testing.T) {
	network := &NetworkConfig{
		CreateIngress: true,
		IngressPath:   "/api",
		CustomDomain:  "api.example.com",
		TLSEnabled:    true,
		Annotations: map[string]string{
			"nginx.ingress.kubernetes.io/rewrite-target": "/",
		},
	}

	assert.True(t, network.CreateIngress)
	assert.Equal(t, "/api", network.IngressPath)
	assert.Equal(t, "api.example.com", network.CustomDomain)
	assert.True(t, network.TLSEnabled)
	assert.Equal(t, "/", network.Annotations["nginx.ingress.kubernetes.io/rewrite-target"])
}

func TestPod_Structure(t *testing.T) {
	pod := Pod{
		Name:      "test-app-abc123",
		Status:    "Running",
		NodeName:  "node-1",
		IP:        "10.0.0.1",
		StartTime: time.Now(),
		Restarts:  0,
	}

	assert.Equal(t, "test-app-abc123", pod.Name)
	assert.Equal(t, "Running", pod.Status)
	assert.Equal(t, "node-1", pod.NodeName)
	assert.Equal(t, "10.0.0.1", pod.IP)
	assert.Equal(t, 0, pod.Restarts)
}

func TestApplicationMetrics(t *testing.T) {
	metrics := &ApplicationMetrics{
		ApplicationID: "app-123",
		Timestamp:     time.Now(),
		PodMetrics: []PodMetrics{
			{
				PodName:     "app-123-abc",
				CPUUsage:    0.5,
				MemoryUsage: 256.0,
				NetworkIn:   1.5,
				NetworkOut:  2.0,
			},
			{
				PodName:     "app-123-def",
				CPUUsage:    0.3,
				MemoryUsage: 200.0,
				NetworkIn:   1.0,
				NetworkOut:  1.5,
			},
		},
		AggregateUsage: AggregateResourceUsage{
			TotalCPU:      0.8,
			TotalMemory:   456.0,
			AverageCPU:    0.4,
			AverageMemory: 228.0,
		},
	}

	assert.Equal(t, "app-123", metrics.ApplicationID)
	assert.Len(t, metrics.PodMetrics, 2)
	assert.Equal(t, 0.5, metrics.PodMetrics[0].CPUUsage)
	assert.Equal(t, 0.8, metrics.AggregateUsage.TotalCPU)
	assert.Equal(t, 0.4, metrics.AggregateUsage.AverageCPU)
}

func TestLogQuery(t *testing.T) {
	now := time.Now()
	query := LogQuery{
		ApplicationID: "app-123",
		PodName:       "app-123-abc",
		Container:     "main",
		Since:         now.Add(-1 * time.Hour),
		Until:         now,
		Limit:         100,
		Follow:        false,
	}

	assert.Equal(t, "app-123", query.ApplicationID)
	assert.Equal(t, "app-123-abc", query.PodName)
	assert.Equal(t, "main", query.Container)
	assert.Equal(t, 100, query.Limit)
	assert.False(t, query.Follow)
}

func TestCreateApplicationRequest(t *testing.T) {
	req := CreateApplicationRequest{
		Name: "new-app",
		Type: ApplicationTypeStateful,
		Source: ApplicationSource{
			Type:  SourceTypeImage,
			Image: "postgres:14",
		},
		Config: ApplicationConfig{
			Replicas: 1,
			Port:     5432,
			EnvVars: map[string]string{
				"POSTGRES_DB":       "mydb",
				"POSTGRES_USER":     "user",
				"POSTGRES_PASSWORD": "pass",
			},
			Resources: ResourceRequests{
				CPURequest:    "250m",
				CPULimit:      "1000m",
				MemoryRequest: "512Mi",
				MemoryLimit:   "2Gi",
			},
			Storage: &StorageConfig{
				Size:         "20Gi",
				StorageClass: "standard",
				MountPath:    "/var/lib/postgresql/data",
			},
		},
		ProjectID:  "proj-123",
		NodePoolID: "pool-dedicated",
	}

	assert.Equal(t, "new-app", req.Name)
	assert.Equal(t, ApplicationTypeStateful, req.Type)
	assert.Equal(t, "postgres:14", req.Source.Image)
	assert.Equal(t, 1, req.Config.Replicas)
	assert.NotNil(t, req.Config.Storage)
	assert.Equal(t, "20Gi", req.Config.Storage.Size)
	assert.Equal(t, "pool-dedicated", req.NodePoolID)
}

func TestUpdateApplicationRequest(t *testing.T) {
	replicas := 5
	req := UpdateApplicationRequest{
		Replicas:     &replicas,
		ImageVersion: "nginx:1.21",
		EnvVars: map[string]string{
			"NEW_VAR": "value",
		},
		Resources: &ResourceRequests{
			CPURequest:    "200m",
			CPULimit:      "1000m",
			MemoryRequest: "256Mi",
			MemoryLimit:   "1Gi",
		},
		NetworkConfig: &NetworkConfig{
			CreateIngress: true,
			CustomDomain:  "new.example.com",
			TLSEnabled:    true,
		},
	}

	assert.NotNil(t, req.Replicas)
	assert.Equal(t, 5, *req.Replicas)
	assert.Equal(t, "nginx:1.21", req.ImageVersion)
	assert.Equal(t, "value", req.EnvVars["NEW_VAR"])
	assert.NotNil(t, req.Resources)
	assert.Equal(t, "200m", req.Resources.CPURequest)
	assert.NotNil(t, req.NetworkConfig)
	assert.Equal(t, "new.example.com", req.NetworkConfig.CustomDomain)
}