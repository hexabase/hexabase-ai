package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

// Client wraps Kubernetes and Knative clients
type Client struct {
	K8sClient     kubernetes.Interface
	KnativeClient versioned.Interface
	Namespace     string
}

// NewClient creates a new function client
func NewClient(k8sClient kubernetes.Interface, knativeClient versioned.Interface, namespace string) *Client {
	return &Client{
		K8sClient:     k8sClient,
		KnativeClient: knativeClient,
		Namespace:     namespace,
	}
}


// GetFunction gets a specific function
func (c *Client) GetFunction(ctx context.Context, name string) (*FunctionInfo, error) {
	svc, err := c.KnativeClient.ServingV1().Services(c.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get function %s: %w", name, err)
	}

	url := ""
	if svc.Status.URL != nil {
		url = svc.Status.URL.String()
	}
	
	// Check if service is ready by looking at conditions
	ready := false
	for _, cond := range svc.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			ready = true
			break
		}
	}
	
	image := ""
	if len(svc.Spec.Template.Spec.Containers) > 0 {
		image = svc.Spec.Template.Spec.Containers[0].Image
	}
	
	return &FunctionInfo{
		Name:      svc.Name,
		Namespace: svc.Namespace,
		URL:       url,
		Ready:     ready,
		Created:   svc.CreationTimestamp.Time,
		Image:     image,
	}, nil
}


// InvokeRequest represents a function invocation request
type InvokeRequest struct {
	Namespace string
	Method    string
	Path      string
	Headers   map[string]string
	Body      []byte
	Async     bool
}

// InvokeResponse represents a function invocation response
type InvokeResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	Duration   time.Duration
}

// InvokeFunction invokes a function
func (c *Client) InvokeFunction(ctx context.Context, name string, req InvokeRequest) (*InvokeResponse, error) {
	// This would typically use an HTTP client to invoke the function
	// For now, this is a placeholder
	return nil, fmt.Errorf("invoke not yet implemented")
}

// GetLogs gets logs for a function
func (c *Client) GetLogs(ctx context.Context, name string, since time.Duration) ([]string, error) {
	// This would get logs from the pods running the function
	// For now, this is a placeholder
	return nil, fmt.Errorf("logs not yet implemented")
}

// FunctionInfo contains information about a function
type FunctionInfo struct {
	Name      string
	Namespace string
	URL       string
	Ready     bool
	Created   time.Time
	Image     string
}

// FunctionDetails represents detailed function information
type FunctionDetails struct {
	FunctionInfo
	Error             string
	Revision          string
	Replicas          int
	AvailableReplicas int
	Conditions        []Condition
	Annotations       map[string]string
	Labels            map[string]string
	Events            []EventConfig
}

// Condition represents a Knative condition
type Condition struct {
	Type    string
	Status  string
	Message string
	Reason  string
}

// EventConfig represents event configuration
type EventConfig struct {
	Type   string
	Source string
}

// APIClient is an alias for Client for backward compatibility
type APIClient = Client

// Function is an alias for FunctionInfo
type Function = FunctionInfo

// NewAPIClient creates a new API client from kubeconfig
func NewAPIClient() (*Client, error) {
	// Build kubeconfig path
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	// Create config
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	// Create Kubernetes client
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Create Knative client
	knativeClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Knative client: %w", err)
	}

	// Get namespace from config or use default
	namespace := "default"
	if ns := os.Getenv("FUNCTION_NAMESPACE"); ns != "" {
		namespace = ns
	}

	return NewClient(k8sClient, knativeClient, namespace), nil
}

// DeleteFunction deletes a function with optional namespace
func (c *Client) DeleteFunction(ctx context.Context, name string, namespace string) error {
	if namespace == "" {
		namespace = c.Namespace
	}
	return c.KnativeClient.ServingV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ListAllFunctions lists functions from all namespaces
func (c *Client) ListAllFunctions(ctx context.Context, limit int) ([]FunctionInfo, error) {
	services, err := c.KnativeClient.ServingV1().Services("").List(ctx, metav1.ListOptions{
		Limit: int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}

	var functions []FunctionInfo
	for _, svc := range services.Items {
		url := ""
		if svc.Status.URL != nil {
			url = svc.Status.URL.String()
		}
		
		// Check if service is ready by looking at conditions
		ready := false
		for _, cond := range svc.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == "True" {
				ready = true
				break
			}
		}
		
		functions = append(functions, FunctionInfo{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			URL:       url,
			Ready:     ready,
			Created:   svc.CreationTimestamp.Time,
		})
	}

	return functions, nil
}

// ListFunctions lists functions in a specific namespace
func (c *Client) ListFunctions(ctx context.Context, namespace string, limit int) ([]FunctionInfo, error) {
	if namespace == "" {
		namespace = c.Namespace
	}
	
	services, err := c.KnativeClient.ServingV1().Services(namespace).List(ctx, metav1.ListOptions{
		Limit: int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}

	var functions []FunctionInfo
	for _, svc := range services.Items {
		url := ""
		if svc.Status.URL != nil {
			url = svc.Status.URL.String()
		}
		
		// Check if service is ready by looking at conditions
		ready := false
		for _, cond := range svc.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == "True" {
				ready = true
				break
			}
		}
		
		functions = append(functions, FunctionInfo{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			URL:       url,
			Ready:     ready,
			Created:   svc.CreationTimestamp.Time,
		})
	}

	return functions, nil
}