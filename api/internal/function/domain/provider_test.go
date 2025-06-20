package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hexabase/hexabase-ai/api/internal/function/domain"
	"github.com/hexabase/hexabase-ai/api/internal/function/repository/mock"
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
func RunProviderTests(t *testing.T, provider domain.Provider) {
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

func testFunctionLifecycle(t *testing.T, provider domain.Provider) {
	ctx := context.Background()
	
	// Create function
	spec := &domain.FunctionSpec{
		Namespace:   "test-ns",
		Name:        "lifecycle-test",
		Runtime:     domain.RuntimePython,
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
	updateSpec := &domain.FunctionSpec{
		Namespace:   spec.Namespace,
		Name:        spec.Name,
		Runtime:     domain.RuntimePython38,
		Handler:     "main.new_handler",
		SourceCode:  "def new_handler(): pass",
		Environment: map[string]string{"TEST": "false", "NEW": "var"},
	}
	
	updated, err := provider.UpdateFunction(ctx, spec.Name, updateSpec)
	require.NoError(t, err)
	assert.Equal(t, domain.RuntimePython38, updated.Runtime)
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

func testVersionManagement(t *testing.T, provider domain.Provider) {
	ctx := context.Background()
	
	// Create function
	spec := &domain.FunctionSpec{
		Namespace: "test-ns",
		Name:      "version-test",
		Runtime:   domain.RuntimeNode,
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
	newVersion := &domain.FunctionVersionDef{
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

func testTriggerManagement(t *testing.T, provider domain.Provider) {
	ctx := context.Background()
	
	// Create function
	spec := &domain.FunctionSpec{
		Namespace: "test-ns",
		Name:      "trigger-test",
		Runtime:   domain.RuntimeGo,
		Handler:   "main",
		SourceCode: "package main\n\nfunc main() {}",
	}
	
	_, err := provider.CreateFunction(ctx, spec)
	require.NoError(t, err)
	
	// Test different trigger types
	triggers := []*domain.FunctionTrigger{
		{
			Name: "http-trigger",
			Type: domain.TriggerHTTP,
			Config: map[string]string{
				"method": "GET",
				"path":   "/test",
			},
		},
		{
			Name: "schedule-trigger",
			Type: domain.TriggerSchedule,
			Config: map[string]string{
				"cron": "0 */5 * * * *",
			},
		},
		{
			Name: "event-trigger",
			Type: domain.TriggerEvent,
			Config: map[string]string{
				"event_type": "test.event",
				"source":     "test-source",
			},
		},
	}
	
	// Test trigger lifecycle for each type
	for _, trigger := range triggers {
		t.Run(string(trigger.Type), func(t *testing.T) {
			// Create trigger
		err := provider.CreateTrigger(ctx, spec.Name, trigger)
			if err != nil {
				// Some providers may not support all trigger types
				if provErr, ok := err.(*domain.ProviderError); ok && provErr.Code == domain.ErrCodeNotSupported {
					t.Skipf("Provider does not support trigger type %s", trigger.Type)
					return
				}
		require.NoError(t, err)
	}
	
	// List triggers
			listTriggers, err := provider.ListTriggers(ctx, spec.Name)
	require.NoError(t, err)
			
			found := false
			for _, tr := range listTriggers {
				if tr.Name == trigger.Name {
					found = true
					assert.Equal(t, trigger.Type, tr.Type)
					break
				}
			}
			assert.True(t, found, "Created trigger should be in list")
	
	// Update trigger
			updatedTrigger := *trigger
			updatedTrigger.Config = map[string]string{
				"updated": "true",
			}
			err = provider.UpdateTrigger(ctx, spec.Name, trigger.Name, &updatedTrigger)
			if err != nil {
				// Some providers may not support trigger updates
				if provErr, ok := err.(*domain.ProviderError); ok && provErr.Code == domain.ErrCodeNotSupported {
					t.Logf("Provider does not support trigger updates for type %s", trigger.Type)
				} else {
	require.NoError(t, err)
				}
			}
	
	// Delete trigger
			err = provider.DeleteTrigger(ctx, spec.Name, trigger.Name)
	require.NoError(t, err)
		})
	}
	
	// Cleanup
	err = provider.DeleteFunction(ctx, spec.Name)
	require.NoError(t, err)
}

func testFunctionInvocation(t *testing.T, provider domain.Provider) {
	ctx := context.Background()
	
	// Create function
	spec := &domain.FunctionSpec{
		Namespace:  "test-ns",
		Name:       "invoke-test",
		Runtime:    domain.RuntimePython,
		Handler:    "main.handler",
		SourceCode: "def handler(): return {'status': 'ok'}",
	}
	
	_, err := provider.CreateFunction(ctx, spec)
	require.NoError(t, err)
	
	// Test synchronous invocation
	invokeReq := &domain.InvokeRequest{
		Method: "POST",
		Body:   []byte(`{"test": "data"}`),
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
	}
	
	response, err := provider.InvokeFunction(ctx, spec.Name, invokeReq)
	require.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.NotEmpty(t, response.Body)
	assert.Greater(t, response.Duration, time.Duration(0))
	
	// Test asynchronous invocation (if supported)
	invocationID, err := provider.InvokeFunctionAsync(ctx, spec.Name, invokeReq)
	if err != nil {
		// Some providers may not support async invocation
		if provErr, ok := err.(*domain.ProviderError); ok && provErr.Code == domain.ErrCodeNotSupported {
			t.Skip("Provider does not support async invocation")
		} else {
	require.NoError(t, err)
		}
	} else {
	assert.NotEmpty(t, invocationID)
	
	// Get invocation status
	status, err := provider.GetInvocationStatus(ctx, invocationID)
	require.NoError(t, err)
		assert.NotEmpty(t, status.Status)
	assert.Equal(t, invocationID, status.InvocationID)
	}
	
	// Cleanup
	err = provider.DeleteFunction(ctx, spec.Name)
	require.NoError(t, err)
}

func testErrorHandling(t *testing.T, provider domain.Provider) {
	ctx := context.Background()
	
	// Test getting non-existent function
	_, err := provider.GetFunction(ctx, "non-existent")
	assert.Error(t, err)
	if provErr, ok := err.(*domain.ProviderError); ok {
		assert.True(t, provErr.IsNotFound())
	}
	
	// Test creating function with invalid spec
	invalidSpec := &domain.FunctionSpec{
		Name: "", // Empty name should cause error
	}
	
	_, err = provider.CreateFunction(ctx, invalidSpec)
	assert.Error(t, err)
	
	// Test updating non-existent function
	updateSpec := &domain.FunctionSpec{
		Namespace: "test-ns",
		Name:      "non-existent",
		Runtime:   domain.RuntimePython,
		Handler:   "main.handler",
	}
	
	_, err = provider.UpdateFunction(ctx, "non-existent", updateSpec)
	assert.Error(t, err)
	
	// Test deleting non-existent function
	err = provider.DeleteFunction(ctx, "non-existent")
	// Some providers may not error on deleting non-existent functions
	// so we don't assert error here
}

func testProviderCapabilities(t *testing.T, provider domain.Provider) {
	capabilities := provider.GetCapabilities()
	require.NotNil(t, capabilities)
	assert.NotEmpty(t, capabilities.Name)
	assert.NotEmpty(t, capabilities.Version)
	assert.NotEmpty(t, capabilities.SupportedRuntimes)
	assert.NotEmpty(t, capabilities.SupportedTriggerTypes)
}

func TestProviderHealthCheck(t *testing.T) {
	ctx := context.Background()
	provider := mock.NewFunctionProvider()
	
	err := provider.HealthCheck(ctx)
	assert.NoError(t, err)
}