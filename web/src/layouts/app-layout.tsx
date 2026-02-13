import { useEffect, useState } from "react";
import { Outlet } from "react-router";
import { Sidebar } from "@/components/sidebar";
import { Header } from "@/components/header";
import { useWorkspaceUsage } from "@/hooks/use-settings";
import { useUIStore } from "@/stores/ui-store";

export function AppLayout() {
  const { data: usage } = useWorkspaceUsage();
  const collapsed = useUIStore((s) => s.sidebarCollapsed);
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    const mql = window.matchMedia("(max-width: 767px)");
    const sync = () => setIsMobile(mql.matches);
    sync();
    mql.addEventListener("change", sync);
    return () => mql.removeEventListener("change", sync);
  }, []);

  const showMobileBetaNotice = isMobile && usage && !usage.limits.mobile_pwa_enabled;

  return (
    <>
      <Sidebar />
      <main
        className="min-h-screen px-4 pb-24 pt-4 transition-all duration-300 md:px-8 md:py-6"
        style={{ marginLeft: isMobile ? 0 : collapsed ? 72 : 220 }}
      >
        {showMobileBetaNotice && (
          <div className="mb-3 rounded-lg border border-accent-rose/30 bg-accent-rose/10 px-3 py-2 text-xs text-accent-rose">
            Beta mobile ainda nao habilitado para este workspace.
          </div>
        )}
        <Header />
        <Outlet />
      </main>
    </>
  );
}
