import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router";
import { Target, Activity, CheckCircle2, Plus, X } from "lucide-react";
import { toast } from "sonner";
import {
  useOnboardingStatus,
  useOnboardingAreas,
  useOnboardingGoals,
  useOnboardingHabits,
  useOnboardingSkipHabits,
  useOnboardingProject,
  useOnboardingFirstTask,
  useOnboardingComplete,
  useAreaTemplates,
  useGoalSuggestions,
  type OnboardingGoalInput,
} from "@/hooks/use-onboarding";
import { useAreas } from "@/hooks/use-areas";
import { useGoals } from "@/hooks/use-goals";
import { useProjects } from "@/hooks/use-projects";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { cn } from "@/lib/cn";
import { areaIconMap } from "@/lib/icons";
import type { AreaTemplate, GoalPeriod, LifeArea } from "@/types/api";

type AreaTemplateLike = Pick<AreaTemplate, "name" | "slug" | "icon" | "color" | "weight">;

const colorTw: Record<string, { bg: string; text: string }> = {
  green: { bg: "bg-accent-green/10", text: "text-accent-green" },
  orange: { bg: "bg-accent-orange/10", text: "text-accent-orange" },
  sand: { bg: "bg-accent-sand/10", text: "text-accent-sand" },
  rose: { bg: "bg-accent-rose/10", text: "text-accent-rose" },
  sage: { bg: "bg-accent-sage/10", text: "text-accent-sage" },
  blue: { bg: "bg-accent-blue/10", text: "text-accent-blue" },
  sky: { bg: "bg-accent-sky/10", text: "text-accent-sky" },
  violet: { bg: "bg-accent-violet/10", text: "text-accent-violet" },
};

const onboardingColorOptions = ["green", "orange", "sand", "rose", "sage", "blue", "sky", "violet"] as const;
const onboardingIconOptions = ["heart", "briefcase", "dollar-sign", "users", "book", "sun", "dumbbell", "brain", "home", "gamepad-2", "sparkles"] as const;

interface CustomArea {
  id: string;
  name: string;
  slug: string;
  icon: string;
  color: string;
  weight: number;
}

