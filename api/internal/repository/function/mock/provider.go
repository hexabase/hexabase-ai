package mock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/function"
)

// FunctionProvider implements the function.Provider interface for testing
type FunctionProvider struct {
	functions        map[string]*function.FunctionDef
	versions         map[string][]*function.FunctionVersionDef
	triggers         map[string][]*function.FunctionTrigger
	invocations      map[string]*function.InvocationStatus
	mu               sync.RWMutex
	failureMode      bool
	simulatedLatency time.Duration
	capabilities     *function.Capabilities
	invocationCount  int
}

// NewFunctionProvider creates a new mock provider instance
func NewFunctionProvider() *FunctionProvider {
	return &FunctionProvider{
		functions:   make(map[string]*function.FunctionDef),
		versions:    make(map[string][]*function.FunctionVersionDef),
		triggers:    make(map[string][]*function.FunctionTrigger),
		invocations: make(map[string]*function.InvocationStatus),
		capabilities: &function.Capabilities{
			Name:        "mock",
			Version:     "1.0.0",
			Description: "Mock provider for testing",
			SupportedRuntimes: []function.Runtime{
				function.RuntimeGo,
				function.RuntimePython,
				function.RuntimeNode,
			},
			SupportedTriggerTypes: []function.TriggerType{
				function.TriggerHTTP,
				function.TriggerSchedule,
				function.TriggerEvent,
			},
			SupportsVersioning:      true,
			SupportsAsync:           true,
			SupportsLogs:            true,
			SupportsMetrics:         true,
			SupportsEnvironmentVars: true,
			MaxMemoryMB:             2048,
			MaxTimeoutSecs:          900,
			TypicalColdStartMs:      100,
		},
	}
}

// SetFailureMode enables or disables failure mode for testing
func (m *FunctionProvider) SetFailureMode(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failureMode = enabled
}

// SetSimulatedLatency sets artificial latency for operations
func (m *FunctionProvider) SetSimulatedLatency(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.simulatedLatency = latency
}

// CreateFunction creates a new function
func (m *FunctionProvider) CreateFunction(ctx context.Context, spec *function.FunctionSpec) (*function.FunctionDef, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return nil, function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Validate input
	if spec.Name == "" {
		return nil, function.NewProviderError(function.ErrCodeInvalidInput, "function name is required")
	}

	key := fmt.Sprintf("%s/%s", spec.Namespace, spec.Name)
	if _, exists := m.functions[key]; exists {
		return nil, function.NewProviderError(function.ErrCodeAlreadyExists, "function already exists")
	}

	// Create function
	fn := &function.FunctionDef{
		Name:        spec.Name,
		Namespace:   spec.Namespace,
		Runtime:     spec.Runtime,
		Handler:     spec.Handler,
		Status:      function.FunctionDefStatusReady,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Labels:      spec.Labels,
		Annotations: spec.Annotations,
	}

	m.functions[key] = fn

	// Create initial version
	version := &function.FunctionVersionDef{
		ID:           fmt.Sprintf("v%d-%d", 1, time.Now().Unix()),
		FunctionName: spec.Name,
		Version:      1,
		SourceCode:   spec.SourceCode,
		Image:        spec.Image,
		BuildStatus:  function.FunctionBuildStatusSuccess,
		CreatedAt:    time.Now(),
		IsActive:     true,
	}
	m.versions[key] = []*function.FunctionVersionDef{version}
	fn.ActiveVersion = version.ID

	// Return a copy of the function
	result := *fn
	return &result, nil
}

