import { useState } from "react";
import { useNavigate } from "react-router";
import { Target, Activity, CheckCircle2, Plus, X } from "lucide-react";
import {
  useOnboardingStatus, useOnboardingAreas, useOnboardingGoals,
  useOnboardingHabits, useOnboardingComplete, useAreaTemplates, useGoalSuggestions,
} from "@/hooks/use-onboarding";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/cn";
import { areaIconMap } from "@/lib/icons";
import type { AreaTemplate, GoalPeriod } from "@/types/api";

const colorTw: Record<string, { bg: string; text: string }> = {
  green:  { bg: "bg-accent-green/10",  text: "text-accent-green" },
  orange: { bg: "bg-accent-orange/10", text: "text-accent-orange" },
  sand:   { bg: "bg-accent-sand/10",   text: "text-accent-sand" },
  rose:   { bg: "bg-accent-rose/10",   text: "text-accent-rose" },
  sage:   { bg: "bg-accent-sage/10",   text: "text-accent-sage" },
  blue:   { bg: "bg-accent-blue/10",   text: "text-accent-blue" },
  sky:    { bg: "bg-accent-sky/10",    text: "text-accent-sky" },
  violet: { bg: "bg-accent-violet/10", text: "text-accent-violet" },
};

function StepProgress({ current, total }: { current: number; total: number }) {
  return (
    <div className="mb-8 flex items-center gap-2">
      {Array.from({ length: total }, (_, i) => (
        <div key={i} className={cn("h-1.5 flex-1 rounded-full transition-colors", i <= current ? "bg-accent-orange" : "bg-border")} />
      ))}
    </div>
  );
}

