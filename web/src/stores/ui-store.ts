import { create } from "zustand";

type Theme = "light" | "dark" | "system";

interface UIState {
  sidebarCollapsed: boolean;
  theme: Theme;
  toggleSidebar: () => void;
  setTheme: (theme: Theme) => void;
  cycleTheme: () => void;
}

function getInitialTheme(): Theme {
  if (typeof window === "undefined") return "system";
  return (localStorage.getItem("theme") as Theme) ?? "system";
}

export const useUIStore = create<UIState>()((set, get) => ({
  sidebarCollapsed: false,
  theme: getInitialTheme(),
  toggleSidebar: () => set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),
  setTheme: (theme) => {
    localStorage.setItem("theme", theme);
    set({ theme });
  },
  cycleTheme: () => {
    const order: Theme[] = ["light", "dark", "system"];
    const next = order[(order.indexOf(get().theme) + 1) % 3]!;
    get().setTheme(next);
  },
}));
