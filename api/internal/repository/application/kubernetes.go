package application

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// KubernetesRepository implements the Kubernetes operations for applications
type KubernetesRepository struct {
	clientset        kubernetes.Interface
	metricsClientset versioned.Interface
}

// NewKubernetesRepository creates a new Kubernetes repository
func NewKubernetesRepository(clientset kubernetes.Interface, metricsClientset versioned.Interface) application.KubernetesRepository {
	return &KubernetesRepository{
		clientset:        clientset,
		metricsClientset: metricsClientset,
	}
}

// CreateDeployment creates a new deployment
func (r *KubernetesRepository) CreateDeployment(ctx context.Context, workspaceID, projectID string, spec application.DeploymentSpec) error {
	namespace := r.getNamespace(workspaceID, projectID)
	
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        spec.Name,
			Labels:      spec.Labels,
			Annotations: spec.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(int32(spec.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: spec.Labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: spec.Labels,
				},
				Spec: corev1.PodSpec{
					NodeSelector: spec.NodeSelector,
					Containers: []corev1.Container{
						{
							Name:  spec.Name,
							Image: spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(spec.Port),
								},
							},
							Env:       r.convertEnvVars(spec.EnvVars),
							Resources: r.convertResources(spec.Resources),
						},
					},
				},
			},
		},
	}

	_, err := r.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

// UpdateDeployment updates an existing deployment
func (r *KubernetesRepository) UpdateDeployment(ctx context.Context, workspaceID, projectID, name string, spec application.DeploymentSpec) error {
	namespace := r.getNamespace(workspaceID, projectID)
	
	deployment, err := r.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Update fields
	if spec.Replicas > 0 {
		deployment.Spec.Replicas = int32Ptr(int32(spec.Replicas))
	}
	if spec.Image != "" {
		deployment.Spec.Template.Spec.Containers[0].Image = spec.Image
	}

	_, err = r.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	return err
}

// DeleteDeployment deletes a deployment
func (r *KubernetesRepository) DeleteDeployment(ctx context.Context, workspaceID, projectID, name string) error {
	namespace := r.getNamespace(workspaceID, projectID)
	return r.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetDeploymentStatus gets the status of a deployment
func (r *KubernetesRepository) GetDeploymentStatus(ctx context.Context, workspaceID, projectID, name string) (*application.DeploymentStatus, error) {
	namespace := r.getNamespace(workspaceID, projectID)
	
	deployment, err := r.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	status := &application.DeploymentStatus{
		Replicas:          int(deployment.Status.Replicas),
		UpdatedReplicas:   int(deployment.Status.UpdatedReplicas),
		ReadyReplicas:     int(deployment.Status.ReadyReplicas),
		AvailableReplicas: int(deployment.Status.AvailableReplicas),
	}

	for _, cond := range deployment.Status.Conditions {
		status.Conditions = append(status.Conditions, application.DeploymentCondition{
			Type:               string(cond.Type),
			Status:             string(cond.Status),
			LastUpdateTime:     cond.LastUpdateTime.Time,
			LastTransitionTime: cond.LastTransitionTime.Time,
			Reason:             cond.Reason,
			Message:            cond.Message,
		})
	}

	return status, nil
}

// CreateStatefulSet creates a new statefulset
func (r *KubernetesRepository) CreateStatefulSet(ctx context.Context, workspaceID, projectID string, spec application.StatefulSetSpec) error {
	namespace := r.getNamespace(workspaceID, projectID)
	
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        spec.Name,
			Labels:      spec.Labels,
			Annotations: spec.Annotations,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32Ptr(int32(spec.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: spec.Labels,
			},
			ServiceName: spec.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: spec.Labels,
				},
				Spec: corev1.PodSpec{
					NodeSelector: spec.NodeSelector,
					Containers: []corev1.Container{
						{
							Name:  spec.Name,
							Image: spec.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(spec.Port),
								},
							},
							Env:       r.convertEnvVars(spec.EnvVars),
							Resources: r.convertResources(spec.Resources),
						},
					},
				},
			},
		},
	}

	// Add volume claim template if storage is configured
	if spec.VolumeClaimSpec.Name != "" {
		statefulSet.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: spec.VolumeClaimSpec.Name,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.PersistentVolumeAccessMode(spec.VolumeClaimSpec.AccessMode),
					},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse(spec.VolumeClaimSpec.Size),
						},
					},
					StorageClassName: &spec.VolumeClaimSpec.StorageClass,
				},
			},
		}
	}

	_, err := r.clientset.AppsV1().StatefulSets(namespace).Create(ctx, statefulSet, metav1.CreateOptions{})
	return err
}

