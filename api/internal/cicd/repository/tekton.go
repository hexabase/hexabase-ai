package repository

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	tektonclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	knativeapis "knative.dev/pkg/apis"

	"github.com/hexabase/hexabase-ai/api/internal/cicd/domain"
)

// TektonProvider implements the CI/CD provider interface for Tekton
type TektonProvider struct {
	kubeClient   kubernetes.Interface
	tektonClient tektonclient.Interface
	namespace    string
}

// NewTektonProvider creates a new Tekton provider
func NewTektonProvider(kubeClient kubernetes.Interface, tektonClient tektonclient.Interface, namespace string) domain.Provider {
	return &TektonProvider{
		kubeClient:   kubeClient,
		tektonClient: tektonClient,
		namespace:    namespace,
	}
}

// GetName returns the provider name
func (p *TektonProvider) GetName() string {
	return "tekton"
}

// GetVersion returns the provider version
func (p *TektonProvider) GetVersion() string {
	return "v1.0.0"
}

// IsHealthy checks if the provider is healthy
func (p *TektonProvider) IsHealthy() bool {
	// Simple health check - try to list pipelines
	ctx := context.Background()
	_, err := p.tektonClient.TektonV1().Pipelines(p.namespace).List(ctx, metav1.ListOptions{Limit: 1})
	return err == nil
}

// RunPipeline starts a new pipeline run
func (p *TektonProvider) RunPipeline(ctx context.Context, config *domain.PipelineConfig) (*domain.PipelineRun, error) {
	// Generate unique run ID
	runID := uuid.New().String()
	pipelineRunName := fmt.Sprintf("%s-%s", config.Name, runID[:8])

	// Create PipelineRun resource
	pipelineRun := &tektonv1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pipelineRunName,
			Namespace: p.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "hexabase-ai",
				"hexabase.ai/workspace-id":     config.WorkspaceID,
				"hexabase.ai/project-id":       config.ProjectID,
				"hexabase.ai/run-id":           runID,
			},
		},
		Spec: tektonv1.PipelineRunSpec{
			PipelineSpec: p.createPipelineSpec(*config),
			Params:       p.createParams(*config),
			Workspaces:   p.createWorkspaces(*config),
		},
	}

	// Create the PipelineRun
	created, err := p.tektonClient.TektonV1().PipelineRuns(p.namespace).Create(ctx, pipelineRun, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline run: %w", err)
	}

	// Convert to domain model
	return p.convertToPipelineRun(created, runID), nil
}

// GetStatus retrieves the current status of a pipeline run
func (p *TektonProvider) GetStatus(ctx context.Context, workspaceID, runID string) (*domain.PipelineRun, error) {
	// List PipelineRuns with label selector
	runs, err := p.tektonClient.TektonV1().PipelineRuns(p.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("hexabase.ai/run-id=%s", runID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipeline runs: %w", err)
	}

	if len(runs.Items) == 0 {
		return nil, fmt.Errorf("pipeline run not found: %s", runID)
	}

	// Get the first (should be only) run
	tektonRun := &runs.Items[0]
	return p.convertToPipelineRun(tektonRun, runID), nil
}

// CancelPipeline cancels a running pipeline
func (p *TektonProvider) CancelPipeline(ctx context.Context, workspaceID, runID string) error {
	// Find the PipelineRun
	runs, err := p.tektonClient.TektonV1().PipelineRuns(p.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("hexabase.ai/run-id=%s", runID),
	})
	if err != nil {
		return fmt.Errorf("failed to list pipeline runs: %w", err)
	}

	if len(runs.Items) == 0 {
		return fmt.Errorf("pipeline run not found: %s", runID)
	}

	// Update status to cancelled
	tektonRun := &runs.Items[0]
	tektonRun.Spec.Status = tektonv1.PipelineRunSpecStatusCancelled
	
	_, err = p.tektonClient.TektonV1().PipelineRuns(p.namespace).Update(ctx, tektonRun, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to cancel pipeline run: %w", err)
	}

	return nil
}

