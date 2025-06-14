import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded text-sm font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-hexa-green focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-50 disabled:cursor-not-allowed",
  {
    variants: {
      variant: {
        // Primary - Hexa Green (main CTA)
        default: "bg-hexa-green text-white hover:bg-hexa-green-500 active:bg-hexa-green-700 shadow-dp-2",
        primary: "bg-hexa-green text-white hover:bg-hexa-green-500 active:bg-hexa-green-700 shadow-dp-2",
        
        // Secondary - Hexa Pink
        secondary: "bg-hexa-pink text-white hover:bg-hexa-pink-500 active:bg-hexa-pink-700 shadow-dp-2",
        
        // Cancel - Gray
        cancel: "bg-gray-cancel text-white hover:bg-gray-cancel-hover shadow-dp-2",
        
        // Destructive/Delete - Error colors
        destructive: "bg-error text-white hover:bg-error-hover shadow-dp-2",
        delete: "bg-error text-white hover:bg-error-hover shadow-dp-2",
        
        // Outline variants
        outline: "border-2 border-hexa-green text-hexa-green hover:bg-hexa-green hover:text-white",
        "outline-secondary": "border-2 border-hexa-pink text-hexa-pink hover:bg-hexa-pink hover:text-white",
        
        // Ghost - minimal style
        ghost: "text-text-primary hover:bg-muted hover:text-text-primary",
        
        // Link style
        link: "text-hexa-green underline-offset-4 hover:underline hover:text-hexa-green-500",
        
        // Disabled state (handled by disabled prop, but included for completeness)
        disabled: "bg-gray-disabled text-gray-500 cursor-not-allowed",
      },
      size: {
        default: "h-10 px-st py-md", // 16px padding, 8px vertical
        sm: "h-8 px-sm py-xs text-xs", // Small
        lg: "h-12 px-lg py-md text-base", // Large
        xl: "h-14 px-xxl py-st text-base", // Extra large
        icon: "h-10 w-10",
        "icon-sm": "h-8 w-8",
        "icon-lg": "h-12 w-12",
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