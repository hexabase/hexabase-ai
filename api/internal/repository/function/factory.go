package function

import (
	"context"
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/hexabase/hexabase-ai/api/internal/domain/function"
	"github.com/hexabase/hexabase-ai/api/internal/repository/function/fission"
	"github.com/hexabase/hexabase-ai/api/internal/repository/function/knative"
	"github.com/hexabase/hexabase-ai/api/internal/repository/function/mock"
)

// ProviderFactory creates function providers based on configuration
type ProviderFactory struct {
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) *ProviderFactory {
	return &ProviderFactory{
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
	}
}

// CreateProvider creates a provider instance based on configuration
func (f *ProviderFactory) CreateProvider(ctx context.Context, providerConfig function.ProviderConfig) (function.Provider, error) {
	switch providerConfig.Type {
	case function.ProviderTypeKnative:
		// Extract namespace from config or use default
		namespace := "default"
		if ns, ok := providerConfig.Config["namespace"].(string); ok {
			namespace = ns
		}
		return knative.NewProvider(f.kubeClient, f.dynamicClient, namespace), nil
		
	case function.ProviderTypeFission:
		// Extract required config for Fission
		endpoint, ok := providerConfig.Config["endpoint"].(string)
		if !ok {
			return nil, fmt.Errorf("fission provider requires 'endpoint' in config")
		}
		namespace := "fission-function"
		if ns, ok := providerConfig.Config["namespace"].(string); ok {
			namespace = ns
		}
		return fission.NewProvider(endpoint, namespace), nil
		
	case function.ProviderTypeMock:
		// Mock provider for testing
		return mock.NewFunctionProvider(), nil
		
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerConfig.Type)
	}
}

// GetAvailableProviders returns the list of available provider types
func (f *ProviderFactory) GetAvailableProviders() []function.ProviderType {
	return []function.ProviderType{
		function.ProviderTypeKnative,
		function.ProviderTypeFission,
		function.ProviderTypeMock, // For testing
	}
}

// GetSupportedProviders returns the list of supported provider types
func (f *ProviderFactory) GetSupportedProviders() []function.ProviderType {
	return f.GetAvailableProviders()
}

// ValidateProviderConfig validates the configuration for a specific provider type
func (f *ProviderFactory) ValidateProviderConfig(providerType function.ProviderType, config map[string]interface{}) error {
	switch providerType {
	case function.ProviderTypeKnative:
		// Knative doesn't require additional config beyond k8s client
		return nil
		
	case function.ProviderTypeFission:
		// Fission requires endpoint configuration
		if _, ok := config["endpoint"]; !ok {
			return fmt.Errorf("fission provider requires 'endpoint' in config")
		}
		return nil
		
	case function.ProviderTypeMock:
		// Mock provider doesn't require any config
		return nil
		
	default:
		return fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// GetProviderCapabilities returns the capabilities for a specific provider type
// without instantiating the provider
func (f *ProviderFactory) GetProviderCapabilities(providerType function.ProviderType) (*function.Capabilities, error) {
	switch providerType {
	case function.ProviderTypeKnative:
		return &function.Capabilities{
			Name:        "knative",
			Version:     "1.0.0",
			Description: "Knative-based serverless platform",
			SupportsVersioning: true,
			SupportedRuntimes: []function.Runtime{
				function.RuntimeGo,
				function.RuntimePython,
				function.RuntimeNode,
				function.RuntimeJava,
				function.RuntimeDotNet,
				function.RuntimePHP,
				function.RuntimeRuby,
			},
			SupportedTriggerTypes: []function.TriggerType{
				function.TriggerHTTP,
				function.TriggerEvent, // Via Knative Eventing
			},
			SupportsAsync:           true,
			SupportsLogs:            false,
			SupportsMetrics:         true,
			SupportsEnvironmentVars: true,
			SupportsCustomImages:    true,
			MaxMemoryMB:            8192,
			MaxTimeoutSecs:         600,
			MaxPayloadSizeMB:       100,
			TypicalColdStartMs:     2000, // 2-5 seconds cold start
			SupportsScaleToZero:    true,
			SupportsAutoScaling:    true,
			SupportsHTTPS:          true,
		}, nil
		
	case function.ProviderTypeFission:
		return &function.Capabilities{
			Name:        "fission",
			Version:     "1.0.0", 
			Description: "Fission lightweight serverless platform",
			SupportsVersioning: true,
			SupportedRuntimes: []function.Runtime{
				function.RuntimeGo,
				function.RuntimePython,
				function.RuntimeNode,
				function.RuntimeJava,
				function.RuntimeDotNet,
				function.RuntimePHP,
				function.RuntimeRuby,
			},
			SupportedTriggerTypes: []function.TriggerType{
				function.TriggerHTTP,
				function.TriggerSchedule, // Fission has time triggers
				function.TriggerEvent,    // Via message queue triggers
				function.TriggerMessageQueue,
			},
			SupportsAsync:           true,
			SupportsLogs:            true,
			SupportsMetrics:         true,
			SupportsEnvironmentVars: true,
			SupportsCustomImages:    true,
			MaxMemoryMB:            4096,
			MaxTimeoutSecs:         300,
			MaxPayloadSizeMB:       50,
			TypicalColdStartMs:     100, // 50-200ms cold start
			SupportsScaleToZero:    true,
			SupportsAutoScaling:    true,
			SupportsHTTPS:          true,
			SupportsWarmPool:       true,
		}, nil
		
	case function.ProviderTypeMock:
		mockProvider := mock.NewFunctionProvider()
		return mockProvider.GetCapabilities(), nil
		
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}