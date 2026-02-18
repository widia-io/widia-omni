import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router";
import { Plus, LayoutGrid, Search, X, GripVertical } from "lucide-react";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import {
  DndContext,
  closestCenter,
  PointerSensor,
  KeyboardSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  rectSortingStrategy,
  useSortable,
  sortableKeyboardCoordinates,
  arrayMove,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { restrictToParentElement } from "@dnd-kit/modifiers";
import {
  getAreaIcon,
  getAreaIconWithFallback,
  isRawAreaIcon,
} from "@/lib/icons";
import { useAreas, useReorderArea } from "@/hooks/use-areas";
import { useWorkspaceUsage } from "@/hooks/use-settings";
import { AreaFormDialog } from "@/components/areas/area-form-dialog";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { cn } from "@/lib/cn";
import type { LifeArea, AreaWithStats, WorkspaceUsage } from "@/types/api";

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

function getAreaDisplayName(area: Pick<AreaWithStats, "name" | "slug"> | Pick<LifeArea, "name" | "slug">) {
  const name = area.name?.trim();
  if (name) return name;
  const slug = area.slug?.trim();
  if (!slug) return "Área sem nome";
  return slug
    .split("-")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function FilterChip({
  label, isActive, onClick,
}: {
  label: string;
  isActive: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium transition-colors",
        isActive
          ? "bg-accent-orange/10 text-accent-orange border-accent-orange/20"
          : "bg-bg-card border-border text-text-muted hover:border-text-muted",
      )}
    >
      {label}
    </button>
  );
}

function AreaUsageBadge({ usage }: { usage?: WorkspaceUsage }) {
  if (!usage) return null;

  const used = usage.counters.areas_count;
  const max = usage.limits.max_areas;
  const isUnlimited = max === -1;
  const ratio = isUnlimited ? 0 : max > 0 ? used / max : 0;
  const isFull = !isUnlimited && ratio >= 1;
  const isWarning = !isUnlimited && ratio >= 0.8 && !isFull;

  const accentColor = isFull
    ? "text-accent-rose"
    : isWarning
      ? "text-accent-orange"
      : "text-text-muted";

  const barColor = isFull
    ? "bg-accent-rose"
    : isWarning
      ? "bg-accent-orange"
      : "bg-accent-green";

  return (
    <div className="flex items-center gap-2.5 rounded-full border border-border/60 px-3 py-1">
      <div className="flex items-baseline gap-1">
        <span className={cn("font-mono text-xs font-semibold tabular-nums", accentColor)}>
          {used}
        </span>
        <span className="text-[10px] text-text-muted">/</span>
        <span className="font-mono text-[10px] text-text-muted">
          {isUnlimited ? "∞" : max}
        </span>
        <span className="text-[10px] text-text-muted">áreas</span>
      </div>
      {!isUnlimited && (
        <div className="h-1 w-14 overflow-hidden rounded-full bg-border/50">
          <div
            className={cn("h-full rounded-full transition-all duration-700 ease-out", barColor)}
            style={{ width: `${Math.min(ratio * 100, 100)}%` }}
          />
        </div>
      )}
    </div>
  );
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
        "mb-4 flex flex-wrap items-center justify-between gap-3 rounded-[12px] border px-4 py-3",
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

function SortableAreaCard({ area, index }: { area: AreaWithStats; index: number }) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: area.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    animationDelay: `${0.15 + index * 0.05}s`,
  };

  const navigate = useNavigate();
  const c = getColorClasses(area.color);
  const score = area.area_score ?? 0;
  const inactive = area.is_active === false;
  const Icon = getAreaIcon(area.icon);
  const FallbackIcon = getAreaIconWithFallback(area);
  const rawIcon = area.icon?.trim();
  const showRawIcon = isRawAreaIcon(rawIcon);
  const displayName = getAreaDisplayName(area);

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className={cn(
        "group relative cursor-pointer overflow-hidden rounded-[14px] border border-border bg-bg-card p-[18px_20px] transition-all duration-[250ms] hover:-translate-y-0.5 hover:border-transparent animate-in",
        isDragging && "z-50 shadow-lg opacity-90",
        inactive && "opacity-60",
      )}
      onClick={() => { if (!isDragging) navigate(`/areas/${area.id}`); }}
    >
      <div className={cn("absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r opacity-0 transition-opacity group-hover:opacity-100", c.bar)} />

      {inactive && (
        <span className="absolute top-2 right-2 rounded-full bg-border px-1.5 py-0.5 text-[10px] font-medium text-text-muted">
          Inativa
        </span>
      )}

      <div
        className="absolute top-2 left-2 z-10 flex h-8 w-8 items-center justify-center rounded opacity-40 transition-opacity group-hover:opacity-100 cursor-grab active:cursor-grabbing"
      >
        <GripVertical size={16} className="text-text-muted" />
      </div>

      <div className="flex items-center justify-between mb-3">
        <div className={cn("flex h-[34px] w-[34px] items-center justify-center rounded-[9px]", c.bg)}>
          {Icon ? <Icon size={18} className={c.text} /> : showRawIcon ? <span className="text-lg">{rawIcon}</span> : <FallbackIcon size={18} className={c.text} />}
        </div>
        <span className={cn("font-mono text-xl font-bold", c.text)}>{score}</span>
      </div>

      <div className="text-sm font-semibold">{displayName}</div>

      <div className="mt-2 text-xs text-text-muted">
        {area.tasks_pending} tarefas · {area.goals_count} metas · {area.projects_count} projetos
      </div>

      <div className="h-[3px] mt-3 overflow-hidden rounded-[3px] bg-border">
        <div
          className={cn("h-full rounded-[3px] bg-gradient-to-r transition-all duration-[1.5s]", c.bar)}
          style={{ width: `${score}%` }}
        />
      </div>
    </div>
  );
}

