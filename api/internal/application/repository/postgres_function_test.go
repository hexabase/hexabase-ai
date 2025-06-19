package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/hexabase/hexabase-ai/api/internal/application/domain"
	"github.com/hexabase/hexabase-ai/api/internal/shared/db"
	"gorm.io/driver/sqlite"
)

func setupFunctionTestDB(t *testing.T) *gorm.DB {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the necessary tables
	err = testDB.AutoMigrate(
		&db.Workspace{},
		&db.Project{},
		&db.Application{},
		&db.FunctionVersion{},
		&db.FunctionInvocation{},
		&db.FunctionEvent{},
	)
	require.NoError(t, err)

	// Create test workspace and project
	workspace := &db.Workspace{
		ID:             "ws-test",
		Name:           "Test Workspace",
		OrganizationID: "org-test",
	}
	err = testDB.Create(workspace).Error
	require.NoError(t, err)

	project := &db.Project{
		ID:          "proj-test",
		WorkspaceID: "ws-test",
		Name:        "Test Project",
	}
	err = testDB.Create(project).Error
	require.NoError(t, err)

	return testDB
}

func TestCreateFunctionVersion(t *testing.T) {
	gormDB := setupFunctionTestDB(t)
	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	// Create a function application first
	app := &domain.Application{
		ID:          "app-func-1",
		WorkspaceID: "ws-test",
		ProjectID:   "proj-test",
		Name:        "test-function",
		Type:        domain.ApplicationTypeFunction,
		Status:      domain.ApplicationStatusPending,
		Source: domain.ApplicationSource{
			Type: domain.SourceTypeImage,
		},
		Config:              domain.ApplicationConfig{},
		FunctionRuntime:     domain.FunctionRuntimePython39,
		FunctionHandler:     "main.handler",
		FunctionTimeout:     300,
		FunctionMemory:      256,
		FunctionTriggerType: domain.FunctionTriggerHTTP,
	}

	err := repo.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Test creating a function version
	version := &domain.FunctionVersion{
		ApplicationID: app.ID,
		VersionNumber: 1,
		SourceCode:    "def handler(event, context):\n    return {'statusCode': 200}",
		SourceType:    domain.FunctionSourceInline,
		BuildStatus:   domain.FunctionBuildPending,
		IsActive:      true,
	}

	err = repo.CreateFunctionVersion(ctx, version)
	assert.NoError(t, err)
	assert.NotEmpty(t, version.ID)
	assert.True(t, version.IsActive)
}

func TestGetFunctionVersion(t *testing.T) {
	gormDB := setupFunctionTestDB(t)
	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	// Create test data
	version := &db.FunctionVersion{
		ID:            "fv-test-1",
		ApplicationID: "app-func-1",
		VersionNumber: 1,
		SourceCode:    "test code",
		SourceType:    "inline",
		BuildStatus:   "success",
		IsActive:      true,
	}
	err := gormDB.Create(version).Error
	require.NoError(t, err)

	// Test getting the version
	result, err := repo.GetFunctionVersion(ctx, version.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, version.ID, result.ID)
	assert.Equal(t, version.SourceCode, result.SourceCode)
}

func TestGetActiveFunctionVersion(t *testing.T) {
	gormDB := setupFunctionTestDB(t)
	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	appID := "app-func-2"

	// Create multiple versions
	versions := []db.FunctionVersion{
		{
			ID:            "fv-1",
			ApplicationID: appID,
			VersionNumber: 1,
			SourceType:    "inline",
			BuildStatus:   "success",
			IsActive:      false,
		},
		{
			ID:            "fv-2",
			ApplicationID: appID,
			VersionNumber: 2,
			SourceType:    "inline",
			BuildStatus:   "success",
			IsActive:      true,
		},
	}

	for _, v := range versions {
		err := gormDB.Create(&v).Error
		require.NoError(t, err)
	}

	// Test getting active version
	result, err := repo.GetActiveFunctionVersion(ctx, appID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "fv-2", result.ID)
	assert.True(t, result.IsActive)
}

func TestSetActiveFunctionVersion(t *testing.T) {
	gormDB := setupFunctionTestDB(t)
	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	appID := "app-func-3"

	// Create multiple versions
	versions := []db.FunctionVersion{
		{
			ID:            "fv-old",
			ApplicationID: appID,
			VersionNumber: 1,
			SourceType:    "inline",
			BuildStatus:   "success",
			IsActive:      true,
		},
		{
			ID:            "fv-new",
			ApplicationID: appID,
			VersionNumber: 2,
			SourceType:    "inline",
			BuildStatus:   "success",
			IsActive:      false,
		},
	}

	for _, v := range versions {
		err := gormDB.Create(&v).Error
		require.NoError(t, err)
	}

	// Test setting new active version
	err := repo.SetActiveFunctionVersion(ctx, appID, "fv-new")
	assert.NoError(t, err)

	// Verify old version is no longer active
	var oldVersion db.FunctionVersion
	err = gormDB.First(&oldVersion, "id = ?", "fv-old").Error
	require.NoError(t, err)
	assert.False(t, oldVersion.IsActive)

	// Verify new version is active
	var newVersion db.FunctionVersion
	err = gormDB.First(&newVersion, "id = ?", "fv-new").Error
	require.NoError(t, err)
	assert.True(t, newVersion.IsActive)
}

