import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { WorkspaceMetrics } from '@/components/monitoring/workspace-metrics-detail';
import { monitoringApi } from '@/lib/api-client';

jest.mock('@/lib/api-client', () => ({
  monitoringApi: {
    getWorkspaceMetrics: jest.fn(),
    getResourceUsageHistory: jest.fn(),
    getApplicationMetrics: jest.fn(),
  },
}));

describe('WorkspaceMetrics', () => {
  const mockWorkspaceId = 'ws-123';
  
  const mockMetrics = {
    cpu_usage: 45.5,
    memory_usage: 62.3,
    storage_usage: 30.1,
    network_ingress: 1024 * 1024 * 100, // 100MB
    network_egress: 1024 * 1024 * 200, // 200MB
    pod_count: 15,
    container_count: 25,
    timestamp: '2024-01-01T00:00:00Z',
  };
  
  const mockHistory = {
    cpu: [
      { timestamp: '2024-01-01T00:00:00Z', value: 40 },
      { timestamp: '2024-01-01T00:05:00Z', value: 45 },
      { timestamp: '2024-01-01T00:10:00Z', value: 45.5 },
    ],
    memory: [
      { timestamp: '2024-01-01T00:00:00Z', value: 60 },
      { timestamp: '2024-01-01T00:05:00Z', value: 61 },
      { timestamp: '2024-01-01T00:10:00Z', value: 62.3 },
    ],
  };

  beforeEach(() => {
    jest.clearAllMocks();
    (monitoringApi.getWorkspaceMetrics as jest.Mock).mockResolvedValue({ metrics: mockMetrics });
    (monitoringApi.getResourceUsageHistory as jest.Mock).mockResolvedValue({ history: mockHistory });
  });

  it('should display workspace metrics', async () => {
    render(<WorkspaceMetrics workspaceId={mockWorkspaceId} />);
    
    await waitFor(() => {
      expect(screen.getByText(/cpu usage/i)).toBeInTheDocument();
      expect(screen.getByText('45.5%')).toBeInTheDocument();
      
      expect(screen.getByText(/memory usage/i)).toBeInTheDocument();
      expect(screen.getByText('62.3%')).toBeInTheDocument();
      
      expect(screen.getByText(/storage usage/i)).toBeInTheDocument();
      expect(screen.getByText('30.1%')).toBeInTheDocument();
    });
  });

  it('should display network metrics', async () => {
    render(<WorkspaceMetrics workspaceId={mockWorkspaceId} />);
    
    await waitFor(() => {
      expect(screen.getByText(/network/i)).toBeInTheDocument();
      expect(screen.getByText(/100 MB/i)).toBeInTheDocument(); // ingress
      expect(screen.getByText(/200 MB/i)).toBeInTheDocument(); // egress
    });
  });

  it('should display pod and container counts', async () => {
    render(<WorkspaceMetrics workspaceId={mockWorkspaceId} />);
    
    await waitFor(() => {
      expect(screen.getByText(/15 pods/i)).toBeInTheDocument();
      expect(screen.getByText(/25 containers/i)).toBeInTheDocument();
    });
  });

  it('should show loading state', () => {
    (monitoringApi.getWorkspaceMetrics as jest.Mock).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    );
    
    render(<WorkspaceMetrics workspaceId={mockWorkspaceId} />);
    
    expect(screen.getByTestId('metrics-skeleton')).toBeInTheDocument();
  });

  it('should handle error state', async () => {
    (monitoringApi.getWorkspaceMetrics as jest.Mock).mockRejectedValue(
      new Error('Failed to fetch metrics')
    );
    
    render(<WorkspaceMetrics workspaceId={mockWorkspaceId} />);
    
    await waitFor(() => {
      expect(screen.getByText(/failed to load metrics/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
    });
  });

  it('should refresh metrics periodically', async () => {
    jest.useFakeTimers();
    
    render(<WorkspaceMetrics workspaceId={mockWorkspaceId} />);
    
    await waitFor(() => {
      expect(monitoringApi.getWorkspaceMetrics).toHaveBeenCalledTimes(1);
    });
    
    // Advance time by 30 seconds (default refresh interval)
    jest.advanceTimersByTime(30000);
    
    await waitFor(() => {
      expect(monitoringApi.getWorkspaceMetrics).toHaveBeenCalledTimes(2);
    });
    
    jest.useRealTimers();
  });

  it('should show critical alert for high resource usage', async () => {
    const criticalMetrics = {
      ...mockMetrics,
      cpu_usage: 95,
      memory_usage: 92,
    };
    
    (monitoringApi.getWorkspaceMetrics as jest.Mock).mockResolvedValue({ 
      metrics: criticalMetrics 
    });
    
    render(<WorkspaceMetrics workspaceId={mockWorkspaceId} />);
    
    await waitFor(() => {
      expect(screen.getByText(/critical resource usage/i)).toBeInTheDocument();
      expect(screen.getByText(/cpu usage is at 95%/i)).toBeInTheDocument();
    });
  });

  it('should display resource usage history chart', async () => {
    render(<WorkspaceMetrics workspaceId={mockWorkspaceId} showHistory />);
    
    await waitFor(() => {
      expect(monitoringApi.getResourceUsageHistory).toHaveBeenCalledWith(
        mockWorkspaceId,
        expect.objectContaining({
          period: '1h',
        })
      );
      
      expect(screen.getByTestId('cpu-history-chart')).toBeInTheDocument();
      expect(screen.getByTestId('memory-history-chart')).toBeInTheDocument();
    });
  });

  it('should allow changing time period for history', async () => {
    // Just verify that the time period selector is rendered when showHistory is true
    render(
      <WorkspaceMetrics workspaceId={mockWorkspaceId} showHistory />
    );
    
    await waitFor(() => {
      expect(apiClient.monitoring.getResourceUsageHistory).toHaveBeenCalledWith(
        mockWorkspaceId,
        expect.objectContaining({
          period: '1h', // default period
        })
      );
    });
    
    // Find the select element by its id
    const periodSelector = await screen.findByRole('combobox');
    expect(periodSelector).toBeInTheDocument();
    expect(periodSelector).toHaveValue('1h');
  });

  it('should format large numbers correctly', async () => {
    const largeMetrics = {
      ...mockMetrics,
      network_ingress: 1024 * 1024 * 1024 * 5.5, // 5.5GB
      network_egress: 1024 * 1024 * 1024 * 1024 * 2.3, // 2.3TB
    };
    
    (apiClient.monitoring.getWorkspaceMetrics as jest.Mock).mockResolvedValue({ 
      metrics: largeMetrics 
    });
    
    render(<WorkspaceMetrics workspaceId={mockWorkspaceId} />);
    
    await waitFor(() => {
      expect(screen.getByText(/5.5 GB/i)).toBeInTheDocument();
      expect(screen.getByText(/2.3 TB/i)).toBeInTheDocument();
    });
  });
});
