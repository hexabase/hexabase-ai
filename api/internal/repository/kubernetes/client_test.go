package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestKubernetesRepository_GetPodMetrics(t *testing.T) {
	ctx := context.Background()
	clientset := fake.NewSimpleClientset()
	repo := NewKubernetesRepository(clientset)

	t.Run("get pod metrics", func(t *testing.T) {
		metrics, err := repo.GetPodMetrics(ctx, "test-namespace")
		assert.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Len(t, metrics.Items, 1)
		assert.Equal(t, "app-pod-1", metrics.Items[0].Name)
		assert.Equal(t, "100m", metrics.Items[0].CPU.Value)
		assert.Equal(t, "256Mi", metrics.Items[0].Memory.Value)
	})
}

func TestKubernetesRepository_GetNodeMetrics(t *testing.T) {
	ctx := context.Background()
	clientset := fake.NewSimpleClientset()
	repo := NewKubernetesRepository(clientset)

	t.Run("get node metrics", func(t *testing.T) {
		metrics, err := repo.GetNodeMetrics(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Len(t, metrics.Items, 1)
		assert.Equal(t, "node-1", metrics.Items[0].Name)
		assert.Equal(t, "2000m", metrics.Items[0].CPU.Value)
		assert.Equal(t, "8Gi", metrics.Items[0].Memory.Value)
		assert.Equal(t, "100Gi", metrics.Items[0].Storage.Value)
	})
}

func TestKubernetesRepository_GetNamespaceResourceQuota(t *testing.T) {
	ctx := context.Background()
	namespace := "test-namespace"

	t.Run("get resource quota from existing quota", func(t *testing.T) {
		// Create test namespace with resource quota
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}

		quota := &corev1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-quota",
				Namespace: namespace,
			},
			Spec: corev1.ResourceQuotaSpec{
				Hard: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10"),
					corev1.ResourceMemory: resource.MustParse("10Gi"),
					corev1.ResourcePods:   resource.MustParse("100"),
				},
			},
			Status: corev1.ResourceQuotaStatus{
				Hard: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("10"),
					corev1.ResourceMemory: resource.MustParse("10Gi"),
					corev1.ResourcePods:   resource.MustParse("100"),
				},
				Used: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
					corev1.ResourcePods:   resource.MustParse("10"),
				},
			},
		}

		clientset := fake.NewSimpleClientset(ns, quota)
		repo := NewKubernetesRepository(clientset)

		result, err := repo.GetNamespaceResourceQuota(ctx, namespace)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-quota", result.Name)
		assert.Equal(t, namespace, result.Namespace)
		assert.Equal(t, "10", result.Hard["cpu"])
		assert.Equal(t, "10Gi", result.Hard["memory"])
		assert.Equal(t, "100", result.Hard["pods"])
		assert.Equal(t, "2", result.Used["cpu"])
		assert.Equal(t, "2Gi", result.Used["memory"])
		assert.Equal(t, "10", result.Used["pods"])
	})

	t.Run("get default quota when none exists", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		repo := NewKubernetesRepository(clientset)

		result, err := repo.GetNamespaceResourceQuota(ctx, namespace)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "default", result.Name)
		assert.Equal(t, namespace, result.Namespace)
		// Should return default values
		assert.Equal(t, "2", result.Hard["cpu"])
		assert.Equal(t, "4Gi", result.Hard["memory"])
		assert.Equal(t, "50", result.Hard["pods"])
	})
}

func TestKubernetesRepository_GetClusterInfo(t *testing.T) {
	ctx := context.Background()

	t.Run("get cluster info", func(t *testing.T) {
		// Create test resources
		nodes := []corev1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "node-2"},
			},
		}

		pods := []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-1", Namespace: "default"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-2", Namespace: "kube-system"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-3", Namespace: "test"},
			},
		}

		namespaces := []corev1.Namespace{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "default"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "kube-system"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
			},
		}

		clientset := fake.NewSimpleClientset()
		
		// Add nodes
		for _, node := range nodes {
			_, err := clientset.CoreV1().Nodes().Create(ctx, &node, metav1.CreateOptions{})
			assert.NoError(t, err)
		}
		
		// Add pods
		for _, pod := range pods {
			_, err := clientset.CoreV1().Pods(pod.Namespace).Create(ctx, &pod, metav1.CreateOptions{})
			assert.NoError(t, err)
		}
		
		// Add namespaces
		for _, ns := range namespaces {
			_, err := clientset.CoreV1().Namespaces().Create(ctx, &ns, metav1.CreateOptions{})
			assert.NoError(t, err)
		}

		repo := NewKubernetesRepository(clientset)

		info, err := repo.GetClusterInfo(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, 2, info.NodeCount)
		assert.Equal(t, 3, info.PodCount)
		assert.Equal(t, 3, info.NamespaceCount)
		assert.NotEmpty(t, info.Version)
	})
}

