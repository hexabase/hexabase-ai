package mock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hexabase/hexabase-ai/api/internal/function/domain"
)

func TestMockProvider_FunctionLifecycle(t *testing.T) {
	ctx := context.Background()
	provider := NewFunctionProvider()

	// Create function
	spec := &domain.FunctionSpec{
		Name:      "test-function",
		Namespace: "test-namespace",
		Runtime:   domain.RuntimePython,
		Handler:   "main.handler",
		SourceCode: "def handler(event, context):\n    return {'statusCode': 200}",
		Environment: map[string]string{
			"ENV_VAR": "test",
		},
		Resources: domain.FunctionResourceRequirements{
			Memory: "256Mi",
			CPU:    "100m",
		},
		Timeout: 30,
		Labels: map[string]string{
			"app": "test",
		},
	}

	// Test CreateFunction
	fn, err := provider.CreateFunction(ctx, spec)
	require.NoError(t, err)
	assert.NotNil(t, fn)
	assert.Equal(t, spec.Name, fn.Name)
	assert.Equal(t, spec.Namespace, fn.Namespace)
	assert.Equal(t, spec.Runtime, fn.Runtime)
	assert.Equal(t, spec.Handler, fn.Handler)
	assert.Equal(t, domain.FunctionDefStatusReady, fn.Status)
	assert.NotEmpty(t, fn.ActiveVersion)
	assert.NotZero(t, fn.CreatedAt)

	// Test GetFunction
	retrieved, err := provider.GetFunction(ctx, spec.Name)
	require.NoError(t, err)
	assert.Equal(t, fn.Name, retrieved.Name)
	assert.Equal(t, fn.Namespace, retrieved.Namespace)

	// Test ListFunctions
	functions, err := provider.ListFunctions(ctx, spec.Namespace)
	require.NoError(t, err)
	assert.Len(t, functions, 1)
	assert.Equal(t, fn.Name, functions[0].Name)

	// Test UpdateFunction
	time.Sleep(100 * time.Millisecond) // Ensure UpdatedAt will be different
	updateSpec := &domain.FunctionSpec{
		Name:      spec.Name,
		Namespace: spec.Namespace,
		Runtime:   domain.RuntimePython38,
		Handler:   "app.handler",
		SourceCode: "def handler(event, context):\n    return {'statusCode': 201}",
	}

	updated, err := provider.UpdateFunction(ctx, spec.Name, updateSpec)
	require.NoError(t, err)
	assert.Equal(t, domain.RuntimePython38, updated.Runtime)
	assert.Equal(t, "app.handler", updated.Handler)
	assert.True(t, updated.UpdatedAt.After(fn.UpdatedAt), 
		"UpdatedAt should be after CreatedAt: updated=%v, created=%v", 
		updated.UpdatedAt, fn.UpdatedAt)

	// Test DeleteFunction
	err = provider.DeleteFunction(ctx, spec.Name)
	require.NoError(t, err)

	// Verify deletion
	_, err = provider.GetFunction(ctx, spec.Name)
	assert.Error(t, err)
	perr, ok := err.(*domain.ProviderError)
	assert.True(t, ok)
	assert.True(t, perr.IsNotFound())
}

func TestMockProvider_VersionManagement(t *testing.T) {
	ctx := context.Background()
	provider := NewFunctionProvider()

	// Create function first
	spec := &domain.FunctionSpec{
		Name:       "version-test",
		Namespace:  "test-namespace",
		Runtime:    domain.RuntimeNode,
		Handler:    "index.handler",
		SourceCode: "exports.handler = async () => ({ statusCode: 200 });",
	}

	fn, err := provider.CreateFunction(ctx, spec)
	require.NoError(t, err)
	initialVersion := fn.ActiveVersion

	// List initial versions
	versions, err := provider.ListVersions(ctx, spec.Name)
	require.NoError(t, err)
	assert.Len(t, versions, 1)
	assert.Equal(t, initialVersion, versions[0].ID)

	// Create new version
	newVersion := &domain.FunctionVersionDef{
		FunctionName: spec.Name,
		SourceCode:   "exports.handler = async () => ({ statusCode: 201 });",
		Image:        "node:16-alpine",
	}

	err = provider.CreateVersion(ctx, spec.Name, newVersion)
	require.NoError(t, err)

	// List versions again
	versions, err = provider.ListVersions(ctx, spec.Name)
	require.NoError(t, err)
	assert.Len(t, versions, 2)

	// Get specific version
	v2 := versions[1]
	retrieved, err := provider.GetVersion(ctx, spec.Name, v2.ID)
	require.NoError(t, err)
	assert.Equal(t, v2.ID, retrieved.ID)
	assert.Equal(t, 2, retrieved.Version)

	// Set active version
	err = provider.SetActiveVersion(ctx, spec.Name, v2.ID)
	require.NoError(t, err)

	// Verify active version changed
	fn, err = provider.GetFunction(ctx, spec.Name)
	require.NoError(t, err)
	assert.Equal(t, v2.ID, fn.ActiveVersion)
	assert.NotEqual(t, initialVersion, fn.ActiveVersion)
}

