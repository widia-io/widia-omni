import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import { useAuthStore } from "@/stores/auth-store";
import type { UserProfile, WorkspaceInvite, WorkspaceInviteWithURL, WorkspaceListItem, WorkspaceMemberSummary, WorkspaceRole } from "@/types/api";

export function useWorkspaces() {
  return useQuery({
    queryKey: ["workspaces"],
    queryFn: () => api<WorkspaceListItem[]>("/api/v1/workspaces"),
  });
}

export function useSwitchWorkspace() {
  const qc = useQueryClient();
  const setUser = useAuthStore((s) => s.setUser);

  return useMutation({
    mutationFn: async (workspaceID: string) =>
      api<{ status: string }>("/api/v1/workspace/switch", {
        method: "POST",
        body: JSON.stringify({ workspace_id: workspaceID }),
      }),
    onSuccess: async () => {
      await qc.invalidateQueries();
      const profile = await api<UserProfile>("/api/v1/me");
      setUser(profile);
    },
  });
}

export function useWorkspaceMembers() {
  return useQuery({
    queryKey: ["workspace", "members"],
    queryFn: () => api<WorkspaceMemberSummary[]>("/api/v1/workspace/members"),
  });
}

export function useRemoveWorkspaceMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (userID: string) =>
      api<void>(`/api/v1/workspace/members/${userID}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["workspace", "members"] });
      qc.invalidateQueries({ queryKey: ["workspace", "usage"] });
      qc.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
}

interface CreateInviteInput {
  email: string;
  role?: WorkspaceRole;
}

export function useWorkspaceInvites(limit = 20, offset = 0) {
  return useQuery({
    queryKey: ["workspace", "invites", limit, offset],
    queryFn: () =>
      api<WorkspaceInvite[]>("/api/v1/workspace/invites", {
        params: {
          limit: String(limit),
          offset: String(offset),
        },
      }),
  });
}

export function useCreateWorkspaceInvite() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateInviteInput) =>
      api<WorkspaceInviteWithURL>("/api/v1/workspace/invites", {
        method: "POST",
        body: JSON.stringify(input),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["workspace", "invites"] });
      qc.invalidateQueries({ queryKey: ["workspace", "usage"] });
    },
  });
}

export function useRevokeWorkspaceInvite() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (inviteID: string) =>
      api<void>(`/api/v1/workspace/invites/${inviteID}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["workspace", "invites"] });
    },
  });
}

export function useAcceptWorkspaceInvite() {
  const qc = useQueryClient();
  const setUser = useAuthStore((s) => s.setUser);

  return useMutation({
    mutationFn: (token: string) =>
      api<{ status: string; workspace_id: string }>("/api/v1/workspace/invites/accept", {
        method: "POST",
        body: JSON.stringify({ token }),
      }),
    onSuccess: async () => {
      await qc.invalidateQueries();
      const profile = await api<UserProfile>("/api/v1/me");
      setUser(profile);
    },
  });
}