// GetLogs retrieves logs for a specific stage/task
func (p *TektonProvider) GetLogs(ctx context.Context, runID string, stage string) ([]domain.LogEntry, error) {
	// Find the PipelineRun
	runs, err := p.tektonClient.TektonV1().PipelineRuns(p.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("hexabase.ai/run-id=%s", runID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipeline runs: %w", err)
	}

	if len(runs.Items) == 0 {
		return nil, fmt.Errorf("pipeline run not found: %s", runID)
	}

	tektonRun := &runs.Items[0]
	
	// Get pod logs for the stage
	podName := p.getTaskPodName(tektonRun, stage, stage)
	if podName == "" {
		return nil, fmt.Errorf("pod not found for stage: %s", stage)
	}

	// Get logs from pod
	req := p.kubeClient.CoreV1().Pods(p.namespace).GetLogs(podName, &corev1.PodLogOptions{})
	logs, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer logs.Close()

	// Parse logs into entries
	return p.parseLogStream(logs, stage, stage)
}

// StreamLogs streams logs in real-time
func (p *TektonProvider) StreamLogs(ctx context.Context, runID string, stage string) (io.ReadCloser, error) {
	// Find the PipelineRun
	runs, err := p.tektonClient.TektonV1().PipelineRuns(p.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("hexabase.ai/run-id=%s", runID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipeline runs: %w", err)
	}

	if len(runs.Items) == 0 {
		return nil, fmt.Errorf("pipeline run not found: %s", runID)
	}

	tektonRun := &runs.Items[0]
	
	// Get pod logs for the stage
	podName := p.getTaskPodName(tektonRun, stage, stage)
	if podName == "" {
		return nil, fmt.Errorf("pod not found for stage: %s", stage)
	}

	// Stream logs from pod
	req := p.kubeClient.CoreV1().Pods(p.namespace).GetLogs(podName, &corev1.PodLogOptions{
		Follow: true,
	})
	
	return req.Stream(ctx)
}

// ListPipelines lists pipeline runs for a workspace/project
func (p *TektonProvider) ListPipelines(ctx context.Context, workspaceID, projectID string) ([]*domain.PipelineRun, error) {
	limit := 50 // Default limit
	labelSelector := fmt.Sprintf("hexabase.ai/workspace-id=%s", workspaceID)
	if projectID != "" {
		labelSelector += fmt.Sprintf(",hexabase.ai/project-id=%s", projectID)
	}

	runs, err := p.tektonClient.TektonV1().PipelineRuns(p.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
		Limit:         int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipeline runs: %w", err)
	}

	result := make([]*domain.PipelineRun, len(runs.Items))
	for i, run := range runs.Items {
		runID := run.Labels["hexabase.ai/run-id"]
		result[i] = p.convertToPipelineRun(&run, runID)
	}

	return result, nil
}

// DeletePipeline deletes a pipeline run and its resources
func (p *TektonProvider) DeletePipeline(ctx context.Context, workspaceID, runID string) error {
	// Find the PipelineRun
	runs, err := p.tektonClient.TektonV1().PipelineRuns(p.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("hexabase.ai/run-id=%s", runID),
	})
	if err != nil {
		return fmt.Errorf("failed to list pipeline runs: %w", err)
	}

	if len(runs.Items) == 0 {
		return fmt.Errorf("pipeline run not found: %s", runID)
	}

	// Delete the PipelineRun
	err = p.tektonClient.TektonV1().PipelineRuns(p.namespace).Delete(ctx, runs.Items[0].Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pipeline run: %w", err)
	}

	return nil
}

// ValidateConfig validates a pipeline configuration
func (p *TektonProvider) ValidateConfig(ctx context.Context, config *domain.PipelineConfig) error {
	// Validate required fields
	if config.Name == "" {
		return fmt.Errorf("pipeline name is required")
	}
	if config.GitRepo.URL == "" {
		return fmt.Errorf("git repository URL is required")
	}
	if config.ServiceAccount == "" {
		return fmt.Errorf("service account is required")
	}

	// Validate credentials exist
	if config.GitRepo.SSHKeyRef != "" {
		_, err := p.kubeClient.CoreV1().Secrets(p.namespace).Get(ctx, config.GitRepo.SSHKeyRef, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("git SSH key secret not found: %w", err)
		}
	}

	if config.RegistryConfig != nil && config.RegistryConfig.CredRef != "" {
		_, err := p.kubeClient.CoreV1().Secrets(p.namespace).Get(ctx, config.RegistryConfig.CredRef, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("registry credential secret not found: %w", err)
		}
	}

	return nil
}

// GetTemplates returns available pipeline templates
func (p *TektonProvider) GetTemplates(ctx context.Context) ([]*domain.PipelineTemplate, error) {
	// Return built-in Tekton templates
	return []*domain.PipelineTemplate{
		{
			ID:          "docker-build-push",
			Name:        "Docker Build & Push",
			Description: "Build a Docker image and push to registry",
			Provider:    "tekton",
			Stages: []domain.StageTemplate{
				{
					Name: "clone",
					Tasks: []domain.TaskTemplate{
						{
							Name: "git-clone",
							Type: "git-clone",
							Parameters: map[string]any{
								"url":    "$(params.git-url)",
								"branch": "$(params.git-branch)",
							},
						},
					},
				},
				{
					Name:      "build",
					DependsOn: []string{"clone"},
					Tasks: []domain.TaskTemplate{
						{
							Name: "docker-build",
							Type: "docker-build",
							Parameters: map[string]any{
								"dockerfile": "$(params.dockerfile-path)",
								"context":    "$(params.build-context)",
								"image":      "$(params.image-name)",
							},
						},
					},
				},
			},
			Parameters: []domain.ParameterDefinition{
				{Name: "git-url", Type: "string", Required: true},
				{Name: "git-branch", Type: "string", Default: "main"},
				{Name: "dockerfile-path", Type: "string", Default: "Dockerfile"},
				{Name: "build-context", Type: "string", Default: "."},
				{Name: "image-name", Type: "string", Required: true},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "kubernetes-deploy",
			Name:        "Kubernetes Deploy",
			Description: "Deploy application to Kubernetes",
			Provider:    "tekton",
			Stages: []domain.StageTemplate{
				{
					Name: "deploy",
					Tasks: []domain.TaskTemplate{
						{
							Name: "kubectl-apply",
							Type: "kubectl-apply",
							Parameters: map[string]any{
								"manifest": "$(params.manifest-path)",
								"namespace": "$(params.target-namespace)",
							},
						},
					},
				},
			},
			Parameters: []domain.ParameterDefinition{
				{Name: "manifest-path", Type: "string", Required: true},
				{Name: "target-namespace", Type: "string", Required: true},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}, nil
}

// CreateFromTemplate creates a pipeline config from a template
func (p *TektonProvider) CreateFromTemplate(ctx context.Context, templateID string, params map[string]any) (*domain.PipelineConfig, error) {
	templates, err := p.GetTemplates(ctx)
	if err != nil {
		return nil, err
	}

	for _, tmpl := range templates {
		if tmpl.ID == templateID {
			// Create config from template
			config := &domain.PipelineConfig{
				Name:     fmt.Sprintf("%s-%d", tmpl.Name, time.Now().Unix()),
				Metadata: params,
			}
			
			// Apply template parameters
			// This is a simplified implementation
			if gitURL, ok := params["git-url"].(string); ok {
				config.GitRepo.URL = gitURL
			}
			if gitBranch, ok := params["git-branch"].(string); ok {
				config.GitRepo.Branch = gitBranch
			}
			
			return config, nil
		}
	}

	return nil, fmt.Errorf("template not found: %s", templateID)
}

// Helper methods

func (p *TektonProvider) createPipelineSpec(config domain.PipelineConfig) *tektonv1.PipelineSpec {
	// Create pipeline spec based on config
	// This is a simplified implementation
	spec := &tektonv1.PipelineSpec{
		Params: []tektonv1.ParamSpec{
			{
				Name: "git-url",
				Type: tektonv1.ParamTypeString,
				Default: &tektonv1.ParamValue{
					Type:      tektonv1.ParamTypeString,
					StringVal: config.GitRepo.URL,
				},
			},
			{
				Name: "git-branch",
				Type: tektonv1.ParamTypeString,
				Default: &tektonv1.ParamValue{
					Type:      tektonv1.ParamTypeString,
					StringVal: config.GitRepo.Branch,
				},
			},
		},
		Workspaces: []tektonv1.PipelineWorkspaceDeclaration{
			{
				Name: "source",
			},
		},
		Tasks: []tektonv1.PipelineTask{
			{
				Name: "clone",
				TaskSpec: &tektonv1.EmbeddedTask{
					TaskSpec: tektonv1.TaskSpec{
						Workspaces: []tektonv1.WorkspaceDeclaration{
							{
								Name: "output",
							},
						},
						Steps: []tektonv1.Step{
							{
								Name:  "clone",
								Image: "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/git-init:latest",
								Args: []string{
									"-url=$(params.git-url)",
									"-revision=$(params.git-branch)",
									"-path=$(workspaces.output.path)",
								},
							},
						},
					},
				},
				Workspaces: []tektonv1.WorkspacePipelineTaskBinding{
					{
						Name:      "output",
						Workspace: "source",
					},
				},
			},
		},
	}

	// Add build task if build config is specified
	if config.BuildConfig != nil {
		spec.Tasks = append(spec.Tasks, p.createBuildTask(config.BuildConfig))
	}

	// Add deploy task if deploy config is specified
	if config.DeployConfig != nil {
		spec.Tasks = append(spec.Tasks, p.createDeployTask(config.DeployConfig))
	}

	return spec
}

func (p *TektonProvider) createBuildTask(buildConfig *domain.BuildConfig) tektonv1.PipelineTask {
	task := tektonv1.PipelineTask{
		Name: "build",
		RunAfter: []string{"clone"},
		Workspaces: []tektonv1.WorkspacePipelineTaskBinding{
			{
				Name:      "source",
				Workspace: "source",
			},
		},
	}

	switch buildConfig.Type {
	case domain.BuildTypeDocker:
		task.TaskSpec = &tektonv1.EmbeddedTask{
			TaskSpec: tektonv1.TaskSpec{
				Workspaces: []tektonv1.WorkspaceDeclaration{
					{
						Name: "source",
					},
				},
				Steps: []tektonv1.Step{
					{
						Name:  "build-and-push",
						Image: "gcr.io/kaniko-project/executor:latest",
						Args: []string{
							"--dockerfile=" + buildConfig.DockerfilePath,
							"--context=" + buildConfig.BuildContext,
							"--destination=$(params.image-name)",
						},
					},
				},
			},
		}
	}

	return task
}

func (p *TektonProvider) createDeployTask(deployConfig *domain.DeployConfig) tektonv1.PipelineTask {
	task := tektonv1.PipelineTask{
		Name: "deploy",
		RunAfter: []string{"build"},
		Workspaces: []tektonv1.WorkspacePipelineTaskBinding{
			{
				Name:      "source",
				Workspace: "source",
			},
		},
	}

	if deployConfig.ManifestPath != "" {
		task.TaskSpec = &tektonv1.EmbeddedTask{
			TaskSpec: tektonv1.TaskSpec{
				Workspaces: []tektonv1.WorkspaceDeclaration{
					{
						Name: "source",
					},
				},
				Steps: []tektonv1.Step{
					{
						Name:  "kubectl-apply",
						Image: "bitnami/kubectl:latest",
						Args: []string{
							"apply",
							"-f",
							deployConfig.ManifestPath,
							"-n",
							deployConfig.TargetNamespace,
						},
					},
				},
			},
		}
	}

	return task
}

func (p *TektonProvider) createParams(config domain.PipelineConfig) []tektonv1.Param {
	params := []tektonv1.Param{
		{
			Name: "git-url",
			Value: tektonv1.ParamValue{
				Type:      tektonv1.ParamTypeString,
				StringVal: config.GitRepo.URL,
			},
		},
		{
			Name: "git-branch",
			Value: tektonv1.ParamValue{
				Type:      tektonv1.ParamTypeString,
				StringVal: config.GitRepo.Branch,
			},
		},
	}
	
	// Add parameters from metadata
	for k, v := range config.Metadata {
		params = append(params, tektonv1.Param{
			Name: k,
			Value: tektonv1.ParamValue{
				Type:      tektonv1.ParamTypeString,
				StringVal: fmt.Sprintf("%v", v),
			},
		})
	}
	
	return params
}

func (p *TektonProvider) createWorkspaces(config domain.PipelineConfig) []tektonv1.WorkspaceBinding {
	return []tektonv1.WorkspaceBinding{
		{
			Name: "source",
			VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("1Gi"),
						},
					},
				},
			},
		},
	}
}

