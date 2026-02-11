import { useState, useRef, useEffect } from "react";
import { Bell } from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";
import { getGreeting, formatDateShort } from "@/lib/format";
import { useNotificationCount } from "@/hooks/use-notifications";
import { useDashboard } from "@/hooks/use-dashboard";
import { NotificationPanel } from "@/components/notification-panel";

export function Header() {
  const user = useAuthStore((s) => s.user);
  const { data: unreadCount } = useNotificationCount();
  const { data: dash } = useDashboard();
  const [showNotifs, setShowNotifs] = useState(false);
  const panelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) setShowNotifs(false);
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  const initials = (user?.display_name ?? "U")
    .split(" ")
    .map((w) => w[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  return (
    <header className="mb-8 flex items-start justify-between animate-in">
      <div>
        <h1 className="text-[28px] font-bold tracking-tight">
          {getGreeting()},{" "}
          <span className="bg-gradient-to-r from-accent-orange to-accent-sand bg-clip-text text-transparent">
            {user?.display_name?.split(" ")[0] ?? "Usuário"}
          </span>
        </h1>
        <p className="mt-1 font-serif text-sm text-text-secondary">
          {dash
            ? `${dash.today_tasks - dash.completed_today} tarefas pendentes · ${dash.habits_today} hábitos hoje`
            : "Gerencie suas metas e hábitos"}
        </p>
      </div>
      <div className="flex items-center gap-3">
        <div className="rounded-lg border border-border bg-bg-card px-3.5 py-2 font-mono text-[13px] text-text-secondary">
          {formatDateShort(new Date())}
        </div>
        <div className="relative" ref={panelRef}>
          <button
            onClick={() => setShowNotifs((p) => !p)}
            className="relative flex h-10 w-10 items-center justify-center rounded-lg border border-border bg-bg-card text-text-muted transition-colors hover:text-text-primary"
          >
            <Bell className="h-[18px] w-[18px]" />
            {(unreadCount ?? 0) > 0 && (
              <span className="absolute -right-1 -top-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-accent-orange px-1 text-[10px] font-bold text-bg-primary">
                {unreadCount}
              </span>
            )}
          </button>
          {showNotifs && (
            <div className="absolute right-0 top-12 z-50">
              <NotificationPanel />
            </div>
          )}
        </div>
        <div className="flex h-[38px] w-[38px] items-center justify-center rounded-[10px] bg-gradient-to-br from-accent-orange to-accent-rose text-sm font-semibold text-text-primary">
          {initials}
        </div>
      </div>
    </header>
  );
}
