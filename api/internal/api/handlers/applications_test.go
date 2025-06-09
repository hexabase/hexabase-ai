package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockApplicationService is a mock implementation of the application service
type MockApplicationService struct {
	mock.Mock
}

func (m *MockApplicationService) CreateApplication(ctx context.Context, workspaceID string, req application.CreateApplicationRequest) (*application.Application, error) {
	args := m.Called(ctx, workspaceID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.Application), args.Error(1)
}

func (m *MockApplicationService) GetApplication(ctx context.Context, applicationID string) (*application.Application, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.Application), args.Error(1)
}

func (m *MockApplicationService) ListApplications(ctx context.Context, workspaceID, projectID string) ([]application.Application, error) {
	args := m.Called(ctx, workspaceID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]application.Application), args.Error(1)
}

func (m *MockApplicationService) UpdateApplication(ctx context.Context, applicationID string, req application.UpdateApplicationRequest) (*application.Application, error) {
	args := m.Called(ctx, applicationID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.Application), args.Error(1)
}

func (m *MockApplicationService) DeleteApplication(ctx context.Context, applicationID string) error {
	args := m.Called(ctx, applicationID)
	return args.Error(0)
}

func (m *MockApplicationService) StartApplication(ctx context.Context, applicationID string) error {
	args := m.Called(ctx, applicationID)
	return args.Error(0)
}

func (m *MockApplicationService) StopApplication(ctx context.Context, applicationID string) error {
	args := m.Called(ctx, applicationID)
	return args.Error(0)
}

func (m *MockApplicationService) RestartApplication(ctx context.Context, applicationID string) error {
	args := m.Called(ctx, applicationID)
	return args.Error(0)
}

func (m *MockApplicationService) ScaleApplication(ctx context.Context, applicationID string, replicas int) error {
	args := m.Called(ctx, applicationID, replicas)
	return args.Error(0)
}

func (m *MockApplicationService) ListPods(ctx context.Context, applicationID string) ([]application.Pod, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]application.Pod), args.Error(1)
}

func (m *MockApplicationService) RestartPod(ctx context.Context, applicationID, podName string) error {
	args := m.Called(ctx, applicationID, podName)
	return args.Error(0)
}

func (m *MockApplicationService) GetPodLogs(ctx context.Context, query application.LogQuery) ([]application.LogEntry, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]application.LogEntry), args.Error(1)
}

func (m *MockApplicationService) StreamPodLogs(ctx context.Context, query application.LogQuery) (io.ReadCloser, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockApplicationService) GetApplicationMetrics(ctx context.Context, applicationID string) (*application.ApplicationMetrics, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.ApplicationMetrics), args.Error(1)
}

func (m *MockApplicationService) GetApplicationEvents(ctx context.Context, applicationID string, limit int) ([]application.ApplicationEvent, error) {
	args := m.Called(ctx, applicationID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]application.ApplicationEvent), args.Error(1)
}

func (m *MockApplicationService) UpdateNetworkConfig(ctx context.Context, applicationID string, config application.NetworkConfig) error {
	args := m.Called(ctx, applicationID, config)
	return args.Error(0)
}

func (m *MockApplicationService) GetApplicationEndpoints(ctx context.Context, applicationID string) ([]application.Endpoint, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]application.Endpoint), args.Error(1)
}

func (m *MockApplicationService) UpdateNodeAffinity(ctx context.Context, applicationID string, nodeSelector map[string]string) error {
	args := m.Called(ctx, applicationID, nodeSelector)
	return args.Error(0)
}

func (m *MockApplicationService) MigrateToNode(ctx context.Context, applicationID, targetNodeID string) error {
	args := m.Called(ctx, applicationID, targetNodeID)
	return args.Error(0)
}

func (m *MockApplicationService) CreateCronJob(ctx context.Context, app *application.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationService) UpdateCronJobSchedule(ctx context.Context, applicationID, newSchedule string) error {
	args := m.Called(ctx, applicationID, newSchedule)
	return args.Error(0)
}

