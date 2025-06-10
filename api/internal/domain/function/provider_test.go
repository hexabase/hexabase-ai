package function_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hexabase/hexabase-ai/api/internal/domain/function"
	"github.com/hexabase/hexabase-ai/api/internal/repository/function/mock"
)

// TestProviderContract runs the contract test suite against a provider
func TestProviderContract(t *testing.T) {
	// Test with mock provider
	t.Run("MockProvider", func(t *testing.T) {
		provider := mock.NewFunctionProvider()
		RunProviderTests(t, provider)
	})
}

// RunProviderTests runs all provider tests
func RunProviderTests(t *testing.T, provider function.Provider) {
	t.Run("FunctionLifecycle", func(t *testing.T) {
		testFunctionLifecycle(t, provider)
	})
	t.Run("VersionManagement", func(t *testing.T) {
		testVersionManagement(t, provider)
	})
	t.Run("TriggerManagement", func(t *testing.T) {
		testTriggerManagement(t, provider)
	})
	t.Run("FunctionInvocation", func(t *testing.T) {
		testFunctionInvocation(t, provider)
	})
	t.Run("ErrorHandling", func(t *testing.T) {
		testErrorHandling(t, provider)
	})
	t.Run("ProviderCapabilities", func(t *testing.T) {
		testProviderCapabilities(t, provider)
	})
}

func testFunctionLifecycle(t *testing.T, provider function.Provider) {
	ctx := context.Background()
	
	// Create function
	spec := &function.FunctionSpec{
		Namespace:   "test-ns",
		Name:        "lifecycle-test",
		Runtime:     function.RuntimePython,
		Handler:     "main.handler",
		SourceCode:  "def handler(): pass",
		Environment: map[string]string{"TEST": "true"},
		Timeout:     30,
		Labels:      map[string]string{"app": "test"},
		Annotations: map[string]string{"description": "test function"},
	}
	
	created, err := provider.CreateFunction(ctx, spec)
	require.NoError(t, err)
	assert.Equal(t, spec.Namespace, created.Namespace)
	assert.Equal(t, spec.Name, created.Name)
	assert.Equal(t, spec.Runtime, created.Runtime)
	assert.NotEmpty(t, created.ActiveVersion)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)
	
	// Get function
	retrieved, err := provider.GetFunction(ctx, spec.Name)
	require.NoError(t, err)
	assert.Equal(t, created.Name, retrieved.Name)
	
	// List functions
	functions, err := provider.ListFunctions(ctx, spec.Namespace)
	require.NoError(t, err)
	assert.True(t, len(functions) > 0)
	found := false
	for _, fn := range functions {
		if fn.Name == spec.Name {
			found = true
			break
		}
	}
	assert.True(t, found, "Created function should be in list")
	
	// Get the function before update to capture the original UpdatedAt
	beforeUpdate, err := provider.GetFunction(ctx, spec.Name)
	require.NoError(t, err)
	originalUpdateTime := beforeUpdate.UpdatedAt
	
	// Wait a bit to ensure UpdatedAt will be different
	time.Sleep(100 * time.Millisecond)
	
	// Update function
	updateSpec := &function.FunctionSpec{
		Namespace:   spec.Namespace,
		Name:        spec.Name,
		Runtime:     function.RuntimePython38,
		Handler:     "main.new_handler",
		SourceCode:  "def new_handler(): pass",
		Environment: map[string]string{"TEST": "false", "NEW": "var"},
	}
	
	updated, err := provider.UpdateFunction(ctx, spec.Name, updateSpec)
	require.NoError(t, err)
	assert.Equal(t, function.RuntimePython38, updated.Runtime)
	assert.NotZero(t, updated.UpdatedAt)
	assert.True(t, updated.UpdatedAt.After(originalUpdateTime), 
		"UpdatedAt should be after original UpdatedAt: updated=%v, original=%v", 
		updated.UpdatedAt, originalUpdateTime)
	
	// Delete function
	err = provider.DeleteFunction(ctx, spec.Name)
	require.NoError(t, err)
	
	// Verify deletion
	_, err = provider.GetFunction(ctx, spec.Name)
	assert.Error(t, err)
}

func testVersionManagement(t *testing.T, provider function.Provider) {
	ctx := context.Background()
	
	// Create function
	spec := &function.FunctionSpec{
		Namespace: "test-ns",
		Name:      "version-test",
		Runtime:   function.RuntimeNode,
		Handler:   "index.handler",
		SourceCode:    "exports.handler = () => 'v1'",
	}
	
	fn, err := provider.CreateFunction(ctx, spec)
	require.NoError(t, err)
	initialVersion := fn.ActiveVersion
	
	// List initial versions
	versions, err := provider.ListVersions(ctx, spec.Name)
	require.NoError(t, err)
	assert.Equal(t, 1, len(versions))
	
	// Create new version
	newVersion := &function.FunctionVersionDef{
		FunctionName: spec.Name,
		SourceCode:   "exports.handler = () => 'v2'",
	}
	
	err = provider.CreateVersion(ctx, spec.Name, newVersion)
	require.NoError(t, err)
	
	// List versions to get the new version ID
	versions, err = provider.ListVersions(ctx, spec.Name)
	require.NoError(t, err)
	assert.Equal(t, 2, len(versions))
	
	// Get the second version (most recent)
	v2 := versions[1]
	retrieved, err := provider.GetVersion(ctx, spec.Name, v2.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, retrieved.Version)
	
	// Set active version
	err = provider.SetActiveVersion(ctx, spec.Name, v2.ID)
	require.NoError(t, err)
	
	// Verify active version changed
	fn, err = provider.GetFunction(ctx, spec.Name)
	require.NoError(t, err)
	assert.Equal(t, v2.ID, fn.ActiveVersion)
	assert.NotEqual(t, initialVersion, fn.ActiveVersion)
	
	// Cleanup
	err = provider.DeleteFunction(ctx, spec.Name)
	require.NoError(t, err)
}

