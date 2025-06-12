'use client';

import { WorkspaceList } from '@/components/workspaces/workspace-list';

export default function WorkspacesPage() {
  return (
    <div className="container mx-auto py-8 px-4">
      <WorkspaceList />
    </div>
  );
}