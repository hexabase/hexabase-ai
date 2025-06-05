import { io, Socket } from 'socket.io-client';
import Cookies from 'js-cookie';

export interface WorkspaceStatusUpdate {
  workspace_id: string;
  organization_id: string;
  status: string;
  message?: string;
  timestamp: string;
}

export interface TaskProgressUpdate {
  task_id: string;
  workspace_id: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  progress?: number;
  message?: string;
  error?: string;
  timestamp: string;
}

export interface VClusterHealthUpdate {
  workspace_id: string;
  healthy: boolean;
  components: Record<string, string>;
  resource_usage: Record<string, string>;
  timestamp: string;
}

export interface ProjectUpdate {
  project_id: string;
  type: 'status' | 'resource' | 'namespace' | 'member' | 'activity';
  data: any;
  timestamp: string;
}

export interface NamespaceUpdate {
  namespace_id: string;
  project_id: string;
  type: 'created' | 'updated' | 'deleted' | 'usage';
  data: any;
  timestamp: string;
}

class WebSocketService {
  private socket: Socket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private listeners: Map<string, Set<(...args: any[]) => void>> = new Map();

  connect(organizationId?: string) {
    if (this.socket?.connected) {
      return;
    }

    const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080';
    const token = Cookies.get('hexabase_access_token');

    this.socket = io(wsUrl, {
      auth: {
        token,
        organization_id: organizationId,
      },
      transports: ['websocket'],
      reconnection: true,
      reconnectionAttempts: this.maxReconnectAttempts,
      reconnectionDelay: this.reconnectDelay,
    });

    this.setupEventHandlers();
  }

  private setupEventHandlers() {
    if (!this.socket) return;

    this.socket.on('connect', () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      this.emit('connected', { timestamp: new Date().toISOString() });
    });

    this.socket.on('disconnect', (reason) => {
      console.log('WebSocket disconnected:', reason);
      this.emit('disconnected', { reason, timestamp: new Date().toISOString() });
    });

    this.socket.on('connect_error', (error) => {
      console.error('WebSocket connection error:', error);
      this.reconnectAttempts++;
      
      if (this.reconnectAttempts >= this.maxReconnectAttempts) {
        this.emit('connection_failed', { 
          error: error.message, 
          attempts: this.reconnectAttempts,
          timestamp: new Date().toISOString() 
        });
      }
    });

    // Workspace status updates
    this.socket.on('workspace:status', (data: WorkspaceStatusUpdate) => {
      this.emit('workspace:status', data);
    });

    // Task progress updates
    this.socket.on('task:progress', (data: TaskProgressUpdate) => {
      this.emit('task:progress', data);
    });

    // vCluster health updates
    this.socket.on('vcluster:health', (data: VClusterHealthUpdate) => {
      this.emit('vcluster:health', data);
    });

    // Project updates
    this.socket.on('project:update', (data: ProjectUpdate) => {
      this.emit('project:update', data);
      // Emit specific event for the project
      this.emit(`project:update:${data.project_id}`, data);
    });

    // Namespace updates
    this.socket.on('namespace:update', (data: NamespaceUpdate) => {
      this.emit('namespace:update', data);
      // Emit specific events for the project and namespace
      this.emit(`project:namespace:${data.project_id}`, data);
      this.emit(`namespace:update:${data.namespace_id}`, data);
    });

    // Project activity events
    this.socket.on('project:activity', (data: any) => {
      this.emit('project:activity', data);
      // Emit specific event for the project
      this.emit(`project:activity:${data.project_id}`, data);
    });

    // Error events
    this.socket.on('error', (error: any) => {
      console.error('WebSocket error:', error);
      this.emit('error', { error, timestamp: new Date().toISOString() });
    });
  }

  disconnect() {
    if (this.socket) {
      this.socket.disconnect();
      this.socket = null;
    }
    this.listeners.clear();
  }

  // Subscribe to workspace updates
  subscribeToWorkspace(workspaceId: string) {
    if (!this.socket?.connected) {
      console.warn('WebSocket not connected');
      return;
    }

    this.socket.emit('subscribe:workspace', { workspace_id: workspaceId });
  }

  unsubscribeFromWorkspace(workspaceId: string) {
    if (!this.socket?.connected) return;

    this.socket.emit('unsubscribe:workspace', { workspace_id: workspaceId });
  }

  // Subscribe to organization updates
  subscribeToOrganization(organizationId: string) {
    if (!this.socket?.connected) {
      console.warn('WebSocket not connected');
      return;
    }

    this.socket.emit('subscribe:organization', { organization_id: organizationId });
  }

  unsubscribeFromOrganization(organizationId: string) {
    if (!this.socket?.connected) return;

    this.socket.emit('unsubscribe:organization', { organization_id: organizationId });
  }

  // Subscribe to task updates
  subscribeToTask(taskId: string) {
    if (!this.socket?.connected) {
      console.warn('WebSocket not connected');
      return;
    }

    this.socket.emit('subscribe:task', { task_id: taskId });
  }

  unsubscribeFromTask(taskId: string) {
    if (!this.socket?.connected) return;

    this.socket.emit('unsubscribe:task', { task_id: taskId });
  }

  // Subscribe to project updates
  subscribeToProject(projectId: string) {
    if (!this.socket?.connected) {
      console.warn('WebSocket not connected');
      return;
    }

    this.socket.emit('subscribe:project', { project_id: projectId });
  }

  unsubscribeFromProject(projectId: string) {
    if (!this.socket?.connected) return;

    this.socket.emit('unsubscribe:project', { project_id: projectId });
  }

  // Subscribe to namespace updates
  subscribeToNamespace(namespaceId: string) {
    if (!this.socket?.connected) {
      console.warn('WebSocket not connected');
      return;
    }

    this.socket.emit('subscribe:namespace', { namespace_id: namespaceId });
  }

  unsubscribeFromNamespace(namespaceId: string) {
    if (!this.socket?.connected) return;

    this.socket.emit('unsubscribe:namespace', { namespace_id: namespaceId });
  }

  // Event listener management
  on(event: string, callback: (...args: any[]) => void) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event)!.add(callback);

    // Return unsubscribe function
    return () => {
      const callbacks = this.listeners.get(event);
      if (callbacks) {
        callbacks.delete(callback);
        if (callbacks.size === 0) {
          this.listeners.delete(event);
        }
      }
    };
  }

  off(event: string, callback?: (...args: any[]) => void) {
    if (!callback) {
      this.listeners.delete(event);
    } else {
      const callbacks = this.listeners.get(event);
      if (callbacks) {
        callbacks.delete(callback);
        if (callbacks.size === 0) {
          this.listeners.delete(event);
        }
      }
    }
  }

  private emit(event: string, data: any) {
    const callbacks = this.listeners.get(event);
    if (callbacks) {
      callbacks.forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in WebSocket event handler for ${event}:`, error);
        }
      });
    }
  }

  // Check connection status
  isConnected(): boolean {
    return this.socket?.connected || false;
  }

  // Get socket instance (for advanced use cases)
  getSocket(): Socket | null {
    return this.socket;
  }
}

// Export singleton instance
export const wsService = new WebSocketService();

// React hook for WebSocket
export function useWebSocket() {
  return wsService;
}