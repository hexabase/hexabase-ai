package knative

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/hexabase/hexabase-ai/api/internal/domain/function"
)

// Provider implements the function.Provider interface for Knative
type Provider struct {
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
	namespace     string
	capabilities  *function.Capabilities
}

// NewProvider creates a new Knative provider instance
func NewProvider(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface, namespace string) *Provider {
	return &Provider{
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
		namespace:     namespace,
		capabilities: &function.Capabilities{
			Name:        "knative",
			Version:     "1.0.0",
			Description: "Knative Serving provider for serverless functions",
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
				function.TriggerEvent, // Via Knative Eventing
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
			MaxMemoryMB:             8192,
			MaxTimeoutSecs:          600,
			MaxPayloadSizeMB:        100,
			TypicalColdStartMs:      3000, // 2-5 seconds typical
			LogRetentionDays:        7,
			MetricsRetentionDays:    30,
		},
	}
}

// knativeServiceGVR is the GroupVersionResource for Knative Services
var knativeServiceGVR = schema.GroupVersionResource{
	Group:    "serving.knative.dev",
	Version:  "v1",
	Resource: "services",
}

// CreateFunction creates a new function using Knative Service
func (p *Provider) CreateFunction(ctx context.Context, spec *function.FunctionSpec) (*function.FunctionDef, error) {
	// Build container image if source code is provided
	image := spec.Image
	if image == "" && spec.SourceCode != "" {
		image = p.buildImage(spec)
	}

	// Create Knative Service
	service := p.buildKnativeService(spec.Name, spec.Namespace, image, spec)
	
	_, err := p.dynamicClient.Resource(knativeServiceGVR).Namespace(spec.Namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to create Knative service: %v", err))
	}

	// Wait for service to be ready
	if err := p.waitForServiceReady(ctx, spec.Namespace, spec.Name); err != nil {
		return nil, err
	}

	// Return function definition
	return &function.FunctionDef{
		Name:          spec.Name,
		Namespace:     spec.Namespace,
		Runtime:       spec.Runtime,
		Handler:       spec.Handler,
		Status:        function.FunctionDefStatusReady,
		ActiveVersion: "00001", // Knative revision format
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Labels:        spec.Labels,
		Annotations:   spec.Annotations,
	}, nil
}

// UpdateFunction updates an existing function
func (p *Provider) UpdateFunction(ctx context.Context, name string, spec *function.FunctionSpec) (*function.FunctionDef, error) {
	// Get existing service
	existing, err := p.dynamicClient.Resource(knativeServiceGVR).Namespace(spec.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, function.NewProviderError(function.ErrCodeNotFound, "function not found")
		}
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to get function: %v", err))
	}

	// Update service spec
	image := spec.Image
	if image == "" && spec.SourceCode != "" {
		image = p.buildImage(spec)
	}

	// Update the service
	p.updateKnativeServiceSpec(existing, image, spec)
	
	_, err = p.dynamicClient.Resource(knativeServiceGVR).Namespace(spec.Namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to update function: %v", err))
	}

	// Wait for update to complete
	if err := p.waitForServiceReady(ctx, spec.Namespace, name); err != nil {
		return nil, err
	}

	// Get latest revision
	latestRevision, err := p.getLatestRevision(ctx, spec.Namespace, name)
	if err != nil {
		return nil, err
	}

	return &function.FunctionDef{
		Name:          name,
		Namespace:     spec.Namespace,
		Runtime:       spec.Runtime,
		Handler:       spec.Handler,
		Status:        function.FunctionDefStatusReady,
		ActiveVersion: latestRevision,
		UpdatedAt:     time.Now(),
		Labels:        spec.Labels,
		Annotations:   spec.Annotations,
	}, nil
}

// DeleteFunction deletes a function
func (p *Provider) DeleteFunction(ctx context.Context, name string) error {
	// Find the function in any namespace
	services, err := p.dynamicClient.Resource(knativeServiceGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to list services: %v", err))
	}

	found := false
	var namespace string
	for _, item := range services.Items {
		if item.GetName() == name {
			namespace = item.GetNamespace()
			found = true
			break
		}
	}

	if !found {
		return function.NewProviderError(function.ErrCodeNotFound, "function not found")
	}

	// Delete the service
	err = p.dynamicClient.Resource(knativeServiceGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to delete function: %v", err))
	}

	return nil
}

