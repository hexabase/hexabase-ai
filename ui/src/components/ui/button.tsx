import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-lg text-sm font-medium ring-offset-background transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 disabled:cursor-not-allowed",
  {
    variants: {
      variant: {
        default: "bg-primary-600 text-white hover:bg-primary-700 hover:shadow-md active:bg-primary-800",
        destructive: "bg-danger-500 text-white hover:bg-danger-600 hover:shadow-md active:bg-danger-700",
        outline: "border border-gray-300 bg-white hover:bg-gray-50 hover:border-gray-400 text-gray-700",
        secondary: "bg-gray-100 text-gray-700 hover:bg-gray-200 hover:text-gray-900",
        ghost: "hover:bg-gray-100 hover:text-gray-900 text-gray-600",
        link: "text-primary-600 underline-offset-4 hover:underline hover:text-primary-700",
        success: "bg-success-600 text-white hover:bg-success-700 hover:shadow-md active:bg-success-800",
        warning: "bg-warning-500 text-white hover:bg-warning-600 hover:shadow-md active:bg-warning-700",
      },
      size: {
        default: "h-10 px-4 py-2",
        sm: "h-8 rounded-md px-3 text-xs",
        lg: "h-12 rounded-lg px-6 text-base",
        xl: "h-14 rounded-xl px-8 text-base",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => {
    return (
      <button
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    );
  }
);
Button.displayName = "Button";

export { Button, buttonVariants };