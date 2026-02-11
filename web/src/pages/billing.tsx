import { Check, CreditCard, ExternalLink } from "lucide-react";
import { usePlans, useSubscription, useCheckout, usePortal } from "@/hooks/use-billing";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { formatCurrency } from "@/lib/format";
import { cn } from "@/lib/cn";
import type { SubscriptionStatus } from "@/types/api";

const statusLabel: Record<SubscriptionStatus, string> = {
  trialing: "Em teste",
  active: "Ativo",
  past_due: "Pagamento pendente",
  canceled: "Cancelado",
  paused: "Pausado",
  unpaid: "Não pago",
};

const statusColor: Record<string, "green" | "orange" | "rose" | "default"> = {
  active: "green", trialing: "blue" as "default", past_due: "orange", canceled: "rose", paused: "default", unpaid: "rose",
};

export function Component() {
  const { data: plans, isLoading: loadingPlans } = usePlans();
  const { data: sub, isLoading: loadingSub } = useSubscription();
  const checkout = useCheckout();
  const portal = usePortal();

  if (loadingPlans || loadingSub) return <Skeleton className="h-96 rounded-[14px]" />;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Plano e Cobrança</h1>

      {sub && (
        <div className="mb-8 rounded-[14px] border border-border bg-bg-card p-6">
          <div className="flex items-center justify-between mb-4">
            <div>
              <div className="text-xs text-text-muted mb-1">Plano atual</div>
              <div className="text-lg font-bold capitalize">{sub.tier}</div>
            </div>
            <Badge variant={statusColor[sub.status]}>{statusLabel[sub.status]}</Badge>
          </div>
          {sub.current_period_end && (
            <p className="text-sm text-text-secondary mb-4">
              Próximo vencimento: {new Date(sub.current_period_end).toLocaleDateString("pt-BR")}
            </p>
          )}
          <Button variant="outline" size="sm" onClick={() => portal.mutate()} disabled={portal.isPending}>
            <CreditCard className="h-4 w-4" /> Gerenciar assinatura <ExternalLink className="h-3 w-3" />
          </Button>
        </div>
      )}

      <h2 className="text-lg font-semibold mb-4">Todos os planos</h2>
      <div className="grid gap-4 md:grid-cols-3">
        {(plans ?? []).map((plan) => {
          const isCurrent = sub?.tier === plan.tier;
          return (
            <div key={plan.id} className={cn(
              "flex flex-col rounded-[14px] border p-6",
              isCurrent ? "border-accent-orange bg-accent-orange/5" : "border-border bg-bg-card",
            )}>
              <h3 className="text-lg font-bold">{plan.name}</h3>
              <div className="mt-2 flex items-baseline gap-1">
                <span className="text-2xl font-bold">{formatCurrency(plan.price_monthly)}</span>
                <span className="text-sm text-text-muted">/mês</span>
              </div>
              <ul className="mt-4 flex-1 space-y-2 text-sm text-text-secondary">
                <li className="flex items-center gap-2"><Check className="h-3.5 w-3.5 text-accent-green" />{plan.limits.max_areas < 0 || plan.limits.max_areas >= 999 ? "Áreas ilimitadas" : `${plan.limits.max_areas} áreas`}</li>
                <li className="flex items-center gap-2"><Check className="h-3.5 w-3.5 text-accent-green" />{plan.limits.max_goals < 0 || plan.limits.max_goals >= 999 ? "Metas ilimitadas" : `${plan.limits.max_goals} metas`}</li>
                {plan.limits.finance_enabled && <li className="flex items-center gap-2"><Check className="h-3.5 w-3.5 text-accent-green" />Módulo financeiro</li>}
                {plan.limits.ai_insights && <li className="flex items-center gap-2"><Check className="h-3.5 w-3.5 text-accent-green" />AI Insights</li>}
                {plan.limits.export_enabled && <li className="flex items-center gap-2"><Check className="h-3.5 w-3.5 text-accent-green" />Exportação de dados</li>}
              </ul>
              {isCurrent ? (
                <Button variant="outline" className="mt-4 w-full" disabled>Plano atual</Button>
              ) : (
                <Button className="mt-4 w-full" onClick={() => checkout.mutate(plan.id)} disabled={checkout.isPending}>
                  Atualizar plano
                </Button>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
