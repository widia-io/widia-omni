import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { LifeArea, AreaWithStats, AreaSummary } from "@/types/api";

export function useAreas() {
  return useQuery({
    queryKey: ["areas"],
    queryFn: () => api<AreaWithStats[]>("/api/v1/areas", { params: { include: "stats" } }),
  });
}

export function useAreaSummary(id: string) {
  return useQuery({
    queryKey: ["areas", id, "summary"],
    queryFn: () => api<AreaSummary>(`/api/v1/areas/${id}/summary`),
    enabled: !!id,
  });
}

export function useCreateArea() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<LifeArea>) =>
      api<LifeArea>("/api/v1/areas", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["areas"] }),
  });
}

export function useUpdateArea() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: Partial<LifeArea> & { id: string }) =>
      api<LifeArea>(`/api/v1/areas/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["areas"] }),
  });
}

export function useReorderArea() {
  return useMutation({
    mutationFn: ({ id, sort_order }: { id: string; sort_order: number }) =>
      api<{ status: string }>(`/api/v1/areas/${id}/reorder`, {
        method: "PATCH",
        body: JSON.stringify({ sort_order }),
      }),
  });
}

export function useDeleteArea() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/v1/areas/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["areas"] }),
  });
}
