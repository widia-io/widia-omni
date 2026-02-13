import { Link, NavLink, useLocation } from "react-router";
import {
  LayoutDashboard,
  Target,
  Activity,
  DollarSign,
  Calendar,
  PenLine,
  Settings,
  Sun,
  Moon,
  Monitor,
  ChevronsRight,
  ChevronsLeft,
} from "lucide-react";
import { cn } from "@/lib/cn";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { useUIStore } from "@/stores/ui-store";
import { Logo } from "@/components/logo";

const navItems = [
  { icon: LayoutDashboard, to: "/dashboard", label: "Dashboard" },
  { icon: Target, to: "/goals", label: "Metas" },
  { icon: Activity, to: "/habits", label: "Hábitos" },
  { icon: DollarSign, to: "/finances", label: "Finanças" },
  { icon: Calendar, to: "/tasks", label: "Tarefas" },
  { icon: PenLine, to: "/journal", label: "Journal" },
] as const;

const mobileNavItems = [
  { icon: LayoutDashboard, to: "/dashboard", label: "Dashboard" },
  { icon: Calendar, to: "/tasks", label: "Tarefas" },
  { icon: Target, to: "/goals", label: "Metas" },
  { icon: Activity, to: "/habits", label: "Hábitos" },
  { icon: DollarSign, to: "/finances", label: "Finanças" },
  { icon: Settings, to: "/settings", label: "Ajustes" },
] as const;

const themeIcons = { light: Sun, dark: Moon, system: Monitor } as const;
const themeLabels = { light: "Claro", dark: "Escuro", system: "Sistema" } as const;

function NavItemLink({ icon: Icon, to, label, collapsed }: { icon: typeof LayoutDashboard; to: string; label: string; collapsed: boolean }) {
  const { pathname } = useLocation();
  const isActive = pathname === to || pathname.startsWith(to + "/");

  const className = cn(
    "relative flex h-11 items-center rounded-xl text-text-muted transition-all duration-200 hover:bg-bg-card hover:text-text-secondary",
    collapsed ? "w-11 justify-center" : "w-full gap-3 px-3",
    isActive && "bg-accent-orange/15 text-accent-orange shadow-sm ring-1 ring-accent-orange/20",
  );

  const content = (
    <>
      {isActive && (
        <span className={cn("absolute h-6 w-[3px] rounded-r-[3px] bg-accent-orange", collapsed ? "-left-[14px]" : "-left-[16px]")} />
      )}
      <Icon className={cn("h-5 w-5 shrink-0", isActive && "scale-110")} />
      {!collapsed && <span className="truncate text-sm font-medium">{label}</span>}
    </>
  );

  if (collapsed) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <Link to={to} className={className}>{content}</Link>
        </TooltipTrigger>
        <TooltipContent side="right">{label}</TooltipContent>
      </Tooltip>
    );
  }

  return <Link to={to} className={className}>{content}</Link>;
}

export function Sidebar() {
  const theme = useUIStore((s) => s.theme);
  const cycleTheme = useUIStore((s) => s.cycleTheme);
  const collapsed = useUIStore((s) => s.sidebarCollapsed);
  const toggleSidebar = useUIStore((s) => s.toggleSidebar);
  const ThemeIcon = themeIcons[theme];
  const ToggleIcon = collapsed ? ChevronsRight : ChevronsLeft;

  return (
    <>
      <TooltipProvider delayDuration={0}>
        <nav
          className={cn(
            "fixed left-0 top-0 z-20 hidden h-screen flex-col border-r border-border bg-bg-secondary py-5 transition-all duration-300 md:flex",
            collapsed ? "w-[72px] items-center gap-2" : "w-[220px] gap-1 px-4",
          )}
        >
          <div className={cn("mb-5", collapsed ? "flex justify-center" : "flex items-center gap-3 px-1")}>
            <Logo size={40} />
            {!collapsed && <span className="text-lg font-bold text-text-primary">Widia</span>}
          </div>

          <div className={cn("flex flex-col", collapsed ? "items-center gap-2" : "gap-1")}>
            {navItems.map((item) => (
              <NavItemLink key={item.to} {...item} collapsed={collapsed} />
            ))}
          </div>

          <div className="flex-1" />

          {/* Toggle */}
          {collapsed ? (
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  onClick={toggleSidebar}
                  className="flex h-11 w-11 items-center justify-center rounded-xl text-text-muted transition-all duration-200 hover:bg-bg-card hover:text-text-secondary"
                >
                  <ToggleIcon className="h-5 w-5" />
                </button>
              </TooltipTrigger>
              <TooltipContent side="right">Expandir</TooltipContent>
            </Tooltip>
          ) : (
            <button
              onClick={toggleSidebar}
              className="flex h-11 w-full items-center gap-3 rounded-xl px-3 text-text-muted transition-all duration-200 hover:bg-bg-card hover:text-text-secondary"
            >
              <ToggleIcon className="h-5 w-5 shrink-0" />
              <span className="truncate text-sm font-medium">Recolher</span>
            </button>
          )}

          {/* Theme */}
          {collapsed ? (
            <Tooltip>
              <TooltipTrigger asChild>
                <button
                  onClick={cycleTheme}
                  className="flex h-11 w-11 items-center justify-center rounded-xl text-text-muted transition-all duration-200 hover:bg-bg-card hover:text-text-secondary"
                >
                  <ThemeIcon className="h-5 w-5" />
                </button>
              </TooltipTrigger>
              <TooltipContent side="right">Tema: {themeLabels[theme]}</TooltipContent>
            </Tooltip>
          ) : (
            <button
              onClick={cycleTheme}
              className="flex h-11 w-full items-center gap-3 rounded-xl px-3 text-text-muted transition-all duration-200 hover:bg-bg-card hover:text-text-secondary"
            >
              <ThemeIcon className="h-5 w-5 shrink-0" />
              <span className="truncate text-sm font-medium">Tema: {themeLabels[theme]}</span>
            </button>
          )}

          {/* Settings */}
          <NavItemLink icon={Settings} to="/settings" label="Configurações" collapsed={collapsed} />
        </nav>
      </TooltipProvider>

      <nav className="fixed bottom-0 left-0 right-0 z-30 border-t border-border bg-bg-card/95 px-2 pb-2 pt-1 backdrop-blur md:hidden">
        <div className="grid grid-cols-6 gap-1">
          {mobileNavItems.map(({ icon: Icon, to, label }) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                cn(
                  "flex flex-col items-center justify-center rounded-lg py-2 text-[10px] text-text-muted",
                  isActive && "bg-accent-orange/10 text-accent-orange",
                )
              }
            >
              <Icon className="mb-1 h-4 w-4" />
              <span className="truncate">{label}</span>
            </NavLink>
          ))}
        </div>
        <button
          onClick={cycleTheme}
          className="mx-auto mt-1 flex items-center gap-1 rounded-md px-2 py-1 text-[10px] text-text-muted"
        >
          <ThemeIcon className="h-3 w-3" />
          Tema {themeLabels[theme]}
        </button>
      </nav>
    </>
  );
}
