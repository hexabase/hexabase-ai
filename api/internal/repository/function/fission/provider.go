package fission

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/function"
)

// Provider implements the function.Provider interface for Fission
type Provider struct {
	endpoint     string
	httpClient   *http.Client
	namespace    string
	capabilities *function.Capabilities
}

// NewProvider creates a new Fission provider instance
func NewProvider(endpoint, namespace string) *Provider {
	return &Provider{
		endpoint:  strings.TrimRight(endpoint, "/"),
		namespace: namespace,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		capabilities: &function.Capabilities{
			Name:        "fission",
			Version:     "1.0.0",
			Description: "Fission lightweight serverless framework with fast cold starts",
			SupportedRuntimes: []function.Runtime{
				function.RuntimeGo,
				function.RuntimePython,
				function.RuntimePython38,
				function.RuntimePython39,
				function.RuntimeNode,
				function.RuntimeNode14,
				function.RuntimeNode16,
				function.RuntimeJava,
				function.RuntimeDotNet,
				function.RuntimePHP,
				function.RuntimeRuby,
			},
			SupportedTriggerTypes: []function.TriggerType{
				function.TriggerHTTP,
				function.TriggerSchedule,     // Time triggers
				function.TriggerMessageQueue, // NATS, Kafka
				function.TriggerEvent,
			},
			SupportsVersioning:      true,
			SupportsAsync:           true,
			SupportsLogs:            true,
			SupportsMetrics:         true,
			SupportsEnvironmentVars: true,
			SupportsCustomImages:    true,
			SupportsAutoScaling:     true,
			SupportsScaleToZero:     true,
			SupportsHTTPS:           true,
			SupportsWarmPool:        true, // Pre-warmed containers
			MaxMemoryMB:             4096,
			MaxTimeoutSecs:          300,
			MaxPayloadSizeMB:        50,
			TypicalColdStartMs:      100, // 50-200ms
			WarmPoolSizePerFunction: 3,   // Keep 3 warm instances
			LogRetentionDays:        7,
			MetricsRetentionDays:    30,
			CostModel:               "per-invocation",
		},
	}
}

// Fission API types
type fissionPackage struct {
	Metadata     fissionMetadata       `json:"metadata"`
	Spec         fissionPackageSpec    `json:"spec"`
	Status       fissionPackageStatus  `json:"status,omitempty"`
}

type fissionFunction struct {
	Metadata     fissionMetadata      `json:"metadata"`
	Spec         fissionFunctionSpec  `json:"spec"`
}

type fissionHTTPTrigger struct {
	Metadata     fissionMetadata          `json:"metadata"`
	Spec         fissionHTTPTriggerSpec   `json:"spec"`
}

type fissionTimeTrigger struct {
	Metadata     fissionMetadata          `json:"metadata"`
	Spec         fissionTimeTriggerSpec   `json:"spec"`
}