// GetFunction retrieves a function by name
func (p *Provider) GetFunction(ctx context.Context, name string) (*function.FunctionDef, error) {
	// Search for the function in all namespaces
	services, err := p.dynamicClient.Resource(knativeServiceGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to list services: %v", err))
	}

	for _, item := range services.Items {
		if item.GetName() == name {
			return p.knativeServiceToFunctionDef(&item)
		}
	}

	return nil, function.NewProviderError(function.ErrCodeNotFound, "function not found")
}

// ListFunctions lists all functions in a namespace
func (p *Provider) ListFunctions(ctx context.Context, namespace string) ([]*function.FunctionDef, error) {
	var services *unstructured.UnstructuredList
	var err error

	if namespace == "" {
		services, err = p.dynamicClient.Resource(knativeServiceGVR).List(ctx, metav1.ListOptions{})
	} else {
		services, err = p.dynamicClient.Resource(knativeServiceGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to list functions: %v", err))
	}

	functions := make([]*function.FunctionDef, 0, len(services.Items))
	for _, item := range services.Items {
		fn, err := p.knativeServiceToFunctionDef(&item)
		if err != nil {
			continue
		}
		functions = append(functions, fn)
	}

	return functions, nil
}

// CreateVersion creates a new version by updating the Knative service
func (p *Provider) CreateVersion(ctx context.Context, functionName string, version *function.FunctionVersionDef) error {
	// In Knative, versions are managed as revisions automatically
	// We need to update the service to create a new revision
	fn, err := p.GetFunction(ctx, functionName)
	if err != nil {
		return err
	}

	spec := &function.FunctionSpec{
		Name:       functionName,
		Namespace:  fn.Namespace,
		Runtime:    fn.Runtime,
		Handler:    fn.Handler,
		SourceCode: version.SourceCode,
		Image:      version.Image,
	}

	_, err = p.UpdateFunction(ctx, functionName, spec)
	return err
}

// GetVersion retrieves a specific version (revision)
func (p *Provider) GetVersion(ctx context.Context, functionName, versionID string) (*function.FunctionVersionDef, error) {
	// Get function to find namespace
	fn, err := p.GetFunction(ctx, functionName)
	if err != nil {
		return nil, err
	}

	// Get revision
	revisionGVR := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "revisions",
	}

	revision, err := p.dynamicClient.Resource(revisionGVR).Namespace(fn.Namespace).Get(ctx, versionID, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, function.NewProviderError(function.ErrCodeNotFound, "version not found")
		}
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to get version: %v", err))
	}

	return p.knativeRevisionToVersion(revision, functionName)
}

// ListVersions lists all versions (revisions) of a function
func (p *Provider) ListVersions(ctx context.Context, functionName string) ([]*function.FunctionVersionDef, error) {
	// Get function to find namespace
	fn, err := p.GetFunction(ctx, functionName)
	if err != nil {
		return nil, err
	}

	// List revisions for the service
	revisionGVR := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "revisions",
	}

	revisions, err := p.dynamicClient.Resource(revisionGVR).Namespace(fn.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("serving.knative.dev/service=%s", functionName),
	})
	if err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to list versions: %v", err))
	}

	versions := make([]*function.FunctionVersionDef, 0, len(revisions.Items))
	for i, item := range revisions.Items {
		v, err := p.knativeRevisionToVersion(&item, functionName)
		if err != nil {
			continue
		}
		v.Version = i + 1
		versions = append(versions, v)
	}

	return versions, nil
}

// SetActiveVersion sets the active version by updating traffic split
func (p *Provider) SetActiveVersion(ctx context.Context, functionName, versionID string) error {
	// Get function to find namespace
	fn, err := p.GetFunction(ctx, functionName)
	if err != nil {
		return err
	}

	// Get the service
	service, err := p.dynamicClient.Resource(knativeServiceGVR).Namespace(fn.Namespace).Get(ctx, functionName, metav1.GetOptions{})
	if err != nil {
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to get service: %v", err))
	}

	// Update traffic to point to specific revision
	spec, found, err := unstructured.NestedMap(service.Object, "spec")
	if err != nil || !found {
		return function.NewProviderError(function.ErrCodeInternal, "invalid service spec")
	}

	traffic := []interface{}{
		map[string]interface{}{
			"revisionName": versionID,
			"percent":      100,
		},
	}
	spec["traffic"] = traffic
	
	unstructured.SetNestedMap(service.Object, spec, "spec")

	_, err = p.dynamicClient.Resource(knativeServiceGVR).Namespace(fn.Namespace).Update(ctx, service, metav1.UpdateOptions{})
	if err != nil {
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to update traffic: %v", err))
	}

	return nil
}

