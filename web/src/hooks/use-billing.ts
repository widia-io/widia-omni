import { useQuery, useMutation } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { Plan, Subscription } from "@/types/api";

export function usePlans() {
  return useQuery({
    queryKey: ["billing", "plans"],
    queryFn: () => api<Plan[]>("/api/v1/billing/plans"),
  });
}

export function useSubscription() {
  return useQuery({
    queryKey: ["billing", "subscription"],
    queryFn: () => api<Subscription>("/api/v1/billing/subscription"),
  });
}

export function useCheckout() {
  return useMutation({
    mutationFn: (planId: string) =>
      api<{ url: string }>("/api/v1/billing/checkout", {
        method: "POST",
        body: JSON.stringify({ plan_id: planId }),
      }),
    onSuccess: (data) => {
      window.location.href = data.url;
    },
  });
}

export function usePortal() {
  return useMutation({
    mutationFn: () =>
      api<{ url: string }>("/api/v1/billing/portal", { method: "POST" }),
    onSuccess: (data) => {
      window.location.href = data.url;
    },
  });
}
