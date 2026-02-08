import { RouterProvider } from "react-router";
import { QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "sonner";
import { queryClient } from "@/config/query-client";
import { router } from "@/routes";

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      <Toaster
        theme="dark"
        toastOptions={{
          style: {
            background: "#1f1f1c",
            border: "1px solid #2e2e28",
            color: "#faf9f5",
          },
        }}
      />
    </QueryClientProvider>
  );
}
