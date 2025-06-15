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
          "flex h-10 w-full rounded px-3 py-2 text-sm transition-all duration-200",
          // Default state
          "border border-[color:hsl(var(--border))] bg-[color:hsl(var(--background))] text-[color:hsl(var(--foreground))] placeholder:text-[color:hsl(var(--muted-foreground))]",
          // Hover state
          "hover:bg-[color:hsl(var(--muted)/0.1)]",
          // Focus state
          "focus-visible:outline-none focus-visible:border-hexa-green focus-visible:ring-2 focus-visible:ring-hexa-green focus-visible:ring-offset-2 focus-visible:ring-offset-[color:hsl(var(--background))]",
          // Disabled state
          "disabled:cursor-not-allowed disabled:opacity-50",
          // Error state
          error && "border-[#FF7979] hover:border-[#FF7979] focus-visible:border-[#FF7979] focus-visible:ring-[#FF7979]",
          // File input specific styles
          "file:border-0 file:bg-transparent file:text-sm file:font-medium",
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