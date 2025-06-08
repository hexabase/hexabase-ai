package observability

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"gopkg.in/yaml.v2"
)

// ObservabilityManager manages the hybrid observability stack
type ObservabilityManager struct {
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
}

// WorkspacePlan represents the plan type for a workspace
type WorkspacePlan struct {
	Type        string // "shared" or "dedicated"
	WorkspaceID string
	OrgID       string
	Features    []string
}

// ObservabilityConfig holds configuration for observability setup
type ObservabilityConfig struct {
	SharedPrometheusNamespace string
	ClickHouseConfig          ClickHouseConfig
	GrafanaConfig             GrafanaConfig
	OIDCConfig                OIDCConfig
}

// ClickHouseConfig holds ClickHouse configuration
type ClickHouseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// GrafanaConfig holds Grafana configuration
type GrafanaConfig struct {
	Domain        string
	AdminPassword string
}

// OIDCConfig holds OIDC configuration for SSO
type OIDCConfig struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	APIURL       string
}

// TemplateData holds data for template substitution
type TemplateData struct {
	WorkspaceID         string
	OrgID               string
	Domain              string
	OIDCClientID        string
	OIDCClientSecret    string
	OIDCAuthURL         string
	OIDCTokenURL        string
	OIDCAPIURL          string
	GrafanaAdminPassword string
}

// NewObservabilityManager creates a new observability manager
func NewObservabilityManager(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) *ObservabilityManager {
	return &ObservabilityManager{
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
	}
}

// SetupWorkspaceObservability sets up observability stack based on workspace plan
func (om *ObservabilityManager) SetupWorkspaceObservability(ctx context.Context, plan *WorkspacePlan, config *ObservabilityConfig) error {
	switch plan.Type {
	case "shared":
		return om.setupSharedPlanObservability(ctx, plan, config)
	case "dedicated":
		return om.setupDedicatedPlanObservability(ctx, plan, config)
	default:
		return fmt.Errorf("unknown plan type: %s", plan.Type)
	}
}

// setupSharedPlanObservability sets up prometheus-agent for shared plans
func (om *ObservabilityManager) setupSharedPlanObservability(ctx context.Context, plan *WorkspacePlan, config *ObservabilityConfig) error {
	namespace := fmt.Sprintf("vcluster-%s", plan.WorkspaceID)

	// Check if namespace exists
	_, err := om.kubeClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return fmt.Errorf("vCluster namespace %s not found", namespace)
	}

	// Load prometheus-agent template
	templateData := &TemplateData{
		WorkspaceID: plan.WorkspaceID,
		OrgID:       plan.OrgID,
	}

	// Apply prometheus-agent configuration
	agentYAML, err := om.renderTemplate(prometheusAgentTemplate, templateData)
	if err != nil {
		return fmt.Errorf("failed to render prometheus-agent template: %w", err)
	}

	if err := om.applyYAMLManifest(ctx, agentYAML, namespace); err != nil {
		return fmt.Errorf("failed to apply prometheus-agent: %w", err)
	}

	// Setup promtail for log forwarding
	promtailYAML, err := om.renderTemplate(promtailTemplate, templateData)
	if err != nil {
		return fmt.Errorf("failed to render promtail template: %w", err)
	}

	if err := om.applyYAMLManifest(ctx, promtailYAML, namespace); err != nil {
		return fmt.Errorf("failed to apply promtail: %w", err)
	}

	return nil
}

