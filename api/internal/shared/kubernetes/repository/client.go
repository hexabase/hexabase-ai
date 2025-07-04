package kubernetes

import (
	"context"
	"fmt"
	"strings"

	kubernetes "github.com/hexabase/hexabase-ai/api/internal/shared/kubernetes/domain"
	workspaceDomain "github.com/hexabase/hexabase-ai/api/internal/workspace/domain"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// kubernetesRepository implements kubernetes.Repository
type kubernetesRepository struct {
	client        k8s.Interface
	metricsClient versioned.Interface
}

// NewKubernetesRepository creates a new Kubernetes repository
func NewKubernetesRepository(client k8s.Interface) kubernetes.Repository {
	// In production, you'd also initialize metrics client here
	return &kubernetesRepository{
		client: client,
	}
}

// GetPodMetrics retrieves metrics for pods in a namespace
func (r *kubernetesRepository) GetPodMetrics(ctx context.Context, namespace string) (*kubernetes.PodMetricsList, error) {
	// This would use the metrics API in production
	// For now, return mock data
	return &kubernetes.PodMetricsList{
		Items: []kubernetes.PodMetrics{
			{
				Name:      "app-pod-1",
				Namespace: namespace,
				CPU:       kubernetes.ResourceMetric{Value: "100m", Unit: "millicores"},
				Memory:    kubernetes.ResourceMetric{Value: "256Mi", Unit: "MiB"},
				Timestamp: "2023-01-01T00:00:00Z",
			},
		},
	}, nil
}

// GetNodeMetrics retrieves metrics for nodes
func (r *kubernetesRepository) GetNodeMetrics(ctx context.Context) (*kubernetes.NodeMetricsList, error) {
	// Mock implementation
	return &kubernetes.NodeMetricsList{
		Items: []kubernetes.NodeMetrics{
			{
				Name:      "node-1",
				CPU:       kubernetes.ResourceMetric{Value: "2000m", Unit: "millicores"},
				Memory:    kubernetes.ResourceMetric{Value: "8Gi", Unit: "GiB"},
				Storage:   kubernetes.ResourceMetric{Value: "100Gi", Unit: "GiB"},
				Timestamp: "2023-01-01T00:00:00Z",
			},
		},
	}, nil
}

// GetNamespaceResourceQuota gets resource quota for a namespace
func (r *kubernetesRepository) GetNamespaceResourceQuota(ctx context.Context, namespace string) (*kubernetes.ResourceQuota, error) {
	quotas, err := r.client.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list resource quotas: %w", err)
	}

	if len(quotas.Items) == 0 {
		// Return default quota if none exists
		return &kubernetes.ResourceQuota{
			Name:      "default",
			Namespace: namespace,
			Hard: map[string]string{
				"cpu":    "2",
				"memory": "4Gi",
				"pods":   "50",
			},
			Used: map[string]string{
				"cpu":    "0.5",
				"memory": "1Gi",
				"pods":   "5",
			},
		}, nil
	}

	// Convert the first quota
	quota := quotas.Items[0]
	hard := make(map[string]string)
	used := make(map[string]string)

	for k, v := range quota.Status.Hard {
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

// GetClusterInfo retrieves general cluster information
func (r *kubernetesRepository) GetClusterInfo(ctx context.Context) (*kubernetes.ClusterInfo, error) {
	// Get version info
	version, err := r.client.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	// Count nodes
	nodes, err := r.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Count pods across all namespaces
	pods, err := r.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Count namespaces
	namespaces, err := r.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	return &kubernetes.ClusterInfo{
		Version:        version.String(),
		Platform:       version.Platform,
		NodeCount:      len(nodes.Items),
		PodCount:       len(pods.Items),
		NamespaceCount: len(namespaces.Items),
	}, nil
}

// getVClusterClient is a placeholder for retrieving a clientset scoped to a vCluster.
func (r *kubernetesRepository) getVClusterClient(ctx context.Context, workspaceID string) (k8s.Interface, string, error) {
	// In a real implementation, this would use the vcluster syncer to get the REST config
	// for the specific vcluster's control plane.
	// For now, it returns the host clientset and a conventional namespace name.
	namespace := "vcluster-" + workspaceID
	return r.client, namespace, nil
}

// ListVClusterNodes lists the nodes available to a specific vCluster.
func (r *kubernetesRepository) ListVClusterNodes(ctx context.Context, workspaceID string) ([]workspaceDomain.Node, error) {
	clientset, _, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get vcluster client: %w", err)
	}
	
	nodeList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes for workspace %s: %w", workspaceID, err)
	}

	var nodes []workspaceDomain.Node
	for _, item := range nodeList.Items {
		// Determine node status
		status := "NotReady"
		for _, cond := range item.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				status = "Ready"
				break
			}
		}
		
		nodes = append(nodes, workspaceDomain.Node{
			Name:   item.Name,
			Status: status,
			CPU:    item.Status.Capacity.Cpu().String(),
			Memory: item.Status.Capacity.Memory().String(),
		})
	}

	return nodes, nil
}

