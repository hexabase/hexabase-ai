'use client';

import { useState } from 'react';
import { Workspace } from '@/lib/api-client';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Download, Trash2, Server, AlertCircle, Clock } from 'lucide-react';
import { cn } from '@/lib/utils';

interface WorkspaceCardProps {
  workspace: Workspace;
  plan?: {
    id: string;
    name: string;
    description: string;
    price: number;
    currency: string;
    resource_limits?: {
      cpu: string;
      memory: string;
      storage: string;
    };
  };
  onClick: (workspaceId: string) => void;
  onDelete?: (workspaceId: string) => void;
  onDownloadKubeconfig?: (workspaceId: string) => void;
}

export function WorkspaceCard({
  workspace,
  plan,
  onClick,
  onDelete,
  onDownloadKubeconfig,
}: WorkspaceCardProps) {
  const [isDeleting, setIsDeleting] = useState(false);
  const [isDownloading, setIsDownloading] = useState(false);

  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case 'active':
        return 'default';
      case 'provisioning':
        return 'secondary';
      case 'error':
        return 'destructive';
      case 'suspended':
        return 'outline';
      default:
        return 'default';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <Server className="h-3 w-3" />;
      case 'provisioning':
        return <Clock className="h-3 w-3" />;
      case 'error':
        return <AlertCircle className="h-3 w-3" />;
      default:
        return null;
    }
  };

  const handleCardClick = (e: React.MouseEvent) => {
    // Don't trigger card click if clicking on action buttons
    if ((e.target as HTMLElement).closest('button')) {
      return;
    }
    onClick(workspace.id);
  };

  const handleDelete = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!onDelete) return;
    
    setIsDeleting(true);
    try {
      await onDelete(workspace.id);
    } finally {
      setIsDeleting(false);
    }
  };

  const handleDownloadKubeconfig = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!onDownloadKubeconfig) return;
    
    setIsDownloading(true);
    try {
      await onDownloadKubeconfig(workspace.id);
    } finally {
      setIsDownloading(false);
    }
  };

  return (
    <Card 
      className="cursor-pointer hover:shadow-lg transition-shadow"
      onClick={handleCardClick}
      data-testid={`workspace-card-${workspace.id}`}
    >
      <CardHeader>
        <div className="flex justify-between items-start">
          <CardTitle className="text-lg">{workspace.name}</CardTitle>
          <Badge 
            variant={getStatusBadgeVariant(workspace.vcluster_status) as any}
            className={cn(
              'flex items-center gap-1',
              `bg-${getStatusBadgeVariant(workspace.vcluster_status)}`
            )}
          >
            {getStatusIcon(workspace.vcluster_status)}
            {workspace.vcluster_status}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {plan && (
            <div className="space-y-2">
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Plan</span>
                <span className="font-medium">{plan.name}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Price</span>
                <span className="font-medium">
                  {plan.price === 0 ? 'Free' : `$${plan.price}/month`}
                </span>
              </div>
              {plan.resource_limits && (
                <div className="pt-2 border-t">
                  <p className="text-sm text-muted-foreground mb-1">Resources</p>
                  <div className="text-sm space-y-1">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">CPU</span>
                      <span>{plan.resource_limits.cpu}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Memory</span>
                      <span>{plan.resource_limits.memory}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">Storage</span>
                      <span>{plan.resource_limits.storage}</span>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}
          
          <div className="flex gap-2 pt-2">
            {onDownloadKubeconfig && (
              <Button
                size="sm"
                variant="outline"
                onClick={handleDownloadKubeconfig}
                disabled={workspace.vcluster_status !== 'active' || isDownloading}
                data-testid={`download-kubeconfig-${workspace.id}`}
              >
                <Download className="h-4 w-4 mr-1" />
                {isDownloading ? 'Downloading...' : 'Kubeconfig'}
              </Button>
            )}
            {onDelete && (
              <Button
                size="sm"
                variant="outline"
                onClick={handleDelete}
                disabled={isDeleting}
                data-testid={`delete-${workspace.id}`}
              >
                <Trash2 className="h-4 w-4 mr-1" />
                {isDeleting ? 'Deleting...' : 'Delete'}
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}