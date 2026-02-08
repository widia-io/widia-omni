import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { Goal } from "@/types/api";

export function useGoals(params?: Record<string, string>) {
  return useQuery({
    queryKey: ["goals", params],
    queryFn: () => api<Goal[]>("/api/v1/goals", { params }),
  });
}

export function useCreateGoal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Goal>) =>
      api<Goal>("/api/v1/goals", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["goals"] }),
  });
}

export function useUpdateGoal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: Partial<Goal> & { id: string }) =>
      api<Goal>(`/api/v1/goals/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["goals"] }),
  });
}

export function useDeleteGoal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/v1/goals/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["goals"] }),
  });
}

export function useUpdateProgress() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, current_value }: { id: string; current_value: number }) =>
      api<Goal>(`/api/v1/goals/${id}/progress`, { method: "PATCH", body: JSON.stringify({ current_value }) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["goals"] }),
  });
}
