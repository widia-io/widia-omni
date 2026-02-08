import { type HTMLAttributes } from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";

const badgeVariants = cva(
  "inline-flex items-center rounded-[4px] px-2 py-0.5 font-mono text-[11px] font-medium",
  {
    variants: {
      variant: {
        default: "bg-bg-elevated text-text-secondary",
        orange: "bg-accent-orange-soft text-accent-orange",
        blue: "bg-accent-blue-soft text-accent-blue",
        green: "bg-accent-green-soft text-accent-green",
        sand: "bg-accent-sand-soft text-accent-sand",
        rose: "bg-accent-rose-soft text-accent-rose",
        sage: "bg-accent-sage-soft text-accent-sage",
      },
    },
    defaultVariants: { variant: "default" },
  },
);

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement>, VariantProps<typeof badgeVariants> {}

export function Badge({ className, variant, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ variant, className }))} {...props} />;
}