func testTriggerManagement(t *testing.T, provider function.Provider) {
	ctx := context.Background()
	
	// Create function
	spec := &function.FunctionSpec{
		Namespace: "test-ns",
		Name:      "trigger-test",
		Runtime:   function.RuntimeGo,
		Handler:   "main",
		SourceCode: "package main\n\nfunc main() {}",
	}
	
	_, err := provider.CreateFunction(ctx, spec)
	require.NoError(t, err)
	
	// Test different trigger types
	triggers := []*function.FunctionTrigger{
		{
			Name: "http-trigger",
			Type: function.TriggerHTTP,
			Config: map[string]string{
				"path":   "/api/test",
				"method": "POST",
			},
		},
		{
			Name: "schedule-trigger",
			Type: function.TriggerSchedule,
			Config: map[string]string{
				"cron": "*/5 * * * *",
			},
		},
	}
	
	// Create triggers
	for _, trigger := range triggers {
		err := provider.CreateTrigger(ctx, spec.Name, trigger)
		require.NoError(t, err)
	}
	
	// List triggers
	listed, err := provider.ListTriggers(ctx, spec.Name)
	require.NoError(t, err)
	assert.Equal(t, len(triggers), len(listed))
	
	// Update trigger
	updatedTrigger := &function.FunctionTrigger{
		Name: "http-trigger",
		Type: function.TriggerHTTP,
		Config: map[string]string{
			"path":   "/api/v2/test",
			"method": "GET",
		},
	}
	
	err = provider.UpdateTrigger(ctx, spec.Name, "http-trigger", updatedTrigger)
	require.NoError(t, err)
	
	// Delete trigger
	err = provider.DeleteTrigger(ctx, spec.Name, "schedule-trigger")
	require.NoError(t, err)
	
	// Verify deletion
	listed, err = provider.ListTriggers(ctx, spec.Name)
	require.NoError(t, err)
	assert.Equal(t, 1, len(listed))
	assert.Equal(t, "http-trigger", listed[0].Name)
	
	// Cleanup
	err = provider.DeleteFunction(ctx, spec.Name)
	require.NoError(t, err)
}

func testFunctionInvocation(t *testing.T, provider function.Provider) {
	ctx := context.Background()
	
	// Create function
	spec := &function.FunctionSpec{
		Namespace: "test-ns",
		Name:      "invoke-test",
		Runtime:   function.RuntimePython,
		Handler:   "main.handler",
		SourceCode:    "def handler(event): return {'status': 'ok'}",
	}
	
	_, err := provider.CreateFunction(ctx, spec)
	require.NoError(t, err)
	
	// Invoke function synchronously
	request := &function.InvokeRequest{
		Method: "POST",
		Path:   "/",
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: []byte(`{"test": true}`),
	}
	
	response, err := provider.InvokeFunction(ctx, spec.Name, request)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.InvocationID)
	assert.True(t, response.Duration > 0)
	
	// Test async invocation
	invocationID, err := provider.InvokeFunctionAsync(ctx, spec.Name, request)
	require.NoError(t, err)
	assert.NotEmpty(t, invocationID)
	
	// Get invocation status
	status, err := provider.GetInvocationStatus(ctx, invocationID)
	require.NoError(t, err)
	assert.Equal(t, invocationID, status.InvocationID)
	
	// Cleanup
	err = provider.DeleteFunction(ctx, spec.Name)
	require.NoError(t, err)
}

func testErrorHandling(t *testing.T, provider function.Provider) {
	ctx := context.Background()
	
	// Test operations on non-existent function
	_, err := provider.GetFunction(ctx, "non-existent")
	assert.Error(t, err)
	
	err = provider.DeleteFunction(ctx, "non-existent")
	assert.Error(t, err)
	
	_, err = provider.ListVersions(ctx, "non-existent")
	assert.Error(t, err)
	
	err = provider.SetActiveVersion(ctx, "non-existent", "v1")
	assert.Error(t, err)
	
	// Test duplicate creation
	spec := &function.FunctionSpec{
		Namespace: "test-ns",
		Name:      "duplicate-test",
		Runtime:   function.RuntimeNode,
		Handler:   "index.handler",
		SourceCode: "exports.handler = () => {}",
	}
	
	_, err = provider.CreateFunction(ctx, spec)
	require.NoError(t, err)
	
	_, err = provider.CreateFunction(ctx, spec)
	assert.Error(t, err)
	
	// Cleanup
	err = provider.DeleteFunction(ctx, spec.Name)
	require.NoError(t, err)
}

func testProviderCapabilities(t *testing.T, provider function.Provider) {
	caps := provider.GetCapabilities()
	assert.NotNil(t, caps)
	
	// Basic capabilities should be defined
	assert.NotEmpty(t, caps.Name)
	assert.NotEmpty(t, caps.Version)
	assert.NotEmpty(t, caps.SupportedRuntimes)
	assert.NotEmpty(t, caps.SupportedTriggerTypes)
	assert.True(t, caps.MaxTimeoutSecs > 0)
	assert.True(t, caps.MaxMemoryMB > 0)
	
	// Test helper methods
	assert.True(t, caps.HasRuntime(function.RuntimePython))
	assert.True(t, caps.HasTriggerType(function.TriggerHTTP))
}

func TestProviderHealthCheck(t *testing.T) {
	ctx := context.Background()
	provider := mock.NewFunctionProvider()
	
	err := provider.HealthCheck(ctx)
	assert.NoError(t, err)
}