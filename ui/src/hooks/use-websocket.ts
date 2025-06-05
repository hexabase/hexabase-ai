import { useEffect, useCallback, useRef } from 'react';
import { wsService, WorkspaceStatusUpdate, TaskProgressUpdate, VClusterHealthUpdate } from '@/lib/websocket';

export interface UseWebSocketOptions {
  organizationId?: string;
  workspaceId?: string;
  taskId?: string;
  autoConnect?: boolean;
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const { organizationId, workspaceId, taskId, autoConnect = true } = options;
  const connectedRef = useRef(false);

  // Connect to WebSocket
  useEffect(() => {
    if (autoConnect && organizationId && !connectedRef.current) {
      wsService.connect(organizationId);
      connectedRef.current = true;
    }

    return () => {
      if (connectedRef.current) {
        wsService.disconnect();
        connectedRef.current = false;
      }
    };
  }, [organizationId, autoConnect]);

  // Subscribe to workspace updates
  useEffect(() => {
    if (workspaceId && wsService.isConnected()) {
      wsService.subscribeToWorkspace(workspaceId);
      
      return () => {
        wsService.unsubscribeFromWorkspace(workspaceId);
      };
    }
  }, [workspaceId]);

  // Subscribe to task updates
  useEffect(() => {
    if (taskId && wsService.isConnected()) {
      wsService.subscribeToTask(taskId);
      
      return () => {
        wsService.unsubscribeFromTask(taskId);
      };
    }
  }, [taskId]);

  const onWorkspaceStatus = useCallback((callback: (data: WorkspaceStatusUpdate) => void) => {
    return wsService.on('workspace:status', callback);
  }, []);

  const onTaskProgress = useCallback((callback: (data: TaskProgressUpdate) => void) => {
    return wsService.on('task:progress', callback);
  }, []);

  const onVClusterHealth = useCallback((callback: (data: VClusterHealthUpdate) => void) => {
    return wsService.on('vcluster:health', callback);
  }, []);

  const onConnected = useCallback((callback: (data: any) => void) => {
    return wsService.on('connected', callback);
  }, []);

  const onDisconnected = useCallback((callback: (data: any) => void) => {
    return wsService.on('disconnected', callback);
  }, []);

  const onError = useCallback((callback: (data: any) => void) => {
    return wsService.on('error', callback);
  }, []);

  return {
    isConnected: wsService.isConnected(),
    onWorkspaceStatus,
    onTaskProgress,
    onVClusterHealth,
    onConnected,
    onDisconnected,
    onError,
    connect: (orgId?: string) => wsService.connect(orgId || organizationId),
    disconnect: () => wsService.disconnect(),
    subscribeToWorkspace: (wsId: string) => wsService.subscribeToWorkspace(wsId),
    unsubscribeFromWorkspace: (wsId: string) => wsService.unsubscribeFromWorkspace(wsId),
    subscribeToTask: (tid: string) => wsService.subscribeToTask(tid),
    unsubscribeFromTask: (tid: string) => wsService.unsubscribeFromTask(tid),
  };
}

// Hook for workspace-specific updates
export function useWorkspaceUpdates(organizationId: string, workspaceId: string) {
  const { onWorkspaceStatus, onVClusterHealth, ...rest } = useWebSocket({
    organizationId,
    workspaceId,
    autoConnect: true,
  });

  return {
    onWorkspaceStatus,
    onVClusterHealth,
    ...rest,
  };
}

// Hook for task monitoring
export function useTaskMonitoring(taskId: string, organizationId?: string) {
  const { onTaskProgress, ...rest } = useWebSocket({
    organizationId,
    taskId,
    autoConnect: true,
  });

  return {
    onTaskProgress,
    ...rest,
  };
}

// Hook for organization-wide updates
export function useOrganizationUpdates(organizationId: string) {
  const websocket = useWebSocket({
    organizationId,
    autoConnect: true,
  });

  useEffect(() => {
    if (websocket.isConnected) {
      wsService.subscribeToOrganization(organizationId);
      
      return () => {
        wsService.unsubscribeFromOrganization(organizationId);
      };
    }
  }, [organizationId, websocket.isConnected]);

  return websocket;
}