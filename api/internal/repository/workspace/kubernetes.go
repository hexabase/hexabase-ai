package workspace

import (
	"context"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-kaas/api/internal/domain/workspace"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/core/v1"
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
func NewKubernetesRepository(clientset kubernetes.Interface, dynamicClient dynamic.Interface, config *rest.Config) workspace.KubernetesRepository {
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
					"hexabase.io/workspace-id": workspaceID,
					"hexabase.io/plan":         plan,
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

func (r *kubernetesRepository) GetVClusterStatus(ctx context.Context, workspaceID string) (*workspace.ClusterInfo, error) {
	vclusterGVR := schema.GroupVersionResource{
		Group:    "cluster.loft.sh",
		Version:  "v1alpha1",
		Resource: "virtualclusters",
	}

	vcluster, err := r.dynamicClient.Resource(vclusterGVR).Namespace("hexabase-vclusters").Get(ctx, workspaceID, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get vCluster: %w", err)
	}

	// Extract status
	status, ok := vcluster.Object["status"].(map[string]interface{})
	if !ok {
		return &workspace.ClusterInfo{
			Status: "unknown",
		}, nil
	}

	info := &workspace.ClusterInfo{
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
		}
	}

	return info, nil
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

func (r *kubernetesRepository) ScaleVCluster(ctx context.Context, workspaceID string, replicas int32) error {
	// Scale vCluster statefulset
	statefulsetName := fmt.Sprintf("%s-vcluster", workspaceID)
	statefulset, err := r.clientset.AppsV1().StatefulSets("hexabase-vclusters").Get(ctx, statefulsetName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get statefulset: %w", err)
	}

	statefulset.Spec.Replicas = &replicas
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
		info, err := r.GetVClusterStatus(ctx, workspaceID)
		if err != nil {
			return err
		}

		if info.Status == "ready" {
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

func (r *kubernetesRepository) UpdateOIDCConfig(ctx context.Context, workspaceID string) error {
	// Update OIDC configuration when members change
	// This would involve updating RBAC rules in the vCluster
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

func (r *kubernetesRepository) GetResourceMetrics(ctx context.Context, workspaceID string) (map[string]float64, error) {
	// Get vCluster metrics
	// This would typically query metrics-server or Prometheus
	
	// For now, return mock data
	metrics := map[string]float64{
		"cpu":     0.5,  // 0.5 cores
		"memory":  1024, // 1GB
		"storage": 5120, // 5GB
		"pods":    10,
	}

	return metrics, nil
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