// UpdateStatefulSet updates an existing statefulset
func (r *KubernetesRepository) UpdateStatefulSet(ctx context.Context, workspaceID, projectID, name string, spec application.StatefulSetSpec) error {
	namespace := r.getNamespace(workspaceID, projectID)
	
	statefulSet, err := r.clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Update fields
	if spec.Replicas > 0 {
		statefulSet.Spec.Replicas = int32Ptr(int32(spec.Replicas))
	}
	if spec.Image != "" {
		statefulSet.Spec.Template.Spec.Containers[0].Image = spec.Image
	}

	_, err = r.clientset.AppsV1().StatefulSets(namespace).Update(ctx, statefulSet, metav1.UpdateOptions{})
	return err
}

// DeleteStatefulSet deletes a statefulset
func (r *KubernetesRepository) DeleteStatefulSet(ctx context.Context, workspaceID, projectID, name string) error {
	namespace := r.getNamespace(workspaceID, projectID)
	return r.clientset.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetStatefulSetStatus gets the status of a statefulset
func (r *KubernetesRepository) GetStatefulSetStatus(ctx context.Context, workspaceID, projectID, name string) (*application.StatefulSetStatus, error) {
	namespace := r.getNamespace(workspaceID, projectID)
	
	statefulSet, err := r.clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	status := &application.StatefulSetStatus{
		Replicas:        int(statefulSet.Status.Replicas),
		ReadyReplicas:   int(statefulSet.Status.ReadyReplicas),
		CurrentReplicas: int(statefulSet.Status.CurrentReplicas),
		UpdatedReplicas: int(statefulSet.Status.UpdatedReplicas),
	}

	for _, cond := range statefulSet.Status.Conditions {
		status.Conditions = append(status.Conditions, application.StatefulSetCondition{
			Type:               string(cond.Type),
			Status:             string(cond.Status),
			LastTransitionTime: cond.LastTransitionTime.Time,
			Reason:             cond.Reason,
			Message:            cond.Message,
		})
	}

	return status, nil
}

// CreateService creates a new service
func (r *KubernetesRepository) CreateService(ctx context.Context, workspaceID, projectID string, spec application.ServiceSpec) error {
	namespace := r.getNamespace(workspaceID, projectID)
	
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: spec.Name,
		},
		Spec: corev1.ServiceSpec{
			Selector: spec.Selector,
			Type:     corev1.ServiceType(spec.Type),
			Ports: []corev1.ServicePort{
				{
					Port:       int32(spec.Port),
					TargetPort: intstr.FromInt(spec.TargetPort),
				},
			},
		},
	}

	_, err := r.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

// DeleteService deletes a service
func (r *KubernetesRepository) DeleteService(ctx context.Context, workspaceID, projectID, name string) error {
	namespace := r.getNamespace(workspaceID, projectID)
	return r.clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// GetServiceEndpoints gets the endpoints for a service
func (r *KubernetesRepository) GetServiceEndpoints(ctx context.Context, workspaceID, projectID, name string) ([]application.Endpoint, error) {
	namespace := r.getNamespace(workspaceID, projectID)
	
	service, err := r.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var endpoints []application.Endpoint
	
	// For ClusterIP services
	if service.Spec.Type == corev1.ServiceTypeClusterIP {
		endpoints = append(endpoints, application.Endpoint{
			Type: "service",
			URL:  fmt.Sprintf("%s.%s.svc.cluster.local:%d", service.Name, namespace, service.Spec.Ports[0].Port),
		})
	}

	// For LoadBalancer services
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer && len(service.Status.LoadBalancer.Ingress) > 0 {
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			if ingress.IP != "" {
				endpoints = append(endpoints, application.Endpoint{
					Type: "loadbalancer",
					URL:  fmt.Sprintf("http://%s:%d", ingress.IP, service.Spec.Ports[0].Port),
				})
			}
			if ingress.Hostname != "" {
				endpoints = append(endpoints, application.Endpoint{
					Type: "loadbalancer",
					URL:  fmt.Sprintf("http://%s:%d", ingress.Hostname, service.Spec.Ports[0].Port),
				})
			}
		}
	}

	// Check for ingress
	ingresses, err := r.clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, ingress := range ingresses.Items {
			for _, rule := range ingress.Spec.Rules {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil && path.Backend.Service.Name == name {
						protocol := "http"
						if len(ingress.Spec.TLS) > 0 {
							protocol = "https"
						}
						endpoints = append(endpoints, application.Endpoint{
							Type: "ingress",
							URL:  fmt.Sprintf("%s://%s%s", protocol, rule.Host, path.Path),
						})
					}
				}
			}
		}
	}

	return endpoints, nil
}

