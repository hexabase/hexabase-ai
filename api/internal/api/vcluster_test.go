package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/db"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// VClusterTestSuite provides test fixtures for vCluster lifecycle tests
type VClusterTestSuite struct {
	suite.Suite
	db            *gorm.DB
	handlers      *Handlers
	router        *gin.Engine
	authUser      *db.User
	authOrg       *db.Organization
	authToken     string
	testWorkspace *db.Workspace
	testPlan      *db.Plan
	vclusterHandler *VClusterHandler
}

func TestVClusterTestSuite(t *testing.T) {
	suite.Run(t, new(VClusterTestSuite))
}

func (suite *VClusterTestSuite) SetupSuite() {
	// Create in-memory SQLite database
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Auto-migrate all models
	err = database.AutoMigrate(
		&db.User{},
		&db.Organization{},
		&db.OrganizationUser{},
		&db.Plan{},
		&db.Workspace{},
		&db.VClusterProvisioningTask{},
		&db.Project{},
		&db.Group{},
		&db.MonitoringTarget{},
		&db.Role{},
		&db.RoleBinding{},
	)
	suite.Require().NoError(err)
	
	// Verify tables exist
	migrator := database.Migrator()
	suite.Require().True(migrator.HasTable("v_cluster_provisioning_tasks"))
	suite.Require().True(migrator.HasTable("organization_users"))

	suite.db = database

	// Setup config and logger
	cfg := &config.Config{
		K8s: config.K8sConfig{
			ConfigPath:        "/tmp/kubeconfig",
			InCluster:         false,
			VClusterNamespace: "vcluster",
		},
	}
	logger := zap.NewNop()

	// Initialize handlers with the same database instance
	suite.handlers = NewHandlers(suite.db, cfg, logger)
	suite.vclusterHandler = NewVClusterHandler(suite.db, cfg, logger)

	// Setup router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Add test auth middleware
	suite.router.Use(func(c *gin.Context) {
		if suite.authUser != nil {
			c.Set("user_id", suite.authUser.ID)
		}
		c.Next()
	})

	// Setup routes
	api := suite.router.Group("/api/v1")
	{
		orgs := api.Group("/organizations/:orgId")
		{
			workspaces := orgs.Group("/workspaces")
			{
				workspaces.POST("/", suite.handlers.Workspaces.CreateWorkspace)
				workspaces.GET("/", suite.handlers.Workspaces.ListWorkspaces)
				workspaces.GET("/:wsId", suite.handlers.Workspaces.GetWorkspace)
				workspaces.PUT("/:wsId", suite.handlers.Workspaces.UpdateWorkspace)
				workspaces.DELETE("/:wsId", suite.handlers.Workspaces.DeleteWorkspace)
				workspaces.GET("/:wsId/kubeconfig", suite.handlers.Workspaces.GetKubeconfig)
				
				// vCluster lifecycle routes
				vcluster := workspaces.Group("/:wsId/vcluster")
				{
					vcluster.POST("/provision", suite.vclusterHandler.ProvisionVCluster)
					vcluster.GET("/status", suite.vclusterHandler.GetVClusterStatus)
					vcluster.POST("/start", suite.vclusterHandler.StartVCluster)
					vcluster.POST("/stop", suite.vclusterHandler.StopVCluster)
					vcluster.DELETE("/", suite.vclusterHandler.DestroyVCluster)
					vcluster.GET("/health", suite.vclusterHandler.GetVClusterHealth)
					vcluster.POST("/upgrade", suite.vclusterHandler.UpgradeVCluster)
					vcluster.GET("/logs", suite.vclusterHandler.GetVClusterLogs)
					vcluster.POST("/backup", suite.vclusterHandler.BackupVCluster)
					vcluster.POST("/restore", suite.vclusterHandler.RestoreVCluster)
				}
			}
		}
		
		// Task management
		tasks := api.Group("/tasks")
		{
			tasks.GET("/", suite.vclusterHandler.ListTasks)
			tasks.GET("/:taskId", suite.vclusterHandler.GetTask)
			tasks.POST("/:taskId/retry", suite.vclusterHandler.RetryTask)
		}
	}

	// Create test data
	suite.setupTestData()
}

