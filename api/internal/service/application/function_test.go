package application

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


func TestCreateFunction(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful function creation", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		service := NewService(mockRepo, mockK8s)

		req := application.CreateFunctionRequest{
			Name:          "test-function",
			ProjectID:     "proj-123",
			Runtime:       application.FunctionRuntimePython39,
			Handler:       "main.handler",
			SourceCode:    base64.StdEncoding.EncodeToString([]byte("def handler(event, context):\n    return {'statusCode': 200}")),
			SourceType:    application.FunctionSourceInline,
			Timeout:       300,
			Memory:        256,
			TriggerType:   application.FunctionTriggerHTTP,
			TriggerConfig: map[string]interface{}{"path": "/api/function"},
		}

		// Mock repository calls
		mockRepo.On("CreateApplication", ctx, mock.MatchedBy(func(app *application.Application) bool {
			return app.Name == req.Name &&
				app.Type == application.ApplicationTypeFunction &&
				app.FunctionRuntime == req.Runtime &&
				app.FunctionHandler == req.Handler
		})).Return(nil)

		mockRepo.On("CreateFunctionVersion", ctx, mock.MatchedBy(func(v *application.FunctionVersion) bool {
			return v.VersionNumber == 1 &&
				v.SourceCode == req.SourceCode &&
				v.SourceType == req.SourceType &&
				v.IsActive == true
		})).Return(nil)

		// Execute
		app, err := service.CreateFunction(ctx, "ws-123", req)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, app)
		assert.Equal(t, req.Name, app.Name)
		assert.Equal(t, application.ApplicationTypeFunction, app.Type)
		
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("Function creation with validation error", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		service := NewService(mockRepo, mockK8s)

		req := application.CreateFunctionRequest{
			Name:        "", // Invalid empty name
			ProjectID:   "proj-123",
			Runtime:     application.FunctionRuntimePython39,
			Handler:     "main.handler",
			SourceCode:  base64.StdEncoding.EncodeToString([]byte("def handler(event, context):\n    return {'statusCode': 200}")),
			SourceType:  application.FunctionSourceInline,
		}

		// Execute
		app, err := service.CreateFunction(ctx, "ws-123", req)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, app)
		assert.Contains(t, err.Error(), "name is required")
		
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})
}

func TestDeployFunctionVersion(t *testing.T) {
	ctx := context.Background()

	t.Run("Deploy new function version", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		service := NewService(mockRepo, mockK8s)

		appID := "app-func-123"
		sourceCode := base64.StdEncoding.EncodeToString([]byte("def handler(event, context):\n    return {'statusCode': 201}"))

		// Mock getting the application
		app := &application.Application{
			ID:                  appID,
			WorkspaceID:         "ws-123",
			ProjectID:           "proj-123",
			Name:                "test-function",
			Type:                application.ApplicationTypeFunction,
			Status:              application.ApplicationStatusRunning,
			FunctionRuntime:     application.FunctionRuntimePython39,
			FunctionHandler:     "main.handler",
			FunctionTimeout:     300,
			FunctionMemory:      256,
			FunctionTriggerType: application.FunctionTriggerHTTP,
		}
		mockRepo.On("GetApplication", ctx, appID).Return(app, nil)

		// Mock getting existing versions
		existingVersions := []application.FunctionVersion{
			{ID: "fv-1", ApplicationID: appID, VersionNumber: 1, IsActive: true},
		}
		mockRepo.On("GetFunctionVersions", ctx, appID).Return(existingVersions, nil)

		// Mock creating new version
		mockRepo.On("CreateFunctionVersion", ctx, mock.MatchedBy(func(v *application.FunctionVersion) bool {
			return v.ApplicationID == appID &&
				v.VersionNumber == 2 &&
				v.SourceCode == sourceCode &&
				v.SourceType == application.FunctionSourceInline &&
				v.BuildStatus == application.FunctionBuildPending &&
				v.IsActive == false
		})).Return(nil)

		// Mock build process
		mockRepo.On("UpdateFunctionVersion", ctx, mock.MatchedBy(func(v *application.FunctionVersion) bool {
			return v.BuildStatus == application.FunctionBuildBuilding
		})).Return(nil).Once()

		mockRepo.On("UpdateFunctionVersion", ctx, mock.MatchedBy(func(v *application.FunctionVersion) bool {
			return v.BuildStatus == application.FunctionBuildSuccess &&
				v.ImageURI != ""
		})).Return(nil).Once()

		// Execute
		version, err := service.DeployFunctionVersion(ctx, appID, sourceCode)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, version)
		assert.Equal(t, 2, version.VersionNumber)
		assert.Equal(t, sourceCode, version.SourceCode)
		
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})
}