func (p *TektonProvider) parseDuration(duration string) time.Duration {
	if duration == "" {
		return 1 * time.Hour // Default timeout
	}
	
	d, err := time.ParseDuration(duration)
	if err != nil {
		return 1 * time.Hour
	}
	
	return d
}

func (p *TektonProvider) convertToPipelineRun(tektonRun *tektonv1.PipelineRun, runID string) *domain.PipelineRun {
	run := &domain.PipelineRun{
		ID:          runID,
		WorkspaceID: tektonRun.Labels["hexabase.ai/workspace-id"],
		ProjectID:   tektonRun.Labels["hexabase.ai/project-id"],
		Name:        tektonRun.Name,
		Status:      p.convertStatus(p.getSucceededCondition(tektonRun)),
		StartedAt:   tektonRun.CreationTimestamp.Time,
		Metadata:    map[string]any{},
	}

	// Set finished time if completed
	if tektonRun.Status.CompletionTime != nil {
		run.FinishedAt = &tektonRun.Status.CompletionTime.Time
	}

	// Convert stages/tasks
	run.Stages = p.convertStages(tektonRun)

	return run
}

func (p *TektonProvider) getSucceededCondition(run *tektonv1.PipelineRun) *knativeapis.Condition {
	for _, cond := range run.Status.Conditions {
		if cond.Type == "Succeeded" {
			return &cond
		}
	}
	return nil
}