func (suite *VClusterTestSuite) setupTestData() {
	// Create test user
	suite.authUser = &db.User{
		ID:          "user-vcluster-test",
		Email:       "vcluster@test.com",
		DisplayName: "VCluster Test User",
		Provider:    "google",
	}
	suite.Require().NoError(suite.db.Create(suite.authUser).Error)

	// Create test organization
	suite.authOrg = &db.Organization{
		ID:   "org-vcluster-test",
		Name: "VCluster Test Org",
	}
	suite.Require().NoError(suite.db.Create(suite.authOrg).Error)

	// Create organization membership
	orgUser := &db.OrganizationUser{
		UserID:         suite.authUser.ID,
		OrganizationID: suite.authOrg.ID,
		Role:           "admin",
	}
	suite.Require().NoError(suite.db.Create(orgUser).Error)

	// Create test plan
	suite.testPlan = &db.Plan{
		ID:            "plan-basic",
		Name:          "Basic Plan",
		Description:   "Basic plan for testing",
		Price:         10.00,
		Currency:      "usd",
		StripePriceID: "price_test_123",
		ResourceLimits: `{
			"cpu_limit": "2",
			"memory_limit": "4Gi",
			"storage_limit": "10Gi",
			"max_projects": 5
		}`,
		IsActive: true,
	}
	suite.Require().NoError(suite.db.Create(suite.testPlan).Error)

	// Create test workspace
	suite.testWorkspace = &db.Workspace{
		ID:             "ws-vcluster-test",
		OrganizationID: suite.authOrg.ID,
		Name:           "VCluster Test Workspace",
		PlanID:         suite.testPlan.ID,
		VClusterStatus: "PENDING_CREATION",
		VClusterConfig: `{
			"version": "0.15.0",
			"resources": {
				"cpu": "2",
				"memory": "4Gi"
			}
		}`,
		DedicatedNodeConfig: "{}",
	}
	suite.Require().NoError(suite.db.Create(suite.testWorkspace).Error)

	// Set auth token for testing
	suite.authToken = "test-token-123"
}

func (suite *VClusterTestSuite) TearDownSuite() {
	// Clean up database
	sqlDB, err := suite.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (suite *VClusterTestSuite) TearDownTest() {
	// Clean up tasks after each test
	suite.db.Where("1=1").Delete(&db.VClusterProvisioningTask{})
}

// Test VCluster Provisioning
func (suite *VClusterTestSuite) TestProvisionVCluster() {
	tests := []struct {
		name           string
		workspaceID    string
		payload        map[string]interface{}
		expectedStatus int
		expectedFields []string
		setupFunc      func()
	}{
		{
			name:        "successful vcluster provisioning",
			workspaceID: suite.testWorkspace.ID,
			payload: map[string]interface{}{
				"version": "0.15.0",
				"resources": map[string]interface{}{
					"cpu":    "2",
					"memory": "4Gi",
				},
				"features": []string{"networking", "storage"},
			},
			expectedStatus: http.StatusAccepted,
			expectedFields: []string{"task_id", "status", "message"},
		},
		{
			name:           "provision already running vcluster",
			workspaceID:    suite.testWorkspace.ID,
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusConflict,
			setupFunc: func() {
				suite.db.Model(suite.testWorkspace).Update("v_cluster_status", "RUNNING")
			},
		},
		{
			name:           "provision non-existent workspace",
			workspaceID:    "ws-nonexistent",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			// Reset workspace status
			defer func() {
				suite.db.Model(suite.testWorkspace).Update("v_cluster_status", "PENDING_CREATION")
			}()

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/vcluster/provision", suite.authOrg.ID, tt.workspaceID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusAccepted {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)

				for _, field := range tt.expectedFields {
					suite.Contains(response, field)
				}

				// Verify task was created
				var task db.VClusterProvisioningTask
				err = suite.db.Where("workspace_id = ? AND task_type = ?", tt.workspaceID, "CREATE").First(&task).Error
				suite.NoError(err)
				suite.Equal("PENDING", task.Status)
			}
		})
	}
}

