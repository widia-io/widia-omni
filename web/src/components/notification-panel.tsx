import { formatRelative } from "@/lib/format";
import { useNotifications, useMarkRead, useMarkAllRead } from "@/hooks/use-notifications";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/cn";
import type { NotificationType } from "@/types/api";

const typeIcon: Record<NotificationType, string> = {
  weekly_review: "📊",
  streak_at_risk: "🔥",
  goal_deadline: "🎯",
  trial_ending: "⏰",
  plan_changed: "💳",
  score_update: "📈",
  habit_reminder: "✅",
  system: "ℹ️",
};

export function NotificationPanel() {
  const { data: notifications, isLoading } = useNotifications();
  const markRead = useMarkRead();
  const markAll = useMarkAllRead();

  if (isLoading) return <div className="p-4"><Skeleton className="h-20 rounded-lg" /></div>;

  const items = notifications ?? [];

  return (
    <div className="w-80 rounded-[14px] border border-border bg-bg-card shadow-lg">
      <div className="flex items-center justify-between border-b border-border px-4 py-3">
        <span className="text-sm font-semibold">Notificações</span>
        {items.some((n) => !n.is_read) && (
          <button onClick={() => markAll.mutate()} className="text-xs text-accent-orange hover:underline">
            Marcar todas como lidas
          </button>
        )}
      </div>
      <div className="max-h-80 overflow-y-auto">
        {items.length === 0 && (
          <div className="p-6 text-center text-sm text-text-muted">Nenhuma notificação</div>
        )}
        {items.map((n) => (
          <button
            key={n.id}
            onClick={() => !n.is_read && markRead.mutate(n.id)}
            className={cn(
              "flex w-full gap-3 border-b border-border/50 px-4 py-3 text-left transition-colors hover:bg-bg-primary/50",
              !n.is_read && "bg-accent-orange/5",
            )}
          >
            <span className="mt-0.5 text-base">{typeIcon[n.type]}</span>
            <div className="flex-1 min-w-0">
              <div className={cn("text-sm", !n.is_read && "font-medium")}>{n.title}</div>
              {n.body && <div className="mt-0.5 text-xs text-text-muted line-clamp-2">{n.body}</div>}
              <div className="mt-1 text-[11px] text-text-muted">{formatRelative(n.created_at)}</div>
            </div>
            {!n.is_read && <span className="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-accent-orange" />}
          </button>
        ))}
      </div>
    </div>
  );
}
