"use client";

import { useAuth } from "@/lib/auth-context";
import { LoadingPage } from "@/components/ui/loading";
import LoginPage from "@/components/login-page";
import DashboardPage from "@/components/dashboard-page";

export default function Home() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <LoadingPage message="Loading Hexabase KaaS..." />;
  }

  if (!isAuthenticated) {
    return <LoginPage />;
  }

  return <DashboardPage />;
}
