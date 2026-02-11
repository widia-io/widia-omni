import { NavLink } from "react-router";
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

export function Sidebar() {
  const theme = useUIStore((s) => s.theme);
  const cycleTheme = useUIStore((s) => s.cycleTheme);
  const ThemeIcon = themeIcons[theme];

  return (
    <>
      <TooltipProvider delayDuration={0}>
        <nav className="fixed left-0 top-0 z-20 hidden h-screen w-[72px] flex-col items-center gap-2 border-r border-border bg-bg-secondary py-5 md:flex">
          <div className="mb-5">
            <Logo size={40} />
          </div>

          {navItems.map(({ icon: Icon, to, label }) => (
            <Tooltip key={to}>
              <TooltipTrigger asChild>
                <NavLink
                  to={to}
                  className={({ isActive }) =>
                    cn(
                      "relative flex h-11 w-11 items-center justify-center rounded-xl text-text-muted transition-all duration-200 hover:bg-bg-card hover:text-text-secondary",
                      isActive && "bg-accent-orange/15 text-accent-orange shadow-sm ring-1 ring-accent-orange/20",
                    )
                  }
                >
                  {({ isActive }) => (
                    <>
                      {isActive && (
                        <span className="absolute -left-[14px] h-6 w-[3px] rounded-r-[3px] bg-accent-orange" />
                      )}
                      <Icon className={cn("h-5 w-5", isActive && "scale-110")} />
                    </>
                  )}
                </NavLink>
              </TooltipTrigger>
              <TooltipContent side="right">{label}</TooltipContent>
            </Tooltip>
          ))}

          <div className="flex-1" />

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

          <Tooltip>
            <TooltipTrigger asChild>
              <NavLink
                to="/settings"
                className={({ isActive }) =>
                  cn(
                    "relative flex h-11 w-11 items-center justify-center rounded-xl text-text-muted transition-all duration-200 hover:bg-bg-card hover:text-text-secondary",
                    isActive && "bg-accent-orange/15 text-accent-orange shadow-sm ring-1 ring-accent-orange/20",
                  )
                }
              >
                {({ isActive }) => (
                  <>
                    {isActive && (
                      <span className="absolute -left-[14px] h-6 w-[3px] rounded-r-[3px] bg-accent-orange" />
                    )}
                    <Settings className={cn("h-5 w-5", isActive && "scale-110")} />
                  </>
                )}
              </NavLink>
            </TooltipTrigger>
            <TooltipContent side="right">Configurações</TooltipContent>
          </Tooltip>
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