export function Component() {
  const navigate = useNavigate();
  const { data: areas, isLoading } = useAreas();
  const { data: usage } = useWorkspaceUsage();
  const qc = useQueryClient();
  const reorderArea = useReorderArea();
  const [editArea, setEditArea] = useState<LifeArea | undefined>();
  const [orderedAreas, setOrderedAreas] = useState<AreaWithStats[]>([]);
  const [isReordering, setIsReordering] = useState(false);
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [showInactive, setShowInactive] = useState(false);
  const maxAreas = usage?.limits.max_areas ?? null;
  const usedAreas = usage?.counters.areas_count ?? null;
  const isUnlimited = maxAreas === -1;
  const usageRatio = !isUnlimited && maxAreas && usedAreas !== null ? usedAreas / maxAreas : 0;
  const areaLimitReached = Boolean(usage) && !isUnlimited && maxAreas !== null && usedAreas !== null && usedAreas >= maxAreas;
  const areaLimitWarning = Boolean(usage) && !isUnlimited && maxAreas !== null && usedAreas !== null && usageRatio >= 0.8 && !areaLimitReached;
  const showPlanGate = areaLimitReached || areaLimitWarning;

  // Keyboard shortcut: N to open create dialog
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      if ((e.target as HTMLElement)?.isContentEditable) return;
      if (e.key === "n" || e.key === "N") {
        e.preventDefault();
        if (areaLimitReached) {
          toast.error("Limite de áreas atingido. Faça upgrade para criar mais.");
          return;
        }
        setOpen(true);
      }
    }
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [areaLimitReached]);

  useEffect(() => {
    setOrderedAreas(areas ?? []);
  }, [areas]);

  const filtered = orderedAreas.filter((a) => {
    if (!showInactive && !a.is_active) return false;
    if (search && !a.name.toLowerCase().includes(search.toLowerCase())) return false;
    return true;
  });

  const hasFilters = search !== "" || showInactive;
  const isEmpty = filtered.length === 0;

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  const handleDragEnd = useCallback(async (event: DragEndEvent) => {
    const { active, over } = event;
    if (isReordering || !over || active.id === over.id || !orderedAreas.length || !filtered.length) return;

    const filteredIDs = filtered.map((a) => a.id);
    const fromIndex = filteredIDs.findIndex((id) => id === active.id);
    const toIndex = filteredIDs.findIndex((id) => id === over.id);
    if (fromIndex === -1 || toIndex === -1) return;

    const reorderedFilteredIDs = arrayMove(filteredIDs, fromIndex, toIndex);
    const filteredByID = new Map(filtered.map((area) => [area.id, area]));
    const reorderedFilteredAreas = reorderedFilteredIDs
      .map((id) => filteredByID.get(id))
      .filter((area): area is AreaWithStats => Boolean(area));

    let cursor = 0;
    const reconstructed = orderedAreas.map((area) => {
      if (!filteredByID.has(area.id)) return area;
      const next = reorderedFilteredAreas[cursor];
      cursor += 1;
      return next ?? area;
    });

    const movedID = String(active.id);
    const newSortOrder = reconstructed.findIndex((a) => a.id === movedID) + 1;
    if (newSortOrder <= 0) return;

    const previous = orderedAreas;
    const optimistic = reconstructed.map((area, i) => ({ ...area, sort_order: i + 1 }));
    setOrderedAreas(optimistic);
    qc.setQueryData<AreaWithStats[]>(["areas"], optimistic);
    setIsReordering(true);

    try {
      await reorderArea.mutateAsync({ id: movedID, sort_order: newSortOrder });
    } catch {
      setOrderedAreas(previous);
      qc.setQueryData<AreaWithStats[]>(["areas"], previous);
      toast.error("Erro ao reordenar áreas");
    } finally {
      setIsReordering(false);
    }
  }, [orderedAreas, filtered, isReordering, reorderArea, qc]);

  if (isLoading) {
    return (
      <div>
        <div className="mb-6 flex items-center justify-between">
          <Skeleton className="h-8 w-40" />
          <Skeleton className="h-8 w-28" />
        </div>
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-32 rounded-[14px]" />)}
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-bold">Áreas de Vida</h1>
          <AreaUsageBadge usage={usage} />
        </div>
        <Dialog open={open} onOpenChange={(v) => { setOpen(v); if (!v) setEditArea(undefined); }}>
          <DialogTrigger asChild>
            <Button size="sm" disabled={areaLimitReached} title={areaLimitReached ? "Limite de áreas atingido" : undefined}>
              <Plus className="h-4 w-4" /> Nova área
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-3xl gap-0 overflow-hidden p-0">
            <DialogHeader className="sr-only"><DialogTitle>{editArea ? "Editar área" : "Nova área"}</DialogTitle></DialogHeader>
            <AreaFormDialog area={editArea} onClose={() => setOpen(false)} />
          </DialogContent>
        </Dialog>
      </div>

      {showPlanGate && maxAreas !== null && usedAreas !== null && (
        <AreaPlanGate
          used={usedAreas}
          max={maxAreas}
          isFull={areaLimitReached}
          onUpgrade={() => navigate("/billing")}
        />
      )}

      {/* Filter bar */}
      <div className="mb-4 flex flex-wrap items-center gap-2">
        <FilterChip label="Ativas" isActive={!showInactive} onClick={() => setShowInactive(false)} />
        <FilterChip label="Inativas" isActive={showInactive} onClick={() => setShowInactive(true)} />
        <div className="relative ml-1">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-text-muted" />
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar áreas..."
            className="w-44 rounded-full border border-border bg-bg-card py-1 pl-8 pr-2 text-xs outline-none placeholder:text-text-muted focus:border-text-muted"
          />
        </div>
        {hasFilters && (
          <button
            onClick={() => { setSearch(""); setShowInactive(false); }}
            className="inline-flex items-center gap-1 text-xs text-text-muted hover:text-text-primary transition-colors"
          >
            <X size={12} /> Limpar
          </button>
        )}
      </div>

      {/* Area grid / empty state */}
      {isEmpty ? (
        <div className="flex flex-col items-center gap-3 py-16 text-text-muted">
          <LayoutGrid className="h-10 w-10" />
          <div className="text-center">
            <p className="text-sm font-medium text-text-primary">Organize sua vida em áreas</p>
            <p className="mt-1 text-xs">Crie sua primeira área para começar.</p>
            <button
              onClick={() => {
                if (areaLimitReached) {
                  toast.error("Limite de áreas atingido. Faça upgrade para criar mais.");
                  return;
                }
                setOpen(true);
              }}
              className="mt-3 text-xs font-medium text-accent-orange transition-opacity hover:opacity-80 disabled:cursor-not-allowed disabled:opacity-50"
              disabled={areaLimitReached}
            >
              {areaLimitReached ? "Limite de áreas atingido" : "+ Nova área"}
            </button>
          </div>
        </div>
      ) : (
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragEnd={handleDragEnd}
          modifiers={[restrictToParentElement]}
        >
          <SortableContext items={filtered.map((a) => a.id)} strategy={rectSortingStrategy}>
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
              {filtered.map((area, i) => (
                <SortableAreaCard key={area.id} area={area} index={i} />
              ))}
            </div>
          </SortableContext>
        </DndContext>
      )}
    </div>
  );
}