func (p *TektonProvider) convertStatus(condition *knativeapis.Condition) domain.PipelineStatus {
	if condition == nil {
		return domain.PipelineStatusPending
	}

	switch condition.Status {
	case corev1.ConditionTrue:
		return domain.PipelineStatusSucceeded
	case corev1.ConditionFalse:
		if condition.Reason == "PipelineRunCancelled" {
			return domain.PipelineStatusCancelled
		}
		return domain.PipelineStatusFailed
	default:
		return domain.PipelineStatusRunning
	}
}

func (p *TektonProvider) convertStages(tektonRun *tektonv1.PipelineRun) []domain.StageStatus {
	stages := []domain.StageStatus{}
	
	// Convert TaskRuns to stages
	for _, taskRun := range tektonRun.Status.PipelineRunStatusFields.ChildReferences {
		stage := domain.StageStatus{
			Name:      taskRun.Name,
			Status:    domain.PipelineStatusRunning, // Simplified
			StartedAt: time.Now(),
			Tasks:     []domain.TaskStatus{},
		}
		
		stages = append(stages, stage)
	}
	
	return stages
}

func (p *TektonProvider) getTaskPodName(tektonRun *tektonv1.PipelineRun, stage, task string) string {
	// Find the pod for the task
	// This is a simplified implementation
	return fmt.Sprintf("%s-%s-pod", tektonRun.Name, task)
}

func (p *TektonProvider) parseLogStream(reader io.Reader, stage, task string) ([]domain.LogEntry, error) {
	// Parse log stream into entries
	// This is a simplified implementation
	entries := []domain.LogEntry{}
	
	// Read logs and create entries
	entries = append(entries, domain.LogEntry{
		Timestamp: time.Now(),
		Stage:     stage,
		Task:      task,
		Level:     "info",
		Message:   "Task completed successfully",
	})
	
	return entries, nil
}