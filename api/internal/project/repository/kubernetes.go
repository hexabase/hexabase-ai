package repository

import (
	"context"
	"fmt"

	"github.com/hexabase/hexabase-ai/api/internal/project/domain"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

// NewKubernetesRepository creates a new Kubernetes project repository
func NewKubernetesRepository(clientset kubernetes.Interface, dynamicClient dynamic.Interface, config *rest.Config) domain.KubernetesRepository {
	return &kubernetesRepository{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		config:        config,
	}
}

func (r *kubernetesRepository) CreateNamespace(ctx context.Context, workspaceID string, namespaceName string, labels map[string]string) error {
	// Get vCluster client
	vClusterClient, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Create namespace
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespaceName,
			Labels: labels,
		},
	}

	_, err = vClusterClient.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	return nil
}

func (r *kubernetesRepository) DeleteNamespace(ctx context.Context, workspaceID, namespaceName string) error {
	// Get vCluster client
	vClusterClient, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return err
	}

	err = vClusterClient.CoreV1().Namespaces().Delete(ctx, namespaceName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	return nil
}

func (r *kubernetesRepository) GetNamespace(ctx context.Context, workspaceID, namespaceName string) (map[string]interface{}, error) {
	// Get vCluster client
	vClusterClient, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	ns, err := vClusterClient.CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	namespace := map[string]interface{}{
		"name":   ns.Name,
		"labels": ns.Labels,
		"status": string(ns.Status.Phase),
		"creationTimestamp": ns.CreationTimestamp.Time,
		"uid": string(ns.UID),
	}

	return namespace, nil
}

func (r *kubernetesRepository) ListNamespaces(ctx context.Context, workspaceID string) ([]string, error) {
	// Get vCluster client
	vClusterClient, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	namespaces, err := vClusterClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var names []string
	for _, ns := range namespaces.Items {
		names = append(names, ns.Name)
	}

	return names, nil
}

func (r *kubernetesRepository) ConfigureHNC(ctx context.Context, workspaceID, parentNamespace, childNamespace string) error {
	// This implementation was moved down in the file
	// Get vCluster client with dynamic client
	vClusterDynamic, err := r.getVClusterDynamicClient(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Create HNC SubnamespaceAnchor
	hncGVR := schema.GroupVersionResource{
		Group:    "hnc.x-k8s.io",
		Version:  "v1alpha2",
		Resource: "subnamespaceanchors",
	}

	anchor := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "hnc.x-k8s.io/v1alpha2",
			"kind":       "SubnamespaceAnchor",
			"metadata": map[string]interface{}{
				"name":      childNamespace,
				"namespace": parentNamespace,
			},
		},
	}

	_, err = vClusterDynamic.Resource(hncGVR).Namespace(parentNamespace).Create(ctx, anchor, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create HNC anchor: %w", err)
	}

	return nil
}

func (r *kubernetesRepository) ApplyResourceQuota(ctx context.Context, workspaceID, namespaceName string, quota *domain.ResourceQuota) error {
	// Get vCluster client
	vClusterClient, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Create resource quota
	resourceQuota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "project-quota",
			Namespace: namespaceName,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{},
		},
	}

	// Set limits
	if quota.CPU != "" {
		resourceQuota.Spec.Hard[corev1.ResourceCPU] = resource.MustParse(quota.CPU)
	}
	if quota.Memory != "" {
		resourceQuota.Spec.Hard[corev1.ResourceMemory] = resource.MustParse(quota.Memory)
	}
	if quota.Storage != "" {
		resourceQuota.Spec.Hard[corev1.ResourceStorage] = resource.MustParse(quota.Storage)
	}
	if quota.Pods > 0 {
		resourceQuota.Spec.Hard[corev1.ResourcePods] = *resource.NewQuantity(int64(quota.Pods), resource.DecimalSI)
	}
	if quota.Services > 0 {
		resourceQuota.Spec.Hard[corev1.ResourceServices] = *resource.NewQuantity(int64(quota.Services), resource.DecimalSI)
	}
	if quota.PersistentVolumeClaims > 0 {
		resourceQuota.Spec.Hard[corev1.ResourcePersistentVolumeClaims] = *resource.NewQuantity(int64(quota.PersistentVolumeClaims), resource.DecimalSI)
	}

	// Check if quota exists
	existing, err := vClusterClient.CoreV1().ResourceQuotas(namespaceName).Get(ctx, "project-quota", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create new quota
			_, err = vClusterClient.CoreV1().ResourceQuotas(namespaceName).Create(ctx, resourceQuota, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create resource quota: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get resource quota: %w", err)
		}
	} else {
		// Update existing quota
		existing.Spec = resourceQuota.Spec
		_, err = vClusterClient.CoreV1().ResourceQuotas(namespaceName).Update(ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update resource quota: %w", err)
		}
	}

	return nil
}

