import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AIChat } from '@/components/aiops/ai-chat';
import { apiClient } from '@/lib/api-client';

jest.mock('@/lib/api-client', () => ({
  apiClient: {
    aiops: {
      chat: jest.fn(),
      getSuggestions: jest.fn(),
      analyzeMetrics: jest.fn(),
    },
  },
}));

describe('AIChat', () => {
  const user = userEvent.setup();
  const mockWorkspaceId = 'ws-123';
  
  const mockChatResponse = {
    message: 'Based on your current resource usage, everything looks healthy.',
    suggestions: [
      'Consider enabling autoscaling for better resource efficiency',
      'Your memory usage pattern suggests you could reduce allocation by 20%',
    ],
    context: {
      analyzed_metrics: {
        cpu_usage: 45,
        memory_usage: 60,
      },
    },
  };

  beforeEach(() => {
    jest.clearAllMocks();
    (apiClient.aiops.chat as jest.Mock).mockResolvedValue(mockChatResponse);
    (apiClient.aiops.getSuggestions as jest.Mock).mockResolvedValue({
      suggestions: [
        {
          id: 'sug-1',
          type: 'optimization',
          title: 'Enable HPA for Frontend App',
          description: 'Your frontend app shows variable load patterns',
          priority: 'medium',
          estimated_savings: '~20% resource cost',
        },
        {
          id: 'sug-2',
          type: 'security',
          title: 'Update container images',
          description: '3 containers are running outdated images',
          priority: 'high',
        },
      ],
    });
  });

  it('should display AI chat interface', () => {
    render(<AIChat workspaceId={mockWorkspaceId} />);
    
    expect(screen.getByText(/ai assistant/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/ask ai about your infrastructure/i)).toBeInTheDocument();
  });

  it('should send message and display response', async () => {
    render(<AIChat workspaceId={mockWorkspaceId} />);
    
    const input = screen.getByPlaceholderText(/ask ai about your infrastructure/i);
    const sendButton = screen.getByRole('button', { name: /send/i });
    
    await user.type(input, 'How is my resource usage?');
    await user.click(sendButton);
    
    await waitFor(() => {
      expect(apiClient.aiops.chat).toHaveBeenCalledWith({
        message: 'How is my resource usage?',
        context: {
          workspace_id: mockWorkspaceId,
        },
      });
      
      expect(screen.getByText(/everything looks healthy/i)).toBeInTheDocument();
    });
  });

  it('should display AI suggestions', async () => {
    render(<AIChat workspaceId={mockWorkspaceId} />);
    
    // Open suggestions panel
    const suggestionsButton = screen.getByRole('button', { name: /view suggestions/i });
    await user.click(suggestionsButton);
    
    await waitFor(() => {
      expect(screen.getByText(/enable hpa for frontend app/i)).toBeInTheDocument();
      expect(screen.getByText(/update container images/i)).toBeInTheDocument();
      expect(screen.getByText(/~20% resource cost/i)).toBeInTheDocument();
    });
  });

  it('should show loading state while sending message', async () => {
    (apiClient.aiops.chat as jest.Mock).mockImplementation(
      () => new Promise(resolve => setTimeout(resolve, 1000))
    );
    
    render(<AIChat workspaceId={mockWorkspaceId} />);
    
    const input = screen.getByPlaceholderText(/ask ai about your infrastructure/i);
    const sendButton = screen.getByRole('button', { name: /send/i });
    
    await user.type(input, 'Test message');
    await user.click(sendButton);
    
    expect(screen.getByTestId('ai-loading')).toBeInTheDocument();
    expect(sendButton).toBeDisabled();
  });

  it('should handle chat errors gracefully', async () => {
    (apiClient.aiops.chat as jest.Mock).mockRejectedValue(
      new Error('AI service unavailable')
    );
    
    render(<AIChat workspaceId={mockWorkspaceId} />);
    
    const input = screen.getByPlaceholderText(/ask ai about your infrastructure/i);
    const sendButton = screen.getByRole('button', { name: /send/i });
    
    await user.type(input, 'Test message');
    await user.click(sendButton);
    
    await waitFor(() => {
      expect(screen.getByText(/failed to get ai response/i)).toBeInTheDocument();
    });
  });

  it('should maintain chat history', async () => {
    render(<AIChat workspaceId={mockWorkspaceId} />);
    
    const input = screen.getByPlaceholderText(/ask ai about your infrastructure/i);
    const sendButton = screen.getByRole('button', { name: /send/i });
    
    // Send first message
    await user.type(input, 'First question');
    await user.click(sendButton);
    
    await waitFor(() => {
      expect(screen.getByText('First question')).toBeInTheDocument();
    });
    
    // Send second message
    await user.clear(input);
    await user.type(input, 'Second question');
    await user.click(sendButton);
    
    await waitFor(() => {
      expect(screen.getByText('First question')).toBeInTheDocument();
      expect(screen.getByText('Second question')).toBeInTheDocument();
    });
  });

  it('should provide quick action buttons', async () => {
    render(<AIChat workspaceId={mockWorkspaceId} />);
    
    const quickActions = [
      'Check resource usage',
      'Analyze costs',
      'Security scan',
      'Performance tips',
    ];
    
    quickActions.forEach(action => {
      expect(screen.getByRole('button', { name: action })).toBeInTheDocument();
    });
    
    // Click a quick action
    const resourceButton = screen.getByRole('button', { name: /check resource usage/i });
    await user.click(resourceButton);
    
    await waitFor(() => {
      expect(apiClient.aiops.chat).toHaveBeenCalledWith({
        message: 'Check resource usage',
        context: {
          workspace_id: mockWorkspaceId,
        },
      });
    });
  });

  it('should analyze current metrics when requested', async () => {
    (apiClient.aiops.analyzeMetrics as jest.Mock).mockResolvedValue({
      analysis: {
        summary: 'Your infrastructure is running efficiently',
        findings: [
          {
            type: 'optimization',
            description: 'CPU usage is optimal at 45%',
            severity: 'info',
          },
          {
            type: 'warning',
            description: 'Memory usage trending upward',
            severity: 'warning',
          },
        ],
        recommendations: [
          'Consider implementing memory limits',
          'Enable monitoring alerts for memory > 80%',
        ],
      },
    });
    
    render(<AIChat workspaceId={mockWorkspaceId} />);
    
    const analyzeButton = screen.getByRole('button', { name: /analyze metrics/i });
    await user.click(analyzeButton);
    
    await waitFor(() => {
      expect(apiClient.aiops.analyzeMetrics).toHaveBeenCalledWith(mockWorkspaceId);
      expect(screen.getByText(/infrastructure is running efficiently/i)).toBeInTheDocument();
      expect(screen.getByText(/memory usage trending upward/i)).toBeInTheDocument();
    });
  });

  it('should export chat history', async () => {
    render(<AIChat workspaceId={mockWorkspaceId} />);
    
    // Send some messages first
    const input = screen.getByPlaceholderText(/ask ai about your infrastructure/i);
    await user.type(input, 'Test message');
    await user.click(screen.getByRole('button', { name: /send/i }));
    
    await waitFor(() => {
      expect(screen.getByText('Test message')).toBeInTheDocument();
    });
    
    // Export chat
    const exportButton = screen.getByRole('button', { name: /export chat/i });
    await user.click(exportButton);
    
    // Should trigger download
    expect(screen.getByText(/chat history exported/i)).toBeInTheDocument();
  });

  it('should clear chat history', async () => {
    render(<AIChat workspaceId={mockWorkspaceId} />);
    
    // Send a message
    const input = screen.getByPlaceholderText(/ask ai about your infrastructure/i);
    await user.type(input, 'Test message');
    await user.click(screen.getByRole('button', { name: /send/i }));
    
    await waitFor(() => {
      expect(screen.getByText('Test message')).toBeInTheDocument();
    });
    
    // Clear chat
    const clearButton = screen.getByRole('button', { name: /clear chat/i });
    await user.click(clearButton);
    
    // Confirm clear
    const confirmButton = screen.getByRole('button', { name: /confirm clear/i });
    await user.click(confirmButton);
    
    expect(screen.queryByText('Test message')).not.toBeInTheDocument();
  });
});