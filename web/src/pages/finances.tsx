import { useState } from "react";
import { format } from "date-fns";
import { Plus } from "lucide-react";
import { useTransactions, useCreateTransaction, useDeleteTransaction, useCategories, useCreateCategory, useDeleteCategory, useFinanceSummary, useBudgets, useUpsertBudget } from "@/hooks/use-finances";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Progress } from "@/components/ui/progress";
import { formatCurrency } from "@/lib/format";
import type { TransactionType } from "@/types/api";

const currentMonth = format(new Date(), "yyyy-MM");

const typeColor: Record<string, "green" | "rose" | "blue" | "sand"> = {
  income: "green", expense: "rose", investment: "blue", transfer: "sand",
};
const typeLabel: Record<string, string> = {
  income: "Receita", expense: "Despesa", investment: "Investimento", transfer: "Transferência",
};

export function Component() {
  const [month] = useState(currentMonth);
  const { data: transactions, isLoading } = useTransactions({ limit: "50" });
  const { data: categories } = useCategories();
  const { data: summary } = useFinanceSummary(month);
  const { data: budgets } = useBudgets(month);
  const createTx = useCreateTransaction();
  const deleteTx = useDeleteTransaction();
  const createCat = useCreateCategory();
  const deleteCat = useDeleteCategory();
  const upsertBudget = useUpsertBudget();

  const [txOpen, setTxOpen] = useState(false);
  const [txTitle, setTxTitle] = useState("");
  const [txAmount, setTxAmount] = useState("");
  const [txType, setTxType] = useState<TransactionType>("expense");

  const [catOpen, setCatOpen] = useState(false);
  const [catName, setCatName] = useState("");
  const [catType, setCatType] = useState<TransactionType>("expense");

  const [budgetCatId, setBudgetCatId] = useState("");
  const [budgetAmount, setBudgetAmount] = useState("");

  if (isLoading) return <Skeleton className="h-96 rounded-[14px]" />;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Finanças</h1>

      <Tabs defaultValue="transactions">
        <TabsList>
          <TabsTrigger value="transactions">Transações</TabsTrigger>
          <TabsTrigger value="categories">Categorias</TabsTrigger>
          <TabsTrigger value="budgets">Orçamentos</TabsTrigger>
          <TabsTrigger value="summary">Resumo</TabsTrigger>
        </TabsList>

        {/* Transactions */}
        <TabsContent value="transactions">
          <div className="mb-4 flex justify-end">
            <Dialog open={txOpen} onOpenChange={setTxOpen}>
              <DialogTrigger asChild>
                <Button size="sm"><Plus className="h-4 w-4" /> Nova transação</Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader><DialogTitle>Nova transação</DialogTitle></DialogHeader>
                <form onSubmit={(e) => { e.preventDefault(); createTx.mutate({ description: txTitle, amount: Number(txAmount), type: txType, date: format(new Date(), "yyyy-MM-dd") }, { onSuccess: () => setTxOpen(false) }); }} className="space-y-4">
                  <div className="space-y-2"><Label>Descrição</Label><Input value={txTitle} onChange={(e) => setTxTitle(e.target.value)} required /></div>
                  <div className="space-y-2"><Label>Valor</Label><Input type="number" step="0.01" value={txAmount} onChange={(e) => setTxAmount(e.target.value)} required /></div>
                  <div className="space-y-2">
                    <Label>Tipo</Label>
                    <Select value={txType} onValueChange={(v) => setTxType(v as TransactionType)}>
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="income">Receita</SelectItem>
                        <SelectItem value="expense">Despesa</SelectItem>
                        <SelectItem value="investment">Investimento</SelectItem>
                        <SelectItem value="transfer">Transferência</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <Button type="submit" className="w-full" disabled={createTx.isPending}>Criar</Button>
                </form>
              </DialogContent>
            </Dialog>
          </div>
          <div className="space-y-2">
            {(transactions ?? []).map((tx) => (
              <div key={tx.id} className="flex items-center gap-3 rounded-[14px] border border-border bg-bg-card px-4 py-3">
                <div className="flex-1">
                  <div className="text-sm">{tx.description ?? "Sem descrição"}</div>
                  <div className="text-xs text-text-muted">{tx.date.slice(0, 10)}</div>
                </div>
                <Badge variant={typeColor[tx.type]}>{typeLabel[tx.type]}</Badge>
                <span className="font-mono text-sm font-semibold">{formatCurrency(tx.amount)}</span>
                <button onClick={() => deleteTx.mutate(tx.id)} className="text-xs text-text-muted hover:text-accent-rose">×</button>
              </div>
            ))}
          </div>
        </TabsContent>

        {/* Categories */}
        <TabsContent value="categories">
          <div className="mb-4 flex justify-end">
            <Dialog open={catOpen} onOpenChange={setCatOpen}>
              <DialogTrigger asChild><Button size="sm"><Plus className="h-4 w-4" /> Nova categoria</Button></DialogTrigger>
              <DialogContent>
                <DialogHeader><DialogTitle>Nova categoria</DialogTitle></DialogHeader>
                <form onSubmit={(e) => { e.preventDefault(); createCat.mutate({ name: catName, type: catType }, { onSuccess: () => setCatOpen(false) }); }} className="space-y-4">
                  <div className="space-y-2"><Label>Nome</Label><Input value={catName} onChange={(e) => setCatName(e.target.value)} required /></div>
                  <div className="space-y-2">
                    <Label>Tipo</Label>
                    <Select value={catType} onValueChange={(v) => setCatType(v as TransactionType)}>
                      <SelectTrigger><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="income">Receita</SelectItem>
                        <SelectItem value="expense">Despesa</SelectItem>
                        <SelectItem value="investment">Investimento</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <Button type="submit" className="w-full" disabled={createCat.isPending}>Criar</Button>
                </form>
              </DialogContent>
            </Dialog>
          </div>
          <div className="grid grid-cols-2 gap-3 lg:grid-cols-3">
            {(categories ?? []).map((cat) => (
              <div key={cat.id} className="rounded-[14px] border border-border bg-bg-card p-4 flex items-center justify-between">
                <div>
                  <div className="font-semibold text-sm">{cat.name}</div>
                  <Badge variant={typeColor[cat.type]} className="mt-1">{typeLabel[cat.type]}</Badge>
                </div>
                {!cat.is_system && <button onClick={() => deleteCat.mutate(cat.id)} className="text-xs text-text-muted hover:text-accent-rose">×</button>}
              </div>
            ))}
          </div>
        </TabsContent>

        {/* Budgets */}
        <TabsContent value="budgets">
          <div className="mb-4 flex gap-2">
            <Select value={budgetCatId} onValueChange={setBudgetCatId}>
              <SelectTrigger className="w-48"><SelectValue placeholder="Categoria" /></SelectTrigger>
              <SelectContent>
                {(categories ?? []).filter(c => c.type === "expense").map(c => <SelectItem key={c.id} value={c.id}>{c.name}</SelectItem>)}
              </SelectContent>
            </Select>
            <Input type="number" placeholder="Valor" className="w-32" value={budgetAmount} onChange={(e) => setBudgetAmount(e.target.value)} />
            <Button size="sm" onClick={() => { if (budgetCatId && budgetAmount) upsertBudget.mutate({ category_id: budgetCatId, month, amount: Number(budgetAmount) }); }}>
              Definir
            </Button>
          </div>
          <div className="space-y-3">
            {(budgets ?? []).map((b) => (
              <div key={b.id} className="rounded-[14px] border border-border bg-bg-card p-4">
                <div className="flex justify-between mb-2">
                  <span className="text-sm font-medium">{b.category_id?.slice(0, 8) ?? "Geral"}</span>
                  <span className="font-mono text-sm">{formatCurrency(b.amount)}</span>
                </div>
                <Progress value={50} />
              </div>
            ))}
          </div>
        </TabsContent>

        {/* Summary */}
        <TabsContent value="summary">
          {summary && (
            <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
              {[
                { label: "Receitas", value: summary.total_income, color: "text-accent-green" },
                { label: "Despesas", value: summary.total_expenses, color: "text-accent-rose" },
                { label: "Investimentos", value: summary.total_investments, color: "text-accent-blue" },
                { label: "Saldo", value: summary.net_balance, color: "text-accent-orange" },
              ].map((s) => (
                <div key={s.label} className="rounded-[14px] border border-border bg-bg-card p-5">
                  <div className="text-xs text-text-muted mb-1">{s.label}</div>
                  <div className={`font-mono text-xl font-bold ${s.color}`}>{formatCurrency(s.value)}</div>
                </div>
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
