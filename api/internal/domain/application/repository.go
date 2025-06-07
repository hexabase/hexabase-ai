package application

import (
	"context"
	"io"
	"time"
)

// Repository defines the interface for application data persistence
type Repository interface {
	// Application CRUD operations
	CreateApplication(ctx context.Context, app *Application) error
	GetApplication(ctx context.Context, id string) (*Application, error)
	GetApplicationByName(ctx context.Context, workspaceID, projectID, name string) (*Application, error)
	ListApplications(ctx context.Context, workspaceID, projectID string) ([]Application, error)
	UpdateApplication(ctx context.Context, app *Application) error
	DeleteApplication(ctx context.Context, id string) error

	// Application event operations
	CreateEvent(ctx context.Context, event *ApplicationEvent) error
	ListEvents(ctx context.Context, applicationID string, limit int) ([]ApplicationEvent, error)

	// Query operations
	GetApplicationsByNode(ctx context.Context, nodeID string) ([]Application, error)
	GetApplicationsByStatus(ctx context.Context, workspaceID string, status ApplicationStatus) ([]Application, error)
}

// KubernetesRepository defines the interface for Kubernetes operations
type KubernetesRepository interface {
	// Deployment operations
	CreateDeployment(ctx context.Context, workspaceID, projectID string, spec DeploymentSpec) error
	UpdateDeployment(ctx context.Context, workspaceID, projectID, name string, spec DeploymentSpec) error
	DeleteDeployment(ctx context.Context, workspaceID, projectID, name string) error
	GetDeploymentStatus(ctx context.Context, workspaceID, projectID, name string) (*DeploymentStatus, error)

	// StatefulSet operations (for stateful apps)
	CreateStatefulSet(ctx context.Context, workspaceID, projectID string, spec StatefulSetSpec) error
	UpdateStatefulSet(ctx context.Context, workspaceID, projectID, name string, spec StatefulSetSpec) error
	DeleteStatefulSet(ctx context.Context, workspaceID, projectID, name string) error
	GetStatefulSetStatus(ctx context.Context, workspaceID, projectID, name string) (*StatefulSetStatus, error)

	// Service operations
	CreateService(ctx context.Context, workspaceID, projectID string, spec ServiceSpec) error
	DeleteService(ctx context.Context, workspaceID, projectID, name string) error
	GetServiceEndpoints(ctx context.Context, workspaceID, projectID, name string) ([]Endpoint, error)

	// Ingress operations
	CreateIngress(ctx context.Context, workspaceID, projectID string, spec IngressSpec) error
	UpdateIngress(ctx context.Context, workspaceID, projectID, name string, spec IngressSpec) error
	DeleteIngress(ctx context.Context, workspaceID, projectID, name string) error

	// PVC operations (for stateful apps)
	CreatePVC(ctx context.Context, workspaceID, projectID string, spec PVCSpec) error
	DeletePVC(ctx context.Context, workspaceID, projectID, name string) error

	// Pod operations
	ListPods(ctx context.Context, workspaceID, projectID string, selector map[string]string) ([]Pod, error)
	GetPodLogs(ctx context.Context, workspaceID, projectID, podName, container string, opts LogOptions) ([]LogEntry, error)
	StreamPodLogs(ctx context.Context, workspaceID, projectID, podName, container string, opts LogOptions) (io.ReadCloser, error)
	RestartPod(ctx context.Context, workspaceID, projectID, podName string) error

	// Metrics operations
	GetPodMetrics(ctx context.Context, workspaceID, projectID string, podNames []string) ([]PodMetrics, error)
}

// DeploymentSpec represents the specification for a Kubernetes Deployment
type DeploymentSpec struct {
	Name         string
	Replicas     int
	Image        string
	Port         int
	EnvVars      map[string]string
	Resources    ResourceRequests
	NodeSelector map[string]string
	Labels       map[string]string
	Annotations  map[string]string
}

// StatefulSetSpec represents the specification for a Kubernetes StatefulSet
type StatefulSetSpec struct {
	Name             string
	Replicas         int
	Image            string
	Port             int
	EnvVars          map[string]string
	Resources        ResourceRequests
	NodeSelector     map[string]string
	Labels           map[string]string
	Annotations      map[string]string
	VolumeClaimSpec  PVCSpec
}

// ServiceSpec represents the specification for a Kubernetes Service
type ServiceSpec struct {
	Name        string
	Port        int
	TargetPort  int
	Selector    map[string]string
	Type        string // ClusterIP, NodePort, LoadBalancer
}

// IngressSpec represents the specification for a Kubernetes Ingress
type IngressSpec struct {
	Name        string
	Host        string
	Path        string
	ServiceName string
	ServicePort int
	TLSEnabled  bool
	Annotations map[string]string
}

// PVCSpec represents the specification for a PersistentVolumeClaim
type PVCSpec struct {
	Name         string
	Size         string
	StorageClass string
	AccessMode   string // ReadWriteOnce, ReadOnlyMany, ReadWriteMany
}

// DeploymentStatus represents the status of a Kubernetes Deployment
type DeploymentStatus struct {
	Replicas          int
	UpdatedReplicas   int
	ReadyReplicas     int
	AvailableReplicas int
	Conditions        []DeploymentCondition
}

// StatefulSetStatus represents the status of a Kubernetes StatefulSet
type StatefulSetStatus struct {
	Replicas        int
	ReadyReplicas   int
	CurrentReplicas int
	UpdatedReplicas int
	Conditions      []StatefulSetCondition
}

// DeploymentCondition represents a condition of a Deployment
type DeploymentCondition struct {
	Type               string
	Status             string
	LastUpdateTime     time.Time
	LastTransitionTime time.Time
	Reason             string
	Message            string
}

// StatefulSetCondition represents a condition of a StatefulSet
type StatefulSetCondition struct {
	Type               string
	Status             string
	LastTransitionTime time.Time
	Reason             string
	Message            string
}

// LogOptions represents options for retrieving logs
type LogOptions struct {
	Since     time.Time
	Until     time.Time
	Limit     int
	Follow    bool
	Previous  bool
}