func TestCreateFunctionInvocation(t *testing.T) {
	gormDB := setupFunctionTestDB(t)
	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	invocation := &domain.FunctionInvocation{
		ApplicationID: "app-func-1",
		VersionID:     "fv-1",
		InvocationID:  "inv-test-1",
		TriggerSource: "http",
		RequestMethod: "POST",
		RequestPath:   "/api/test",
		StartedAt:     time.Now(),
	}

	err := repo.CreateFunctionInvocation(ctx, invocation)
	assert.NoError(t, err)
	assert.NotEmpty(t, invocation.ID)
}

func TestGetFunctionInvocations(t *testing.T) {
	gormDB := setupFunctionTestDB(t)
	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	appID := "app-func-4"

	// Create test invocations
	for i := 0; i < 5; i++ {
		inv := &db.FunctionInvocation{
			ApplicationID:  appID,
			InvocationID:   "inv-" + string(rune('0'+i)),
			TriggerSource:  "http",
			ResponseStatus: 200,
			DurationMs:     100 + i*10,
			StartedAt:      time.Now().Add(time.Duration(-i) * time.Hour),
		}
		err := gormDB.Create(inv).Error
		require.NoError(t, err)
	}

	// Test pagination
	result, total, err := repo.GetFunctionInvocations(ctx, appID, 3, 0)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, 5, total)
}

func TestCreateFunctionEvent(t *testing.T) {
	gormDB := setupFunctionTestDB(t)
	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	event := &domain.FunctionEvent{
		ApplicationID: "app-func-1",
		EventType:     "webhook.github",
		EventSource:   "github",
		EventData: map[string]interface{}{
			"action": "push",
			"repo":   "test/repo",
		},
		ProcessingStatus: "pending",
	}

	err := repo.CreateFunctionEvent(ctx, event)
	assert.NoError(t, err)
	assert.NotEmpty(t, event.ID)
}

func TestGetPendingFunctionEvents(t *testing.T) {
	gormDB := setupFunctionTestDB(t)
	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	appID := "app-func-5"

	// Create test events with different statuses
	events := []db.FunctionEvent{
		{
			ApplicationID:    appID,
			EventType:        "test",
			EventSource:      "test",
			EventData:        db.JSON(`{"test": true}`),
			ProcessingStatus: "pending",
		},
		{
			ApplicationID:    appID,
			EventType:        "test",
			EventSource:      "test",
			EventData:        db.JSON(`{"test": true}`),
			ProcessingStatus: "retry",
			RetryCount:       1,
		},
		{
			ApplicationID:    appID,
			EventType:        "test",
			EventSource:      "test",
			EventData:        db.JSON(`{"test": true}`),
			ProcessingStatus: "success",
		},
	}

	for _, e := range events {
		err := gormDB.Create(&e).Error
		require.NoError(t, err)
	}

	// Test getting pending events
	result, err := repo.GetPendingFunctionEvents(ctx, appID, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 2) // Only pending and retry status
}

func TestUpdateFunctionInvocation(t *testing.T) {
	gormDB := setupFunctionTestDB(t)
	repo := NewPostgresRepository(gormDB)
	ctx := context.Background()

	// Create test invocation
	inv := &db.FunctionInvocation{
		ID:            "fi-test-1",
		ApplicationID: "app-func-1",
		InvocationID:  "inv-update-1",
		TriggerSource: "http",
		StartedAt:     time.Now(),
	}
	err := gormDB.Create(inv).Error
	require.NoError(t, err)

	// Update invocation
	completedAt := time.Now()
	update := &domain.FunctionInvocation{
		ID:              inv.ID,
		ResponseStatus:  200,
		ResponseBody:    "success",
		DurationMs:      150,
		CompletedAt:     &completedAt,
	}

	err = repo.UpdateFunctionInvocation(ctx, update)
	assert.NoError(t, err)

	// Verify update
	var updated db.FunctionInvocation
	err = gormDB.First(&updated, "id = ?", inv.ID).Error
	require.NoError(t, err)
	assert.Equal(t, 200, updated.ResponseStatus)
	assert.Equal(t, 150, updated.DurationMs)
	assert.NotNil(t, updated.CompletedAt)
}