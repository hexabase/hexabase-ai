package function

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// KnativeIntegration handles Knative Serving operations
type KnativeIntegration struct {
	k8sClient     kubernetes.Interface
	dynamicClient dynamic.Interface
}

// NewKnativeIntegration creates a new Knative integration
func NewKnativeIntegration(k8sClient kubernetes.Interface, dynamicClient dynamic.Interface) *KnativeIntegration {
	return &KnativeIntegration{
		k8sClient:     k8sClient,
		dynamicClient: dynamicClient,
	}
}

// KnativeService represents a Knative Service configuration
type KnativeService struct {
	Name            string
	Namespace       string
	Image           string
	Port            int32
	Env             map[string]string
	Resources       ResourceRequirements
	Autoscaling     AutoscalingConfig
	TimeoutSeconds  int64
	Concurrency     int64
	Labels          map[string]string
	Annotations     map[string]string
}

// ResourceRequirements defines resource limits and requests
type ResourceRequirements struct {
	Requests ResourceList
	Limits   ResourceList
}

// ResourceList defines CPU and memory resources
type ResourceList struct {
	CPU    string
	Memory string
}

// AutoscalingConfig defines autoscaling parameters
type AutoscalingConfig struct {
	MinScale              int
	MaxScale              int
	Target                int
	Metric                string
	ScaleDownDelay       string
	StableWindow         string
	InitialScale         int
	ScaleToZeroEnabled   bool
	TargetUtilization    int
}

// CreateKnativeService creates a new Knative Service
func (k *KnativeIntegration) CreateKnativeService(ctx context.Context, svc *KnativeService) error {
	knativeService := k.buildKnativeService(svc)
	
	gvr := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "services",
	}
	
	_, err := k.dynamicClient.Resource(gvr).Namespace(svc.Namespace).Create(
		ctx,
		knativeService,
		metav1.CreateOptions{},
	)
	
	return err
}

// UpdateKnativeService updates an existing Knative Service
func (k *KnativeIntegration) UpdateKnativeService(ctx context.Context, svc *KnativeService) error {
	knativeService := k.buildKnativeService(svc)
	
	gvr := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "services",
	}
	
	_, err := k.dynamicClient.Resource(gvr).Namespace(svc.Namespace).Update(
		ctx,
		knativeService,
		metav1.UpdateOptions{},
	)
	
	return err
}

// DeleteKnativeService deletes a Knative Service
func (k *KnativeIntegration) DeleteKnativeService(ctx context.Context, namespace, name string) error {
	gvr := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "services",
	}
	
	return k.dynamicClient.Resource(gvr).Namespace(namespace).Delete(
		ctx,
		name,
		metav1.DeleteOptions{},
	)
}

// GetKnativeService retrieves a Knative Service
func (k *KnativeIntegration) GetKnativeService(ctx context.Context, namespace, name string) (*unstructured.Unstructured, error) {
	gvr := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "services",
	}
	
	return k.dynamicClient.Resource(gvr).Namespace(namespace).Get(
		ctx,
		name,
		metav1.GetOptions{},
	)
}

// GetServiceURL retrieves the URL of a Knative Service
func (k *KnativeIntegration) GetServiceURL(ctx context.Context, namespace, name string) (string, error) {
	svc, err := k.GetKnativeService(ctx, namespace, name)
	if err != nil {
		return "", err
	}
	
	status, found, err := unstructured.NestedMap(svc.Object, "status")
	if err != nil || !found {
		return "", fmt.Errorf("service status not found")
	}
	
	url, found, err := unstructured.NestedString(status, "url")
	if err != nil || !found {
		return "", fmt.Errorf("service URL not found")
	}
	
	return url, nil
}