// CreateIngress creates a new ingress
func (r *KubernetesRepository) CreateIngress(ctx context.Context, workspaceID, projectID string, spec application.IngressSpec) error {
	namespace := r.getNamespace(workspaceID, projectID)
	
	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        spec.Name,
			Annotations: spec.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: spec.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     spec.Path,
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: spec.ServiceName,
											Port: networkingv1.ServiceBackendPort{
												Number: int32(spec.ServicePort),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if spec.TLSEnabled {
		ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts:      []string{spec.Host},
				SecretName: spec.Name + "-tls",
			},
		}
	}

	_, err := r.clientset.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	return err
}

// UpdateIngress updates an existing ingress
func (r *KubernetesRepository) UpdateIngress(ctx context.Context, workspaceID, projectID, name string, spec application.IngressSpec) error {
	namespace := r.getNamespace(workspaceID, projectID)
	
	ingress, err := r.clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Update annotations
	if spec.Annotations != nil {
		ingress.Annotations = spec.Annotations
	}

	// Update host if specified
	if spec.Host != "" && len(ingress.Spec.Rules) > 0 {
		ingress.Spec.Rules[0].Host = spec.Host
	}

	_, err = r.clientset.NetworkingV1().Ingresses(namespace).Update(ctx, ingress, metav1.UpdateOptions{})
	return err
}

// DeleteIngress deletes an ingress
func (r *KubernetesRepository) DeleteIngress(ctx context.Context, workspaceID, projectID, name string) error {
	namespace := r.getNamespace(workspaceID, projectID)
	return r.clientset.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// CreatePVC creates a new persistent volume claim
func (r *KubernetesRepository) CreatePVC(ctx context.Context, workspaceID, projectID string, spec application.PVCSpec) error {
	namespace := r.getNamespace(workspaceID, projectID)
	
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: spec.Name,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.PersistentVolumeAccessMode(spec.AccessMode),
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(spec.Size),
				},
			},
			StorageClassName: &spec.StorageClass,
		},
	}

	_, err := r.clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	return err
}

// DeletePVC deletes a persistent volume claim
func (r *KubernetesRepository) DeletePVC(ctx context.Context, workspaceID, projectID, name string) error {
	namespace := r.getNamespace(workspaceID, projectID)
	return r.clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ListPods lists pods matching the selector
func (r *KubernetesRepository) ListPods(ctx context.Context, workspaceID, projectID string, selector map[string]string) ([]application.Pod, error) {
	namespace := r.getNamespace(workspaceID, projectID)
	
	labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{
		MatchLabels: selector,
	})

	podList, err := r.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	var pods []application.Pod
	for _, p := range podList.Items {
		pod := application.Pod{
			Name:      p.Name,
			Status:    string(p.Status.Phase),
			NodeName:  p.Spec.NodeName,
			IP:        p.Status.PodIP,
			Restarts:  0,
		}

		if p.Status.StartTime != nil {
			pod.StartTime = p.Status.StartTime.Time
		}

		// Count restarts
		for _, cs := range p.Status.ContainerStatuses {
			pod.Restarts += int(cs.RestartCount)
		}

		pods = append(pods, pod)
	}

	return pods, nil
}

// GetPodLogs gets logs for a pod
func (r *KubernetesRepository) GetPodLogs(ctx context.Context, workspaceID, projectID, podName, container string, opts application.LogOptions) ([]application.LogEntry, error) {
	namespace := r.getNamespace(workspaceID, projectID)
	
	logOpts := &corev1.PodLogOptions{
		Container: container,
		Previous:  opts.Previous,
	}

	if !opts.Since.IsZero() {
		sinceTime := metav1.NewTime(opts.Since)
		logOpts.SinceTime = &sinceTime
	}

	if opts.Limit > 0 {
		tailLines := int64(opts.Limit)
		logOpts.TailLines = &tailLines
	}

	req := r.clientset.CoreV1().Pods(namespace).GetLogs(podName, logOpts)
	logs, err := req.Do(ctx).Raw()
	if err != nil {
		return nil, err
	}

	// Parse logs into entries
	// This is a simplified implementation - in production you'd parse the actual log format
	var entries []application.LogEntry
	timestamp := time.Now()
	for _, line := range splitLines(string(logs)) {
		if line != "" {
			entries = append(entries, application.LogEntry{
				Timestamp: timestamp,
				PodName:   podName,
				Container: container,
				Message:   line,
			})
		}
	}

	return entries, nil
}

