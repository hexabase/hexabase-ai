package function

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/function"
)

// ConfigRepository implements configuration management for function providers
type ConfigRepository struct {
	db *sql.DB
}

// NewConfigRepository creates a new configuration repository
func NewConfigRepository(db *sql.DB) *ConfigRepository {
	return &ConfigRepository{
		db: db,
	}
}

// GetWorkspaceProviderConfig retrieves the provider configuration for a workspace
func (r *ConfigRepository) GetWorkspaceProviderConfig(ctx context.Context, workspaceID string) (*function.ProviderConfig, error) {
	query := `
		SELECT provider_type, config, created_at, updated_at
		FROM workspace_provider_configs
		WHERE workspace_id = $1
	`

	var providerType string
	var configJSON []byte
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, workspaceID).Scan(
		&providerType, &configJSON, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No config found, use default
		}
		return nil, fmt.Errorf("failed to get workspace provider config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &function.ProviderConfig{
		Type:   function.ProviderType(providerType),
		Config: config,
	}, nil
}

// UpdateWorkspaceProviderConfig updates or inserts the provider configuration for a workspace
func (r *ConfigRepository) UpdateWorkspaceProviderConfig(ctx context.Context, workspaceID string, config *function.ProviderConfig) error {
	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		INSERT INTO workspace_provider_configs (workspace_id, provider_type, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (workspace_id) DO UPDATE
		SET provider_type = EXCLUDED.provider_type,
		    config = EXCLUDED.config,
		    updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query,
		workspaceID,
		string(config.Type),
		configJSON,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to update workspace provider config: %w", err)
	}

	return nil
}

// GetDefaultProviderConfig returns the default provider configuration
func (r *ConfigRepository) GetDefaultProviderConfig() *function.ProviderConfig {
	return &function.ProviderConfig{
		Type: function.ProviderTypeFission,
		Config: map[string]interface{}{
			"endpoint":  "http://controller.fission.svc.cluster.local",
			"namespace": "fission-function",
		},
	}
}

// ValidateProviderConfig validates a provider configuration
func (r *ConfigRepository) ValidateProviderConfig(providerType function.ProviderType, config map[string]interface{}) error {
	switch providerType {
	case function.ProviderTypeFission:
		// Fission requires endpoint
		if _, ok := config["endpoint"].(string); !ok {
			return fmt.Errorf("fission provider requires 'endpoint' in config")
		}
		return nil

	case function.ProviderTypeKnative:
		// Knative doesn't require additional config beyond k8s client
		return nil

	case function.ProviderTypeMock:
		// Mock provider doesn't require any config
		return nil

	default:
		return fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// GetProviderFeatureFlags returns feature flags for a specific provider
func (r *ConfigRepository) GetProviderFeatureFlags(ctx context.Context, workspaceID string) (map[string]bool, error) {
	// Get provider config first
	config, err := r.GetWorkspaceProviderConfig(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	if config == nil {
		config = r.GetDefaultProviderConfig()
	}

	// Return feature flags based on provider type
	switch config.Type {
	case function.ProviderTypeFission:
		return map[string]bool{
			"warm_pool":       true,
			"fast_cold_start": true,
			"logs_supported":  true,
			"async_supported": true,
			"custom_images":   true,
		}, nil

	case function.ProviderTypeKnative:
		return map[string]bool{
			"warm_pool":       false,
			"fast_cold_start": false,
			"logs_supported":  true,
			"async_supported": false, // Not directly supported
			"custom_images":   true,
		}, nil

	default:
		return map[string]bool{}, nil
	}
}