func (m *MockApplicationService) TriggerCronJob(ctx context.Context, applicationID string) error {
	args := m.Called(ctx, applicationID)
	return args.Error(0)
}

func (m *MockApplicationService) GetCronJobExecutions(ctx context.Context, applicationID string, limit, offset int) ([]application.CronJobExecution, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]application.CronJobExecution), args.Int(1), args.Error(2)
}

func (m *MockApplicationService) GetCronJobStatus(ctx context.Context, applicationID string) (*application.CronJobStatus, error) {
	args := m.Called(ctx, applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.CronJobStatus), args.Error(1)
}

// Function-related methods
func (m *MockApplicationService) CreateFunction(ctx context.Context, workspaceID string, req application.CreateFunctionRequest) (*application.Application, error) {
	args := m.Called(ctx, workspaceID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.Application), args.Error(1)
}

func (m *MockApplicationService) DeployFunctionVersion(ctx context.Context, applicationID string, sourceCode string) (*application.FunctionVersion, error) {
	args := m.Called(ctx, applicationID, sourceCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.FunctionVersion), args.Error(1)
}

func (m *MockApplicationService) GetFunctionVersions(ctx context.Context, applicationID string) ([]application.FunctionVersion, error) {
	args := m.Called(ctx, applicationID)
	return args.Get(0).([]application.FunctionVersion), args.Error(1)
}

func (m *MockApplicationService) SetActiveFunctionVersion(ctx context.Context, applicationID, versionID string) error {
	args := m.Called(ctx, applicationID, versionID)
	return args.Error(0)
}

func (m *MockApplicationService) InvokeFunction(ctx context.Context, applicationID string, req application.InvokeFunctionRequest) (*application.InvokeFunctionResponse, error) {
	args := m.Called(ctx, applicationID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.InvokeFunctionResponse), args.Error(1)
}

func (m *MockApplicationService) GetFunctionInvocations(ctx context.Context, applicationID string, limit, offset int) ([]application.FunctionInvocation, int, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]application.FunctionInvocation), args.Int(1), args.Error(2)
}

func (m *MockApplicationService) GetFunctionEvents(ctx context.Context, applicationID string, limit int) ([]application.FunctionEvent, error) {
	args := m.Called(ctx, applicationID, limit)
	return args.Get(0).([]application.FunctionEvent), args.Error(1)
}

func (m *MockApplicationService) ProcessFunctionEvent(ctx context.Context, eventID string) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

// ApplicationHandlerTestSuite tests the application handlers
type ApplicationHandlerTestSuite struct {
	suite.Suite
	router      *gin.Engine
	handler     *ApplicationHandler
	mockService *MockApplicationService
}

func (suite *ApplicationHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.mockService = new(MockApplicationService)
	suite.handler = &ApplicationHandler{
		appService: suite.mockService,
	}

	// Setup routes for testing
	v1 := suite.router.Group("/api/v1")
	
	// Setup auth middleware mock
	v1.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Set("org_id", "test-org")
		c.Next()
	})

	// Organization scoped routes
	orgs := v1.Group("/organizations/:orgId")
	workspaces := orgs.Group("/workspaces/:wsId")
	applications := workspaces.Group("/applications")
	{
		applications.POST("/", suite.handler.CreateApplication)
		applications.GET("/", suite.handler.ListApplications)
		applications.GET("/:appId", suite.handler.GetApplication)
		applications.PUT("/:appId", suite.handler.UpdateApplication)
		applications.DELETE("/:appId", suite.handler.DeleteApplication)

		// Application operations
		applications.POST("/:appId/start", suite.handler.StartApplication)
		applications.POST("/:appId/stop", suite.handler.StopApplication)
		applications.POST("/:appId/restart", suite.handler.RestartApplication)
		applications.POST("/:appId/scale", suite.handler.ScaleApplication)

		// Pod operations
		applications.GET("/:appId/pods", suite.handler.ListPods)
		applications.POST("/:appId/pods/:podName/restart", suite.handler.RestartPod)
		applications.GET("/:appId/logs", suite.handler.GetPodLogs)
		applications.GET("/:appId/logs/stream", suite.handler.StreamPodLogs)

		// Monitoring
		applications.GET("/:appId/metrics", suite.handler.GetApplicationMetrics)
		applications.GET("/:appId/events", suite.handler.GetApplicationEvents)

		// Network operations
		applications.PUT("/:appId/network", suite.handler.UpdateNetworkConfig)
		applications.GET("/:appId/endpoints", suite.handler.GetApplicationEndpoints)

		// Node operations
		applications.PUT("/:appId/node-affinity", suite.handler.UpdateNodeAffinity)
		applications.POST("/:appId/migrate", suite.handler.MigrateToNode)
	}
}

