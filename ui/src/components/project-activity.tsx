'use client';

import { useState, useEffect } from 'react';
import { Clock, Users, Folder, Settings, Archive, Plus, Minus, Edit } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import {
  ProjectActivity,
  projectActivityApi,
} from '@/lib/api-client';
import { LoadingSpinner } from '@/components/ui/loading';
import { useWebSocket } from '@/hooks/use-websocket';

interface ProjectActivityProps {
  organizationId: string;
  projectId: string;
}

const activityIcons = {
  member_added: Users,
  member_removed: Users,
  member_role_changed: Users,
  namespace_created: Folder,
  namespace_deleted: Folder,
  quota_changed: Settings,
  project_created: Plus,
  project_updated: Edit,
  project_archived: Archive,
};

const activityColors = {
  member_added: 'text-green-600',
  member_removed: 'text-red-600',
  member_role_changed: 'text-blue-600',
  namespace_created: 'text-green-600',
  namespace_deleted: 'text-red-600',
  quota_changed: 'text-yellow-600',
  project_created: 'text-green-600',
  project_updated: 'text-blue-600',
  project_archived: 'text-gray-600',
};

export function ProjectActivityTimeline({ organizationId, projectId }: ProjectActivityProps) {
  const { toast } = useToast();
  const [activities, setActivities] = useState<ProjectActivity[]>([]);
  const [loading, setLoading] = useState(true);
  const [limit, setLimit] = useState(50);

  // Subscribe to project activity updates
  const { connected } = useWebSocket({
    organizationId,
    autoConnect: true,
  });

  useEffect(() => {
    loadActivities();
  }, [organizationId, projectId, limit]);

  // Listen for real-time activity updates
  useEffect(() => {
    if (!connected) return;

    const handleProjectActivity = (activity: ProjectActivity) => {
      if (activity.project_id === projectId) {
        setActivities((prev) => [activity, ...prev]);
      }
    };

    // Subscribe to project activity events
    window.addEventListener(`project:activity:${projectId}`, handleProjectActivity as any);

    return () => {
      window.removeEventListener(`project:activity:${projectId}`, handleProjectActivity as any);
    };
  }, [connected, projectId]);

  const loadActivities = async () => {
    try {
      const response = await projectActivityApi.list(organizationId, projectId, limit);
      setActivities(response.activities);
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to load project activity',
        variant: 'destructive',
      });
      console.error('Error loading activities:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatRelativeTime = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 1) return 'just now';
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    if (days < 7) return `${days}d ago`;
    return date.toLocaleDateString();
  };

  const formatActivityDescription = (activity: ProjectActivity) => {
    // Add metadata details to description if available
    if (activity.metadata) {
      switch (activity.type) {
        case 'member_role_changed':
          return `${activity.description} from ${activity.metadata.old_role} to ${activity.metadata.new_role}`;
        case 'quota_changed':
          return `${activity.description}: ${activity.metadata.resource} changed from ${activity.metadata.old_value} to ${activity.metadata.new_value}`;
        default:
          return activity.description;
      }
    }
    return activity.description;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <div>
          <h3 className="text-lg font-semibold">Activity Timeline</h3>
          <p className="text-sm text-gray-500">Recent changes and events in this project</p>
        </div>
        <div className="flex items-center gap-2">
          <Clock className="h-4 w-4 text-gray-400" />
          <select
            value={limit}
            onChange={(e) => setLimit(Number(e.target.value))}
            className="text-sm border border-gray-300 rounded-md px-2 py-1"
          >
            <option value={25}>Last 25</option>
            <option value={50}>Last 50</option>
            <option value={100}>Last 100</option>
            <option value={200}>Last 200</option>
          </select>
        </div>
      </div>

      {activities.length === 0 ? (
        <div className="text-center py-12 bg-gray-50 rounded-lg">
          <Clock className="h-12 w-12 text-gray-400 mx-auto mb-3" />
          <p className="text-gray-500">No activity yet</p>
          <p className="text-sm text-gray-400 mt-1">
            Activities will appear here as changes are made to the project
          </p>
        </div>
      ) : (
        <div className="flow-root">
          <ul role="list" className="-mb-8">
            {activities.map((activity, activityIdx) => {
              const Icon = activityIcons[activity.type] || Clock;
              const colorClass = activityColors[activity.type] || 'text-gray-600';
              
              return (
                <li key={activity.id}>
                  <div className="relative pb-8">
                    {activityIdx !== activities.length - 1 ? (
                      <span
                        className="absolute top-4 left-4 -ml-px h-full w-0.5 bg-gray-200"
                        aria-hidden="true"
                      />
                    ) : null}
                    <div className="relative flex space-x-3">
                      <div>
                        <span
                          className={`h-8 w-8 rounded-full flex items-center justify-center ring-8 ring-white ${
                            colorClass.replace('text-', 'bg-').replace('600', '100')
                          }`}
                        >
                          <Icon className={`h-4 w-4 ${colorClass}`} aria-hidden="true" />
                        </span>
                      </div>
                      <div className="flex min-w-0 flex-1 justify-between space-x-4 pt-1.5">
                        <div>
                          <p className="text-sm text-gray-900">
                            {formatActivityDescription(activity)}{' '}
                            <span className="font-medium text-gray-700">
                              by {activity.user_name || activity.user_email}
                            </span>
                          </p>
                        </div>
                        <div className="whitespace-nowrap text-right text-sm text-gray-500">
                          <time dateTime={activity.created_at}>
                            {formatRelativeTime(activity.created_at)}
                          </time>
                        </div>
                      </div>
                    </div>
                  </div>
                </li>
              );
            })}
          </ul>
        </div>
      )}

      {/* Load More */}
      {activities.length === limit && (
        <div className="text-center pt-4">
          <button
            onClick={() => setLimit(limit + 50)}
            className="text-sm text-blue-600 hover:text-blue-800"
          >
            Load more activities
          </button>
        </div>
      )}
    </div>
  );
}