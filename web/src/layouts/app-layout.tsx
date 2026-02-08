import { Outlet } from "react-router";
import { Sidebar } from "@/components/sidebar";
import { Header } from "@/components/header";

export function AppLayout() {
  return (
    <>
      <Sidebar />
      <main className="ml-[72px] min-h-screen px-8 py-6">
        <Header />
        <Outlet />
      </main>
    </>
  );
}