func (r *kubernetesRepository) GetNamespaceResourceUsage(ctx context.Context, workspaceID, namespaceName string) (*domain.ResourceUsage, error) {
	// Get vCluster client
	vClusterClient, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Get resource quota status
	quota, err := vClusterClient.CoreV1().ResourceQuotas(namespaceName).Get(ctx, "project-quota", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// No quota, return empty usage
			return &domain.ResourceUsage{
				CPU:    "0",
				Memory: "0",
				Pods:   0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get resource quota: %w", err)
	}

	// Extract usage from quota status
	cpuUsed := "0"
	memoryUsed := "0"
	podsUsed := 0

	if used, ok := quota.Status.Used[corev1.ResourceCPU]; ok {
		cpuUsed = used.String()
	}
	if used, ok := quota.Status.Used[corev1.ResourceMemory]; ok {
		memoryUsed = used.String()
	}
	if used, ok := quota.Status.Used[corev1.ResourcePods]; ok {
		podsUsed = int(used.Value())
	}

	return &domain.ResourceUsage{
		CPU:    cpuUsed,
		Memory: memoryUsed,
		Pods:   podsUsed,
	}, nil
}

func (r *kubernetesRepository) ApplyRBAC(ctx context.Context, workspaceID, namespaceName, userID, role string) error {
	// Get vCluster client
	vClusterClient, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Determine Kubernetes role based on project role
	k8sRole := "view" // Default to read-only
	switch role {
	case "admin":
		k8sRole = "admin"
	case "editor":
		k8sRole = "edit"
	case "viewer":
		k8sRole = "view"
	}

	// Create RoleBinding
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("user-%s", userID),
			Namespace: namespaceName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "User",
				Name: userID,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: k8sRole,
		},
	}

	// Check if role binding exists
	existing, err := vClusterClient.RbacV1().RoleBindings(namespaceName).Get(ctx, roleBinding.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create new role binding
			_, err = vClusterClient.RbacV1().RoleBindings(namespaceName).Create(ctx, roleBinding, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create role binding: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get role binding: %w", err)
		}
	} else {
		// Update existing role binding
		existing.RoleRef = roleBinding.RoleRef
		_, err = vClusterClient.RbacV1().RoleBindings(namespaceName).Update(ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update role binding: %w", err)
		}
	}

	return nil
}

func (r *kubernetesRepository) RemoveRBAC(ctx context.Context, workspaceID, namespaceName, userID string) error {
	// Get vCluster client
	vClusterClient, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Delete RoleBinding
	roleBindingName := fmt.Sprintf("user-%s", userID)
	err = vClusterClient.RbacV1().RoleBindings(namespaceName).Delete(ctx, roleBindingName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete role binding: %w", err)
	}

	return nil
}

func (r *kubernetesRepository) CreateResourceQuota(ctx context.Context, workspaceID, namespace string, quota *domain.ResourceQuota) error {
	// Get vCluster client
	vClusterClient, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Create resource quota
	resourceQuota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "project-quota",
			Namespace: namespace,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceCPU:                        resource.MustParse(quota.CPU),
				corev1.ResourceMemory:                     resource.MustParse(quota.Memory),
				corev1.ResourceStorage:                    resource.MustParse(quota.Storage),
				corev1.ResourcePods:                       resource.MustParse(fmt.Sprintf("%d", quota.Pods)),
				corev1.ResourceServices:                   resource.MustParse(fmt.Sprintf("%d", quota.Services)),
				corev1.ResourcePersistentVolumeClaims:     resource.MustParse(fmt.Sprintf("%d", quota.PersistentVolumeClaims)),
			},
		},
	}

	_, err = vClusterClient.CoreV1().ResourceQuotas(namespace).Create(ctx, resourceQuota, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create resource quota: %w", err)
	}

	return nil
}

