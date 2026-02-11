import { WeeklyStats } from "@/components/dashboard/weekly-stats";
import { LifeScoreRing } from "@/components/dashboard/life-score-ring";
import { AreasGrid } from "@/components/dashboard/areas-grid";
import { GoalsCard } from "@/components/dashboard/goals-card";
import { HabitsHeatmap } from "@/components/dashboard/habits-heatmap";
import { TodayFocus } from "@/components/dashboard/today-focus";

export function Component() {
  return (
    <div>
      {/* Weekly Stats bar */}
      <WeeklyStats />

      {/* Top section: Life Score + Areas */}
      <div className="mb-4 grid grid-cols-1 gap-4 md:mb-5 md:gap-5 lg:grid-cols-[300px_1fr]">
        <LifeScoreRing />
        <AreasGrid />
      </div>

      {/* Bottom grid: Goals | Habits | Today */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 md:gap-5 lg:grid-cols-[1fr_1fr_340px]">
        <GoalsCard />
        <HabitsHeatmap />
        <TodayFocus />
      </div>
    </div>
  );
}