func TestKubernetesRepository_CheckComponentHealth(t *testing.T) {
	ctx := context.Background()

	t.Run("check component health", func(t *testing.T) {
		// Create system pods
		systemPods := []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kube-apiserver-master",
					Namespace: "kube-system",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "etcd-master",
					Namespace: "kube-system",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kube-controller-manager-master",
					Namespace: "kube-system",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kube-scheduler-master",
					Namespace: "kube-system",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "coredns-1234",
					Namespace: "kube-system",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
		}

		clientset := fake.NewSimpleClientset()
		
		// Create namespace first
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kube-system",
			},
		}
		_, err := clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		assert.NoError(t, err)
		
		// Add system pods
		for _, pod := range systemPods {
			_, err := clientset.CoreV1().Pods("kube-system").Create(ctx, &pod, metav1.CreateOptions{})
			assert.NoError(t, err)
		}

		repo := NewKubernetesRepository(clientset)

		health, err := repo.CheckComponentHealth(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, health)
		
		// Check API server
		assert.True(t, health["api-server"].Healthy)
		assert.Equal(t, "api-server", health["api-server"].Name)
		
		// Check other components
		expectedComponents := []string{"etcd", "controller-manager", "scheduler", "dns"}
		for _, component := range expectedComponents {
			status, exists := health[component]
			assert.True(t, exists, "Component %s should exist", component)
			assert.True(t, status.Healthy, "Component %s should be healthy", component)
			assert.Equal(t, component, status.Name)
		}
	})

	t.Run("check component health with unhealthy component", func(t *testing.T) {
		// Create system pods with one unhealthy
		systemPods := []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "etcd-master",
					Namespace: "kube-system",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed, // Unhealthy
				},
			},
		}

		clientset := fake.NewSimpleClientset()
		
		// Create namespace first
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kube-system",
			},
		}
		_, err := clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		assert.NoError(t, err)
		
		// Add system pods
		for _, pod := range systemPods {
			_, err := clientset.CoreV1().Pods("kube-system").Create(ctx, &pod, metav1.CreateOptions{})
			assert.NoError(t, err)
		}

		repo := NewKubernetesRepository(clientset)

		health, err := repo.CheckComponentHealth(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, health)
		
		// Check etcd is unhealthy
		assert.False(t, health["etcd"].Healthy)
		assert.Contains(t, health["etcd"].Message, "unhealthy")
	})
}

// Test helper functions
func TestContainsPrefix(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		prefix   string
		expected bool
	}{
		{
			name:     "contains prefix",
			s:        "kube-apiserver-master",
			prefix:   "kube-apiserver",
			expected: true,
		},
		{
			name:     "does not contain prefix",
			s:        "etcd-master",
			prefix:   "kube-apiserver",
			expected: false,
		},
		{
			name:     "empty string",
			s:        "",
			prefix:   "kube",
			expected: false,
		},
		{
			name:     "prefix longer than string",
			s:        "kube",
			prefix:   "kube-apiserver",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsPrefix(tt.s, tt.prefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseResourceValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected float64
	}{
		{
			name:     "parse CPU millicores",
			value:    "100m",
			expected: 100,
		},
		{
			name:     "parse CPU cores",
			value:    "2",
			expected: 2000,
		},
		{
			name:     "parse memory Mi",
			value:    "256Mi",
			expected: 268435.456,
		},
		{
			name:     "parse memory Gi", 
			value:    "4Gi",
			expected: 4294967.296,
		},
		{
			name:     "invalid value",
			value:    "invalid",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseResourceValue(tt.value)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}