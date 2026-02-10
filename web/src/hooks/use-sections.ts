import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { Section } from "@/types/api";

export function useSections(params?: Record<string, string>) {
  return useQuery({
    queryKey: ["sections", params],
    queryFn: () => api<Section[]>("/api/v1/sections", { params }),
  });
}

export function useCreateSection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { area_id: string; name: string }) =>
      api<Section>("/api/v1/sections", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sections"] }),
  });
}

export function useUpdateSection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string; area_id: string; name: string }) =>
      api<Section>(`/api/v1/sections/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sections"] }),
  });
}

export function useDeleteSection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/v1/sections/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["sections"] });
      qc.invalidateQueries({ queryKey: ["tasks"] });
    },
  });
}

export function useReorderSection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, position }: { id: string; position: number }) =>
      api<Section>(`/api/v1/sections/${id}/reorder`, { method: "PATCH", body: JSON.stringify({ position }) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["sections"] }),
  });
}
