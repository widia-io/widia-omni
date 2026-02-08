import { useState } from "react";
import { Plus, Star } from "lucide-react";
import { useTasks, useCreateTask, useUpdateTask, useDeleteTask, useCompleteTask, useToggleFocus } from "@/hooks/use-tasks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Skeleton } from "@/components/ui/skeleton";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { cn } from "@/lib/cn";
import type { Task, TaskPriority } from "@/types/api";

function TaskFormDialog({ task, onClose }: { task?: Task; onClose: () => void }) {
  const create = useCreateTask();
  const update = useUpdateTask();
  const [title, setTitle] = useState(task?.title ?? "");
  const [priority, setPriority] = useState<TaskPriority>(task?.priority ?? "medium");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (task) {
      update.mutate({ id: task.id, title, priority }, { onSuccess: onClose });
    } else {
      create.mutate({ title, priority }, { onSuccess: onClose });
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label>Titulo</Label>
        <Input value={title} onChange={(e) => setTitle(e.target.value)} required />
      </div>
      <div className="space-y-2">
        <Label>Prioridade</Label>
        <Select value={priority} onValueChange={(v) => setPriority(v as TaskPriority)}>
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="low">Baixa</SelectItem>
            <SelectItem value="medium">Media</SelectItem>
            <SelectItem value="high">Alta</SelectItem>
            <SelectItem value="critical">Critica</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <Button type="submit" className="w-full" disabled={create.isPending || update.isPending}>
        {task ? "Salvar" : "Criar tarefa"}
      </Button>
    </form>
  );
}

const priorityColor: Record<string, "default" | "green" | "orange" | "rose"> = {
  low: "default", medium: "blue" as "default", high: "orange", critical: "rose",
};

export function Component() {
  const [filter, setFilter] = useState<Record<string, string>>({});
  const { data: tasks, isLoading } = useTasks(filter);
  const deleteTask = useDeleteTask();
  const completeTask = useCompleteTask();
  const toggleFocus = useToggleFocus();
  const [editTask, setEditTask] = useState<Task | undefined>();
  const [open, setOpen] = useState(false);

  if (isLoading) return <div className="space-y-3">{[1,2,3].map(i => <Skeleton key={i} className="h-14 rounded-[14px]" />)}</div>;

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold">Tarefas</h1>
        <div className="flex gap-2">
          <Select value={filter.is_completed ?? ""} onValueChange={(v) => setFilter(p => ({ ...p, is_completed: v }))}>
            <SelectTrigger className="w-36"><SelectValue placeholder="Status" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="false">Pendentes</SelectItem>
              <SelectItem value="true">Concluidas</SelectItem>
            </SelectContent>
          </Select>
          <Dialog open={open} onOpenChange={(v) => { setOpen(v); if (!v) setEditTask(undefined); }}>
            <DialogTrigger asChild>
              <Button size="sm"><Plus className="h-4 w-4" /> Nova tarefa</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader><DialogTitle>{editTask ? "Editar tarefa" : "Nova tarefa"}</DialogTitle></DialogHeader>
              <TaskFormDialog task={editTask} onClose={() => setOpen(false)} />
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <div className="space-y-2">
        {(tasks ?? []).map((task) => (
          <div key={task.id} className="flex items-center gap-3 rounded-[14px] border border-border bg-bg-card px-4 py-3">
            <Checkbox
              checked={task.is_completed}
              onCheckedChange={() => completeTask.mutate(task.id)}
            />
            <div className="flex-1 cursor-pointer" onClick={() => { setEditTask(task); setOpen(true); }}>
              <span className={cn("text-sm", task.is_completed && "text-text-muted line-through")}>{task.title}</span>
            </div>
            <button onClick={() => toggleFocus.mutate(task.id)} className={cn("transition-colors", task.is_focus ? "text-accent-orange" : "text-text-muted hover:text-accent-orange")}>
              <Star className="h-4 w-4" fill={task.is_focus ? "currentColor" : "none"} />
            </button>
            <Badge variant={priorityColor[task.priority]}>{task.priority}</Badge>
            <button onClick={() => deleteTask.mutate(task.id)} className="text-xs text-text-muted hover:text-accent-rose transition-colors">×</button>
          </div>
        ))}
      </div>
    </div>
  );
}
