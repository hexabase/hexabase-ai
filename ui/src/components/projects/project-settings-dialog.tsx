'use client';

import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Settings, Users, Cpu, Package } from 'lucide-react';
import { type Project } from '@/lib/api-client';

interface ProjectSettingsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  project: Project;
  onUpdate: (project: Project) => void;
}

export function ProjectSettingsDialog({ open, onOpenChange, project }: ProjectSettingsDialogProps) {
  const [activeTab, setActiveTab] = useState<'general' | 'resources' | 'permissions'>('general');

  const handleClose = () => {
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[700px]" data-testid="project-settings-modal">
        <DialogHeader>
          <DialogTitle>Project Settings</DialogTitle>
          <DialogDescription>
            Configure settings and permissions for {project.name}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Settings Tabs */}
          <div className="border-b border-gray-200">
            <nav className="flex space-x-8" aria-label="Tabs">
              <button
                className={`whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm ${
                  activeTab === 'general'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
                onClick={() => setActiveTab('general')}
                data-testid="general-settings-tab"
              >
                <Settings className="h-4 w-4 mr-2 inline" />
                General
              </button>
              <button
                className={`whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm ${
                  activeTab === 'resources'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
                onClick={() => setActiveTab('resources')}
                data-testid="resource-limits-tab"
              >
                <Cpu className="h-4 w-4 mr-2 inline" />
                Resource Limits
              </button>
              <button
                className={`whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm ${
                  activeTab === 'permissions'
                    ? 'border-primary-500 text-primary-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700'
                }`}
                onClick={() => setActiveTab('permissions')}
                data-testid="permissions-tab"
              >
                <Users className="h-4 w-4 mr-2 inline" />
                Permissions
              </button>
            </nav>
          </div>

          {/* Tab Content */}
          <div className="space-y-6">
            {activeTab === 'general' && (
              <div className="space-y-4">
                <Card>
                  <CardHeader>
                    <CardTitle>Project Information</CardTitle>
                    <CardDescription>
                      Basic information about this project
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div>
                      <label className="text-sm font-medium text-gray-700">Project Name</label>
                      <p className="text-sm text-gray-900 mt-1">{project.name}</p>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-gray-700">Description</label>
                      <p className="text-sm text-gray-900 mt-1">
                        {project.description || 'No description provided'}
                      </p>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-gray-700">Status</label>
                      <div className="mt-1">
                        <Badge variant="default">{project.status}</Badge>
                      </div>
                    </div>
                    <div>
                      <label className="text-sm font-medium text-gray-700">Workspace</label>
                      <p className="text-sm text-gray-900 mt-1">{project.workspace_name}</p>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>Danger Zone</CardTitle>
                    <CardDescription>
                      Irreversible and destructive actions
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="border border-red-200 rounded-lg p-4 bg-red-50">
                      <h4 className="text-sm font-medium text-red-900 mb-2">Archive Project</h4>
                      <p className="text-sm text-red-700 mb-4">
                        Archive this project and all its namespaces. This action can be undone later.
                      </p>
                      <Button variant="outline" className="text-red-600 border-red-300 hover:bg-red-50">
                        Archive Project
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              </div>
            )}

            {activeTab === 'resources' && (
              <div className="space-y-4">
                <Card>
                  <CardHeader>
                    <CardTitle>Resource Limits</CardTitle>
                    <CardDescription>
                      Overall resource limits for this project across all namespaces
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="text-center py-8 text-gray-500">
                      <Cpu className="mx-auto h-12 w-12 mb-4" />
                      <p>Resource limit configuration coming soon</p>
                    </div>
                  </CardContent>
                </Card>
              </div>
            )}

            {activeTab === 'permissions' && (
              <div className="space-y-4">
                <Card>
                  <CardHeader>
                    <CardTitle>Access Control</CardTitle>
                    <CardDescription>
                      Manage who can access and modify this project
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="text-center py-8 text-gray-500">
                      <Users className="mx-auto h-12 w-12 mb-4" />
                      <p>Permission management coming soon</p>
                    </div>
                  </CardContent>
                </Card>
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="flex justify-end space-x-3 pt-4 border-t border-gray-200">
            <Button variant="outline" onClick={handleClose}>
              Close
            </Button>
            <Button onClick={handleClose}>
              Save Changes
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}