import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded text-sm font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-hexa-green focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-50 disabled:cursor-not-allowed",
  {
    variants: {
      variant: {
        // Primary - Hexa Green (main CTA)
        default: "bg-hexa-green text-white hover:bg-hexa-green-500 active:bg-hexa-green-700 shadow-md",
        primary: "bg-hexa-green text-white hover:bg-hexa-green-500 active:bg-hexa-green-700 shadow-md",
        
        // Secondary - Hexa Pink
        secondary: "bg-hexa-pink text-white hover:bg-hexa-pink-500 active:bg-hexa-pink-700 shadow-md",
        
        // Cancel - Gray
        cancel: "bg-[#9e9e9e] text-white hover:bg-[#b5b5b5] shadow-md",
        
        // Destructive/Delete - Error colors
        destructive: "bg-[#FF7979] text-white hover:bg-[#FF9B9B] shadow-md",
        delete: "bg-[#FF7979] text-white hover:bg-[#FF9B9B] shadow-md",
        
        // Outline variants
        outline: "border-2 border-hexa-green text-hexa-green hover:bg-hexa-green hover:text-white",
        "outline-secondary": "border-2 border-hexa-pink text-hexa-pink hover:bg-hexa-pink hover:text-white",
        
        // Ghost - minimal style
        ghost: "text-[color:hsl(var(--foreground))] hover:bg-[color:hsl(var(--muted))] hover:text-[color:hsl(var(--foreground))]",
        
        // Link style
        link: "text-hexa-green underline-offset-4 hover:underline hover:text-hexa-green-500",
        
        // Disabled state (handled by disabled prop, but included for completeness)
        disabled: "bg-[#e6e6e6] text-gray-500 cursor-not-allowed",
      },
      size: {
        default: "h-10 px-4 py-2", // 16px padding, 8px vertical
        sm: "h-8 px-3 py-1 text-xs", // Small
        lg: "h-12 px-6 py-2 text-base", // Large
        xl: "h-14 px-8 py-4 text-base", // Extra large
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