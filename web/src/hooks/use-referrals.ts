import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { ReferralAttribution, ReferralCode, ReferralCredit, ReferralMe } from "@/types/api";

export function useReferralMe(enabled = true) {
  return useQuery({
    queryKey: ["referrals", "me"],
    queryFn: () => api<ReferralMe>("/api/v1/referrals/me"),
    enabled,
  });
}

export function useRegenerateReferralCode() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () =>
      api<ReferralCode>("/api/v1/referrals/regenerate", {
        method: "POST",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["referrals", "me"] });
    },
  });
}

export function useReferralAttributions(limit = 20, offset = 0, enabled = true) {
  return useQuery({
    queryKey: ["referrals", "attributions", limit, offset],
    queryFn: () =>
      api<ReferralAttribution[]>("/api/v1/referrals/attributions", {
        params: { limit: String(limit), offset: String(offset) },
      }),
    enabled,
  });
}

export function useReferralCredits(limit = 20, offset = 0, enabled = true) {
  return useQuery({
    queryKey: ["referrals", "credits", limit, offset],
    queryFn: () =>
      api<ReferralCredit[]>("/api/v1/referrals/credits", {
        params: { limit: String(limit), offset: String(offset) },
      }),
    enabled,
  });
}