// UpdateFunction updates an existing function
func (m *FunctionProvider) UpdateFunction(ctx context.Context, name string, spec *function.FunctionSpec) (*function.FunctionDef, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return nil, function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	key := fmt.Sprintf("%s/%s", spec.Namespace, name)
	fn, exists := m.functions[key]
	if !exists {
		return nil, function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	// Update function
	fn.Runtime = spec.Runtime
	fn.Handler = spec.Handler
	fn.UpdatedAt = time.Now()
	if spec.Labels != nil {
		fn.Labels = spec.Labels
	}
	if spec.Annotations != nil {
		fn.Annotations = spec.Annotations
	}

	// Return a copy of the function
	result := *fn
	return &result, nil
}

// DeleteFunction deletes a function
func (m *FunctionProvider) DeleteFunction(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Find and delete function across all namespaces
	deleted := false
	for key := range m.functions {
		if m.functions[key].Name == name {
			delete(m.functions, key)
			delete(m.versions, key)
			delete(m.triggers, key)
			deleted = true
		}
	}

	if !deleted {
		return function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	return nil
}

// GetFunction retrieves a function by name
func (m *FunctionProvider) GetFunction(ctx context.Context, name string) (*function.FunctionDef, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return nil, function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Search for function by name
	for _, fn := range m.functions {
		if fn.Name == name {
			return fn, nil
		}
	}

	return nil, function.NewProviderError(function.ErrCodeNotFound, "function not found")
}

// ListFunctions lists all functions in a namespace
func (m *FunctionProvider) ListFunctions(ctx context.Context, namespace string) ([]*function.FunctionDef, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return nil, function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	var functions []*function.FunctionDef
	for _, fn := range m.functions {
		if fn.Namespace == namespace || namespace == "" {
			functions = append(functions, fn)
		}
	}

	return functions, nil
}

// CreateVersion creates a new version of a function
func (m *FunctionProvider) CreateVersion(ctx context.Context, functionName string, version *function.FunctionVersionDef) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Find function
	var key string
	for k, fn := range m.functions {
		if fn.Name == functionName {
			key = k
			break
		}
	}

	if key == "" {
		return function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	// Add version
	versions := m.versions[key]
	version.Version = len(versions) + 1
	version.ID = fmt.Sprintf("v%d-%d", version.Version, time.Now().Unix())
	version.CreatedAt = time.Now()
	version.BuildStatus = function.FunctionBuildStatusSuccess

	m.versions[key] = append(versions, version)

	return nil
}

// GetVersion retrieves a specific version of a function
func (m *FunctionProvider) GetVersion(ctx context.Context, functionName, versionID string) (*function.FunctionVersionDef, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return nil, function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Find function
	var key string
	for k, fn := range m.functions {
		if fn.Name == functionName {
			key = k
			break
		}
	}

	if key == "" {
		return nil, function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	// Find version
	for _, v := range m.versions[key] {
		if v.ID == versionID {
			return v, nil
		}
	}

	return nil, function.NewProviderError(function.ErrCodeNotFound, "version not found")
}

// ListVersions lists all versions of a function
func (m *FunctionProvider) ListVersions(ctx context.Context, functionName string) ([]*function.FunctionVersionDef, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return nil, function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Find function
	var key string
	for k, fn := range m.functions {
		if fn.Name == functionName {
			key = k
			break
		}
	}

	if key == "" {
		return nil, function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	return m.versions[key], nil
}

// SetActiveVersion sets the active version of a function
func (m *FunctionProvider) SetActiveVersion(ctx context.Context, functionName, versionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Find function
	var fn *function.FunctionDef
	var key string
	for k, f := range m.functions {
		if f.Name == functionName {
			fn = f
			key = k
			break
		}
	}

	if fn == nil {
		return function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	// Verify version exists
	versionExists := false
	for _, v := range m.versions[key] {
		v.IsActive = false // Deactivate all versions
		if v.ID == versionID {
			v.IsActive = true
			versionExists = true
		}
	}

	if !versionExists {
		return function.NewProviderError(function.ErrCodeNotFound, "version not found")
	}

	fn.ActiveVersion = versionID
	fn.UpdatedAt = time.Now()

	return nil
}

// CreateTrigger creates a new trigger for a function
func (m *FunctionProvider) CreateTrigger(ctx context.Context, functionName string, trigger *function.FunctionTrigger) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Find function
	var key string
	for k, fn := range m.functions {
		if fn.Name == functionName {
			key = k
			break
		}
	}

	if key == "" {
		return function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	// Check if trigger already exists
	for _, t := range m.triggers[key] {
		if t.Name == trigger.Name {
			return function.NewProviderError(function.ErrCodeAlreadyExists, "trigger already exists")
		}
	}

	trigger.FunctionName = functionName
	trigger.CreatedAt = time.Now()
	trigger.UpdatedAt = time.Now()

	m.triggers[key] = append(m.triggers[key], trigger)

	return nil
}

// UpdateTrigger updates an existing trigger
func (m *FunctionProvider) UpdateTrigger(ctx context.Context, functionName, triggerName string, trigger *function.FunctionTrigger) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Find function
	var key string
	for k, fn := range m.functions {
		if fn.Name == functionName {
			key = k
			break
		}
	}

	if key == "" {
		return function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	// Find and update trigger
	for i, t := range m.triggers[key] {
		if t.Name == triggerName {
			trigger.Name = triggerName
			trigger.FunctionName = functionName
			trigger.CreatedAt = t.CreatedAt
			trigger.UpdatedAt = time.Now()
			m.triggers[key][i] = trigger
			return nil
		}
	}

	return function.NewProviderError(function.ErrCodeNotFound, "trigger not found")
}

// DeleteTrigger deletes a trigger from a function
func (m *FunctionProvider) DeleteTrigger(ctx context.Context, functionName, triggerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Find function
	var key string
	for k, fn := range m.functions {
		if fn.Name == functionName {
			key = k
			break
		}
	}

	if key == "" {
		return function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	// Find and delete trigger
	triggers := m.triggers[key]
	for i, t := range triggers {
		if t.Name == triggerName {
			m.triggers[key] = append(triggers[:i], triggers[i+1:]...)
			return nil
		}
	}

	return function.NewProviderError(function.ErrCodeNotFound, "trigger not found")
}

// ListTriggers lists all triggers for a function
func (m *FunctionProvider) ListTriggers(ctx context.Context, functionName string) ([]*function.FunctionTrigger, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return nil, function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Find function
	var key string
	for k, fn := range m.functions {
		if fn.Name == functionName {
			key = k
			break
		}
	}

	if key == "" {
		return nil, function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	return m.triggers[key], nil
}

// InvokeFunction invokes a function synchronously
func (m *FunctionProvider) InvokeFunction(ctx context.Context, name string, req *function.InvokeRequest) (*function.InvokeResponse, error) {
	m.mu.Lock()
	m.invocationCount++
	invocationID := fmt.Sprintf("inv-%d-%d", m.invocationCount, time.Now().Unix())
	m.mu.Unlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return nil, function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Verify function exists
	m.mu.RLock()
	functionExists := false
	for _, fn := range m.functions {
		if fn.Name == name {
			functionExists = true
			break
		}
	}
	m.mu.RUnlock()

	if !functionExists {
		return nil, function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	// Simulate function execution
	executionTime := 50 * time.Millisecond
	if m.simulatedLatency > 0 {
		executionTime = m.simulatedLatency
	}
	time.Sleep(executionTime)

	// Create response
	response := &function.InvokeResponse{
		StatusCode: 200,
		Headers: map[string][]string{
			"Content-Type":    {"application/json"},
			"X-Function-Name": {name},
			"X-Invocation-Id": {invocationID},
		},
		Body:         []byte(`{"message":"Hello from mock function","invocationId":"` + invocationID + `"}`),
		Duration:     executionTime,
		ColdStart:    m.invocationCount <= 1,
		InvocationID: invocationID,
	}

	return response, nil
}

// InvokeFunctionAsync invokes a function asynchronously
func (m *FunctionProvider) InvokeFunctionAsync(ctx context.Context, name string, req *function.InvokeRequest) (string, error) {
	m.mu.Lock()
	m.invocationCount++
	invocationID := fmt.Sprintf("async-inv-%d-%d", m.invocationCount, time.Now().Unix())
	m.mu.Unlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return "", function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Verify function exists
	m.mu.RLock()
	functionExists := false
	for _, fn := range m.functions {
		if fn.Name == name {
			functionExists = true
			break
		}
	}
	m.mu.RUnlock()

	if !functionExists {
		return "", function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	// Create invocation status
	status := &function.InvocationStatus{
		InvocationID: invocationID,
		Status:       "running",
		StartedAt:    time.Now(),
	}

	m.mu.Lock()
	m.invocations[invocationID] = status
	m.mu.Unlock()

	// Simulate async execution
	go func() {
		time.Sleep(100 * time.Millisecond)
		
		m.mu.Lock()
		defer m.mu.Unlock()
		
		completedAt := time.Now()
		status.CompletedAt = &completedAt
		status.Status = "completed"
		status.Result = &function.InvokeResponse{
			StatusCode:   200,
			Headers:      map[string][]string{"Content-Type": {"application/json"}},
			Body:         []byte(`{"message":"Async execution completed"}`),
			Duration:     100 * time.Millisecond,
			InvocationID: invocationID,
		}
	}()

	return invocationID, nil
}

// GetInvocationStatus retrieves the status of an async invocation
func (m *FunctionProvider) GetInvocationStatus(ctx context.Context, invocationID string) (*function.InvocationStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return nil, function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	status, exists := m.invocations[invocationID]
	if !exists {
		return nil, function.NewProviderError(function.ErrCodeNotFound, "invocation not found")
	}

	return status, nil
}

// GetFunctionURL returns the URL for a function
func (m *FunctionProvider) GetFunctionURL(ctx context.Context, name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return "", function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Verify function exists
	functionExists := false
	for _, fn := range m.functions {
		if fn.Name == name {
			functionExists = true
			break
		}
	}

	if !functionExists {
		return "", function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	return fmt.Sprintf("http://mock.provider.local/functions/%s", name), nil
}

// GetFunctionLogs retrieves logs for a function
func (m *FunctionProvider) GetFunctionLogs(ctx context.Context, name string, opts *function.LogOptions) ([]*function.LogEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return nil, function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Verify function exists
	functionExists := false
	for _, fn := range m.functions {
		if fn.Name == name {
			functionExists = true
			break
		}
	}

	if !functionExists {
		return nil, function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	// Generate mock logs
	logs := []*function.LogEntry{
		{
			Timestamp: time.Now().Add(-5 * time.Minute),
			Level:     "info",
			Message:   fmt.Sprintf("Function %s initialized", name),
		},
		{
			Timestamp: time.Now().Add(-3 * time.Minute),
			Level:     "info",
			Message:   fmt.Sprintf("Function %s received request", name),
		},
		{
			Timestamp: time.Now().Add(-1 * time.Minute),
			Level:     "info",
			Message:   fmt.Sprintf("Function %s completed successfully", name),
		},
	}

	return logs, nil
}

// GetFunctionMetrics retrieves metrics for a function
func (m *FunctionProvider) GetFunctionMetrics(ctx context.Context, name string, opts *function.MetricOptions) (*function.Metrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.simulatedLatency > 0 {
		time.Sleep(m.simulatedLatency)
	}

	if m.failureMode {
		return nil, function.NewProviderError(function.ErrCodeInternal, "simulated failure")
	}

	// Verify function exists
	functionExists := false
	for _, fn := range m.functions {
		if fn.Name == name {
			functionExists = true
			break
		}
	}

	if !functionExists {
		return nil, function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	// Generate mock metrics
	metrics := &function.Metrics{
		Invocations: int64(m.invocationCount),
		Errors:      0,
		Duration: function.MetricStats{
			Min: 10,
			Max: 200,
			Avg: 50,
			P50: 45,
			P95: 150,
			P99: 190,
		},
		ColdStarts: 1,
		Concurrency: function.MetricStats{
			Min: 0,
			Max: 5,
			Avg: 2,
			P50: 2,
			P95: 4,
			P99: 5,
		},
	}

	return metrics, nil
}

// GetCapabilities returns the provider's capabilities
func (m *FunctionProvider) GetCapabilities() *function.Capabilities {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.capabilities
}

// HealthCheck performs a health check on the provider
func (m *FunctionProvider) HealthCheck(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.failureMode {
		return function.NewProviderError(function.ErrCodeInternal, "provider unhealthy")
	}

	return nil
}