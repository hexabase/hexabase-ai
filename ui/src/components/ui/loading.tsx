import { cn } from "@/lib/utils";

interface LoadingSpinnerProps {
  size?: "sm" | "md" | "lg";
  className?: string;
}

export function LoadingSpinner({ size = "md", className }: LoadingSpinnerProps) {
  const sizeClasses = {
    sm: "h-4 w-4 border-2",
    md: "h-6 w-6 border-2", 
    lg: "h-10 w-10 border-3",
  };

  return (
    <div
      className={cn(
        "animate-spin rounded-full border-gray-200 border-t-primary-600",
        sizeClasses[size],
        className
      )}
    />
  );
}

interface LoadingPageProps {
  message?: string;
}

export function LoadingPage({ message = "Loading..." }: LoadingPageProps) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="text-center">
        <div className="relative">
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="h-32 w-32 bg-primary-200 rounded-full animate-pulse"></div>
          </div>
          <div className="relative">
            <LoadingSpinner size="lg" className="mx-auto mb-4" />
          </div>
        </div>
        <p className="text-lg font-medium text-gray-700 mt-8">{message}</p>
      </div>
    </div>
  );
}