package kubernetes

import (
	"context"
)

// Repository defines the interface for Kubernetes cluster operations
type Repository interface {
	// GetPodMetrics retrieves metrics for pods in a namespace
	GetPodMetrics(ctx context.Context, namespace string) (*PodMetricsList, error)
	
	// GetNodeMetrics retrieves metrics for nodes
	GetNodeMetrics(ctx context.Context) (*NodeMetricsList, error)
	
	// GetNamespaceResourceQuota gets resource quota for a namespace
	GetNamespaceResourceQuota(ctx context.Context, namespace string) (*ResourceQuota, error)
	
	// GetClusterInfo retrieves general cluster information
	GetClusterInfo(ctx context.Context) (*ClusterInfo, error)
	
	// CheckComponentHealth checks health of cluster components
	CheckComponentHealth(ctx context.Context) (map[string]ComponentStatus, error)
}

// PodMetricsList contains metrics for multiple pods
type PodMetricsList struct {
	Items []PodMetrics `json:"items"`
}

// PodMetrics contains resource usage metrics for a pod
type PodMetrics struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	CPU       ResourceMetric    `json:"cpu"`
	Memory    ResourceMetric    `json:"memory"`
	Timestamp string            `json:"timestamp"`
}

// ResourceMetric represents a resource metric value
type ResourceMetric struct {
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

// NodeMetricsList contains metrics for multiple nodes
type NodeMetricsList struct {
	Items []NodeMetrics `json:"items"`
}

// NodeMetrics contains resource usage metrics for a node
type NodeMetrics struct {
	Name      string         `json:"name"`
	CPU       ResourceMetric `json:"cpu"`
	Memory    ResourceMetric `json:"memory"`
	Storage   ResourceMetric `json:"storage"`
	Timestamp string         `json:"timestamp"`
}

// ResourceQuota represents namespace resource quota
type ResourceQuota struct {
	Name      string                       `json:"name"`
	Namespace string                       `json:"namespace"`
	Hard      map[string]string            `json:"hard"`
	Used      map[string]string            `json:"used"`
}

// ClusterInfo contains general cluster information
type ClusterInfo struct {
	Version      string `json:"version"`
	Platform     string `json:"platform"`
	NodeCount    int    `json:"node_count"`
	PodCount     int    `json:"pod_count"`
	NamespaceCount int  `json:"namespace_count"`
}

// ComponentStatus represents the status of a cluster component
type ComponentStatus struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Message string `json:"message"`
}