// Test VCluster Status
func (suite *VClusterTestSuite) TestGetVClusterStatus() {
	tests := []struct {
		name           string
		workspaceID    string
		setupStatus    string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "get status for pending vcluster",
			workspaceID:    suite.testWorkspace.ID,
			setupStatus:    "PENDING_CREATION",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"status", "workspace", "cluster_info"},
		},
		{
			name:           "get status for running vcluster",
			workspaceID:    suite.testWorkspace.ID,
			setupStatus:    "RUNNING",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"status", "workspace", "cluster_info", "health"},
		},
		{
			name:           "get status for non-existent workspace",
			workspaceID:    "ws-nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.setupStatus != "" {
				suite.db.Model(suite.testWorkspace).Update("v_cluster_status", tt.setupStatus)
			}

			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/vcluster/status", suite.authOrg.ID, tt.workspaceID), nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)

				for _, field := range tt.expectedFields {
					suite.Contains(response, field)
				}
			}
		})
	}
}

// Test VCluster Start/Stop
func (suite *VClusterTestSuite) TestStartStopVCluster() {
	tests := []struct {
		name           string
		endpoint       string
		workspaceID    string
		setupStatus    string
		expectedStatus int
		expectedTaskType string
	}{
		{
			name:           "start stopped vcluster",
			endpoint:       "start",
			workspaceID:    suite.testWorkspace.ID,
			setupStatus:    "STOPPED",
			expectedStatus: http.StatusAccepted,
			expectedTaskType: "START",
		},
		{
			name:           "stop running vcluster",
			endpoint:       "stop",
			workspaceID:    suite.testWorkspace.ID,
			setupStatus:    "RUNNING",
			expectedStatus: http.StatusAccepted,
			expectedTaskType: "STOP",
		},
		{
			name:           "start already running vcluster",
			endpoint:       "start",
			workspaceID:    suite.testWorkspace.ID,
			setupStatus:    "RUNNING",
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.db.Model(suite.testWorkspace).Update("v_cluster_status", tt.setupStatus)

			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/vcluster/%s", suite.authOrg.ID, tt.workspaceID, tt.endpoint), nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusAccepted {
				// Verify task was created
				var task db.VClusterProvisioningTask
				err := suite.db.Where("workspace_id = ? AND task_type = ?", tt.workspaceID, tt.expectedTaskType).First(&task).Error
				suite.NoError(err)
				suite.Equal("PENDING", task.Status)
			}
		})
	}
}

// Test VCluster Destruction
func (suite *VClusterTestSuite) TestDestroyVCluster() {
	tests := []struct {
		name           string
		workspaceID    string
		setupStatus    string
		expectedStatus int
		checkDeletion  bool
	}{
		{
			name:           "destroy running vcluster",
			workspaceID:    suite.testWorkspace.ID,
			setupStatus:    "RUNNING",
			expectedStatus: http.StatusAccepted,
			checkDeletion:  true,
		},
		{
			name:           "destroy already deleting vcluster",
			workspaceID:    suite.testWorkspace.ID,
			setupStatus:    "DELETING",
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "destroy non-existent workspace",
			workspaceID:    "ws-nonexistent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.setupStatus != "" {
				suite.db.Model(suite.testWorkspace).Update("v_cluster_status", tt.setupStatus)
			}

			req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/vcluster", suite.authOrg.ID, tt.workspaceID), nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.checkDeletion {
				// Verify task was created
				var task db.VClusterProvisioningTask
				err := suite.db.Where("workspace_id = ? AND task_type = ?", tt.workspaceID, "DELETE").First(&task).Error
				suite.NoError(err)
				suite.Equal("PENDING", task.Status)

				// Verify workspace status updated
				var workspace db.Workspace
				suite.db.First(&workspace, "id = ?", tt.workspaceID)
				suite.Equal("DELETING", workspace.VClusterStatus)
			}
		})
	}
}

