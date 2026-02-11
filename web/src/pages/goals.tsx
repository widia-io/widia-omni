import { useState } from "react";
import { Plus } from "lucide-react";
import { useGoals, useCreateGoal, useUpdateGoal, useDeleteGoal } from "@/hooks/use-goals";
import { useAreas } from "@/hooks/use-areas";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import type { Goal, GoalPeriod } from "@/types/api";

function GoalFormDialog({ goal, onClose }: { goal?: Goal; onClose: () => void }) {
  const { data: areas } = useAreas();
  const create = useCreateGoal();
  const update = useUpdateGoal();
  const [title, setTitle] = useState(goal?.title ?? "");
  const [period, setPeriod] = useState<GoalPeriod>(goal?.period ?? "quarterly");
  const [areaId, setAreaId] = useState(goal?.area_id ?? "");
  const [target, setTarget] = useState(String(goal?.target_value ?? 100));

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const today = new Date().toISOString().slice(0, 10);
    const data = {
      title, period,
      area_id: areaId || null,
      target_value: Number(target),
      start_date: goal?.start_date ?? today,
      end_date: goal?.end_date ?? today,
    };
    if (goal) {
      update.mutate({ id: goal.id, ...data }, { onSuccess: onClose });
    } else {
      create.mutate(data, { onSuccess: onClose });
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label>Título</Label>
        <Input value={title} onChange={(e) => setTitle(e.target.value)} required />
      </div>
      <div className="space-y-2">
        <Label>Período</Label>
        <Select value={period} onValueChange={(v) => setPeriod(v as GoalPeriod)}>
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="weekly">Semanal</SelectItem>
            <SelectItem value="monthly">Mensal</SelectItem>
            <SelectItem value="quarterly">Trimestral</SelectItem>
            <SelectItem value="yearly">Anual</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <Label>Área</Label>
        <Select value={areaId ?? ""} onValueChange={setAreaId}>
          <SelectTrigger><SelectValue placeholder="Selecione" /></SelectTrigger>
          <SelectContent>
            {(areas ?? []).map((a) => <SelectItem key={a.id} value={a.id}>{a.name}</SelectItem>)}
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <Label>Meta (valor)</Label>
        <Input type="number" value={target} onChange={(e) => setTarget(e.target.value)} />
      </div>
      <Button type="submit" className="w-full" disabled={create.isPending || update.isPending}>
        {goal ? "Salvar" : "Criar meta"}
      </Button>
    </form>
  );
}

const statusBadge: Record<string, { v: "green" | "orange" | "rose"; label: string }> = {
  on_track: { v: "green", label: "On Track" },
  at_risk: { v: "orange", label: "At Risk" },
  behind: { v: "rose", label: "Behind" },
  completed: { v: "green", label: "Concluído" },
  not_started: { v: "default" as "green", label: "Não Iniciado" },
};

export function Component() {
  const [filter, setFilter] = useState<Record<string, string>>({});
  const { data: goals, isLoading } = useGoals(filter);
  const deleteGoal = useDeleteGoal();
  const [editGoal, setEditGoal] = useState<Goal | undefined>();
  const [open, setOpen] = useState(false);

  if (isLoading) return <div className="space-y-3">{[1,2,3].map(i => <Skeleton key={i} className="h-20 rounded-[14px]" />)}</div>;

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold">Metas</h1>
        <div className="flex gap-2">
          <Select value={filter.period ?? ""} onValueChange={(v) => setFilter(p => ({ ...p, period: v }))}>
            <SelectTrigger className="w-36"><SelectValue placeholder="Período" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="weekly">Semanal</SelectItem>
              <SelectItem value="monthly">Mensal</SelectItem>
              <SelectItem value="quarterly">Trimestral</SelectItem>
              <SelectItem value="yearly">Anual</SelectItem>
            </SelectContent>
          </Select>
          <Dialog open={open} onOpenChange={(v) => { setOpen(v); if (!v) setEditGoal(undefined); }}>
            <DialogTrigger asChild>
              <Button size="sm"><Plus className="h-4 w-4" /> Nova meta</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader><DialogTitle>{editGoal ? "Editar meta" : "Nova meta"}</DialogTitle></DialogHeader>
              <GoalFormDialog goal={editGoal} onClose={() => setOpen(false)} />
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <div className="space-y-3">
        {(goals ?? []).map((goal) => {
          const pct = goal.target_value ? Math.round((goal.current_value / goal.target_value) * 100) : 0;
          const badge = statusBadge[goal.status];
          return (
            <div
              key={goal.id}
              className="rounded-[14px] border border-border bg-bg-card p-5 cursor-pointer hover:border-border-hover transition-colors"
              onClick={() => { setEditGoal(goal); setOpen(true); }}
            >
              <div className="flex items-center justify-between mb-2">
                <span className="font-serif font-medium">{goal.title}</span>
                <div className="flex items-center gap-2">
                  <span className="font-mono text-sm text-accent-orange">{pct}%</span>
                  <button onClick={(e) => { e.stopPropagation(); deleteGoal.mutate(goal.id); }} className="text-xs text-text-muted hover:text-accent-rose">Excluir</button>
                </div>
              </div>
              <Progress value={pct} className="mb-2" />
              <div className="flex gap-2">
                <Badge variant="blue">{goal.period}</Badge>
                {badge && <Badge variant={badge.v}>{badge.label}</Badge>}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
