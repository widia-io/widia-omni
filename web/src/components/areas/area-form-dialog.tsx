import { type ReactNode, useEffect, useMemo, useState } from "react";
import { Check, ChevronDown, ChevronRight } from "lucide-react";
import { toast } from "sonner";
import { useCreateArea, useUpdateArea } from "@/hooks/use-areas";
import { useWorkspaceUsage } from "@/hooks/use-settings";
import { getAreaIcon, getAreaIconWithFallback, resolveAreaIconKey } from "@/lib/icons";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { IconPicker } from "@/components/ui/icon-picker";
import { cn } from "@/lib/cn";
import type { LifeArea } from "@/types/api";

type AreaFormDialogProps = {
  area?: LifeArea;
  onClose: () => void;
};

const COLOR_OPTIONS = [
  {
    value: "green",
    label: "Verde",
    bg: "bg-accent-green-soft",
    text: "text-accent-green",
    swatch: "bg-accent-green",
    bar: "from-accent-green to-accent-sage",
  },
  {
    value: "orange",
    label: "Laranja",
    bg: "bg-accent-orange-soft",
    text: "text-accent-orange",
    swatch: "bg-accent-orange",
    bar: "from-accent-orange to-accent-sand",
  },
  {
    value: "blue",
    label: "Azul",
    bg: "bg-accent-blue-soft",
    text: "text-accent-blue",
    swatch: "bg-accent-blue",
    bar: "from-accent-blue to-accent-sage",
  },
  {
    value: "rose",
    label: "Rosa",
    bg: "bg-accent-rose-soft",
    text: "text-accent-rose",
    swatch: "bg-accent-rose",
    bar: "from-accent-rose to-accent-orange",
  },
  {
    value: "sand",
    label: "Areia",
    bg: "bg-accent-sand-soft",
    text: "text-accent-sand",
    swatch: "bg-accent-sand",
    bar: "from-accent-sand to-accent-blue",
  },
  {
    value: "sage",
    label: "Sálvia",
    bg: "bg-accent-sage-soft",
    text: "text-accent-sage",
    swatch: "bg-accent-sage",
    bar: "from-accent-sage to-accent-green",
  },
] as const;