function slugify(value: string): string {
  return value
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

function StepProgress({ current, total }: { current: number; total: number }) {
  return (
    <div className="mb-8 flex items-center gap-2">
      {Array.from({ length: total }, (_, i) => (
        <div key={i} className={cn("h-1.5 flex-1 rounded-full transition-colors", i <= current ? "bg-accent-orange" : "bg-border")} />
      ))}
    </div>
  );
}

function GoalSuggestionsForArea({
  areaSlug,
  onAdd,
}: {
  areaSlug: string;
  onAdd: (title: string, period: GoalPeriod) => void;
}) {
  const { data: suggestions } = useGoalSuggestions("pt-BR", areaSlug);
  if (!suggestions || suggestions.length === 0) return null;

  return (
    <div className="flex flex-wrap gap-2">
      {suggestions.map((s) => (
        <button
          key={`${areaSlug}:${s.title}`}
          onClick={() => onAdd(s.title, s.period as GoalPeriod)}
          className="flex items-center gap-1.5 rounded-full border border-border bg-bg-card px-3 py-1.5 text-xs text-text-secondary transition-colors hover:border-accent-orange hover:text-text-primary"
        >
          <Plus className="h-3 w-3" />
          {s.title}
        </button>
      ))}
    </div>
  );
}

export function Component() {
  const navigate = useNavigate();
  const { data: status, isLoading } = useOnboardingStatus();
  const { data: templates, isLoading: templatesLoading } = useAreaTemplates("pt-BR");
  const { data: areasData } = useAreas();
  const { data: goalsData } = useGoals();
  const { data: projectsData } = useProjects();

  const submitAreas = useOnboardingAreas();
  const submitGoals = useOnboardingGoals();
  const submitHabits = useOnboardingHabits();
  const skipHabits = useOnboardingSkipHabits();
  const submitProject = useOnboardingProject();
  const submitFirstTask = useOnboardingFirstTask();
  const complete = useOnboardingComplete();

  const [step, setStep] = useState<number | null>(null);
  const [selectedSlugs, setSelectedSlugs] = useState<Set<string>>(new Set());
  const [customAreas, setCustomAreas] = useState<CustomArea[]>([]);
  const [createdAreas, setCreatedAreas] = useState<LifeArea[]>([]);

  const [customModalOpen, setCustomModalOpen] = useState(false);
  const [customName, setCustomName] = useState("");
  const [customIcon, setCustomIcon] = useState<string>("heart");
  const [customColor, setCustomColor] = useState<string>("blue");

  const [goalTitle, setGoalTitle] = useState("");
  const [goalAreaId, setGoalAreaId] = useState("");
  const [goals, setGoals] = useState<OnboardingGoalInput[]>([]);

  const [habitName, setHabitName] = useState("");
  const [habits, setHabits] = useState<string[]>([]);

  const [projectTitle, setProjectTitle] = useState("");
  const [projectAreaId, setProjectAreaId] = useState("");
  const [projectGoalId, setProjectGoalId] = useState("");
  const [createdProjectId, setCreatedProjectId] = useState("");

  const [firstTaskProjectId, setFirstTaskProjectId] = useState("");
  const [firstTaskTitle, setFirstTaskTitle] = useState("");

  const areaTemplates = templates ?? [];
  const onboardingAreas = createdAreas.length > 0 ? createdAreas : (areasData ?? []);
  const onboardingGoals = goalsData ?? [];
  const onboardingProjects = projectsData ?? [];

  useEffect(() => {
    if (status?.completed) {
      navigate("/dashboard", { replace: true });
    }
  }, [status?.completed, navigate]);

  useEffect(() => {
    const firstArea = onboardingAreas.at(0);
    if (!goalAreaId && onboardingAreas.length > 0) {
      setGoalAreaId(firstArea?.id ?? "");
    }
    if (!projectAreaId && onboardingAreas.length > 0) {
      setProjectAreaId(firstArea?.id ?? "");
    }
  }, [goalAreaId, projectAreaId, onboardingAreas]);

  useEffect(() => {
    const firstProject = onboardingProjects.at(0);
    if (!firstTaskProjectId) {
      if (createdProjectId) {
        setFirstTaskProjectId(createdProjectId);
      } else if (onboardingProjects.length > 0) {
        setFirstTaskProjectId(firstProject?.id ?? "");
      }
    }
  }, [firstTaskProjectId, createdProjectId, onboardingProjects]);

  const templateSelections: AreaTemplateLike[] = useMemo(
    () => areaTemplates.filter((a) => selectedSlugs.has(a.slug)),
    [areaTemplates, selectedSlugs],
  );

  const selectedAreasForSubmit: AreaTemplateLike[] = useMemo(
    () => [
      ...templateSelections,
      ...customAreas.map((a) => ({
        name: a.name,
        slug: a.slug,
        icon: a.icon,
        color: a.color,
        weight: a.weight,
      })),
    ],
    [templateSelections, customAreas],
  );

  const onboardingAreaByID = useMemo(
    () => new Map(onboardingAreas.map((area) => [area.id, area])),
    [onboardingAreas],
  );
  const onboardingAreaBySlug = useMemo(
    () => new Map(onboardingAreas.map((area) => [area.slug, area])),
    [onboardingAreas],
  );

  const projectGoals = useMemo(
    () => onboardingGoals.filter((goal) => !projectAreaId || goal.area_id === projectAreaId),
    [onboardingGoals, projectAreaId],
  );

  if (isLoading || templatesLoading) return <Skeleton className="h-96 rounded-[14px]" />;

  const derivedStep = status?.steps?.first_task
    ? 5
    : status?.steps?.project
      ? 4
      : status?.steps?.habits
        ? 3
        : status?.steps?.goals
          ? 2
          : status?.steps?.areas
            ? 1
            : 0;
  const currentStep = step ?? derivedStep;

  function toggleArea(slug: string) {
    setSelectedSlugs((prev) => {
      const next = new Set(prev);
      next.has(slug) ? next.delete(slug) : next.add(slug);
      return next;
    });
  }

  function openCustomModal() {
    setCustomName("");
    setCustomIcon("heart");
    setCustomColor("blue");
    setCustomModalOpen(true);
  }

  function addCustomArea() {
    const name = customName.trim();
    if (!name) return;
    const slug = slugify(name);
    if (!slug) return;
    const slugAlreadyUsed = areaTemplates.some((area) => area.slug === slug) || customAreas.some((area) => area.slug === slug);
    if (slugAlreadyUsed) {
      toast.error("Já existe uma área com esse nome.");
      return;
    }

    const area: CustomArea = {
      id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
      name,
      slug,
      icon: customIcon,
      color: customColor,
      weight: 1,
    };
    setCustomAreas((prev) => [...prev, area]);
    setCustomModalOpen(false);
  }

  function removeCustomArea(id: string) {
    setCustomAreas((prev) => prev.filter((a) => a.id !== id));
  }

  function handleAreasSubmit() {
    const data = selectedAreasForSubmit.map((a, index) => ({
      name: a.name,
      slug: a.slug,
      icon: a.icon,
      color: a.color,
      weight: a.weight,
      sort_order: index + 1,
    }));

    submitAreas.mutate(data, {
      onSuccess: (areas) => {
        setCreatedAreas(areas);
        const firstArea = areas.at(0);
        if (areas.length > 0) {
          setGoalAreaId(firstArea?.id ?? "");
          setProjectAreaId(firstArea?.id ?? "");
        }
        setStep(1);
      },
    });
  }

  function addGoal(areaID: string, title: string, period: GoalPeriod) {
    const normalized = title.trim();
    if (!normalized) return;
    const duplicate = goals.some((g) => g.area_id === areaID && g.title.toLowerCase() === normalized.toLowerCase());
    if (duplicate) return;
    setGoals((prev) => [...prev, { area_id: areaID, title: normalized, period }]);
  }

  function handleGoalsSubmit() {
    if (goals.length === 0) return;
    submitGoals.mutate(goals, { onSuccess: () => setStep(2) });
  }

  function handleHabitsSubmit() {
    const data = habits.map((name) => ({ name, frequency: "daily" as const, target_per_week: 7 }));
    submitHabits.mutate(data, { onSuccess: () => setStep(3) });
  }

  function handleSkipHabits() {
    skipHabits.mutate(undefined, { onSuccess: () => setStep(3) });
  }

  function handleProjectSubmit() {
    if (!projectTitle.trim()) return;
    submitProject.mutate(
      {
        title: projectTitle.trim(),
        area_id: projectAreaId || null,
        goal_id: projectGoalId || null,
        color: "#E07A5F",
      },
      {
        onSuccess: (project) => {
          setCreatedProjectId(project.id);
          setFirstTaskProjectId(project.id);
          setStep(4);
        },
      },
    );
  }

  function handleFirstTaskSubmit() {
    if (!firstTaskProjectId || !firstTaskTitle.trim()) return;
    submitFirstTask.mutate(
      { project_id: firstTaskProjectId, title: firstTaskTitle.trim() },
      { onSuccess: () => setStep(5) },
    );
  }

  function handleComplete() {
    complete.mutate(undefined, {
      onSuccess: () => navigate("/dashboard"),
      onError: (err: any) => {
        const missingSteps = Array.isArray(err?.body?.missing_steps) ? err.body.missing_steps.join(", ") : null;
        if (missingSteps) {
          toast.error(`Ainda faltam etapas: ${missingSteps}`);
          return;
        }
        toast.error(err?.message ?? "Erro ao concluir onboarding");
      },
    });
  }

  function renderAreaIcon(area: { icon: string; color: string }) {
    const Icon = areaIconMap[area.icon] ?? Target;
    const { bg, text } = colorTw[area.color] ?? { bg: "bg-accent-blue/10", text: "text-accent-blue" };
    return (
      <div className={cn("flex h-8 w-8 items-center justify-center rounded-lg", bg, text)}>
        <Icon className="h-4 w-4" />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-lg py-12">
      <StepProgress current={currentStep} total={6} />

      {currentStep === 0 && (
        <div>
          <h2 className="mb-1 text-2xl font-bold">Escolha suas áreas de vida</h2>
          <p className="mb-6 font-serif text-sm text-text-secondary">Selecione as áreas que quer acompanhar e adicione as suas.</p>
          <div className="grid grid-cols-2 gap-3">
            {areaTemplates.map((area) => (
              <button
                key={area.slug}
                onClick={() => toggleArea(area.slug)}
                className={cn(
                  "flex items-center gap-3 rounded-[14px] border p-4 text-left transition-colors",
                  selectedSlugs.has(area.slug) ? "border-accent-orange bg-accent-orange/5" : "border-border bg-bg-card",
                )}
              >
                {renderAreaIcon(area)}
                <span className="text-sm font-medium">{area.name}</span>
              </button>
            ))}
          </div>

          {customAreas.length > 0 && (
            <div className="mt-4 space-y-2">
              {customAreas.map((area) => (
                <div key={area.id} className="flex items-center justify-between rounded-lg border border-border bg-bg-card px-3 py-2">
                  <div className="flex items-center gap-2">
                    {renderAreaIcon(area)}
                    <span className="text-sm">{area.name}</span>
                  </div>
                  <button onClick={() => removeCustomArea(area.id)} className="text-text-muted hover:text-accent-rose">
                    <X className="h-3.5 w-3.5" />
                  </button>
                </div>
              ))}
            </div>
          )}

          <Button variant="outline" className="mt-4 w-full" onClick={openCustomModal}>
            <Plus className="h-4 w-4" /> Adicionar área personalizada
          </Button>

          <Button className="mt-3 w-full" disabled={selectedAreasForSubmit.length === 0 || submitAreas.isPending} onClick={handleAreasSubmit}>
            Continuar
          </Button>
        </div>
      )}

      {currentStep === 1 && (
        <div>
          <h2 className="mb-1 text-2xl font-bold">Defina suas primeiras metas</h2>
          <p className="mb-6 font-serif text-sm text-text-secondary">Cada meta precisa estar vinculada a uma área.</p>

          {selectedAreasForSubmit.length > 0 && (
            <div className="mb-6 space-y-4">
              {selectedAreasForSubmit.map((area) => {
                const linkedArea = onboardingAreaBySlug.get(area.slug);
                return (
                  <div key={`${area.slug}-${area.name}`}>
                    <div className="mb-2 flex items-center gap-2">
                      {renderAreaIcon(area)}
                      <span className="text-sm font-medium">{area.name}</span>
                    </div>
                    {linkedArea ? (
                      <GoalSuggestionsForArea
                        areaSlug={area.slug}
                        onAdd={(title, period) => addGoal(linkedArea.id, title, period)}
                      />
                    ) : null}
                  </div>
                );
              })}
            </div>
          )}

          <div className="mb-3 flex gap-2">
            <Input
              value={goalTitle}
              onChange={(e) => setGoalTitle(e.target.value)}
              placeholder="Meta personalizada..."
              onKeyDown={(e) => {
                if (e.key === "Enter" && goalTitle.trim() && goalAreaId) {
                  addGoal(goalAreaId, goalTitle, "quarterly");
                  setGoalTitle("");
                }
              }}
            />
            <Select value={goalAreaId} onValueChange={setGoalAreaId}>
              <SelectTrigger className="w-44">
                <SelectValue placeholder="Área" />
              </SelectTrigger>
              <SelectContent>
                {onboardingAreas.map((area) => (
                  <SelectItem key={area.id} value={area.id}>{area.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              size="sm"
              onClick={() => {
                if (goalTitle.trim() && goalAreaId) {
                  addGoal(goalAreaId, goalTitle, "quarterly");
                  setGoalTitle("");
                }
              }}
            >
              Adicionar
            </Button>
          </div>

          {goals.length > 0 && (
            <div className="space-y-2">
              {goals.map((g, i) => (
                <div key={`${g.area_id}:${g.title}:${i}`} className="flex items-center justify-between rounded-lg border border-border bg-bg-card px-4 py-2 text-sm">
                  <div className="flex items-center gap-2">
                    <span>{g.title}</span>
                    <span className="rounded-full bg-bg-tertiary px-2 py-0.5 text-xs text-text-muted">{g.period}</span>
                    <span className="rounded-full bg-accent-blue/10 px-2 py-0.5 text-xs text-accent-blue">
                      {onboardingAreaByID.get(g.area_id)?.name ?? "Área"}
                    </span>
                  </div>
                  <button onClick={() => setGoals((prev) => prev.filter((_, j) => j !== i))} className="text-text-muted hover:text-accent-rose">
                    <X className="h-3.5 w-3.5" />
                  </button>
                </div>
              ))}
            </div>
          )}

          <Button className="mt-6 w-full" disabled={goals.length === 0 || submitGoals.isPending} onClick={handleGoalsSubmit}>
            Continuar
          </Button>
        </div>
      )}

      {currentStep === 2 && (
        <div>
          <h2 className="mb-1 text-2xl font-bold">Crie seus hábitos</h2>
          <p className="mb-6 font-serif text-sm text-text-secondary">Opcional: você pode pular por agora.</p>
          <div className="mb-4 flex gap-2">
            <Input
              value={habitName}
              onChange={(e) => setHabitName(e.target.value)}
              placeholder="Ex: Meditar 10 min"
              onKeyDown={(e) => {
                if (e.key === "Enter" && habitName.trim()) {
                  setHabits((prev) => [...prev, habitName.trim()]);
                  setHabitName("");
                }
              }}
            />
            <Button
              size="sm"
              onClick={() => {
                if (habitName.trim()) {
                  setHabits((prev) => [...prev, habitName.trim()]);
                  setHabitName("");
                }
              }}
            >
              Adicionar
            </Button>
          </div>
          <div className="space-y-2">
            {habits.map((h, i) => (
              <div key={`${h}-${i}`} className="flex items-center justify-between rounded-lg border border-border bg-bg-card px-4 py-2 text-sm">
                <div className="flex items-center gap-2"><Activity className="h-3.5 w-3.5 text-accent-blue" />{h}</div>
                <button onClick={() => setHabits((prev) => prev.filter((_, j) => j !== i))} className="text-text-muted hover:text-accent-rose">
                  <X className="h-3.5 w-3.5" />
                </button>
              </div>
            ))}
          </div>
          <div className="mt-6 grid grid-cols-2 gap-3">
            <Button variant="outline" disabled={skipHabits.isPending} onClick={handleSkipHabits}>Pular por agora</Button>
            <Button disabled={habits.length === 0 || submitHabits.isPending} onClick={handleHabitsSubmit}>Continuar</Button>
          </div>
        </div>
      )}

      {currentStep === 3 && (
        <div>
          <h2 className="mb-1 text-2xl font-bold">Crie seu primeiro projeto</h2>
          <p className="mb-6 font-serif text-sm text-text-secondary">Conecte projeto, área e meta para organizar a execução.</p>

          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-xs text-text-muted">Título do projeto</label>
              <Input value={projectTitle} onChange={(e) => setProjectTitle(e.target.value)} placeholder="Ex: Projeto de saúde 90 dias" />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <label className="text-xs text-text-muted">Área</label>
                <Select value={projectAreaId} onValueChange={(v) => { setProjectAreaId(v); setProjectGoalId(""); }}>
                  <SelectTrigger><SelectValue placeholder="Selecione" /></SelectTrigger>
                  <SelectContent>
                    {onboardingAreas.map((area) => (
                      <SelectItem key={area.id} value={area.id}>{area.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <label className="text-xs text-text-muted">Meta (opcional)</label>
                <Select value={projectGoalId || "__none__"} onValueChange={(v) => setProjectGoalId(v === "__none__" ? "" : v)}>
                  <SelectTrigger><SelectValue placeholder="Nenhuma" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__none__">Nenhuma</SelectItem>
                    {projectGoals.map((goal) => (
                      <SelectItem key={goal.id} value={goal.id}>{goal.title}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>

          <Button className="mt-6 w-full" disabled={!projectTitle.trim() || submitProject.isPending} onClick={handleProjectSubmit}>
            Continuar
          </Button>
        </div>
      )}

      {currentStep === 4 && (
        <div>
          <h2 className="mb-1 text-2xl font-bold">Crie a primeira tarefa</h2>
          <p className="mb-6 font-serif text-sm text-text-secondary">Finalizamos a estrutura com sua primeira execução.</p>

          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-xs text-text-muted">Projeto</label>
              <Select value={firstTaskProjectId} onValueChange={setFirstTaskProjectId}>
                <SelectTrigger><SelectValue placeholder="Selecione um projeto" /></SelectTrigger>
                <SelectContent>
                  {onboardingProjects.map((project) => (
                    <SelectItem key={project.id} value={project.id}>{project.title}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <label className="text-xs text-text-muted">Título da tarefa</label>
              <Input
                value={firstTaskTitle}
                onChange={(e) => setFirstTaskTitle(e.target.value)}
                placeholder="Ex: Definir escopo e próximas 3 ações"
                onKeyDown={(e) => {
                  if (e.key === "Enter") handleFirstTaskSubmit();
                }}
              />
            </div>
          </div>

          <Button className="mt-6 w-full" disabled={!firstTaskProjectId || !firstTaskTitle.trim() || submitFirstTask.isPending} onClick={handleFirstTaskSubmit}>
            Continuar
          </Button>
        </div>
      )}

      {currentStep === 5 && (
        <div className="py-12 text-center">
          <CheckCircle2 className="mx-auto mb-4 h-16 w-16 text-accent-green" />
          <h2 className="mb-2 text-2xl font-bold">Tudo pronto!</h2>
          <p className="mb-8 font-serif text-text-secondary">Você concluiu a jornada com área, meta, projeto e tarefa.</p>
          <Button size="lg" onClick={handleComplete} disabled={complete.isPending}>
            Ir para o Dashboard
          </Button>
        </div>
      )}

      <Dialog open={customModalOpen} onOpenChange={setCustomModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Nova área personalizada</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-xs text-text-muted">Nome</label>
              <Input value={customName} onChange={(e) => setCustomName(e.target.value)} placeholder="Ex: Estudos" />
            </div>

            <div className="space-y-2">
              <label className="text-xs text-text-muted">Ícone</label>
              <Select value={customIcon} onValueChange={setCustomIcon}>
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  {onboardingIconOptions.map((iconKey) => (
                    <SelectItem key={iconKey} value={iconKey}>{iconKey}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <label className="text-xs text-text-muted">Cor</label>
              <div className="flex flex-wrap gap-2">
                {onboardingColorOptions.map((color) => (
                  <button
                    key={color}
                    type="button"
                    onClick={() => setCustomColor(color)}
                    className={cn(
                      "h-7 w-7 rounded-full border-2",
                      customColor === color ? "border-text-primary" : "border-transparent",
                      `bg-accent-${color}`,
                    )}
                  />
                ))}
              </div>
            </div>

            <Button className="w-full" onClick={addCustomArea} disabled={!customName.trim()}>
              Adicionar área
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
