import { useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router";
import { addMonths, endOfMonth, format, startOfMonth } from "date-fns";
import { ptBR } from "date-fns/locale";
import {
  ArrowDownRight,
  ArrowUpRight,
  ChevronLeft,
  ChevronRight,
  DollarSign,
  Filter,
  Lock,
  PiggyBank,
  Plus,
  Sparkles,
  Tag,
  Trash2,
  TrendingUp,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import {
  useTransactions,
  useCreateTransaction,
  useDeleteTransaction,
  useCategories,
  useCreateCategory,
  useDeleteCategory,
  useFinanceSummary,
  useBudgets,
  useUpsertBudget,
  useDeleteBudget,
} from "@/hooks/use-finances";
import { useAreas } from "@/hooks/use-areas";
import { useWorkspaceUsage } from "@/hooks/use-settings";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select";
import { Progress } from "@/components/ui/progress";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { formatCurrency, formatPercent } from "@/lib/format";
import { cn } from "@/lib/cn";
import type { FinanceCategory, Transaction, TransactionType, WorkspaceUsage } from "@/types/api";

const currentMonth = format(new Date(), "yyyy-MM");

const TYPE_LABEL: Record<TransactionType, string> = {
  income: "Receita",
  expense: "Despesa",
  investment: "Investimento",
  transfer: "Transferência",
};

const TYPE_BADGE: Record<TransactionType, "green" | "rose" | "blue" | "sand"> = {
  income: "green",
  expense: "rose",
  investment: "blue",
  transfer: "sand",
};

const TYPE_AMOUNT_CLASS: Record<TransactionType, string> = {
  income: "text-accent-green",
  expense: "text-accent-rose",
  investment: "text-accent-blue",
  transfer: "text-accent-sand",
};

const TYPE_ICON: Record<TransactionType, LucideIcon> = {
  income: ArrowUpRight,
  expense: ArrowDownRight,
  investment: TrendingUp,
  transfer: Wallet,
};

const TYPE_ICON_BG: Record<TransactionType, string> = {
  income: "bg-accent-green/12 text-accent-green",
  expense: "bg-accent-rose/12 text-accent-rose",
  investment: "bg-accent-blue/12 text-accent-blue",
  transfer: "bg-accent-sand/12 text-accent-sand",
};

const TYPE_CHIP_ACTIVE: Record<"all" | TransactionType, string> = {
  all: "bg-accent-orange/10 text-accent-orange border-accent-orange/20",
  income: "bg-accent-green/10 text-accent-green border-accent-green/20",
  expense: "bg-accent-rose/10 text-accent-rose border-accent-rose/20",
  investment: "bg-accent-blue/10 text-accent-blue border-accent-blue/20",
  transfer: "bg-accent-sand/10 text-accent-sand border-accent-sand/20",
};

function monthStringToDate(month: string): Date {
  const [yearRaw, monthRaw] = month.split("-");
  const year = Number(yearRaw) || new Date().getFullYear();
  const monthIndex = (Number(monthRaw) || 1) - 1;
  return new Date(year, monthIndex, 1);
}

function signedAmount(tx: Transaction): string {
  if (tx.type === "income") return `+${formatCurrency(tx.amount)}`;
  if (tx.type === "expense" || tx.type === "investment") return `-${formatCurrency(tx.amount)}`;
  return formatCurrency(tx.amount);
}

function monthLabel(month: string): string {
  const value = format(monthStringToDate(month), "MMMM 'de' yyyy", { locale: ptBR });
  return value.charAt(0).toUpperCase() + value.slice(1);
}

function dateHeading(date: string): string {
  const value = format(new Date(`${date}T00:00:00`), "EEEE, dd 'de' MMM", { locale: ptBR });
  return value.charAt(0).toUpperCase() + value.slice(1);
}

function budgetStatus(percentage: number) {
  if (percentage >= 100) {
    return {
      label: "Estourado",
      badge: "rose" as const,
      textClass: "text-accent-rose",
      barClass: "bg-accent-rose",
    };
  }
  if (percentage >= 85) {
    return {
      label: "Atenção",
      badge: "orange" as const,
      textClass: "text-accent-orange",
      barClass: "bg-accent-orange",
    };
  }
  return {
    label: "No controle",
    badge: "green" as const,
    textClass: "text-accent-green",
    barClass: "bg-accent-green",
  };
}

function FilterChip({
  label,
  isActive,
  activeClass,
  onClick,
}: {
  label: string;
  isActive: boolean;
  activeClass: string;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "inline-flex items-center rounded-full border px-3 py-1 text-xs font-medium transition-colors",
        isActive
          ? activeClass
          : "border-border bg-bg-card text-text-muted hover:border-border-hover hover:text-text-secondary",
      )}
    >
      {label}
    </button>
  );
}

function FinanceUsageBadge({ usage }: { usage?: WorkspaceUsage }) {
  if (!usage) return null;

  const used = usage.counters.transactions_month_count;
  const max = usage.limits.max_transactions_per_month;
  if (max === 0) {
    return (
      <div className="rounded-full border border-accent-rose/30 bg-accent-rose/8 px-3 py-1 text-[11px] font-medium text-accent-rose">
        Finanças bloqueado no plano atual
      </div>
    );
  }

  const isUnlimited = max === -1;
  const ratio = isUnlimited ? 0 : max > 0 ? used / max : 0;
  const isFull = !isUnlimited && ratio >= 1;
  const isWarning = !isUnlimited && ratio >= 0.8 && !isFull;

  const accentColor = isFull
    ? "text-accent-rose"
    : isWarning
      ? "text-accent-orange"
      : "text-text-muted";

  const barColor = isFull
    ? "bg-accent-rose"
    : isWarning
      ? "bg-accent-orange"
      : "bg-accent-green";

  return (
    <div className="flex items-center gap-2.5 rounded-full border border-border/60 px-3 py-1">
      <div className="flex items-baseline gap-1">
        <span className={cn("font-mono text-xs font-semibold tabular-nums", accentColor)}>{used}</span>
        <span className="text-[10px] text-text-muted">/</span>
        <span className="font-mono text-[10px] text-text-muted">{isUnlimited ? "∞" : max}</span>
        <span className="text-[10px] text-text-muted">mês</span>
      </div>
      {!isUnlimited && (
        <div className="h-1 w-14 overflow-hidden rounded-full bg-border/50">
          <div
            className={cn("h-full rounded-full transition-all duration-700 ease-out", barColor)}
            style={{ width: `${Math.min(ratio * 100, 100)}%` }}
          />
        </div>
      )}
    </div>
  );
}

