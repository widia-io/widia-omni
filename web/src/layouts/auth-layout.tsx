import { Outlet } from "react-router";
import { Logo } from "@/components/logo";

export function AuthLayout() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-bg-primary p-4">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <Logo size={48} className="mx-auto mb-4" />
          <h2 className="font-sans text-xl font-bold text-text-primary">MeuFoco</h2>
        </div>
        <Outlet />
      </div>
    </div>
  );
}
