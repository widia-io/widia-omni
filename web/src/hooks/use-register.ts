import { useMutation } from "@tanstack/react-query";
import { useNavigate } from "react-router";
import { api } from "@/lib/api-client";
import { useAuthStore } from "@/stores/auth-store";
import type { AuthResponse, RegisterRequest, UserProfile } from "@/types/api";

export function useRegister() {
  const navigate = useNavigate();
  const { setTokens, setUser } = useAuthStore();

  return useMutation({
    mutationFn: async (data: RegisterRequest) => {
      const auth = await api<AuthResponse>("/auth/register", {
        method: "POST",
        body: JSON.stringify(data),
      });
      setTokens(auth.access_token, auth.refresh_token);

      const profile = await api<UserProfile>("/api/v1/me");
      setUser(profile);
      return profile;
    },
    onSuccess: () => {
      navigate("/onboarding", { replace: true });
    },
  });
}