export function Component() {
  const [month, setMonth] = useState(currentMonth);
  const [tab, setTab] = useState("transactions");
  const [quickOpen, setQuickOpen] = useState(false);

  const [txTypeFilter, setTxTypeFilter] = useState<"all" | TransactionType>("all");
  const [txCategoryFilter, setTxCategoryFilter] = useState("all");
  const [txAreaFilter, setTxAreaFilter] = useState("all");
  const [txTagFilter, setTxTagFilter] = useState("");

  const [txDescription, setTxDescription] = useState("");
  const [txAmount, setTxAmount] = useState("");
  const [txType, setTxType] = useState<TransactionType>("expense");
  const [txCategoryId, setTxCategoryId] = useState("none");
  const [txAreaId, setTxAreaId] = useState("none");
  const [txDate, setTxDate] = useState(format(new Date(), "yyyy-MM-dd"));
  const [txTags, setTxTags] = useState("");
  const quickInputRef = useRef<HTMLInputElement>(null);

  const [quickTxDescription, setQuickTxDescription] = useState("");
  const [quickTxAmount, setQuickTxAmount] = useState("");
  const [quickTxType, setQuickTxType] = useState<TransactionType>("expense");
  const [quickTxCategoryId, setQuickTxCategoryId] = useState("none");
  const [quickTxAreaId, setQuickTxAreaId] = useState("none");
  const [quickTxDate, setQuickTxDate] = useState(format(new Date(), "yyyy-MM-dd"));
  const [quickTxTags, setQuickTxTags] = useState("");

  const [catName, setCatName] = useState("");
  const [catType, setCatType] = useState<TransactionType>("expense");

  const [budgetCategoryId, setBudgetCategoryId] = useState("none");
  const [budgetAmount, setBudgetAmount] = useState("");
  const { data: usage, isLoading: isLoadingUsage } = useWorkspaceUsage();

  const shouldShowFinanceGate = Boolean(usage && !usage.limits.finance_enabled);
  const financeQueryEnabled = isLoadingUsage ? false : usage ? usage.limits.finance_enabled : true;

  const monthDate = useMemo(() => monthStringToDate(month), [month]);
  const dateFrom = useMemo(() => format(startOfMonth(monthDate), "yyyy-MM-dd"), [monthDate]);
  const dateTo = useMemo(() => format(endOfMonth(monthDate), "yyyy-MM-dd"), [monthDate]);

  const txParams = useMemo(() => {
    const params: Record<string, string> = {
      limit: "200",
      date_from: dateFrom,
      date_to: dateTo,
    };
    if (txTypeFilter !== "all") params.type = txTypeFilter;
    if (txCategoryFilter !== "all") params.category_id = txCategoryFilter;
    if (txAreaFilter !== "all") params.area_id = txAreaFilter;
    if (txTagFilter.trim()) params.tag = txTagFilter.trim();
    return params;
  }, [dateFrom, dateTo, txTypeFilter, txCategoryFilter, txAreaFilter, txTagFilter]);

  const { data: transactions, isLoading: isLoadingTransactions } = useTransactions(txParams, financeQueryEnabled);
  const { data: categories, isLoading: isLoadingCategories } = useCategories(financeQueryEnabled);
  const { data: areas } = useAreas();
  const { data: summary, isLoading: isLoadingSummary } = useFinanceSummary(month, financeQueryEnabled);
  const { data: budgets, isLoading: isLoadingBudgets } = useBudgets(month, financeQueryEnabled);

  const createTx = useCreateTransaction();
  const deleteTx = useDeleteTransaction();
  const createCat = useCreateCategory();
  const deleteCat = useDeleteCategory();
  const upsertBudget = useUpsertBudget();
  const deleteBudget = useDeleteBudget();

  const txList = transactions ?? [];
  const categoriesList = categories ?? [];
  const areasList = areas ?? [];
  const budgetsList = budgets ?? [];

  const categoryById = useMemo(() => {
    return new Map(categoriesList.map((cat) => [cat.id, cat]));
  }, [categoriesList]);

  const areaById = useMemo(() => {
    return new Map(areasList.map((area) => [area.id, area]));
  }, [areasList]);

  const categoryTotals = useMemo(() => {
    const totals = new Map<string, number>();
    for (const row of summary?.by_category ?? []) {
      if (!row.category_id) continue;
      totals.set(row.category_id, (totals.get(row.category_id) ?? 0) + row.total);
    }
    return totals;
  }, [summary]);

  const groupedCategories = useMemo<Record<TransactionType, FinanceCategory[]>>(
    () => ({
      income: categoriesList.filter((cat) => cat.type === "income"),
      expense: categoriesList.filter((cat) => cat.type === "expense"),
      investment: categoriesList.filter((cat) => cat.type === "investment"),
      transfer: categoriesList.filter((cat) => cat.type === "transfer"),
    }),
    [categoriesList],
  );

  const txCategories = useMemo(
    () => categoriesList.filter((cat) => cat.type === txType),
    [categoriesList, txType],
  );

  const quickTxCategories = useMemo(
    () => categoriesList.filter((cat) => cat.type === quickTxType),
    [categoriesList, quickTxType],
  );

  const expenseCategories = useMemo(
    () => categoriesList.filter((cat) => cat.type === "expense"),
    [categoriesList],
  );

  const groupedTransactions = useMemo(() => {
    const groups: Array<{ date: string; items: Transaction[] }> = [];
    let currentDate = "";
    let currentItems: Transaction[] = [];

    for (const tx of txList) {
      const txDateKey = tx.date.slice(0, 10);
      if (txDateKey !== currentDate) {
        if (currentItems.length > 0 && currentDate) {
          groups.push({ date: currentDate, items: currentItems });
        }
        currentDate = txDateKey;
        currentItems = [tx];
      } else {
        currentItems.push(tx);
      }
    }

    if (currentItems.length > 0 && currentDate) {
      groups.push({ date: currentDate, items: currentItems });
    }

    return groups;
  }, [txList]);

  const budgetIdByKey = useMemo(() => {
    return new Map(budgetsList.map((budget) => [budget.category_id ?? "none", budget.id]));
  }, [budgetsList]);

  const budgetRows = useMemo(() => {
    const rows = (summary?.budget_status ?? []).map((status) => {
      const key = status.category_id ?? "none";
      return {
        key,
        budgetId: budgetIdByKey.get(key),
        categoryName:
          status.category_name
          ?? (status.category_id ? categoryById.get(status.category_id)?.name : "Geral")
          ?? "Geral",
        budgetAmount: status.budget_amount,
        spentAmount: status.spent_amount,
        remaining: status.remaining,
        percentage: status.percentage,
      };
    });

    for (const budget of budgetsList) {
      const key = budget.category_id ?? "none";
      if (rows.some((row) => row.key === key)) continue;

      rows.push({
        key,
        budgetId: budget.id,
        categoryName: budget.category_id ? categoryById.get(budget.category_id)?.name ?? "Sem categoria" : "Geral",
        budgetAmount: budget.amount,
        spentAmount: 0,
        remaining: budget.amount,
        percentage: 0,
      });
    }

    return rows.sort((a, b) => b.percentage - a.percentage);
  }, [summary, budgetIdByKey, budgetsList, categoryById]);

  const topCategories = useMemo(() => {
    return (summary?.by_category ?? []).slice(0, 8);
  }, [summary]);

  const hasTxFilters =
    txTypeFilter !== "all"
    || txCategoryFilter !== "all"
    || txAreaFilter !== "all"
    || txTagFilter.trim().length > 0;

  const recurringCount = useMemo(() => txList.filter((tx) => tx.is_recurring).length, [txList]);

  const tagCount = useMemo(() => {
    const tags = new Set<string>();
    for (const tx of txList) {
      for (const tag of tx.tags ?? []) tags.add(tag);
    }
    return tags.size;
  }, [txList]);

  const savingsRate = useMemo(() => {
    if (!summary || summary.total_income <= 0) return 0;
    return (summary.net_balance / summary.total_income) * 100;
  }, [summary]);

  const insights = useMemo(() => {
    const items: string[] = [];

    if (!summary) {
      return ["Registre transações para gerar insights automáticos de fluxo financeiro."];
    }

    if (summary.net_balance >= 0) {
      items.push(`Saldo positivo de ${formatCurrency(summary.net_balance)} no período.`);
    } else {
      items.push(`Saldo negativo de ${formatCurrency(Math.abs(summary.net_balance))}; reduzir despesas é prioridade.`);
    }

    const topExpense = [...(summary.by_category ?? [])]
      .filter((item) => item.type === "expense")
      .sort((a, b) => b.total - a.total)[0];

    if (topExpense) {
      items.push(`Maior centro de custo: ${topExpense.category_name ?? "Sem categoria"} (${formatCurrency(topExpense.total)}).`);
    }

    const overBudgetCount = budgetRows.filter((row) => row.percentage >= 100).length;
    if (overBudgetCount > 0) {
      items.push(`${overBudgetCount} orçamento(s) estourado(s) neste mês.`);
    } else if (budgetRows.length > 0) {
      items.push("Nenhum orçamento estourado até agora.");
    }

    if (summary.total_income > 0) {
      items.push(`Taxa de poupança atual: ${formatPercent(savingsRate)}.`);
    }

    return items;
  }, [summary, budgetRows, savingsRate]);

  const overviewCards = [
    {
      label: "Receitas",
      icon: ArrowUpRight,
      value: summary?.total_income ?? 0,
      valueClass: "text-accent-green",
      iconClass: "text-accent-green",
    },
    {
      label: "Despesas",
      icon: ArrowDownRight,
      value: summary?.total_expenses ?? 0,
      valueClass: "text-accent-rose",
      iconClass: "text-accent-rose",
    },
    {
      label: "Investimentos",
      icon: TrendingUp,
      value: summary?.total_investments ?? 0,
      valueClass: "text-accent-blue",
      iconClass: "text-accent-blue",
    },
    {
      label: "Saldo",
      icon: Wallet,
      value: summary?.net_balance ?? 0,
      valueClass: (summary?.net_balance ?? 0) >= 0 ? "text-accent-green" : "text-accent-rose",
      iconClass: (summary?.net_balance ?? 0) >= 0 ? "text-accent-green" : "text-accent-rose",
    },
  ] as const;

  function shiftMonth(delta: -1 | 1) {
    setMonth(format(addMonths(monthDate, delta), "yyyy-MM"));
  }

  function resetQuickForm() {
    setQuickTxDescription("");
    setQuickTxAmount("");
    setQuickTxType("expense");
    setQuickTxCategoryId("none");
    setQuickTxAreaId("none");
    setQuickTxDate(format(new Date(), "yyyy-MM-dd"));
    setQuickTxTags("");
  }

  function resetFilters() {
    setTxTypeFilter("all");
    setTxCategoryFilter("all");
    setTxAreaFilter("all");
    setTxTagFilter("");
  }

  function parseTags(value: string): string[] {
    return value
      .split(",")
      .map((part) => part.trim())
      .filter(Boolean);
  }

  function handleCreateTransaction(e: React.FormEvent) {
    e.preventDefault();

    const parsedAmount = Number(txAmount.replace(",", "."));
    if (!txDescription.trim() || !Number.isFinite(parsedAmount) || parsedAmount <= 0) return;

    const tags = parseTags(txTags);

    createTx.mutate(
      {
        description: txDescription.trim(),
        amount: parsedAmount,
        type: txType,
        date: txDate,
        ...(txCategoryId !== "none" && { category_id: txCategoryId }),
        ...(txAreaId !== "none" && { area_id: txAreaId }),
        ...(tags.length > 0 && { tags }),
      },
      {
        onSuccess: () => {
          setTxDescription("");
          setTxAmount("");
          setTxCategoryId("none");
          setTxAreaId("none");
          setTxTags("");
        },
      },
    );
  }

  function handleCreateQuickTransaction(e: React.FormEvent) {
    e.preventDefault();

    const parsedAmount = Number(quickTxAmount.replace(",", "."));
    if (!quickTxDescription.trim() || !Number.isFinite(parsedAmount) || parsedAmount <= 0) return;

    const tags = parseTags(quickTxTags);

    createTx.mutate(
      {
        description: quickTxDescription.trim(),
        amount: parsedAmount,
        type: quickTxType,
        date: quickTxDate,
        ...(quickTxCategoryId !== "none" && { category_id: quickTxCategoryId }),
        ...(quickTxAreaId !== "none" && { area_id: quickTxAreaId }),
        ...(tags.length > 0 && { tags }),
      },
      {
        onSuccess: () => {
          resetQuickForm();
          setQuickOpen(false);
        },
      },
    );
  }

  function handleCreateCategory(e: React.FormEvent) {
    e.preventDefault();
    if (!catName.trim()) return;

    createCat.mutate(
      {
        name: catName.trim(),
        type: catType,
      },
      {
        onSuccess: () => setCatName(""),
      },
    );
  }

  function handleUpsertBudget(e: React.FormEvent) {
    e.preventDefault();

    const parsedAmount = Number(budgetAmount.replace(",", "."));
    if (!Number.isFinite(parsedAmount) || parsedAmount <= 0) return;

    upsertBudget.mutate(
      {
        month,
        amount: parsedAmount,
        ...(budgetCategoryId !== "none" && { category_id: budgetCategoryId }),
      },
      {
        onSuccess: () => {
          setBudgetAmount("");
          setBudgetCategoryId("none");
        },
      },
    );
  }

  useEffect(() => {
    if (!quickOpen) return;
    const timeout = window.setTimeout(() => quickInputRef.current?.focus(), 10);
    return () => window.clearTimeout(timeout);
  }, [quickOpen]);

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.defaultPrevented) return;
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      if (shouldShowFinanceGate) return;

      const target = e.target as HTMLElement | null;
      const tag = target?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      if (target?.isContentEditable) return;

      if (e.key === "f" || e.key === "F") {
        e.preventDefault();
        setQuickOpen(true);
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [shouldShowFinanceGate]);

  useEffect(() => {
    if (shouldShowFinanceGate && quickOpen) {
      setQuickOpen(false);
      resetQuickForm();
    }
  }, [quickOpen, shouldShowFinanceGate]);

  if (isLoadingUsage) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-semibold text-text-primary">Finanças</h1>
        </div>
        <Skeleton className="h-52 rounded-[14px]" />
        <Skeleton className="h-80 rounded-[14px]" />
      </div>
    );
  }

  if (shouldShowFinanceGate && usage) {
    const txLimit = usage.limits.max_transactions_per_month;

    return (
      <div className="space-y-5">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <h1 className="text-xl font-semibold text-text-primary">Finanças</h1>
            <FinanceUsageBadge usage={usage} />
          </div>
        </div>

        <section className="rounded-[14px] border border-accent-rose/25 bg-gradient-to-br from-accent-rose-soft via-bg-card to-accent-orange-soft p-6">
          <div className="max-w-2xl">
            <div className="mb-3 inline-flex h-10 w-10 items-center justify-center rounded-full bg-accent-rose/14 text-accent-rose">
              <Lock className="h-5 w-5" />
            </div>

            <h2 className="text-lg font-semibold text-text-primary">Finanças indisponível no seu plano atual</h2>
            <p className="mt-2 text-sm text-text-secondary">
              Este workspace está sem acesso ao módulo financeiro. Faça upgrade para liberar transações, categorias,
              orçamentos e análises automáticas.
            </p>

            <div className="mt-3 rounded-xl border border-border/70 bg-bg-card/80 px-3 py-2 text-xs text-text-secondary">
              Limite atual de transações: <span className="font-mono text-text-primary">{txLimit}</span> por mês
            </div>

            <div className="mt-4 flex flex-wrap items-center gap-2">
              <Button asChild>
                <Link to="/billing">
                  <Sparkles className="h-4 w-4" />
                  Ver planos e liberar finanças
                </Link>
              </Button>
              <Button asChild variant="outline">
                <Link to="/dashboard">Voltar ao dashboard</Link>
              </Button>
            </div>
          </div>
        </section>
      </div>
    );
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-semibold text-text-primary">Finanças</h1>
          <FinanceUsageBadge usage={usage} />
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <button
            onClick={() => setQuickOpen(true)}
            className="inline-flex items-center gap-2 rounded-full border border-border bg-bg-card px-3 py-1.5 text-xs font-medium text-text-secondary transition-colors hover:border-border-hover hover:text-text-primary"
          >
            <Plus className="h-3.5 w-3.5 text-accent-green" />
            Lançamento rápido
            <span className="inline-flex items-center rounded border border-border bg-bg-secondary px-1.5 py-0.5 font-mono text-[10px] text-text-muted">
              F
            </span>
          </button>

          <div className="flex items-center gap-1 rounded-full border border-border bg-bg-card p-1">
            <button
              onClick={() => shiftMonth(-1)}
              className="rounded-full p-1 text-text-muted transition-colors hover:bg-bg-secondary hover:text-text-primary"
              aria-label="Mês anterior"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <span className="min-w-[145px] text-center text-sm font-medium capitalize text-text-primary">
              {monthLabel(month)}
            </span>
            <button
              onClick={() => shiftMonth(1)}
              className="rounded-full p-1 text-text-muted transition-colors hover:bg-bg-secondary hover:text-text-primary"
              aria-label="Próximo mês"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>

      <section className="rounded-[14px] border border-accent-green/20 bg-gradient-to-br from-accent-green-soft via-bg-card to-accent-sage-soft p-4 sm:p-5">
        <div className="mb-4 flex flex-wrap items-start justify-between gap-3">
          <div>
            <p className="text-[10px] font-semibold uppercase tracking-wider text-text-muted">Visão Mensal</p>
            <h2 className="mt-1 text-lg font-semibold text-text-primary">Resumo de {monthLabel(month)}</h2>
            <p className="mt-1 text-xs text-text-secondary">
              Fluxo consolidado de receitas, despesas e investimentos do período.
            </p>
          </div>

          {summary && (
            <Badge
              variant={summary.net_balance >= 0 ? "green" : "rose"}
              className="rounded-full px-3 py-1 text-xs"
            >
              {summary.net_balance >= 0 ? "Mês positivo" : "Mês em alerta"}
            </Badge>
          )}
        </div>

        <div className="grid grid-cols-2 gap-2 lg:grid-cols-4">
          {isLoadingSummary && !summary && [1, 2, 3, 4].map((i) => (
            <Skeleton key={i} className="h-20 rounded-xl" />
          ))}

          {!isLoadingSummary && overviewCards.map((card) => {
            const Icon = card.icon;
            return (
              <div key={card.label} className="rounded-xl border border-border/70 bg-bg-card/90 px-4 py-3">
                <div className="mb-1.5 flex items-center justify-between">
                  <span className="text-xs text-text-muted">{card.label}</span>
                  <Icon className={cn("h-4 w-4", card.iconClass)} />
                </div>
                <div className={cn("font-mono text-base font-semibold", card.valueClass)}>
                  {formatCurrency(card.value)}
                </div>
              </div>
            );
          })}
        </div>
      </section>

      <Tabs value={tab} onValueChange={setTab}>
        <TabsList className="h-auto w-full justify-start gap-1 overflow-x-auto p-1.5">
          <TabsTrigger value="transactions" className="text-xs sm:text-sm">Transações</TabsTrigger>
          <TabsTrigger value="categories" className="text-xs sm:text-sm">Categorias</TabsTrigger>
          <TabsTrigger value="budgets" className="text-xs sm:text-sm">Orçamentos</TabsTrigger>
          <TabsTrigger value="summary" className="text-xs sm:text-sm">Resumo</TabsTrigger>
        </TabsList>

        <TabsContent value="transactions" className="space-y-4">
          <div className="rounded-[14px] border border-border bg-bg-card p-4">
            <div className="mb-3 flex items-center gap-2">
              <Plus className="h-4 w-4 text-accent-green" />
              <h3 className="text-sm font-semibold text-text-primary">Lançamento rápido</h3>
            </div>

            <form onSubmit={handleCreateTransaction} className="space-y-2.5">
              <div className="grid gap-2 lg:grid-cols-[1.8fr_0.95fr_0.9fr_1fr_1fr_auto]">
                <Input
                  value={txDescription}
                  onChange={(e) => setTxDescription(e.target.value)}
                  placeholder="Ex.: Assinatura de ferramenta"
                  required
                />

                <Input
                  value={txAmount}
                  onChange={(e) => setTxAmount(e.target.value)}
                  type="number"
                  step="0.01"
                  min="0"
                  placeholder="Valor"
                  required
                />

                <Select
                  value={txType}
                  onValueChange={(value) => {
                    setTxType(value as TransactionType);
                    setTxCategoryId("none");
                  }}
                >
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="income">Receita</SelectItem>
                    <SelectItem value="expense">Despesa</SelectItem>
                    <SelectItem value="investment">Investimento</SelectItem>
                    <SelectItem value="transfer">Transferência</SelectItem>
                  </SelectContent>
                </Select>

                <Select value={txCategoryId} onValueChange={setTxCategoryId}>
                  <SelectTrigger><SelectValue placeholder="Categoria" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none">Sem categoria</SelectItem>
                    {txCategories.map((cat) => (
                      <SelectItem key={cat.id} value={cat.id}>{cat.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Input type="date" value={txDate} onChange={(e) => setTxDate(e.target.value)} />

                <Button
                  type="submit"
                  disabled={createTx.isPending || !txDescription.trim() || !txAmount.trim()}
                  className="h-10"
                >
                  <Plus className="h-3.5 w-3.5" />
                  Registrar
                </Button>
              </div>

              <div className="grid gap-2 lg:grid-cols-[1fr_2fr]">
                <Select value={txAreaId} onValueChange={setTxAreaId}>
                  <SelectTrigger><SelectValue placeholder="Área" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none">Sem área</SelectItem>
                    {areasList.map((area) => (
                      <SelectItem key={area.id} value={area.id}>{area.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Input
                  value={txTags}
                  onChange={(e) => setTxTags(e.target.value)}
                  placeholder="Tags separadas por vírgula (ex.: fixo, software, pessoal)"
                />
              </div>
            </form>
          </div>

          <div className="rounded-[14px] border border-border bg-bg-card p-3">
            <div className="flex flex-wrap items-center gap-2">
              <span className="inline-flex items-center gap-1.5 text-[11px] font-semibold uppercase tracking-wider text-text-muted">
                <Filter className="h-3.5 w-3.5" />
                Filtros
              </span>

              <FilterChip
                label="Todas"
                isActive={txTypeFilter === "all"}
                activeClass={TYPE_CHIP_ACTIVE.all}
                onClick={() => setTxTypeFilter("all")}
              />
              <FilterChip
                label="Receitas"
                isActive={txTypeFilter === "income"}
                activeClass={TYPE_CHIP_ACTIVE.income}
                onClick={() => setTxTypeFilter("income")}
              />
              <FilterChip
                label="Despesas"
                isActive={txTypeFilter === "expense"}
                activeClass={TYPE_CHIP_ACTIVE.expense}
                onClick={() => setTxTypeFilter("expense")}
              />
              <FilterChip
                label="Investimentos"
                isActive={txTypeFilter === "investment"}
                activeClass={TYPE_CHIP_ACTIVE.investment}
                onClick={() => setTxTypeFilter("investment")}
              />
              <FilterChip
                label="Transferências"
                isActive={txTypeFilter === "transfer"}
                activeClass={TYPE_CHIP_ACTIVE.transfer}
                onClick={() => setTxTypeFilter("transfer")}
              />

              <div className="min-w-[180px] flex-1 lg:flex-none">
                <Select value={txCategoryFilter} onValueChange={setTxCategoryFilter}>
                  <SelectTrigger><SelectValue placeholder="Categoria" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Todas as categorias</SelectItem>
                    {categoriesList.map((cat) => (
                      <SelectItem key={cat.id} value={cat.id}>{cat.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="min-w-[160px] flex-1 lg:flex-none">
                <Select value={txAreaFilter} onValueChange={setTxAreaFilter}>
                  <SelectTrigger><SelectValue placeholder="Área" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Todas as áreas</SelectItem>
                    {areasList.map((area) => (
                      <SelectItem key={area.id} value={area.id}>{area.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <Input
                value={txTagFilter}
                onChange={(e) => setTxTagFilter(e.target.value)}
                placeholder="Filtrar por tag"
                className="h-9 min-w-[150px] flex-1"
              />

              {hasTxFilters && (
                <button
                  onClick={resetFilters}
                  className="text-xs text-text-muted underline decoration-dotted underline-offset-4 transition-colors hover:text-text-primary"
                >
                  Limpar
                </button>
              )}
            </div>
          </div>

          {isLoadingTransactions && !transactions ? (
            <div className="space-y-2">
              {[1, 2, 3, 4, 5].map((item) => (
                <Skeleton key={item} className="h-16 rounded-[14px]" />
              ))}
            </div>
          ) : groupedTransactions.length === 0 ? (
            <div className="rounded-[14px] border border-border bg-bg-card py-14 text-center">
              <DollarSign className="mx-auto mb-2 h-8 w-8 text-text-muted" />
              <p className="text-sm font-medium text-text-primary">Nenhuma transação no período</p>
              <p className="mt-1 text-xs text-text-muted">Use o lançamento rápido para começar a registrar seu fluxo.</p>
            </div>
          ) : (
            <div className="overflow-hidden rounded-[14px] border border-border bg-bg-card">
              {groupedTransactions.map((group) => (
                <div key={group.date} className="border-b border-border/60 last:border-b-0">
                  <div className="flex items-center justify-between border-b border-border/40 bg-bg-secondary/60 px-4 py-2">
                    <span className="text-[11px] font-semibold uppercase tracking-wider text-text-secondary">
                      {dateHeading(group.date)}
                    </span>
                    <span className="font-mono text-[11px] text-text-muted">{group.items.length}</span>
                  </div>

                  <div>
                    {group.items.map((tx) => {
                      const Icon = TYPE_ICON[tx.type];
                      const category = tx.category_id ? categoryById.get(tx.category_id) : undefined;
                      const area = tx.area_id ? areaById.get(tx.area_id) : undefined;
                      const tags = tx.tags ?? [];

                      return (
                        <div
                          key={tx.id}
                          className="group flex items-start gap-3 border-b border-border/40 px-4 py-3 transition-colors last:border-b-0 hover:bg-bg-secondary/40"
                        >
                          <div
                            className={cn(
                              "mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full",
                              TYPE_ICON_BG[tx.type],
                            )}
                          >
                            <Icon className="h-4 w-4" />
                          </div>

                          <div className="min-w-0 flex-1">
                            <div className="truncate text-sm font-medium text-text-primary">
                              {tx.description?.trim() || "Sem descrição"}
                            </div>

                            <div className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[11px] text-text-muted">
                              <Badge variant={TYPE_BADGE[tx.type]}>{TYPE_LABEL[tx.type]}</Badge>

                              <span>{category?.name ?? "Sem categoria"}</span>

                              {area && (
                                <span className="inline-flex items-center gap-1">
                                  <span className="h-1.5 w-1.5 rounded-full bg-text-muted/60" />
                                  {area.name}
                                </span>
                              )}

                              {tx.is_recurring && (
                                <span className="inline-flex items-center gap-1 text-accent-blue">
                                  <TrendingUp className="h-3 w-3" />
                                  Recorrente
                                </span>
                              )}

                              {tags.length > 0 && (
                                <span className="inline-flex items-center gap-1">
                                  <Tag className="h-3 w-3" />
                                  {tags.slice(0, 3).join(", ")}
                                </span>
                              )}
                            </div>
                          </div>

                          <div className="flex items-center gap-2">
                            <span className={cn("whitespace-nowrap font-mono text-sm font-semibold", TYPE_AMOUNT_CLASS[tx.type])}>
                              {signedAmount(tx)}
                            </span>
                            <button
                              onClick={() => deleteTx.mutate(tx.id)}
                              className="rounded p-1 text-text-muted opacity-0 transition-opacity group-hover:opacity-100 hover:text-accent-rose"
                              disabled={deleteTx.isPending}
                              title="Excluir"
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </button>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="categories" className="space-y-4">
          <form onSubmit={handleCreateCategory} className="rounded-[14px] border border-border bg-bg-card p-4">
            <div className="mb-3 flex items-center gap-2">
              <Plus className="h-4 w-4 text-accent-blue" />
              <h3 className="text-sm font-semibold text-text-primary">Nova categoria</h3>
            </div>

            <div className="grid gap-2 lg:grid-cols-[1.6fr_1fr_auto]">
              <Input
                value={catName}
                onChange={(e) => setCatName(e.target.value)}
                placeholder="Ex.: Ferramentas, Alimentação, Receita de vendas"
                required
              />
              <Select value={catType} onValueChange={(value) => setCatType(value as TransactionType)}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="income">Receita</SelectItem>
                  <SelectItem value="expense">Despesa</SelectItem>
                  <SelectItem value="investment">Investimento</SelectItem>
                  <SelectItem value="transfer">Transferência</SelectItem>
                </SelectContent>
              </Select>
              <Button type="submit" disabled={createCat.isPending || !catName.trim()}>
                Criar
              </Button>
            </div>
          </form>

          {isLoadingCategories && !categories ? (
            <div className="grid grid-cols-1 gap-3 lg:grid-cols-2">
              {[1, 2, 3, 4].map((i) => (
                <Skeleton key={i} className="h-52 rounded-[14px]" />
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
              {(["income", "expense", "investment", "transfer"] as TransactionType[]).map((type) => {
                const list = groupedCategories[type];
                return (
                  <section key={type} className="rounded-[14px] border border-border bg-bg-card p-4">
                    <div className="mb-3 flex items-center justify-between">
                      <h3 className="text-sm font-semibold text-text-primary">{TYPE_LABEL[type]}</h3>
                      <Badge variant={TYPE_BADGE[type]}>{list.length}</Badge>
                    </div>

                    {list.length === 0 ? (
                      <p className="text-xs text-text-muted">Sem categorias desse tipo.</p>
                    ) : (
                      <div className="space-y-1.5">
                        {list.map((cat) => {
                          const monthTotal = categoryTotals.get(cat.id) ?? 0;
                          return (
                            <div
                              key={cat.id}
                              className="group flex items-center justify-between rounded-lg border border-border/60 px-3 py-2 transition-colors hover:bg-bg-secondary/50"
                            >
                              <div className="min-w-0">
                                <p className="truncate text-sm text-text-primary">{cat.name}</p>
                                <p className="mt-0.5 text-[11px] text-text-muted">
                                  {monthTotal > 0 ? `${formatCurrency(monthTotal)} no mês` : "Sem movimentação no mês"}
                                </p>
                              </div>

                              <div className="flex items-center gap-2">
                                {cat.is_system ? (
                                  <Badge variant="default">Sistema</Badge>
                                ) : (
                                  <button
                                    onClick={() => deleteCat.mutate(cat.id)}
                                    className="rounded p-1 text-text-muted opacity-0 transition-opacity group-hover:opacity-100 hover:text-accent-rose"
                                    disabled={deleteCat.isPending}
                                    title="Excluir"
                                  >
                                    <Trash2 className="h-3.5 w-3.5" />
                                  </button>
                                )}
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </section>
                );
              })}
            </div>
          )}
        </TabsContent>

        <TabsContent value="budgets" className="space-y-4">
          <form onSubmit={handleUpsertBudget} className="rounded-[14px] border border-border bg-bg-card p-4">
            <div className="mb-3 flex items-center gap-2">
              <PiggyBank className="h-4 w-4 text-accent-sand" />
              <h3 className="text-sm font-semibold text-text-primary">Definir orçamento mensal</h3>
            </div>

            <div className="grid gap-2 lg:grid-cols-[1.4fr_1fr_auto]">
              <Select value={budgetCategoryId} onValueChange={setBudgetCategoryId}>
                <SelectTrigger><SelectValue placeholder="Categoria" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">Geral (sem categoria)</SelectItem>
                  {expenseCategories.map((cat) => (
                    <SelectItem key={cat.id} value={cat.id}>{cat.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Input
                value={budgetAmount}
                onChange={(e) => setBudgetAmount(e.target.value)}
                type="number"
                step="0.01"
                min="0"
                placeholder="Valor do orçamento"
                required
              />

              <Button type="submit" disabled={upsertBudget.isPending || !budgetAmount.trim()}>
                Salvar
              </Button>
            </div>
          </form>

          {isLoadingBudgets && isLoadingSummary ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-28 rounded-[14px]" />
              ))}
            </div>
          ) : budgetRows.length === 0 ? (
            <div className="rounded-[14px] border border-border bg-bg-card py-14 text-center">
              <PiggyBank className="mx-auto mb-2 h-8 w-8 text-text-muted" />
              <p className="text-sm font-medium text-text-primary">Nenhum orçamento cadastrado</p>
              <p className="mt-1 text-xs text-text-muted">Defina um orçamento por categoria para acompanhar gastos em tempo real.</p>
            </div>
          ) : (
            <div className="space-y-3">
              {budgetRows.map((row) => {
                const status = budgetStatus(row.percentage);
                const progress = Math.min(Math.max(row.percentage, 0), 100);
                const overage = row.spentAmount - row.budgetAmount;

                return (
                  <div key={row.key} className="rounded-[14px] border border-border bg-bg-card p-4">
                    <div className="mb-2 flex flex-wrap items-start justify-between gap-2">
                      <div>
                        <p className="text-sm font-medium text-text-primary">{row.categoryName}</p>
                        <p className="text-xs text-text-muted">
                          Orçamento: {formatCurrency(row.budgetAmount)} · Gasto: {formatCurrency(row.spentAmount)}
                        </p>
                      </div>

                      <div className="flex items-center gap-2">
                        <Badge variant={status.badge}>{status.label}</Badge>
                        <span className={cn("font-mono text-xs", status.textClass)}>{formatPercent(row.percentage)}</span>
                        {row.budgetId && (
                          <button
                            onClick={() => deleteBudget.mutate(row.budgetId!)}
                            className="rounded p-1 text-text-muted transition-colors hover:text-accent-rose"
                            title="Excluir orçamento"
                            disabled={deleteBudget.isPending}
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </button>
                        )}
                      </div>
                    </div>

                    <Progress value={progress} indicatorClassName={status.barClass} />

                    <div className="mt-2 flex items-center justify-between text-xs text-text-secondary">
                      <span>
                        {overage > 0
                          ? `Excesso: ${formatCurrency(overage)}`
                          : `Restante: ${formatCurrency(row.remaining)}`}
                      </span>
                      <span className="font-mono">{formatPercent(row.percentage)}</span>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </TabsContent>

        <TabsContent value="summary" className="space-y-4">
          <div className="grid grid-cols-2 gap-2 lg:grid-cols-4">
            <div className="rounded-[14px] border border-border bg-bg-card p-4">
              <div className="mb-1 inline-flex rounded-lg bg-accent-blue/12 p-2 text-accent-blue">
                <DollarSign className="h-4 w-4" />
              </div>
              <div className="font-mono text-lg font-semibold text-text-primary">{txList.length}</div>
              <p className="text-xs text-text-muted">Transações no mês</p>
            </div>

            <div className="rounded-[14px] border border-border bg-bg-card p-4">
              <div className="mb-1 inline-flex rounded-lg bg-accent-green/12 p-2 text-accent-green">
                <TrendingUp className="h-4 w-4" />
              </div>
              <div className="font-mono text-lg font-semibold text-text-primary">{recurringCount}</div>
              <p className="text-xs text-text-muted">Recorrentes</p>
            </div>

            <div className="rounded-[14px] border border-border bg-bg-card p-4">
              <div className="mb-1 inline-flex rounded-lg bg-accent-sand/12 p-2 text-accent-sand">
                <Tag className="h-4 w-4" />
              </div>
              <div className="font-mono text-lg font-semibold text-text-primary">{tagCount}</div>
              <p className="text-xs text-text-muted">Tags ativas</p>
            </div>

            <div className="rounded-[14px] border border-border bg-bg-card p-4">
              <div className="mb-1 inline-flex rounded-lg bg-accent-orange/12 p-2 text-accent-orange">
                <Wallet className="h-4 w-4" />
              </div>
              <div
                className={cn(
                  "font-mono text-lg font-semibold",
                  savingsRate >= 0 ? "text-accent-green" : "text-accent-rose",
                )}
              >
                {formatPercent(savingsRate)}
              </div>
              <p className="text-xs text-text-muted">Taxa de poupança</p>
            </div>
          </div>

          <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <section className="rounded-[14px] border border-border bg-bg-card p-4">
              <h3 className="mb-3 text-sm font-semibold text-text-primary">Fluxo por categoria</h3>

              {topCategories.length === 0 ? (
                <p className="text-xs text-text-muted">Sem dados de categoria neste período.</p>
              ) : (
                <div className="space-y-2.5">
                  {topCategories.map((row, idx) => {
                    const maxValue = topCategories[0]?.total ?? 0;
                    const width = maxValue > 0 ? (row.total / maxValue) * 100 : 0;
                    const type = row.type as TransactionType;

                    return (
                      <div key={`${row.category_id ?? "none"}-${idx}`}>
                        <div className="mb-1 flex items-center justify-between gap-2 text-xs">
                          <span className="truncate text-text-secondary">
                            {row.category_name ?? "Sem categoria"}
                          </span>
                          <span className={cn("font-mono", TYPE_AMOUNT_CLASS[type])}>
                            {formatCurrency(row.total)}
                          </span>
                        </div>
                        <Progress
                          value={Math.min(Math.max(width, 0), 100)}
                          indicatorClassName={
                            type === "income"
                              ? "bg-accent-green"
                              : type === "expense"
                                ? "bg-accent-rose"
                                : type === "investment"
                                  ? "bg-accent-blue"
                                  : "bg-accent-sand"
                          }
                          className="h-1.5"
                        />
                      </div>
                    );
                  })}
                </div>
              )}
            </section>

            <section className="rounded-[14px] border border-border bg-bg-card p-4">
              <h3 className="mb-3 text-sm font-semibold text-text-primary">Diagnóstico rápido</h3>
              <div className="space-y-2">
                {insights.map((insight) => (
                  <div
                    key={insight}
                    className="rounded-lg border border-border/60 bg-bg-secondary/50 px-3 py-2 text-xs text-text-secondary"
                  >
                    {insight}
                  </div>
                ))}
              </div>
            </section>
          </div>
        </TabsContent>
      </Tabs>

      <Dialog
        open={quickOpen}
        onOpenChange={(open) => {
          setQuickOpen(open);
          if (!open) resetQuickForm();
        }}
      >
        <DialogContent className="max-w-xl">
          <DialogTitle>Novo lançamento rápido</DialogTitle>

          <form onSubmit={handleCreateQuickTransaction} className="mt-3 space-y-3">
            <Input
              ref={quickInputRef}
              value={quickTxDescription}
              onChange={(e) => setQuickTxDescription(e.target.value)}
              placeholder="Descrição da transação"
              required
            />

            <div className="grid gap-2 sm:grid-cols-2">
              <Input
                value={quickTxAmount}
                onChange={(e) => setQuickTxAmount(e.target.value)}
                type="number"
                step="0.01"
                min="0"
                placeholder="Valor"
                required
              />
              <Input
                type="date"
                value={quickTxDate}
                onChange={(e) => setQuickTxDate(e.target.value)}
              />
            </div>

            <div className="grid gap-2 sm:grid-cols-2">
              <Select
                value={quickTxType}
                onValueChange={(value) => {
                  setQuickTxType(value as TransactionType);
                  setQuickTxCategoryId("none");
                }}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="income">Receita</SelectItem>
                  <SelectItem value="expense">Despesa</SelectItem>
                  <SelectItem value="investment">Investimento</SelectItem>
                  <SelectItem value="transfer">Transferência</SelectItem>
                </SelectContent>
              </Select>

              <Select value={quickTxCategoryId} onValueChange={setQuickTxCategoryId}>
                <SelectTrigger><SelectValue placeholder="Categoria" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">Sem categoria</SelectItem>
                  {quickTxCategories.map((cat) => (
                    <SelectItem key={cat.id} value={cat.id}>{cat.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="grid gap-2 sm:grid-cols-2">
              <Select value={quickTxAreaId} onValueChange={setQuickTxAreaId}>
                <SelectTrigger><SelectValue placeholder="Área" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">Sem área</SelectItem>
                  {areasList.map((area) => (
                    <SelectItem key={area.id} value={area.id}>{area.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Input
                value={quickTxTags}
                onChange={(e) => setQuickTxTags(e.target.value)}
                placeholder="Tags (vírgula)"
              />
            </div>

            <div className="flex items-center justify-between gap-2 pt-1">
              <span className="text-[11px] text-text-muted">
                Atalho global nesta página: <span className="font-mono text-text-secondary">F</span>
              </span>
              <div className="flex items-center gap-2">
                <Button
                  type="button"
                  variant="ghost"
                  onClick={() => {
                    setQuickOpen(false);
                    resetQuickForm();
                  }}
                >
                  Cancelar
                </Button>
                <Button
                  type="submit"
                  disabled={createTx.isPending || !quickTxDescription.trim() || !quickTxAmount.trim()}
                >
                  Registrar
                </Button>
              </div>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
