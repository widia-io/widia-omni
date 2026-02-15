import { useState } from "react";
import { Plus, Archive, ArchiveRestore } from "lucide-react";
import {
  useProjects,
  useCreateProject,
  useUpdateProject,
  useDeleteProject,
  useArchiveProject,
  useUnarchiveProject,
} from "@/hooks/use-projects";
import { useAreas } from "@/hooks/use-areas";
import { useGoals } from "@/hooks/use-goals";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import type { Project, ProjectStatus } from "@/types/api";

const statusBadge: Record<ProjectStatus, { v: "blue" | "green" | "orange" | "rose" | "default"; label: string }> = {
  planning: { v: "blue", label: "Planejamento" },
  active: { v: "green", label: "Ativo" },
  paused: { v: "orange", label: "Pausado" },
  completed: { v: "green", label: "Concluído" },
  cancelled: { v: "rose", label: "Cancelado" },
};

const COLOR_OPTIONS = [
  { value: "#E07A5F", label: "Terracota" },
  { value: "#81B29A", label: "Sage" },
  { value: "#F2CC8F", label: "Sand" },
  { value: "#3D405B", label: "Charcoal" },
  { value: "#5B8DBE", label: "Blue" },
  { value: "#E4959E", label: "Rose" },
];

function ProjectFormDialog({ project, onClose }: { project?: Project; onClose: () => void }) {
  const { data: areas } = useAreas();
  const { data: goals } = useGoals();
  const create = useCreateProject();
  const update = useUpdateProject();
  const [title, setTitle] = useState(project?.title ?? "");
  const [description, setDescription] = useState(project?.description ?? "");
  const [status, setStatus] = useState<ProjectStatus>(project?.status ?? "planning");
  const [areaId, setAreaId] = useState(project?.area_id ?? "");
  const [goalId, setGoalId] = useState(project?.goal_id ?? "");
  const [color, setColor] = useState(project?.color ?? "#E07A5F");
  const [startDate, setStartDate] = useState(project?.start_date ?? "");
  const [targetDate, setTargetDate] = useState(project?.target_date ?? "");

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const data = {
      title,
      description: description || null,
      status,
      color,
      area_id: areaId || null,
      goal_id: goalId || null,
      start_date: startDate || null,
      target_date: targetDate || null,
    };
    if (project) {
      update.mutate({ id: project.id, ...data }, { onSuccess: onClose });
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
        <Label>Descrição</Label>
        <Input value={description} onChange={(e) => setDescription(e.target.value)} />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-2">
          <Label>Status</Label>
          <Select value={status} onValueChange={(v) => setStatus(v as ProjectStatus)}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              {Object.entries(statusBadge).map(([k, { label }]) => (
                <SelectItem key={k} value={k}>{label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label>Cor</Label>
          <div className="flex gap-1.5 pt-1">
            {COLOR_OPTIONS.map((c) => (
              <button
                key={c.value}
                type="button"
                onClick={() => setColor(c.value)}
                className="h-7 w-7 rounded-full border-2 transition-transform"
                style={{
                  backgroundColor: c.value,
                  borderColor: color === c.value ? c.value : "transparent",
                  transform: color === c.value ? "scale(1.15)" : "scale(1)",
                }}
              />
            ))}
          </div>
        </div>
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-2">
          <Label>Área</Label>
          <Select value={areaId ?? ""} onValueChange={setAreaId}>
            <SelectTrigger><SelectValue placeholder="Nenhuma" /></SelectTrigger>
            <SelectContent>
              {(areas ?? []).map((a) => <SelectItem key={a.id} value={a.id}>{a.name}</SelectItem>)}
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label>Meta</Label>
          <Select value={goalId ?? ""} onValueChange={setGoalId}>
            <SelectTrigger><SelectValue placeholder="Nenhuma" /></SelectTrigger>
            <SelectContent>
              {(goals ?? []).map((g) => <SelectItem key={g.id} value={g.id}>{g.title}</SelectItem>)}
            </SelectContent>
          </Select>
        </div>
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-2">
          <Label>Início</Label>
          <Input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} />
        </div>
        <div className="space-y-2">
          <Label>Prazo</Label>
          <Input type="date" value={targetDate} onChange={(e) => setTargetDate(e.target.value)} />
        </div>
      </div>
      <Button type="submit" className="w-full" disabled={create.isPending || update.isPending}>
        {project ? "Salvar" : "Criar projeto"}
      </Button>
    </form>
  );
}

export function Component() {
  const [filter, setFilter] = useState<Record<string, string>>({});
  const { data: projects, isLoading } = useProjects(filter);
  const { data: areas } = useAreas();
  const { data: goals } = useGoals();
  const deleteProject = useDeleteProject();
  const archiveProject = useArchiveProject();
  const unarchiveProject = useUnarchiveProject();
  const [editProject, setEditProject] = useState<Project | undefined>();
  const [open, setOpen] = useState(false);

  const areaMap = new Map((areas ?? []).map((a) => [a.id, a.name]));
  const goalMap = new Map((goals ?? []).map((g) => [g.id, g.title]));

  if (isLoading) {
    return (
      <div className="space-y-3">
        {[1, 2, 3].map((i) => <Skeleton key={i} className="h-28 rounded-[14px]" />)}
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold">Projetos</h1>
        <div className="flex gap-2">
          <Select value={filter.status ?? ""} onValueChange={(v) => setFilter((p) => ({ ...p, status: v }))}>
            <SelectTrigger className="w-40"><SelectValue placeholder="Status" /></SelectTrigger>
            <SelectContent>
              {Object.entries(statusBadge).map(([k, { label }]) => (
                <SelectItem key={k} value={k}>{label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Dialog open={open} onOpenChange={(v) => { setOpen(v); if (!v) setEditProject(undefined); }}>
            <DialogTrigger asChild>
              <Button size="sm"><Plus className="h-4 w-4" /> Novo projeto</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader><DialogTitle>{editProject ? "Editar projeto" : "Novo projeto"}</DialogTitle></DialogHeader>
              <ProjectFormDialog project={editProject} onClose={() => setOpen(false)} />
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <div className="space-y-3">
        {(projects ?? []).map((project) => {
          const pct = project.tasks_total > 0
            ? Math.round((project.tasks_completed / project.tasks_total) * 100)
            : 0;
          const badge = statusBadge[project.status];
          const areaName = project.area_id ? areaMap.get(project.area_id) : null;
          const goalTitle = project.goal_id ? goalMap.get(project.goal_id) : null;

          return (
            <div
              key={project.id}
              className="rounded-[14px] border border-border bg-bg-card p-5 cursor-pointer hover:border-border-hover transition-colors"
              onClick={() => { setEditProject(project); setOpen(true); }}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span
                    className="h-3 w-3 rounded-full shrink-0"
                    style={{ backgroundColor: project.color }}
                  />
                  <span className="font-serif font-medium">{project.title}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="font-mono text-sm text-text-muted">
                    {project.tasks_completed}/{project.tasks_total}
                  </span>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      if (project.is_archived) {
                        unarchiveProject.mutate(project.id);
                      } else {
                        archiveProject.mutate(project.id);
                      }
                    }}
                    className="text-text-muted hover:text-text-secondary"
                    title={project.is_archived ? "Desarquivar" : "Arquivar"}
                  >
                    {project.is_archived ? <ArchiveRestore className="h-4 w-4" /> : <Archive className="h-4 w-4" />}
                  </button>
                  <button
                    onClick={(e) => { e.stopPropagation(); deleteProject.mutate(project.id); }}
                    className="text-xs text-text-muted hover:text-accent-rose"
                  >
                    Excluir
                  </button>
                </div>
              </div>

              {project.tasks_total > 0 && (
                <Progress value={pct} className="mb-2" />
              )}

              <div className="flex gap-2 flex-wrap">
                <Badge variant={badge.v}>{badge.label}</Badge>
                {areaName && <Badge variant="sage">{areaName}</Badge>}
                {goalTitle && <Badge variant="blue">{goalTitle}</Badge>}
                {project.is_archived && <Badge variant="default">Arquivado</Badge>}
              </div>
            </div>
          );
        })}

        {(projects ?? []).length === 0 && (
          <div className="rounded-[14px] border border-dashed border-border p-10 text-center text-text-muted">
            Nenhum projeto encontrado. Crie o primeiro!
          </div>
        )}
      </div>
    </div>
  );
}
