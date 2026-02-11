import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api-client";
import type { Task, TaskPriority } from "@/types/api";

export function useTasks(params?: Record<string, string>) {
  return useQuery({
    queryKey: ["tasks", params],
    queryFn: () => api<Task[]>("/api/v1/tasks", { params }),
  });
}

export interface CreateTaskInput {
  title: string;
  priority?: TaskPriority;
  description?: string;
  area_id?: string;
  goal_id?: string;
  parent_id?: string;
  section_id?: string;
  due_date?: string;
  duration_minutes?: number;
  is_focus?: boolean;
  label_ids?: string[];
}

export interface UpdateTaskInput extends CreateTaskInput {
  id: string;
}

export function useCreateTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateTaskInput) =>
      api<Task>("/api/v1/tasks", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tasks"] });
      qc.invalidateQueries({ queryKey: ["workspace", "usage"] });
      toast.success("Tarefa criada");
    },
    onError: (err) => {
      const msg = err.message === "daily task limit reached"
        ? "Limite diario de tarefas atingido. Faca upgrade para criar mais."
        : "Erro ao criar tarefa";
      toast.error(msg);
    },
  });
}

export function useUpdateTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: UpdateTaskInput) =>
      api<Task>(`/api/v1/tasks/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tasks"] }),
  });
}

export function useDeleteTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/v1/tasks/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tasks"] });
      qc.invalidateQueries({ queryKey: ["workspace", "usage"] });
      toast.success("Tarefa excluida");
    },
  });
}

export function useCompleteTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<Task>(`/api/v1/tasks/${id}/complete`, { method: "PATCH" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["tasks"] });
      toast.success("Feito!");
    },
  });
}

export function useToggleFocus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<Task>(`/api/v1/tasks/${id}/focus`, { method: "PATCH" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tasks"] }),
  });
}

export function useReopenTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<Task>(`/api/v1/tasks/${id}/reopen`, { method: "PATCH" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tasks"] }),
  });
}

export function useReorderTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, position }: { id: string; position: number }) =>
      api<Task>(`/api/v1/tasks/${id}/reorder`, { method: "PATCH", body: JSON.stringify({ position }) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["tasks"] }),
  });
}
