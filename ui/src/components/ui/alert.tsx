import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const alertVariants = cva(
  "relative w-full rounded-lg border p-st shadow-dp-2 [&>svg]:absolute [&>svg]:left-st [&>svg]:top-st [&>svg]:text-text-primary [&>svg~*]:pl-plus",
  {
    variants: {
      variant: {
        default: "bg-background-sidebar border-border text-text-primary",
        success: "bg-hexa-green-900/10 border-hexa-green-600 text-hexa-green-200 [&>svg]:text-hexa-green",
        destructive:
          "bg-error/10 border-error text-error [&>svg]:text-error",
        warning: "bg-hexa-pink-900/10 border-hexa-pink-600 text-hexa-pink-200 [&>svg]:text-hexa-pink",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

const Alert = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & VariantProps<typeof alertVariants>
>(({ className, variant, ...props }, ref) => (
  <div
    ref={ref}
    role="alert"
    className={cn(alertVariants({ variant }), className)}
    {...props}
  />
))
Alert.displayName = "Alert"

const AlertTitle = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLHeadingElement>
>(({ className, ...props }, ref) => (
  <h5
    ref={ref}
    className={cn("mb-xs font-semibold leading-none tracking-tight text-heading-base", className)}
    {...props}
  />
))
AlertTitle.displayName = "AlertTitle"

const AlertDescription = React.forwardRef<
  HTMLParagraphElement,
  React.HTMLAttributes<HTMLParagraphElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("text-body-sm [&_p]:leading-relaxed", className)}
    {...props}
  />
))
AlertDescription.displayName = "AlertDescription"

export { Alert, AlertTitle, AlertDescription }