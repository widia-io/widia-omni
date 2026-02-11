import { useState } from "react";
import { useNavigate } from "react-router";
import { Target, Activity, CheckCircle2 } from "lucide-react";
import { useOnboardingStatus, useOnboardingAreas, useOnboardingGoals, useOnboardingHabits, useOnboardingComplete } from "@/hooks/use-onboarding";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/cn";

const SUGGESTED_AREAS = [
  { name: "Saúde", icon: "heart", color: "green", tw: { bg: "bg-accent-green/10", text: "text-accent-green" } },
  { name: "Carreira", icon: "briefcase", color: "orange", tw: { bg: "bg-accent-orange/10", text: "text-accent-orange" } },
  { name: "Finanças", icon: "dollar-sign", color: "sand", tw: { bg: "bg-accent-sand/10", text: "text-accent-sand" } },
  { name: "Relacionamentos", icon: "users", color: "blue", tw: { bg: "bg-accent-blue/10", text: "text-accent-blue" } },
  { name: "Desenvolvimento Pessoal", icon: "book", color: "sage", tw: { bg: "bg-accent-sage/10", text: "text-accent-sage" } },
  { name: "Lazer", icon: "sun", color: "rose", tw: { bg: "bg-accent-rose/10", text: "text-accent-rose" } },
];

function StepProgress({ current, total }: { current: number; total: number }) {
  return (
    <div className="mb-8 flex items-center gap-2">
      {Array.from({ length: total }, (_, i) => (
        <div key={i} className={cn("h-1.5 flex-1 rounded-full transition-colors", i <= current ? "bg-accent-orange" : "bg-border")} />
      ))}
    </div>
  );
}

export function Component() {
  const navigate = useNavigate();
  const { data: status, isLoading } = useOnboardingStatus();
  const submitAreas = useOnboardingAreas();
  const submitGoals = useOnboardingGoals();
  const submitHabits = useOnboardingHabits();
  const complete = useOnboardingComplete();

  const [step, setStep] = useState<number | null>(null);

  const [selectedAreas, setSelectedAreas] = useState<Set<number>>(new Set());
  const [goalTitle, setGoalTitle] = useState("");
  const [goals, setGoals] = useState<string[]>([]);
  const [habitName, setHabitName] = useState("");
  const [habits, setHabits] = useState<string[]>([]);

  if (isLoading) return <Skeleton className="h-96 rounded-[14px]" />;

  const derivedStep = status?.steps?.habits ? 3 : status?.steps?.goals ? 2 : status?.steps?.areas ? 1 : 0;
  const currentStep = step ?? derivedStep;

  function handleAreasSubmit() {
    const areas = [...selectedAreas]
      .map((i) => SUGGESTED_AREAS[i])
      .filter((a): a is (typeof SUGGESTED_AREAS)[number] => !!a)
      .map((a) => ({ name: a.name, icon: a.icon, color: a.color, weight: 1 }));
    submitAreas.mutate(areas, { onSuccess: () => setStep(1) });
  }

  function handleGoalsSubmit() {
    const data = goals.map((title) => ({ title, period: "quarterly" as const }));
    submitGoals.mutate(data, { onSuccess: () => setStep(2) });
  }

  function handleHabitsSubmit() {
    const data = habits.map((name) => ({ name, frequency: "daily" as const, target_per_week: 7 }));
    submitHabits.mutate(data, { onSuccess: () => setStep(3) });
  }

  function handleComplete() {
    complete.mutate(undefined, { onSuccess: () => navigate("/dashboard") });
  }

  return (
    <div className="mx-auto max-w-lg py-12">
      <StepProgress current={currentStep} total={4} />

      {currentStep === 0 && (
        <div>
          <h2 className="mb-1 text-2xl font-bold">Escolha suas áreas de vida</h2>
          <p className="mb-6 font-serif text-sm text-text-secondary">Selecione as áreas que quer acompanhar.</p>
          <div className="grid grid-cols-2 gap-3">
            {SUGGESTED_AREAS.map((area, i) => (
              <button
                key={area.name}
                onClick={() => setSelectedAreas((prev) => { const next = new Set(prev); next.has(i) ? next.delete(i) : next.add(i); return next; })}
                className={cn(
                  "flex items-center gap-3 rounded-[14px] border p-4 text-left transition-colors",
                  selectedAreas.has(i) ? "border-accent-orange bg-accent-orange/5" : "border-border bg-bg-card",
                )}
              >
                <div className={cn("flex h-8 w-8 items-center justify-center rounded-lg", area.tw.bg, area.tw.text)}>
                  <Target className="h-4 w-4" />
                </div>
                <span className="text-sm font-medium">{area.name}</span>
              </button>
            ))}
          </div>
          <Button className="mt-6 w-full" disabled={selectedAreas.size === 0 || submitAreas.isPending} onClick={handleAreasSubmit}>
            Continuar
          </Button>
        </div>
      )}

      {currentStep === 1 && (
        <div>
          <h2 className="mb-1 text-2xl font-bold">Defina suas primeiras metas</h2>
          <p className="mb-6 font-serif text-sm text-text-secondary">Adicione pelo menos uma meta para começar.</p>
          <div className="mb-4 flex gap-2">
            <Input value={goalTitle} onChange={(e) => setGoalTitle(e.target.value)} placeholder="Ex: Ler 12 livros" />
            <Button size="sm" onClick={() => { if (goalTitle.trim()) { setGoals((p) => [...p, goalTitle.trim()]); setGoalTitle(""); } }}>
              Adicionar
            </Button>
          </div>
          <div className="space-y-2">
            {goals.map((g, i) => (
              <div key={i} className="flex items-center justify-between rounded-lg border border-border bg-bg-card px-4 py-2 text-sm">
                <span>{g}</span>
                <button onClick={() => setGoals((p) => p.filter((_, j) => j !== i))} className="text-text-muted hover:text-accent-rose">x</button>
              </div>
            ))}
          </div>
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
            <Input value={habitName} onChange={(e) => setHabitName(e.target.value)} placeholder="Ex: Meditar 10 min" />
            <Button size="sm" onClick={() => { if (habitName.trim()) { setHabits((p) => [...p, habitName.trim()]); setHabitName(""); } }}>
              Adicionar
            </Button>
          </div>
          <div className="space-y-2">
            {habits.map((h, i) => (
              <div key={i} className="flex items-center justify-between rounded-lg border border-border bg-bg-card px-4 py-2 text-sm">
                <div className="flex items-center gap-2"><Activity className="h-3.5 w-3.5 text-accent-blue" />{h}</div>
                <button onClick={() => setHabits((p) => p.filter((_, j) => j !== i))} className="text-text-muted hover:text-accent-rose">x</button>
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
