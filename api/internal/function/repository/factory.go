package repository

import (
	"context"
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/hexabase/hexabase-ai/api/internal/function/domain"
	"github.com/hexabase/hexabase-ai/api/internal/function/repository/fission"
	"github.com/hexabase/hexabase-ai/api/internal/function/repository/knative"
	"github.com/hexabase/hexabase-ai/api/internal/function/repository/mock"
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
func (f *ProviderFactory) CreateProvider(ctx context.Context, providerConfig domain.ProviderConfig) (domain.Provider, error) {
	switch providerConfig.Type {
	case domain.ProviderTypeKnative:
		// Extract namespace from config or use default
		namespace := "default"
		if ns, ok := providerConfig.Config["namespace"].(string); ok {
			namespace = ns
		}
		return knative.NewProvider(f.kubeClient, f.dynamicClient, namespace), nil
		
	case domain.ProviderTypeFission:
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
		
	case domain.ProviderTypeMock:
		// Mock provider for testing
		return mock.NewFunctionProvider(), nil
		
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerConfig.Type)
	}
}

// GetAvailableProviders returns the list of available provider types
func (f *ProviderFactory) GetAvailableProviders() []domain.ProviderType {
	return []domain.ProviderType{
		domain.ProviderTypeKnative,
		domain.ProviderTypeFission,
		domain.ProviderTypeMock, // For testing
	}
}

// GetSupportedProviders returns the list of supported provider types
func (f *ProviderFactory) GetSupportedProviders() []domain.ProviderType {
	return f.GetAvailableProviders()
}

// ValidateProviderConfig validates the configuration for a specific provider type
func (f *ProviderFactory) ValidateProviderConfig(providerType domain.ProviderType, config map[string]interface{}) error {
	switch providerType {
	case domain.ProviderTypeKnative:
		// Knative doesn't require additional config beyond k8s client
		return nil
		
	case domain.ProviderTypeFission:
		// Fission requires endpoint configuration
		if _, ok := config["endpoint"]; !ok {
			return fmt.Errorf("fission provider requires 'endpoint' in config")
		}
		return nil
		
	case domain.ProviderTypeMock:
		// Mock provider doesn't require any config
		return nil
		
	default:
		return fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// GetProviderCapabilities returns the capabilities for a specific provider type
// without instantiating the provider
func (f *ProviderFactory) GetProviderCapabilities(providerType domain.ProviderType) (*domain.Capabilities, error) {
	switch providerType {
	case domain.ProviderTypeKnative:
		return &domain.Capabilities{
			Name:        "knative",
			Version:     "1.0.0",
			Description: "Knative-based serverless platform",
			SupportsVersioning: true,
			SupportedRuntimes: []domain.Runtime{
				domain.RuntimeGo,
				domain.RuntimePython,
				domain.RuntimeNode,
				domain.RuntimeJava,
				domain.RuntimeDotNet,
				domain.RuntimePHP,
				domain.RuntimeRuby,
			},
			SupportedTriggerTypes: []domain.TriggerType{
				domain.TriggerHTTP,
				domain.TriggerEvent, // Via Knative Eventing
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
		
	case domain.ProviderTypeFission:
		return &domain.Capabilities{
			Name:        "fission",
			Version:     "1.0.0", 
			Description: "Fission lightweight serverless platform",
			SupportsVersioning: true,
			SupportedRuntimes: []domain.Runtime{
				domain.RuntimeGo,
				domain.RuntimePython,
				domain.RuntimeNode,
				domain.RuntimeJava,
				domain.RuntimeDotNet,
				domain.RuntimePHP,
				domain.RuntimeRuby,
			},
			SupportedTriggerTypes: []domain.TriggerType{
				domain.TriggerHTTP,
				domain.TriggerSchedule, // Fission has time triggers
				domain.TriggerEvent,    // Via message queue triggers
				domain.TriggerMessageQueue,
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
		
	case domain.ProviderTypeMock:
		mockProvider := mock.NewFunctionProvider()
		return mockProvider.GetCapabilities(), nil
		
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}