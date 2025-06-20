package domain

import (
	"context"
)

// ProviderFactory defines the interface for creating function providers
type ProviderFactory interface {
	// CreateProvider creates a provider instance based on configuration
	CreateProvider(ctx context.Context, config ProviderConfig) (Provider, error)
	
	// GetSupportedProviders returns a list of supported provider types
	GetSupportedProviders() []ProviderType
}