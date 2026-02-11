import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { api } from "@/lib/api-client";
import type { FinanceCategory, Transaction, Budget, FinanceSummary } from "@/types/api";

export function useFinanceSummary(month: string, enabled = true) {
  return useQuery({
    queryKey: ["finances", "summary", month],
    queryFn: () => api<FinanceSummary>("/api/v1/finances/summary", { params: { month } }),
    enabled: !!month && enabled,
  });
}

export function useTransactions(params?: Record<string, string>, enabled = true) {
  return useQuery({
    queryKey: ["finances", "transactions", params],
    queryFn: () => api<Transaction[]>("/api/v1/finances/transactions", { params }),
    enabled,
  });
}

export interface CreateTransactionInput {
  category_id?: string;
  area_id?: string;
  type: Transaction["type"];
  amount: number;
  description?: string;
  date: string;
  is_recurring?: boolean;
  recurrence_rule?: string;
  tags?: string[];
}

export interface UpdateTransactionInput extends CreateTransactionInput {
  id: string;
}

export function useCreateTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateTransactionInput) =>
      api<Transaction>("/api/v1/finances/transactions", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["finances"] });
      qc.invalidateQueries({ queryKey: ["workspace", "usage"] });
      toast.success("Transação registrada");
    },
    onError: (err) => {
      const msg = err.message === "monthly transaction limit reached"
        ? "Limite mensal de transações atingido. Faça upgrade para continuar."
        : err.message === "finance not available on your plan"
          ? "Finanças não disponível no seu plano atual."
        : "Erro ao registrar transação";
      toast.error(msg);
    },
  });
}

export function useDeleteTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/v1/finances/transactions/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["finances"] });
      toast.success("Transação removida");
    },
    onError: () => toast.error("Erro ao remover transação"),
  });
}

export function useCategories(enabled = true) {
  return useQuery({
    queryKey: ["finances", "categories"],
    queryFn: () => api<FinanceCategory[]>("/api/v1/finances/categories"),
    enabled,
  });
}

export interface CreateCategoryInput {
  name: string;
  type: FinanceCategory["type"];
  color?: string;
  icon?: string;
  parent_id?: string;
}

export interface UpdateCategoryInput extends Omit<CreateCategoryInput, "type"> {
  id: string;
}

export function useCreateCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateCategoryInput) =>
      api<FinanceCategory>("/api/v1/finances/categories", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["finances", "categories"] });
      toast.success("Categoria criada");
    },
    onError: () => toast.error("Erro ao criar categoria"),
  });
}

export function useDeleteCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/v1/finances/categories/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["finances", "categories"] });
      toast.success("Categoria removida");
    },
    onError: () => toast.error("Erro ao remover categoria"),
  });
}

export function useUpdateTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: UpdateTransactionInput) =>
      api<Transaction>(`/api/v1/finances/transactions/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["finances"] });
      toast.success("Transação atualizada");
    },
    onError: () => toast.error("Erro ao atualizar transação"),
  });
}

export function useUpdateCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: UpdateCategoryInput) =>
      api<FinanceCategory>(`/api/v1/finances/categories/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["finances", "categories"] });
      toast.success("Categoria atualizada");
    },
    onError: () => toast.error("Erro ao atualizar categoria"),
  });
}

export function useBudgets(month: string, enabled = true) {
  return useQuery({
    queryKey: ["finances", "budgets", month],
    queryFn: () => api<Budget[]>("/api/v1/finances/budgets", { params: { month } }),
    enabled: !!month && enabled,
  });
}

export interface UpsertBudgetInput {
  category_id?: string;
  month: string;
  amount: number;
}

export function useUpsertBudget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpsertBudgetInput) =>
      api<Budget>("/api/v1/finances/budgets", { method: "POST", body: JSON.stringify(data) }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["finances"] });
      toast.success("Orçamento salvo");
    },
    onError: () => toast.error("Erro ao salvar orçamento"),
  });
}

export function useDeleteBudget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/v1/finances/budgets/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["finances"] });
      toast.success("Orçamento removido");
    },
    onError: () => toast.error("Erro ao remover orçamento"),
  });
}
