package proxmox

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hexabase/hexabase-kaas/api/internal/domain/node"
)

// Common errors
var (
	ErrVMNotFound          = errors.New("VM not found")
	ErrTemplateNotFound    = errors.New("template not found")
	ErrInsufficientResources = errors.New("insufficient resources")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrTimeout             = errors.New("operation timeout")
)

// HTTPClient interface for HTTP operations
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client implements the ProxmoxRepository interface
type Client struct {
	baseURL      string
	username     string
	tokenID      string
	tokenSecret  string
	httpClient   HTTPClient
	defaultNode  string
}

// ClientOption defines options for configuring the client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client HTTPClient) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		// This only works if httpClient is a *http.Client
		if httpClient, ok := c.httpClient.(*http.Client); ok {
			httpClient.Timeout = timeout
		}
	}
}

// WithDefaultNode sets the default Proxmox node
func WithDefaultNode(node string) ClientOption {
	return func(c *Client) {
		c.defaultNode = node
	}
}

// NewClient creates a new Proxmox API client
func NewClient(baseURL, username, tokenID, tokenSecret string, opts ...ClientOption) *Client {
	client := &Client{
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		username:    username,
		tokenID:     tokenID,
		tokenSecret: tokenSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Proxmox often uses self-signed certs
				},
			},
		},
		defaultNode: "pve",
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// CheckConnection verifies the connection to Proxmox
func (c *Client) CheckConnection(ctx context.Context) error {
	resp, err := c.request(ctx, "GET", "/api2/json/version", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrAuthenticationFailed
	}

	return nil
}

