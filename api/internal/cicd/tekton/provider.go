// Package tekton provides a CI/CD provider implementation using Tekton.
package tekton

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/cicd/domain"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	tektonclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/apis"
)

// TektonProvider implements the domain.Provider interface using Tekton.
type TektonProvider struct {
	tektonClient  tektonclient.Interface
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
}

// NewTektonProvider creates a new TektonProvider.
// It requires a Tekton client, a standard Kubernetes client, and a dynamic client.
func NewTektonProvider(tektonClient tektonclient.Interface, kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) *TektonProvider {
	return &TektonProvider{
		tektonClient:  tektonClient,
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
	}
}

// GetResource retrieves a specific resource from Kubernetes.
func (p *TektonProvider) GetResource(ctx context.Context, projectID string, namespace string, groupVersionResource schema.GroupVersionResource, name string) (*unstructured.Unstructured, error) {
	return p.dynamicClient.Resource(groupVersionResource).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

// ListResources lists resources of a specific type in a namespace.
func (p *TektonProvider) ListResources(ctx context.Context, projectID string, namespace string, groupVersionResource schema.GroupVersionResource) (*unstructured.UnstructuredList, error) {
	return p.dynamicClient.Resource(groupVersionResource).Namespace(namespace).List(ctx, metav1.ListOptions{})
}

// CreateResource creates a new resource in Kubernetes.
func (p *TektonProvider) CreateResource(ctx context.Context, projectID string, namespace string, groupVersionResource schema.GroupVersionResource, resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return p.dynamicClient.Resource(groupVersionResource).Namespace(namespace).Create(ctx, resource, metav1.CreateOptions{})
}

// UpdateResource updates an existing resource in Kubernetes.
func (p *TektonProvider) UpdateResource(ctx context.Context, projectID string, namespace string, groupVersionResource schema.GroupVersionResource, resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return p.dynamicClient.Resource(groupVersionResource).Namespace(namespace).Update(ctx, resource, metav1.UpdateOptions{})
}

// DeleteResource deletes a resource from Kubernetes.
func (p *TektonProvider) DeleteResource(ctx context.Context, projectID string, namespace string, groupVersionResource schema.GroupVersionResource, name string) error {
	return p.dynamicClient.Resource(groupVersionResource).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// CreateFromTemplate creates a pipeline from a template.
func (p *TektonProvider) CreateFromTemplate(ctx context.Context, templateID string, params map[string]any) (*domain.PipelineConfig, error) {
	// TODO: Implement template-based pipeline creation
	// For now, return a basic pipeline config
	return &domain.PipelineConfig{
		Name: fmt.Sprintf("Pipeline from template %s", templateID),
		GitRepo: domain.GitConfig{
			URL:    "https://github.com/example/repo.git",
			Branch: "main",
		},
		Metadata: params,
	}, nil
}

// ValidateConfig validates the pipeline configuration.
func (p *TektonProvider) ValidateConfig(ctx context.Context, config *domain.PipelineConfig) error {
	if config.Name == "" {
		return fmt.Errorf("pipeline name is required")
	}
	if config.GitRepo.URL == "" {
		return fmt.Errorf("git repository URL is required")
	}
	return nil
}

// RunPipeline runs a pipeline with the given configuration.
func (p *TektonProvider) RunPipeline(ctx context.Context, config *domain.PipelineConfig) (*domain.PipelineRun, error) {
	namespace := "default" // TODO: derive from project/workspace
	runName := fmt.Sprintf("%s-run-%s", config.Name, strings.ToLower(uuid.New().String()[:8]))

	// Build pipeline parameters
	params := []pipelinev1.Param{
		{Name: "repo-url", Value: pipelinev1.ParamValue{Type: pipelinev1.ParamTypeString, StringVal: config.GitRepo.URL}},
		{Name: "revision", Value: pipelinev1.ParamValue{Type: pipelinev1.ParamTypeString, StringVal: config.GitRepo.Branch}},
		{Name: "app-name", Value: pipelinev1.ParamValue{Type: pipelinev1.ParamTypeString, StringVal: config.Name}},
	}

	// Add image ref if registry config is provided
	if config.RegistryConfig != nil {
		imageRef := fmt.Sprintf("%s/%s/%s:latest", config.RegistryConfig.URL, config.RegistryConfig.Namespace, config.Name)
		params = append(params, pipelinev1.Param{
			Name:  "image-ref",
			Value: pipelinev1.ParamValue{Type: pipelinev1.ParamTypeString, StringVal: imageRef},
		})
	}

	pipelineRun := &pipelinev1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "hexabase",
				"hexabase.ai/workspace-id":     config.WorkspaceID,
				"hexabase.ai/project-id":       config.ProjectID,
				"hexabase.ai/app-name":         config.Name,
			},
		},
		Spec: pipelinev1.PipelineRunSpec{
			PipelineRef: &pipelinev1.PipelineRef{Name: "build-and-push-pipeline"},
			Params:      params,
		},
	}

	createdPR, err := p.tektonClient.TektonV1().PipelineRuns(namespace).Create(ctx, pipelineRun, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Tekton PipelineRun: %w", err)
	}

	return &domain.PipelineRun{
		ID:          createdPR.Name,
		WorkspaceID: config.WorkspaceID,
		ProjectID:   config.ProjectID,
		Name:        config.Name,
		Status:      domain.PipelineStatusPending,
		StartedAt:   time.Now(),
		Metadata:    config.Metadata,
	}, nil
}