// ScaleVClusterDeployment scales a deployment within a vCluster.
func (r *kubernetesRepository) ScaleVClusterDeployment(ctx context.Context, workspaceID, deploymentName string, replicas int) error {
	clientset, namespace, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to get vcluster client: %w", err)
	}

	scale, err := clientset.AppsV1().Deployments(namespace).GetScale(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get scale for deployment %s in namespace %s: %w", deploymentName, namespace, err)
	}

	scale.Spec.Replicas = int32(replicas)

	_, err = clientset.AppsV1().Deployments(namespace).UpdateScale(ctx, deploymentName, scale, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update scale for deployment %s in namespace %s: %w", deploymentName, namespace, err)
	}

	return nil
}

// CheckComponentHealth checks health of cluster components
func (r *kubernetesRepository) CheckComponentHealth(ctx context.Context) (map[string]kubernetes.ComponentStatus, error) {
	components := make(map[string]kubernetes.ComponentStatus)

	// Check core components by verifying system pods
	systemPods, err := r.client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list system pods: %w", err)
	}

	// Check API server
	components["api-server"] = kubernetes.ComponentStatus{
		Name:    "api-server",
		Healthy: true,
		Message: "API server is responding",
	}

	// Check other components based on system pods
	componentNames := map[string]string{
		"etcd":               "etcd",
		"controller-manager": "kube-controller-manager",
		"scheduler":          "kube-scheduler",
		"dns":                "coredns",
	}

	for component, podPrefix := range componentNames {
		found := false
		allHealthy := true
		
		for _, pod := range systemPods.Items {
			if containsPrefix(pod.Name, podPrefix) {
				found = true
				if pod.Status.Phase != corev1.PodRunning {
					allHealthy = false
					break
				}
			}
		}

		if found && allHealthy {
			components[component] = kubernetes.ComponentStatus{
				Name:    component,
				Healthy: true,
				Message: fmt.Sprintf("%s is running", component),
			}
		} else if found {
			components[component] = kubernetes.ComponentStatus{
				Name:    component,
				Healthy: false,
				Message: fmt.Sprintf("%s has unhealthy pods", component),
			}
		} else {
			components[component] = kubernetes.ComponentStatus{
				Name:    component,
				Healthy: false,
				Message: fmt.Sprintf("%s not found", component),
			}
		}
	}

	return components, nil
}

// Helper function
func containsPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// parseResourceValue converts resource strings to float64
func parseResourceValue(value string) float64 {
	quantity, err := resource.ParseQuantity(value)
	if err != nil {
		return 0
	}
	
	// For CPU resources, return in millicores
	if strings.HasSuffix(value, "m") {
		// Already in millicores
		return float64(quantity.MilliValue())
	} else if strings.HasSuffix(value, "Mi") {
		// Memory in MiB - convert to KiB: 1 MiB = 1024 KiB
		return float64(quantity.Value()) / 1024
	} else if strings.HasSuffix(value, "Gi") {
		// Memory in GiB - convert to KiB: 1 GiB = 1024*1024 KiB
		return float64(quantity.Value()) / 1024
	} else if strings.Contains(value, "i") || strings.Contains(value, "G") || strings.Contains(value, "M") || strings.Contains(value, "K") {
		// Other memory resources - return in kilobytes
		return float64(quantity.Value()) / 1024
	} else {
		// CPU cores - convert to millicores
		return float64(quantity.MilliValue())
	}
}