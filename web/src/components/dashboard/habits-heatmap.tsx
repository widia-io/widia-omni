import { useMemo } from "react";
import { format, subWeeks, startOfWeek, addDays } from "date-fns";
import { useHabits, useHabitEntries, useHabitStreaks } from "@/hooks/use-habits";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/cn";

const cellColors: Record<string, Record<number, string>> = {
  green: { 1: "bg-accent-green/20", 2: "bg-accent-green/50", 3: "bg-accent-green" },
  blue: { 1: "bg-accent-blue/20", 2: "bg-accent-blue/50", 3: "bg-accent-blue" },
  orange: { 1: "bg-accent-orange/20", 2: "bg-accent-orange/50", 3: "bg-accent-orange" },
  sand: { 1: "bg-accent-sand/20", 2: "bg-accent-sand/50", 3: "bg-accent-sand" },
  sage: { 1: "bg-accent-sage/20", 2: "bg-accent-sage/50", 3: "bg-accent-sage" },
  rose: { 1: "bg-accent-rose/20", 2: "bg-accent-rose/50", 3: "bg-accent-rose" },
};

const dayLabels = ["S", "T", "Q", "Q", "S", "S", "D"];

export function HabitsHeatmap() {
  const { data: habits, isLoading: loadingHabits } = useHabits();
  const { data: streaks } = useHabitStreaks();

  const today = new Date();
  const from = format(startOfWeek(subWeeks(today, 3), { weekStartsOn: 1 }), "yyyy-MM-dd");
  const to = format(today, "yyyy-MM-dd");
  const { data: entries, isLoading: loadingEntries } = useHabitEntries(from, to);

  const entryMap = useMemo(() => {
    const m = new Map<string, number>();
    (entries ?? []).forEach((e) => {
      m.set(`${e.habit_id}:${e.date.slice(0, 10)}`, e.intensity);
    });
    return m;
  }, [entries]);

  const weeks = useMemo(() => {
    const result: string[][] = [];
    for (let w = 3; w >= 0; w--) {
      const weekStart = startOfWeek(subWeeks(today, w), { weekStartsOn: 1 });
      const days = Array.from({ length: 7 }, (_, i) => format(addDays(weekStart, i), "yyyy-MM-dd"));
      result.push(days);
    }
    return result;
  }, []);

  if (loadingHabits || loadingEntries) {
    return <Skeleton className="h-80 rounded-[14px]" />;
  }

  return (
    <Card>
      <CardHeader>
        <span className="h-1.5 w-1.5 rounded-full bg-accent-green animate-pulse" />
        <CardTitle>Habitos · Ultimas 4 Semanas</CardTitle>
      </CardHeader>
      <CardContent>
        {/* Day headers */}
        <div className="flex gap-[3px] ml-[90px] mb-1.5">
          {weeks.map((_, wi) =>
            dayLabels.map((d, di) => (
              <div key={`${wi}-${di}`} className={cn("w-[18px] text-center font-mono text-[9px] text-text-muted", wi > 0 && di === 0 && "ml-1")}>
                {d}
              </div>
            )),
          )}
        </div>

        {/* Habit rows */}
        <div className="flex flex-col gap-2.5">
          {(habits ?? []).slice(0, 6).map((habit) => {
            const colors = cellColors[habit.color] ?? cellColors["green"]!;
            return (
              <div key={habit.id} className="flex items-center gap-2.5">
                <div className="w-20 shrink-0 text-right font-serif text-xs text-text-secondary truncate">
                  {habit.name}
                </div>
                <div className="flex gap-[3px] flex-1">
                  {weeks.map((weekDays, wi) =>
                    weekDays.map((date) => {
                      const intensity = entryMap.get(`${habit.id}:${date}`) ?? 0;
                      return (
                        <div
                          key={`${wi}-${date}`}
                          className={cn(
                            "h-[18px] w-[18px] rounded-[4px] transition-all duration-150 hover:scale-130 hover:z-10",
                            intensity === 0 ? "bg-border opacity-30" : colors[Math.min(intensity, 3)],
                          )}
                        />
                      );
                    }),
                  )}
                </div>
              </div>
            );
          })}
        </div>

        {/* Streaks */}
        {streaks && streaks.length > 0 && (
          <div className="mt-4 flex items-center gap-4 border-t border-border pt-3.5">
            {streaks.slice(0, 3).map((s) => (
              <div key={s.habit_id} className="flex items-center gap-1.5">
                <span className="text-sm">🔥</span>
                <span className="font-mono text-lg font-bold text-accent-orange">{s.current_streak}</span>
                <span className="text-[11px] text-text-muted leading-tight">dias<br />{s.name}</span>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
