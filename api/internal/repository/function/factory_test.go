package function

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
	
	"github.com/hexabase/hexabase-ai/api/internal/domain/function"
)

func TestProviderFactory_CreateProvider(t *testing.T) {
	ctx := context.Background()
	kubeClient := kubefake.NewSimpleClientset()
	// For now, we'll pass nil since we're not actually using the dynamic client in tests
	var dynamicClient *dynamicfake.FakeDynamicClient
	
	factory := NewProviderFactory(kubeClient, dynamicClient)
	
	t.Run("CreateMockProvider", func(t *testing.T) {
		provider, err := factory.CreateProvider(ctx, function.ProviderTypeMock, nil)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})
	
	t.Run("CreateKnativeProvider", func(t *testing.T) {
		// Currently not implemented
		_, err := factory.CreateProvider(ctx, function.ProviderTypeKnative, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})
	
	t.Run("CreateFissionProvider", func(t *testing.T) {
		// Currently not implemented
		_, err := factory.CreateProvider(ctx, function.ProviderTypeFission, map[string]interface{}{
			"endpoint": "http://fission.example.com",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})
	
	t.Run("CreateUnknownProvider", func(t *testing.T) {
		_, err := factory.CreateProvider(ctx, function.ProviderType("unknown"), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported provider type")
	})
}

func TestProviderFactory_GetAvailableProviders(t *testing.T) {
	factory := NewProviderFactory(nil, nil)
	
	providers := factory.GetAvailableProviders()
	assert.Contains(t, providers, function.ProviderTypeKnative)
	assert.Contains(t, providers, function.ProviderTypeFission)
	assert.Contains(t, providers, function.ProviderTypeMock)
}

func TestProviderFactory_ValidateProviderConfig(t *testing.T) {
	factory := NewProviderFactory(nil, nil)
	
	tests := []struct {
		name         string
		providerType function.ProviderType
		config       map[string]interface{}
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "Knative - Valid",
			providerType: function.ProviderTypeKnative,
			config:       nil,
			expectError:  false,
		},
		{
			name:         "Fission - Valid",
			providerType: function.ProviderTypeFission,
			config:       map[string]interface{}{"endpoint": "http://fission.example.com"},
			expectError:  false,
		},
		{
			name:         "Fission - Missing Endpoint",
			providerType: function.ProviderTypeFission,
			config:       map[string]interface{}{},
			expectError:  true,
			errorMsg:     "endpoint",
		},
		{
			name:         "Mock - Valid",
			providerType: function.ProviderTypeMock,
			config:       nil,
			expectError:  false,
		},
		{
			name:         "Unknown Provider",
			providerType: function.ProviderType("unknown"),
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
		caps, err := factory.GetProviderCapabilities(function.ProviderTypeKnative)
		require.NoError(t, err)
		assert.True(t, caps.SupportsVersioning)
		// Check that Knative doesn't support schedule triggers
		assert.False(t, caps.HasTriggerType(function.TriggerSchedule))
		// Check that Knative supports HTTP triggers
		assert.True(t, caps.HasTriggerType(function.TriggerHTTP))
	})
	
	t.Run("Fission Capabilities", func(t *testing.T) {
		caps, err := factory.GetProviderCapabilities(function.ProviderTypeFission)
		require.NoError(t, err)
		assert.True(t, caps.SupportsVersioning)
		// Check that Fission supports schedule triggers
		assert.True(t, caps.HasTriggerType(function.TriggerSchedule))
		// Check that Fission supports HTTP triggers
		assert.True(t, caps.HasTriggerType(function.TriggerHTTP))
		// Check that Python runtime is supported
		assert.True(t, caps.HasRuntime(function.RuntimePython))
	})
	
	t.Run("Mock Capabilities", func(t *testing.T) {
		caps, err := factory.GetProviderCapabilities(function.ProviderTypeMock)
		require.NoError(t, err)
		assert.True(t, caps.SupportsVersioning)
		// Check that Mock supports schedule triggers
		assert.True(t, caps.HasTriggerType(function.TriggerSchedule))
		// Check that Mock supports HTTP triggers
		assert.True(t, caps.HasTriggerType(function.TriggerHTTP))
	})
	
	t.Run("Unknown Provider", func(t *testing.T) {
		_, err := factory.GetProviderCapabilities(function.ProviderType("unknown"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown provider type")
	})
}