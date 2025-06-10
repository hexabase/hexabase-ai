package function

// Capabilities represents the capabilities of a function provider
type Capabilities struct {
	// Provider information
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`

	// Supported features
	SupportedRuntimes      []Runtime     `json:"supported_runtimes"`
	SupportedTriggerTypes  []TriggerType `json:"supported_trigger_types"`
	SupportsVersioning     bool          `json:"supports_versioning"`
	SupportsAsync          bool          `json:"supports_async"`
	SupportsLogs           bool          `json:"supports_logs"`
	SupportsMetrics        bool          `json:"supports_metrics"`
	SupportsTracing        bool          `json:"supports_tracing"`
	SupportsSecrets        bool          `json:"supports_secrets"`
	SupportsEnvironmentVars bool         `json:"supports_environment_vars"`
	SupportsCustomDomains  bool          `json:"supports_custom_domains"`
	SupportsPrivateRegistry bool         `json:"supports_private_registry"`

	// Resource limits
	MaxMemoryMB      int `json:"max_memory_mb"`
	MaxTimeoutSecs   int `json:"max_timeout_secs"`
	MaxPayloadSizeMB int `json:"max_payload_size_mb"`
	MaxConcurrency   int `json:"max_concurrency"`

	// Build capabilities
	SupportsBuildFromSource bool     `json:"supports_build_from_source"`
	SupportedBuildPacks     []string `json:"supported_build_packs,omitempty"`
	SupportsCustomImages    bool     `json:"supports_custom_images"`

	// Scaling capabilities
	SupportsAutoScaling     bool `json:"supports_auto_scaling"`
	SupportsScaleToZero     bool `json:"supports_scale_to_zero"`
	MinScaleDownDelaySecs   int  `json:"min_scale_down_delay_secs,omitempty"`
	MaxScaleUpRate          int  `json:"max_scale_up_rate,omitempty"` // replicas per second

	// Network capabilities
	SupportsHTTPS           bool `json:"supports_https"`
	SupportsWebSockets      bool `json:"supports_websockets"`
	SupportsGRPC            bool `json:"supports_grpc"`
	SupportsCustomHeaders   bool `json:"supports_custom_headers"`
	
	// Cold start characteristics
	TypicalColdStartMs      int  `json:"typical_cold_start_ms"`
	SupportsWarmPool        bool `json:"supports_warm_pool"`
	WarmPoolSizePerFunction int  `json:"warm_pool_size_per_function,omitempty"`

	// Observability
	LogRetentionDays        int      `json:"log_retention_days"`
	MetricsRetentionDays    int      `json:"metrics_retention_days"`
	SupportedLogFormats     []string `json:"supported_log_formats,omitempty"`
	SupportedMetricTypes    []string `json:"supported_metric_types,omitempty"`

	// Security
	SupportsIAMIntegration  bool     `json:"supports_iam_integration"`
	SupportsNetworkPolicies bool     `json:"supports_network_policies"`
	SupportedAuthMethods    []string `json:"supported_auth_methods,omitempty"`

	// Cost model hints
	CostModel            string  `json:"cost_model,omitempty"` // e.g., "per-invocation", "per-gb-second"
	FreeTierInvocations  int64   `json:"free_tier_invocations,omitempty"`
	FreeTierGBSeconds    float64 `json:"free_tier_gb_seconds,omitempty"`
}

// HasRuntime checks if a runtime is supported
func (c *Capabilities) HasRuntime(runtime Runtime) bool {
	for _, r := range c.SupportedRuntimes {
		if r == runtime {
			return true
		}
	}
	return false
}

// HasTriggerType checks if a trigger type is supported
func (c *Capabilities) HasTriggerType(triggerType TriggerType) bool {
	for _, t := range c.SupportedTriggerTypes {
		if t == triggerType {
			return true
		}
	}
	return false
}

// IsFeatureSupported checks if a feature is supported based on feature name
func (c *Capabilities) IsFeatureSupported(feature string) bool {
	switch feature {
	case "versioning":
		return c.SupportsVersioning
	case "async":
		return c.SupportsAsync
	case "logs":
		return c.SupportsLogs
	case "metrics":
		return c.SupportsMetrics
	case "tracing":
		return c.SupportsTracing
	case "secrets":
		return c.SupportsSecrets
	case "environment_vars":
		return c.SupportsEnvironmentVars
	case "custom_domains":
		return c.SupportsCustomDomains
	case "private_registry":
		return c.SupportsPrivateRegistry
	case "build_from_source":
		return c.SupportsBuildFromSource
	case "custom_images":
		return c.SupportsCustomImages
	case "auto_scaling":
		return c.SupportsAutoScaling
	case "scale_to_zero":
		return c.SupportsScaleToZero
	case "https":
		return c.SupportsHTTPS
	case "websockets":
		return c.SupportsWebSockets
	case "grpc":
		return c.SupportsGRPC
	case "warm_pool":
		return c.SupportsWarmPool
	case "iam_integration":
		return c.SupportsIAMIntegration
	case "network_policies":
		return c.SupportsNetworkPolicies
	default:
		return false
	}
}