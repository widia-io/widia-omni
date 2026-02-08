import { Link } from "react-router";
import { Button } from "@/components/ui/button";
import { Check } from "lucide-react";
import { cn } from "@/lib/cn";

interface PricingCardProps {
  name: string;
  price: string;
  period?: string;
  features: string[];
  cta: string;
  highlighted?: boolean;
}

export function PricingCard({ name, price, period, features, cta, highlighted }: PricingCardProps) {
  return (
    <div className={cn(
      "flex flex-col rounded-[14px] border p-6",
      highlighted
        ? "border-accent-orange bg-gradient-to-b from-accent-orange/5 to-transparent"
        : "border-border bg-bg-card",
    )}>
      {highlighted && (
        <span className="mb-4 w-fit rounded-full bg-accent-orange/10 px-3 py-0.5 text-xs font-medium text-accent-orange">
          Popular
        </span>
      )}
      <h3 className="text-lg font-bold">{name}</h3>
      <div className="mt-2 flex items-baseline gap-1">
        <span className="text-3xl font-bold">{price}</span>
        {period && <span className="text-sm text-text-muted">{period}</span>}
      </div>
      <ul className="mt-6 flex-1 space-y-3">
        {features.map((f) => (
          <li key={f} className="flex items-start gap-2 text-sm text-text-secondary">
            <Check className="mt-0.5 h-4 w-4 shrink-0 text-accent-green" />
            {f}
          </li>
        ))}
      </ul>
      <Button variant={highlighted ? "default" : "outline"} className="mt-6 w-full" asChild>
        <Link to="/register">{cta}</Link>
      </Button>
    </div>
  );
}