// buildKnativeService builds the Knative Service unstructured object
func (k *KnativeIntegration) buildKnativeService(svc *KnativeService) *unstructured.Unstructured {
	// Build annotations
	annotations := make(map[string]string)
	for k, v := range svc.Annotations {
		annotations[k] = v
	}
	
	// Add autoscaling annotations
	if svc.Autoscaling.MinScale >= 0 {
		annotations["autoscaling.knative.dev/min-scale"] = fmt.Sprintf("%d", svc.Autoscaling.MinScale)
	}
	if svc.Autoscaling.MaxScale > 0 {
		annotations["autoscaling.knative.dev/max-scale"] = fmt.Sprintf("%d", svc.Autoscaling.MaxScale)
	}
	if svc.Autoscaling.Target > 0 {
		annotations["autoscaling.knative.dev/target"] = fmt.Sprintf("%d", svc.Autoscaling.Target)
	}
	if svc.Autoscaling.Metric != "" {
		annotations["autoscaling.knative.dev/metric"] = svc.Autoscaling.Metric
	}
	if svc.Autoscaling.ScaleDownDelay != "" {
		annotations["autoscaling.knative.dev/scale-down-delay"] = svc.Autoscaling.ScaleDownDelay
	}
	if svc.Autoscaling.StableWindow != "" {
		annotations["autoscaling.knative.dev/window"] = svc.Autoscaling.StableWindow
	}
	if svc.Autoscaling.InitialScale > 0 {
		annotations["autoscaling.knative.dev/initial-scale"] = fmt.Sprintf("%d", svc.Autoscaling.InitialScale)
	}
	
	// Build environment variables
	env := []interface{}{}
	for k, v := range svc.Env {
		env = append(env, map[string]interface{}{
			"name":  k,
			"value": v,
		})
	}
	
	// Build container spec
	container := map[string]interface{}{
		"image": svc.Image,
		"ports": []interface{}{
			map[string]interface{}{
				"containerPort": svc.Port,
			},
		},
		"env": env,
	}
	
	// Add resources if specified
	if svc.Resources.Requests.CPU != "" || svc.Resources.Requests.Memory != "" ||
		svc.Resources.Limits.CPU != "" || svc.Resources.Limits.Memory != "" {
		resources := map[string]interface{}{}
		
		if svc.Resources.Requests.CPU != "" || svc.Resources.Requests.Memory != "" {
			requests := map[string]interface{}{}
			if svc.Resources.Requests.CPU != "" {
				requests["cpu"] = svc.Resources.Requests.CPU
			}
			if svc.Resources.Requests.Memory != "" {
				requests["memory"] = svc.Resources.Requests.Memory
			}
			resources["requests"] = requests
		}
		
		if svc.Resources.Limits.CPU != "" || svc.Resources.Limits.Memory != "" {
			limits := map[string]interface{}{}
			if svc.Resources.Limits.CPU != "" {
				limits["cpu"] = svc.Resources.Limits.CPU
			}
			if svc.Resources.Limits.Memory != "" {
				limits["memory"] = svc.Resources.Limits.Memory
			}
			resources["limits"] = limits
		}
		
		container["resources"] = resources
	}
	
	// Build the Knative Service object
	knativeService := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "serving.knative.dev/v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":        svc.Name,
				"namespace":   svc.Namespace,
				"labels":      svc.Labels,
				"annotations": annotations,
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"annotations": annotations,
						"labels":      svc.Labels,
					},
					"spec": map[string]interface{}{
						"containers":              []interface{}{container},
						"timeoutSeconds":          svc.TimeoutSeconds,
						"containerConcurrency":    svc.Concurrency,
					},
				},
			},
		},
	}
	
	return knativeService
}

// WaitForServiceReady waits for a Knative Service to become ready
func (k *KnativeIntegration) WaitForServiceReady(ctx context.Context, namespace, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		svc, err := k.GetKnativeService(ctx, namespace, name)
		if err != nil {
			return err
		}
		
		conditions, found, err := unstructured.NestedSlice(svc.Object, "status", "conditions")
		if err != nil || !found {
			time.Sleep(2 * time.Second)
			continue
		}
		
		for _, c := range conditions {
			condition, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			
			if condition["type"] == "Ready" && condition["status"] == "True" {
				return nil
			}
		}
		
		time.Sleep(2 * time.Second)
	}
	
	return fmt.Errorf("timeout waiting for service to be ready")
}

// GetServiceRevisions gets all revisions for a Knative Service
func (k *KnativeIntegration) GetServiceRevisions(ctx context.Context, namespace, serviceName string) ([]string, error) {
	gvr := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "revisions",
	}
	
	revisions, err := k.dynamicClient.Resource(gvr).Namespace(namespace).List(
		ctx,
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("serving.knative.dev/service=%s", serviceName),
		},
	)
	if err != nil {
		return nil, err
	}
	
	var revisionNames []string
	for _, item := range revisions.Items {
		revisionNames = append(revisionNames, item.GetName())
	}
	
	return revisionNames, nil
}

// SetTrafficSplit sets traffic distribution between revisions
func (k *KnativeIntegration) SetTrafficSplit(ctx context.Context, namespace, serviceName string, traffic []TrafficTarget) error {
	svc, err := k.GetKnativeService(ctx, namespace, serviceName)
	if err != nil {
		return err
	}
	
	// Build traffic array
	trafficArray := []interface{}{}
	for _, t := range traffic {
		target := map[string]interface{}{
			"percent": t.Percent,
		}
		if t.RevisionName != "" {
			target["revisionName"] = t.RevisionName
		}
		if t.Tag != "" {
			target["tag"] = t.Tag
		}
		if t.LatestRevision {
			target["latestRevision"] = true
		}
		trafficArray = append(trafficArray, target)
	}
	
	// Update the service with new traffic configuration
	if err := unstructured.SetNestedSlice(svc.Object, trafficArray, "spec", "traffic"); err != nil {
		return err
	}
	
	gvr := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  "v1",
		Resource: "services",
	}
	
	_, err = k.dynamicClient.Resource(gvr).Namespace(namespace).Update(
		ctx,
		svc,
		metav1.UpdateOptions{},
	)
	
	return err
}

// TrafficTarget defines traffic routing to a specific revision
type TrafficTarget struct {
	RevisionName   string
	Percent        int64
	Tag            string
	LatestRevision bool
}