func TestMockProvider_TriggerManagement(t *testing.T) {
	ctx := context.Background()
	provider := NewFunctionProvider()

	// Create function first
	spec := &domain.FunctionSpec{
		Name:       "trigger-test",
		Namespace:  "test-namespace",
		Runtime:    domain.RuntimeGo,
		Handler:    "main",
		SourceCode: "package main\n\nfunc main() {}",
	}

	_, err := provider.CreateFunction(ctx, spec)
	require.NoError(t, err)

	// Create HTTP trigger
	httpTrigger := &domain.FunctionTrigger{
		Name:    "http-trigger",
		Type:    domain.TriggerHTTP,
		Enabled: true,
		Config: map[string]string{
			"path":   "/api/test",
			"method": "POST",
		},
	}

	err = provider.CreateTrigger(ctx, spec.Name, httpTrigger)
	require.NoError(t, err)

	// Create schedule trigger
	scheduleTrigger := &domain.FunctionTrigger{
		Name:    "schedule-trigger",
		Type:    domain.TriggerSchedule,
		Enabled: true,
		Config: map[string]string{
			"cron": "0 */5 * * *",
		},
	}

	err = provider.CreateTrigger(ctx, spec.Name, scheduleTrigger)
	require.NoError(t, err)

	// List triggers
	triggers, err := provider.ListTriggers(ctx, spec.Name)
	require.NoError(t, err)
	assert.Len(t, triggers, 2)

	// Update trigger
	httpTrigger.Config["method"] = "GET"
	err = provider.UpdateTrigger(ctx, spec.Name, "http-trigger", httpTrigger)
	require.NoError(t, err)

	// Verify update
	triggers, err = provider.ListTriggers(ctx, spec.Name)
	require.NoError(t, err)
	for _, tr := range triggers {
		if tr.Name == "http-trigger" {
			assert.Equal(t, "GET", tr.Config["method"])
			break
		}
	}

	// Delete trigger
	err = provider.DeleteTrigger(ctx, spec.Name, "schedule-trigger")
	require.NoError(t, err)

	// Verify deletion
	triggers, err = provider.ListTriggers(ctx, spec.Name)
	require.NoError(t, err)
	assert.Len(t, triggers, 1)
	assert.Equal(t, "http-trigger", triggers[0].Name)
}