// CreateVM creates a new VM by cloning a template
func (c *Client) CreateVM(ctx context.Context, spec node.VMSpec) (*node.ProxmoxVMInfo, error) {
	// Find next available VMID
	vmid, err := c.getNextVMID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get next VMID: %w", err)
	}

	targetNode := spec.TargetNode
	if targetNode == "" {
		targetNode = c.defaultNode
	}

	// Clone the template
	cloneData := map[string]interface{}{
		"newid":  vmid,
		"name":   spec.Name,
		"full":   true,
		"target": targetNode,
	}

	clonePath := fmt.Sprintf("/api2/json/nodes/%s/qemu/%d/clone", targetNode, spec.TemplateID)
	resp, err := c.request(ctx, "POST", clonePath, cloneData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrTemplateNotFound
	}
	if resp.StatusCode >= 500 {
		return nil, ErrInsufficientResources
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	// Wait for clone task to complete
	var cloneResp struct {
		Data string `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cloneResp); err != nil {
		return nil, err
	}

	if err := c.waitForTask(ctx, targetNode, cloneResp.Data); err != nil {
		return nil, fmt.Errorf("clone task failed: %w", err)
	}

	// Update VM configuration
	configData := map[string]interface{}{
		"cores":  spec.CPUCores,
		"memory": spec.MemoryMB,
		"scsihw": "virtio-scsi-pci",
		"net0":   fmt.Sprintf("virtio,bridge=%s", spec.NetworkBridge),
	}

	if spec.VLAN > 0 {
		configData["net0"] = fmt.Sprintf("virtio,bridge=%s,tag=%d", spec.NetworkBridge, spec.VLAN)
	}

	configPath := fmt.Sprintf("/api2/json/nodes/%s/qemu/%d/config", targetNode, vmid)
	resp, err = c.request(ctx, "PUT", configPath, configData)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	// Set cloud-init configuration
	if err := c.SetCloudInitConfig(ctx, vmid, spec.CloudInit); err != nil {
		return nil, fmt.Errorf("failed to set cloud-init config: %w", err)
	}

	// Start the VM
	if err := c.StartVM(ctx, vmid); err != nil {
		return nil, fmt.Errorf("failed to start VM: %w", err)
	}

	// Get VM info
	return c.GetVM(ctx, vmid)
}

// GetVM retrieves VM information
func (c *Client) GetVM(ctx context.Context, vmid int) (*node.ProxmoxVMInfo, error) {
	path := fmt.Sprintf("/api2/json/nodes/%s/qemu/%d/status/current", c.defaultNode, vmid)
	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrVMNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result struct {
		Data node.ProxmoxVMInfo `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	result.Data.VMID = vmid
	result.Data.Node = c.defaultNode

	return &result.Data, nil
}

// StartVM starts a VM
func (c *Client) StartVM(ctx context.Context, vmid int) error {
	return c.vmAction(ctx, vmid, "start")
}

// StopVM stops a VM
func (c *Client) StopVM(ctx context.Context, vmid int) error {
	return c.vmAction(ctx, vmid, "stop")
}

// RebootVM reboots a VM
func (c *Client) RebootVM(ctx context.Context, vmid int) error {
	return c.vmAction(ctx, vmid, "reboot")
}

// DeleteVM deletes a VM
func (c *Client) DeleteVM(ctx context.Context, vmid int) error {
	// First stop the VM if it's running
	status, err := c.GetVMStatus(ctx, vmid)
	if err != nil && err != ErrVMNotFound {
		return err
	}

	if status == "running" {
		if err := c.StopVM(ctx, vmid); err != nil {
			return fmt.Errorf("failed to stop VM before deletion: %w", err)
		}
		// Wait for VM to stop
		time.Sleep(5 * time.Second)
	}

	path := fmt.Sprintf("/api2/json/nodes/%s/qemu/%d", c.defaultNode, vmid)
	resp, err := c.request(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

// UpdateVMConfig updates VM configuration
func (c *Client) UpdateVMConfig(ctx context.Context, vmid int, config node.VMConfig) error {
	data := make(map[string]interface{})
	if config.CPUCores > 0 {
		data["cores"] = config.CPUCores
	}
	if config.MemoryMB > 0 {
		data["memory"] = config.MemoryMB
	}
	if config.Name != "" {
		data["name"] = config.Name
	}

	path := fmt.Sprintf("/api2/json/nodes/%s/qemu/%d/config", c.defaultNode, vmid)
	resp, err := c.request(ctx, "PUT", path, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

// GetVMStatus returns the current status of a VM
func (c *Client) GetVMStatus(ctx context.Context, vmid int) (string, error) {
	vm, err := c.GetVM(ctx, vmid)
	if err != nil {
		return "", err
	}
	return vm.Status, nil
}

// SetCloudInitConfig sets cloud-init configuration for a VM
func (c *Client) SetCloudInitConfig(ctx context.Context, vmid int, config node.CloudInitConfig) error {
	data := make(map[string]interface{})

	if config.UserData != "" {
		// In a real implementation, we would upload the user-data to Proxmox storage
		// For now, we'll reference it as a custom snippet
		data["cicustom"] = fmt.Sprintf("user=local:snippets/user-%d.yml", vmid)
	}

	if len(config.SSHKeys) > 0 {
		data["sshkeys"] = strings.Join(config.SSHKeys, "\n")
	}

	if config.IPAddress != "" {
		ipconfig := fmt.Sprintf("ip=%s", config.IPAddress)
		if config.Gateway != "" {
			ipconfig += fmt.Sprintf(",gw=%s", config.Gateway)
		}
		data["ipconfig0"] = ipconfig
	}

	if len(config.Nameservers) > 0 {
		data["nameserver"] = strings.Join(config.Nameservers, " ")
	}

	if len(data) == 0 {
		return nil
	}

	path := fmt.Sprintf("/api2/json/nodes/%s/qemu/%d/config", c.defaultNode, vmid)
	resp, err := c.request(ctx, "PUT", path, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

// GetVMResourceUsage retrieves current resource usage for a VM
func (c *Client) GetVMResourceUsage(ctx context.Context, vmid int) (*node.VMResourceUsage, error) {
	vm, err := c.GetVM(ctx, vmid)
	if err != nil {
		return nil, err
	}

	return &node.VMResourceUsage{
		CPUUsage:    vm.CPU * 100, // Convert to percentage
		MemoryUsage: vm.Mem,       // Current memory usage in bytes
		DiskUsage:   vm.Disk,      // Current disk usage in bytes
		NetworkIn:   vm.NetIn,
		NetworkOut:  vm.NetOut,
	}, nil
}

// CloneTemplate clones a template to create a new VM
func (c *Client) CloneTemplate(ctx context.Context, templateID int, name string) (int, error) {
	spec := node.VMSpec{
		Name:       name,
		TemplateID: templateID,
		TargetNode: c.defaultNode,
	}

	vm, err := c.CreateVM(ctx, spec)
	if err != nil {
		return 0, err
	}

	return vm.VMID, nil
}

// ListTemplates lists available VM templates
func (c *Client) ListTemplates(ctx context.Context) ([]node.VMTemplate, error) {
	path := fmt.Sprintf("/api2/json/nodes/%s/qemu", c.defaultNode)
	resp, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			VMID     int    `json:"vmid"`
			Name     string `json:"name"`
			Template int    `json:"template"`
			Status   string `json:"status"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var templates []node.VMTemplate
	for _, vm := range result.Data {
		if vm.Template == 1 {
			templates = append(templates, node.VMTemplate{
				ID:   vm.VMID,
				Name: vm.Name,
			})
		}
	}

	return templates, nil
}

// Helper methods

func (c *Client) vmAction(ctx context.Context, vmid int, action string) error {
	path := fmt.Sprintf("/api2/json/nodes/%s/qemu/%d/status/%s", c.defaultNode, vmid, action)
	resp, err := c.request(ctx, "POST", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrVMNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	var result struct {
		Data string `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// Wait for task to complete
	return c.waitForTask(ctx, c.defaultNode, result.Data)
}

func (c *Client) waitForTask(ctx context.Context, node, upid string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			path := fmt.Sprintf("/api2/json/nodes/%s/tasks/%s/status", node, upid)
			resp, err := c.request(ctx, "GET", path, nil)
			if err != nil {
				return err
			}

			var result struct {
				Data struct {
					Status     string `json:"status"`
					ExitStatus string `json:"exitstatus"`
				} `json:"data"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				resp.Body.Close()
				return err
			}
			resp.Body.Close()

			if result.Data.Status == "stopped" {
				if result.Data.ExitStatus != "OK" {
					return fmt.Errorf("task failed with status: %s", result.Data.ExitStatus)
				}
				return nil
			}
		}
	}
}

func (c *Client) getNextVMID(ctx context.Context) (int, error) {
	resp, err := c.request(ctx, "GET", "/api2/json/cluster/nextid", nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Data int `json:"data,string"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// Fallback to a default if the API doesn't support nextid
		return 100, nil
	}

	return result.Data, nil
}

func (c *Client) request(ctx context.Context, method, path string, data interface{}) (*http.Response, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	// Add authentication header
	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s!%s=%s", c.username, c.tokenID, c.tokenSecret))
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ErrTimeout
		}
		return nil, err
	}

	return resp, nil
}

func (c *Client) parseError(resp *http.Response) error {
	var errResp struct {
		Errors map[string]interface{} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("HTTP %d: failed to parse error response", resp.StatusCode)
	}

	if authErr, ok := errResp.Errors["auth"]; ok {
		return fmt.Errorf("%w: %v", ErrAuthenticationFailed, authErr)
	}

	errStrs := make([]string, 0, len(errResp.Errors))
	for key, val := range errResp.Errors {
		errStrs = append(errStrs, fmt.Sprintf("%s: %v", key, val))
	}

	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.Join(errStrs, ", "))
}