import { useState } from "react";
import { useParams, useNavigate, Link } from "react-router";
import { ArrowLeft, Pencil, Trash2, Target, CheckSquare, FolderKanban, ChevronRight } from "lucide-react";
import { toast } from "sonner";
import { areaIconMap } from "@/lib/icons";
import { cn } from "@/lib/cn";
import { useAreaSummary, useDeleteArea } from "@/hooks/use-areas";
import { useGoals } from "@/hooks/use-goals";
import { useTasks } from "@/hooks/use-tasks";
import { useProjects } from "@/hooks/use-projects";
import { AreaFormDialog } from "@/pages/areas";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import type { AreaStats } from "@/types/api";

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

export function Component() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: summary, isLoading } = useAreaSummary(id!);
  const { data: goals } = useGoals({ area_id: id! });
  const { data: tasks } = useTasks({ area_id: id!, is_completed: "false" });
  const { data: projects } = useProjects({ area_id: id! });
  const deleteArea = useDeleteArea();

  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);

  if (isLoading || !summary) {
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

  const { area, stats } = summary;
  const c = getColorClasses(area.color);
  const Icon = areaIconMap[area.icon];
  const pendingTasks = (tasks ?? []).slice(0, 5);
  const activeGoals = (goals ?? []).filter((g) => g.status !== "completed" && g.status !== "cancelled");
  const activeProjects = (projects ?? []).filter((p) => p.status === "active" || p.status === "planning");

  function handleDelete() {
    deleteArea.mutate(area.id, {
      onSuccess: () => { toast.success("Área excluída"); navigate("/areas"); },
      onError: () => toast.error("Erro ao excluir área"),
    });
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button onClick={() => navigate("/areas")} className="text-text-muted hover:text-text-primary transition-colors">
            <ArrowLeft size={20} />
          </button>
          <div className={cn("flex h-10 w-10 items-center justify-center rounded-[9px]", c.bg)}>
            {Icon ? <Icon size={22} className={c.text} /> : <span className="text-xl">{area.icon}</span>}
          </div>
          <h1 className="text-2xl font-bold">{area.name}</h1>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={() => setEditOpen(true)}>
            <Pencil className="h-3.5 w-3.5" /> Editar
          </Button>
          <Button variant="ghost" size="sm" onClick={() => setDeleteOpen(true)} className="text-accent-rose hover:text-accent-rose">
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>

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
        <DialogContent>
          <DialogHeader><DialogTitle>Editar area</DialogTitle></DialogHeader>
          <AreaFormDialog area={area} onClose={() => setEditOpen(false)} />
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
