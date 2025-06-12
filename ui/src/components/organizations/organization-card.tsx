'use client';

import { Organization } from '@/lib/api-client';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Edit2, Trash2, Users, Layers } from 'lucide-react';
import { cn } from '@/lib/utils';

interface OrganizationCardProps {
  organization: Organization & { 
    member_count?: number; 
    workspace_count?: number;
  };
  isActive: boolean;
  onClick: (orgId: string) => void;
  onEdit?: (org: Organization) => void;
  onDelete?: (orgId: string) => void;
}

export function OrganizationCard({
  organization,
  isActive,
  onClick,
  onEdit,
  onDelete,
}: OrganizationCardProps) {
  const handleClick = () => {
    onClick(organization.id);
  };

  const handleEdit = (e: React.MouseEvent) => {
    e.stopPropagation();
    onEdit?.(organization);
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDelete?.(organization.id);
  };

  return (
    <Card
      data-testid={`org-card-${organization.id}`}
      className={cn(
        'cursor-pointer transition-all hover:shadow-md',
        isActive && 'ring-2 ring-primary'
      )}
      onClick={handleClick}
    >
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <div className="flex items-center gap-2">
          <h3 className="text-lg font-semibold">{organization.name}</h3>
          {isActive && (
            <Badge variant="secondary" className="text-xs">
              Active
            </Badge>
          )}
        </div>
        <Badge variant={organization.role === 'admin' ? 'default' : 'secondary'}>
          {organization.role}
        </Badge>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            {organization.member_count !== undefined && (
              <div className="flex items-center gap-1">
                <Users className="h-4 w-4" />
                <span>{organization.member_count} members</span>
              </div>
            )}
            {organization.workspace_count !== undefined && (
              <div className="flex items-center gap-1">
                <Layers className="h-4 w-4" />
                <span>{organization.workspace_count} workspaces</span>
              </div>
            )}
          </div>
          {organization.role === 'admin' && (onEdit || onDelete) && (
            <div className="flex items-center gap-2">
              {onEdit && (
                <Button
                  variant="ghost"
                  size="icon"
                  data-testid={`edit-${organization.id}`}
                  onClick={handleEdit}
                >
                  <Edit2 className="h-4 w-4" />
                </Button>
              )}
              {onDelete && (
                <Button
                  variant="ghost"
                  size="icon"
                  data-testid={`delete-${organization.id}`}
                  onClick={handleDelete}
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              )}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}