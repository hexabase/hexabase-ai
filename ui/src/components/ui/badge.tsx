import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex items-center rounded px-sm py-xs text-xs font-medium transition-colors",
  {
    variants: {
      variant: {
        // High importance variants
        default: "bg-hexa-green text-white",
        primary: "bg-hexa-green text-white",
        secondary: "bg-hexa-pink text-white",
        gray: "bg-gray text-white",
        // Low importance
        subtle: "bg-black text-white",
        // Status variants
        success: "bg-hexa-green text-white",
        warning: "bg-hexa-pink text-white",
        error: "bg-error text-white",
        info: "bg-primary text-white",
        // Outline variants
        outline: "border border-border text-text-primary bg-transparent",
        "outline-primary": "border border-hexa-green text-hexa-green bg-transparent",
        "outline-secondary": "border border-hexa-pink text-hexa-pink bg-transparent",
      },
      size: {
        default: "px-sm py-xs text-xs",
        sm: "px-xs py-0.5 text-[10px]",
        lg: "px-md py-sm text-sm",
      }
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  )
}

export { Badge, badgeVariants }