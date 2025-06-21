package service

import (
	"context"
	"io"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/application/domain"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateApplication(ctx context.Context, app *domain.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockRepository) GetApplication(ctx context.Context, id string) (*domain.Application, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Application), args.Error(1)
}

func (m *MockRepository) GetApplicationByName(ctx context.Context, workspaceID, projectID, name string) (*domain.Application, error) {
	args := m.Called(ctx, workspaceID, projectID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Application), args.Error(1)
}

func (m *MockRepository) ListApplications(ctx context.Context, workspaceID, projectID string) ([]domain.Application, error) {
	args := m.Called(ctx, workspaceID, projectID)
	return args.Get(0).([]domain.Application), args.Error(1)
}

func (m *MockRepository) UpdateApplication(ctx context.Context, app *domain.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockRepository) DeleteApplication(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) CreateEvent(ctx context.Context, event *domain.ApplicationEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockRepository) ListEvents(ctx context.Context, applicationID string, limit int) ([]domain.ApplicationEvent, error) {
	args := m.Called(ctx, applicationID, limit)
	return args.Get(0).([]domain.ApplicationEvent), args.Error(1)
}

func (m *MockRepository) GetApplicationsByNode(ctx context.Context, nodeID string) ([]domain.Application, error) {
	args := m.Called(ctx, nodeID)
	return args.Get(0).([]domain.Application), args.Error(1)
}

func (m *MockRepository) GetApplicationsByStatus(ctx context.Context, workspaceID string, status domain.ApplicationStatus) ([]domain.Application, error) {
	args := m.Called(ctx, workspaceID, status)
	return args.Get(0).([]domain.Application), args.Error(1)
}

func (m *MockRepository) Create(ctx context.Context, app *domain.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockRepository) GetCronJobExecutions(ctx context.Context, applicationID string, limit, offset int) ([]domain.CronJobExecution, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]domain.CronJobExecution), args.Int(1), args.Error(2)
}

func (m *MockRepository) CreateCronJobExecution(ctx context.Context, execution *domain.CronJobExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockRepository) UpdateCronJobExecution(ctx context.Context, executionID string, completedAt *time.Time, status domain.CronJobExecutionStatus, exitCode *int, logs string) error {
	args := m.Called(ctx, executionID, completedAt, status, exitCode, logs)
	return args.Error(0)
}

func (m *MockRepository) UpdateCronSchedule(ctx context.Context, applicationID, schedule string) error {
	args := m.Called(ctx, applicationID, schedule)
	return args.Error(0)
}

func (m *MockRepository) GetCronJobExecutionByID(ctx context.Context, executionID string) (*domain.CronJobExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CronJobExecution), args.Error(1)
}

// MockKubernetesRepository is a mock implementation of the KubernetesRepository interface
type MockKubernetesRepository struct {
	mock.Mock
}

func (m *MockKubernetesRepository) CreateDeployment(ctx context.Context, workspaceID, projectID string, spec domain.DeploymentSpec) error {
	args := m.Called(ctx, workspaceID, projectID, spec)
	return args.Error(0)
}

func (m *MockKubernetesRepository) UpdateDeployment(ctx context.Context, workspaceID, projectID, name string, spec domain.DeploymentSpec) error {
	args := m.Called(ctx, workspaceID, projectID, name, spec)
	return args.Error(0)
}

func (m *MockKubernetesRepository) DeleteDeployment(ctx context.Context, workspaceID, projectID, name string) error {
	args := m.Called(ctx, workspaceID, projectID, name)
	return args.Error(0)
}

func (m *MockKubernetesRepository) GetDeploymentStatus(ctx context.Context, workspaceID, projectID, name string) (*domain.DeploymentStatus, error) {
	args := m.Called(ctx, workspaceID, projectID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DeploymentStatus), args.Error(1)
}

func (m *MockKubernetesRepository) CreateStatefulSet(ctx context.Context, workspaceID, projectID string, spec domain.StatefulSetSpec) error {
	args := m.Called(ctx, workspaceID, projectID, spec)
	return args.Error(0)
}

func (m *MockKubernetesRepository) UpdateStatefulSet(ctx context.Context, workspaceID, projectID, name string, spec domain.StatefulSetSpec) error {
	args := m.Called(ctx, workspaceID, projectID, name, spec)
	return args.Error(0)
}

