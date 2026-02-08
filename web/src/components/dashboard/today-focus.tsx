import { useTasks, useCompleteTask } from "@/hooks/use-tasks";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/cn";
import { MoodSelector } from "./mood-selector";

export function TodayFocus() {
  const { data: tasks, isLoading } = useTasks({ is_focus: "true" });
  const complete = useCompleteTask();

  if (isLoading) {
    return <Skeleton className="h-80 rounded-[14px]" />;
  }

  return (
    <Card>
      <CardHeader>
        <span className="h-1.5 w-1.5 rounded-full bg-accent-rose animate-pulse" />
        <CardTitle>Foco de Hoje</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col">
          {(tasks ?? []).map((task) => (
            <div key={task.id} className="flex items-start gap-3 border-b border-border py-3 last:border-b-0">
              <button
                onClick={() => !task.is_completed && complete.mutate(task.id)}
                className={cn(
                  "mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-md border-2 transition-all",
                  task.is_completed
                    ? "border-accent-green bg-accent-green"
                    : "border-border hover:border-accent-orange",
                )}
              >
                {task.is_completed && <span className="text-[11px] font-bold text-bg-primary">✓</span>}
              </button>
              <div className="flex-1">
                <div className={cn("font-serif text-[13px] font-medium leading-relaxed", task.is_completed && "text-text-muted line-through")}>
                  {task.title}
                </div>
                {task.area_id && (
                  <Badge variant="blue" className="mt-1">{task.area_id.slice(0, 6)}</Badge>
                )}
              </div>
            </div>
          ))}
        </div>
        <MoodSelector />
      </CardContent>
    </Card>
  );
}