func (suite *ApplicationHandlerTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *ApplicationHandlerTestSuite) TestCreateApplication() {
	workspaceID := uuid.New().String()
	projectID := uuid.New().String()
	
	req := application.CreateApplicationRequest{
		Name:      "test-app",
		Type:      application.ApplicationTypeStateless,
		ProjectID: projectID,
		Source: application.ApplicationSource{
			Type:  application.SourceTypeImage,
			Image: "nginx:latest",
		},
		Config: application.ApplicationConfig{
			Replicas: 3,
			Resources: application.ResourceRequests{
				CPURequest:    "100m",
				MemoryRequest: "128Mi",
			},
		},
	}

	expectedApp := &application.Application{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Name:        req.Name,
		Type:        req.Type,
		Status:      application.ApplicationStatusDeploying,
		Source:      req.Source,
		Config:      req.Config,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	suite.mockService.On("CreateApplication", mock.Anything, workspaceID, req).Return(expectedApp, nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/organizations/test-org/workspaces/"+workspaceID+"/applications/", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, r)

	// Debug: print response if not expected status
	if w.Code != http.StatusCreated {
		suite.T().Logf("Response status: %d, body: %s", w.Code, w.Body.String())
		suite.T().Logf("Location header: %s", w.Header().Get("Location"))
	}

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response application.Application
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedApp.ID, response.ID)
	assert.Equal(suite.T(), expectedApp.Name, response.Name)
}

func (suite *ApplicationHandlerTestSuite) TestCreateApplication_InvalidType() {
	workspaceID := uuid.New().String()
	
	req := map[string]interface{}{
		"name":       "test-app",
		"type":       "invalid-type",
		"project_id": uuid.New().String(),
		"source": map[string]interface{}{
			"type":  "image",
			"image": "nginx:latest",
		},
	}

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/organizations/test-org/workspaces/"+workspaceID+"/applications/", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "invalid application type")
}

func (suite *ApplicationHandlerTestSuite) TestGetApplication() {
	workspaceID := uuid.New().String()
	appID := uuid.New().String()

	expectedApp := &application.Application{
		ID:          appID,
		WorkspaceID: workspaceID,
		Name:        "test-app",
		Type:        application.ApplicationTypeStateless,
		Status:      application.ApplicationStatusRunning,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	suite.mockService.On("GetApplication", mock.Anything, appID).Return(expectedApp, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/organizations/test-org/workspaces/"+workspaceID+"/applications/"+appID, nil)

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response application.Application
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedApp.ID, response.ID)
}

func (suite *ApplicationHandlerTestSuite) TestGetApplication_NotFound() {
	workspaceID := uuid.New().String()
	appID := uuid.New().String()

	suite.mockService.On("GetApplication", mock.Anything, appID).Return(nil, errors.New("application not found"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/organizations/test-org/workspaces/"+workspaceID+"/applications/"+appID, nil)

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *ApplicationHandlerTestSuite) TestListApplications() {
	workspaceID := uuid.New().String()
	projectID := uuid.New().String()

	expectedApps := []application.Application{
		{
			ID:          uuid.New().String(),
			WorkspaceID: workspaceID,
			ProjectID:   projectID,
			Name:        "app-1",
			Type:        application.ApplicationTypeStateless,
			Status:      application.ApplicationStatusRunning,
		},
		{
			ID:          uuid.New().String(),
			WorkspaceID: workspaceID,
			ProjectID:   projectID,
			Name:        "app-2",
			Type:        application.ApplicationTypeStateful,
			Status:      application.ApplicationStatusRunning,
		},
	}

	suite.mockService.On("ListApplications", mock.Anything, workspaceID, projectID).Return(expectedApps, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/organizations/test-org/workspaces/"+workspaceID+"/applications/?project_id="+projectID, nil)

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response []application.Application
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), response, 2)
}

func (suite *ApplicationHandlerTestSuite) TestUpdateApplication() {
	workspaceID := uuid.New().String()
	appID := uuid.New().String()
	
	replicas := 5
	req := application.UpdateApplicationRequest{
		Replicas:     &replicas,
		ImageVersion: "nginx:1.21",
	}

	updatedApp := &application.Application{
		ID:          appID,
		WorkspaceID: workspaceID,
		Name:        "test-app",
		Status:      application.ApplicationStatusUpdating,
		UpdatedAt:   time.Now(),
	}

	suite.mockService.On("UpdateApplication", mock.Anything, appID, req).Return(updatedApp, nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/organizations/test-org/workspaces/"+workspaceID+"/applications/"+appID, bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *ApplicationHandlerTestSuite) TestDeleteApplication() {
	workspaceID := uuid.New().String()
	appID := uuid.New().String()

	suite.mockService.On("DeleteApplication", mock.Anything, appID).Return(nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/v1/organizations/test-org/workspaces/"+workspaceID+"/applications/"+appID, nil)

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
}

func (suite *ApplicationHandlerTestSuite) TestScaleApplication() {
	workspaceID := uuid.New().String()
	appID := uuid.New().String()
	
	req := struct {
		Replicas int `json:"replicas"`
	}{
		Replicas: 10,
	}

	suite.mockService.On("ScaleApplication", mock.Anything, appID, req.Replicas).Return(nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/organizations/test-org/workspaces/"+workspaceID+"/applications/"+appID+"/scale", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *ApplicationHandlerTestSuite) TestListPods() {
	workspaceID := uuid.New().String()
	appID := uuid.New().String()

	expectedPods := []application.Pod{
		{
			Name:      "app-pod-1",
			Status:    "Running",
			NodeName:  "node-1",
			IP:        "10.0.0.1",
			StartTime: time.Now().Add(-1 * time.Hour),
			Restarts:  0,
		},
		{
			Name:      "app-pod-2",
			Status:    "Running",
			NodeName:  "node-2",
			IP:        "10.0.0.2",
			StartTime: time.Now().Add(-2 * time.Hour),
			Restarts:  1,
		},
	}

	suite.mockService.On("ListPods", mock.Anything, appID).Return(expectedPods, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/organizations/test-org/workspaces/"+workspaceID+"/applications/"+appID+"/pods", nil)

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response []application.Pod
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), response, 2)
}

func (suite *ApplicationHandlerTestSuite) TestGetApplicationMetrics() {
	workspaceID := uuid.New().String()
	appID := uuid.New().String()

	expectedMetrics := &application.ApplicationMetrics{
		ApplicationID: appID,
		Timestamp:     time.Now(),
		PodMetrics: []application.PodMetrics{
			{
				PodName:     "app-pod-1",
				CPUUsage:    0.5,
				MemoryUsage: 128.5,
			},
		},
		AggregateUsage: application.AggregateResourceUsage{
			TotalCPU:    0.5,
			TotalMemory: 128.5,
		},
	}

	suite.mockService.On("GetApplicationMetrics", mock.Anything, appID).Return(expectedMetrics, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/organizations/test-org/workspaces/"+workspaceID+"/applications/"+appID+"/metrics", nil)

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response application.ApplicationMetrics
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedMetrics.ApplicationID, response.ApplicationID)
}

func TestApplicationHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationHandlerTestSuite))
}