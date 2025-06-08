package monitoring

import (
	"context"

	"github.com/hexabase/hexabase-ai/api/internal/domain/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	k8sClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type kubernetesRepository struct {
	client        k8sClient.Interface
	dynamicClient dynamic.Interface
	config        *rest.Config
}

// NewKubernetesRepository creates a new Kubernetes repository for monitoring
func NewKubernetesRepository(
	client k8sClient.Interface,
	dynamicClient dynamic.Interface,
	config *rest.Config,
) kubernetes.Repository {
	return &kubernetesRepository{
		client:        client,
		dynamicClient: dynamicClient,
		config:        config,
	}
}

func (r *kubernetesRepository) GetPodMetrics(ctx context.Context, namespace string) (*kubernetes.PodMetricsList, error) {
	// TODO: Implement using metrics-server API
	// For now, return a mock response
	return &kubernetes.PodMetricsList{
		Items: []kubernetes.PodMetrics{
			{
				Name:      "example-pod",
				Namespace: namespace,
				CPU: kubernetes.ResourceMetric{
					Value: "100m",
					Unit:  "millicores",
				},
				Memory: kubernetes.ResourceMetric{
					Value: "128Mi",
					Unit:  "bytes",
				},
				Timestamp: "2024-01-01T00:00:00Z",
			},
		},
	}, nil
}

func (r *kubernetesRepository) GetNodeMetrics(ctx context.Context) (*kubernetes.NodeMetricsList, error) {
	// TODO: Implement using metrics-server API
	// For now, return a mock response
	return &kubernetes.NodeMetricsList{
		Items: []kubernetes.NodeMetrics{
			{
				Name: "node-1",
				CPU: kubernetes.ResourceMetric{
					Value: "2000m",
					Unit:  "millicores",
				},
				Memory: kubernetes.ResourceMetric{
					Value: "4Gi",
					Unit:  "bytes",
				},
				Storage: kubernetes.ResourceMetric{
					Value: "100Gi",
					Unit:  "bytes",
				},
				Timestamp: "2024-01-01T00:00:00Z",
			},
		},
	}, nil
}

func (r *kubernetesRepository) GetNamespaceResourceQuota(ctx context.Context, namespace string) (*kubernetes.ResourceQuota, error) {
	quotas, err := r.client.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(quotas.Items) == 0 {
		// Return default quota if none exists
		return &kubernetes.ResourceQuota{
			Name:      "default",
			Namespace: namespace,
			Hard: map[string]string{
				"requests.cpu":    "2",
				"requests.memory": "4Gi",
				"limits.cpu":      "4",
				"limits.memory":   "8Gi",
				"pods":            "10",
			},
			Used: map[string]string{
				"requests.cpu":    "1",
				"requests.memory": "2Gi",
				"limits.cpu":      "2",
				"limits.memory":   "4Gi",
				"pods":            "5",
			},
		}, nil
	}

	quota := quotas.Items[0]
	hard := make(map[string]string)
	used := make(map[string]string)

	for k, v := range quota.Spec.Hard {
		hard[string(k)] = v.String()
	}

	for k, v := range quota.Status.Used {
		used[string(k)] = v.String()
	}

	return &kubernetes.ResourceQuota{
		Name:      quota.Name,
		Namespace: quota.Namespace,
		Hard:      hard,
		Used:      used,
	}, nil
}

func (r *kubernetesRepository) GetClusterInfo(ctx context.Context) (*kubernetes.ClusterInfo, error) {
	version, err := r.client.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}

	nodes, err := r.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	namespaces, err := r.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods, err := r.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return &kubernetes.ClusterInfo{
		Version:        version.String(),
		Platform:       "kubernetes",
		NodeCount:      len(nodes.Items),
		PodCount:       len(pods.Items),
		NamespaceCount: len(namespaces.Items),
	}, nil
}

func (r *kubernetesRepository) CheckComponentHealth(ctx context.Context) (map[string]kubernetes.ComponentStatus, error) {
	components := make(map[string]kubernetes.ComponentStatus)

	// Check component statuses
	componentStatuses, err := r.client.CoreV1().ComponentStatuses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, cs := range componentStatuses.Items {
		healthy := true
		message := "Component is healthy"

		for _, condition := range cs.Conditions {
			if condition.Type == v1.ComponentHealthy && condition.Status != v1.ConditionTrue {
				healthy = false
				message = condition.Message
				break
			}
		}

		components[cs.Name] = kubernetes.ComponentStatus{
			Name:    cs.Name,
			Healthy: healthy,
			Message: message,
		}
	}

	// If no component statuses available, check basic cluster health
	if len(components) == 0 {
		// Try to get nodes as a basic health check
		_, err := r.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
		if err != nil {
			components["api-server"] = kubernetes.ComponentStatus{
				Name:    "api-server",
				Healthy: false,
				Message: err.Error(),
			}
		} else {
			components["api-server"] = kubernetes.ComponentStatus{
				Name:    "api-server",
				Healthy: true,
				Message: "API server is responding",
			}
		}
	}

	return components, nil
}