func TestMockProvider_Invocation(t *testing.T) {
	ctx := context.Background()
	provider := NewFunctionProvider()

	// Create function
	spec := &domain.FunctionSpec{
		Name:       "invoke-test",
		Namespace:  "test-namespace",
		Runtime:    domain.RuntimePython,
		Handler:    "main.handler",
		SourceCode: "def handler(event, context):\n    return {'statusCode': 200}",
	}

	_, err := provider.CreateFunction(ctx, spec)
	require.NoError(t, err)

	// Test synchronous invocation
	req := &domain.InvokeRequest{
		Method: "POST",
		Path:   "/test",
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: []byte(`{"test": "data"}`),
	}

	resp, err := provider.InvokeFunction(ctx, spec.Name, req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotEmpty(t, resp.Body)
	assert.NotEmpty(t, resp.InvocationID)
	assert.Greater(t, resp.Duration, time.Duration(0))

	// Test asynchronous invocation
	invocationID, err := provider.InvokeFunctionAsync(ctx, spec.Name, req)
	require.NoError(t, err)
	assert.NotEmpty(t, invocationID)

	// Check invocation status (initially running)
	status, err := provider.GetInvocationStatus(ctx, invocationID)
	require.NoError(t, err)
	assert.Equal(t, invocationID, status.InvocationID)
	assert.Equal(t, "running", status.Status)
	assert.NotZero(t, status.StartedAt)

	// Wait for completion
	time.Sleep(150 * time.Millisecond)

	// Check invocation status (should be completed)
	status, err = provider.GetInvocationStatus(ctx, invocationID)
	require.NoError(t, err)
	assert.Equal(t, "completed", status.Status)
	assert.NotNil(t, status.CompletedAt)
	assert.NotNil(t, status.Result)
	assert.Equal(t, 200, status.Result.StatusCode)

	// Test function URL
	url, err := provider.GetFunctionURL(ctx, spec.Name)
	require.NoError(t, err)
	assert.Contains(t, url, spec.Name)
}

func TestMockProvider_LogsAndMetrics(t *testing.T) {
	ctx := context.Background()
	provider := NewFunctionProvider()

	// Create function
	spec := &domain.FunctionSpec{
		Name:       "logs-test",
		Namespace:  "test-namespace",
		Runtime:    domain.RuntimeNode,
		Handler:    "index.handler",
		SourceCode: "exports.handler = async () => ({ statusCode: 200 });",
	}

	_, err := provider.CreateFunction(ctx, spec)
	require.NoError(t, err)

	// Invoke function to generate some activity
	req := &domain.InvokeRequest{
		Method: "GET",
		Path:   "/",
	}
	_, err = provider.InvokeFunction(ctx, spec.Name, req)
	require.NoError(t, err)

	// Get logs
	logOpts := &domain.LogOptions{
		Limit: 10,
	}
	logs, err := provider.GetFunctionLogs(ctx, spec.Name, logOpts)
	require.NoError(t, err)
	assert.NotEmpty(t, logs)
	assert.Greater(t, len(logs), 0)

	// Get metrics
	metricOpts := &domain.MetricOptions{
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   time.Now(),
	}
	metrics, err := provider.GetFunctionMetrics(ctx, spec.Name, metricOpts)
	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Greater(t, metrics.Invocations, int64(0))
	assert.Greater(t, metrics.Duration.Avg, float64(0))
}

func TestMockProvider_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	provider := NewFunctionProvider()

	// Test creating function with empty name
	spec := &domain.FunctionSpec{
		Name:      "",
		Namespace: "test-namespace",
		Runtime:   domain.RuntimePython,
		Handler:   "main.handler",
	}

	_, err := provider.CreateFunction(ctx, spec)
	assert.Error(t, err)
	perr, ok := err.(*domain.ProviderError)
	assert.True(t, ok)
	assert.Equal(t, domain.ErrCodeInvalidInput, perr.Code)

	// Test getting non-existent function
	_, err = provider.GetFunction(ctx, "non-existent")
	assert.Error(t, err)
	perr, ok = err.(*domain.ProviderError)
	assert.True(t, ok)
	assert.True(t, perr.IsNotFound())

	// Test duplicate function creation
	spec.Name = "duplicate-test"
	_, err = provider.CreateFunction(ctx, spec)
	require.NoError(t, err)

	_, err = provider.CreateFunction(ctx, spec)
	assert.Error(t, err)
	perr, ok = err.(*domain.ProviderError)
	assert.True(t, ok)
	assert.True(t, perr.IsAlreadyExists())

	// Test failure mode
	provider.SetFailureMode(true)
	_, err = provider.GetFunction(ctx, "duplicate-test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated failure")
}

func TestMockProvider_Capabilities(t *testing.T) {
	provider := NewFunctionProvider()
	capabilities := provider.GetCapabilities()

	assert.NotNil(t, capabilities)
	assert.Equal(t, "mock", capabilities.Name)
	assert.NotEmpty(t, capabilities.Version)
	assert.NotEmpty(t, capabilities.SupportedRuntimes)
	assert.NotEmpty(t, capabilities.SupportedTriggerTypes)
	assert.True(t, capabilities.SupportsVersioning)
	assert.True(t, capabilities.SupportsAsync)
	assert.True(t, capabilities.SupportsLogs)
	assert.True(t, capabilities.SupportsMetrics)
	assert.Greater(t, capabilities.MaxMemoryMB, 0)
	assert.Greater(t, capabilities.MaxTimeoutSecs, 0)
}

func TestMockProvider_HealthCheck(t *testing.T) {
	ctx := context.Background()
	provider := NewFunctionProvider()

	// Normal health check
	err := provider.HealthCheck(ctx)
	assert.NoError(t, err)

	// Health check with failure mode
	provider.SetFailureMode(true)
	err = provider.HealthCheck(ctx)
	assert.Error(t, err)
}

func TestMockProvider_SimulatedLatency(t *testing.T) {
	ctx := context.Background()
	provider := NewFunctionProvider()
	provider.SetSimulatedLatency(100 * time.Millisecond)

	spec := &domain.FunctionSpec{
		Name:       "latency-test",
		Namespace:  "test-namespace",
		Runtime:    domain.RuntimeGo,
		Handler:    "main",
		SourceCode: "package main\n\nfunc main() {}",
	}

	start := time.Now()
	_, err := provider.CreateFunction(ctx, spec)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
}