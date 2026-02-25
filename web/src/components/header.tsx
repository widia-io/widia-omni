import { useState, useRef, useEffect } from "react";
import { Link } from "react-router";
import { Bell } from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";
import { getGreeting, formatDateShort } from "@/lib/format";
import { useNotificationCount } from "@/hooks/use-notifications";
import { useDashboard } from "@/hooks/use-dashboard";
import { useSubscription } from "@/hooks/use-billing";
import { useSwitchWorkspace, useWorkspaces } from "@/hooks/use-workspaces";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { NotificationPanel } from "@/components/notification-panel";
import type { PlanTier } from "@/types/api";

const tierLabel: Record<PlanTier, string> = {
  free: "Free",
  pro: "Pro",
  premium: "Premium",
};

const tierBadge: Record<PlanTier, "default" | "orange" | "blue"> = {
  free: "default",
  pro: "orange",
  premium: "blue",
};

export function Header() {
  const user = useAuthStore((s) => s.user);
  const { data: unreadCount } = useNotificationCount();
  const { data: dash } = useDashboard();
  const { data: subscription } = useSubscription();
  const { data: workspaces } = useWorkspaces();
  const switchWorkspace = useSwitchWorkspace();
  const [showNotifs, setShowNotifs] = useState(false);
  const panelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (panelRef.current && !panelRef.current.contains(e.target as Node)) setShowNotifs(false);
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  const defaultWorkspace = workspaces?.find((w) => w.is_default) ?? workspaces?.[0];

  const initials = (user?.display_name ?? "U")
    .split(" ")
    .map((w) => w[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  return (
    <header className="relative z-40 animate-in mb-6 flex flex-col gap-4 md:mb-8 md:flex-row md:items-start md:justify-between">
      <div>
        <h1 className="text-[28px] font-bold tracking-tight">
          {getGreeting()},{" "}
          <span className="bg-gradient-to-r from-accent-orange to-accent-sand bg-clip-text text-transparent">
            {user?.display_name?.split(" ")[0] ?? "Usuário"}
          </span>
        </h1>
        <p className="mt-1 font-serif text-sm text-text-secondary">
          {dash
            ? `${dash.today_tasks} tarefas pendentes · ${dash.habits_today} hábitos hoje · ${dash.active_projects} projetos ativos`
            : "Gerencie suas metas e hábitos"}
        </p>
      </div>
      <div className="flex flex-wrap items-center gap-2 md:gap-3">
        {workspaces && workspaces.length > 0 && (
          <Select
            value={defaultWorkspace?.id}
            onValueChange={(workspaceID) => switchWorkspace.mutate(workspaceID)}
            disabled={switchWorkspace.isPending}
          >
            <SelectTrigger className="h-10 w-[170px] max-w-[65vw] bg-bg-card text-xs">
              <SelectValue placeholder="Workspace" />
            </SelectTrigger>
            <SelectContent>
              {workspaces.map((ws) => (
                <SelectItem key={ws.id} value={ws.id}>
                  {ws.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
        {subscription && (
          <Link to="/billing" title="Gerenciar plano">
            <Badge variant={tierBadge[subscription.tier]} className="rounded-full px-2.5 py-1 text-[11px]">
              {tierLabel[subscription.tier]}
            </Badge>
          </Link>
        )}
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
            <div className="absolute right-0 top-12 z-[60]">
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
