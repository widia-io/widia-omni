import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { FinanceCategory, Transaction, Budget, FinanceSummary } from "@/types/api";

export function useFinanceSummary(month: string) {
  return useQuery({
    queryKey: ["finances", "summary", month],
    queryFn: () => api<FinanceSummary>("/api/v1/finances/summary", { params: { month } }),
    enabled: !!month,
  });
}

export function useTransactions(params?: Record<string, string>) {
  return useQuery({
    queryKey: ["finances", "transactions", params],
    queryFn: () => api<Transaction[]>("/api/v1/finances/transactions", { params }),
  });
}

export function useCreateTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Transaction>) =>
      api<Transaction>("/api/v1/finances/transactions", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["finances"] }),
  });
}

export function useDeleteTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/v1/finances/transactions/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["finances"] }),
  });
}

export function useCategories() {
  return useQuery({
    queryKey: ["finances", "categories"],
    queryFn: () => api<FinanceCategory[]>("/api/v1/finances/categories"),
  });
}

export function useCreateCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<FinanceCategory>) =>
      api<FinanceCategory>("/api/v1/finances/categories", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["finances", "categories"] }),
  });
}

export function useDeleteCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/v1/finances/categories/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["finances", "categories"] }),
  });
}

export function useBudgets(month: string) {
  return useQuery({
    queryKey: ["finances", "budgets", month],
    queryFn: () => api<Budget[]>("/api/v1/finances/budgets", { params: { month } }),
    enabled: !!month,
  });
}

export function useUpsertBudget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Budget>) =>
      api<Budget>("/api/v1/finances/budgets", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["finances"] }),
  });
}