// StreamPodLogs streams logs for a pod
func (r *KubernetesRepository) StreamPodLogs(ctx context.Context, workspaceID, projectID, podName, container string, opts application.LogOptions) (io.ReadCloser, error) {
	namespace := r.getNamespace(workspaceID, projectID)
	
	logOpts := &corev1.PodLogOptions{
		Container: container,
		Follow:    opts.Follow,
		Previous:  opts.Previous,
	}

	if !opts.Since.IsZero() {
		sinceTime := metav1.NewTime(opts.Since)
		logOpts.SinceTime = &sinceTime
	}

	req := r.clientset.CoreV1().Pods(namespace).GetLogs(podName, logOpts)
	return req.Stream(ctx)
}

// RestartPod restarts a pod by deleting it
func (r *KubernetesRepository) RestartPod(ctx context.Context, workspaceID, projectID, podName string) error {
	namespace := r.getNamespace(workspaceID, projectID)
	return r.clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
}

// GetPodMetrics gets metrics for pods
func (r *KubernetesRepository) GetPodMetrics(ctx context.Context, workspaceID, projectID string, podNames []string) ([]application.PodMetrics, error) {
	namespace := r.getNamespace(workspaceID, projectID)
	
	var metrics []application.PodMetrics
	for _, podName := range podNames {
		podMetrics, err := r.metricsClientset.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			continue // Skip pods without metrics
		}

		var cpuUsage, memoryUsage float64
		for _, container := range podMetrics.Containers {
			cpu := container.Usage[corev1.ResourceCPU]
			memory := container.Usage[corev1.ResourceMemory]
			
			cpuUsage += float64(cpu.MilliValue()) / 1000 // Convert to cores
			memoryUsage += float64(memory.Value()) / (1024 * 1024) // Convert to MB
		}

		metrics = append(metrics, application.PodMetrics{
			PodName:     podName,
			CPUUsage:    cpuUsage,
			MemoryUsage: memoryUsage,
		})
	}

	return metrics, nil
}

// Helper functions

func (r *KubernetesRepository) getNamespace(workspaceID, projectID string) string {
	// In a real implementation, you would map workspace/project to actual Kubernetes namespace
	// For now, we'll use a simple convention
	if projectID != "" {
		return fmt.Sprintf("ws-%s-proj-%s", workspaceID, projectID)
	}
	return fmt.Sprintf("ws-%s", workspaceID)
}

func (r *KubernetesRepository) convertEnvVars(envVars map[string]string) []corev1.EnvVar {
	var env []corev1.EnvVar
	for k, v := range envVars {
		env = append(env, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return env
}

func (r *KubernetesRepository) convertResources(resources application.ResourceRequests) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(resources.CPURequest),
			corev1.ResourceMemory: resource.MustParse(resources.MemoryRequest),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(resources.CPULimit),
			corev1.ResourceMemory: resource.MustParse(resources.MemoryLimit),
		},
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// CreateCronJob creates a new Kubernetes CronJob
func (r *KubernetesRepository) CreateCronJob(ctx context.Context, workspaceID, projectID string, spec application.CronJobSpec) error {
	namespace := r.getNamespace(workspaceID, projectID)
	
	// Build CronJob resource
	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:        spec.Name,
			Namespace:   namespace,
			Labels:      spec.Labels,
			Annotations: spec.Annotations,
		},
		Spec: batchv1.CronJobSpec{
			Schedule:          spec.Schedule,
			ConcurrencyPolicy: batchv1.ConcurrencyPolicy(spec.ConcurrencyPolicy),
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: spec.Labels,
						},
						Spec: r.buildCronJobPodSpec(spec),
					},
					BackoffLimit:            int32Ptr(3),
					TTLSecondsAfterFinished: int32Ptr(86400), // 24 hours
				},
			},
		},
	}

	_, err := r.clientset.BatchV1().CronJobs(namespace).Create(ctx, cronJob, metav1.CreateOptions{})
	return err
}

