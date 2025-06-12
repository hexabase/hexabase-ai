'use client';

import { ProtectedRoute } from '@/components/protected-route';
import { ReactNode } from 'react';

export default function DashboardLayout({ children }: { children: ReactNode }) {
  return (
    <ProtectedRoute requireOrganization>
      <div className="min-h-screen bg-gray-50">
        {/* Navigation header will go here */}
        <main>{children}</main>
      </div>
    </ProtectedRoute>
  );
}