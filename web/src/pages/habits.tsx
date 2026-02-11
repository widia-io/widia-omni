import { useState } from "react";
import { Plus, Check } from "lucide-react";
import { format } from "date-fns";
import { useHabits, useCreateHabit, useUpdateHabit, useDeleteHabit, useCheckIn, useHabitStreaks } from "@/hooks/use-habits";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { cn } from "@/lib/cn";
import type { Habit, HabitFrequency } from "@/types/api";

function HabitFormDialog({ habit, onClose }: { habit?: Habit; onClose: () => void }) {
  const create = useCreateHabit();
  const update = useUpdateHabit();
  const [name, setName] = useState(habit?.name ?? "");
  const [frequency, setFrequency] = useState<HabitFrequency>(habit?.frequency ?? "daily");
  const [target, setTarget] = useState(String(habit?.target_per_week ?? 5));
  const [color, setColor] = useState(habit?.color ?? "green");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const data = { name, frequency, target_per_week: Number(target), color };
    if (habit) {
      update.mutate({ id: habit.id, ...data }, { onSuccess: onClose });
    } else {
      create.mutate(data, { onSuccess: onClose });
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label>Nome</Label>
        <Input value={name} onChange={(e) => setName(e.target.value)} required />
      </div>
      <div className="space-y-2">
        <Label>Frequência</Label>
        <Select value={frequency} onValueChange={(v) => setFrequency(v as HabitFrequency)}>
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="daily">Diário</SelectItem>
            <SelectItem value="weekly">Semanal</SelectItem>
            <SelectItem value="custom">Personalizado</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <Label>Meta por semana</Label>
        <Input type="number" min="1" max="7" value={target} onChange={(e) => setTarget(e.target.value)} />
      </div>
      <div className="space-y-2">
        <Label>Cor</Label>
        <div className="flex gap-2">
          {["green", "blue", "orange", "sand", "sage", "rose"].map((c) => (
            <button key={c} type="button" onClick={() => setColor(c)} className={cn("h-8 w-8 rounded-lg border-2 transition-colors", color === c ? "border-text-primary" : "border-transparent", `bg-accent-${c}`)} />
          ))}
        </div>
      </div>
      <Button type="submit" className="w-full" disabled={create.isPending || update.isPending}>
        {habit ? "Salvar" : "Criar hábito"}
      </Button>
    </form>
  );
}

export function Component() {
  const { data: habits, isLoading } = useHabits();
  const { data: streaks } = useHabitStreaks();
  const deleteHabit = useDeleteHabit();
  const checkIn = useCheckIn();
  const [editHabit, setEditHabit] = useState<Habit | undefined>();
  const [open, setOpen] = useState(false);

  if (isLoading) return <div className="space-y-3">{[1,2,3].map(i => <Skeleton key={i} className="h-16 rounded-[14px]" />)}</div>;

  const streakMap = new Map((streaks ?? []).map(s => [s.habit_id, s]));
  const today = format(new Date(), "yyyy-MM-dd");

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold">Hábitos</h1>
        <Dialog open={open} onOpenChange={(v) => { setOpen(v); if (!v) setEditHabit(undefined); }}>
          <DialogTrigger asChild>
            <Button size="sm"><Plus className="h-4 w-4" /> Novo hábito</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader><DialogTitle>{editHabit ? "Editar hábito" : "Novo hábito"}</DialogTitle></DialogHeader>
            <HabitFormDialog habit={editHabit} onClose={() => setOpen(false)} />
          </DialogContent>
        </Dialog>
      </div>

      <div className="space-y-3">
        {(habits ?? []).map((habit) => {
          const streak = streakMap.get(habit.id);
          return (
            <div key={habit.id} className="flex items-center gap-4 rounded-[14px] border border-border bg-bg-card p-4">
              <button
                onClick={() => checkIn.mutate({ id: habit.id, date: today, intensity: 3 })}
                className={cn("flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border-2 transition-all", `border-accent-${habit.color} hover:bg-accent-${habit.color}`)}
              >
                <Check className="h-5 w-5" />
              </button>
              <div className="flex-1 cursor-pointer" onClick={() => { setEditHabit(habit); setOpen(true); }}>
                <div className="font-semibold">{habit.name}</div>
                <div className="text-xs text-text-muted">{habit.frequency} · {habit.target_per_week}x/sem</div>
              </div>
              <div className="flex items-center gap-3">
                {streak && streak.current_streak > 0 && (
                  <Badge variant="orange">🔥 {streak.current_streak} dias</Badge>
                )}
                <button onClick={() => deleteHabit.mutate(habit.id)} className="text-xs text-text-muted hover:text-accent-rose transition-colors">
                  Excluir
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
