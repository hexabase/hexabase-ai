'use client';

import { OrganizationList } from '@/components/organizations/organization-list';

export default function OrganizationsPage() {
  return (
    <div className="container mx-auto py-8 px-4">
      <OrganizationList />
    </div>
  );
}