import { useState } from "react";
import { useParams, useNavigate, Link } from "react-router";
import { ArrowLeft, Pencil, Trash2, Target, CheckSquare, FolderKanban, ChevronRight, Plus } from "lucide-react";
import { toast } from "sonner";
import { getAreaIcon, getAreaIconWithFallback, isRawAreaIcon } from "@/lib/icons";
import { cn } from "@/lib/cn";
import { useAreaSummary, useDeleteArea } from "@/hooks/use-areas";
import { useGoals, useCreateGoal } from "@/hooks/use-goals";
import { useTasks, useCreateTask } from "@/hooks/use-tasks";
import { useProjects, useCreateProject } from "@/hooks/use-projects";
import { useWorkspaceUsage } from "@/hooks/use-settings";
import { AreaFormDialog } from "@/components/areas/area-form-dialog";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import type { AreaStats, GoalPeriod } from "@/types/api";

const colorMap: Record<string, { bg: string; text: string; bar: string }> = {
  green: { bg: "bg-accent-green-soft", text: "text-accent-green", bar: "from-accent-green to-accent-sage" },
  orange: { bg: "bg-accent-orange-soft", text: "text-accent-orange", bar: "from-accent-orange to-accent-sand" },
  blue: { bg: "bg-accent-blue-soft", text: "text-accent-blue", bar: "from-accent-blue to-accent-sage" },
  rose: { bg: "bg-accent-rose-soft", text: "text-accent-rose", bar: "from-accent-rose to-accent-orange" },
  sand: { bg: "bg-accent-sand-soft", text: "text-accent-sand", bar: "from-accent-sand to-accent-blue" },
  sage: { bg: "bg-accent-sage-soft", text: "text-accent-sage", bar: "from-accent-sage to-accent-green" },
};

function getColorClasses(color: string) {
  return colorMap[color] ?? colorMap["orange"]!;
}

