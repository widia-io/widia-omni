import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { OnboardingStatus, LifeArea, Goal, Habit, AreaTemplate, GoalSuggestion } from "@/types/api";

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
    mutationFn: (data: Array<Partial<LifeArea>>) =>
      api<LifeArea[]>("/api/v1/onboarding/areas", { method: "POST", body: JSON.stringify({ areas: data }) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["onboarding"] }),
  });
}

export function useOnboardingGoals() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Array<Partial<Goal>>) =>
      api<Goal[]>("/api/v1/onboarding/goals", { method: "POST", body: JSON.stringify({ goals: data }) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["onboarding"] }),
  });
}

export function useOnboardingHabits() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Array<Partial<Habit>>) =>
      api<Habit[]>("/api/v1/onboarding/habits", { method: "POST", body: JSON.stringify({ habits: data }) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["onboarding"] }),
  });
}

export function useOnboardingComplete() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api("/api/v1/onboarding/complete", { method: "POST" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["onboarding"] }),
  });
}
