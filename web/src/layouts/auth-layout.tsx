import { Outlet } from "react-router";

export function AuthLayout() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-bg-primary p-4">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-accent-orange to-accent-sand font-mono text-lg font-bold text-bg-primary">
            MC
          </div>
          <h2 className="font-sans text-xl font-bold text-text-primary">Mission Control</h2>
        </div>
        <Outlet />
      </div>
    </div>
  );
}
