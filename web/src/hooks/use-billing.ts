import { useQuery, useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api-client";
import type { Plan, Subscription } from "@/types/api";

type CheckoutInterval = "monthly" | "yearly";

interface CheckoutPayload {
  tier: Plan["tier"];
  interval: CheckoutInterval;
}

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
    mutationFn: (payload: CheckoutPayload) =>
      api<{ url: string }>("/api/v1/billing/checkout", {
        method: "POST",
        body: JSON.stringify(payload),
      }),
    onSuccess: (data) => {
      window.location.href = data.url;
    },
    onError: (err: Error) => {
      const msg = err.message === "plan not found or no stripe price configured"
        ? "Plano sem preço Stripe configurado para este ciclo."
        : "Não foi possível iniciar o checkout.";
      toast.error(msg);
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
    onError: () => {
      toast.error("Não foi possível abrir o portal de assinatura.");
    },
  });
}