func (m *MockKubernetesRepository) DeleteStatefulSet(ctx context.Context, workspaceID, projectID, name string) error {
	args := m.Called(ctx, workspaceID, projectID, name)
	return args.Error(0)
}

func (m *MockKubernetesRepository) GetStatefulSetStatus(ctx context.Context, workspaceID, projectID, name string) (*domain.StatefulSetStatus, error) {
	args := m.Called(ctx, workspaceID, projectID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.StatefulSetStatus), args.Error(1)
}

func (m *MockKubernetesRepository) CreateService(ctx context.Context, workspaceID, projectID string, spec domain.ServiceSpec) error {
	args := m.Called(ctx, workspaceID, projectID, spec)
	return args.Error(0)
}

func (m *MockKubernetesRepository) DeleteService(ctx context.Context, workspaceID, projectID, name string) error {
	args := m.Called(ctx, workspaceID, projectID, name)
	return args.Error(0)
}

func (m *MockKubernetesRepository) GetServiceEndpoints(ctx context.Context, workspaceID, projectID, name string) ([]domain.Endpoint, error) {
	args := m.Called(ctx, workspaceID, projectID, name)
	return args.Get(0).([]domain.Endpoint), args.Error(1)
}

func (m *MockKubernetesRepository) CreateIngress(ctx context.Context, workspaceID, projectID string, spec domain.IngressSpec) error {
	args := m.Called(ctx, workspaceID, projectID, spec)
	return args.Error(0)
}

func (m *MockKubernetesRepository) UpdateIngress(ctx context.Context, workspaceID, projectID, name string, spec domain.IngressSpec) error {
	args := m.Called(ctx, workspaceID, projectID, name, spec)
	return args.Error(0)
}

func (m *MockKubernetesRepository) DeleteIngress(ctx context.Context, workspaceID, projectID, name string) error {
	args := m.Called(ctx, workspaceID, projectID, name)
	return args.Error(0)
}

func (m *MockKubernetesRepository) CreatePVC(ctx context.Context, workspaceID, projectID string, spec domain.PVCSpec) error {
	args := m.Called(ctx, workspaceID, projectID, spec)
	return args.Error(0)
}

func (m *MockKubernetesRepository) DeletePVC(ctx context.Context, workspaceID, projectID, name string) error {
	args := m.Called(ctx, workspaceID, projectID, name)
	return args.Error(0)
}

func (m *MockKubernetesRepository) ListPods(ctx context.Context, workspaceID, projectID string, selector map[string]string) ([]domain.Pod, error) {
	args := m.Called(ctx, workspaceID, projectID, selector)
	return args.Get(0).([]domain.Pod), args.Error(1)
}

func (m *MockKubernetesRepository) GetPodLogs(ctx context.Context, workspaceID, projectID, podName, container string, opts domain.LogOptions) ([]domain.LogEntry, error) {
	args := m.Called(ctx, workspaceID, projectID, podName, container, opts)
	return args.Get(0).([]domain.LogEntry), args.Error(1)
}

func (m *MockKubernetesRepository) StreamPodLogs(ctx context.Context, workspaceID, projectID, podName, container string, opts domain.LogOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, workspaceID, projectID, podName, container, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockKubernetesRepository) RestartPod(ctx context.Context, workspaceID, projectID, podName string) error {
	args := m.Called(ctx, workspaceID, projectID, podName)
	return args.Error(0)
}

func (m *MockKubernetesRepository) GetPodMetrics(ctx context.Context, workspaceID, projectID string, podNames []string) ([]domain.PodMetrics, error) {
	args := m.Called(ctx, workspaceID, projectID, podNames)
	return args.Get(0).([]domain.PodMetrics), args.Error(1)
}

func (m *MockKubernetesRepository) CreateCronJob(ctx context.Context, workspaceID, projectID string, spec domain.CronJobSpec) error {
	args := m.Called(ctx, workspaceID, projectID, spec)
	return args.Error(0)
}

func (m *MockKubernetesRepository) UpdateCronJob(ctx context.Context, workspaceID, projectID, name string, spec domain.CronJobSpec) error {
	args := m.Called(ctx, workspaceID, projectID, name, spec)
	return args.Error(0)
}

func (m *MockKubernetesRepository) DeleteCronJob(ctx context.Context, workspaceID, projectID, name string) error {
	args := m.Called(ctx, workspaceID, projectID, name)
	return args.Error(0)
}

