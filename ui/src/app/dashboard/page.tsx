'use client';

import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { useAuth } from '@/lib/auth-context';

export default function DashboardPage() {
  const router = useRouter();
  const { activeOrganization } = useAuth();

  useEffect(() => {
    if (activeOrganization) {
      router.replace(`/dashboard/organizations/${activeOrganization.id}`);
    } else {
      router.replace('/dashboard/organizations');
    }
  }, [activeOrganization, router]);

  return (
    <div className="flex h-screen items-center justify-center">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
    </div>
  );
}