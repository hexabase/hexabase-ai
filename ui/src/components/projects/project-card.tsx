"use client";

import { useState } from "react";
import { Project } from "@/lib/api-client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Edit,
  Trash2,
  Package,
  Server,
  HardDrive,
  Cpu,
  MemoryStick,
  AlertCircle,
  Clock,
  CheckCircle,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface ProjectCardProps {
  project: Project;
  onClick: (projectId: string) => void;
  onEdit?: (project: Project) => void;
  onDelete?: (projectId: string) => void;
}

export function ProjectCard({
  project,
  onClick,
  onEdit,
  onDelete,
}: ProjectCardProps) {
  const [isDeleting, setIsDeleting] = useState(false);

  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case "active":
        return "success";
      case "creating":
        return "warning";
      case "error":
        return "error";
      case "suspended":
        return "secondary";
      default:
        return "outline";
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "active":
        return <CheckCircle className="h-3 w-3" />;
      case "creating":
        return <Clock className="h-3 w-3" />;
      case "error":
        return <AlertCircle className="h-3 w-3" />;
      default:
        return null;
    }
  };

  const handleCardClick = (e: React.MouseEvent) => {
    // Don't trigger card click if clicking on action buttons
    if ((e.target as HTMLElement).closest("button")) {
      return;
    }
    onClick(project.id);
  };

  const handleEdit = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (onEdit) {
      onEdit(project);
    }
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!onDelete) return;

    setIsDeleting(true);
    onDelete(project.id);
    setIsDeleting(false);
  };

  return (
    <Card
      className="cursor-pointer hover:shadow-lg transition-shadow"
      onClick={handleCardClick}
      data-testid={`project-card-${project.id}`}
    >
      <CardHeader>
        <div className="flex justify-between items-start">
          <div>
            <CardTitle className="text-lg">{project.name}</CardTitle>
            <p className="text-sm text-muted-foreground mt-1">
              {project.namespace}
            </p>
          </div>
          <Badge
            variant={getStatusBadgeVariant(project.status)}
            className={cn("flex items-center gap-1")}
          >
            {getStatusIcon(project.status)}
            {project.status}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            {project.description || "No description"}
          </p>

          {project.resources && (
            <div className="grid grid-cols-3 gap-2 text-sm">
              <div className="text-center">
                <Package className="h-4 w-4 mx-auto mb-1 text-muted-foreground" />
                <div className="font-semibold">
                  {project.resources.deployments}
                </div>
                <div className="text-xs text-muted-foreground">Deployments</div>
              </div>
              <div className="text-center">
                <Server className="h-4 w-4 mx-auto mb-1 text-muted-foreground" />
                <div className="font-semibold">
                  {project.resources.services}
                </div>
                <div className="text-xs text-muted-foreground">Services</div>
              </div>
              <div className="text-center">
                <HardDrive className="h-4 w-4 mx-auto mb-1 text-muted-foreground" />
                <div className="font-semibold">{project.resources.pods}</div>
                <div className="text-xs text-muted-foreground">Pods</div>
              </div>
            </div>
          )}

          {project.resource_quota && (
            <div className="pt-3 border-t">
              <p className="text-xs font-medium text-muted-foreground mb-2">
                Resource Quotas
              </p>
              <div className="flex flex-wrap gap-2 text-xs">
                <div className="flex items-center gap-1">
                  <Cpu className="h-3 w-3" />
                  <span>{project.resource_quota.cpu} CPU</span>
                </div>
                <div className="flex items-center gap-1">
                  <MemoryStick className="h-3 w-3" />
                  <span>{project.resource_quota.memory} Memory</span>
                </div>
                <div className="flex items-center gap-1">
                  <HardDrive className="h-3 w-3" />
                  <span>{project.resource_quota.storage} Storage</span>
                </div>
              </div>
            </div>
          )}

          <div className="flex gap-2 pt-2">
            {onEdit && (
              <Button
                size="sm"
                variant="outline"
                onClick={handleEdit}
                data-testid={`edit-${project.id}`}
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
                disabled={isDeleting}
                data-testid={`delete-${project.id}`}
              >
                <Trash2 className="h-4 w-4 mr-1" />
                {isDeleting ? "Deleting..." : "Delete"}
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
