import { RouterProvider } from "react-router";
import { QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "sonner";
import { queryClient } from "@/config/query-client";
import { router } from "@/routes";
import { useTheme } from "@/hooks/use-theme";

function ThemeSync() {
  useTheme();
  return null;
}

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeSync />
      <RouterProvider router={router} />
      <Toaster
        theme="system"
        toastOptions={{
          style: {
            background: "var(--card)",
            border: "1px solid var(--border)",
            color: "var(--text)",
          },
        }}
      />
    </QueryClientProvider>
  );
}
