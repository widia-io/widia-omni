import { useGoals } from "@/hooks/use-goals";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";

const statusVariant: Record<string, "green" | "orange" | "rose"> = {
  on_track: "green",
  at_risk: "orange",
  behind: "rose",
  not_started: "default" as "green",
  completed: "green",
};

const statusLabel: Record<string, string> = {
  on_track: "On Track",
  at_risk: "At Risk",
  behind: "Behind",
  not_started: "Nao Iniciado",
  completed: "Concluido",
};

export function GoalsCard() {
  const { data: goals, isLoading } = useGoals({ period: "quarterly" });

  if (isLoading) {
    return <Skeleton className="h-80 rounded-[14px]" />;
  }

  return (
    <Card>
      <CardHeader>
        <span className="h-1.5 w-1.5 rounded-full bg-accent-orange animate-pulse" />
        <CardTitle>Metas</CardTitle>
      </CardHeader>
      <CardContent>
        {(goals ?? []).slice(0, 5).map((goal) => {
          const pct = goal.target_value
            ? Math.round((goal.current_value / goal.target_value) * 100)
            : 0;

          return (
            <div key={goal.id} className="border-b border-border py-3.5 last:border-b-0">
              <div className="flex items-center justify-between mb-2">
                <span className="font-serif text-sm font-medium">{goal.title}</span>
                <span className="font-mono text-xs text-accent-orange">{pct}%</span>
              </div>
              <div className="h-1 overflow-hidden rounded bg-border">
                <div
                  className="h-full rounded bg-gradient-to-r from-accent-orange to-accent-sand transition-all duration-[1.5s]"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <div className="mt-1.5 flex gap-2">
                <Badge variant="blue">{goal.period === "quarterly" ? "Q" : goal.period[0]?.toUpperCase()}</Badge>
                <Badge variant={statusVariant[goal.status] ?? "default"}>
                  {statusLabel[goal.status] ?? goal.status}
                </Badge>
              </div>
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
}
