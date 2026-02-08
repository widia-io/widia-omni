import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { Habit, HabitEntry, HabitStreak } from "@/types/api";

export function useHabits() {
  return useQuery({
    queryKey: ["habits"],
    queryFn: () => api<Habit[]>("/api/v1/habits"),
  });
}

export function useHabitEntries(from: string, to: string) {
  return useQuery({
    queryKey: ["habits", "entries", from, to],
    queryFn: () => api<HabitEntry[]>("/api/v1/habits/entries", { params: { from, to } }),
    enabled: !!from && !!to,
  });
}

export function useHabitStreaks() {
  return useQuery({
    queryKey: ["habits", "streaks"],
    queryFn: () => api<HabitStreak[]>("/api/v1/habits/streaks"),
  });
}

export function useCheckIn() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string; date: string; intensity: number; notes?: string }) =>
      api<HabitEntry>(`/api/v1/habits/${id}/check-in`, { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["habits"] }),
  });
}

export function useCreateHabit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Habit>) =>
      api<Habit>("/api/v1/habits", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["habits"] }),
  });
}

export function useUpdateHabit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: Partial<Habit> & { id: string }) =>
      api<Habit>(`/api/v1/habits/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["habits"] }),
  });
}

export function useDeleteHabit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/v1/habits/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["habits"] }),
  });
}