function GoalSuggestionsForArea({ areaSlug, onAdd }: { areaSlug: string; onAdd: (title: string, period: GoalPeriod) => void }) {
  const { data: suggestions } = useGoalSuggestions("pt-BR", areaSlug);
  if (!suggestions || suggestions.length === 0) return null;

  return (
    <div className="flex flex-wrap gap-2">
      {suggestions.map((s) => (
        <button
          key={s.title}
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
  const submitAreas = useOnboardingAreas();
  const submitGoals = useOnboardingGoals();
  const submitHabits = useOnboardingHabits();
  const complete = useOnboardingComplete();

  const [step, setStep] = useState<number | null>(null);
  const [selectedSlugs, setSelectedSlugs] = useState<Set<string>>(new Set());
  const [goalTitle, setGoalTitle] = useState("");
  const [goals, setGoals] = useState<Array<{ title: string; period: GoalPeriod }>>([]);
  const [habitName, setHabitName] = useState("");
  const [habits, setHabits] = useState<string[]>([]);

  if (isLoading || templatesLoading) return <Skeleton className="h-96 rounded-[14px]" />;

  const areas = templates ?? [];
  const derivedStep = status?.steps?.habits ? 3 : status?.steps?.goals ? 2 : status?.steps?.areas ? 1 : 0;
  const currentStep = step ?? derivedStep;

  const selectedAreas = areas.filter((a) => selectedSlugs.has(a.slug));

  function toggleArea(slug: string) {
    setSelectedSlugs((prev) => {
      const next = new Set(prev);
      next.has(slug) ? next.delete(slug) : next.add(slug);
      return next;
    });
  }

  function handleAreasSubmit() {
    const data = selectedAreas.map((a) => ({ name: a.name, icon: a.icon, color: a.color, weight: a.weight }));
    submitAreas.mutate(data, { onSuccess: () => setStep(1) });
  }

  function addGoal(title: string, period: GoalPeriod) {
    if (goals.some((g) => g.title === title)) return;
    setGoals((p) => [...p, { title, period }]);
  }

  function handleGoalsSubmit() {
    submitGoals.mutate(goals, { onSuccess: () => setStep(2) });
  }

  function handleHabitsSubmit() {
    const data = habits.map((name) => ({ name, frequency: "daily" as const, target_per_week: 7 }));
    submitHabits.mutate(data, { onSuccess: () => setStep(3) });
  }

  function handleComplete() {
    complete.mutate(undefined, { onSuccess: () => navigate("/dashboard") });
  }

  function renderAreaIcon(area: AreaTemplate) {
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
      <StepProgress current={currentStep} total={4} />

      {currentStep === 0 && (
        <div>
          <h2 className="mb-1 text-2xl font-bold">Escolha suas áreas de vida</h2>
          <p className="mb-6 font-serif text-sm text-text-secondary">Selecione as áreas que quer acompanhar.</p>
          <div className="grid grid-cols-2 gap-3">
            {areas.map((area) => (
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
          <Button className="mt-6 w-full" disabled={selectedSlugs.size === 0 || submitAreas.isPending} onClick={handleAreasSubmit}>
            Continuar
          </Button>
        </div>
      )}

      {currentStep === 1 && (
        <div>
          <h2 className="mb-1 text-2xl font-bold">Defina suas primeiras metas</h2>
          <p className="mb-6 font-serif text-sm text-text-secondary">Escolha sugestões ou adicione metas personalizadas.</p>

          {selectedAreas.length > 0 && (
            <div className="mb-6 space-y-4">
              {selectedAreas.map((area) => (
                <div key={area.slug}>
                  <div className="mb-2 flex items-center gap-2">
                    {renderAreaIcon(area)}
                    <span className="text-sm font-medium">{area.name}</span>
                  </div>
                  <GoalSuggestionsForArea areaSlug={area.slug} onAdd={addGoal} />
                </div>
              ))}
            </div>
          )}

          <div className="mb-4 flex gap-2">
            <Input value={goalTitle} onChange={(e) => setGoalTitle(e.target.value)} placeholder="Meta personalizada..." onKeyDown={(e) => {
              if (e.key === "Enter" && goalTitle.trim()) { addGoal(goalTitle.trim(), "quarterly"); setGoalTitle(""); }
            }} />
            <Button size="sm" onClick={() => { if (goalTitle.trim()) { addGoal(goalTitle.trim(), "quarterly"); setGoalTitle(""); } }}>
              Adicionar
            </Button>
          </div>

          {goals.length > 0 && (
            <div className="space-y-2">
              {goals.map((g, i) => (
                <div key={i} className="flex items-center justify-between rounded-lg border border-border bg-bg-card px-4 py-2 text-sm">
                  <div className="flex items-center gap-2">
                    <span>{g.title}</span>
                    <span className="rounded-full bg-bg-tertiary px-2 py-0.5 text-xs text-text-muted">{g.period}</span>
                  </div>
                  <button onClick={() => setGoals((p) => p.filter((_, j) => j !== i))} className="text-text-muted hover:text-accent-rose">
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
          <p className="mb-6 font-serif text-sm text-text-secondary">Quais hábitos quer cultivar no dia a dia?</p>
          <div className="mb-4 flex gap-2">
            <Input value={habitName} onChange={(e) => setHabitName(e.target.value)} placeholder="Ex: Meditar 10 min" onKeyDown={(e) => {
              if (e.key === "Enter" && habitName.trim()) { setHabits((p) => [...p, habitName.trim()]); setHabitName(""); }
            }} />
            <Button size="sm" onClick={() => { if (habitName.trim()) { setHabits((p) => [...p, habitName.trim()]); setHabitName(""); } }}>
              Adicionar
            </Button>
          </div>
          <div className="space-y-2">
            {habits.map((h, i) => (
              <div key={i} className="flex items-center justify-between rounded-lg border border-border bg-bg-card px-4 py-2 text-sm">
                <div className="flex items-center gap-2"><Activity className="h-3.5 w-3.5 text-accent-blue" />{h}</div>
                <button onClick={() => setHabits((p) => p.filter((_, j) => j !== i))} className="text-text-muted hover:text-accent-rose">
                  <X className="h-3.5 w-3.5" />
                </button>
              </div>
            ))}
          </div>
          <Button className="mt-6 w-full" disabled={habits.length === 0 || submitHabits.isPending} onClick={handleHabitsSubmit}>
            Continuar
          </Button>
        </div>
      )}

      {currentStep === 3 && (
        <div className="text-center py-12">
          <CheckCircle2 className="mx-auto mb-4 h-16 w-16 text-accent-green" />
          <h2 className="mb-2 text-2xl font-bold">Tudo pronto!</h2>
          <p className="mb-8 font-serif text-text-secondary">Sua conta está configurada. Vamos ao seu dashboard.</p>
          <Button size="lg" onClick={handleComplete} disabled={complete.isPending}>
            Ir para o Dashboard
          </Button>
        </div>
      )}
    </div>
  );
}
