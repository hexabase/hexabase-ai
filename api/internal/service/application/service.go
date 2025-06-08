package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/domain/application"
)

// Service implements the application service interface
type Service struct {
	repo    application.Repository
	k8s     application.KubernetesRepository
	k8sRepo application.KubernetesRepository // Alias for backward compatibility
}

// NewService creates a new application service
func NewService(repo application.Repository, k8s application.KubernetesRepository) application.Service {
	return &Service{
		repo:    repo,
		k8s:     k8s,
		k8sRepo: k8s, // Set alias
	}
}

// CreateApplication creates a new application
func (s *Service) CreateApplication(ctx context.Context, workspaceID string, req application.CreateApplicationRequest) (*application.Application, error) {
	// Validate request
	if !req.Type.IsValid() {
		return nil, errors.New("invalid application type")
	}
	if !req.Source.Type.IsValid() {
		return nil, errors.New("invalid source type")
	}
	if req.Source.Type == application.SourceTypeImage && req.Source.Image == "" {
		return nil, errors.New("image is required for image source type")
	}
	if req.Source.Type == application.SourceTypeGit && req.Source.GitURL == "" {
		return nil, errors.New("git URL is required for git source type")
	}

	// Check if application name already exists
	existing, _ := s.repo.GetApplicationByName(ctx, workspaceID, req.ProjectID, req.Name)
	if existing != nil {
		return nil, fmt.Errorf("application with name %s already exists", req.Name)
	}

	// Create application entity
	app := &application.Application{
		ID:          uuid.New().String(),
		WorkspaceID: workspaceID,
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Type:        req.Type,
		Status:      application.ApplicationStatusPending,
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
	if req.Type == application.ApplicationTypeCronJob {
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
	event := &application.ApplicationEvent{
		ID:            uuid.New().String(),
		ApplicationID: app.ID,
		Type:          "deployment.started",
		Message:       "Application deployment started",
		Timestamp:     time.Now(),
	}
	s.repo.CreateEvent(ctx, event)

	// Update status to deploying
	app.Status = application.ApplicationStatusDeploying
	s.repo.UpdateApplication(ctx, app)

	// Deploy to Kubernetes
	go s.deployApplication(context.Background(), app)

	return app, nil
}

// deployApplication handles the actual deployment to Kubernetes
func (s *Service) deployApplication(ctx context.Context, app *application.Application) {
	var err error
	defer func() {
		if err != nil {
			app.Status = application.ApplicationStatusError
			s.repo.UpdateApplication(ctx, app)
			event := &application.ApplicationEvent{
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
	case application.ApplicationTypeStateless:
		err = s.deployStatelessApp(ctx, app)
	case application.ApplicationTypeStateful:
		err = s.deployStatefulApp(ctx, app)
	case application.ApplicationTypeCronJob:
		// CronJob deployment is handled separately through CreateCronJob
		app.Status = application.ApplicationStatusRunning
		s.repo.UpdateApplication(ctx, app)
		return
	case application.ApplicationTypeFunction:
		// Function deployment will be implemented later
		err = errors.New("function type not yet implemented")
	default:
		err = fmt.Errorf("unknown application type: %s", app.Type)
	}

	if err != nil {
		return
	}

	// Create service
	serviceSpec := application.ServiceSpec{
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
		ingressSpec := application.IngressSpec{
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
			event := &application.ApplicationEvent{
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
	app.Status = application.ApplicationStatusRunning
	s.repo.UpdateApplication(ctx, app)

	// Get endpoints
	endpoints, _ := s.k8s.GetServiceEndpoints(ctx, app.WorkspaceID, app.ProjectID, app.Name)
	app.Endpoints = endpoints
	s.repo.UpdateApplication(ctx, app)

	// Create success event
	event := &application.ApplicationEvent{
		ID:            uuid.New().String(),
		ApplicationID: app.ID,
		Type:          "deployment.succeeded",
		Message:       "Application deployed successfully",
		Timestamp:     time.Now(),
	}
	s.repo.CreateEvent(ctx, event)
}

func (s *Service) deployStatelessApp(ctx context.Context, app *application.Application) error {
	deploymentSpec := application.DeploymentSpec{
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

func (s *Service) deployStatefulApp(ctx context.Context, app *application.Application) error {
	// Create PVC first if storage is configured
	if app.Config.Storage != nil {
		pvcSpec := application.PVCSpec{
			Name:         app.Name + "-data",
			Size:         app.Config.Storage.Size,
			StorageClass: app.Config.Storage.StorageClass,
			AccessMode:   "ReadWriteOnce",
		}
		if err := s.k8s.CreatePVC(ctx, app.WorkspaceID, app.ProjectID, pvcSpec); err != nil {
			return err
		}
	}

	statefulSetSpec := application.StatefulSetSpec{
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
		statefulSetSpec.VolumeClaimSpec = application.PVCSpec{
			Name:         app.Name + "-data",
			Size:         app.Config.Storage.Size,
			StorageClass: app.Config.Storage.StorageClass,
			AccessMode:   "ReadWriteOnce",
		}
	}

	return s.k8s.CreateStatefulSet(ctx, app.WorkspaceID, app.ProjectID, statefulSetSpec)
}

// GetApplication retrieves an application by ID
func (s *Service) GetApplication(ctx context.Context, applicationID string) (*application.Application, error) {
	return s.repo.GetApplication(ctx, applicationID)
}

// ListApplications lists all applications in a workspace/project
func (s *Service) ListApplications(ctx context.Context, workspaceID, projectID string) ([]application.Application, error) {
	return s.repo.ListApplications(ctx, workspaceID, projectID)
}

// UpdateApplication updates an application
func (s *Service) UpdateApplication(ctx context.Context, applicationID string, req application.UpdateApplicationRequest) (*application.Application, error) {
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	// Check if status allows updates
	if !app.Status.CanTransition(application.ApplicationStatusUpdating) {
		return nil, fmt.Errorf("cannot update application in status %s", app.Status)
	}

	// Update status
	app.Status = application.ApplicationStatusUpdating
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
	event := &application.ApplicationEvent{
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

func (s *Service) updateKubernetesResources(ctx context.Context, app *application.Application, req application.UpdateApplicationRequest) {
	var err error
	defer func() {
		if err != nil {
			app.Status = application.ApplicationStatusError
		} else {
			app.Status = application.ApplicationStatusRunning
		}
		s.repo.UpdateApplication(ctx, app)

		eventType := "update.succeeded"
		message := "Application updated successfully"
		if err != nil {
			eventType = "update.failed"
			message = fmt.Sprintf("Update failed: %v", err)
		}

		event := &application.ApplicationEvent{
			ID:            uuid.New().String(),
			ApplicationID: app.ID,
			Type:          eventType,
			Message:       message,
			Timestamp:     time.Now(),
		}
		s.repo.CreateEvent(ctx, event)
	}()

	// Update deployment or statefulset
	if app.Type == application.ApplicationTypeStateless {
		deploymentSpec := application.DeploymentSpec{
			Name:     app.Name,
			Replicas: app.Config.Replicas,
		}
		if req.ImageVersion != "" {
			deploymentSpec.Image = req.ImageVersion
		}
		err = s.k8s.UpdateDeployment(ctx, app.WorkspaceID, app.ProjectID, app.Name, deploymentSpec)
	} else {
		statefulSetSpec := application.StatefulSetSpec{
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
		ingressSpec := application.IngressSpec{
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
	if !app.Status.CanTransition(application.ApplicationStatusDeleting) {
		return fmt.Errorf("cannot delete application in status %s", app.Status)
	}

	// Update status
	app.Status = application.ApplicationStatusDeleting
	s.repo.UpdateApplication(ctx, app)

	// Create deletion event
	event := &application.ApplicationEvent{
		ID:            uuid.New().String(),
		ApplicationID: app.ID,
		Type:          "deletion.started",
		Message:       "Application deletion started",
		Timestamp:     time.Now(),
	}
	s.repo.CreateEvent(ctx, event)

	// Delete Kubernetes resources
	if app.Type == application.ApplicationTypeStateless {
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

	if app.Status != application.ApplicationStatusStopped {
		return fmt.Errorf("application is not in stopped state")
	}

	// Re-deploy the application
	app.Status = application.ApplicationStatusDeploying
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

	if !app.Status.CanTransition(application.ApplicationStatusStopping) {
		return fmt.Errorf("cannot stop application in status %s", app.Status)
	}

	app.Status = application.ApplicationStatusStopping
	s.repo.UpdateApplication(ctx, app)

	// Scale to 0 replicas
	if app.Type == application.ApplicationTypeStateless {
		deploymentSpec := application.DeploymentSpec{
			Name:     app.Name,
			Replicas: 0,
		}
		err = s.k8s.UpdateDeployment(ctx, app.WorkspaceID, app.ProjectID, app.Name, deploymentSpec)
	} else {
		statefulSetSpec := application.StatefulSetSpec{
			Name:     app.Name,
			Replicas: 0,
		}
		err = s.k8s.UpdateStatefulSet(ctx, app.WorkspaceID, app.ProjectID, app.Name, statefulSetSpec)
	}

	if err != nil {
		app.Status = application.ApplicationStatusError
	} else {
		app.Status = application.ApplicationStatusStopped
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

	if app.Status != application.ApplicationStatusRunning {
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
	event := &application.ApplicationEvent{
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
	if app.Type == application.ApplicationTypeStateful && replicas > 1 {
		return errors.New("stateful applications cannot scale beyond 1 replica")
	}

	// Update configuration
	req := application.UpdateApplicationRequest{
		Replicas: &replicas,
	}

	_, err = s.UpdateApplication(ctx, applicationID, req)
	return err
}

// ListPods lists all pods for an application
func (s *Service) ListPods(ctx context.Context, applicationID string) ([]application.Pod, error) {
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
func (s *Service) GetPodLogs(ctx context.Context, query application.LogQuery) ([]application.LogEntry, error) {
	app, err := s.repo.GetApplication(ctx, query.ApplicationID)
	if err != nil {
		return nil, err
	}

	opts := application.LogOptions{
		Since:    query.Since,
		Until:    query.Until,
		Limit:    query.Limit,
		Follow:   query.Follow,
		Previous: false,
	}

	return s.k8s.GetPodLogs(ctx, app.WorkspaceID, app.ProjectID, query.PodName, query.Container, opts)
}

// StreamPodLogs streams logs for an application
func (s *Service) StreamPodLogs(ctx context.Context, query application.LogQuery) (io.ReadCloser, error) {
	app, err := s.repo.GetApplication(ctx, query.ApplicationID)
	if err != nil {
		return nil, err
	}

	opts := application.LogOptions{
		Since:    query.Since,
		Until:    query.Until,
		Limit:    query.Limit,
		Follow:   query.Follow,
		Previous: false,
	}

	return s.k8s.StreamPodLogs(ctx, app.WorkspaceID, app.ProjectID, query.PodName, query.Container, opts)
}

// GetApplicationMetrics retrieves metrics for an application
func (s *Service) GetApplicationMetrics(ctx context.Context, applicationID string) (*application.ApplicationMetrics, error) {
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

	return &application.ApplicationMetrics{
		ApplicationID: applicationID,
		Timestamp:     time.Now(),
		PodMetrics:    podMetrics,
		AggregateUsage: application.AggregateResourceUsage{
			TotalCPU:      totalCPU,
			TotalMemory:   totalMemory,
			AverageCPU:    avgCPU,
			AverageMemory: avgMemory,
		},
	}, nil
}

// GetApplicationEvents retrieves events for an application
func (s *Service) GetApplicationEvents(ctx context.Context, applicationID string, limit int) ([]application.ApplicationEvent, error) {
	return s.repo.ListEvents(ctx, applicationID, limit)
}

// UpdateNetworkConfig updates network configuration for an application
func (s *Service) UpdateNetworkConfig(ctx context.Context, applicationID string, config application.NetworkConfig) error {
	req := application.UpdateApplicationRequest{
		NetworkConfig: &config,
	}
	_, err := s.UpdateApplication(ctx, applicationID, req)
	return err
}

// GetApplicationEndpoints retrieves endpoints for an application
func (s *Service) GetApplicationEndpoints(ctx context.Context, applicationID string) ([]application.Endpoint, error) {
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
func (s *Service) CreateCronJob(ctx context.Context, app *application.Application) error {
	// Validate CronJob specific fields
	if app.Type != application.ApplicationTypeCronJob {
		return errors.New("application type must be cronjob")
	}
	if app.CronSchedule == "" {
		return errors.New("cron schedule is required")
	}
	if len(app.CronCommand) == 0 && app.TemplateAppID == "" {
		return errors.New("cron command or template app ID is required")
	}

	// Set initial status
	app.Status = application.ApplicationStatusPending
	
	// Create in repository (will handle template app copying)
	if err := s.repo.Create(ctx, app); err != nil {
		return fmt.Errorf("failed to create cronjob application: %w", err)
	}

	// Create CronJob in Kubernetes
	cronJobSpec := application.CronJobSpec{
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
		app.Status = application.ApplicationStatusError
		s.repo.UpdateApplication(ctx, app)
		return fmt.Errorf("failed to create kubernetes cronjob: %w", err)
	}

	// Update status to running
	app.Status = application.ApplicationStatusRunning
	return s.repo.UpdateApplication(ctx, app)
}

// UpdateCronJobSchedule updates the schedule of a CronJob
func (s *Service) UpdateCronJobSchedule(ctx context.Context, applicationID, newSchedule string) error {
	// Get application
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	if app.Type != application.ApplicationTypeCronJob {
		return errors.New("application is not a cronjob")
	}

	// Update schedule in repository
	if err := s.repo.UpdateCronSchedule(ctx, applicationID, newSchedule); err != nil {
		return fmt.Errorf("failed to update cron schedule: %w", err)
	}

	// Update CronJob in Kubernetes
	cronJobSpec := application.CronJobSpec{
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
func (s *Service) TriggerCronJob(ctx context.Context, applicationID string) error {
	// Get application
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return err
	}

	if app.Type != application.ApplicationTypeCronJob {
		return errors.New("application is not a cronjob")
	}

	// Trigger in Kubernetes
	if err := s.k8s.TriggerCronJob(ctx, app.WorkspaceID, app.ProjectID, app.Name); err != nil {
		return fmt.Errorf("failed to trigger cronjob: %w", err)
	}

	// Create execution record
	execution := &application.CronJobExecution{
		ID:            uuid.New().String(),
		ApplicationID: applicationID,
		JobName:       fmt.Sprintf("%s-manual-%d", app.Name, time.Now().Unix()),
		StartedAt:     time.Now(),
		Status:        application.CronJobExecutionStatusRunning,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.CreateCronJobExecution(ctx, execution); err != nil {
		// Log error but don't fail the trigger
		fmt.Printf("failed to create execution record: %v\n", err)
	}

	return nil
}

// GetCronJobExecutions retrieves executions for a CronJob
func (s *Service) GetCronJobExecutions(ctx context.Context, applicationID string, limit, offset int) ([]application.CronJobExecution, int, error) {
	// Verify application exists and is a CronJob
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, 0, err
	}

	if app.Type != application.ApplicationTypeCronJob {
		return nil, 0, errors.New("application is not a cronjob")
	}

	return s.repo.GetCronJobExecutions(ctx, applicationID, limit, offset)
}

// GetCronJobStatus retrieves the status of a CronJob from Kubernetes
func (s *Service) GetCronJobStatus(ctx context.Context, applicationID string) (*application.CronJobStatus, error) {
	// Get application
	app, err := s.repo.GetApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	if app.Type != application.ApplicationTypeCronJob {
		return nil, errors.New("application is not a cronjob")
	}

	return s.k8s.GetCronJobStatus(ctx, app.WorkspaceID, app.ProjectID, app.Name)
}