// UpdateCronJob updates an existing Kubernetes CronJob
func (r *KubernetesRepository) UpdateCronJob(ctx context.Context, workspaceID, projectID, name string, spec application.CronJobSpec) error {
	namespace := r.getNamespace(workspaceID, projectID)
	
	// Get existing CronJob
	cronJob, err := r.clientset.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get CronJob: %w", err)
	}
	
	// Update spec
	cronJob.Spec.Schedule = spec.Schedule
	cronJob.Spec.ConcurrencyPolicy = batchv1.ConcurrencyPolicy(spec.ConcurrencyPolicy)
	cronJob.Spec.JobTemplate.Spec.Template.Spec = r.buildCronJobPodSpec(spec)
	if spec.Labels != nil {
		cronJob.Labels = spec.Labels
		cronJob.Spec.JobTemplate.Spec.Template.Labels = spec.Labels
	}
	if spec.Annotations != nil {
		cronJob.Annotations = spec.Annotations
	}
	
	// Update CronJob
	_, err = r.clientset.BatchV1().CronJobs(namespace).Update(ctx, cronJob, metav1.UpdateOptions{})
	return err
}

// DeleteCronJob deletes a Kubernetes CronJob
func (r *KubernetesRepository) DeleteCronJob(ctx context.Context, workspaceID, projectID, name string) error {
	namespace := r.getNamespace(workspaceID, projectID)
	
	// Delete CronJob (will also delete associated Jobs)
	deletePolicy := metav1.DeletePropagationForeground
	return r.clientset.BatchV1().CronJobs(namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

// GetCronJobStatus retrieves the status of a Kubernetes CronJob
func (r *KubernetesRepository) GetCronJobStatus(ctx context.Context, workspaceID, projectID, name string) (*application.CronJobStatus, error) {
	namespace := r.getNamespace(workspaceID, projectID)
	
	// Get CronJob
	cronJob, err := r.clientset.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get CronJob: %w", err)
	}
	
	// Convert to domain status
	status := &application.CronJobStatus{
		Schedule:           cronJob.Spec.Schedule,
		LastScheduleTime:   nil,
		LastSuccessfulTime: nil,
		Active:             make([]application.ObjectReference, 0),
	}
	
	// Set LastScheduleTime if available
	if cronJob.Status.LastScheduleTime != nil {
		status.LastScheduleTime = &cronJob.Status.LastScheduleTime.Time
	}
	
	// Set LastSuccessfulTime if available
	if cronJob.Status.LastSuccessfulTime != nil {
		status.LastSuccessfulTime = &cronJob.Status.LastSuccessfulTime.Time
	}
	
	// Convert active job references
	for _, ref := range cronJob.Status.Active {
		status.Active = append(status.Active, application.ObjectReference{
			Name:      ref.Name,
			Namespace: ref.Namespace,
			UID:       string(ref.UID),
		})
	}
	
	return status, nil
}

// TriggerCronJob manually triggers a CronJob by creating a Job from the CronJob template
func (r *KubernetesRepository) TriggerCronJob(ctx context.Context, workspaceID, projectID, name string) error {
	namespace := r.getNamespace(workspaceID, projectID)
	
	// Get CronJob
	cronJob, err := r.clientset.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get CronJob: %w", err)
	}
	
	// Create Job from CronJob template
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-manual-%d", name, time.Now().Unix()),
			Namespace: namespace,
			Labels: map[string]string{
				"cronjob-name": name,
				"triggered-by": "manual",
			},
			Annotations: map[string]string{
				"hexabase.io/triggered-at": time.Now().Format(time.RFC3339),
			},
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}
	
	// Create the Job
	_, err = r.clientset.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	return err
}

// buildCronJobPodSpec builds a PodSpec from CronJobSpec
func (r *KubernetesRepository) buildCronJobPodSpec(spec application.CronJobSpec) corev1.PodSpec {
	// Convert environment variables
	envVars := r.convertEnvVars(spec.EnvVars)
	
	// Build container
	container := corev1.Container{
		Name:      spec.Name,
		Image:     spec.Image,
		Command:   spec.Command,
		Args:      spec.Args,
		Env:       envVars,
		Resources: r.convertResources(spec.Resources),
	}
	
	// Build pod spec
	restartPolicy := corev1.RestartPolicyOnFailure
	if spec.RestartPolicy != "" {
		restartPolicy = corev1.RestartPolicy(spec.RestartPolicy)
	}
	
	return corev1.PodSpec{
		Containers:    []corev1.Container{container},
		RestartPolicy: restartPolicy,
		NodeSelector:  spec.NodeSelector,
	}
}