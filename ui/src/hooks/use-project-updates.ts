import { useEffect, useState, useCallback } from 'react';
import { useWebSocket } from './use-websocket';
import { ProjectUpdate, NamespaceUpdate } from '@/lib/websocket';

interface UseProjectUpdatesOptions {
  projectId: string;
  organizationId?: string;
  autoConnect?: boolean;
  onProjectUpdate?: (update: ProjectUpdate) => void;
  onNamespaceUpdate?: (update: NamespaceUpdate) => void;
  onActivityUpdate?: (activity: any) => void;
}

export function useProjectUpdates({
  projectId,
  organizationId,
  autoConnect = true,
  onProjectUpdate,
  onNamespaceUpdate,
  onActivityUpdate,
}: UseProjectUpdatesOptions) {
  const [latestUpdate, setLatestUpdate] = useState<ProjectUpdate | null>(null);
  const [latestNamespaceUpdate, setLatestNamespaceUpdate] = useState<NamespaceUpdate | null>(null);
  
  const { connected, subscribeToProject, unsubscribeFromProject } = useWebSocket({
    organizationId,
    projectId,
    autoConnect,
  });

  // Handle project updates
  const handleProjectUpdate = useCallback((update: ProjectUpdate) => {
    if (update.project_id === projectId) {
      setLatestUpdate(update);
      onProjectUpdate?.(update);
    }
  }, [projectId, onProjectUpdate]);

  // Handle namespace updates
  const handleNamespaceUpdate = useCallback((update: NamespaceUpdate) => {
    if (update.project_id === projectId) {
      setLatestNamespaceUpdate(update);
      onNamespaceUpdate?.(update);
    }
  }, [projectId, onNamespaceUpdate]);

  // Handle activity updates
  const handleActivityUpdate = useCallback((activity: any) => {
    if (activity.project_id === projectId) {
      onActivityUpdate?.(activity);
    }
  }, [projectId, onActivityUpdate]);

  useEffect(() => {
    if (!connected || !projectId) return;

    // Subscribe to project updates
    subscribeToProject(projectId);

    // Set up event listeners
    const projectUpdateHandler = (e: CustomEvent) => handleProjectUpdate(e.detail);
    const namespaceUpdateHandler = (e: CustomEvent) => handleNamespaceUpdate(e.detail);
    const activityUpdateHandler = (e: CustomEvent) => handleActivityUpdate(e.detail);

    window.addEventListener(`project:update:${projectId}`, projectUpdateHandler as any);
    window.addEventListener(`project:namespace:${projectId}`, namespaceUpdateHandler as any);
    window.addEventListener(`project:activity:${projectId}`, activityUpdateHandler as any);

    return () => {
      // Unsubscribe and remove listeners
      unsubscribeFromProject(projectId);
      window.removeEventListener(`project:update:${projectId}`, projectUpdateHandler as any);
      window.removeEventListener(`project:namespace:${projectId}`, namespaceUpdateHandler as any);
      window.removeEventListener(`project:activity:${projectId}`, activityUpdateHandler as any);
    };
  }, [
    connected,
    projectId,
    subscribeToProject,
    unsubscribeFromProject,
    handleProjectUpdate,
    handleNamespaceUpdate,
    handleActivityUpdate,
  ]);

  return {
    connected,
    latestUpdate,
    latestNamespaceUpdate,
  };
}