// Test VCluster Health
func (suite *VClusterTestSuite) TestGetVClusterHealth() {
	tests := []struct {
		name           string
		workspaceID    string
		setupStatus    string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "health check for running vcluster",
			workspaceID:    suite.testWorkspace.ID,
			setupStatus:    "RUNNING",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"healthy", "components", "resource_usage"},
		},
		{
			name:           "health check for pending vcluster",
			workspaceID:    suite.testWorkspace.ID,
			setupStatus:    "PENDING_CREATION",
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.db.Model(suite.testWorkspace).Update("v_cluster_status", tt.setupStatus)

			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/vcluster/health", suite.authOrg.ID, tt.workspaceID), nil)
			req.Header.Set("Authorization", "Bearer "+suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)

				for _, field := range tt.expectedFields {
					suite.Contains(response, field)
				}
			}
		})
	}
}

// Test VCluster Upgrade
func (suite *VClusterTestSuite) TestUpgradeVCluster() {
	tests := []struct {
		name           string
		workspaceID    string
		payload        map[string]interface{}
		setupStatus    string
		expectedStatus int
	}{
		{
			name:        "upgrade running vcluster",
			workspaceID: suite.testWorkspace.ID,
			payload: map[string]interface{}{
				"target_version": "0.16.0",
				"strategy":       "rolling",
			},
			setupStatus:    "RUNNING",
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "upgrade non-running vcluster",
			workspaceID:    suite.testWorkspace.ID,
			setupStatus:    "PENDING_CREATION",
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.db.Model(suite.testWorkspace).Update("v_cluster_status", tt.setupStatus)

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/vcluster/upgrade", suite.authOrg.ID, tt.workspaceID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.authToken)

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusAccepted {
				// Verify task was created
				var task db.VClusterProvisioningTask
				err := suite.db.Where("workspace_id = ? AND task_type = ?", tt.workspaceID, "UPGRADE").First(&task).Error
				suite.NoError(err)
			}
		})
	}
}

// Test Task Management
func (suite *VClusterTestSuite) TestListTasks() {
	// Create test tasks
	task1 := &db.VClusterProvisioningTask{
		WorkspaceID: suite.testWorkspace.ID,
		TaskType:    "CREATE",
		Status:      "COMPLETED",
		Payload:     "{}",
	}
	task2 := &db.VClusterProvisioningTask{
		WorkspaceID: suite.testWorkspace.ID,
		TaskType:    "DELETE",
		Status:      "PENDING",
		Payload:     "{}",
	}
	suite.Require().NoError(suite.db.Create(task1).Error)
	suite.Require().NoError(suite.db.Create(task2).Error)

	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)

	suite.Contains(response, "tasks")
	tasks := response["tasks"].([]interface{})
	suite.GreaterOrEqual(len(tasks), 2)
}

// Test Task Retry
func (suite *VClusterTestSuite) TestRetryTask() {
	// Create failed task
	task := &db.VClusterProvisioningTask{
		WorkspaceID:  suite.testWorkspace.ID,
		TaskType:     "CREATE",
		Status:       "FAILED",
		Payload:      "{}",
		ErrorMessage: &[]string{"Connection timeout"}[0],
	}
	suite.Require().NoError(suite.db.Create(task).Error)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/tasks/%s/retry", task.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	// Verify task status reset
	var updatedTask db.VClusterProvisioningTask
	suite.db.First(&updatedTask, "id = ?", task.ID)
	suite.Equal("PENDING", updatedTask.Status)
	suite.Nil(updatedTask.ErrorMessage)
}

// Test VCluster Backup/Restore
func (suite *VClusterTestSuite) TestBackupRestoreVCluster() {
	suite.db.Model(suite.testWorkspace).Update("v_cluster_status", "RUNNING")

	// Test backup
	backupPayload := map[string]interface{}{
		"backup_name": "test-backup-1",
		"retention":   "30d",
	}
	body, _ := json.Marshal(backupPayload)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/vcluster/backup", suite.authOrg.ID, suite.testWorkspace.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusAccepted, w.Code)

	// Test restore
	restorePayload := map[string]interface{}{
		"backup_name": "test-backup-1",
		"strategy":    "replace",
	}
	body, _ = json.Marshal(restorePayload)
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/organizations/%s/workspaces/%s/vcluster/restore", suite.authOrg.ID, suite.testWorkspace.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authToken)

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusAccepted, w.Code)
}

