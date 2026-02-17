import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type {
  AreaTemplate,
  Goal,
  GoalPeriod,
  GoalSuggestion,
  Habit,
  LifeArea,
  OnboardingStatus,
  Project,
  Task,
} from "@/types/api";

export interface OnboardingAreaInput {
  name: string;
  slug?: string;
  icon: string;
  color: string;
  weight?: number;
  sort_order?: number;
}

export interface OnboardingGoalInput {
  area_id: string;
  title: string;
  period: GoalPeriod;
  start_date?: string;
  end_date?: string;
}

export interface OnboardingHabitInput {
  area_id?: string;
  name: string;
  color?: string;
  frequency?: "daily" | "weekly" | "custom";
  target_per_week?: number;
}

export interface OnboardingProjectInput {
  area_id?: string | null;
  goal_id?: string | null;
  title: string;
  description?: string | null;
  color?: string;
  icon?: string;
  start_date?: string | null;
  target_date?: string | null;
}

export interface OnboardingFirstTaskInput {
  project_id: string;
  title: string;
  description?: string;
  due_date?: string;
}

export function useOnboardingStatus() {
  return useQuery({
    queryKey: ["onboarding"],
    queryFn: () => api<OnboardingStatus>("/api/v1/onboarding/status"),
  });
}

export function useAreaTemplates(locale = "pt-BR") {
  return useQuery({
    queryKey: ["onboarding", "area-templates", locale],
    queryFn: () => api<AreaTemplate[]>(`/api/v1/onboarding/area-templates?locale=${locale}`),
  });
}

export function useGoalSuggestions(locale = "pt-BR", areaSlug?: string) {
  return useQuery({
    queryKey: ["onboarding", "goal-suggestions", locale, areaSlug],
    queryFn: () => api<GoalSuggestion[]>(`/api/v1/onboarding/goal-suggestions?locale=${locale}&area_slug=${areaSlug}`),
    enabled: !!areaSlug,
  });
}

export function useOnboardingAreas() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: OnboardingAreaInput[]) =>
      api<LifeArea[]>("/api/v1/onboarding/areas", { method: "POST", body: JSON.stringify({ areas: data }) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["onboarding"] }),
  });
}

export function useOnboardingGoals() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: OnboardingGoalInput[]) =>
      api<Goal[]>("/api/v1/onboarding/goals", { method: "POST", body: JSON.stringify({ goals: data }) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["onboarding"] });
      qc.invalidateQueries({ queryKey: ["goals"] });
    },
  });
}

export function useOnboardingHabits() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: OnboardingHabitInput[]) =>
      api<Habit[]>("/api/v1/onboarding/habits", { method: "POST", body: JSON.stringify({ habits: data }) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["onboarding"] }),
  });
}

export function useOnboardingSkipHabits() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api<{ status: string }>("/api/v1/onboarding/habits/skip", { method: "POST" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["onboarding"] }),
  });
}

export function useOnboardingProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: OnboardingProjectInput) =>
      api<Project>("/api/v1/onboarding/project", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["onboarding"] });
      qc.invalidateQueries({ queryKey: ["projects"] });
    },
  });
}

export function useOnboardingFirstTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: OnboardingFirstTaskInput) =>
      api<Task>("/api/v1/onboarding/first-task", {
        method: "POST",
        body: JSON.stringify({ ...data, priority: "medium" }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["onboarding"] });
      qc.invalidateQueries({ queryKey: ["tasks"] });
      qc.invalidateQueries({ queryKey: ["projects"] });
    },
  });
}

export function useOnboardingComplete() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api("/api/v1/onboarding/complete", { method: "POST" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["onboarding"] }),
  });
}