// setupDedicatedPlanObservability sets up full observability stack for dedicated plans
func (om *ObservabilityManager) setupDedicatedPlanObservability(ctx context.Context, plan *WorkspacePlan, config *ObservabilityConfig) error {
	vclusterNamespace := fmt.Sprintf("vcluster-%s", plan.WorkspaceID)

	// Check if vCluster namespace exists
	_, err := om.kubeClient.CoreV1().Namespaces().Get(ctx, vclusterNamespace, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return fmt.Errorf("vCluster namespace %s not found", vclusterNamespace)
	}

	templateData := &TemplateData{
		WorkspaceID:          plan.WorkspaceID,
		OrgID:                plan.OrgID,
		Domain:               config.GrafanaConfig.Domain,
		OIDCClientID:         config.OIDCConfig.ClientID,
		OIDCClientSecret:     config.OIDCConfig.ClientSecret,
		OIDCAuthURL:          config.OIDCConfig.AuthURL,
		OIDCTokenURL:         config.OIDCConfig.TokenURL,
		OIDCAPIURL:           config.OIDCConfig.APIURL,
		GrafanaAdminPassword: config.GrafanaConfig.AdminPassword,
	}

	// Apply dedicated observability stack
	dedicatedYAML, err := om.renderTemplate(dedicatedStackTemplate, templateData)
	if err != nil {
		return fmt.Errorf("failed to render dedicated stack template: %w", err)
	}

	// Apply to the vCluster (using vCluster kubeconfig)
	if err := om.applyToVCluster(ctx, plan.WorkspaceID, dedicatedYAML); err != nil {
		return fmt.Errorf("failed to apply dedicated stack to vCluster: %w", err)
	}

	return nil
}

// RemoveWorkspaceObservability removes observability stack for a workspace
func (om *ObservabilityManager) RemoveWorkspaceObservability(ctx context.Context, plan *WorkspacePlan) error {
	switch plan.Type {
	case "shared":
		return om.removeSharedPlanObservability(ctx, plan)
	case "dedicated":
		return om.removeDedicatedPlanObservability(ctx, plan)
	default:
		return fmt.Errorf("unknown plan type: %s", plan.Type)
	}
}

// removeSharedPlanObservability removes prometheus-agent and promtail
func (om *ObservabilityManager) removeSharedPlanObservability(ctx context.Context, plan *WorkspacePlan) error {
	namespace := fmt.Sprintf("vcluster-%s", plan.WorkspaceID)

	// Remove prometheus-agent deployment
	err := om.kubeClient.AppsV1().Deployments(namespace).Delete(ctx, "prometheus-agent", metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete prometheus-agent: %w", err)
	}

	// Remove promtail deployment
	err = om.kubeClient.AppsV1().Deployments(namespace).Delete(ctx, "promtail", metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete promtail: %w", err)
	}

	// Remove related services and configmaps
	_ = om.kubeClient.CoreV1().Services(namespace).Delete(ctx, "prometheus-agent", metav1.DeleteOptions{})
	_ = om.kubeClient.CoreV1().ConfigMaps(namespace).Delete(ctx, "prometheus-agent-config", metav1.DeleteOptions{})
	_ = om.kubeClient.CoreV1().ConfigMaps(namespace).Delete(ctx, "promtail-config", metav1.DeleteOptions{})

	return nil
}

// removeDedicatedPlanObservability removes dedicated observability stack
func (om *ObservabilityManager) removeDedicatedPlanObservability(ctx context.Context, plan *WorkspacePlan) error {
	// For dedicated plans, the observability stack is inside the vCluster
	// When vCluster is deleted, the observability stack is automatically removed
	// This method is for cases where we need to remove just the observability components
	
	// Would need vCluster client access to remove individual components
	// For now, this is handled by vCluster deletion
	return nil
}

// renderTemplate renders a YAML template with the provided data
func (om *ObservabilityManager) renderTemplate(templateStr string, data *TemplateData) (string, error) {
	tmpl, err := template.New("observability").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// applyYAMLManifest applies a YAML manifest to the cluster
func (om *ObservabilityManager) applyYAMLManifest(ctx context.Context, yamlContent string, namespace string) error {
	// Split YAML into individual documents
	documents := strings.Split(yamlContent, "---")

	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		// Parse YAML to unstructured object
		var obj unstructured.Unstructured
		if err := yaml.Unmarshal([]byte(doc), &obj.Object); err != nil {
			return fmt.Errorf("failed to unmarshal YAML: %w", err)
		}

		// Set namespace if not set
		if obj.GetNamespace() == "" && namespace != "" {
			obj.SetNamespace(namespace)
		}

		// Get GVR for the object
		gvr, err := om.getGVRForObject(&obj)
		if err != nil {
			return fmt.Errorf("failed to get GVR for object: %w", err)
		}

		// Apply the object
		if err := om.applyUnstructuredObject(ctx, &obj, gvr); err != nil {
			return fmt.Errorf("failed to apply object %s/%s: %w", obj.GetKind(), obj.GetName(), err)
		}
	}

	return nil
}