func TestInvokeFunction(t *testing.T) {
	ctx := context.Background()

	t.Run("Invoke function successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		service := NewService(mockRepo, mockK8s)

		appID := "app-func-123"

		// Mock getting the application
		app := &application.Application{
			ID:                  appID,
			WorkspaceID:         "ws-123",
			ProjectID:           "proj-123",
			Name:                "test-function",
			Type:                application.ApplicationTypeFunction,
			Status:              application.ApplicationStatusRunning,
			FunctionRuntime:     application.FunctionRuntimePython39,
			FunctionTriggerType: application.FunctionTriggerHTTP,
		}
		mockRepo.On("GetApplication", ctx, appID).Return(app, nil)

		// Mock getting active version
		activeVersion := &application.FunctionVersion{
			ID:            "fv-active",
			ApplicationID: appID,
			VersionNumber: 2,
			IsActive:      true,
			ImageURI:      "registry.example.com/functions/test-function:v2",
		}
		mockRepo.On("GetActiveFunctionVersion", ctx, appID).Return(activeVersion, nil)

		// Mock getting Knative service URL
		functionURL := "https://test-function.ws-123.example.com"
		mockK8s.On("GetKnativeServiceURL", ctx, "ws-123", "proj-123", "test-function").Return(functionURL, nil)

		// Mock creating invocation record
		mockRepo.On("CreateFunctionInvocation", ctx, mock.MatchedBy(func(inv *application.FunctionInvocation) bool {
			return inv.ApplicationID == appID &&
				inv.VersionID == activeVersion.ID &&
				inv.TriggerSource == "http"
		})).Return(nil)

		// Mock updating invocation after completion
		mockRepo.On("UpdateFunctionInvocation", ctx, mock.MatchedBy(func(inv *application.FunctionInvocation) bool {
			return inv.ResponseStatus == 200 &&
				inv.DurationMs > 0
		})).Return(nil)

		// Prepare request
		req := application.InvokeFunctionRequest{
			Method: "POST",
			Path:   "/test",
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: base64.StdEncoding.EncodeToString([]byte(`{"key": "value"}`)),
		}

		// Execute
		resp, err := service.InvokeFunction(ctx, appID, req)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.InvocationID)
		
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("Invoke function with no active version", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		service := NewService(mockRepo, mockK8s)

		appID := "app-func-123"

		// Mock getting the application
		app := &application.Application{
			ID:                  appID,
			Type:                application.ApplicationTypeFunction,
			Status:              application.ApplicationStatusRunning,
			FunctionTriggerType: application.FunctionTriggerHTTP,
		}
		mockRepo.On("GetApplication", ctx, appID).Return(app, nil)

		// Mock no active version
		mockRepo.On("GetActiveFunctionVersion", ctx, appID).Return(nil, nil)

		// Prepare request
		req := application.InvokeFunctionRequest{
			Method: "POST",
			Path:   "/test",
		}

		// Execute
		resp, err := service.InvokeFunction(ctx, appID, req)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "no active version")
		
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})
}