func (r *kubernetesRepository) UpdateResourceQuota(ctx context.Context, workspaceID, namespace string, quota *domain.ResourceQuota) error {
	// Get vCluster client
	vClusterClient, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Get existing quota
	existing, err := vClusterClient.CoreV1().ResourceQuotas(namespace).Get(ctx, "project-quota", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create new quota
			return r.CreateResourceQuota(ctx, workspaceID, namespace, quota)
		}
		return fmt.Errorf("failed to get resource quota: %w", err)
	}

	// Update quota
	existing.Spec.Hard = corev1.ResourceList{
		corev1.ResourceCPU:                        resource.MustParse(quota.CPU),
		corev1.ResourceMemory:                     resource.MustParse(quota.Memory),
		corev1.ResourceStorage:                    resource.MustParse(quota.Storage),
		corev1.ResourcePods:                       resource.MustParse(fmt.Sprintf("%d", quota.Pods)),
		corev1.ResourceServices:                   resource.MustParse(fmt.Sprintf("%d", quota.Services)),
		corev1.ResourcePersistentVolumeClaims:     resource.MustParse(fmt.Sprintf("%d", quota.PersistentVolumeClaims)),
	}

	_, err = vClusterClient.CoreV1().ResourceQuotas(namespace).Update(ctx, existing, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update resource quota: %w", err)
	}

	return nil
}

func (r *kubernetesRepository) GetResourceQuota(ctx context.Context, workspaceID, namespace string) (*domain.ResourceQuota, error) {
	// Get vCluster client
	vClusterClient, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Get resource quota
	resourceQuota, err := vClusterClient.CoreV1().ResourceQuotas(namespace).Get(ctx, "project-quota", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get resource quota: %w", err)
	}

	// Convert to domain model
	cpuQuantity := resourceQuota.Spec.Hard[corev1.ResourceCPU]
	memoryQuantity := resourceQuota.Spec.Hard[corev1.ResourceMemory]
	storageQuantity := resourceQuota.Spec.Hard[corev1.ResourceStorage]
	podsQuantity := resourceQuota.Spec.Hard[corev1.ResourcePods]
	servicesQuantity := resourceQuota.Spec.Hard[corev1.ResourceServices]
	pvcsQuantity := resourceQuota.Spec.Hard[corev1.ResourcePersistentVolumeClaims]
	
	quota := &domain.ResourceQuota{
		CPU:                    cpuQuantity.String(),
		Memory:                 memoryQuantity.String(),
		Storage:                storageQuantity.String(),
		Pods:                   int(podsQuantity.Value()),
		Services:               int(servicesQuantity.Value()),
		PersistentVolumeClaims: int(pvcsQuantity.Value()),
	}

	return quota, nil
}

func (r *kubernetesRepository) DeleteResourceQuota(ctx context.Context, workspaceID, namespace string) error {
	// Get vCluster client
	vClusterClient, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return err
	}

	err = vClusterClient.CoreV1().ResourceQuotas(namespace).Delete(ctx, "project-quota", metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete resource quota: %w", err)
	}

	return nil
}

func (r *kubernetesRepository) GetNamespaceUsage(ctx context.Context, workspaceID, namespaceName string) (*domain.NamespaceUsage, error) {
	// Get vCluster client
	vClusterClient, err := r.getVClusterClient(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Get resource quota status
	quota, err := vClusterClient.CoreV1().ResourceQuotas(namespaceName).Get(ctx, "project-quota", metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// No quota, return empty usage
			return &domain.NamespaceUsage{
				CPU:     "0",
				Memory:  "0",
				Storage: "0",
				Pods:    0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get resource quota: %w", err)
	}

	// Extract usage from quota status
	cpuUsed := "0"
	memoryUsed := "0"
	storageUsed := "0"
	podsUsed := 0

	if used, ok := quota.Status.Used[corev1.ResourceCPU]; ok {
		cpuUsed = used.String()
	}
	if used, ok := quota.Status.Used[corev1.ResourceMemory]; ok {
		memoryUsed = used.String()
	}
	if used, ok := quota.Status.Used[corev1.ResourceStorage]; ok {
		storageUsed = used.String()
	}
	if used, ok := quota.Status.Used[corev1.ResourcePods]; ok {
		podsUsed = int(used.Value())
	}

	return &domain.NamespaceUsage{
		CPU:     cpuUsed,
		Memory:  memoryUsed,
		Storage: storageUsed,
		Pods:    podsUsed,
	}, nil
}

// Helper functions

func (r *kubernetesRepository) getVClusterClient(ctx context.Context, workspaceID string) (kubernetes.Interface, error) {
	// Get vCluster kubeconfig
	kubeconfig, err := r.getVClusterKubeconfig(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Create client config
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	// Create client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vCluster client: %w", err)
	}

	return client, nil
}

func (r *kubernetesRepository) getVClusterDynamicClient(ctx context.Context, workspaceID string) (dynamic.Interface, error) {
	// Get vCluster kubeconfig
	kubeconfig, err := r.getVClusterKubeconfig(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Create client config
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	// Create dynamic client
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vCluster dynamic client: %w", err)
	}

	return client, nil
}

func (r *kubernetesRepository) getVClusterKubeconfig(ctx context.Context, workspaceID string) (string, error) {
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