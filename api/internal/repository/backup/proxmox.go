package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
	"github.com/hexabase/hexabase-ai/api/internal/repository/proxmox"
)

// ProxmoxRepository implements the backup ProxmoxRepository using Proxmox API
type ProxmoxRepository struct {
	client *proxmox.Client
}

// NewProxmoxRepository creates a new Proxmox backup repository
func NewProxmoxRepository(client *proxmox.Client) backup.ProxmoxRepository {
	return &ProxmoxRepository{
		client: client,
	}
}

// CreateStorage creates a new storage configuration in Proxmox
func (r *ProxmoxRepository) CreateStorage(ctx context.Context, nodeID string, config backup.ProxmoxStorageConfig) (string, error) {
	// Build request data
	data := map[string]interface{}{
		"storage": config.Name,
		"type":    config.Type,
		"content": joinContent(config.Content),
	}

	// Add type-specific parameters
	switch config.Type {
	case "dir":
		if config.Path == "" {
			return "", fmt.Errorf("path is required for directory storage")
		}
		data["path"] = config.Path
	case "nfs":
		if config.Server == "" || config.Export == "" {
			return "", fmt.Errorf("server and export are required for NFS storage")
		}
		data["server"] = config.Server
		data["export"] = config.Export
	case "cephfs":
		if config.Server == "" {
			return "", fmt.Errorf("monitor hosts are required for CephFS storage")
		}
		data["monhost"] = config.Server
		data["username"] = config.Username
		data["keyring"] = config.KeyRing
	}

	// Add additional options
	for k, v := range config.Options {
		data[k] = v
	}

	// Create storage
	resp, err := r.client.PostJSON(ctx, fmt.Sprintf("/storage"), data)
	if err != nil {
		return "", fmt.Errorf("failed to create storage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return config.Name, nil
}

// DeleteStorage deletes a storage configuration from Proxmox
func (r *ProxmoxRepository) DeleteStorage(ctx context.Context, nodeID, storageID string) error {
	resp, err := r.client.Delete(ctx, fmt.Sprintf("/storage/%s", storageID))
	if err != nil {
		return fmt.Errorf("failed to delete storage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetStorageInfo retrieves information about a storage
func (r *ProxmoxRepository) GetStorageInfo(ctx context.Context, nodeID, storageID string) (*backup.ProxmoxStorageInfo, error) {
	// Get storage configuration
	configResp, err := r.client.Get(ctx, fmt.Sprintf("/storage/%s", storageID))
	if err != nil {
		return nil, fmt.Errorf("failed to get storage config: %w", err)
	}
	defer configResp.Body.Close()

	if configResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("storage not found")
	}

	var configResult struct {
		Data struct {
			Storage string   `json:"storage"`
			Type    string   `json:"type"`
			Content string   `json:"content"`
			Shared  int      `json:"shared"`
			Disable int      `json:"disable"`
		} `json:"data"`
	}
	if err := json.NewDecoder(configResp.Body).Decode(&configResult); err != nil {
		return nil, fmt.Errorf("failed to decode storage config: %w", err)
	}

	// Get storage status
	statusResp, err := r.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/status", nodeID, storageID))
	if err != nil {
		return nil, fmt.Errorf("failed to get storage status: %w", err)
	}
	defer statusResp.Body.Close()

	var statusResult struct {
		Data struct {
			Total     int64 `json:"total"`
			Used      int64 `json:"used"`
			Available int64 `json:"avail"`
			Active    int   `json:"active"`
		} `json:"data"`
	}
	if err := json.NewDecoder(statusResp.Body).Decode(&statusResult); err != nil {
		return nil, fmt.Errorf("failed to decode storage status: %w", err)
	}

	info := &backup.ProxmoxStorageInfo{
		ID:        configResult.Data.Storage,
		Type:      configResult.Data.Type,
		Total:     statusResult.Data.Total,
		Used:      statusResult.Data.Used,
		Available: statusResult.Data.Available,
		Active:    statusResult.Data.Active == 1,
		Enabled:   configResult.Data.Disable == 0,
		Shared:    configResult.Data.Shared == 1,
		Content:   splitContent(configResult.Data.Content),
	}

	return info, nil
}

// ResizeStorage resizes a storage (if supported by the storage type)
func (r *ProxmoxRepository) ResizeStorage(ctx context.Context, nodeID, storageID string, newSizeGB int) error {
	// Note: Direct storage resize is not typically supported in Proxmox
	// This would need to be handled at the underlying storage level
	// For now, we'll return an error indicating this limitation
	return fmt.Errorf("storage resize not supported via Proxmox API")
}

// CreateBackupVolume creates a backup volume on the storage
func (r *ProxmoxRepository) CreateBackupVolume(ctx context.Context, nodeID, storageID string, volumeSize int64) (string, error) {
	// Generate volume ID
	volumeID := fmt.Sprintf("backup-%s-%d", nodeID, time.Now().Unix())

	// Create volume
	data := map[string]interface{}{
		"vmid":     "0", // 0 for unassigned volumes
		"filename": volumeID,
		"size":     fmt.Sprintf("%dG", volumeSize/(1024*1024*1024)),
		"format":   "raw",
	}

	resp, err := r.client.PostJSON(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content", nodeID, storageID), data)
	if err != nil {
		return "", fmt.Errorf("failed to create backup volume: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return volumeID, nil
}

// DeleteBackupVolume deletes a backup volume
func (r *ProxmoxRepository) DeleteBackupVolume(ctx context.Context, nodeID, storageID, volumeID string) error {
	resp, err := r.client.Delete(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content/%s:%s", nodeID, storageID, storageID, volumeID))
	if err != nil {
		return fmt.Errorf("failed to delete backup volume: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetVolumeInfo retrieves information about a volume
func (r *ProxmoxRepository) GetVolumeInfo(ctx context.Context, nodeID, storageID, volumeID string) (*backup.ProxmoxVolumeInfo, error) {
	resp, err := r.client.Get(ctx, fmt.Sprintf("/nodes/%s/storage/%s/content", nodeID, storageID))
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Volid  string `json:"volid"`
			Size   int64  `json:"size"`
			Used   int64  `json:"used"`
			Format string `json:"format"`
			Path   string `json:"path"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode volume list: %w", err)
	}

	// Find the specific volume
	for _, vol := range result.Data {
		if vol.Volid == fmt.Sprintf("%s:%s", storageID, volumeID) {
			return &backup.ProxmoxVolumeInfo{
				ID:     volumeID,
				Name:   volumeID,
				Size:   vol.Size,
				Used:   vol.Used,
				Format: vol.Format,
				Path:   vol.Path,
			}, nil
		}
	}

	return nil, fmt.Errorf("volume not found")
}

// SetStorageQuota sets a quota for the storage
func (r *ProxmoxRepository) SetStorageQuota(ctx context.Context, nodeID, storageID string, quotaGB int) error {
	// Note: Proxmox doesn't directly support per-storage quotas via API
	// This would typically be handled at the filesystem level or through
	// resource pools. For now, we'll store this as a custom property
	data := map[string]interface{}{
		"maxfiles": strconv.Itoa(quotaGB * 1000), // Use maxfiles as a proxy for quota
	}

	resp, err := r.client.PutJSON(ctx, fmt.Sprintf("/storage/%s", storageID), data)
	if err != nil {
		return fmt.Errorf("failed to set storage quota: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetStorageQuota retrieves the quota for a storage
func (r *ProxmoxRepository) GetStorageQuota(ctx context.Context, nodeID, storageID string) (*backup.ProxmoxStorageQuota, error) {
	// Get storage info first
	info, err := r.GetStorageInfo(ctx, nodeID, storageID)
	if err != nil {
		return nil, err
	}

	// Get the quota from storage config
	resp, err := r.client.Get(ctx, fmt.Sprintf("/storage/%s", storageID))
	if err != nil {
		return nil, fmt.Errorf("failed to get storage config: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			MaxFiles string `json:"maxfiles"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode storage config: %w", err)
	}

	// Convert maxfiles back to quota
	quotaGB := 0
	if result.Data.MaxFiles != "" {
		if maxFiles, err := strconv.Atoi(result.Data.MaxFiles); err == nil {
			quotaGB = maxFiles / 1000
		}
	}

	// Calculate used GB from bytes
	usedGB := int(info.Used / (1024 * 1024 * 1024))
	
	return &backup.ProxmoxStorageQuota{
		StorageID: storageID,
		QuotaGB:   quotaGB,
		UsedGB:    usedGB,
		Available: quotaGB - usedGB,
	}, nil
}

// Helper functions

func joinContent(content []string) string {
	result := ""
	for i, c := range content {
		if i > 0 {
			result += ","
		}
		result += c
	}
	return result
}

func splitContent(content string) []string {
	if content == "" {
		return []string{}
	}
	result := []string{}
	for _, c := range strings.Split(content, ",") {
		c = strings.TrimSpace(c)
		if c != "" {
			result = append(result, c)
		}
	}
	return result
}