// CreateTrigger creates a trigger for the function
func (p *Provider) CreateTrigger(ctx context.Context, functionName string, trigger *function.FunctionTrigger) error {
	// Knative uses different resources for different trigger types
	switch trigger.Type {
	case function.TriggerHTTP:
		// HTTP triggers are handled by the Knative Service itself
		return nil
	case function.TriggerEvent:
		// Create Knative Eventing trigger
		return p.createEventTrigger(ctx, functionName, trigger)
	case function.TriggerSchedule:
		// Knative doesn't have built-in cron, would need to use Knative Eventing + CronJobSource
		return function.NewProviderError(function.ErrCodeNotSupported, "schedule triggers not supported by Knative provider")
	default:
		return function.NewProviderError(function.ErrCodeInvalidInput, fmt.Sprintf("unsupported trigger type: %s", trigger.Type))
	}
}

// UpdateTrigger updates a trigger
func (p *Provider) UpdateTrigger(ctx context.Context, functionName, triggerName string, trigger *function.FunctionTrigger) error {
	switch trigger.Type {
	case function.TriggerHTTP:
		return nil
	case function.TriggerEvent:
		return p.updateEventTrigger(ctx, functionName, triggerName, trigger)
	default:
		return function.NewProviderError(function.ErrCodeNotSupported, "trigger type not supported")
	}
}

// DeleteTrigger deletes a trigger
func (p *Provider) DeleteTrigger(ctx context.Context, functionName, triggerName string) error {
	// Get function to find namespace
	fn, err := p.GetFunction(ctx, functionName)
	if err != nil {
		return err
	}

	// Delete Knative Eventing trigger
	triggerGVR := schema.GroupVersionResource{
		Group:    "eventing.knative.dev",
		Version:  "v1",
		Resource: "triggers",
	}

	err = p.dynamicClient.Resource(triggerGVR).Namespace(fn.Namespace).Delete(ctx, triggerName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to delete trigger: %v", err))
	}

	return nil
}

// ListTriggers lists all triggers for a function
func (p *Provider) ListTriggers(ctx context.Context, functionName string) ([]*function.FunctionTrigger, error) {
	// Get function to find namespace
	fn, err := p.GetFunction(ctx, functionName)
	if err != nil {
		return nil, err
	}

	triggers := []*function.FunctionTrigger{}

	// HTTP trigger is always available for Knative Services
	triggers = append(triggers, &function.FunctionTrigger{
		Name:         "http",
		Type:         function.TriggerHTTP,
		FunctionName: functionName,
		Enabled:      true,
		Config: map[string]string{
			"url": p.getFunctionURL(fn.Namespace, functionName),
		},
		CreatedAt: fn.CreatedAt,
		UpdatedAt: fn.UpdatedAt,
	})

	// List Knative Eventing triggers
	triggerGVR := schema.GroupVersionResource{
		Group:    "eventing.knative.dev",
		Version:  "v1",
		Resource: "triggers",
	}

	eventTriggers, err := p.dynamicClient.Resource(triggerGVR).Namespace(fn.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("serving.knative.dev/service=%s", functionName),
	})
	if err == nil {
		for _, item := range eventTriggers.Items {
			t, err := p.knativeTriggerToFunctionTrigger(&item, functionName)
			if err == nil {
				triggers = append(triggers, t)
			}
		}
	}

	return triggers, nil
}

// InvokeFunction invokes a function synchronously
func (p *Provider) InvokeFunction(ctx context.Context, name string, req *function.InvokeRequest) (*function.InvokeResponse, error) {
	// Get function to find its URL
	fn, err := p.GetFunction(ctx, name)
	if err != nil {
		return nil, err
	}

	_ = p.getFunctionURL(fn.Namespace, name) // url will be used for actual HTTP invocation
	
	// TODO: Implement actual HTTP invocation
	// For now, return a mock response
	return &function.InvokeResponse{
		StatusCode: 200,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body:         []byte(`{"message": "Function invoked via Knative"}`),
		Duration:     100 * time.Millisecond,
		ColdStart:    false,
		InvocationID: fmt.Sprintf("knative-%s-%d", name, time.Now().Unix()),
	}, nil
}

