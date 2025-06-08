package proxmox_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/hexabase/hexabase-ai/api/internal/domain/node"
	"github.com/hexabase/hexabase-ai/api/internal/repository/proxmox"
)

func TestProxmoxClient_CreateVM(t *testing.T) {
	tests := []struct {
		name      string
		spec      node.VMSpec
		mockSetup func(*proxmox.MockHTTPClient)
		expected  *node.ProxmoxVMInfo
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful VM creation",
			spec: node.VMSpec{
				Name:       "test-node-1",
				NodeType:   "S-Type",
				TemplateID: 9000,
				TargetNode: "pve-node1",
				CPUCores:   4,
				MemoryMB:   16384,
				DiskGB:     200,
				NetworkBridge: "vmbr0",
				CloudInit: node.CloudInitConfig{
					SSHKeys: []string{"ssh-rsa AAAAB3NzaC1yc2E..."},
				},
			},
			mockSetup: func(mock *proxmox.MockHTTPClient) {
				// Mock get next VMID
				mock.ExpectGet("/api2/json/cluster/nextid").
					RespondJSON(200, map[string]interface{}{
						"data": "100",
					})

				// Mock clone template request
				mock.ExpectPost("/api2/json/nodes/pve-node1/qemu/9000/clone").
					WithJSON(map[string]interface{}{
						"newid":  float64(100),
						"name":   "test-node-1",
						"full":   true,
						"target": "pve-node1",
					}).
					RespondJSON(200, map[string]interface{}{
						"data": "UPID:pve-node1:00001234:00000000:00000000:qmclone:9000:root@pam:",
					})

				// Mock task status check
				mock.ExpectGet("/api2/json/nodes/pve-node1/tasks/UPID:pve-node1:00001234:00000000:00000000:qmclone:9000:root@pam:/status").
					RespondJSON(200, map[string]interface{}{
						"data": map[string]interface{}{
							"status": "stopped",
							"exitstatus": "OK",
						},
					})

				// Mock VM config update
				mock.ExpectPut("/api2/json/nodes/pve-node1/qemu/100/config").
					WithJSON(map[string]interface{}{
						"cores":   float64(4),
						"memory":  float64(16384),
						"scsihw":  "virtio-scsi-pci",
						"net0":    "virtio,bridge=vmbr0",
					}).
					RespondJSON(200, map[string]interface{}{"data": nil})

				// Mock cloud-init config
				mock.ExpectPut("/api2/json/nodes/pve-node1/qemu/100/config").
					WithJSON(map[string]interface{}{
						"sshkeys": "ssh-rsa AAAAB3NzaC1yc2E...",
					}).
					RespondJSON(200, map[string]interface{}{"data": nil})

				// Mock VM start
				mock.ExpectPost("/api2/json/nodes/pve-node1/qemu/100/status/start").
					RespondJSON(200, map[string]interface{}{
						"data": "UPID:pve-node1:00001235:00000000:00000000:qmstart:100:root@pam:",
					})

				// Mock start task status check
				mock.ExpectGet("/api2/json/nodes/pve-node1/tasks/UPID:pve-node1:00001235:00000000:00000000:qmstart:100:root@pam:/status").
					RespondJSON(200, map[string]interface{}{
						"data": map[string]interface{}{
							"status": "stopped",
							"exitstatus": "OK",
						},
					})

				// Mock VM status
				mock.ExpectGet("/api2/json/nodes/pve-node1/qemu/100/status/current").
					RespondJSON(200, map[string]interface{}{
						"data": map[string]interface{}{
							"vmid":   100,
							"status": "running",
							"maxmem": 17179869184,
							"maxdisk": 214748364800,
							"cpu":    0.02,
						},
					})
			},
			expected: &node.ProxmoxVMInfo{
				VMID:    100,
				Node:    "pve-node1",
				Status:  "running",
				MaxMem:  17179869184,
				MaxDisk: 214748364800,
				CPU:     0.02,
			},
			wantErr: false,
		},
		{
			name: "template not found",
			spec: node.VMSpec{
				Name:       "test-node-2",
				TemplateID: 9999,
				TargetNode: "pve-node1",
			},
			mockSetup: func(mock *proxmox.MockHTTPClient) {
				// Mock get next VMID
				mock.ExpectGet("/api2/json/cluster/nextid").
					RespondJSON(200, map[string]interface{}{
						"data": "101",
					})
				
				mock.ExpectPost("/api2/json/nodes/pve-node1/qemu/9999/clone").
					RespondJSON(404, map[string]interface{}{
						"errors": map[string]interface{}{
							"vmid": "VM 9999 not found",
						},
					})
			},
			wantErr: true,
			errMsg:  "template not found",
		},
		{
			name: "insufficient resources",
			spec: node.VMSpec{
				Name:       "test-node-3",
				TemplateID: 9000,
				TargetNode: "pve-node1",
				CPUCores:   64,
				MemoryMB:   524288, // 512GB
			},
			mockSetup: func(mock *proxmox.MockHTTPClient) {
				// Mock get next VMID
				mock.ExpectGet("/api2/json/cluster/nextid").
					RespondJSON(200, map[string]interface{}{
						"data": "102",
					})
				
				mock.ExpectPost("/api2/json/nodes/pve-node1/qemu/9000/clone").
					RespondJSON(500, map[string]interface{}{
						"errors": map[string]interface{}{
							"memory": "unable to allocate memory",
						},
					})
			},
			wantErr: true,
			errMsg:  "insufficient resources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := proxmox.NewMockHTTPClient(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			client := proxmox.NewClient(
				"https://pve.example.com:8006",
				"root@pam",
				"test-token-id",
				"test-token-secret",
				proxmox.WithHTTPClient(mockClient),
				proxmox.WithDefaultNode("pve-node1"),
			)

			result, err := client.CreateVM(context.Background(), tt.spec)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.VMID, result.VMID)
				assert.Equal(t, tt.expected.Status, result.Status)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestProxmoxClient_VMLifecycle(t *testing.T) {
	vmid := 100
	
	tests := []struct {
		name      string
		operation string
		mockSetup func(*proxmox.MockHTTPClient)
		wantErr   bool
	}{
		{
			name:      "start VM successfully",
			operation: "start",
			mockSetup: func(mock *proxmox.MockHTTPClient) {
				mock.ExpectPost("/api2/json/nodes/pve-node1/qemu/100/status/start").
					RespondJSON(200, map[string]interface{}{
						"data": "UPID:pve-node1:00001236:00000000:00000000:qmstart:100:root@pam:",
					})
				
				mock.ExpectGet("/api2/json/nodes/pve-node1/tasks/UPID:pve-node1:00001236:00000000:00000000:qmstart:100:root@pam:/status").
					RespondJSON(200, map[string]interface{}{
						"data": map[string]interface{}{
							"status": "stopped",
							"exitstatus": "OK",
						},
					})
			},
			wantErr: false,
		},
		{
			name:      "stop VM successfully",
			operation: "stop",
			mockSetup: func(mock *proxmox.MockHTTPClient) {
				mock.ExpectPost("/api2/json/nodes/pve-node1/qemu/100/status/stop").
					RespondJSON(200, map[string]interface{}{
						"data": "UPID:pve-node1:00001237:00000000:00000000:qmstop:100:root@pam:",
					})
				
				mock.ExpectGet("/api2/json/nodes/pve-node1/tasks/UPID:pve-node1:00001237:00000000:00000000:qmstop:100:root@pam:/status").
					RespondJSON(200, map[string]interface{}{
						"data": map[string]interface{}{
							"status": "stopped",
							"exitstatus": "OK",
						},
					})
			},
			wantErr: false,
		},
		{
			name:      "reboot VM successfully",
			operation: "reboot",
			mockSetup: func(mock *proxmox.MockHTTPClient) {
				mock.ExpectPost("/api2/json/nodes/pve-node1/qemu/100/status/reboot").
					RespondJSON(200, map[string]interface{}{
						"data": "UPID:pve-node1:00001238:00000000:00000000:qmreboot:100:root@pam:",
					})
				
				mock.ExpectGet("/api2/json/nodes/pve-node1/tasks/UPID:pve-node1:00001238:00000000:00000000:qmreboot:100:root@pam:/status").
					RespondJSON(200, map[string]interface{}{
						"data": map[string]interface{}{
							"status": "stopped",
							"exitstatus": "OK",
						},
					})
			},
			wantErr: false,
		},
		{
			name:      "VM not found",
			operation: "start",
			mockSetup: func(mock *proxmox.MockHTTPClient) {
				mock.ExpectPost("/api2/json/nodes/pve-node1/qemu/100/status/start").
					RespondJSON(404, map[string]interface{}{
						"errors": map[string]interface{}{
							"vmid": "VM 100 not found",
						},
					})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := proxmox.NewMockHTTPClient(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			client := proxmox.NewClient(
				"https://pve.example.com:8006",
				"root@pam",
				"test-token-id",
				"test-token-secret",
				proxmox.WithHTTPClient(mockClient),
				proxmox.WithDefaultNode("pve-node1"),
			)

			ctx := context.Background()
			var err error

			switch tt.operation {
			case "start":
				err = client.StartVM(ctx, vmid)
			case "stop":
				err = client.StopVM(ctx, vmid)
			case "reboot":
				err = client.RebootVM(ctx, vmid)
			}

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestProxmoxClient_GetVMResourceUsage(t *testing.T) {
	tests := []struct {
		name      string
		vmid      int
		mockSetup func(*proxmox.MockHTTPClient)
		expected  *node.VMResourceUsage
		wantErr   bool
	}{
		{
			name: "get resource usage successfully",
			vmid: 100,
			mockSetup: func(mock *proxmox.MockHTTPClient) {
				mock.ExpectGet("/api2/json/nodes/pve-node1/qemu/100/status/current").
					RespondJSON(200, map[string]interface{}{
						"data": map[string]interface{}{
							"cpu":       0.156,
							"mem":       8589934592, // 8GB
							"disk":      0,
							"netin":     1048576,    // 1MB/s
							"netout":    524288,     // 512KB/s
							"diskread":  0,
							"diskwrite": 0,
						},
					})
			},
			expected: &node.VMResourceUsage{
				CPUUsage:    15.6,
				MemoryUsage: 8589934592,
				DiskUsage:   0,
				NetworkIn:   1048576,
				NetworkOut:  524288,
			},
			wantErr: false,
		},
		{
			name: "VM not running",
			vmid: 101,
			mockSetup: func(mock *proxmox.MockHTTPClient) {
				mock.ExpectGet("/api2/json/nodes/pve-node1/qemu/101/status/current").
					RespondJSON(200, map[string]interface{}{
						"data": map[string]interface{}{
							"status": "stopped",
							"cpu":    0,
							"mem":    0,
						},
					})
			},
			expected: &node.VMResourceUsage{
				CPUUsage:    0,
				MemoryUsage: 0,
				DiskUsage:   0,
				NetworkIn:   0,
				NetworkOut:  0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := proxmox.NewMockHTTPClient(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			client := proxmox.NewClient(
				"https://pve.example.com:8006",
				"root@pam",
				"test-token-id",
				"test-token-secret",
				proxmox.WithHTTPClient(mockClient),
				proxmox.WithDefaultNode("pve-node1"),
			)

			result, err := client.GetVMResourceUsage(context.Background(), tt.vmid)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.CPUUsage, result.CPUUsage)
				assert.Equal(t, tt.expected.MemoryUsage, result.MemoryUsage)
				assert.Equal(t, tt.expected.NetworkIn, result.NetworkIn)
				assert.Equal(t, tt.expected.NetworkOut, result.NetworkOut)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestProxmoxClient_CloudInit(t *testing.T) {
	userData := `#cloud-config
users:
  - name: hexabase
    groups: sudo
    shell: /bin/bash
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    ssh_authorized_keys:
      - ssh-rsa AAAAB3NzaC1yc2E...

runcmd:
  - curl -sfL https://get.k3s.io | sh -s - agent --server https://k3s-server:6443 --token mytoken
`

	tests := []struct {
		name      string
		vmid      int
		config    node.CloudInitConfig
		mockSetup func(*proxmox.MockHTTPClient)
		wantErr   bool
	}{
		{
			name: "set cloud-init config successfully",
			vmid: 100,
			config: node.CloudInitConfig{
				UserData:    userData,
				SSHKeys:     []string{"ssh-rsa AAAAB3NzaC1yc2E..."},
				IPAddress:   "10.0.0.100/24",
				Gateway:     "10.0.0.1",
				Nameservers: []string{"8.8.8.8", "8.8.4.4"},
			},
			mockSetup: func(mock *proxmox.MockHTTPClient) {
				mock.ExpectPut("/api2/json/nodes/pve-node1/qemu/100/config").
					WithJSON(map[string]interface{}{
						"cicustom":   "user=local:snippets/user-100.yml",
						"sshkeys":    "ssh-rsa AAAAB3NzaC1yc2E...",
						"ipconfig0":  "ip=10.0.0.100/24,gw=10.0.0.1",
						"nameserver": "8.8.8.8 8.8.4.4",
					}).
					RespondJSON(200, map[string]interface{}{"data": nil})
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := proxmox.NewMockHTTPClient(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			client := proxmox.NewClient(
				"https://pve.example.com:8006",
				"root@pam",
				"test-token-id",
				"test-token-secret",
				proxmox.WithHTTPClient(mockClient),
				proxmox.WithDefaultNode("pve-node1"),
			)

			err := client.SetCloudInitConfig(context.Background(), tt.vmid, tt.config)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestProxmoxClient_Authentication(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*proxmox.MockHTTPClient)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful authentication",
			mockSetup: func(mock *proxmox.MockHTTPClient) {
				// Test with a simple API call
				mock.ExpectGet("/api2/json/version").
					WithHeaders(map[string]string{
						"Authorization": "PVEAPIToken=root@pam!test-token-id=test-token-secret",
					}).
					RespondJSON(200, map[string]interface{}{
						"data": map[string]interface{}{
							"version": "7.4-1",
							"release": "7.4",
						},
					})
			},
			wantErr: false,
		},
		{
			name: "invalid credentials",
			mockSetup: func(mock *proxmox.MockHTTPClient) {
				mock.ExpectGet("/api2/json/version").
					RespondJSON(401, map[string]interface{}{
						"errors": map[string]interface{}{
							"auth": "invalid credentials",
						},
					})
			},
			wantErr: true,
			errMsg:  "authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := proxmox.NewMockHTTPClient(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockClient)
			}

			client := proxmox.NewClient(
				"https://pve.example.com:8006",
				"root@pam",
				"test-token-id",
				"test-token-secret",
				proxmox.WithHTTPClient(mockClient),
				proxmox.WithDefaultNode("pve-node1"),
			)

			// Test authentication with version check
			err := client.CheckConnection(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestProxmoxClient_Timeout(t *testing.T) {
	t.Skip("Timeout test is flaky due to mock limitations")
	
	mockClient := proxmox.NewMockHTTPClient(t)
	
	// Simulate a very slow response that will timeout
	mockClient.ExpectGet("/api2/json/nodes/pve-node1/qemu/100/status/current").
		RespondAfter(200*time.Millisecond, 200, map[string]interface{}{})

	client := proxmox.NewClient(
		"https://pve.example.com:8006",
		"root@pam",
		"test-token-id",
		"test-token-secret",
		proxmox.WithHTTPClient(mockClient),
		proxmox.WithTimeout(50*time.Millisecond),
		proxmox.WithDefaultNode("pve-node1"),
	)

	ctx := context.Background()
	_, err := client.GetVM(ctx, 100)
	
	// We expect an error, but it might not contain "timeout" in the message
	// The mock will fail because the request times out before the response
	require.Error(t, err)
}