package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/hexabase-ai/api/internal/application/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApplicationHandler_CreateCronJob(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		request        domain.CreateApplicationRequest
		expectedStatus int
		expectedError  bool
		setupMocks     func(*MockApplicationService)
	}{
		{
			name: "successful cronjob creation",
			request: domain.CreateApplicationRequest{
				ProjectID: "proj-123",
				Name:      "daily-backup",
				Type:      domain.ApplicationTypeCronJob,
				Source: domain.ApplicationSource{
					Type:  domain.SourceTypeImage,
					Image: "backup-tool:latest",
				},
				Config: domain.ApplicationConfig{
					Resources: domain.ResourceRequests{
						CPURequest:    "100m",
						MemoryRequest: "256Mi",
						CPULimit:      "500m",
						MemoryLimit:   "1Gi",
					},
					EnvVars: map[string]string{
						"BACKUP_PATH": "/data",
					},
				},
				CronSchedule: "0 2 * * *",
				CronCommand:  []string{"/usr/bin/backup.sh"},
				CronArgs:     []string{"--full", "--compress"},
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
			setupMocks: func(svc *MockApplicationService) {
				svc.On("CreateApplication", mock.Anything, "ws-123", mock.MatchedBy(func(req domain.CreateApplicationRequest) bool {
					return req.Name == "daily-backup" &&
						req.Type == domain.ApplicationTypeCronJob &&
						req.CronSchedule == "0 2 * * *"
				})).Return(&domain.Application{
					ID:           "app-789",
					WorkspaceID:  "ws-123",
					ProjectID:    "proj-123",
					Name:         "daily-backup",
					Type:         domain.ApplicationTypeCronJob,
					Status:       domain.ApplicationStatusRunning,
					CronSchedule: "0 2 * * *",
					CronCommand:  []string{"/usr/bin/backup.sh"},
					CronArgs:     []string{"--full", "--compress"},
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil)
			},
		},
		{
			name: "cronjob with template",
			request: domain.CreateApplicationRequest{
				ProjectID:     "proj-123",
				Name:          "scheduled-task",
				Type:          domain.ApplicationTypeCronJob,
				CronSchedule:  "*/5 * * * *",
				TemplateAppID: "template-123",
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
			setupMocks: func(svc *MockApplicationService) {
				svc.On("CreateApplication", mock.Anything, "ws-123", mock.MatchedBy(func(req domain.CreateApplicationRequest) bool {
					return req.Name == "scheduled-task" &&
						req.TemplateAppID == "template-123"
				})).Return(&domain.Application{
					ID:            "app-890",
					WorkspaceID:   "ws-123",
					ProjectID:     "proj-123",
					Name:          "scheduled-task",
					Type:          domain.ApplicationTypeCronJob,
					Status:        domain.ApplicationStatusRunning,
					CronSchedule:  "*/5 * * * *",
					TemplateAppID: "template-123",
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}, nil)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockApplicationService)
			handler := &ApplicationHandler{
				appService: mockService,
			}
			
			tt.setupMocks(mockService)
			
			// Create request
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/workspaces/ws-123/applications", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			// Create response recorder
			w := httptest.NewRecorder()
			
			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "wsId", Value: "ws-123"},
			}
			
			// Call handler
			handler.CreateApplication(c)
			
			// Assert response
			if w.Code != tt.expectedStatus {
				t.Logf("Response body: %s", w.Body.String())
			}
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if !tt.expectedError {
				var response domain.Application
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, domain.ApplicationTypeCronJob, response.Type)
			}
			
			mockService.AssertExpectations(t)
		})
	}
}

func TestApplicationHandler_UpdateCronJobSchedule(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		applicationID  string
		request        domain.UpdateCronScheduleRequest
		expectedStatus int
		expectedError  bool
		setupMocks     func(*MockApplicationService)
	}{
		{
			name:          "successful schedule update",
			applicationID: "app-123",
			request: domain.UpdateCronScheduleRequest{
				Schedule: "0 4 * * *",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			setupMocks: func(svc *MockApplicationService) {
				svc.On("UpdateCronJobSchedule", mock.Anything, "app-123", "0 4 * * *").Return(nil)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockApplicationService)
			handler := &ApplicationHandler{
				appService: mockService,
			}
			
			tt.setupMocks(mockService)
			
			// Create request
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("PUT", "/api/v1/applications/"+tt.applicationID+"/schedule", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			
			// Create response recorder
			w := httptest.NewRecorder()
			
			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "appId", Value: tt.applicationID},
			}
			
			// Call handler
			handler.UpdateCronJobSchedule(c)
			
			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			mockService.AssertExpectations(t)
		})
	}
}

func TestApplicationHandler_TriggerCronJob(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		applicationID  string
		expectedStatus int
		expectedError  bool
		setupMocks     func(*MockApplicationService)
	}{
		{
			name:           "successful trigger",
			applicationID:  "app-123",
			expectedStatus: http.StatusOK,
			expectedError:  false,
			setupMocks: func(svc *MockApplicationService) {
				req := &domain.TriggerCronJobRequest{ApplicationID: "app-123"}
				execution := &domain.CronJobExecution{
					ID:            "cje-123",
					ApplicationID: "app-123",
					Status:        domain.CronJobExecutionStatusSucceeded,
				}
				svc.On("TriggerCronJob", mock.Anything, req).Return(execution, nil)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockApplicationService)
			handler := &ApplicationHandler{
				appService: mockService,
			}
			
			tt.setupMocks(mockService)
			
			// Create request
			req := httptest.NewRequest("POST", "/api/v1/applications/"+tt.applicationID+"/trigger", nil)
			
			// Create response recorder
			w := httptest.NewRecorder()
			
			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "appId", Value: tt.applicationID},
			}
			
			// Call handler
			handler.TriggerCronJob(c)
			
			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			mockService.AssertExpectations(t)
		})
	}
}

func TestApplicationHandler_GetCronJobExecutions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now()
	
	tests := []struct {
		name           string
		applicationID  string
		queryParams    map[string]string
		expectedStatus int
		expectedError  bool
		setupMocks     func(*MockApplicationService)
	}{
		{
			name:          "successful get executions",
			applicationID: "app-123",
			queryParams: map[string]string{
				"page":     "1",
				"per_page": "10",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			setupMocks: func(svc *MockApplicationService) {
				executions := []domain.CronJobExecution{
					{
						ID:            "exec-1",
						ApplicationID: "app-123",
						JobName:       "daily-backup-1234567890",
						StartedAt:     now.Add(-1 * time.Hour),
						CompletedAt:   &now,
						Status:        domain.CronJobExecutionStatusSucceeded,
						ExitCode:      intPtr(0),
						Logs:          "Backup completed successfully",
					},
				}
				svc.On("GetCronJobExecutions", mock.Anything, "app-123", 10, 0).
					Return(executions, 1, nil)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockApplicationService)
			handler := &ApplicationHandler{
				appService: mockService,
			}
			
			tt.setupMocks(mockService)
			
			// Create request
			req := httptest.NewRequest("GET", "/api/v1/applications/"+tt.applicationID+"/executions", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
			
			// Create response recorder
			w := httptest.NewRecorder()
			
			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{
				{Key: "appId", Value: tt.applicationID},
			}
			
			// Call handler
			handler.GetCronJobExecutions(c)
			
			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if !tt.expectedError {
				var response domain.CronJobExecutionList
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, response.Total, 0)
			}
			
			mockService.AssertExpectations(t)
		})
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}