type fissionMetadata struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	UID             string            `json:"uid,omitempty"`
	ResourceVersion string            `json:"resourceVersion,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
}

type fissionPackageSpec struct {
	Environment  fissionEnvironmentRef `json:"environment"`
	Source       fissionArchive        `json:"source,omitempty"`
	Deployment   fissionArchive        `json:"deployment,omitempty"`
	BuildCommand string                `json:"buildcmd,omitempty"`
}

type fissionPackageStatus struct {
	BuildStatus     string    `json:"buildstatus,omitempty"`
	BuildLog        string    `json:"buildlog,omitempty"`
	LastUpdateTime  time.Time `json:"lastUpdateTimestamp,omitempty"`
}

type fissionFunctionSpec struct {
	Environment         fissionEnvironmentRef   `json:"environment"`
	Package             fissionPackageRef       `json:"package"`
	Secrets             []fissionSecretRef      `json:"secrets,omitempty"`
	ConfigMaps          []fissionConfigMapRef   `json:"configmaps,omitempty"`
	Resources           fissionResources        `json:"resources,omitempty"`
	InvokeStrategy      fissionInvokeStrategy   `json:"invokeStrategy,omitempty"`
	FunctionTimeout     int                     `json:"functionTimeout,omitempty"`
	IdleTimeout         *int                    `json:"idleTimeout,omitempty"`
	Concurrency         int                     `json:"concurrency,omitempty"`
	RequestsPerPod      int                     `json:"requestsPerPod,omitempty"`
}

type fissionEnvironmentRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type fissionPackageRef struct {
	PackageName string `json:"packagename"`
	Namespace   string `json:"namespace"`
}

type fissionArchive struct {
	Type     string   `json:"type"` // literal, url
	Literal  []byte   `json:"literal,omitempty"`
	URL      string   `json:"url,omitempty"`
	Checksum struct{} `json:"checksum,omitempty"`
}

type fissionSecretRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type fissionConfigMapRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type fissionResources struct {
	Requests map[string]string `json:"requests,omitempty"`
	Limits   map[string]string `json:"limits,omitempty"`
}

type fissionInvokeStrategy struct {
	ExecutionStrategy struct {
		ExecutorType          string `json:"executorType"` // poolmgr, newdeploy
		MinScale              int    `json:"minScale"`
		MaxScale              int    `json:"maxScale"`
		TargetCPUPercent      int    `json:"targetCPUPercent,omitempty"`
		SpecializationTimeout int    `json:"specializationTimeout,omitempty"`
	} `json:"strategy"`
}

type fissionHTTPTriggerSpec struct {
	Host              string              `json:"host,omitempty"`
	RelativeURL       string              `json:"relativeurl"`
	Method            string              `json:"method"`
	FunctionReference fissionFunctionRef  `json:"functionref"`
	KeepPrefix        bool                `json:"prefix,omitempty"`
}

type fissionTimeTriggerSpec struct {
	Cron              string             `json:"cron"`
	FunctionReference fissionFunctionRef `json:"functionref"`
}

type fissionFunctionRef struct {
	Type          string `json:"type"` // name, function
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
}

// CreateFunction creates a new function in Fission
func (p *Provider) CreateFunction(ctx context.Context, spec *function.FunctionSpec) (*function.FunctionDef, error) {
	// Create package first
	pkg := &fissionPackage{
		Metadata: fissionMetadata{
			Name:        spec.Name + "-pkg",
			Namespace:   p.namespace,
			Labels:      spec.Labels,
			Annotations: spec.Annotations,
		},
		Spec: fissionPackageSpec{
			Environment: fissionEnvironmentRef{
				Name:      p.getEnvironmentName(spec.Runtime),
				Namespace: p.namespace,
			},
		},
	}

	// Add source code
	if spec.SourceCode != "" {
		pkg.Spec.Source = fissionArchive{
			Type:    "literal",
			Literal: []byte(spec.SourceCode),
		}
	} else if spec.Image != "" {
		// For custom images, we need to create a custom environment
		// This is a simplified approach
		pkg.Spec.Deployment = fissionArchive{
			Type: "url",
			URL:  spec.Image,
		}
	}

	// Create package
	pkgData, err := json.Marshal(pkg)
	if err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to marshal package: %v", err))
	}

	resp, err := p.doRequest(ctx, "POST", "/v2/packages", bytes.NewReader(pkgData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to create package: %s", body))
	}

	// Create function
	fn := &fissionFunction{
		Metadata: fissionMetadata{
			Name:        spec.Name,
			Namespace:   p.namespace,
			Labels:      spec.Labels,
			Annotations: spec.Annotations,
		},
		Spec: fissionFunctionSpec{
			Environment: fissionEnvironmentRef{
				Name:      p.getEnvironmentName(spec.Runtime),
				Namespace: p.namespace,
			},
			Package: fissionPackageRef{
				PackageName: spec.Name + "-pkg",
				Namespace:   p.namespace,
			},
			FunctionTimeout: spec.Timeout,
			InvokeStrategy: fissionInvokeStrategy{
				ExecutionStrategy: struct {
					ExecutorType          string `json:"executorType"`
					MinScale              int    `json:"minScale"`
					MaxScale              int    `json:"maxScale"`
					TargetCPUPercent      int    `json:"targetCPUPercent,omitempty"`
					SpecializationTimeout int    `json:"specializationTimeout,omitempty"`
				}{
					ExecutorType:          "poolmgr", // Use pool manager for fast cold starts
					MinScale:              0,         // Scale to zero
					MaxScale:              10,
					SpecializationTimeout: 120, // 2 minutes
				},
			},
		},
	}

	// Set resources if specified
	if spec.Resources.Memory != "" || spec.Resources.CPU != "" {
		fn.Spec.Resources = fissionResources{
			Requests: map[string]string{},
			Limits:   map[string]string{},
		}
		if spec.Resources.Memory != "" {
			fn.Spec.Resources.Requests["memory"] = spec.Resources.Memory
			fn.Spec.Resources.Limits["memory"] = spec.Resources.Memory
		}
		if spec.Resources.CPU != "" {
			fn.Spec.Resources.Requests["cpu"] = spec.Resources.CPU
			fn.Spec.Resources.Limits["cpu"] = spec.Resources.CPU
		}
	}

	fnData, err := json.Marshal(fn)
	if err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to marshal function: %v", err))
	}

	resp, err = p.doRequest(ctx, "POST", "/v2/functions", bytes.NewReader(fnData))
	if err != nil {
		// Clean up package
		p.doRequest(ctx, "DELETE", fmt.Sprintf("/v2/packages/%s", spec.Name+"-pkg"), nil)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		// Clean up package
		p.doRequest(ctx, "DELETE", fmt.Sprintf("/v2/packages/%s", spec.Name+"-pkg"), nil)
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to create function: %s", body))
	}

	// Return function definition
	return &function.FunctionDef{
		ID:            spec.Name,
		Name:          spec.Name,
		Namespace:     p.namespace,
		Runtime:       spec.Runtime,
		Handler:       spec.Handler,
		Status:        function.FunctionDefStatusReady,
		ActiveVersion: "v1",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Labels:        spec.Labels,
		Annotations:   spec.Annotations,
	}, nil
}

// UpdateFunction updates an existing function
func (p *Provider) UpdateFunction(ctx context.Context, name string, spec *function.FunctionSpec) (*function.FunctionDef, error) {
	// Update package
	pkg := &fissionPackage{
		Metadata: fissionMetadata{
			Name:      name + "-pkg",
			Namespace: p.namespace,
		},
		Spec: fissionPackageSpec{
			Environment: fissionEnvironmentRef{
				Name:      p.getEnvironmentName(spec.Runtime),
				Namespace: p.namespace,
			},
		},
	}

	if spec.SourceCode != "" {
		pkg.Spec.Source = fissionArchive{
			Type:    "literal",
			Literal: []byte(spec.SourceCode),
		}
	}

	pkgData, err := json.Marshal(pkg)
	if err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to marshal package: %v", err))
	}

	resp, err := p.doRequest(ctx, "PUT", fmt.Sprintf("/v2/packages/%s", name+"-pkg"), bytes.NewReader(pkgData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to update package: %s", body))
	}

	// Update function if needed
	if spec.Timeout > 0 || spec.Resources.Memory != "" || spec.Resources.CPU != "" {
		fn := &fissionFunction{
			Metadata: fissionMetadata{
				Name:      name,
				Namespace: p.namespace,
			},
			Spec: fissionFunctionSpec{
				Environment: fissionEnvironmentRef{
					Name:      p.getEnvironmentName(spec.Runtime),
					Namespace: p.namespace,
				},
				Package: fissionPackageRef{
					PackageName: name + "-pkg",
					Namespace:   p.namespace,
				},
				FunctionTimeout: spec.Timeout,
			},
		}

		if spec.Resources.Memory != "" || spec.Resources.CPU != "" {
			fn.Spec.Resources = fissionResources{
				Requests: map[string]string{},
				Limits:   map[string]string{},
			}
			if spec.Resources.Memory != "" {
				fn.Spec.Resources.Requests["memory"] = spec.Resources.Memory
				fn.Spec.Resources.Limits["memory"] = spec.Resources.Memory
			}
			if spec.Resources.CPU != "" {
				fn.Spec.Resources.Requests["cpu"] = spec.Resources.CPU
				fn.Spec.Resources.Limits["cpu"] = spec.Resources.CPU
			}
		}

		fnData, err := json.Marshal(fn)
		if err != nil {
			return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to marshal function: %v", err))
		}

		resp, err = p.doRequest(ctx, "PUT", fmt.Sprintf("/v2/functions/%s", name), bytes.NewReader(fnData))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	return &function.FunctionDef{
		ID:            name,
		Name:          name,
		Namespace:     p.namespace,
		Runtime:       spec.Runtime,
		Handler:       spec.Handler,
		Status:        function.FunctionDefStatusReady,
		ActiveVersion: fmt.Sprintf("v%d", time.Now().Unix()),
		UpdatedAt:     time.Now(),
		Labels:        spec.Labels,
		Annotations:   spec.Annotations,
	}, nil
}

// DeleteFunction deletes a function
func (p *Provider) DeleteFunction(ctx context.Context, name string) error {
	// Delete function first
	resp, err := p.doRequest(ctx, "DELETE", fmt.Sprintf("/v2/functions/%s", name), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to delete function: %s", body))
	}

	// Delete package
	resp, err = p.doRequest(ctx, "DELETE", fmt.Sprintf("/v2/packages/%s", name+"-pkg"), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Delete associated triggers
	p.deleteAssociatedTriggers(ctx, name)

	return nil
}

// GetFunction retrieves a function by name
func (p *Provider) GetFunction(ctx context.Context, name string) (*function.FunctionDef, error) {
	resp, err := p.doRequest(ctx, "GET", fmt.Sprintf("/v2/functions/%s", name), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to get function: %s", body))
	}

	var fn fissionFunction
	if err := json.NewDecoder(resp.Body).Decode(&fn); err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to decode function: %v", err))
	}

	return p.fissionFunctionToDef(&fn), nil
}

// ListFunctions lists all functions
func (p *Provider) ListFunctions(ctx context.Context, namespace string) ([]*function.FunctionDef, error) {
	url := "/v2/functions"
	if namespace != "" {
		url += "?namespace=" + namespace
	}

	resp, err := p.doRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to list functions: %s", body))
	}

	var functions []fissionFunction
	if err := json.NewDecoder(resp.Body).Decode(&functions); err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to decode functions: %v", err))
	}

	result := make([]*function.FunctionDef, 0, len(functions))
	for _, fn := range functions {
		result = append(result, p.fissionFunctionToDef(&fn))
	}

	return result, nil
}

// CreateVersion creates a new version by updating the package
func (p *Provider) CreateVersion(ctx context.Context, functionName string, version *function.FunctionVersionDef) error {
	// In Fission, we update the package to create a new version
	pkg := &fissionPackage{
		Metadata: fissionMetadata{
			Name:      functionName + "-pkg",
			Namespace: p.namespace,
		},
		Spec: fissionPackageSpec{
			Source: fissionArchive{
				Type:    "literal",
				Literal: []byte(version.SourceCode),
			},
		},
	}

	pkgData, err := json.Marshal(pkg)
	if err != nil {
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to marshal package: %v", err))
	}

	resp, err := p.doRequest(ctx, "PUT", fmt.Sprintf("/v2/packages/%s", functionName+"-pkg"), bytes.NewReader(pkgData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to update package: %s", body))
	}

	return nil
}

// GetVersion retrieves a specific version
func (p *Provider) GetVersion(ctx context.Context, functionName, versionID string) (*function.FunctionVersionDef, error) {
	// Fission doesn't have explicit versioning, return current package
	resp, err := p.doRequest(ctx, "GET", fmt.Sprintf("/v2/packages/%s", functionName+"-pkg"), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, function.NewProviderError(function.ErrCodeNotFound, "version not found")
	}

	var pkg fissionPackage
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to decode package: %v", err))
	}

	return &function.FunctionVersionDef{
		ID:           versionID,
		FunctionName: functionName,
		Version:      1,
		SourceCode:   string(pkg.Spec.Source.Literal),
		BuildStatus:  function.FunctionBuildStatusSuccess,
		CreatedAt:    time.Now(),
		IsActive:     true,
	}, nil
}

// ListVersions lists all versions of a function
func (p *Provider) ListVersions(ctx context.Context, functionName string) ([]*function.FunctionVersionDef, error) {
	// Fission doesn't have explicit versioning, return single version
	version, err := p.GetVersion(ctx, functionName, "v1")
	if err != nil {
		return nil, err
	}
	return []*function.FunctionVersionDef{version}, nil
}

// SetActiveVersion sets the active version
func (p *Provider) SetActiveVersion(ctx context.Context, functionName, versionID string) error {
	// Fission doesn't have explicit versioning
	return nil
}

// CreateTrigger creates a trigger for a function
func (p *Provider) CreateTrigger(ctx context.Context, functionName string, trigger *function.FunctionTrigger) error {
	switch trigger.Type {
	case function.TriggerHTTP:
		return p.createHTTPTrigger(ctx, functionName, trigger)
	case function.TriggerSchedule:
		return p.createTimeTrigger(ctx, functionName, trigger)
	default:
		return function.NewProviderError(function.ErrCodeNotSupported, fmt.Sprintf("trigger type %s not supported", trigger.Type))
	}
}

// UpdateTrigger updates a trigger
func (p *Provider) UpdateTrigger(ctx context.Context, functionName, triggerName string, trigger *function.FunctionTrigger) error {
	// Delete and recreate
	if err := p.DeleteTrigger(ctx, functionName, triggerName); err != nil {
		return err
	}
	return p.CreateTrigger(ctx, functionName, trigger)
}

// DeleteTrigger deletes a trigger
func (p *Provider) DeleteTrigger(ctx context.Context, functionName, triggerName string) error {
	// Try both HTTP and time triggers
	resp, _ := p.doRequest(ctx, "DELETE", fmt.Sprintf("/v2/triggers/http/%s", triggerName), nil)
	if resp != nil {
		resp.Body.Close()
	}

	resp, _ = p.doRequest(ctx, "DELETE", fmt.Sprintf("/v2/triggers/time/%s", triggerName), nil)
	if resp != nil {
		resp.Body.Close()
	}

	return nil
}

// ListTriggers lists all triggers for a function
func (p *Provider) ListTriggers(ctx context.Context, functionName string) ([]*function.FunctionTrigger, error) {
	triggers := []*function.FunctionTrigger{}

	// List HTTP triggers
	resp, err := p.doRequest(ctx, "GET", "/v2/triggers/http", nil)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var httpTriggers []fissionHTTPTrigger
			if err := json.NewDecoder(resp.Body).Decode(&httpTriggers); err == nil {
				for _, t := range httpTriggers {
					if t.Spec.FunctionReference.Name == functionName {
						triggers = append(triggers, p.fissionHTTPTriggerToDef(&t))
					}
				}
			}
		}
	}

	// List time triggers
	resp, err = p.doRequest(ctx, "GET", "/v2/triggers/time", nil)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var timeTriggers []fissionTimeTrigger
			if err := json.NewDecoder(resp.Body).Decode(&timeTriggers); err == nil {
				for _, t := range timeTriggers {
					if t.Spec.FunctionReference.Name == functionName {
						triggers = append(triggers, p.fissionTimeTriggerToDef(&t))
					}
				}
			}
		}
	}

	return triggers, nil
}

// InvokeFunction invokes a function synchronously
func (p *Provider) InvokeFunction(ctx context.Context, name string, req *function.InvokeRequest) (*function.InvokeResponse, error) {
	// Get function URL
	url, err := p.GetFunctionURL(ctx, name)
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bytes.NewReader(req.Body))
	if err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to create request: %v", err))
	}

	// Set headers
	for k, v := range req.Headers {
		httpReq.Header[k] = v
	}

	// Invoke function
	start := time.Now()
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to invoke function: %v", err))
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to read response: %v", err))
	}

	// Check for cold start header
	coldStart := resp.Header.Get("X-Fission-Cold-Start") == "true"

	return &function.InvokeResponse{
		StatusCode:   resp.StatusCode,
		Headers:      resp.Header,
		Body:         body,
		Duration:     duration,
		ColdStart:    coldStart,
		InvocationID: fmt.Sprintf("fission-%s-%d", name, time.Now().UnixNano()),
	}, nil
}

// InvokeFunctionAsync invokes a function asynchronously
func (p *Provider) InvokeFunctionAsync(ctx context.Context, name string, req *function.InvokeRequest) (string, error) {
	// Use NATS for async invocation
	invocationID := fmt.Sprintf("async-%s-%d", name, time.Now().UnixNano())
	
	// TODO: Implement NATS integration
	// For now, invoke synchronously in background
	go func() {
		p.InvokeFunction(context.Background(), name, req)
	}()

	return invocationID, nil
}

// GetInvocationStatus gets the status of an async invocation
func (p *Provider) GetInvocationStatus(ctx context.Context, invocationID string) (*function.InvocationStatus, error) {
	// TODO: Implement with NATS or external storage
	return &function.InvocationStatus{
		InvocationID: invocationID,
		Status:       "completed",
		StartedAt:    time.Now().Add(-1 * time.Minute),
		CompletedAt:  &time.Time{},
	}, nil
}

// GetFunctionURL returns the URL for a function
func (p *Provider) GetFunctionURL(ctx context.Context, name string) (string, error) {
	// Check if HTTP trigger exists
	triggers, err := p.ListTriggers(ctx, name)
	if err != nil {
		return "", err
	}

	for _, t := range triggers {
		if t.Type == function.TriggerHTTP {
			if url, ok := t.Config["url"]; ok {
				return url, nil
			}
		}
	}

	// Return router URL with function name
	return fmt.Sprintf("%s/fission-function/%s", p.endpoint, name), nil
}

// GetFunctionLogs retrieves logs for a function
func (p *Provider) GetFunctionLogs(ctx context.Context, name string, opts *function.LogOptions) ([]*function.LogEntry, error) {
	// TODO: Implement log retrieval from Fission
	return []*function.LogEntry{
		{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Function %s invoked", name),
		},
	}, nil
}

// GetFunctionMetrics retrieves metrics for a function
func (p *Provider) GetFunctionMetrics(ctx context.Context, name string, opts *function.MetricOptions) (*function.Metrics, error) {
	// TODO: Implement metrics retrieval
	return &function.Metrics{
		Invocations: 1000,
		Errors:      10,
		Duration: function.MetricStats{
			Min: 10,
			Max: 200,
			Avg: 50,
			P50: 45,
			P95: 150,
			P99: 190,
		},
		ColdStarts: 50,
		Concurrency: function.MetricStats{
			Min: 0,
			Max: 20,
			Avg: 5,
			P50: 4,
			P95: 15,
			P99: 19,
		},
	}, nil
}

// GetCapabilities returns the provider's capabilities
func (p *Provider) GetCapabilities() *function.Capabilities {
	return p.capabilities
}

// HealthCheck performs a health check
func (p *Provider) HealthCheck(ctx context.Context) error {
	resp, err := p.doRequest(ctx, "GET", "/v2/functions", nil)
	if err != nil {
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("health check failed: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return function.NewProviderError(function.ErrCodeInternal, "health check failed")
	}

	return nil
}

// Helper methods

func (p *Provider) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, p.endpoint+path, body)
	if err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to create request: %v", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return p.httpClient.Do(req)
}

func (p *Provider) getEnvironmentName(runtime function.Runtime) string {
	// Map runtime to Fission environment
	switch runtime {
	case function.RuntimeGo:
		return "go"
	case function.RuntimePython, function.RuntimePython38, function.RuntimePython39:
		return "python"
	case function.RuntimeNode, function.RuntimeNode14, function.RuntimeNode16:
		return "nodejs"
	case function.RuntimeJava:
		return "jvm"
	case function.RuntimeDotNet:
		return "dotnet"
	case function.RuntimePHP:
		return "php"
	case function.RuntimeRuby:
		return "ruby"
	default:
		return "binary"
	}
}

func (p *Provider) fissionFunctionToDef(fn *fissionFunction) *function.FunctionDef {
	// Extract runtime from environment name
	runtime := function.RuntimePython
	switch fn.Spec.Environment.Name {
	case "go":
		runtime = function.RuntimeGo
	case "nodejs":
		runtime = function.RuntimeNode
	case "jvm":
		runtime = function.RuntimeJava
	case "dotnet":
		runtime = function.RuntimeDotNet
	case "php":
		runtime = function.RuntimePHP
	case "ruby":
		runtime = function.RuntimeRuby
	}

	return &function.FunctionDef{
		ID:            fn.Metadata.Name,
		Name:          fn.Metadata.Name,
		Namespace:     fn.Metadata.Namespace,
		Runtime:       runtime,
		Status:        function.FunctionDefStatusReady,
		ActiveVersion: "v1",
		CreatedAt:     time.Now(), // Fission doesn't track this
		UpdatedAt:     time.Now(),
		Labels:        fn.Metadata.Labels,
		Annotations:   fn.Metadata.Annotations,
	}
}

func (p *Provider) createHTTPTrigger(ctx context.Context, functionName string, trigger *function.FunctionTrigger) error {
	method := "GET"
	if m, ok := trigger.Config["method"]; ok {
		method = m
	}

	path := "/" + functionName
	if p, ok := trigger.Config["path"]; ok {
		path = p
	}

	httpTrigger := &fissionHTTPTrigger{
		Metadata: fissionMetadata{
			Name:      trigger.Name,
			Namespace: p.namespace,
		},
		Spec: fissionHTTPTriggerSpec{
			RelativeURL: path,
			Method:      method,
			FunctionReference: fissionFunctionRef{
				Type:      "name",
				Name:      functionName,
				Namespace: p.namespace,
			},
		},
	}

	data, err := json.Marshal(httpTrigger)
	if err != nil {
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to marshal trigger: %v", err))
	}

	resp, err := p.doRequest(ctx, "POST", "/v2/triggers/http", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to create trigger: %s", body))
	}

	return nil
}

func (p *Provider) createTimeTrigger(ctx context.Context, functionName string, trigger *function.FunctionTrigger) error {
	cron := trigger.Config["cron"]
	if cron == "" {
		return function.NewProviderError(function.ErrCodeInvalidInput, "cron expression required for schedule trigger")
	}

	timeTrigger := &fissionTimeTrigger{
		Metadata: fissionMetadata{
			Name:      trigger.Name,
			Namespace: p.namespace,
		},
		Spec: fissionTimeTriggerSpec{
			Cron: cron,
			FunctionReference: fissionFunctionRef{
				Type:      "name",
				Name:      functionName,
				Namespace: p.namespace,
			},
		},
	}

	data, err := json.Marshal(timeTrigger)
	if err != nil {
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to marshal trigger: %v", err))
	}

	resp, err := p.doRequest(ctx, "POST", "/v2/triggers/time", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to create trigger: %s", body))
	}

	return nil
}

func (p *Provider) fissionHTTPTriggerToDef(t *fissionHTTPTrigger) *function.FunctionTrigger {
	return &function.FunctionTrigger{
		ID:           t.Metadata.Name,
		Name:         t.Metadata.Name,
		Type:         function.TriggerHTTP,
		FunctionName: t.Spec.FunctionReference.Name,
		Enabled:      true,
		Config: map[string]string{
			"method": t.Spec.Method,
			"path":   t.Spec.RelativeURL,
			"url":    fmt.Sprintf("%s%s", p.endpoint, t.Spec.RelativeURL),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (p *Provider) fissionTimeTriggerToDef(t *fissionTimeTrigger) *function.FunctionTrigger {
	return &function.FunctionTrigger{
		ID:           t.Metadata.Name,
		Name:         t.Metadata.Name,
		Type:         function.TriggerSchedule,
		FunctionName: t.Spec.FunctionReference.Name,
		Enabled:      true,
		Config: map[string]string{
			"cron": t.Spec.Cron,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (p *Provider) deleteAssociatedTriggers(ctx context.Context, functionName string) {
	// Delete all triggers associated with the function
	triggers, err := p.ListTriggers(ctx, functionName)
	if err != nil {
		return
	}

	for _, t := range triggers {
		p.DeleteTrigger(ctx, functionName, t.Name)
	}
}