func TestProcessFunctionEvent(t *testing.T) {
	ctx := context.Background()

	t.Run("Process event successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		service := NewService(mockRepo, mockK8s)

		eventID := "fe-123"
		appID := "app-func-123"

		// Mock getting the event
		event := &application.FunctionEvent{
			ID:               eventID,
			ApplicationID:    appID,
			EventType:        "webhook.github",
			EventSource:      "github",
			EventData:        map[string]interface{}{"action": "push"},
			ProcessingStatus: "pending",
			RetryCount:       0,
		}
		mockRepo.On("GetFunctionEvent", ctx, eventID).Return(event, nil)

		// Mock getting the application
		app := &application.Application{
			ID:                  appID,
			WorkspaceID:         "ws-123",
			ProjectID:           "proj-123",
			Name:                "webhook-handler",
			Type:                application.ApplicationTypeFunction,
			Status:              application.ApplicationStatusRunning,
			FunctionTriggerType: application.FunctionTriggerEvent,
		}
		mockRepo.On("GetApplication", ctx, appID).Return(app, nil)

		// Mock updating event status to processing
		mockRepo.On("UpdateFunctionEvent", ctx, mock.MatchedBy(func(e *application.FunctionEvent) bool {
			return e.ID == eventID && e.ProcessingStatus == "processing"
		})).Return(nil).Once()

		// Mock function invocation (reuse logic from InvokeFunction)
		activeVersion := &application.FunctionVersion{
			ID:       "fv-active",
			IsActive: true,
		}
		mockRepo.On("GetActiveFunctionVersion", ctx, appID).Return(activeVersion, nil)
		mockK8s.On("GetKnativeServiceURL", ctx, "ws-123", "proj-123", "webhook-handler").Return("https://webhook-handler.example.com", nil)
		
		mockRepo.On("CreateFunctionInvocation", ctx, mock.Anything).Return(nil)
		mockRepo.On("UpdateFunctionInvocation", ctx, mock.Anything).Return(nil)

		// Mock updating event status to success
		mockRepo.On("UpdateFunctionEvent", ctx, mock.MatchedBy(func(e *application.FunctionEvent) bool {
			return e.ID == eventID && 
				e.ProcessingStatus == "success" &&
				e.InvocationID != ""
		})).Return(nil).Once()

		// Execute
		err := service.ProcessFunctionEvent(ctx, eventID)
		
		// Assert
		assert.NoError(t, err)
		
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})

	t.Run("Process event with retry on failure", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockK8s := new(MockKubernetesRepository)
		service := NewService(mockRepo, mockK8s)

		eventID := "fe-456"
		appID := "app-func-456"

		// Mock getting the event
		event := &application.FunctionEvent{
			ID:               eventID,
			ApplicationID:    appID,
			EventType:        "webhook.github",
			EventSource:      "github",
			EventData:        map[string]interface{}{"action": "push"},
			ProcessingStatus: "pending",
			RetryCount:       1,
			MaxRetries:       3,
		}
		mockRepo.On("GetFunctionEvent", ctx, eventID).Return(event, nil)

		// Mock getting the application
		app := &application.Application{
			ID:                  appID,
			WorkspaceID:         "ws-123",
			ProjectID:           "proj-123",
			Name:                "webhook-handler",
			Type:                application.ApplicationTypeFunction,
			Status:              application.ApplicationStatusRunning,
			FunctionTriggerType: application.FunctionTriggerEvent,
		}
		mockRepo.On("GetApplication", ctx, appID).Return(app, nil)

		// Mock updating event status to processing
		mockRepo.On("UpdateFunctionEvent", ctx, mock.MatchedBy(func(e *application.FunctionEvent) bool {
			return e.ProcessingStatus == "processing"
		})).Return(nil).Once()

		// Mock function invocation failure
		mockRepo.On("GetActiveFunctionVersion", ctx, appID).Return(nil, errors.New("no active version"))

		// Mock updating event status to retry
		mockRepo.On("UpdateFunctionEvent", ctx, mock.MatchedBy(func(e *application.FunctionEvent) bool {
			return e.ID == eventID && 
				e.ProcessingStatus == "retry" &&
				e.RetryCount == 2 &&
				e.ErrorMessage != ""
		})).Return(nil).Once()

		// Execute
		err := service.ProcessFunctionEvent(ctx, eventID)
		
		// Assert - should not return error even on failure (retry scheduled)
		assert.NoError(t, err)
		
		mockRepo.AssertExpectations(t)
		mockK8s.AssertExpectations(t)
	})
}