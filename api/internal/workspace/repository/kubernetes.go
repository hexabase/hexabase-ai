package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/workspace/domain"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type kubernetesRepository struct {
	clientset     kubernetes.Interface
	dynamicClient dynamic.Interface
	config        *rest.Config
}

// NewKubernetesRepository creates a new Kubernetes workspace repository
func NewKubernetesRepository(clientset kubernetes.Interface, dynamicClient dynamic.Interface, config *rest.Config) domain.KubernetesRepository {
	return &kubernetesRepository{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		config:        config,
	}
}

func (r *kubernetesRepository) CreateVCluster(ctx context.Context, workspaceID string, plan string) error {
	// Define vCluster resource
	vclusterGVR := schema.GroupVersionResource{
		Group:    "cluster.loft.sh",
		Version:  "v1alpha1",
		Resource: "virtualclusters",
	}

	// Create vCluster manifest
	vcluster := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.loft.sh/v1alpha1",
			"kind":       "VirtualCluster",
			"metadata": map[string]interface{}{
				"name":      workspaceID,
				"namespace": "hexabase-vclusters",
				"labels": map[string]interface{}{
					"hexabase.ai/workspace-id": workspaceID,
					"hexabase.ai/plan":         plan,
				},
			},
			"spec": map[string]interface{}{
				"helmRelease": map[string]interface{}{
					"values": getVClusterValues(plan),
				},
			},
		},
	}

	// Create vCluster
	_, err := r.dynamicClient.Resource(vclusterGVR).Namespace("hexabase-vclusters").Create(ctx, vcluster, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create vCluster: %w", err)
	}

	return nil
}

func (r *kubernetesRepository) DeleteVCluster(ctx context.Context, workspaceID string) error {
	vclusterGVR := schema.GroupVersionResource{
		Group:    "cluster.loft.sh",
		Version:  "v1alpha1",
		Resource: "virtualclusters",
	}

	err := r.dynamicClient.Resource(vclusterGVR).Namespace("hexabase-vclusters").Delete(ctx, workspaceID, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete vCluster: %w", err)
	}

	return nil
}

func (r *kubernetesRepository) GetVClusterStatus(ctx context.Context, workspaceID string) (string, error) {
	vclusterGVR := schema.GroupVersionResource{
		Group:    "cluster.loft.sh",
		Version:  "v1alpha1",
		Resource: "virtualclusters",
	}

	vcluster, err := r.dynamicClient.Resource(vclusterGVR).Namespace("hexabase-vclusters").Get(ctx, workspaceID, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get vCluster: %w", err)
	}

	// Extract status
	status, ok := vcluster.Object["status"].(map[string]interface{})
	if !ok {
		return "unknown", nil
	}

	if phase, ok := status["phase"].(string); ok {
		return phase, nil
	}

	if ready, ok := status["ready"].(bool); ok && ready {
		return "ready", nil
	}

	return "pending", nil
}

func (r *kubernetesRepository) GetVClusterKubeconfig(ctx context.Context, workspaceID string) (string, error) {
	// Get vCluster secret containing kubeconfig
	secretName := fmt.Sprintf("vc-%s-kubeconfig", workspaceID)
	secret, err := r.clientset.CoreV1().Secrets("hexabase-vclusters").Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig secret: %w", err)
	}

	kubeconfig, ok := secret.Data["config"]
	if !ok {
		return "", fmt.Errorf("kubeconfig not found in secret")
	}

	return string(kubeconfig), nil
}

func (r *kubernetesRepository) ScaleVCluster(ctx context.Context, workspaceID string, replicas int) error {
	// Scale vCluster statefulset
	statefulsetName := fmt.Sprintf("%s-vcluster", workspaceID)
	statefulset, err := r.clientset.AppsV1().StatefulSets("hexabase-vclusters").Get(ctx, statefulsetName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get statefulset: %w", err)
	}

	replicas32 := int32(replicas)
	statefulset.Spec.Replicas = &replicas32
	_, err = r.clientset.AppsV1().StatefulSets("hexabase-vclusters").Update(ctx, statefulset, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale statefulset: %w", err)
	}

	return nil
}

func (r *kubernetesRepository) WaitForVClusterReady(ctx context.Context, workspaceID string) error {
	timeout := 5 * time.Minute
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		status, err := r.GetVClusterStatus(ctx, workspaceID)
		if err != nil {
			return err
		}

		if status == "ready" {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			// Continue checking
		}
	}

	return fmt.Errorf("vCluster did not become ready within %v", timeout)
}

