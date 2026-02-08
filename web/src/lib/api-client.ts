import { env } from "@/config/env";
import { useAuthStore } from "@/stores/auth-store";

class ApiError extends Error {
  constructor(
    public status: number,
    public body: { error: string },
  ) {
    super(body.error);
    this.name = "ApiError";
  }
}

async function refreshToken(): Promise<boolean> {
  const { refreshToken: rt, setTokens, logout } = useAuthStore.getState();
  if (!rt) return false;

  try {
    const res = await fetch(`${env.apiUrl}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: rt }),
    });
    if (!res.ok) {
      logout();
      return false;
    }
    const data = await res.json();
    setTokens(data.access_token, data.refresh_token);
    return true;
  } catch {
    logout();
    return false;
  }
}

export async function api<T>(
  path: string,
  options: RequestInit & { params?: Record<string, string> } = {},
): Promise<T> {
  const { accessToken } = useAuthStore.getState();
  const { params, ...init } = options;

  let url = `${env.apiUrl}${path}`;
  if (params) {
    const search = new URLSearchParams(
      Object.entries(params).filter(([, v]) => v !== undefined && v !== ""),
    );
    if (search.toString()) url += `?${search}`;
  }

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init.headers as Record<string, string>),
  };

  if (accessToken) {
    headers["Authorization"] = `Bearer ${accessToken}`;
  }

  let res = await fetch(url, { ...init, headers });

  if (res.status === 401 && accessToken) {
    const refreshed = await refreshToken();
    if (refreshed) {
      const newToken = useAuthStore.getState().accessToken;
      headers["Authorization"] = `Bearer ${newToken}`;
      res = await fetch(url, { ...init, headers });
    }
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, body);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}
