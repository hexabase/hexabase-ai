package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"

	"github.com/hexabase/hexabase-ai/api/internal/function/domain"
)

func TestProviderFactory_CreateProvider(t *testing.T) {
	ctx := context.Background()
	kubeClient := kubefake.NewSimpleClientset()
	dynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	
	factory := NewProviderFactory(kubeClient, dynamicClient)
	
	t.Run("CreateMockProvider", func(t *testing.T) {
		config := domain.ProviderConfig{
			Type:   domain.ProviderTypeMock,
			Config: map[string]interface{}{
				"endpoint": "http://mock",
			},
		}
		provider, err := factory.CreateProvider(ctx, config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})
	
	t.Run("CreateKnativeProvider", func(t *testing.T) {
		config := domain.ProviderConfig{
			Type:   domain.ProviderTypeKnative,
			Config: map[string]interface{}{
				"namespace": "knative-serving",
			},
		}
		provider, err := factory.CreateProvider(ctx, config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})
	
	t.Run("CreateFissionProvider", func(t *testing.T) {
		config := domain.ProviderConfig{
			Type: domain.ProviderTypeFission,
			Config: map[string]interface{}{
				"endpoint":  "http://controller.fission",
				"namespace": "fission",
			},
		}
		provider, err := factory.CreateProvider(ctx, config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})
	
	t.Run("CreateUnknownProvider", func(t *testing.T) {
		config := domain.ProviderConfig{
			Type:   "unknown",
		}
		_, err := factory.CreateProvider(ctx, config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported provider type")
	})
}

func TestProviderFactory_GetSupportedProviders(t *testing.T) {
	factory := NewProviderFactory(nil, nil)
	
	providers := factory.GetSupportedProviders()
	assert.Contains(t, providers, domain.ProviderTypeKnative)
	assert.Contains(t, providers, domain.ProviderTypeFission)
	assert.Contains(t, providers, domain.ProviderTypeMock)
}

func TestProviderFactory_ValidateProviderConfig(t *testing.T) {
	factory := NewProviderFactory(nil, nil)
	
	tests := []struct {
		name         string
		providerType domain.ProviderType
		config       map[string]interface{}
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "Knative - Valid",
			providerType: domain.ProviderTypeKnative,
			config:       nil,
			expectError:  false,
		},
		{
			name:         "Fission - Valid",
			providerType: domain.ProviderTypeFission,
			config:       map[string]interface{}{"endpoint": "http://fission.example.com"},
			expectError:  false,
		},
		{
			name:         "Fission - Missing Endpoint",
			providerType: domain.ProviderTypeFission,
			config:       map[string]interface{}{},
			expectError:  true,
			errorMsg:     "endpoint",
		},
		{
			name:         "Mock - Valid",
			providerType: domain.ProviderTypeMock,
			config:       nil,
			expectError:  false,
		},
		{
			name:         "Unknown Provider",
			providerType: domain.ProviderType("unknown"),
			config:       nil,
			expectError:  true,
			errorMsg:     "unknown provider type",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.ValidateProviderConfig(tt.providerType, tt.config)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProviderFactory_GetProviderCapabilities(t *testing.T) {
	factory := NewProviderFactory(nil, nil)
	
	t.Run("Knative Capabilities", func(t *testing.T) {
		caps, err := factory.GetProviderCapabilities(domain.ProviderTypeKnative)
		require.NoError(t, err)
		assert.True(t, caps.SupportsVersioning)
		// Check that Knative doesn't support schedule triggers
		assert.False(t, caps.HasTriggerType(domain.TriggerSchedule))
		// Check that Knative supports HTTP triggers
		assert.True(t, caps.HasTriggerType(domain.TriggerHTTP))
	})
	
	t.Run("Fission Capabilities", func(t *testing.T) {
		caps, err := factory.GetProviderCapabilities(domain.ProviderTypeFission)
		require.NoError(t, err)
		assert.True(t, caps.SupportsVersioning)
		// Check that Fission supports schedule triggers
		assert.True(t, caps.HasTriggerType(domain.TriggerSchedule))
		// Check that Fission supports HTTP triggers
		assert.True(t, caps.HasTriggerType(domain.TriggerHTTP))
		// Check that Python runtime is supported
		assert.True(t, caps.HasRuntime(domain.RuntimePython))
	})
	
	t.Run("Mock Capabilities", func(t *testing.T) {
		caps, err := factory.GetProviderCapabilities(domain.ProviderTypeMock)
		require.NoError(t, err)
		assert.True(t, caps.SupportsVersioning)
		// Check that Mock supports schedule triggers
		assert.True(t, caps.HasTriggerType(domain.TriggerSchedule))
		// Check that Mock supports HTTP triggers
		assert.True(t, caps.HasTriggerType(domain.TriggerHTTP))
	})
	
	t.Run("Unknown Provider", func(t *testing.T) {
		_, err := factory.GetProviderCapabilities(domain.ProviderType("unknown"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown provider type")
	})
}