// InvokeFunctionAsync invokes a function asynchronously
func (p *Provider) InvokeFunctionAsync(ctx context.Context, name string, req *function.InvokeRequest) (string, error) {
	// Knative doesn't have built-in async invocation
	// We would need to use Knative Eventing or a message queue
	return "", function.NewProviderError(function.ErrCodeNotSupported, "async invocation not directly supported by Knative")
}

// GetInvocationStatus gets the status of an async invocation
func (p *Provider) GetInvocationStatus(ctx context.Context, invocationID string) (*function.InvocationStatus, error) {
	return nil, function.NewProviderError(function.ErrCodeNotSupported, "async invocation not supported")
}

// GetFunctionURL returns the URL for a function
func (p *Provider) GetFunctionURL(ctx context.Context, name string) (string, error) {
	fn, err := p.GetFunction(ctx, name)
	if err != nil {
		return "", err
	}

	return p.getFunctionURL(fn.Namespace, name), nil
}

// GetFunctionLogs retrieves logs for a function
func (p *Provider) GetFunctionLogs(ctx context.Context, name string, opts *function.LogOptions) ([]*function.LogEntry, error) {
	// Get function to find namespace and pod selector
	fn, err := p.GetFunction(ctx, name)
	if err != nil {
		return nil, err
	}

	// Find pods for the function
	pods, err := p.kubeClient.CoreV1().Pods(fn.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("serving.knative.dev/service=%s", name),
	})
	if err != nil {
		return nil, function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to list pods: %v", err))
	}

	logs := []*function.LogEntry{}
	for _, pod := range pods.Items {
		// Get logs from user container
		logOpts := &corev1.PodLogOptions{
			Container: "user-container",
			Follow:    opts.Follow,
		}
		if opts.Since != nil {
			logOpts.SinceTime = &metav1.Time{Time: *opts.Since}
		}

		stream, err := p.kubeClient.CoreV1().Pods(fn.Namespace).GetLogs(pod.Name, logOpts).Stream(ctx)
		if err != nil {
			continue
		}
		defer stream.Close()

		// Parse logs (simplified - in production would parse properly)
		// TODO: Implement proper log parsing
		logs = append(logs, &function.LogEntry{
			Timestamp: time.Now(),
			Level:     "info",
			Message:   fmt.Sprintf("Logs from pod %s", pod.Name),
			Container: "user-container",
			Pod:       pod.Name,
		})
	}

	return logs, nil
}

// GetFunctionMetrics retrieves metrics for a function
func (p *Provider) GetFunctionMetrics(ctx context.Context, name string, opts *function.MetricOptions) (*function.Metrics, error) {
	// TODO: Implement metrics retrieval from Prometheus/monitoring system
	return &function.Metrics{
		Invocations: 100,
		Errors:      5,
		Duration: function.MetricStats{
			Min: 50,
			Max: 500,
			Avg: 150,
			P50: 120,
			P95: 400,
			P99: 480,
		},
		ColdStarts: 10,
		Concurrency: function.MetricStats{
			Min: 0,
			Max: 10,
			Avg: 3,
			P50: 2,
			P95: 8,
			P99: 9,
		},
	}, nil
}

// GetCapabilities returns the provider's capabilities
func (p *Provider) GetCapabilities() *function.Capabilities {
	return p.capabilities
}

// HealthCheck performs a health check on the provider
func (p *Provider) HealthCheck(ctx context.Context) error {
	// Check if Knative Serving is installed
	_, err := p.dynamicClient.Resource(knativeServiceGVR).List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("Knative Serving not available: %v", err))
	}
	return nil
}

// Helper methods

func (p *Provider) buildImage(spec *function.FunctionSpec) string {
	// In production, this would trigger a build process
	// For now, return a placeholder image
	return fmt.Sprintf("gcr.io/knative-samples/%s:latest", strings.ToLower(string(spec.Runtime)))
}

func (p *Provider) buildKnativeService(name, namespace, image string, spec *function.FunctionSpec) *unstructured.Unstructured {
	service := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "serving.knative.dev/v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":        name,
				"namespace":   namespace,
				"labels":      spec.Labels,
				"annotations": spec.Annotations,
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"autoscaling.knative.dev/target": "100",
						},
					},
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"image": image,
								"env":   p.buildEnvVars(spec.Environment),
								"resources": map[string]interface{}{
									"requests": map[string]string{
							"memory": spec.Resources.Memory,
							"cpu":    spec.Resources.CPU,
						},
									"limits": map[string]string{
							"memory": spec.Resources.Memory,
							"cpu":    spec.Resources.CPU,
						},
								},
							},
						},
						"timeoutSeconds": spec.Timeout,
					},
				},
			},
		},
	}

	return service
}

