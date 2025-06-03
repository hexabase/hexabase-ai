'use client';

import { Workspace } from '@/lib/api-client';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { 
  Play, 
  Square, 
  MoreHorizontal, 
  ExternalLink,
  Activity,
  Clock,
  AlertCircle,
  CheckCircle
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface WorkspaceCardProps {
  workspace: Workspace;
  onStart?: (workspaceId: string) => void;
  onStop?: (workspaceId: string) => void;
  onView?: (workspaceId: string) => void;
  onDelete?: (workspaceId: string) => void;
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'RUNNING':
      return 'bg-green-100 text-green-800 border-green-200';
    case 'STOPPED':
      return 'bg-gray-100 text-gray-800 border-gray-200';
    case 'PENDING_CREATION':
    case 'STARTING':
    case 'STOPPING':
      return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    case 'ERROR':
      return 'bg-red-100 text-red-800 border-red-200';
    default:
      return 'bg-gray-100 text-gray-800 border-gray-200';
  }
};

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'RUNNING':
      return <CheckCircle className="w-4 h-4" />;
    case 'STOPPED':
      return <Square className="w-4 h-4" />;
    case 'PENDING_CREATION':
    case 'STARTING':
    case 'STOPPING':
      return <Clock className="w-4 h-4" />;
    case 'ERROR':
      return <AlertCircle className="w-4 h-4" />;
    default:
      return <Activity className="w-4 h-4" />;
  }
};

const formatStatus = (status: string) => {
  return status.replace(/_/g, ' ').toLowerCase().replace(/\b\w/g, l => l.toUpperCase());
};

export function WorkspaceCard({ workspace, onStart, onStop, onView, onDelete }: WorkspaceCardProps) {
  const isRunning = workspace.vcluster_status === 'RUNNING';
  const isStopped = workspace.vcluster_status === 'STOPPED';
  const isTransitioning = ['PENDING_CREATION', 'STARTING', 'STOPPING', 'CONFIGURING_HNC'].includes(workspace.vcluster_status);

  return (
    <Card 
      className="hover:shadow-md transition-shadow cursor-pointer"
      data-testid="workspace-card"
      onClick={() => onView?.(workspace.id)}
    >
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-lg font-semibold">{workspace.name}</CardTitle>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button 
              variant="ghost" 
              className="h-8 w-8 p-0"
              onClick={(e) => e.stopPropagation()}
            >
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={(e) => {
              e.stopPropagation();
              onView?.(workspace.id);
            }}>
              <ExternalLink className="mr-2 h-4 w-4" />
              View Details
            </DropdownMenuItem>
            {isRunning && (
              <DropdownMenuItem onClick={(e) => {
                e.stopPropagation();
                onStop?.(workspace.id);
              }}>
                <Square className="mr-2 h-4 w-4" />
                Stop vCluster
              </DropdownMenuItem>
            )}
            {isStopped && (
              <DropdownMenuItem onClick={(e) => {
                e.stopPropagation();
                onStart?.(workspace.id);
              }}>
                <Play className="mr-2 h-4 w-4" />
                Start vCluster
              </DropdownMenuItem>
            )}
            <DropdownMenuItem 
              onClick={(e) => {
                e.stopPropagation();
                onDelete?.(workspace.id);
              }}
              className="text-red-600"
            >
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {/* Status Badge */}
          <div className="flex items-center space-x-2">
            <Badge 
              variant="outline" 
              className={getStatusColor(workspace.vcluster_status)}
              data-testid={`status-${workspace.vcluster_status.toLowerCase()}`}
            >
              {getStatusIcon(workspace.vcluster_status)}
              <span className="ml-1">{formatStatus(workspace.vcluster_status)}</span>
            </Badge>
          </div>

          {/* Workspace Info */}
          <div className="text-sm text-gray-600 space-y-1">
            <div>Plan: <span className="font-medium">{workspace.plan_id}</span></div>
            <div>Created: {new Date(workspace.created_at).toLocaleDateString()}</div>
            {workspace.vcluster_instance_name && (
              <div>Instance: <span className="font-mono text-xs">{workspace.vcluster_instance_name}</span></div>
            )}
          </div>

          {/* Quick Actions */}
          <div className="flex space-x-2 pt-2">
            {isRunning && (
              <Button 
                size="sm" 
                variant="outline"
                onClick={(e) => {
                  e.stopPropagation();
                  onStop?.(workspace.id);
                }}
                data-testid="stop-vcluster"
              >
                <Square className="w-3 h-3 mr-1" />
                Stop
              </Button>
            )}
            {isStopped && (
              <Button 
                size="sm" 
                variant="outline"
                onClick={(e) => {
                  e.stopPropagation();
                  onStart?.(workspace.id);
                }}
                data-testid="start-vcluster"
              >
                <Play className="w-3 h-3 mr-1" />
                Start
              </Button>
            )}
            {isTransitioning && (
              <Button size="sm" variant="outline" disabled>
                <Clock className="w-3 h-3 mr-1" />
                {formatStatus(workspace.vcluster_status)}
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}