func (r *kubernetesRepository) WaitForVClusterDeleted(ctx context.Context, workspaceID string) error {
	timeout := 5 * time.Minute
	deadline := time.Now().Add(timeout)

	vclusterGVR := schema.GroupVersionResource{
		Group:    "cluster.loft.sh",
		Version:  "v1alpha1",
		Resource: "virtualclusters",
	}

	for time.Now().Before(deadline) {
		_, err := r.dynamicClient.Resource(vclusterGVR).Namespace("hexabase-vclusters").Get(ctx, workspaceID, metav1.GetOptions{})
		if err != nil {
			// vCluster not found, deletion complete
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			// Continue checking
		}
	}

	return fmt.Errorf("vCluster was not deleted within %v", timeout)
}

func (r *kubernetesRepository) ConfigureOIDC(ctx context.Context, workspaceID string) error {
	// Get vCluster kubeconfig
	kubeconfig, err := r.GetVClusterKubeconfig(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Create client for vCluster
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	vClusterClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create vCluster client: %w", err)
	}

	// Configure OIDC in vCluster
	// This would typically involve:
	// 1. Creating OIDC provider configuration
	// 2. Setting up RBAC rules
	// 3. Configuring API server flags

	// For now, create a ConfigMap with OIDC settings
	oidcConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "oidc-config",
			Namespace: "kube-system",
		},
		Data: map[string]string{
			"issuer-url":     "https://api.hexabase-kaas.io",
			"client-id":      workspaceID,
			"username-claim": "sub",
			"groups-claim":   "groups",
		},
	}

	_, err = vClusterClient.CoreV1().ConfigMaps("kube-system").Create(ctx, oidcConfig, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create OIDC config: %w", err)
	}

	return nil
}

func (r *kubernetesRepository) UpdateOIDCConfig(ctx context.Context, workspaceID string, config map[string]interface{}) error {
	// Update OIDC configuration when members change
	// This would involve updating RBAC rules in the vCluster
	// For now, just call ConfigureOIDC
	return r.ConfigureOIDC(ctx, workspaceID)
}

func (r *kubernetesRepository) ApplyResourceQuotas(ctx context.Context, workspaceID string, plan string) error {
	// Get vCluster kubeconfig
	kubeconfig, err := r.GetVClusterKubeconfig(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Create client for vCluster
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	vClusterClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create vCluster client: %w", err)
	}

	// Get quota limits based on plan
	limits := getPlanLimits(plan)

	// Create ResourceQuota for default namespace
	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-quota",
			Namespace: "default",
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse(limits.CPU),
				corev1.ResourceMemory:           resource.MustParse(limits.Memory),
				corev1.ResourceStorage:          resource.MustParse(limits.Storage),
				corev1.ResourcePods:             resource.MustParse(limits.Pods),
				corev1.ResourceServices:         resource.MustParse(limits.Services),
				corev1.ResourcePersistentVolumeClaims: resource.MustParse(limits.PVCs),
			},
		},
	}

	_, err = vClusterClient.CoreV1().ResourceQuotas("default").Create(ctx, quota, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create resource quota: %w", err)
	}

	return nil
}

func (r *kubernetesRepository) GetResourceMetrics(ctx context.Context, workspaceID string) (*domain.ResourceUsage, error) {
	// Get vCluster metrics
	// This would typically query metrics-server or Prometheus
	
	// For now, return mock data
	usage := &domain.ResourceUsage{
		CPU: domain.ResourceMetric{
			Used:      0.5,
			Requested: 1.0,
			Limit:     2.0,
			Unit:      "cores",
		},
		Memory: domain.ResourceMetric{
			Used:      1024,
			Requested: 2048,
			Limit:     4096,
			Unit:      "MB",
		},
		Storage: domain.ResourceMetric{
			Used:      5120,
			Requested: 10240,
			Limit:     20480,
			Unit:      "MB",
		},
		Pods: domain.PodMetric{
			Running: 5,
			Pending: 2,
			Failed:  1,
			Total:   8,
		},
		WorkspaceID: workspaceID,
		Timestamp:   time.Now(),
	}

	return usage, nil
}

