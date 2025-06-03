'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { MoreHorizontal, Package, Cpu, Database, Users, Edit, Trash2, AlertTriangle } from 'lucide-react';
import { type Namespace } from '@/lib/api-client';
import { useToast } from '@/hooks/use-toast';

interface NamespaceCardProps {
  namespace: Namespace;
  onUpdate: () => void;
}

export function NamespaceCard({ namespace, onUpdate }: NamespaceCardProps) {
  const { toast } = useToast();
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'bg-green-100 text-green-800';
      case 'inactive':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const calculateUsagePercentage = (usage: string, quota: string) => {
    // Simple calculation for display purposes
    if (usage.includes('cores') && quota.includes('cores')) {
      const usageNum = parseFloat(usage);
      const quotaNum = parseFloat(quota);
      return Math.round((usageNum / quotaNum) * 100);
    }
    if (usage.includes('GB') && quota.includes('GB')) {
      const usageNum = parseFloat(usage);
      const quotaNum = parseFloat(quota);
      return Math.round((usageNum / quotaNum) * 100);
    }
    if (typeof usage === 'number' && typeof quota === 'number') {
      return Math.round((usage / quota) * 100);
    }
    return 0;
  };

  const handleEdit = () => {
    setIsMenuOpen(false);
    toast({
      title: 'Edit Namespace',
      description: 'Edit namespace functionality coming soon',
    });
  };

  const handleDelete = () => {
    setIsMenuOpen(false);
    toast({
      title: 'Delete Namespace',
      description: 'Are you sure you want to delete this namespace?',
    });
  };

  const cpuPercentage = calculateUsagePercentage(namespace.resource_usage.cpu, namespace.resource_quota.cpu);
  const memoryPercentage = calculateUsagePercentage(namespace.resource_usage.memory, namespace.resource_quota.memory);
  const podPercentage = Math.round((namespace.resource_usage.pods / namespace.resource_quota.pods) * 100);

  return (
    <Card 
      className="hover:shadow-md transition-shadow"
      data-testid="namespace-card"
      data-namespace-id={`namespace-card-${namespace.name}`}
    >
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className="h-8 w-8 bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg flex items-center justify-center">
              <Package className="h-4 w-4 text-white" />
            </div>
            <div>
              <CardTitle className="text-base font-semibold">
                {namespace.name}
              </CardTitle>
              <Badge 
                className={getStatusColor(namespace.status)}
                data-testid="namespace-status"
              >
                {namespace.status}
              </Badge>
            </div>
          </div>
          
          <DropdownMenu open={isMenuOpen} onOpenChange={setIsMenuOpen}>
            <DropdownMenuTrigger asChild>
              <Button 
                variant="ghost" 
                size="sm" 
                className="h-8 w-8 p-0"
                data-testid="namespace-actions-menu"
              >
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={handleEdit} data-testid="edit-namespace">
                <Edit className="h-4 w-4 mr-2" />
                Edit Namespace
              </DropdownMenuItem>
              <DropdownMenuItem 
                onClick={handleDelete} 
                className="text-red-600"
                data-testid="delete-namespace"
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Delete Namespace
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* Description */}
        {namespace.description && (
          <p className="text-sm text-gray-600">{namespace.description}</p>
        )}

        {/* Resource Usage */}
        <div className="space-y-3">
          {/* CPU Usage */}
          <div className="space-y-1" data-testid="namespace-cpu-usage">
            <div className="flex items-center justify-between text-sm">
              <div className="flex items-center">
                <Cpu className="h-4 w-4 text-gray-400 mr-2" />
                <span>CPU</span>
              </div>
              <span className="font-medium">
                {namespace.resource_usage.cpu} / {namespace.resource_quota.cpu}
              </span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div 
                className={`h-2 rounded-full transition-all ${
                  cpuPercentage > 80 ? 'bg-red-500' : cpuPercentage > 60 ? 'bg-yellow-500' : 'bg-green-500'
                }`}
                style={{ width: `${Math.min(cpuPercentage, 100)}%` }}
              ></div>
            </div>
            <div className="text-xs text-gray-500">{cpuPercentage}% used</div>
          </div>

          {/* Memory Usage */}
          <div className="space-y-1" data-testid="namespace-memory-usage">
            <div className="flex items-center justify-between text-sm">
              <div className="flex items-center">
                <Database className="h-4 w-4 text-gray-400 mr-2" />
                <span>Memory</span>
              </div>
              <span className="font-medium">
                {namespace.resource_usage.memory} / {namespace.resource_quota.memory}
              </span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div 
                className={`h-2 rounded-full transition-all ${
                  memoryPercentage > 80 ? 'bg-red-500' : memoryPercentage > 60 ? 'bg-yellow-500' : 'bg-green-500'
                }`}
                style={{ width: `${Math.min(memoryPercentage, 100)}%` }}
              ></div>
            </div>
            <div className="text-xs text-gray-500">{memoryPercentage}% used</div>
          </div>

          {/* Pod Count */}
          <div className="flex items-center justify-between text-sm" data-testid="namespace-pod-count">
            <div className="flex items-center">
              <Users className="h-4 w-4 text-gray-400 mr-2" />
              <span>Pods</span>
            </div>
            <span className="font-medium">
              {namespace.resource_usage.pods} / {namespace.resource_quota.pods}
            </span>
          </div>

          {/* Warnings */}
          {(cpuPercentage > 80 || memoryPercentage > 80) && (
            <div className="flex items-center space-x-2 p-2 bg-yellow-50 rounded-lg border border-yellow-200">
              <AlertTriangle className="h-4 w-4 text-yellow-600" />
              <span className="text-xs text-yellow-800">
                High resource usage detected
              </span>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}