func (m *MockKubernetesRepository) GetCronJobStatus(ctx context.Context, workspaceID, projectID, name string) (*domain.CronJobStatus, error) {
	args := m.Called(ctx, workspaceID, projectID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CronJobStatus), args.Error(1)
}

func (m *MockKubernetesRepository) TriggerCronJob(ctx context.Context, workspaceID, projectID, name string) error {
	args := m.Called(ctx, workspaceID, projectID, name)
	return args.Error(0)
}

// Function-specific methods for MockRepository
func (m *MockRepository) CreateFunctionVersion(ctx context.Context, version *domain.FunctionVersion) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *MockRepository) GetFunctionVersion(ctx context.Context, versionID string) (*domain.FunctionVersion, error) {
	args := m.Called(ctx, versionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FunctionVersion), args.Error(1)
}

func (m *MockRepository) GetFunctionVersions(ctx context.Context, applicationID string) ([]domain.FunctionVersion, error) {
	args := m.Called(ctx, applicationID)
	return args.Get(0).([]domain.FunctionVersion), args.Error(1)
}

func (m *MockRepository) GetActiveFunctionVersion(ctx context.Context, applicationID string) (*domain.FunctionVersion, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FunctionVersion), args.Error(1)
}

func (m *MockRepository) UpdateFunctionVersion(ctx context.Context, version *domain.FunctionVersion) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *MockRepository) SetActiveFunctionVersion(ctx context.Context, applicationID, versionID string) error {
	args := m.Called(ctx, applicationID, versionID)
	return args.Error(0)
}

func (m *MockRepository) CreateFunctionInvocation(ctx context.Context, invocation *domain.FunctionInvocation) error {
	args := m.Called(ctx, invocation)
	return args.Error(0)
}

func (m *MockRepository) GetFunctionInvocation(ctx context.Context, invocationID string) (*domain.FunctionInvocation, error) {
	args := m.Called(ctx, invocationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FunctionInvocation), args.Error(1)
}

func (m *MockRepository) GetFunctionInvocations(ctx context.Context, applicationID string, limit, offset int) ([]domain.FunctionInvocation, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]domain.FunctionInvocation), args.Int(1), args.Error(2)
}

func (m *MockRepository) UpdateFunctionInvocation(ctx context.Context, invocation *domain.FunctionInvocation) error {
	args := m.Called(ctx, invocation)
	return args.Error(0)
}

func (m *MockRepository) CreateFunctionEvent(ctx context.Context, event *domain.FunctionEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockRepository) GetFunctionEvent(ctx context.Context, eventID string) (*domain.FunctionEvent, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FunctionEvent), args.Error(1)
}

func (m *MockRepository) GetPendingFunctionEvents(ctx context.Context, applicationID string, limit int) ([]domain.FunctionEvent, error) {
	args := m.Called(ctx, applicationID, limit)
	return args.Get(0).([]domain.FunctionEvent), args.Error(1)
}

func (m *MockRepository) UpdateFunctionEvent(ctx context.Context, event *domain.FunctionEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// Knative-specific methods for MockKubernetesRepository
func (m *MockKubernetesRepository) CreateKnativeService(ctx context.Context, workspaceID, projectID string, spec domain.KnativeServiceSpec) error {
	args := m.Called(ctx, workspaceID, projectID, spec)
	return args.Error(0)
}

func (m *MockKubernetesRepository) UpdateKnativeService(ctx context.Context, workspaceID, projectID, name string, spec domain.KnativeServiceSpec) error {
	args := m.Called(ctx, workspaceID, projectID, name, spec)
	return args.Error(0)
}

func (m *MockKubernetesRepository) DeleteKnativeService(ctx context.Context, workspaceID, projectID, name string) error {
	args := m.Called(ctx, workspaceID, projectID, name)
	return args.Error(0)
}

func (m *MockKubernetesRepository) GetKnativeServiceStatus(ctx context.Context, workspaceID, projectID, name string) (*domain.KnativeServiceStatus, error) {
	args := m.Called(ctx, workspaceID, projectID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.KnativeServiceStatus), args.Error(1)
}

func (m *MockKubernetesRepository) GetKnativeServiceURL(ctx context.Context, workspaceID, projectID, name string) (string, error) {
	args := m.Called(ctx, workspaceID, projectID, name)
	return args.String(0), args.Error(1)
}

// Add missing method to MockRepository
func (m *MockRepository) GetCronJobExecution(ctx context.Context, executionID string) (*domain.CronJobExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CronJobExecution), args.Error(1)
}