function getAreaDisplayName(name: string, slug?: string) {
  const trimmedName = name.trim();
  if (trimmedName) return trimmedName;

  const normalizedSlug = (slug ?? "").trim();
  if (!normalizedSlug) return "Área sem nome";

  return normalizedSlug
    .split("-")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function SidebarField({ label, children, noBorder }: { label: string; children: ReactNode; noBorder?: boolean }) {
  return (
    <div className={cn("px-4 py-3", !noBorder && "border-b border-border/40")}>
      <div className="mb-1.5 text-[11px] font-medium uppercase tracking-wider text-text-muted">{label}</div>
      {children}
    </div>
  );
}

function getAreaMutationErrorMessage(
  err: unknown,
  action: "criar" | "atualizar",
  options?: { areaLimitReached?: boolean },
) {
  if (options?.areaLimitReached && action === "criar") {
    return "Limite de áreas atingido. Faça upgrade para criar mais.";
  }

  if (err instanceof Error) {
    if (err.message === "area limit reached") {
      return "Limite de áreas atingido. Faça upgrade para criar mais.";
    }
    if (err.message === "workspace not found") {
      return "Workspace não encontrado. Atualize a página e tente novamente.";
    }
    if (err.message === "entitlements not loaded") {
      return "Não foi possível validar os limites do plano. Tente novamente em instantes.";
    }
    if (err.message) return err.message;
  }

  if (
    action === "criar" &&
    typeof err === "object" &&
    err !== null &&
    "status" in err &&
    (err as { status?: number }).status === 403 &&
    options?.areaLimitReached
  ) {
    return "Limite de áreas atingido. Faça upgrade para criar mais.";
  }

  return action === "criar" ? "Erro ao criar área" : "Erro ao atualizar área";
}

export function AreaFormDialog({ area, onClose }: AreaFormDialogProps) {
  const create = useCreateArea();
  const update = useUpdateArea();
  const { data: usage } = useWorkspaceUsage();

  const [name, setName] = useState(area?.name ?? "");
  const [icon, setIcon] = useState(resolveAreaIconKey(area?.icon) ?? "heart");
  const [color, setColor] = useState(area?.color ?? "orange");
  const [weight, setWeight] = useState(String(area?.weight ?? 1));
  const [isActive, setIsActive] = useState(area?.is_active ?? true);
  const [showAdvanced, setShowAdvanced] = useState(Boolean(area && area.weight !== 1));

  const isEditing = Boolean(area);
  const isPending = create.isPending || update.isPending;
  const normalizedName = name.trim();
  const maxAreas = usage?.limits.max_areas ?? null;
  const usedAreas = usage?.counters.areas_count ?? null;
  const isUnlimited = maxAreas === -1;
  const areaLimitReached = !isEditing && Boolean(usage) && !isUnlimited && maxAreas !== null && usedAreas !== null && usedAreas >= maxAreas;
  const canSubmit = normalizedName.length > 0 && !isPending && !areaLimitReached;

  const selectedColor = useMemo(() => {
    return COLOR_OPTIONS.find((option) => option.value === color) ?? COLOR_OPTIONS[1];
  }, [color]);

  const iconKey = resolveAreaIconKey(icon) ?? icon;
  const Icon = getAreaIcon(iconKey);
  const FallbackIcon = getAreaIconWithFallback({ icon: iconKey, name: normalizedName, color });
  const displayName = getAreaDisplayName(normalizedName, area?.slug);

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key !== "Escape" || isPending) return;
      e.preventDefault();
      onClose();
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [isPending, onClose]);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!normalizedName || isPending) return;
    if (!isEditing && areaLimitReached) {
      toast.error("Limite de áreas atingido. Faça upgrade para criar mais.");
      return;
    }

    const slug = normalizedName
      .toLowerCase()
      .replace(/\s+/g, "-")
      .replace(/[^a-z0-9-]/g, "");
    const parsedWeight = Number(weight);
    const safeWeight = Number.isFinite(parsedWeight) ? parsedWeight : 1;

    const payload = {
      name: normalizedName,
      slug,
      icon: iconKey,
      color,
      weight: safeWeight,
      is_active: isActive,
    };

    if (isEditing && area) {
      update.mutate(
        { id: area.id, ...payload },
        {
          onSuccess: () => {
            toast.success("Área atualizada");
            onClose();
          },
          onError: (err) => toast.error(getAreaMutationErrorMessage(err, "atualizar")),
        },
      );
      return;
    }

    create.mutate(payload, {
      onSuccess: () => {
        toast.success("Área criada");
        onClose();
      },
      onError: (err) => toast.error(getAreaMutationErrorMessage(err, "criar", { areaLimitReached })),
    });
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col">
      <div className="flex min-h-[360px]">
        <div className="flex flex-1 flex-col gap-4 p-6">
          <div className="space-y-2">
            <label className="text-xs font-medium uppercase tracking-wider text-text-muted">Nome da área</label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Ex.: Saúde e energia"
              autoFocus
              required
              className="h-10 border-border/60"
            />
          </div>

          <div className="overflow-hidden rounded-[14px] border border-border/60 bg-bg-card">
            <div className={cn("h-1 bg-gradient-to-r", selectedColor.bar)} />
            <div className="flex items-start justify-between p-4">
              <div className="flex items-center gap-3">
                <div className={cn("flex h-10 w-10 items-center justify-center rounded-[10px]", selectedColor.bg)}>
                  {Icon ? <Icon size={20} className={selectedColor.text} /> : <FallbackIcon size={20} className={selectedColor.text} />}
                </div>
                <div>
                  <p className="text-sm font-semibold text-text-primary">{displayName}</p>
                  <p className="text-xs text-text-muted">Pré-visualização da área</p>
                </div>
              </div>
              <span
                className={cn(
                  "rounded-full border px-2 py-0.5 text-[11px] font-medium",
                  isActive
                    ? "border-accent-green/25 bg-accent-green/10 text-accent-green"
                    : "border-border bg-bg-secondary text-text-muted",
                )}
              >
                {isActive ? "Ativa" : "Inativa"}
              </span>
            </div>
          </div>

          <div className="border-t border-border/40 pt-3">
            <button
              type="button"
              onClick={() => setShowAdvanced((prev) => !prev)}
              className="flex items-center gap-1.5 text-xs font-medium text-text-muted transition-colors hover:text-text-primary"
            >
              {showAdvanced ? <ChevronDown className="h-3.5 w-3.5" /> : <ChevronRight className="h-3.5 w-3.5" />}
              Avançado
            </button>

            {showAdvanced && (
              <div className="mt-3 rounded-[10px] border border-border/50 bg-bg-secondary/20 p-3">
                <label className="mb-1 block text-xs font-medium text-text-secondary">Peso</label>
                <Input
                  type="number"
                  min="0"
                  max="10"
                  step="0.1"
                  value={weight}
                  onChange={(e) => setWeight(e.target.value)}
                  className="h-9 border-border/60 text-sm"
                />
                <p className="mt-2 text-[11px] text-text-muted">Use para priorização analítica da área (0 a 10).</p>
              </div>
            )}
          </div>
        </div>

        <div className="w-[260px] shrink-0 border-l border-border/60 bg-bg-secondary/20">
          <SidebarField label="Ícone">
            <IconPicker value={iconKey} onChange={setIcon} />
          </SidebarField>

          <SidebarField label="Cor">
            <div className="grid grid-cols-3 gap-2">
              {COLOR_OPTIONS.map((option) => (
                <button
                  key={option.value}
                  type="button"
                  onClick={() => setColor(option.value)}
                  className={cn(
                    "rounded-lg border px-2 py-2 text-center text-[11px] transition-colors",
                    color === option.value
                      ? "border-text-primary bg-bg-card text-text-primary"
                      : "border-border/60 text-text-muted hover:border-border hover:text-text-secondary",
                  )}
                >
                  <span className={cn("mx-auto mb-1 block h-3 w-3 rounded-full", option.swatch)} />
                  <span className="inline-flex items-center gap-1">
                    {color === option.value && <Check className="h-3 w-3" />}
                    {option.label}
                  </span>
                </button>
              ))}
            </div>
          </SidebarField>

          <SidebarField label="Status" noBorder>
            <div className="grid grid-cols-2 gap-1.5">
              <button
                type="button"
                onClick={() => setIsActive(true)}
                className={cn(
                  "rounded-lg border px-2 py-1.5 text-xs font-medium transition-colors",
                  isActive
                    ? "border-accent-green/25 bg-accent-green/10 text-accent-green"
                    : "border-border/60 text-text-muted hover:border-border hover:text-text-secondary",
                )}
              >
                Ativa
              </button>
              <button
                type="button"
                onClick={() => setIsActive(false)}
                className={cn(
                  "rounded-lg border px-2 py-1.5 text-xs font-medium transition-colors",
                  !isActive
                    ? "border-accent-rose/25 bg-accent-rose/10 text-accent-rose"
                    : "border-border/60 text-text-muted hover:border-border hover:text-text-secondary",
                )}
              >
                Inativa
              </button>
            </div>
          </SidebarField>
        </div>
      </div>

      <div className="flex items-center justify-between border-t border-border/50 px-6 py-4">
        <div className="text-xs text-text-muted">
          <span>Enter para salvar · Esc para fechar</span>
          {!isEditing && usage && !isUnlimited && maxAreas !== null && usedAreas !== null && (
            <p className={cn("mt-1", areaLimitReached ? "text-accent-rose" : "text-text-muted")}>
              Áreas: {usedAreas}/{maxAreas}
            </p>
          )}
        </div>
        <Button type="submit" disabled={!canSubmit}>
          {isEditing ? "Salvar área" : "Criar área"}
        </Button>
      </div>
    </form>
  );
}
