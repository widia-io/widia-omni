import { useEffect, useState } from "react";
import { Outlet } from "react-router";
import { Sidebar } from "@/components/sidebar";
import { Header } from "@/components/header";
import { useWorkspaceUsage } from "@/hooks/use-settings";

export function AppLayout() {
  const { data: usage } = useWorkspaceUsage();
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
      <main className="min-h-screen px-4 pb-24 pt-4 md:ml-[72px] md:px-8 md:py-6">
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
