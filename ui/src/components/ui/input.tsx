import * as React from "react"
import { cn } from "@/lib/utils"

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: boolean
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, error, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          // Base styles
          "flex h-10 w-full rounded px-st py-md text-sm transition-all duration-200",
          // Default state
          "border border-border bg-input text-text-primary placeholder:text-text-placeholder",
          // Hover state
          "hover:bg-input-hover hover:border-border-hover",
          // Focus state
          "focus-visible:outline-none focus-visible:border-hexa-green focus-visible:ring-2 focus-visible:ring-hexa-green focus-visible:ring-offset-2 focus-visible:ring-offset-background",
          // Disabled state
          "disabled:cursor-not-allowed disabled:opacity-50 disabled:bg-input-disabled disabled:border-border-disabled disabled:hover:bg-input-disabled disabled:hover:border-border-disabled",
          // Error state
          error && "border-error hover:border-error focus-visible:border-error focus-visible:ring-error",
          // File input specific styles
          "file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-text-primary",
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
Input.displayName = "Input"

export { Input }