// GetStatus gets the status of a pipeline run.
func (p *TektonProvider) GetStatus(ctx context.Context, workspaceID, runID string) (*domain.PipelineRun, error) {
	// For now, use default namespace
	namespace := "default"
	
	pr, err := p.tektonClient.TektonV1().PipelineRuns(namespace).Get(ctx, runID, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline run: %w", err)
	}
	
	status := domain.PipelineStatusPending
	if pr.Status.GetCondition(apis.ConditionSucceeded).IsTrue() {
		status = domain.PipelineStatusSucceeded
	} else if pr.Status.GetCondition(apis.ConditionSucceeded).IsFalse() {
		status = domain.PipelineStatusFailed
	} else if pr.Status.GetCondition(apis.ConditionSucceeded).IsUnknown() {
		status = domain.PipelineStatusRunning
	}
	
	run := &domain.PipelineRun{
		ID:          runID,
		WorkspaceID: workspaceID,
		Name:        pr.Name,
		Status:      status,
		StartedAt:   pr.Status.StartTime.Time,
	}
	
	if pr.Status.CompletionTime != nil {
		run.FinishedAt = &pr.Status.CompletionTime.Time
	}
	
	return run, nil
}

// ListPipelines lists all pipeline runs for a project.
func (p *TektonProvider) ListPipelines(ctx context.Context, workspaceID, projectID string) ([]*domain.PipelineRun, error) {
	// For now, use default namespace
	namespace := "default"
	
	prs, err := p.tektonClient.TektonV1().PipelineRuns(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipeline runs: %w", err)
	}
	
	var runs []*domain.PipelineRun
	for _, pr := range prs.Items {
		status := domain.PipelineStatusPending
		if pr.Status.GetCondition(apis.ConditionSucceeded).IsTrue() {
			status = domain.PipelineStatusSucceeded
		} else if pr.Status.GetCondition(apis.ConditionSucceeded).IsFalse() {
			status = domain.PipelineStatusFailed
		} else if pr.Status.GetCondition(apis.ConditionSucceeded).IsUnknown() {
			status = domain.PipelineStatusRunning
		}
		
		run := &domain.PipelineRun{
			ID:          pr.Name,
			WorkspaceID: workspaceID,
			ProjectID:   projectID,
			Name:        pr.Name,
			Status:      status,
			StartedAt:   pr.Status.StartTime.Time,
		}
		
		if pr.Status.CompletionTime != nil {
			run.FinishedAt = &pr.Status.CompletionTime.Time
		}
		
		runs = append(runs, run)
	}
	
	return runs, nil
}

// DeletePipeline deletes a pipeline run.
func (p *TektonProvider) DeletePipeline(ctx context.Context, workspaceID, runID string) error {
	// For now, use default namespace
	namespace := "default"
	return p.tektonClient.TektonV1().PipelineRuns(namespace).Delete(ctx, runID, metav1.DeleteOptions{})
}

// GetLogs gets the logs for a pipeline run.
func (p *TektonProvider) GetLogs(ctx context.Context, workspaceID, runID string) ([]domain.LogEntry, error) {
	// TODO: Implement proper log retrieval
	return []domain.LogEntry{}, nil
}

// StreamLogs streams the logs for a pipeline run.
func (p *TektonProvider) StreamLogs(ctx context.Context, workspaceID, runID string) (io.ReadCloser, error) {
	namespace := "default" // TODO: derive from workspaceID
	
	_, err := p.tektonClient.TektonV1().PipelineRuns(namespace).Get(ctx, runID, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not find PipelineRun %s: %w", runID, err)
	}

	// In Tekton, pods are labeled with the PipelineRun name.
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("tekton.dev/pipelineRun=%s", runID),
	}
	pods, err := p.kubeClient.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil || len(pods.Items) == 0 {
		return nil, fmt.Errorf("could not find pod for PipelineRun %s: %w", runID, err)
	}

	// For simplicity, stream logs from the first pod found.
	// A more robust implementation might aggregate logs or let the user select a TaskRun.
	podName := pods.Items[0].Name

	req := p.kubeClient.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Follow: true, // Stream logs
	})

	logStream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to stream logs for pod %s: %w", podName, err)
	}

	return logStream, nil
}

// GetTemplates gets available pipeline templates.
func (p *TektonProvider) GetTemplates(ctx context.Context) ([]*domain.PipelineTemplate, error) {
	// TODO: Implement template listing
	return []*domain.PipelineTemplate{}, nil
}

// GetName returns the provider name.
func (p *TektonProvider) GetName() string {
	return "tekton"
}

// GetVersion returns the provider version.
func (p *TektonProvider) GetVersion() string {
	return "v1"
}

// IsHealthy checks if the provider is healthy.
func (p *TektonProvider) IsHealthy() bool {
	// TODO: Implement health check
	return true
}

// CancelPipeline cancels a running pipeline by deleting the PipelineRun.
func (p *TektonProvider) CancelPipeline(ctx context.Context, workspaceID, runID string) error {
	// For now, use default namespace
	namespace := "default"
	return p.tektonClient.TektonV1().PipelineRuns(namespace).Delete(ctx, runID, metav1.DeleteOptions{})
}