import { createBrowserRouter } from "react-router";
import { ProtectedRoute } from "./protected-route";
import { AppLayout } from "@/layouts/app-layout";
import { AuthLayout } from "@/layouts/auth-layout";
import { PublicLayout } from "@/layouts/public-layout";
import { NotFoundPage } from "@/pages/not-found";

export const router = createBrowserRouter([
  {
    element: <PublicLayout />,
    children: [
      { index: true, lazy: () => import("@/pages/public/landing") },
    ],
  },
  {
    element: <AuthLayout />,
    children: [
      { path: "login", lazy: () => import("@/pages/auth/login") },
      { path: "register", lazy: () => import("@/pages/auth/register") },
      { path: "forgot-password", lazy: () => import("@/pages/auth/forgot-password") },
      { path: "reset-password", lazy: () => import("@/pages/auth/reset-password") },
    ],
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <AppLayout />,
        children: [
          { path: "dashboard", lazy: () => import("@/pages/dashboard") },
          { path: "areas", lazy: () => import("@/pages/areas") },
          { path: "goals", lazy: () => import("@/pages/goals") },
          { path: "habits", lazy: () => import("@/pages/habits") },
          { path: "tasks", lazy: () => import("@/pages/tasks") },
          { path: "journal", lazy: () => import("@/pages/journal") },
          { path: "finances", lazy: () => import("@/pages/finances") },
          { path: "billing", lazy: () => import("@/pages/billing") },
          { path: "billing/success", lazy: () => import("@/pages/billing-success") },
          { path: "billing/cancel", lazy: () => import("@/pages/billing-cancel") },
          { path: "settings", lazy: () => import("@/pages/settings") },
          { path: "onboarding", lazy: () => import("@/pages/onboarding") },
        ],
      },
    ],
  },
  { path: "*", element: <NotFoundPage /> },
]);
