'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { MoreHorizontal, Package, Users, Cpu, Database, Calendar } from 'lucide-react';
import { type Project } from '@/lib/api-client';
import { formatDateTime } from '@/lib/utils';

interface ProjectCardProps {
  project: Project;
  onClick: () => void;
}

export function ProjectCard({ project, onClick }: ProjectCardProps) {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'bg-green-100 text-green-800';
      case 'inactive':
        return 'bg-yellow-100 text-yellow-800';
      case 'archived':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <Card 
      className="hover:shadow-lg transition-shadow cursor-pointer group"
      onClick={onClick}
      data-testid="project-card"
      data-project-id={`project-card-${project.name.toLowerCase().replace(/\s+/g, '-')}`}
      data-status={project.status}
    >
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className="h-10 w-10 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center">
              <Package className="h-5 w-5 text-white" />
            </div>
            <div className="flex-1">
              <CardTitle className="text-lg font-semibold" data-testid="project-name">
                {project.name}
              </CardTitle>
              <div className="flex items-center space-x-2 mt-1">
                <Badge 
                  className={getStatusColor(project.status)}
                  data-testid="project-status"
                >
                  {project.status}
                </Badge>
              </div>
            </div>
          </div>
          
          <DropdownMenu>
            <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
              <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem>
                <Package className="h-4 w-4 mr-2" />
                View Details
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Users className="h-4 w-4 mr-2" />
                Manage Access
              </DropdownMenuItem>
              <DropdownMenuItem className="text-red-600">
                Archive Project
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* Description */}
        {project.description && (
          <CardDescription className="text-sm" data-testid="project-description">
            {project.description}
          </CardDescription>
        )}

        {/* Workspace Info */}
        <div className="flex items-center text-sm text-gray-600" data-testid="project-workspace">
          <div className="h-2 w-2 bg-blue-500 rounded-full mr-2"></div>
          <span>{project.workspace_name || 'Unknown Workspace'}</span>
        </div>

        {/* Resource Usage */}
        {project.resource_usage && (
          <div className="grid grid-cols-2 gap-4 pt-4 border-t border-gray-100">
            <div className="text-center">
              <div className="flex items-center justify-center mb-1">
                <Cpu className="h-4 w-4 text-gray-400 mr-1" />
                <span className="text-xs text-gray-500">CPU</span>
              </div>
              <div className="text-sm font-medium">{project.resource_usage.cpu}</div>
            </div>
            <div className="text-center">
              <div className="flex items-center justify-center mb-1">
                <Database className="h-4 w-4 text-gray-400 mr-1" />
                <span className="text-xs text-gray-500">Memory</span>
              </div>
              <div className="text-sm font-medium">{project.resource_usage.memory}</div>
            </div>
          </div>
        )}

        {/* Namespace Count */}
        <div className="flex items-center justify-between pt-4 border-t border-gray-100">
          <div className="flex items-center text-sm text-gray-600">
            <Package className="h-4 w-4 mr-2" />
            <span data-testid="project-namespace-count">
              {project.namespace_count} namespace{project.namespace_count !== 1 ? 's' : ''}
            </span>
          </div>
          
          {project.resource_usage && (
            <div className="flex items-center text-sm text-gray-600">
              <Users className="h-4 w-4 mr-2" />
              <span>{project.resource_usage.pods} pods</span>
            </div>
          )}
        </div>

        {/* Created Date */}
        <div className="flex items-center text-xs text-gray-500 pt-2">
          <Calendar className="h-3 w-3 mr-1" />
          <span>Created {formatDateTime(project.created_at)}</span>
        </div>
      </CardContent>
    </Card>
  );
}