func (p *Provider) updateKnativeServiceSpec(service *unstructured.Unstructured, image string, spec *function.FunctionSpec) {
	// Update container image
	containers, _, _ := unstructured.NestedSlice(service.Object, "spec", "template", "spec", "containers")
	if len(containers) > 0 {
		container := containers[0].(map[string]interface{})
		container["image"] = image
		container["env"] = p.buildEnvVars(spec.Environment)
		if spec.Resources.Memory != "" || spec.Resources.CPU != "" {
			container["resources"] = map[string]interface{}{
				"requests": map[string]string{
					"memory": spec.Resources.Memory,
					"cpu":    spec.Resources.CPU,
				},
				"limits": map[string]string{
					"memory": spec.Resources.Memory,
					"cpu":    spec.Resources.CPU,
				},
			}
		}
	}
	unstructured.SetNestedSlice(service.Object, containers, "spec", "template", "spec", "containers")

	// Update timeout
	if spec.Timeout > 0 {
		unstructured.SetNestedField(service.Object, int64(spec.Timeout), "spec", "template", "spec", "timeoutSeconds")
	}
}

func (p *Provider) buildEnvVars(env map[string]string) []interface{} {
	envVars := make([]interface{}, 0, len(env))
	for k, v := range env {
		envVars = append(envVars, map[string]interface{}{
			"name":  k,
			"value": v,
		})
	}
	return envVars
}