function getAreaDisplayName(name: string, slug: string) {
  const trimmed = name?.trim();
  if (trimmed) return trimmed;
  const normalizedSlug = slug?.trim();
  if (!normalizedSlug) return "Área sem nome";
  return normalizedSlug
    .split("-")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function StatCard({ label, value, colorClass }: { label: string; value: number | string; colorClass?: string }) {
  return (
    <div className="rounded-[14px] border border-border bg-bg-card p-4">
      <div className={cn("font-mono text-2xl font-bold", colorClass)}>{value}</div>
      <div className="text-xs text-text-muted mt-1">{label}</div>
    </div>
  );
}

function StatsRow({ stats, colorClasses }: { stats: AreaStats; colorClasses: { text: string } }) {
  return (
    <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
      <StatCard label="Score" value={stats.area_score ?? "—"} colorClass={colorClasses.text} />
      <StatCard label="Tarefas pendentes" value={stats.tasks_pending} />
      <StatCard label="Metas ativas" value={stats.goals_active} />
      <StatCard label="Projetos ativos" value={stats.projects_active} />
    </div>
  );
}

function todayISO() {
  return new Date().toISOString().slice(0, 10);
}

function computeGoalEndDate(start: string, period: GoalPeriod) {
  const base = new Date(`${start}T00:00:00`);
  const daysByPeriod: Record<GoalPeriod, number> = {
    weekly: 7,
    monthly: 30,
    quarterly: 90,
    yearly: 365,
  };
  base.setDate(base.getDate() + daysByPeriod[period]);
  return base.toISOString().slice(0, 10);
}

function formatDateForApi(date: string) {
  return new Date(`${date}T23:59:00`).toISOString();
}

function AreaPlanGate({
  used,
  max,
  isFull,
  onUpgrade,
}: {
  used: number;
  max: number;
  isFull: boolean;
  onUpgrade: () => void;
}) {
  return (
    <div
      className={cn(
        "flex flex-wrap items-center justify-between gap-3 rounded-[12px] border px-4 py-3",
        isFull
          ? "border-accent-rose/30 bg-accent-rose/10"
          : "border-accent-orange/30 bg-accent-orange/10",
      )}
    >
      <div className="text-sm">
        <p className={cn("font-medium", isFull ? "text-accent-rose" : "text-accent-orange")}>
          {isFull ? "Limite de áreas atingido" : "Você está próximo do limite de áreas"}
        </p>
        <p className="text-xs text-text-secondary">
          Uso atual: {used}/{max} áreas.
          {isFull ? " A criação de novas áreas está bloqueada no plano atual." : " Faça upgrade para evitar bloqueio."}
        </p>
      </div>
      <Button size="sm" variant="outline" onClick={onUpgrade}>
        Ver planos
      </Button>
    </div>
  );
}

export function Component() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: summary, isLoading, isError } = useAreaSummary(id ?? "");
  const { data: usage } = useWorkspaceUsage();
  const { data: goals } = useGoals({ area_id: id! });
  const { data: tasks } = useTasks({ area_id: id!, is_completed: "false" });
  const { data: projects } = useProjects({ area_id: id! });
  const deleteArea = useDeleteArea();
  const createTask = useCreateTask();
  const createGoal = useCreateGoal();
  const createProject = useCreateProject();

  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [taskOpen, setTaskOpen] = useState(false);
  const [goalOpen, setGoalOpen] = useState(false);
  const [projectOpen, setProjectOpen] = useState(false);
  const [taskTitle, setTaskTitle] = useState("");
  const [taskDueDate, setTaskDueDate] = useState("");
  const [goalTitle, setGoalTitle] = useState("");
  const [goalPeriod, setGoalPeriod] = useState<GoalPeriod>("quarterly");
  const [goalTarget, setGoalTarget] = useState("100");
  const [goalStartDate, setGoalStartDate] = useState(todayISO());
  const [projectTitle, setProjectTitle] = useState("");
  const [projectDescription, setProjectDescription] = useState("");
  const maxAreas = usage?.limits.max_areas ?? null;
  const usedAreas = usage?.counters.areas_count ?? null;
  const isUnlimited = maxAreas === -1;
  const usageRatio = !isUnlimited && maxAreas && usedAreas !== null ? usedAreas / maxAreas : 0;
  const areaLimitReached = Boolean(usage) && !isUnlimited && maxAreas !== null && usedAreas !== null && usedAreas >= maxAreas;
  const areaLimitWarning = Boolean(usage) && !isUnlimited && maxAreas !== null && usedAreas !== null && usageRatio >= 0.8 && !areaLimitReached;
  const showPlanGate = areaLimitReached || areaLimitWarning;

  if (!id) {
    return (
      <div className="flex min-h-[260px] flex-col items-center justify-center gap-3 text-center">
        <p className="text-sm text-text-muted">Área inválida.</p>
        <Button variant="outline" onClick={() => navigate("/areas")}>Voltar para Áreas</Button>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-3">
          <Skeleton className="h-8 w-8" />
          <Skeleton className="h-8 w-48" />
        </div>
        <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
          {[1, 2, 3, 4].map((i) => <Skeleton key={i} className="h-20 rounded-[14px]" />)}
        </div>
        <Skeleton className="h-48 rounded-[14px]" />
      </div>
    );
  }

  if (isError || !summary) {
    return (
      <div className="flex min-h-[300px] flex-col items-center justify-center gap-3 text-center">
        <p className="text-base font-medium text-text-primary">Não foi possível carregar esta área.</p>
        <p className="text-sm text-text-muted">Tente novamente ou volte para a lista de áreas.</p>
        <Button variant="outline" onClick={() => navigate("/areas")}>Voltar para Áreas</Button>
      </div>
    );
  }

  const { area, stats } = summary;
  const c = getColorClasses(area.color);
  const Icon = getAreaIcon(area.icon);
  const FallbackIcon = getAreaIconWithFallback(area);
  const rawIcon = area.icon?.trim();
  const showRawIcon = isRawAreaIcon(rawIcon);
  const displayName = getAreaDisplayName(area.name, area.slug);
  const pendingTasks = (tasks ?? []).slice(0, 5);
  const activeGoals = (goals ?? []).filter((g) => g.status !== "completed" && g.status !== "cancelled");
  const activeProjects = (projects ?? []).filter((p) => p.status === "active" || p.status === "planning");

  function handleDelete() {
    deleteArea.mutate(area.id, {
      onSuccess: () => {
        setDeleteOpen(false);
        toast.success("Área excluída");
        navigate("/areas");
      },
      onError: () => toast.error("Erro ao excluir área"),
    });
  }

  function handleCreateTask(e: React.FormEvent) {
    e.preventDefault();
    const title = taskTitle.trim();
    if (!title) return;

    createTask.mutate(
      {
        title,
        area_id: area.id,
        ...(taskDueDate && { due_date: formatDateForApi(taskDueDate) }),
      },
      {
        onSuccess: () => {
          setTaskTitle("");
          setTaskDueDate("");
          setTaskOpen(false);
        },
      },
    );
  }

  function handleCreateGoal(e: React.FormEvent) {
    e.preventDefault();
    const title = goalTitle.trim();
    if (!title) return;

    const targetValue = Number(goalTarget);
    const startDate = goalStartDate || todayISO();
    const endDate = computeGoalEndDate(startDate, goalPeriod);

    createGoal.mutate(
      {
        title,
        period: goalPeriod,
        area_id: area.id,
        target_value: Number.isFinite(targetValue) ? targetValue : 100,
        start_date: startDate,
        end_date: endDate,
      },
      {
        onSuccess: () => {
          toast.success("Meta criada");
          setGoalTitle("");
          setGoalTarget("100");
          setGoalPeriod("quarterly");
          setGoalStartDate(todayISO());
          setGoalOpen(false);
        },
        onError: (err) => {
          if (err instanceof Error && err.message === "goal limit reached") {
            toast.error("Limite de metas atingido. Faça upgrade para criar mais.");
            return;
          }
          toast.error("Erro ao criar meta");
        },
      },
    );
  }

  function handleCreateProject(e: React.FormEvent) {
    e.preventDefault();
    const title = projectTitle.trim();
    if (!title) return;

    createProject.mutate(
      {
        title,
        area_id: area.id,
        description: projectDescription.trim() || null,
      },
      {
        onSuccess: () => {
          toast.success("Projeto criado");
          setProjectTitle("");
          setProjectDescription("");
          setProjectOpen(false);
        },
        onError: (err) => {
          if (err instanceof Error && err.message === "project limit reached") {
            toast.error("Limite de projetos atingido. Faça upgrade para criar mais.");
            return;
          }
          toast.error("Erro ao criar projeto");
        },
      },
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex min-w-0 items-center gap-3">
          <button
            onClick={() => navigate("/areas")}
            className="shrink-0 text-text-muted hover:text-text-primary transition-colors"
          >
            <ArrowLeft size={20} />
          </button>
          <div className={cn("flex h-10 w-10 shrink-0 items-center justify-center rounded-[9px]", c.bg)}>
            {Icon ? <Icon size={22} className={c.text} /> : showRawIcon ? <span className="text-xl">{rawIcon}</span> : <FallbackIcon size={22} className={c.text} />}
          </div>
          <h1 className="truncate text-xl font-bold sm:text-2xl">{displayName}</h1>
        </div>
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
          <div className="grid grid-cols-3 gap-2 sm:flex">
            <Button variant="outline" size="sm" onClick={() => setTaskOpen(true)} className="w-full sm:w-auto" aria-label="Nova tarefa">
              <Plus className="h-3.5 w-3.5" />
              <span className="sm:hidden">Tarefa</span>
              <span className="hidden sm:inline">Nova tarefa</span>
            </Button>
            <Button variant="outline" size="sm" onClick={() => setGoalOpen(true)} className="w-full sm:w-auto" aria-label="Nova meta">
              <Plus className="h-3.5 w-3.5" />
              <span className="sm:hidden">Meta</span>
              <span className="hidden sm:inline">Nova meta</span>
            </Button>
            <Button variant="outline" size="sm" onClick={() => setProjectOpen(true)} className="w-full sm:w-auto" aria-label="Novo projeto">
              <Plus className="h-3.5 w-3.5" />
              <span className="sm:hidden">Projeto</span>
              <span className="hidden sm:inline">Novo projeto</span>
            </Button>
          </div>
          <div className="flex items-center justify-end gap-2">
            <Button variant="outline" size="sm" onClick={() => setEditOpen(true)} aria-label="Editar área">
              <Pencil className="h-3.5 w-3.5" />
              <span className="hidden sm:inline">Editar</span>
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setDeleteOpen(true)}
              className="text-accent-rose hover:text-accent-rose"
              aria-label="Excluir área"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      </div>

      {showPlanGate && maxAreas !== null && usedAreas !== null && (
        <AreaPlanGate
          used={usedAreas}
          max={maxAreas}
          isFull={areaLimitReached}
          onUpgrade={() => navigate("/billing")}
        />
      )}

      {/* Stats */}
      <StatsRow stats={stats} colorClasses={c} />

      {/* Goals */}
      <section>
        <div className="flex items-center justify-between mb-3">
          <h2 className="flex items-center gap-2 text-sm font-semibold uppercase tracking-[1.5px] text-text-secondary">
            <Target size={16} /> Metas
          </h2>
          <span className="text-xs text-text-muted">{activeGoals.length} ativas</span>
        </div>
        {activeGoals.length === 0 ? (
          <p className="text-sm text-text-muted">Nenhuma meta ativa nesta área.</p>
        ) : (
          <div className="space-y-2">
            {activeGoals.map((goal) => {
              const progress = goal.target_value ? Math.round((goal.current_value / goal.target_value) * 100) : 0;
              return (
                <div key={goal.id} className="rounded-[14px] border border-border bg-bg-card p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium">{goal.title}</span>
                    <span className="text-xs text-text-muted">{progress}%</span>
                  </div>
                  <div className="h-[3px] overflow-hidden rounded-[3px] bg-border">
                    <div className={cn("h-full rounded-[3px] bg-gradient-to-r", c.bar)} style={{ width: `${progress}%` }} />
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </section>

      {/* Tasks */}
      <section>
        <div className="flex items-center justify-between mb-3">
          <h2 className="flex items-center gap-2 text-sm font-semibold uppercase tracking-[1.5px] text-text-secondary">
            <CheckSquare size={16} /> Tarefas pendentes
          </h2>
          {(tasks ?? []).length > 5 && (
            <Link to={`/tasks?area_id=${area.id}`} className="flex items-center gap-1 text-xs text-accent-orange hover:underline">
              Ver todas <ChevronRight size={14} />
            </Link>
          )}
        </div>
        {pendingTasks.length === 0 ? (
          <p className="text-sm text-text-muted">Nenhuma tarefa pendente.</p>
        ) : (
          <div className="space-y-1.5">
            {pendingTasks.map((task) => (
              <div key={task.id} className="flex items-center gap-3 rounded-lg border border-border bg-bg-card px-4 py-2.5">
                <div className={cn("h-2 w-2 rounded-full", task.priority === "critical" ? "bg-accent-rose" : task.priority === "high" ? "bg-accent-orange" : "bg-border")} />
                <span className="text-sm">{task.title}</span>
                {task.due_date && (
                  <span className="ml-auto text-xs text-text-muted">{new Date(task.due_date).toLocaleDateString("pt-BR", { day: "2-digit", month: "short" })}</span>
                )}
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Projects */}
      <section>
        <div className="flex items-center justify-between mb-3">
          <h2 className="flex items-center gap-2 text-sm font-semibold uppercase tracking-[1.5px] text-text-secondary">
            <FolderKanban size={16} /> Projetos
          </h2>
          <span className="text-xs text-text-muted">{activeProjects.length} ativos</span>
        </div>
        {activeProjects.length === 0 ? (
          <p className="text-sm text-text-muted">Nenhum projeto ativo.</p>
        ) : (
          <div className="grid grid-cols-1 gap-2 md:grid-cols-2">
            {activeProjects.map((project) => {
              const progress = project.tasks_total > 0 ? Math.round((project.tasks_completed / project.tasks_total) * 100) : 0;
              return (
                <Link key={project.id} to={`/projects`} className="rounded-[14px] border border-border bg-bg-card p-4 hover:border-transparent transition-colors">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium">{project.title}</span>
                    <span className="text-xs text-text-muted">{progress}%</span>
                  </div>
                  <div className="h-[3px] overflow-hidden rounded-[3px] bg-border">
                    <div className={cn("h-full rounded-[3px] bg-gradient-to-r", c.bar)} style={{ width: `${progress}%` }} />
                  </div>
                  <div className="mt-2 text-xs text-text-muted">
                    {project.tasks_completed}/{project.tasks_total} tarefas
                  </div>
                </Link>
              );
            })}
          </div>
        )}
      </section>

      {/* Edit Dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-w-3xl gap-0 overflow-hidden p-0">
          <DialogHeader className="sr-only"><DialogTitle>Editar área</DialogTitle></DialogHeader>
          <AreaFormDialog area={area} onClose={() => setEditOpen(false)} />
        </DialogContent>
      </Dialog>

      <Dialog open={taskOpen} onOpenChange={setTaskOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Nova tarefa</DialogTitle>
          </DialogHeader>
          <form className="space-y-4" onSubmit={handleCreateTask}>
            <div className="space-y-2">
              <Label htmlFor="area-task-title">Título</Label>
              <Input
                id="area-task-title"
                value={taskTitle}
                onChange={(e) => setTaskTitle(e.target.value)}
                placeholder="Ex.: Marcar consulta anual"
                required
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="area-task-due">Prazo (opcional)</Label>
              <Input
                id="area-task-due"
                type="date"
                value={taskDueDate}
                onChange={(e) => setTaskDueDate(e.target.value)}
              />
            </div>
            <Button type="submit" className="w-full" disabled={createTask.isPending}>
              Criar tarefa nesta área
            </Button>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={goalOpen} onOpenChange={setGoalOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Nova meta</DialogTitle>
          </DialogHeader>
          <form className="space-y-4" onSubmit={handleCreateGoal}>
            <div className="space-y-2">
              <Label htmlFor="area-goal-title">Título</Label>
              <Input
                id="area-goal-title"
                value={goalTitle}
                onChange={(e) => setGoalTitle(e.target.value)}
                placeholder="Ex.: Correr 120 km no trimestre"
                required
                autoFocus
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label>Período</Label>
                <Select value={goalPeriod} onValueChange={(v) => setGoalPeriod(v as GoalPeriod)}>
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
                <Label htmlFor="area-goal-target">Meta (valor)</Label>
                <Input
                  id="area-goal-target"
                  type="number"
                  min="0"
                  value={goalTarget}
                  onChange={(e) => setGoalTarget(e.target.value)}
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="area-goal-start">Início</Label>
              <Input
                id="area-goal-start"
                type="date"
                value={goalStartDate}
                onChange={(e) => setGoalStartDate(e.target.value)}
              />
            </div>
            <Button type="submit" className="w-full" disabled={createGoal.isPending}>
              Criar meta nesta área
            </Button>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={projectOpen} onOpenChange={setProjectOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Novo projeto</DialogTitle>
          </DialogHeader>
          <form className="space-y-4" onSubmit={handleCreateProject}>
            <div className="space-y-2">
              <Label htmlFor="area-project-title">Título</Label>
              <Input
                id="area-project-title"
                value={projectTitle}
                onChange={(e) => setProjectTitle(e.target.value)}
                placeholder="Ex.: Planejar rotina de treino"
                required
                autoFocus
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="area-project-description">Descrição (opcional)</Label>
              <Input
                id="area-project-description"
                value={projectDescription}
                onChange={(e) => setProjectDescription(e.target.value)}
                placeholder="Escopo e próximos passos"
              />
            </div>
            <Button type="submit" className="w-full" disabled={createProject.isPending}>
              Criar projeto nesta área
            </Button>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete Confirm */}
      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title="Excluir área"
        description={`Esta área tem ${stats.tasks_pending} tarefas pendentes, ${stats.goals_active} metas e ${stats.projects_active} projetos. Deseja excluir?`}
        confirmLabel="Excluir"
        variant="destructive"
        onConfirm={handleDelete}
        loading={deleteArea.isPending}
      />
    </div>
  );
}