func (r *kubernetesRepository) GetVClusterInfo(ctx context.Context, workspaceID string) (*domain.ClusterInfo, error) {
	vclusterGVR := schema.GroupVersionResource{
		Group:    "cluster.loft.sh",
		Version:  "v1alpha1",
		Resource: "virtualclusters",
	}
	
	// Get vCluster information
	vcluster, err := r.dynamicClient.Resource(vclusterGVR).Namespace("hexabase-vclusters").Get(ctx, workspaceID, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get vCluster: %w", err)
	}

	// Extract status
	status, ok := vcluster.Object["status"].(map[string]interface{})
	if !ok {
		return &domain.ClusterInfo{
			Status: "unknown",
		}, nil
	}

	info := &domain.ClusterInfo{
		Status: "unknown",
	}

	if phase, ok := status["phase"].(string); ok {
		info.Status = phase
	}

	if ready, ok := status["ready"].(bool); ok && ready {
		info.Status = "ready"
	}

	// Get endpoint from service
	svc, err := r.clientset.CoreV1().Services("hexabase-vclusters").Get(ctx, workspaceID, metav1.GetOptions{})
	if err == nil {
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			info.Endpoint = fmt.Sprintf("https://%s:443", svc.Status.LoadBalancer.Ingress[0].Hostname)
			info.APIServer = info.Endpoint
		}
	}

	// Get kubeconfig
	kubeconfig, err := r.GetVClusterKubeconfig(ctx, workspaceID)
	if err == nil {
		info.KubeConfig = kubeconfig
	}

	return info, nil
}

// Helper functions

func getVClusterValues(plan string) map[string]interface{} {
	// Return Helm values based on plan
	baseValues := map[string]interface{}{
		"syncer": map[string]interface{}{
			"extraArgs": []string{
				"--enable-storage-classes",
				"--sync-all-nodes",
			},
		},
		"isolation": map[string]interface{}{
			"enabled": true,
		},
	}

	// Adjust resources based on plan
	switch plan {
	case "starter":
		baseValues["syncer"].(map[string]interface{})["resources"] = map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    "500m",
				"memory": "512Mi",
			},
		}
	case "professional":
		baseValues["syncer"].(map[string]interface{})["resources"] = map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    "1",
				"memory": "1Gi",
			},
		}
	case "enterprise":
		baseValues["syncer"].(map[string]interface{})["resources"] = map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    "2",
				"memory": "2Gi",
			},
		}
	}

	return baseValues
}

type planLimits struct {
	CPU      string
	Memory   string
	Storage  string
	Pods     string
	Services string
	PVCs     string
}

func getPlanLimits(plan string) planLimits {
	switch plan {
	case "starter":
		return planLimits{
			CPU:      "2",
			Memory:   "4Gi",
			Storage:  "20Gi",
			Pods:     "20",
			Services: "10",
			PVCs:     "5",
		}
	case "professional":
		return planLimits{
			CPU:      "8",
			Memory:   "16Gi",
			Storage:  "100Gi",
			Pods:     "100",
			Services: "50",
			PVCs:     "20",
		}
	case "enterprise":
		return planLimits{
			CPU:      "32",
			Memory:   "64Gi",
			Storage:  "500Gi",
			Pods:     "500",
			Services: "200",
			PVCs:     "100",
		}
	default:
		// Default to starter limits
		return planLimits{
			CPU:      "2",
			Memory:   "4Gi",
			Storage:  "20Gi",
			Pods:     "20",
			Services: "10",
			PVCs:     "5",
		}
	}
}

// ListVClusterNodes lists nodes in the vCluster
func (r *kubernetesRepository) ListVClusterNodes(ctx context.Context, workspaceID string) ([]domain.Node, error) {
	// Get vCluster kubeconfig
	kubeconfig, err := r.GetVClusterKubeconfig(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Create client for vCluster
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	vClusterClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vCluster client: %w", err)
	}

	// List nodes in vCluster
	nodeList, err := vClusterClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodes := make([]domain.Node, 0, len(nodeList.Items))
	for _, node := range nodeList.Items {
		nodes = append(nodes, domain.Node{
			Name:   node.Name,
			Status: string(node.Status.Phase),
		})
	}

	return nodes, nil
}

// ScaleVClusterDeployment scales a deployment in the vCluster
func (r *kubernetesRepository) ScaleVClusterDeployment(ctx context.Context, workspaceID, deploymentName string, replicas int) error {
	// Get vCluster kubeconfig
	kubeconfig, err := r.GetVClusterKubeconfig(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Create client for vCluster
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	vClusterClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create vCluster client: %w", err)
	}

	// Scale deployment
	scale := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: "default",
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: int32(replicas),
		},
	}

	_, err = vClusterClient.AppsV1().Deployments("default").UpdateScale(ctx, deploymentName, scale, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	return nil
}