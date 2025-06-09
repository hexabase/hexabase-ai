package deployer

import (
	"context"
	"fmt"

	"github.com/hexabase/hexabase-ai/cli/hks-func/pkg/function"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/client/clientset/versioned"
)

// Deployer handles function deployment
type Deployer struct {
	knativeClient versioned.Interface
	namespace     string
}

// NewDeployer creates a new deployer
func NewDeployer(knativeClient versioned.Interface, namespace string) *Deployer {
	return &Deployer{
		knativeClient: knativeClient,
		namespace:     namespace,
	}
}

// New is an alias for NewDeployer
func New(knativeClient versioned.Interface, namespace string) *Deployer {
	return NewDeployer(knativeClient, namespace)
}

// Deploy deploys a function to Knative
func (d *Deployer) Deploy(ctx context.Context, config *function.Config, image string) error {
	service := d.buildKnativeService(config, image)

	// Check if service exists
	existing, err := d.knativeClient.ServingV1().Services(d.namespace).Get(ctx, config.Name, metav1.GetOptions{})
	if err == nil {
		// Update existing service
		existing.Spec = service.Spec
		_, err = d.knativeClient.ServingV1().Services(d.namespace).Update(ctx, existing, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update function: %w", err)
		}
		fmt.Printf("Function %s updated successfully\n", config.Name)
	} else {
		// Create new service
		_, err = d.knativeClient.ServingV1().Services(d.namespace).Create(ctx, service, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to deploy function: %w", err)
		}
		fmt.Printf("Function %s deployed successfully\n", config.Name)
	}

	return nil
}

// buildKnativeService builds a Knative Service from function config
func (d *Deployer) buildKnativeService(config *function.Config, image string) *servingv1.Service {
	service := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: d.namespace,
			Labels:    config.Deploy.Labels,
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: config.Deploy.Labels,
						Annotations: map[string]string{
							"autoscaling.knative.dev/minScale": fmt.Sprintf("%d", config.Deploy.Autoscaling.MinScale),
							"autoscaling.knative.dev/maxScale": fmt.Sprintf("%d", config.Deploy.Autoscaling.MaxScale),
						},
					},
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: image,
									Env:   d.buildEnvVars(config.Environment),
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse(config.Deploy.Resources.CPU),
											corev1.ResourceMemory: resource.MustParse(config.Deploy.Resources.Memory),
										},
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse(config.Deploy.Resources.CPU),
											corev1.ResourceMemory: resource.MustParse(config.Deploy.Resources.Memory),
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

	// Add timeout if specified
	if config.Deploy.Timeout > 0 {
		timeoutSeconds := int64(config.Deploy.Timeout)
		service.Spec.Template.Spec.TimeoutSeconds = &timeoutSeconds
	}

	// Add concurrency limit if specified
	if config.Deploy.Concurrency > 0 {
		concurrency := int64(config.Deploy.Concurrency)
		service.Spec.Template.Spec.ContainerConcurrency = &concurrency
	}

	return service
}

// buildEnvVars converts environment map to Kubernetes env vars
func (d *Deployer) buildEnvVars(env map[string]string) []corev1.EnvVar {
	var envVars []corev1.EnvVar
	for k, v := range env {
		envVars = append(envVars, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return envVars
}