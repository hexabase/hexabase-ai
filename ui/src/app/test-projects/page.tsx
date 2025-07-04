'use client';

import { ProjectList } from '@/components/projects/project-list';

// This is a test page for projects functionality
export default function TestProjectsPage() {
  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Projects Management Test</h1>
          <p className="text-gray-600 mt-2">Testing projects functionality</p>
        </div>
        
        <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-6">Projects</h2>
          <ProjectList organizationId="test-org-123" />
        </div>
      </div>
    </div>
  );
}