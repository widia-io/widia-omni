import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { UserProfile, UserPreferences, Workspace, WorkspaceUsage } from "@/types/api";

export function useProfile() {
  return useQuery({
    queryKey: ["me"],
    queryFn: () => api<UserProfile>("/api/v1/me"),
  });
}

export function useUpdateProfile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<UserProfile>) =>
      api<UserProfile>("/api/v1/me", { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["me"] }),
  });
}

export function usePreferences() {
  return useQuery({
    queryKey: ["me", "preferences"],
    queryFn: () => api<UserPreferences>("/api/v1/me/preferences"),
  });
}

export function useUpdatePreferences() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<UserPreferences>) =>
      api<UserPreferences>("/api/v1/me/preferences", { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["me", "preferences"] }),
  });
}

export function useWorkspace() {
  return useQuery({
    queryKey: ["workspace"],
    queryFn: () => api<Workspace>("/api/v1/workspace"),
  });
}

export function useUpdateWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Workspace>) =>
      api<Workspace>("/api/v1/workspace", { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["workspace"] }),
  });
}

export function useWorkspaceUsage() {
  return useQuery({
    queryKey: ["workspace", "usage"],
    queryFn: () => api<WorkspaceUsage>("/api/v1/workspace/usage"),
  });
}

export function useExportData() {
  return useMutation({
    mutationFn: () => api("/api/v1/me/export", { method: "POST" }),
  });
}

export function useDeleteAccount() {
  return useMutation({
    mutationFn: () => api("/api/v1/me", { method: "DELETE" }),
  });
}
