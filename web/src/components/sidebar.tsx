import { NavLink } from "react-router";
import {
  LayoutDashboard,
  Target,
  Activity,
  DollarSign,
  Calendar,
  PenLine,
  Settings,
} from "lucide-react";
import { cn } from "@/lib/cn";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";

const navItems = [
  { icon: LayoutDashboard, to: "/dashboard", label: "Dashboard" },
  { icon: Target, to: "/goals", label: "Metas" },
  { icon: Activity, to: "/habits", label: "Habitos" },
  { icon: DollarSign, to: "/finances", label: "Financas" },
  { icon: Calendar, to: "/tasks", label: "Tarefas" },
  { icon: PenLine, to: "/journal", label: "Journal" },
] as const;

export function Sidebar() {
  return (
    <TooltipProvider delayDuration={0}>
      <nav className="fixed top-0 left-0 z-10 flex h-screen w-[72px] flex-col items-center gap-2 border-r border-border bg-bg-secondary py-5">
        {/* Logo */}
        <div className="mb-5 flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-accent-orange to-accent-sand font-mono text-sm font-bold text-bg-primary animate-[logoPulse_4s_ease-in-out_infinite]">
          MC
        </div>

        {/* Nav items */}
        {navItems.map(({ icon: Icon, to, label }) => (
          <Tooltip key={to}>
            <TooltipTrigger asChild>
              <NavLink
                to={to}
                className={({ isActive }) =>
                  cn(
                    "relative flex h-11 w-11 items-center justify-center rounded-xl text-text-muted transition-all duration-200 hover:bg-bg-card hover:text-text-secondary",
                    isActive && "bg-accent-orange-soft text-accent-orange",
                  )
                }
              >
                {({ isActive }) => (
                  <>
                    {isActive && (
                      <span className="absolute -left-[14px] h-5 w-[3px] rounded-r-[3px] bg-accent-orange" />
                    )}
                    <Icon className="h-5 w-5" />
                  </>
                )}
              </NavLink>
            </TooltipTrigger>
            <TooltipContent side="right">{label}</TooltipContent>
          </Tooltip>
        ))}

        {/* Spacer */}
        <div className="flex-1" />

        {/* Settings */}
        <Tooltip>
          <TooltipTrigger asChild>
            <NavLink
              to="/settings"
              className={({ isActive }) =>
                cn(
                  "relative flex h-11 w-11 items-center justify-center rounded-xl text-text-muted transition-all duration-200 hover:bg-bg-card hover:text-text-secondary",
                  isActive && "bg-accent-orange-soft text-accent-orange",
                )
              }
            >
              {({ isActive }) => (
                <>
                  {isActive && (
                    <span className="absolute -left-[14px] h-5 w-[3px] rounded-r-[3px] bg-accent-orange" />
                  )}
                  <Settings className="h-5 w-5" />
                </>
              )}
            </NavLink>
          </TooltipTrigger>
          <TooltipContent side="right">Configuracoes</TooltipContent>
        </Tooltip>
      </nav>
    </TooltipProvider>
  );
}
