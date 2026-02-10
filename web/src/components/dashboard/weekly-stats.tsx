import { useDashboard } from "@/hooks/use-dashboard";
import { Skeleton } from "@/components/ui/skeleton";

interface StatBarProps {
  heights: number[];
  color: string;
}

function StatChart({ heights, color }: StatBarProps) {
  return (
    <div className="flex items-end gap-[3px] h-10">
      {heights.map((h, i) => (
        <div
          key={i}
          className="w-1.5 rounded-[3px] transition-all duration-1000"
          style={{ height: `${h}%`, background: `color-mix(in srgb, ${color} ${20 + h * 0.8}%, transparent)` }}
        />
      ))}
    </div>
  );
}

export function WeeklyStats() {
  const { data, isLoading } = useDashboard();

  if (isLoading) {
    return (
      <div className="grid grid-cols-3 gap-5 mb-5">
        {[1, 2, 3].map((i) => <Skeleton key={i} className="h-24 rounded-[14px]" />)}
      </div>
    );
  }

  const stats = [
    {
      label: "Foco Semanal",
      value: data?.completed_today ?? 0,
      suffix: `/${data?.today_tasks ?? 0}`,
      sub: "tarefas concluidas",
      color: "var(--area-indigo)",
      bars: [60, 80, 45, 90, 70, 30, 15],
    },
    {
      label: "Habitos Hoje",
      value: data?.habits_today ?? 0,
      suffix: "",
      sub: "check-ins realizados",
      color: "var(--area-green)",
      bars: [50, 65, 70, 85, 90, 40, 20],
    },
    {
      label: "Streaks Ativos",
      value: data?.current_streaks ?? 0,
      suffix: "",
      sub: "habitos consecutivos",
      color: "var(--accent)",
      bars: [40, 55, 60, 50, 75, 85, 20],
    },
  ];

  return (
    <div className="grid grid-cols-1 gap-5 mb-5 md:grid-cols-3">
      {stats.map((s, i) => (
        <div
          key={i}
          className="flex items-center justify-between rounded-[14px] border border-border bg-bg-card px-[22px] py-5 animate-in"
          style={{ animationDelay: `${0.05 + i * 0.05}s` }}
        >
          <div>
            <div className="text-xs text-text-muted mb-1">{s.label}</div>
            <div className="font-mono text-[28px] font-bold" style={{ color: s.color }}>
              {s.value}
              {s.suffix && <span className="text-base text-text-muted">{s.suffix}</span>}
            </div>
            <div className="font-serif text-xs text-text-secondary mt-0.5">{s.sub}</div>
          </div>
          <StatChart heights={s.bars} color={s.color} />
        </div>
      ))}
    </div>
  );
}
