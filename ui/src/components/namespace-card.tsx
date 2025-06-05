'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Cpu, Database, Package, MoreVertical, Edit, Trash2 } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { namespacesApi, type Namespace } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';

interface NamespaceCardProps {
  namespace: Namespace;
  organizationId?: string;
  projectId?: string;
  onUpdate?: () => void;
}

export function NamespaceCard({ namespace, organizationId, projectId, onUpdate }: NamespaceCardProps) {
  const { toast } = useToast();
  const [deleting, setDeleting] = useState(false);

  const handleDelete = async () => {
    if (!organizationId || !projectId) return;
    
    if (!confirm(`Are you sure you want to delete the namespace "${namespace.name}"? This action cannot be undone.`)) {
      return;
    }

    try {
      setDeleting(true);
      await namespacesApi.delete(organizationId, projectId, namespace.id);
      toast({
        title: 'Success',
        description: 'Namespace deleted successfully',
      });
      onUpdate?.();
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete namespace',
        variant: 'destructive',
      });
    } finally {
      setDeleting(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      'active': { label: 'Active', variant: 'default' as const },
      'inactive': { label: 'Inactive', variant: 'secondary' as const },
    };

    const config = statusConfig[status] || { label: status, variant: 'outline' as const };

    return <Badge variant={config.variant}>{config.label}</Badge>;
  };

  const calculateUsagePercentage = (used: string, limit: string): number => {
    const usedNum = parseFloat(used);
    const limitNum = parseFloat(limit);
    if (limitNum === 0) return 0;
    return Math.round((usedNum / limitNum) * 100);
  };

  const cpuUsage = calculateUsagePercentage(
    namespace.resource_usage.cpu,
    namespace.resource_quota.cpu
  );

  const memoryUsage = calculateUsagePercentage(
    namespace.resource_usage.memory,
    namespace.resource_quota.memory
  );

  const podUsage = Math.round(
    (namespace.resource_usage.pods / namespace.resource_quota.pods) * 100
  );

  return (
    <Card data-testid={`namespace-card-${namespace.id}`}>
      <CardHeader>
        <div className="flex justify-between items-start">
          <div className="flex-1">
            <CardTitle className="text-lg flex items-center gap-2">
              <Package className="h-4 w-4" />
              {namespace.name}
            </CardTitle>
            <CardDescription className="mt-1">
              {namespace.description || 'No description'}
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            {getStatusBadge(namespace.status)}
            {organizationId && projectId && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button 
                    variant="ghost" 
                    size="icon"
                    disabled={deleting}
                    data-testid="namespace-actions-menu"
                  >
                    <MoreVertical className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuLabel>Actions</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem data-testid="edit-namespace">
                    <Edit className="w-4 h-4 mr-2" />
                    Edit Namespace
                  </DropdownMenuItem>
                  <DropdownMenuItem 
                    className="text-red-600"
                    onClick={handleDelete}
                    data-testid="delete-namespace"
                  >
                    <Trash2 className="w-4 h-4 mr-2" />
                    Delete Namespace
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* CPU Usage */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <div className="flex items-center gap-2">
              <Cpu className="h-4 w-4 text-gray-500" />
              <span>CPU</span>
            </div>
            <span className="text-gray-600" data-testid="namespace-cpu-usage">
              {namespace.resource_usage.cpu} / {namespace.resource_quota.cpu}
            </span>
          </div>
          <Progress value={cpuUsage} className="h-2" />
        </div>

        {/* Memory Usage */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <div className="flex items-center gap-2">
              <Database className="h-4 w-4 text-gray-500" />
              <span>Memory</span>
            </div>
            <span className="text-gray-600" data-testid="namespace-memory-usage">
              {namespace.resource_usage.memory} / {namespace.resource_quota.memory}
            </span>
          </div>
          <Progress value={memoryUsage} className="h-2" />
        </div>

        {/* Pod Count */}
        <div className="flex items-center justify-between text-sm">
          <span className="text-gray-600">Pods</span>
          <span className="font-medium" data-testid="namespace-pod-count">
            {namespace.resource_usage.pods} / {namespace.resource_quota.pods}
          </span>
        </div>

        <div className="pt-2 text-xs text-gray-500" data-testid="namespace-status">
          Created {new Date(namespace.created_at).toLocaleDateString()}
        </div>
      </CardContent>
    </Card>
  );
}