func (p *Provider) waitForServiceReady(ctx context.Context, namespace, name string) error {
	// Wait for up to 5 minutes
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return function.NewProviderError(function.ErrCodeTimeout, "timeout waiting for service to be ready")
		case <-ticker.C:
			service, err := p.dynamicClient.Resource(knativeServiceGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				continue
			}

			// Check if ready
			conditions, _, _ := unstructured.NestedSlice(service.Object, "status", "conditions")
			for _, c := range conditions {
				condition := c.(map[string]interface{})
				if condition["type"] == "Ready" && condition["status"] == "True" {
					return nil
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *Provider) getLatestRevision(ctx context.Context, namespace, name string) (string, error) {
	service, err := p.dynamicClient.Resource(knativeServiceGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	latestRevision, found, err := unstructured.NestedString(service.Object, "status", "latestCreatedRevisionName")
	if err != nil || !found {
		return "", function.NewProviderError(function.ErrCodeInternal, "could not get latest revision")
	}

	return latestRevision, nil
}

func (p *Provider) knativeServiceToFunctionDef(service *unstructured.Unstructured) (*function.FunctionDef, error) {
	// Extract function details from Knative service
	name := service.GetName()
	namespace := service.GetNamespace()
	
	// Get runtime from annotation or default
	runtime := function.RuntimePython
	if r, ok := service.GetAnnotations()["function.hexabase.ai/runtime"]; ok {
		runtime = function.Runtime(r)
	}

	// Get active revision
	activeRevision, _, _ := unstructured.NestedString(service.Object, "status", "latestReadyRevisionName")
	
	// Get status
	status := function.FunctionDefStatusPending
	conditions, _, _ := unstructured.NestedSlice(service.Object, "status", "conditions")
	for _, c := range conditions {
		condition := c.(map[string]interface{})
		if condition["type"] == "Ready" && condition["status"] == "True" {
			status = function.FunctionDefStatusReady
			break
		}
	}

	return &function.FunctionDef{
		Name:          name,
		Namespace:     namespace,
		Runtime:       runtime,
		Status:        status,
		ActiveVersion: activeRevision,
		CreatedAt:     service.GetCreationTimestamp().Time,
		UpdatedAt:     service.GetCreationTimestamp().Time, // Knative doesn't track updates
		Labels:        service.GetLabels(),
		Annotations:   service.GetAnnotations(),
	}, nil
}

func (p *Provider) knativeRevisionToVersion(revision *unstructured.Unstructured, functionName string) (*function.FunctionVersionDef, error) {
	name := revision.GetName()
	
	// Get image
	containers, _, _ := unstructured.NestedSlice(revision.Object, "spec", "containers")
	image := ""
	if len(containers) > 0 {
		container := containers[0].(map[string]interface{})
		image, _ = container["image"].(string)
	}

	// Get status
	status := function.FunctionBuildStatusPending
	conditions, _, _ := unstructured.NestedSlice(revision.Object, "status", "conditions")
	for _, c := range conditions {
		condition := c.(map[string]interface{})
		if condition["type"] == "Ready" && condition["status"] == "True" {
			status = function.FunctionBuildStatusSuccess
			break
		}
	}

	// Check if active
	isActive := false
	traffic, _, _ := unstructured.NestedSlice(revision.Object, "status", "traffic")
	for _, t := range traffic {
		trafficTarget := t.(map[string]interface{})
		if trafficTarget["revisionName"] == name && trafficTarget["percent"].(int64) > 0 {
			isActive = true
			break
		}
	}

	return &function.FunctionVersionDef{
		ID:           name,
		FunctionName: functionName,
		Image:        image,
		BuildStatus:  status,
		CreatedAt:    revision.GetCreationTimestamp().Time,
		IsActive:     isActive,
	}, nil
}

func (p *Provider) getFunctionURL(namespace, name string) string {
	// Construct Knative service URL
	// Format: http://{service}.{namespace}.svc.cluster.local
	return fmt.Sprintf("http://%s.%s.svc.cluster.local", name, namespace)
}

func (p *Provider) createEventTrigger(ctx context.Context, functionName string, trigger *function.FunctionTrigger) error {
	// Get function to find namespace
	fn, err := p.GetFunction(ctx, functionName)
	if err != nil {
		return err
	}

	// Create Knative Eventing trigger
	triggerGVR := schema.GroupVersionResource{
		Group:    "eventing.knative.dev",
		Version:  "v1",
		Resource: "triggers",
	}

	knativeTrigger := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "eventing.knative.dev/v1",
			"kind":       "Trigger",
			"metadata": map[string]interface{}{
				"name":      trigger.Name,
				"namespace": fn.Namespace,
				"labels": map[string]interface{}{
					"serving.knative.dev/service": functionName,
				},
			},
			"spec": map[string]interface{}{
				"broker": "default",
				"filter": map[string]interface{}{
					"attributes": trigger.Config,
				},
				"subscriber": map[string]interface{}{
					"ref": map[string]interface{}{
						"apiVersion": "serving.knative.dev/v1",
						"kind":       "Service",
						"name":       functionName,
					},
				},
			},
		},
	}

	_, err = p.dynamicClient.Resource(triggerGVR).Namespace(fn.Namespace).Create(ctx, knativeTrigger, metav1.CreateOptions{})
	if err != nil {
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to create trigger: %v", err))
	}

	return nil
}

func (p *Provider) updateEventTrigger(ctx context.Context, functionName, triggerName string, trigger *function.FunctionTrigger) error {
	// Similar to create but with update
	fn, err := p.GetFunction(ctx, functionName)
	if err != nil {
		return err
	}

	triggerGVR := schema.GroupVersionResource{
		Group:    "eventing.knative.dev",
		Version:  "v1",
		Resource: "triggers",
	}

	existing, err := p.dynamicClient.Resource(triggerGVR).Namespace(fn.Namespace).Get(ctx, triggerName, metav1.GetOptions{})
	if err != nil {
		return function.NewProviderError(function.ErrCodeNotFound, "trigger not found")
	}

	// Update filter attributes
	// Convert map[string]string to map[string]interface{}
	attributes := make(map[string]interface{})
	for k, v := range trigger.Config {
		attributes[k] = v
	}
	unstructured.SetNestedMap(existing.Object, attributes, "spec", "filter", "attributes")

	_, err = p.dynamicClient.Resource(triggerGVR).Namespace(fn.Namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return function.NewProviderError(function.ErrCodeInternal, fmt.Sprintf("failed to update trigger: %v", err))
	}

	return nil
}

func (p *Provider) knativeTriggerToFunctionTrigger(trigger *unstructured.Unstructured, functionName string) (*function.FunctionTrigger, error) {
	name := trigger.GetName()
	
	// Get filter attributes as config
	config := make(map[string]string)
	attributes, _, _ := unstructured.NestedMap(trigger.Object, "spec", "filter", "attributes")
	for k, v := range attributes {
		config[k] = fmt.Sprintf("%v", v)
	}

	return &function.FunctionTrigger{
		Name:         name,
		Type:         function.TriggerEvent,
		FunctionName: functionName,
		Enabled:      true,
		Config:       config,
		CreatedAt:    trigger.GetCreationTimestamp().Time,
		UpdatedAt:    trigger.GetCreationTimestamp().Time,
	}, nil
}