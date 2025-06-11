'use client';

import { useState } from 'react';
import { Application } from '@/lib/api-client';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Edit,
  Trash2,
  Play,
  Pause,
  Globe,
  GitBranch,
  Package,
  Server,
  Cpu,
  MemoryStick,
  Clock,
  AlertCircle,
  CheckCircle,
  Loader2,
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface ApplicationCardProps {
  application: Application;
  onClick: (applicationId: string) => void;
  onEdit?: (application: Application) => void;
  onDelete?: (applicationId: string) => void;
  onStatusChange?: (applicationId: string, newStatus: string) => void;
}

export function ApplicationCard({
  application,
  onClick,
  onEdit,
  onDelete,
  onStatusChange,
}: ApplicationCardProps) {
  const [isUpdating, setIsUpdating] = useState(false);

  const getStatusVariant = (status: string): "default" | "secondary" | "destructive" | "outline" => {
    switch (status) {
      case 'running':
      case 'active':
        return 'default';
      case 'pending':
      case 'creating':
        return 'secondary';
      case 'error':
        return 'destructive';
      case 'suspended':
      case 'terminating':
        return 'outline';
      default:
        return 'secondary';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running':
      case 'active':
        return <CheckCircle className="h-3 w-3" />;
      case 'pending':
      case 'creating':
        return <Loader2 className="h-3 w-3 animate-spin" />;
      case 'error':
        return <AlertCircle className="h-3 w-3" />;
      case 'suspended':
        return <Pause className="h-3 w-3" />;
      case 'terminating':
        return <Clock className="h-3 w-3" />;
      default:
        return null;
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'stateless':
        return <Server className="h-4 w-4" />;
      case 'stateful':
        return <Package className="h-4 w-4" />;
      case 'cronjob':
        return <Clock className="h-4 w-4" />;
      case 'function':
        return <GitBranch className="h-4 w-4" />;
      default:
        return <Server className="h-4 w-4" />;
    }
  };

  const handleCardClick = (e: React.MouseEvent) => {
    if ((e.target as HTMLElement).closest('button')) {
      return;
    }
    onClick(application.id);
  };

  const handleEdit = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (onEdit) {
      onEdit(application);
    }
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (onDelete) {
      onDelete(application.id);
    }
  };

  const handleStatusToggle = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!onStatusChange) return;

    const newStatus = application.status === 'running' ? 'suspended' : 'running';
    setIsUpdating(true);
    try {
      await onStatusChange(application.id, newStatus);
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <Card
      className="cursor-pointer hover:shadow-lg transition-shadow"
      onClick={handleCardClick}
      data-testid={`app-card-${application.id}`}
    >
      <CardHeader>
        <div className="flex justify-between items-start">
          <div className="flex items-center gap-2">
            {getTypeIcon(application.type)}
            <CardTitle className="text-lg">{application.name}</CardTitle>
          </div>
          <Badge
            variant={getStatusVariant(application.status)}
            className="flex items-center gap-1"
          >
            {getStatusIcon(application.status)}
            {application.status}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {/* Source Information */}
          <div className="text-sm text-muted-foreground">
            {application.source_type === 'image' && application.source_image && (
              <div className="flex items-center gap-2">
                <Package className="h-3 w-3" />
                <span>{application.source_image}</span>
              </div>
            )}
            {application.source_type === 'git' && application.source_git_url && (
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <GitBranch className="h-3 w-3" />
                  <span className="truncate">{application.source_git_url}</span>
                </div>
                {application.source_git_ref && (
                  <span className="ml-5 text-xs">{application.source_git_ref}</span>
                )}
              </div>
            )}
          </div>

          {/* Type Badge */}
          <div>
            <Badge variant="outline" className="text-xs">
              {application.type}
            </Badge>
          </div>

          {/* Endpoints */}
          {application.endpoints?.external && (
            <div className="flex items-center gap-2 text-sm">
              <Globe className="h-3 w-3 text-blue-500" />
              <a
                href={application.endpoints.external}
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-500 hover:underline truncate"
                onClick={(e) => e.stopPropagation()}
              >
                {application.endpoints.external.replace(/^https?:\/\//, '')}
              </a>
            </div>
          )}

          {/* Configuration */}
          {application.config && (
            <div className="grid grid-cols-3 gap-2 text-xs">
              {application.config.replicas !== undefined && (
                <div className="text-center">
                  <div className="font-semibold">{application.config.replicas} replicas</div>
                </div>
              )}
              {application.config.cpu && (
                <div className="text-center">
                  <div className="font-semibold">{application.config.cpu} CPU</div>
                </div>
              )}
              {application.config.memory && (
                <div className="text-center">
                  <div className="font-semibold">{application.config.memory} Memory</div>
                </div>
              )}
            </div>
          )}

          {/* Actions */}
          <div className="flex gap-2 pt-2">
            {onStatusChange && (
              <Button
                size="sm"
                variant="outline"
                onClick={handleStatusToggle}
                disabled={isUpdating}
                data-testid={`status-${application.id}`}
              >
                {isUpdating ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : application.status === 'running' ? (
                  <>
                    <Pause className="h-4 w-4 mr-1" />
                    Stop
                  </>
                ) : (
                  <>
                    <Play className="h-4 w-4 mr-1" />
                    Start
                  </>
                )}
              </Button>
            )}
            {onEdit && (
              <Button
                size="sm"
                variant="outline"
                onClick={handleEdit}
                data-testid={`edit-${application.id}`}
              >
                <Edit className="h-4 w-4 mr-1" />
                Edit
              </Button>
            )}
            {onDelete && (
              <Button
                size="sm"
                variant="outline"
                onClick={handleDelete}
                data-testid={`delete-${application.id}`}
              >
                <Trash2 className="h-4 w-4 mr-1" />
                Delete
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}