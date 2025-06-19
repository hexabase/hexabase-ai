package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/application/domain"
)

// Service implements the application service interface
type Service struct {
	repo    domain.Repository
	k8s     domain.KubernetesRepository
	k8sRepo domain.KubernetesRepository // Alias for backward compatibility
	logger  *slog.Logger
}

// NewService creates a new application service
func NewService(repo domain.Repository, k8s domain.KubernetesRepository) domain.Service {
	return &Service{
		repo:    repo,
		k8s:     k8s,
		k8sRepo: k8s, // Set alias
		logger:  slog.Default(),
	}
}

// CreateApplication creates a new application
func (s *Service) CreateApplication(ctx context.Context, workspaceID string, req domain.CreateApplicationRequest) (*domain.Application, error) {
	// Validate request
	if !req.Type.IsValid() {
		return nil, errors.New("invalid application type")
	}
	if !req.Source.Type.IsValid() {
		return nil, errors.New("invalid source type")
	}
	if req.Source.Type == domain.SourceTypeImage && req.Source.Image == "" {
		return nil, errors.New("image is required for image source type")
	}
	if req.Source.Type == domain.SourceTypeGit && req.Source.GitURL == "" {
		return nil, errors.New("git URL is required for git source type")
	}

	// Check if application name already exists
	existing, _ := s.repo.GetApplicationByName(ctx, workspaceID, req.ProjectID, req.Name)
	if existing != nil {
		return nil, fmt.Errorf("application with name %s already exists", req.Name)
	}

	// Create application entity
	app := &domain.Application{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Type:        req.Type,
		Status:      domain.ApplicationStatusPending,
		Source:      req.Source,
		Config:      req.Config,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Apply node affinity if specified
	if req.NodePoolID != "" {
		app.Config.NodeSelector = map[string]string{
			"node-pool": req.NodePoolID,
		}
	}

	// Handle CronJob type specifically
	if req.Type == domain.ApplicationTypeCronJob {
		// Set CronJob specific fields from request
		app.CronSchedule = req.CronSchedule
		app.CronCommand = req.CronCommand
		app.CronArgs = req.CronArgs
		app.TemplateAppID = req.TemplateAppID
		
		// Use CreateCronJob for CronJob type
		return app, s.CreateCronJob(ctx, app)
	}

	// Save to database
	if err := s.repo.CreateApplication(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	// Create deployment event
	event := &domain.ApplicationEvent{
		ID:            uuid.New().String(),
		ApplicationID: app.ID,
		Type:          "deployment.started",
		Message:       "Application deployment started",
		Timestamp:     time.Now(),
	}
	s.repo.CreateEvent(ctx, event)

	// Update status to deploying
	app.Status = domain.ApplicationStatusDeploying
	s.repo.UpdateApplication(ctx, app)

	// Deploy to Kubernetes
	go s.deployApplication(context.Background(), app)

	return app, nil
}

// deployApplication handles the actual deployment to Kubernetes
func (s *Service) deployApplication(ctx context.Context, app *domain.Application) {
	var err error
	defer func() {
		if err != nil {
			app.Status = domain.ApplicationStatusError
			s.repo.UpdateApplication(ctx, app)
			event := &domain.ApplicationEvent{
				ID:            uuid.New().String(),
				ApplicationID: app.ID,
				Type:          "deployment.failed",
				Message:       fmt.Sprintf("Deployment failed: %v", err),
				Details:       err.Error(),
				Timestamp:     time.Now(),
			}
			s.repo.CreateEvent(ctx, event)
		}
	}()

	// Deploy based on application type
	switch app.Type {
	case domain.ApplicationTypeStateless:
		err = s.deployStatelessApp(ctx, app)
	case domain.ApplicationTypeStateful:
		err = s.deployStatefulApp(ctx, app)
	case domain.ApplicationTypeCronJob:
		// CronJob deployment is handled separately through CreateCronJob
		app.Status = domain.ApplicationStatusRunning
		s.repo.UpdateApplication(ctx, app)
		return
	case domain.ApplicationTypeFunction:
		// Function deployment will be implemented later
		err = errors.New("function type not yet implemented")
	default:
		err = fmt.Errorf("unknown application type: %s", app.Type)
	}

	if err != nil {
		return
	}

	// Create service
	serviceSpec := domain.ServiceSpec{
		Name:       app.Name,
		Port:       app.Config.Port,
		TargetPort: app.Config.Port,
		Selector: map[string]string{
			"app": app.Name,
		},
		Type: "ClusterIP",
	}
	if err = s.k8s.CreateService(ctx, app.WorkspaceID, app.ProjectID, serviceSpec); err != nil {
		return
	}

	// Create ingress if requested
	if app.Config.NetworkConfig != nil && app.Config.NetworkConfig.CreateIngress {
		ingressSpec := domain.IngressSpec{
			Name:        app.Name,
			Host:        app.Config.NetworkConfig.CustomDomain,
			Path:        app.Config.NetworkConfig.IngressPath,
			ServiceName: app.Name,
			ServicePort: app.Config.Port,
			TLSEnabled:  app.Config.NetworkConfig.TLSEnabled,
			Annotations: app.Config.NetworkConfig.Annotations,
		}
		if err = s.k8s.CreateIngress(ctx, app.WorkspaceID, app.ProjectID, ingressSpec); err != nil {
			// Ingress creation failure is not critical
			event := &domain.ApplicationEvent{
				ID:            uuid.New().String(),
				ApplicationID: app.ID,
				Type:          "ingress.failed",
				Message:       "Failed to create ingress",
				Details:       err.Error(),
				Timestamp:     time.Now(),
			}
			s.repo.CreateEvent(ctx, event)
		}
	}

	// Update status to running
	app.Status = domain.ApplicationStatusRunning
	s.repo.UpdateApplication(ctx, app)

	// Get endpoints
	endpoints, _ := s.k8s.GetServiceEndpoints(ctx, app.WorkspaceID, app.ProjectID, app.Name)
	app.Endpoints = endpoints
	s.repo.UpdateApplication(ctx, app)

	// Create success event
	event := &domain.ApplicationEvent{
		ID:            uuid.New().String(),
		ApplicationID: app.ID,
		Type:          "deployment.succeeded",
		Message:       "Application deployed successfully",
		Timestamp:     time.Now(),
	}
	s.repo.CreateEvent(ctx, event)
}

func (s *Service) deployStatelessApp(ctx context.Context, app *domain.Application) error {
	deploymentSpec := domain.DeploymentSpec{
		Name:         app.Name,
		Replicas:     app.Config.Replicas,
		Image:        app.Source.Image,
		Port:         app.Config.Port,
		EnvVars:      app.Config.EnvVars,
		Resources:    app.Config.Resources,
		NodeSelector: app.Config.NodeSelector,
		Labels: map[string]string{
			"app": app.Name,
		},
	}
	return s.k8s.CreateDeployment(ctx, app.WorkspaceID, app.ProjectID, deploymentSpec)
}

func (s *Service) deployStatefulApp(ctx context.Context, app *domain.Application) error {
	// Create PVC first if storage is configured
	if app.Config.Storage != nil {
		pvcSpec := domain.PVCSpec{
			Name:         app.Name + "-data",
			Size:         app.Config.Storage.Size,
			StorageClass: app.Config.Storage.StorageClass,
			AccessMode:   "ReadWriteOnce",
		}
		if err := s.k8s.CreatePVC(ctx, app.WorkspaceID, app.ProjectID, pvcSpec); err != nil {
			return err
		}
	}

	statefulSetSpec := domain.StatefulSetSpec{
		Name:         app.Name,
		Replicas:     app.Config.Replicas,
		Image:        app.Source.Image,
		Port:         app.Config.Port,
		EnvVars:      app.Config.EnvVars,
		Resources:    app.Config.Resources,
		NodeSelector: app.Config.NodeSelector,
		Labels: map[string]string{
			"app": app.Name,
		},
	}

	if app.Config.Storage != nil {
		statefulSetSpec.VolumeClaimSpec = domain.PVCSpec{
			Name:         app.Name + "-data",
			Size:         app.Config.Storage.Size,
			StorageClass: app.Config.Storage.StorageClass,
			AccessMode:   "ReadWriteOnce",
		}
	}

	return s.k8s.CreateStatefulSet(ctx, app.WorkspaceID, app.ProjectID, statefulSetSpec)
}

// GetApplication retrieves an application by ID
func (s *Service) GetApplication(ctx context.Context, applicationID string) (*domain.Application, error) {
	return s.repo.GetApplication(ctx, applicationID)
}

// ListApplications lists all applications in a workspace/project
func (s *Service) ListApplications(ctx context.Context, workspaceID, projectID string) ([]domain.Application, error) {
	return s.repo.ListApplications(ctx, workspaceID, projectID)
}

// UpdateApplication updates an application
func (s *Service) UpdateApplication(ctx context.Context, applicationID string, req domain.UpdateApplicationRequest) (*domain.Application, error) {
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	// Check if status allows updates
	if !app.Status.CanTransition(domain.ApplicationStatusUpdating) {
		return nil, fmt.Errorf("cannot update application in status %s", app.Status)
	}

	// Update status
	app.Status = domain.ApplicationStatusUpdating
	app.UpdatedAt = time.Now()

	// Apply updates
	if req.Replicas != nil {
		app.Config.Replicas = *req.Replicas
	}
	if req.ImageVersion != "" {
		app.Source.Image = req.ImageVersion
	}
	if req.EnvVars != nil {
		for k, v := range req.EnvVars {
			app.Config.EnvVars[k] = v
		}
	}
	if req.Resources != nil {
		app.Config.Resources = *req.Resources
	}
	if req.NetworkConfig != nil {
		app.Config.NetworkConfig = req.NetworkConfig
	}

	// Save changes
	if err := s.repo.UpdateApplication(ctx, app); err != nil {
		return nil, err
	}

	// Create update event
	event := &domain.ApplicationEvent{
		ID:            uuid.New().String(),
		ApplicationID: app.ID,
		Type:          "update.started",
		Message:       "Application update started",
		Timestamp:     time.Now(),
	}
	s.repo.CreateEvent(ctx, event)

	// Apply updates to Kubernetes
	go s.updateKubernetesResources(context.Background(), app, req)

	return app, nil
}

func (s *Service) updateKubernetesResources(ctx context.Context, app *domain.Application, req domain.UpdateApplicationRequest) {
	var err error
	defer func() {
		if err != nil {
			app.Status = domain.ApplicationStatusError
		} else {
			app.Status = domain.ApplicationStatusRunning
		}
		s.repo.UpdateApplication(ctx, app)

		eventType := "update.succeeded"
		message := "Application updated successfully"
		if err != nil {
			eventType = "update.failed"
			message = fmt.Sprintf("Update failed: %v", err)
		}

		event := &domain.ApplicationEvent{
			ID:            uuid.New().String(),
			ApplicationID: app.ID,
			Type:          eventType,
			Message:       message,
			Timestamp:     time.Now(),
		}
		s.repo.CreateEvent(ctx, event)
	}()

	// Update deployment or statefulset
	if app.Type == domain.ApplicationTypeStateless {
		deploymentSpec := domain.DeploymentSpec{
			Name:     app.Name,
			Replicas: app.Config.Replicas,
		}
		if req.ImageVersion != "" {
			deploymentSpec.Image = req.ImageVersion
		}
		err = s.k8s.UpdateDeployment(ctx, app.WorkspaceID, app.ProjectID, app.Name, deploymentSpec)
	} else {
		statefulSetSpec := domain.StatefulSetSpec{
			Name:     app.Name,
			Replicas: app.Config.Replicas,
		}
		if req.ImageVersion != "" {
			statefulSetSpec.Image = req.ImageVersion
		}
		err = s.k8s.UpdateStatefulSet(ctx, app.WorkspaceID, app.ProjectID, app.Name, statefulSetSpec)
	}

	// Update ingress if network config changed
	if req.NetworkConfig != nil && req.NetworkConfig.CreateIngress {
		ingressSpec := domain.IngressSpec{
			Name:        app.Name,
			Host:        req.NetworkConfig.CustomDomain,
			Path:        req.NetworkConfig.IngressPath,
			ServiceName: app.Name,
			ServicePort: app.Config.Port,
			TLSEnabled:  req.NetworkConfig.TLSEnabled,
			Annotations: req.NetworkConfig.Annotations,
		}
		s.k8s.UpdateIngress(ctx, app.WorkspaceID, app.ProjectID, app.Name, ingressSpec)
	}
}

// DeleteApplication deletes an application
func (s *Service) DeleteApplication(ctx context.Context, applicationID string) error {
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	// Check if status allows deletion
	if !app.Status.CanTransition(domain.ApplicationStatusDeleting) {
		return fmt.Errorf("cannot delete application in status %s", app.Status)
	}

	// Update status
	app.Status = domain.ApplicationStatusDeleting
	s.repo.UpdateApplication(ctx, app)

	// Create deletion event
	event := &domain.ApplicationEvent{
		ID:            uuid.New().String(),
		ApplicationID: app.ID,
		Type:          "deletion.started",
		Message:       "Application deletion started",
		Timestamp:     time.Now(),
	}
	s.repo.CreateEvent(ctx, event)

	// Delete Kubernetes resources
	if app.Type == domain.ApplicationTypeStateless {
		s.k8s.DeleteDeployment(ctx, app.WorkspaceID, app.ProjectID, app.Name)
	} else {
		s.k8s.DeleteStatefulSet(ctx, app.WorkspaceID, app.ProjectID, app.Name)
		if app.Config.Storage != nil {
			s.k8s.DeletePVC(ctx, app.WorkspaceID, app.ProjectID, app.Name+"-data")
		}
	}

	// Delete service and ingress
	s.k8s.DeleteService(ctx, app.WorkspaceID, app.ProjectID, app.Name)
	s.k8s.DeleteIngress(ctx, app.WorkspaceID, app.ProjectID, app.Name)

	// Delete from database
	return s.repo.DeleteApplication(ctx, applicationID)
}

// StartApplication starts a stopped application
func (s *Service) StartApplication(ctx context.Context, applicationID string) error {
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	if app.Status != domain.ApplicationStatusStopped {
		return fmt.Errorf("application is not in stopped state")
	}

	// Re-deploy the application
	app.Status = domain.ApplicationStatusDeploying
	s.repo.UpdateApplication(ctx, app)

	go s.deployApplication(context.Background(), app)

	return nil
}

// StopApplication stops a running application
func (s *Service) StopApplication(ctx context.Context, applicationID string) error {
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	if !app.Status.CanTransition(domain.ApplicationStatusStopping) {
		return fmt.Errorf("cannot stop application in status %s", app.Status)
	}

	app.Status = domain.ApplicationStatusStopping
	s.repo.UpdateApplication(ctx, app)

	// Scale to 0 replicas
	if app.Type == domain.ApplicationTypeStateless {
		deploymentSpec := domain.DeploymentSpec{
			Name:     app.Name,
			Replicas: 0,
		}
		err = s.k8s.UpdateDeployment(ctx, app.WorkspaceID, app.ProjectID, app.Name, deploymentSpec)
	} else {
		statefulSetSpec := domain.StatefulSetSpec{
			Name:     app.Name,
			Replicas: 0,
		}
		err = s.k8s.UpdateStatefulSet(ctx, app.WorkspaceID, app.ProjectID, app.Name, statefulSetSpec)
	}

	if err != nil {
		app.Status = domain.ApplicationStatusError
	} else {
		app.Status = domain.ApplicationStatusStopped
	}
	s.repo.UpdateApplication(ctx, app)

	return err
}

// RestartApplication restarts an application
func (s *Service) RestartApplication(ctx context.Context, applicationID string) error {
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	if app.Status != domain.ApplicationStatusRunning {
		return fmt.Errorf("can only restart running applications")
	}

	// Get all pods
	pods, err := s.k8s.ListPods(ctx, app.WorkspaceID, app.ProjectID, map[string]string{"app": app.Name})
	if err != nil {
		return err
	}

	// Restart each pod
	for _, pod := range pods {
		if err := s.k8s.RestartPod(ctx, app.WorkspaceID, app.ProjectID, pod.Name); err != nil {
			return err
		}
	}

	// Create restart event
	event := &domain.ApplicationEvent{
		ID:            uuid.New().String(),
		ApplicationID: app.ID,
		Type:          "restart.completed",
		Message:       fmt.Sprintf("Restarted %d pods", len(pods)),
		Timestamp:     time.Now(),
	}
	s.repo.CreateEvent(ctx, event)

	return nil
}

// ScaleApplication scales an application to the specified number of replicas
func (s *Service) ScaleApplication(ctx context.Context, applicationID string, replicas int) error {
	if replicas < 0 {
		return errors.New("invalid replica count")
	}

	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	// Check constraints for stateful apps
	if app.Type == domain.ApplicationTypeStateful && replicas > 1 {
		return errors.New("stateful applications cannot scale beyond 1 replica")
	}

	// Update configuration
	req := domain.UpdateApplicationRequest{
		Replicas: &replicas,
	}

	_, err = s.UpdateApplication(ctx, applicationID, req)
	return err
}

// ListPods lists all pods for an application
func (s *Service) ListPods(ctx context.Context, applicationID string) ([]domain.Pod, error) {
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	selector := map[string]string{"app": app.Name}
	return s.k8s.ListPods(ctx, app.WorkspaceID, app.ProjectID, selector)
}

// RestartPod restarts a specific pod
func (s *Service) RestartPod(ctx context.Context, applicationID, podName string) error {
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	return s.k8s.RestartPod(ctx, app.WorkspaceID, app.ProjectID, podName)
}

// GetPodLogs retrieves logs for an application
func (s *Service) GetPodLogs(ctx context.Context, query domain.LogQuery) ([]domain.LogEntry, error) {
	app, err := s.repo.GetApplication(ctx, query.ApplicationID)
	if err != nil {
		return nil, err
	}

	opts := domain.LogOptions{
		Since:    query.Since,
		Until:    query.Until,
		Limit:    query.Limit,
		Follow:   query.Follow,
		Previous: false,
	}

	return s.k8s.GetPodLogs(ctx, app.WorkspaceID, app.ProjectID, query.PodName, query.Container, opts)
}

// StreamPodLogs streams logs for an application
func (s *Service) StreamPodLogs(ctx context.Context, query domain.LogQuery) (io.ReadCloser, error) {
	app, err := s.repo.GetApplication(ctx, query.ApplicationID)
	if err != nil {
		return nil, err
	}

	opts := domain.LogOptions{
		Since:    query.Since,
		Until:    query.Until,
		Limit:    query.Limit,
		Follow:   query.Follow,
		Previous: false,
	}

	return s.k8s.StreamPodLogs(ctx, app.WorkspaceID, app.ProjectID, query.PodName, query.Container, opts)
}

// GetApplicationMetrics retrieves metrics for an application
func (s *Service) GetApplicationMetrics(ctx context.Context, applicationID string) (*domain.ApplicationMetrics, error) {
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	// Get all pods
	pods, err := s.k8s.ListPods(ctx, app.WorkspaceID, app.ProjectID, map[string]string{"app": app.Name})
	if err != nil {
		return nil, err
	}

	// Get pod names
	podNames := make([]string, len(pods))
	for i, pod := range pods {
		podNames[i] = pod.Name
	}

	// Get metrics
	podMetrics, err := s.k8s.GetPodMetrics(ctx, app.WorkspaceID, app.ProjectID, podNames)
	if err != nil {
		return nil, err
	}

	// Calculate aggregates
	var totalCPU, totalMemory float64
	for _, m := range podMetrics {
		totalCPU += m.CPUUsage
		totalMemory += m.MemoryUsage
	}

	avgCPU := totalCPU
	avgMemory := totalMemory
	if len(podMetrics) > 0 {
		avgCPU = totalCPU / float64(len(podMetrics))
		avgMemory = totalMemory / float64(len(podMetrics))
	}

	return &domain.ApplicationMetrics{
		ApplicationID: applicationID,
		Timestamp:     time.Now(),
		PodMetrics:    podMetrics,
		AggregateUsage: domain.AggregateResourceUsage{
			TotalCPU:      totalCPU,
			TotalMemory:   totalMemory,
			AverageCPU:    avgCPU,
			AverageMemory: avgMemory,
		},
	}, nil
}

// GetApplicationEvents retrieves events for an application
func (s *Service) GetApplicationEvents(ctx context.Context, applicationID string, limit int) ([]domain.ApplicationEvent, error) {
	return s.repo.ListEvents(ctx, applicationID, limit)
}

// UpdateNetworkConfig updates network configuration for an application
func (s *Service) UpdateNetworkConfig(ctx context.Context, applicationID string, config domain.NetworkConfig) error {
	req := domain.UpdateApplicationRequest{
		NetworkConfig: &config,
	}
	_, err := s.UpdateApplication(ctx, applicationID, req)
	return err
}

// GetApplicationEndpoints retrieves endpoints for an application
func (s *Service) GetApplicationEndpoints(ctx context.Context, applicationID string) ([]domain.Endpoint, error) {
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	return s.k8s.GetServiceEndpoints(ctx, app.WorkspaceID, app.ProjectID, app.Name)
}

// UpdateNodeAffinity updates node affinity for an application
func (s *Service) UpdateNodeAffinity(ctx context.Context, applicationID string, nodeSelector map[string]string) error {
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	app.Config.NodeSelector = nodeSelector
	app.UpdatedAt = time.Now()

	return s.repo.UpdateApplication(ctx, app)
}

// MigrateToNode migrates an application to a specific node
func (s *Service) MigrateToNode(ctx context.Context, applicationID, targetNodeID string) error {
	nodeSelector := map[string]string{
		"node-id": targetNodeID,
	}
	return s.UpdateNodeAffinity(ctx, applicationID, nodeSelector)
}

// CreateCronJob creates a new CronJob application
func (s *Service) CreateCronJob(ctx context.Context, app *domain.Application) error {
	// Validate CronJob specific fields
	if app.Type != domain.ApplicationTypeCronJob {
		return errors.New("application type must be cronjob")
	}
	if app.CronSchedule == "" {
		return errors.New("cron schedule is required")
	}
	if len(app.CronCommand) == 0 && app.TemplateAppID == "" {
		return errors.New("cron command or template app ID is required")
	}

	// Set initial status
	app.Status = domain.ApplicationStatusPending
	
	// Create in repository (will handle template app copying)
	if err := s.repo.Create(ctx, app); err != nil {
		return fmt.Errorf("failed to create cronjob application: %w", err)
	}

	// Create CronJob in Kubernetes
	cronJobSpec := domain.CronJobSpec{
		Name:              app.Name,
		Schedule:          app.CronSchedule,
		Image:             app.Source.Image,
		Command:           app.CronCommand,
		Args:              app.CronArgs,
		EnvVars:           app.Config.EnvVars,
		Resources:         app.Config.Resources,
		NodeSelector:      app.Config.NodeSelector,
		Labels:            map[string]string{"app": app.Name, "type": "cronjob"},
		Annotations:       map[string]string{"hexabase.io/app-id": app.ID},
		RestartPolicy:     "OnFailure",
		ConcurrencyPolicy: "Forbid", // Default to forbid concurrent runs
	}

	if err := s.k8s.CreateCronJob(ctx, app.WorkspaceID, app.ProjectID, cronJobSpec); err != nil {
		// Update status to error
		app.Status = domain.ApplicationStatusError
		s.repo.UpdateApplication(ctx, app)
		return fmt.Errorf("failed to create kubernetes cronjob: %w", err)
	}

	// Update status to running
	app.Status = domain.ApplicationStatusRunning
	return s.repo.UpdateApplication(ctx, app)
}

// UpdateCronJobSchedule updates the schedule of a CronJob
func (s *Service) UpdateCronJobSchedule(ctx context.Context, applicationID, newSchedule string) error {
	// Get application
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	if app.Type != domain.ApplicationTypeCronJob {
		return errors.New("application is not a cronjob")
	}

	// Update schedule in repository
	if err := s.repo.UpdateCronSchedule(ctx, applicationID, newSchedule); err != nil {
		return fmt.Errorf("failed to update cron schedule: %w", err)
	}

	// Update CronJob in Kubernetes
	cronJobSpec := domain.CronJobSpec{
		Name:              app.Name,
		Schedule:          newSchedule,
		Image:             app.Source.Image,
		Command:           app.CronCommand,
		Args:              app.CronArgs,
		EnvVars:           app.Config.EnvVars,
		Resources:         app.Config.Resources,
		NodeSelector:      app.Config.NodeSelector,
		Labels:            map[string]string{"app": app.Name, "type": "cronjob"},
		Annotations:       map[string]string{"hexabase.io/app-id": app.ID},
		RestartPolicy:     "OnFailure",
		ConcurrencyPolicy: "Forbid",
	}

	if err := s.k8s.UpdateCronJob(ctx, app.WorkspaceID, app.ProjectID, app.Name, cronJobSpec); err != nil {
		return fmt.Errorf("failed to update kubernetes cronjob: %w", err)
	}

	return nil
}

// TriggerCronJob manually triggers a CronJob
func (s *Service) TriggerCronJob(ctx context.Context, req *domain.TriggerCronJobRequest) (*domain.CronJobExecution, error) {
	// Get application
	app, err := s.repo.GetApplication(ctx, req.ApplicationID)
	if err != nil {
		return nil, err
	}

	if app.Type != domain.ApplicationTypeCronJob {
		return nil, errors.New("application is not a cronjob")
	}

	// Trigger in Kubernetes
	err = s.k8s.TriggerCronJob(ctx, app.WorkspaceID, app.ProjectID, app.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger cronjob: %w", err)
	}

	// Create execution record
	execution := &domain.CronJobExecution{
		ID:            uuid.New().String(),
		ApplicationID: req.ApplicationID,
		JobName:       fmt.Sprintf("%s-manual-%d", app.Name, time.Now().Unix()),
		StartedAt:     time.Now(),
		Status:        domain.CronJobExecutionStatusRunning,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateCronJobExecution(ctx, execution); err != nil {
		// Log error but don't fail the trigger
		if s.logger != nil {
			s.logger.Error("failed to create execution record", "error", err)
		}
	}

	return execution, nil
}

// GetCronJobExecutions retrieves executions for a CronJob
func (s *Service) GetCronJobExecutions(ctx context.Context, applicationID string, limit, offset int) ([]domain.CronJobExecution, int, error) {
	// Verify application exists and is a CronJob
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, 0, err
	}

	if app.Type != domain.ApplicationTypeCronJob {
		return nil, 0, errors.New("application is not a cronjob")
	}

	return s.repo.GetCronJobExecutions(ctx, applicationID, limit, offset)
}

// GetCronJobStatus retrieves the status of a CronJob from Kubernetes
func (s *Service) GetCronJobStatus(ctx context.Context, applicationID string) (*domain.CronJobStatus, error) {
	// Get application
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	if app.Type != domain.ApplicationTypeCronJob {
		return nil, errors.New("application is not a cronjob")
	}

	return s.k8s.GetCronJobStatus(ctx, app.WorkspaceID, app.ProjectID, app.Name)
}

// CreateFunction creates a new serverless function application
func (s *Service) CreateFunction(ctx context.Context, workspaceID string, req domain.CreateFunctionRequest) (*domain.Application, error) {
	// Validate request
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.ProjectID == "" {
		return nil, errors.New("project ID is required")
	}
	if req.Handler == "" {
		return nil, errors.New("handler is required")
	}
	if req.SourceCode == "" {
		return nil, errors.New("source code is required")
	}

	// Set defaults
	if req.Timeout == 0 {
		req.Timeout = 300 // 5 minutes default
	}
	if req.Memory == 0 {
		req.Memory = 256 // 256MB default
	}

	// Create the application
	app := &domain.Application{
		WorkspaceID:         workspaceID,
		ProjectID:           req.ProjectID,
		Name:                req.Name,
		Type:                domain.ApplicationTypeFunction,
		Status:              domain.ApplicationStatusPending,
		FunctionRuntime:     req.Runtime,
		FunctionHandler:     req.Handler,
		FunctionTimeout:     req.Timeout,
		FunctionMemory:      req.Memory,
		FunctionTriggerType: req.TriggerType,
		FunctionTriggerConfig: req.TriggerConfig,
		FunctionEnvVars:     req.EnvVars,
		FunctionSecrets:     req.Secrets,
		Source: domain.ApplicationSource{
			Type: domain.SourceTypeImage, // Will be built from source
		},
		Config: domain.ApplicationConfig{
			Replicas: 0, // Scale to zero when idle
			Resources: domain.ResourceRequests{
				CPURequest:    "100m",
				CPULimit:      "1000m",
				MemoryRequest: fmt.Sprintf("%dMi", req.Memory),
				MemoryLimit:   fmt.Sprintf("%dMi", req.Memory*2),
			},
		},
	}

	// Create application record
	if err := s.repo.CreateApplication(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to create function: %w", err)
	}

	// Create initial version
	version := &domain.FunctionVersion{
		ApplicationID: app.ID,
		VersionNumber: 1,
		SourceCode:    req.SourceCode,
		SourceType:    req.SourceType,
		SourceURL:     req.SourceURL,
		BuildStatus:   domain.FunctionBuildPending,
		IsActive:      true,
	}

	if err := s.repo.CreateFunctionVersion(ctx, version); err != nil {
		// Rollback application creation
		s.repo.DeleteApplication(ctx, app.ID)
		return nil, fmt.Errorf("failed to create function version: %w", err)
	}

	// TODO: Trigger build process asynchronously

	return app, nil
}

// DeployFunctionVersion creates and deploys a new version of a function
func (s *Service) DeployFunctionVersion(ctx context.Context, applicationID string, sourceCode string) (*domain.FunctionVersion, error) {
	// Get application
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	if app.Type != domain.ApplicationTypeFunction {
		return nil, errors.New("application is not a function")
	}

	// Get existing versions to determine next version number
	versions, err := s.repo.GetFunctionVersions(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	nextVersionNumber := 1
	if len(versions) > 0 {
		nextVersionNumber = versions[0].VersionNumber + 1
	}

	// Create new version
	version := &domain.FunctionVersion{
		ApplicationID: applicationID,
		VersionNumber: nextVersionNumber,
		SourceCode:    sourceCode,
		SourceType:    domain.FunctionSourceInline,
		BuildStatus:   domain.FunctionBuildPending,
		IsActive:      false, // Not active until successfully built
	}

	if err := s.repo.CreateFunctionVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// Start build process
	// TODO: Make this configurable for testing
	if s.logger != nil {
		go s.buildFunctionVersion(context.Background(), app, version)
	}

	return version, nil
}

// buildFunctionVersion handles the asynchronous build process
func (s *Service) buildFunctionVersion(ctx context.Context, app *domain.Application, version *domain.FunctionVersion) {
	// Update status to building
	version.BuildStatus = domain.FunctionBuildBuilding
	if err := s.repo.UpdateFunctionVersion(ctx, version); err != nil {
		s.logger.Error("failed to update build status", "error", err)
		return
	}

	// TODO: Implement actual build process
	// For now, simulate a successful build
	imageURI := fmt.Sprintf("registry.local/functions/%s:v%d", app.Name, version.VersionNumber)
	
	// Update with build results
	version.BuildStatus = domain.FunctionBuildSuccess
	version.ImageURI = imageURI
	if err := s.repo.UpdateFunctionVersion(ctx, version); err != nil {
		s.logger.Error("failed to update build results", "error", err)
		return
	}

	// If this is the first version, make it active
	activeVersion, _ := s.repo.GetActiveFunctionVersion(ctx, app.ID)
	if activeVersion == nil {
		s.SetActiveFunctionVersion(ctx, app.ID, version.ID)
	}
}

// GetFunctionVersions retrieves all versions of a function
func (s *Service) GetFunctionVersions(ctx context.Context, applicationID string) ([]domain.FunctionVersion, error) {
	// Verify application exists and is a function
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	if app.Type != domain.ApplicationTypeFunction {
		return nil, errors.New("application is not a function")
	}

	return s.repo.GetFunctionVersions(ctx, applicationID)
}

// SetActiveFunctionVersion sets the active version for a function
func (s *Service) SetActiveFunctionVersion(ctx context.Context, applicationID, versionID string) error {
	// Verify application exists and is a function
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	if app.Type != domain.ApplicationTypeFunction {
		return errors.New("application is not a function")
	}

	// Verify version exists and is built
	version, err := s.repo.GetFunctionVersion(ctx, versionID)
	if err != nil {
		return err
	}

	if version.ApplicationID != applicationID {
		return errors.New("version does not belong to this application")
	}

	if version.BuildStatus != domain.FunctionBuildSuccess {
		return errors.New("can only activate successfully built versions")
	}

	// Set active version
	if err := s.repo.SetActiveFunctionVersion(ctx, applicationID, versionID); err != nil {
		return err
	}

	// Deploy the new version to Knative
	spec := domain.KnativeServiceSpec{
		Name:         app.Name,
		Image:        version.ImageURI,
		EnvVars:      app.FunctionEnvVars,
		Secrets:      app.FunctionSecrets,
		Resources:    app.Config.Resources,
		Labels:       map[string]string{"app": app.Name, "version": fmt.Sprintf("v%d", version.VersionNumber)},
		TimeoutSeconds: app.FunctionTimeout,
	}

	// Update or create Knative service
	if app.Status == domain.ApplicationStatusRunning {
		return s.k8s.UpdateKnativeService(ctx, app.WorkspaceID, app.ProjectID, app.Name, spec)
	} else {
		if err := s.k8s.CreateKnativeService(ctx, app.WorkspaceID, app.ProjectID, spec); err != nil {
			return err
		}
		// Update app status
		app.Status = domain.ApplicationStatusRunning
		return s.repo.UpdateApplication(ctx, app)
	}
}

// InvokeFunction invokes a function synchronously
func (s *Service) InvokeFunction(ctx context.Context, applicationID string, req domain.InvokeFunctionRequest) (*domain.InvokeFunctionResponse, error) {
	// Get application
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	if app.Type != domain.ApplicationTypeFunction {
		return nil, errors.New("application is not a function")
	}

	if app.Status != domain.ApplicationStatusRunning {
		return nil, errors.New("function is not running")
	}

	// Get active version
	activeVersion, err := s.repo.GetActiveFunctionVersion(ctx, applicationID)
	if err != nil || activeVersion == nil {
		return nil, errors.New("no active version found")
	}

	// Get function URL
	_, err = s.k8s.GetKnativeServiceURL(ctx, app.WorkspaceID, app.ProjectID, app.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get function URL: %w", err)
	}

	// Create invocation record
	invocation := &domain.FunctionInvocation{
		ApplicationID:  applicationID,
		VersionID:      activeVersion.ID,
		InvocationID:   uuid.New().String(),
		TriggerSource:  "http",
		RequestMethod:  req.Method,
		RequestPath:    req.Path,
		RequestHeaders: req.Headers,
		RequestBody:    req.Body,
		StartedAt:      time.Now(),
	}

	if err := s.repo.CreateFunctionInvocation(ctx, invocation); err != nil {
		// Log but don't fail
		s.logger.Error("failed to create invocation record", "error", err)
	}

	// TODO: Actually invoke the function via HTTP
	// For now, return a mock response
	response := &domain.InvokeFunctionResponse{
		InvocationID: invocation.InvocationID,
		Status:       200,
		Headers:      map[string][]string{"Content-Type": {"application/json"}},
		Body:         base64.StdEncoding.EncodeToString([]byte(`{"message": "Hello from function"}`)),
		DurationMs:   100,
		ColdStart:    false,
	}

	// Update invocation record
	completedAt := time.Now()
	invocation.ResponseStatus = response.Status
	invocation.ResponseHeaders = response.Headers
	invocation.ResponseBody = response.Body
	invocation.DurationMs = response.DurationMs
	invocation.CompletedAt = &completedAt
	
	if err := s.repo.UpdateFunctionInvocation(ctx, invocation); err != nil {
		s.logger.Error("failed to update invocation record", "error", err)
	}

	return response, nil
}

// GetFunctionInvocations retrieves invocation history for a function
func (s *Service) GetFunctionInvocations(ctx context.Context, applicationID string, limit, offset int) ([]domain.FunctionInvocation, int, error) {
	// Verify application exists and is a function
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, 0, err
	}

	if app.Type != domain.ApplicationTypeFunction {
		return nil, 0, errors.New("application is not a function")
	}

	return s.repo.GetFunctionInvocations(ctx, applicationID, limit, offset)
}

// GetFunctionEvents retrieves pending events for a function
func (s *Service) GetFunctionEvents(ctx context.Context, applicationID string, limit int) ([]domain.FunctionEvent, error) {
	// Verify application exists and is a function
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	if app.Type != domain.ApplicationTypeFunction {
		return nil, errors.New("application is not a function")
	}

	return s.repo.GetPendingFunctionEvents(ctx, applicationID, limit)
}

// ProcessFunctionEvent processes a function event
func (s *Service) ProcessFunctionEvent(ctx context.Context, eventID string) error {
	// Get event
	event, err := s.repo.GetFunctionEvent(ctx, eventID)
	if err != nil {
		return err
	}

	// Check if already processed
	if event.ProcessingStatus == "success" {
		return nil
	}

	// Update status to processing
	event.ProcessingStatus = "processing"
	if err := s.repo.UpdateFunctionEvent(ctx, event); err != nil {
		return err
	}

	// Get application
	app, err := s.repo.GetApplication(ctx, event.ApplicationID)
	if err != nil {
		return s.handleEventError(ctx, event, err)
	}

	// Prepare invocation request
	eventData, _ := json.Marshal(event.EventData)
	req := domain.InvokeFunctionRequest{
		Method: "POST",
		Path:   "/event",
		Headers: map[string][]string{
			"X-Event-Type":   {event.EventType},
			"X-Event-Source": {event.EventSource},
			"X-Event-ID":     {event.ID},
		},
		Body: base64.StdEncoding.EncodeToString(eventData),
	}

	// Invoke function
	resp, err := s.InvokeFunction(ctx, app.ID, req)
	if err != nil {
		return s.handleEventError(ctx, event, err)
	}

	// Update event as processed
	now := time.Now()
	event.ProcessingStatus = "success"
	event.InvocationID = resp.InvocationID
	event.ProcessedAt = &now
	
	return s.repo.UpdateFunctionEvent(ctx, event)
}

// handleEventError handles errors during event processing
func (s *Service) handleEventError(ctx context.Context, event *domain.FunctionEvent, err error) error {
	event.ErrorMessage = err.Error()
	event.RetryCount++

	if event.RetryCount >= event.MaxRetries {
		event.ProcessingStatus = "failed"
	} else {
		event.ProcessingStatus = "retry"
	}

	return s.repo.UpdateFunctionEvent(ctx, event)
}

// UpdateCronJobExecutionStatus updates the status of a CronJob execution
func (s *Service) UpdateCronJobExecutionStatus(ctx context.Context, executionID string, status domain.CronJobExecutionStatus) error {
	execution, err := s.repo.GetCronJobExecution(ctx, executionID)
	if err != nil {
		return err
	}

	execution.Status = status
	execution.UpdatedAt = time.Now()

	var completedAt *time.Time
	var exitCode *int
	if status == domain.CronJobExecutionStatusSucceeded || status == domain.CronJobExecutionStatusFailed {
		now := time.Now()
		completedAt = &now
		execution.CompletedAt = completedAt
		
		if status == domain.CronJobExecutionStatusSucceeded {
			code := 0
			exitCode = &code
		} else {
			code := 1
			exitCode = &code
		}
	}

	return s.repo.UpdateCronJobExecution(ctx, executionID, completedAt, status, exitCode, "")
}