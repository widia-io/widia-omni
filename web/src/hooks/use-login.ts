import { useMutation } from "@tanstack/react-query";
import { useNavigate } from "react-router";
import { api } from "@/lib/api-client";
import { useAuthStore } from "@/stores/auth-store";
import type { AuthResponse, LoginRequest, UserProfile } from "@/types/api";

const PENDING_INVITE_KEY = "pending_invite_token";

export function useLogin() {
  const navigate = useNavigate();
  const { setTokens, setUser } = useAuthStore();

  return useMutation({
    mutationFn: async (data: LoginRequest) => {
      const auth = await api<AuthResponse>("/auth/login", {
        method: "POST",
        body: JSON.stringify(data),
      });
      setTokens(auth.access_token, auth.refresh_token);

      const pendingInviteToken = sessionStorage.getItem(PENDING_INVITE_KEY);
      if (pendingInviteToken) {
        await api<{ status: string; workspace_id: string }>("/api/v1/workspace/invites/accept", {
          method: "POST",
          body: JSON.stringify({ token: pendingInviteToken }),
        });
        sessionStorage.removeItem(PENDING_INVITE_KEY);
      }

      const profile = await api<UserProfile>("/api/v1/me");
      setUser(profile);
      return profile;
    },
    onSuccess: (profile) => {
      if (!profile.onboarding_completed) {
        navigate("/onboarding", { replace: true });
      } else {
        navigate("/dashboard", { replace: true });
      }
    },
  });
}
