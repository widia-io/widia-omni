import { useCurrentScore } from "@/hooks/use-scores";
import { Skeleton } from "@/components/ui/skeleton";
import { format } from "date-fns";
import { ptBR } from "date-fns/locale";

export function LifeScoreRing() {
  const { data: score, isLoading } = useCurrentScore();

  if (isLoading) {
    return <Skeleton className="h-full min-h-[300px] rounded-[14px]" />;
  }

  const value = score?.score ?? 0;
  const circumference = 2 * Math.PI * 65;
  const offset = circumference - (value / 100) * circumference;
  const weekLabel = score?.week_start
    ? `Semana ${format(new Date(score.week_start), "w", { locale: ptBR })} de ${format(new Date(score.week_start), "yyyy")}`
    : "";

  return (
    <div className="relative flex flex-col items-center justify-center overflow-hidden rounded-[14px] border border-border bg-bg-card p-7">
      {/* Radial glow */}
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_30%,#d9775708,transparent_70%)]" />

      {/* SVG Ring */}
      <div className="relative mb-4 h-40 w-40">
        <svg className="h-40 w-40 -rotate-90" viewBox="0 0 140 140">
          <defs>
            <linearGradient id="scoreGrad" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stopColor="#d97757" />
              <stop offset="100%" stopColor="#c4a882" />
            </linearGradient>
          </defs>
          <circle cx="70" cy="70" r="65" fill="none" stroke="var(--color-border)" strokeWidth="8" />
          <circle
            cx="70" cy="70" r="65"
            fill="none"
            stroke="url(#scoreGrad)"
            strokeWidth="8"
            strokeLinecap="round"
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            className="transition-all duration-[2s] ease-out"
          />
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className="bg-gradient-to-r from-accent-orange to-accent-sand bg-clip-text text-[42px] font-extrabold leading-none tracking-tighter text-transparent">
            {value}
          </span>
          <span className="mt-0.5 font-mono text-[10px] uppercase tracking-[2px] text-text-muted">
            Life Score
          </span>
        </div>
      </div>

      <div className="font-serif text-[13px] text-text-secondary">{weekLabel}</div>
    </div>
  );
}
