import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { DashboardData } from "@/types/api";

export function useDashboard() {
  return useQuery({
    queryKey: ["dashboard"],
    queryFn: () => api<DashboardData>("/api/v1/dashboard"),
  });
}
