import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { Project, ProjectSection } from "@/types/api";

export function useProjects(params?: Record<string, string>) {
  return useQuery({
    queryKey: ["projects", params],
    queryFn: () => api<Project[]>("/api/v1/projects", { params }),
  });
}

export function useProject(id: string) {
  return useQuery({
    queryKey: ["projects", id],
    queryFn: () => api<Project>(`/api/v1/projects/${id}`),
    enabled: !!id,
  });
}

export function useProjectSections(projectId?: string) {
  return useQuery({
    queryKey: ["projects", projectId, "sections"],
    queryFn: () => api<ProjectSection[]>(`/api/v1/projects/${projectId}/sections`),
    enabled: !!projectId,
  });
}

export function useCreateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Project>) =>
      api<Project>("/api/v1/projects", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["projects"] }),
  });
}

export function useUpdateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: Partial<Project> & { id: string }) =>
      api<Project>(`/api/v1/projects/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["projects"] }),
  });
}

export function useDeleteProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/v1/projects/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["projects"] }),
  });
}

export function useArchiveProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api(`/api/v1/projects/${id}/archive`, { method: "PATCH" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["projects"] }),
  });
}

export function useUnarchiveProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      api(`/api/v1/projects/${id}/unarchive`, { method: "PATCH" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["projects"] }),
  });
}