// applyToVCluster applies YAML to a vCluster (placeholder - would need vCluster client)
func (om *ObservabilityManager) applyToVCluster(ctx context.Context, workspaceID, yamlContent string) error {
	// This would require vCluster client setup
	// For now, this is a placeholder that would need to:
	// 1. Get vCluster kubeconfig
	// 2. Create client for vCluster
	// 3. Apply YAML to vCluster
	
	// Placeholder implementation
	return fmt.Errorf("vCluster integration not implemented yet")
}

// getGVRForObject gets the GroupVersionResource for an unstructured object
func (om *ObservabilityManager) getGVRForObject(obj *unstructured.Unstructured) (schema.GroupVersionResource, error) {
	gvk := obj.GroupVersionKind()
	
	// Simple mapping for common resources
	mapping := map[string]schema.GroupVersionResource{
		"ConfigMap":              {Group: "", Version: "v1", Resource: "configmaps"},
		"Service":                {Group: "", Version: "v1", Resource: "services"},
		"ServiceAccount":         {Group: "", Version: "v1", Resource: "serviceaccounts"},
		"Deployment":             {Group: "apps", Version: "v1", Resource: "deployments"},
		"PersistentVolumeClaim":  {Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		"ClusterRole":            {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
		"ClusterRoleBinding":     {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},
	}

	if gvr, exists := mapping[gvk.Kind]; exists {
		return gvr, nil
	}

	return schema.GroupVersionResource{}, fmt.Errorf("unknown kind: %s", gvk.Kind)
}

// applyUnstructuredObject applies an unstructured object using server-side apply
func (om *ObservabilityManager) applyUnstructuredObject(ctx context.Context, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) error {
	var resourceClient dynamic.ResourceInterface

	if obj.GetNamespace() != "" {
		resourceClient = om.dynamicClient.Resource(gvr).Namespace(obj.GetNamespace())
	} else {
		resourceClient = om.dynamicClient.Resource(gvr)
	}

	// Try to get existing object
	existing, err := resourceClient.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		// Create new object
		_, err = resourceClient.Create(ctx, obj, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	// Update existing object
	obj.SetResourceVersion(existing.GetResourceVersion())
	_, err = resourceClient.Update(ctx, obj, metav1.UpdateOptions{})
	return err
}

// CheckObservabilityHealth checks the health of observability components
func (om *ObservabilityManager) CheckObservabilityHealth(ctx context.Context, workspaceID string, planType string) (*HealthStatus, error) {
	status := &HealthStatus{
		WorkspaceID: workspaceID,
		PlanType:    planType,
		Components:  make(map[string]ComponentHealth),
	}

	namespace := fmt.Sprintf("vcluster-%s", workspaceID)

	if planType == "shared" {
		// Check prometheus-agent health
		deployment, err := om.kubeClient.AppsV1().Deployments(namespace).Get(ctx, "prometheus-agent", metav1.GetOptions{})
		if err != nil {
			status.Components["prometheus-agent"] = ComponentHealth{
				Status: "unhealthy",
				Error:  err.Error(),
			}
		} else {
			status.Components["prometheus-agent"] = ComponentHealth{
				Status:           "healthy",
				ReadyReplicas:    deployment.Status.ReadyReplicas,
				AvailableReplicas: deployment.Status.AvailableReplicas,
			}
		}

		// Check promtail health
		deployment, err = om.kubeClient.AppsV1().Deployments(namespace).Get(ctx, "promtail", metav1.GetOptions{})
		if err != nil {
			status.Components["promtail"] = ComponentHealth{
				Status: "unhealthy",
				Error:  err.Error(),
			}
		} else {
			status.Components["promtail"] = ComponentHealth{
				Status:           "healthy",
				ReadyReplicas:    deployment.Status.ReadyReplicas,
				AvailableReplicas: deployment.Status.AvailableReplicas,
			}
		}
	} else {
		// For dedicated plans, would need to check components inside vCluster
		status.Components["dedicated-stack"] = ComponentHealth{
			Status: "unknown",
			Error:  "vCluster health checking not implemented",
		}
	}

	return status, nil
}

// HealthStatus represents the health status of observability components
type HealthStatus struct {
	WorkspaceID string                      `json:"workspace_id"`
	PlanType    string                      `json:"plan_type"`
	Components  map[string]ComponentHealth  `json:"components"`
}

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Status            string `json:"status"`
	Error             string `json:"error,omitempty"`
	ReadyReplicas     int32  `json:"ready_replicas,omitempty"`
	AvailableReplicas int32  `json:"available_replicas,omitempty"`
}

// Template constants (loaded from files in production)
const prometheusAgentTemplate = `# Prometheus Agent Template for vCluster (Shared Plans)
# This template is deployed in each vCluster namespace for shared plan workspaces

apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-agent-config
  namespace: vcluster-{{.WorkspaceID}}
data:
  prometheus.yml: |
    global:
      scrape_interval: 30s
      evaluation_interval: 30s
      external_labels:
        workspace_id: '{{.WorkspaceID}}'
        tenant: 'shared'
        vcluster: '{{.WorkspaceID}}'

    scrape_configs:
      # Scrape vCluster API server
      - job_name: 'vcluster-api'
        kubernetes_sd_configs:
          - role: endpoints
            namespaces:
              names: ['kube-system']
        relabel_configs:
          - source_labels: [__meta_kubernetes_service_name]
            action: keep
            regex: kubernetes
          - target_label: workspace_id
            replacement: '{{.WorkspaceID}}'

      # Scrape pods in vCluster
      - job_name: 'vcluster-pods'
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          # Add workspace_id to all metrics
          - target_label: workspace_id
            replacement: '{{.WorkspaceID}}'
          
          # Only scrape pods with prometheus annotations
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
            action: keep
            regex: true
          
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
            action: replace
            target_label: __metrics_path__
            regex: (.+)
          
          - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
            action: replace
            regex: ([^:]+)(?::\d+)?;(\d+)
            replacement: $1:$2
            target_label: __address__

      # Scrape services in vCluster
      - job_name: 'vcluster-services'
        kubernetes_sd_configs:
          - role: service
        relabel_configs:
          - target_label: workspace_id
            replacement: '{{.WorkspaceID}}'
          
          - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
            action: keep
            regex: true

      # Scrape kube-state-metrics if deployed
      - job_name: 'kube-state-metrics'
        kubernetes_sd_configs:
          - role: service
        relabel_configs:
          - source_labels: [__meta_kubernetes_service_name]
            action: keep
            regex: kube-state-metrics
          - target_label: workspace_id
            replacement: '{{.WorkspaceID}}'

    remote_write:
      # Send metrics to shared Prometheus
      - url: "http://prometheus-shared.hexabase-monitoring.svc.cluster.local:9090/api/v1/write"
        queue_config:
          max_samples_per_send: 1000
          max_shards: 10
          batch_send_deadline: 5s
        write_relabel_configs:
          # Ensure workspace_id is always present
          - target_label: workspace_id
            replacement: '{{.WorkspaceID}}'
          
          # Add plan type label
          - target_label: plan_type
            replacement: 'shared'

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus-agent
  namespace: vcluster-{{.WorkspaceID}}
  labels:
    app: prometheus-agent
    workspace_id: {{.WorkspaceID}}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus-agent
  template:
    metadata:
      labels:
        app: prometheus-agent
        workspace_id: {{.WorkspaceID}}
    spec:
      serviceAccountName: prometheus-agent
      containers:
      - name: prometheus-agent
        image: prom/prometheus:v2.45.0
        args:
          - '--config.file=/etc/prometheus/prometheus.yml'
          - '--web.listen-address=0.0.0.0:9090'
          - '--storage.agent.path=/prometheus'
          - '--enable-feature=agent'
          - '--web.enable-lifecycle'
        ports:
        - name: web
          containerPort: 9090
        volumeMounts:
        - name: config
          mountPath: /etc/prometheus
        - name: storage
          mountPath: /prometheus
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /-/healthy
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /-/ready
            port: 9090
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: prometheus-agent-config
      - name: storage
        emptyDir:
          sizeLimit: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: prometheus-agent
  namespace: vcluster-{{.WorkspaceID}}
  labels:
    app: prometheus-agent
    workspace_id: {{.WorkspaceID}}
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9090"
    prometheus.io/path: "/metrics"
spec:
  ports:
  - port: 9090
    targetPort: 9090
    name: web
  selector:
    app: prometheus-agent
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus-agent
  namespace: vcluster-{{.WorkspaceID}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus-agent-{{.WorkspaceID}}
rules:
- apiGroups: [""]
  resources:
  - nodes
  - nodes/proxy
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs: ["get", "list", "watch"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: prometheus-agent-{{.WorkspaceID}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus-agent-{{.WorkspaceID}}
subjects:
- kind: ServiceAccount
  name: prometheus-agent
  namespace: vcluster-{{.WorkspaceID}}`

const promtailTemplate = `# Promtail configuration for shared plans
apiVersion: v1
kind: ConfigMap
metadata:
  name: promtail-config
  namespace: vcluster-{{.WorkspaceID}}
data:
  promtail.yml: |
    server:
      http_listen_port: 3101
      grpc_listen_port: 9096

    positions:
      filename: /tmp/positions.yaml

    clients:
      - url: http://loki.hexabase-monitoring.svc.cluster.local:3100/loki/api/v1/push
        tenant_id: {{.WorkspaceID}}

    scrape_configs:
      - job_name: kubernetes-pods
        kubernetes_sd_configs:
          - role: pod
        pipeline_stages:
          - cri: {}
        relabel_configs:
          - source_labels:
              - __meta_kubernetes_pod_controller_name
            regex: ([0-9a-z-.]+?)(-[0-9a-f]{8,10})?
            target_label: __tmp_controller_name

          - source_labels:
              - __meta_kubernetes_pod_label_app_kubernetes_io_name
              - __meta_kubernetes_pod_label_app
              - __tmp_controller_name
              - __meta_kubernetes_pod_name
            regex: ^;*([^;]+)(;.*)?$
            target_label: app
            replacement: $1

          - source_labels:
              - __meta_kubernetes_pod_label_app_kubernetes_io_instance
              - __meta_kubernetes_pod_label_instance
            regex: ^;*([^;]+)(;.*)?$
            target_label: instance
            replacement: $1

          - source_labels:
              - __meta_kubernetes_pod_label_app_kubernetes_io_component
              - __meta_kubernetes_pod_label_component
            regex: ^;*([^;]+)(;.*)?$
            target_label: component
            replacement: $1

          - replacement: {{.WorkspaceID}}
            target_label: workspace_id
          - replacement: {{.OrgID}}
            target_label: org_id
          - replacement: shared
            target_label: plan_type

---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: promtail
  namespace: vcluster-{{.WorkspaceID}}
  labels:
    app: promtail
    workspace_id: {{.WorkspaceID}}
spec:
  selector:
    matchLabels:
      app: promtail
  template:
    metadata:
      labels:
        app: promtail
        workspace_id: {{.WorkspaceID}}
    spec:
      serviceAccountName: promtail
      containers:
      - name: promtail
        image: grafana/promtail:2.9.0
        args:
          - -config.file=/etc/promtail/promtail.yml
        volumeMounts:
        - name: config
          mountPath: /etc/promtail
        - name: varlog
          mountPath: /var/log
          readOnly: true
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
        env:
        - name: HOSTNAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "256Mi"
            cpu: "200m"
      volumes:
      - name: config
        configMap:
          name: promtail-config
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
      tolerations:
      - key: node-role.kubernetes.io/master
        operator: Exists
        effect: NoSchedule
      - key: node-role.kubernetes.io/control-plane
        operator: Exists
        effect: NoSchedule
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: promtail
  namespace: vcluster-{{.WorkspaceID}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: promtail-{{.WorkspaceID}}
rules:
- apiGroups: [""]
  resources:
  - nodes
  - nodes/proxy
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: promtail-{{.WorkspaceID}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: promtail-{{.WorkspaceID}}
subjects:
- kind: ServiceAccount
  name: promtail
  namespace: vcluster-{{.WorkspaceID}}`

const dedicatedStackTemplate = `# Dedicated Plan Observability Stack Template
# Full Prometheus + Grafana + Loki stack deployed within vCluster

apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-dedicated-config
  namespace: default  # Inside vCluster
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
      evaluation_interval: 15s
      external_labels:
        workspace_id: '{{.WorkspaceID}}'
        tenant: 'dedicated'
        cluster: 'vcluster-{{.WorkspaceID}}'

    rule_files:
      - "alert_rules.yml"
      - "recording_rules.yml"

    alerting:
      alertmanagers:
        - static_configs:
            - targets:
              - alertmanager:9093

    scrape_configs:
      # Scrape Prometheus itself
      - job_name: 'prometheus'
        static_configs:
          - targets: ['localhost:9090']

      # Scrape all pods with prometheus annotations
      - job_name: 'kubernetes-pods'
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
            action: keep
            regex: true
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
            action: replace
            target_label: __metrics_path__
            regex: (.+)
          - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
            action: replace
            regex: ([^:]+)(?::\d+)?;(\d+)
            replacement: $1:$2
            target_label: __address__
          - target_label: workspace_id
            replacement: '{{.WorkspaceID}}'

      # Scrape services
      - job_name: 'kubernetes-services'
        kubernetes_sd_configs:
          - role: service
        relabel_configs:
          - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
            action: keep
            regex: true
          - target_label: workspace_id
            replacement: '{{.WorkspaceID}}'

      # Scrape API server
      - job_name: 'kubernetes-api'
        kubernetes_sd_configs:
          - role: endpoints
            namespaces:
              names: ['kube-system']
        relabel_configs:
          - source_labels: [__meta_kubernetes_service_name]
            action: keep
            regex: kubernetes
          - target_label: workspace_id
            replacement: '{{.WorkspaceID}}'

      # Scrape nodes (if available in vCluster)
      - job_name: 'kubernetes-nodes'
        kubernetes_sd_configs:
          - role: node
        relabel_configs:
          - target_label: workspace_id
            replacement: '{{.WorkspaceID}}'

  alert_rules.yml: |
    groups:
      - name: workspace.rules
        rules:
          - alert: PodCrashLooping
            expr: rate(kube_pod_container_status_restarts_total[5m]) > 0
            for: 5m
            labels:
              severity: warning
              workspace_id: '{{.WorkspaceID}}'
            annotations:
              summary: "Pod is crash looping"
              description: "Pod {{ $labels.pod }} in namespace {{ $labels.namespace }} is restarting frequently"

          - alert: HighCPUUsage
            expr: sum(rate(container_cpu_usage_seconds_total[5m])) by (pod) > 0.8
            for: 10m
            labels:
              severity: warning
              workspace_id: '{{.WorkspaceID}}'
            annotations:
              summary: "High CPU usage detected"
              description: "Pod {{ $labels.pod }} is using {{ $value | humanizePercentage }} CPU"

          - alert: HighMemoryUsage
            expr: sum(container_memory_usage_bytes) by (pod) / sum(container_spec_memory_limit_bytes) by (pod) > 0.9
            for: 10m
            labels:
              severity: warning
              workspace_id: '{{.WorkspaceID}}'
            annotations:
              summary: "High memory usage detected"
              description: "Pod {{ $labels.pod }} is using {{ $value | humanizePercentage }} of memory limit"

  recording_rules.yml: |
    groups:
      - name: workspace.recording
        rules:
          - record: workspace:cpu_usage_rate
            expr: sum(rate(container_cpu_usage_seconds_total[5m])) by (pod, namespace)

          - record: workspace:memory_usage_bytes
            expr: sum(container_memory_usage_bytes) by (pod, namespace)

          - record: workspace:network_receive_bytes_rate
            expr: sum(rate(container_network_receive_bytes_total[5m])) by (pod, namespace)

          - record: workspace:network_transmit_bytes_rate
            expr: sum(rate(container_network_transmit_bytes_total[5m])) by (pod, namespace)

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
  namespace: default
  labels:
    app: prometheus
    workspace_id: {{.WorkspaceID}}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
        workspace_id: {{.WorkspaceID}}
    spec:
      serviceAccountName: prometheus
      containers:
      - name: prometheus
        image: prom/prometheus:v2.45.0
        args:
          - '--config.file=/etc/prometheus/prometheus.yml'
          - '--storage.tsdb.path=/prometheus'
          - '--web.console.libraries=/etc/prometheus/console_libraries'
          - '--web.console.templates=/etc/prometheus/consoles'
          - '--storage.tsdb.retention.time=90d'
          - '--storage.tsdb.retention.size=10GB'
          - '--web.enable-lifecycle'
          - '--web.enable-admin-api'
        ports:
        - name: web
          containerPort: 9090
        volumeMounts:
        - name: config
          mountPath: /etc/prometheus
        - name: storage
          mountPath: /prometheus
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "4Gi"
            cpu: "2"
        livenessProbe:
          httpGet:
            path: /-/healthy
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /-/ready
            port: 9090
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: prometheus-dedicated-config
      - name: storage
        persistentVolumeClaim:
          claimName: prometheus-pvc

---
apiVersion: v1
kind: Service
metadata:
  name: prometheus
  namespace: default
  labels:
    app: prometheus
    workspace_id: {{.WorkspaceID}}
spec:
  ports:
  - port: 9090
    targetPort: 9090
    name: web
  selector:
    app: prometheus

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: prometheus-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi

---
# Grafana for dedicated plans
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-config
  namespace: default
data:
  grafana.ini: |
    [server]
    root_url = https://grafana-{{.WorkspaceID}}.{{.Domain}}
    
    [auth]
    disable_login_form = false
    
    [auth.generic_oauth]
    enabled = true
    name = Hexabase SSO
    allow_sign_up = true
    client_id = {{.OIDCClientID}}
    client_secret = {{.OIDCClientSecret}}
    scopes = openid profile email groups
    auth_url = {{.OIDCAuthURL}}
    token_url = {{.OIDCTokenURL}}
    api_url = {{.OIDCAPIURL}}
    role_attribute_path = groups[?@ == 'workspace:{{.WorkspaceID}}:admin'] && 'Admin' || 'Viewer'
    
    [security]
    admin_user = admin
    admin_password = {{.GrafanaAdminPassword}}
    
    [log]
    level = info

  datasources.yml: |
    apiVersion: 1
    datasources:
      - name: Prometheus
        type: prometheus
        access: proxy
        url: http://prometheus:9090
        isDefault: true
        editable: true
      - name: Loki
        type: loki
        access: proxy
        url: http://loki:3100
        editable: true

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: default
  labels:
    app: grafana
    workspace_id: {{.WorkspaceID}}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
        workspace_id: {{.WorkspaceID}}
    spec:
      containers:
      - name: grafana
        image: grafana/grafana:10.1.0
        ports:
        - containerPort: 3000
        volumeMounts:
        - name: config
          mountPath: /etc/grafana
        - name: storage
          mountPath: /var/lib/grafana
        env:
        - name: GF_INSTALL_PLUGINS
          value: "grafana-piechart-panel,grafana-worldmap-panel"
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "1Gi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: grafana-config
      - name: storage
        persistentVolumeClaim:
          claimName: grafana-pvc

---
apiVersion: v1
kind: Service
metadata:
  name: grafana
  namespace: default
  labels:
    app: grafana
    workspace_id: {{.WorkspaceID}}
spec:
  ports:
  - port: 3000
    targetPort: 3000
    name: web
  selector:
    app: grafana

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: grafana-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi

---
# Service accounts and RBAC
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus
  namespace: default

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus-{{.WorkspaceID}}
rules:
- apiGroups: [""]
  resources:
  - nodes
  - nodes/proxy
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs: ["get", "list", "watch"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: prometheus-{{.WorkspaceID}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus-{{.WorkspaceID}}
subjects:
- kind: ServiceAccount
  name: prometheus
  namespace: default`