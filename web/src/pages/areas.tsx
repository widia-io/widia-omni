import { useState } from "react";
import { useNavigate } from "react-router";
import { Plus } from "lucide-react";
import { areaIconMap } from "@/lib/icons";
import { useAreas, useCreateArea, useUpdateArea } from "@/hooks/use-areas";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { cn } from "@/lib/cn";
import type { LifeArea, AreaWithStats } from "@/types/api";

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

export function AreaFormDialog({ area, onClose }: { area?: LifeArea; onClose: () => void }) {
  const create = useCreateArea();
  const update = useUpdateArea();
  const [name, setName] = useState(area?.name ?? "");
  const [icon, setIcon] = useState(area?.icon ?? "🎯");
  const [color, setColor] = useState(area?.color ?? "orange");
  const [weight, setWeight] = useState(String(area?.weight ?? 1));
  const [isActive, setIsActive] = useState(area?.is_active ?? true);

  const colors = ["green", "orange", "blue", "rose", "sand", "sage"];

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const slug = name.toLowerCase().replace(/\s+/g, "-").replace(/[^a-z0-9-]/g, "");
    const data = { name, slug, icon, color, weight: Number(weight), is_active: isActive };
    if (area) {
      update.mutate({ id: area.id, ...data }, { onSuccess: onClose });
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
        <Label>Icone</Label>
        <Input value={icon} onChange={(e) => setIcon(e.target.value)} />
      </div>
      <div className="space-y-2">
        <Label>Cor</Label>
        <div className="flex gap-2">
          {colors.map((c) => (
            <button key={c} type="button" onClick={() => setColor(c)} className={cn("h-8 w-8 rounded-lg border-2 transition-colors", color === c ? "border-text-primary" : "border-transparent", `bg-accent-${c}`)} />
          ))}
        </div>
      </div>
      <div className="space-y-2">
        <Label>Peso</Label>
        <Input type="number" min="0" max="10" step="0.1" value={weight} onChange={(e) => setWeight(e.target.value)} />
      </div>
      <div className="flex items-center gap-2">
        <Checkbox checked={isActive} onCheckedChange={(v) => setIsActive(Boolean(v))} />
        <Label>Área ativa</Label>
      </div>
      <Button type="submit" className="w-full" disabled={create.isPending || update.isPending}>
        {area ? "Salvar" : "Criar area"}
      </Button>
    </form>
  );
}

function AreaCard({ area, index }: { area: AreaWithStats; index: number }) {
  const navigate = useNavigate();
  const c = getColorClasses(area.color);
  const score = area.area_score ?? 0;

  return (
    <div
      onClick={() => navigate(`/areas/${area.id}`)}
      className={cn(
        "group relative cursor-pointer overflow-hidden rounded-[14px] border border-border bg-bg-card p-[18px_20px] transition-all duration-[250ms] hover:-translate-y-0.5 hover:border-transparent animate-in",
      )}
      style={{ animationDelay: `${0.15 + index * 0.05}s` }}
    >
      <div className={cn("absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r opacity-0 transition-opacity group-hover:opacity-100", c.bar)} />

      <div className="flex items-center justify-between mb-3">
        <div className={cn("flex h-[34px] w-[34px] items-center justify-center rounded-[9px]", c.bg)}>
          {(() => { const Icon = areaIconMap[area.icon]; return Icon ? <Icon size={18} className={c.text} /> : <span className="text-lg">{area.icon}</span>; })()}
        </div>
        <span className={cn("font-mono text-xl font-bold", c.text)}>{score}</span>
      </div>

      <div className="text-sm font-semibold">{area.name}</div>

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
  const { data: areas, isLoading } = useAreas();
  const [editArea, setEditArea] = useState<LifeArea | undefined>();
  const [open, setOpen] = useState(false);

  if (isLoading) {
    return (
      <div>
        <div className="mb-6 flex items-center justify-between">
          <Skeleton className="h-8 w-40" />
          <Skeleton className="h-8 w-28" />
        </div>
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3, 4, 5, 6].map((i) => <Skeleton key={i} className="h-32 rounded-[14px]" />)}
        </div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold">Areas de Vida</h1>
        <Dialog open={open} onOpenChange={(v) => { setOpen(v); if (!v) setEditArea(undefined); }}>
          <DialogTrigger asChild>
            <Button size="sm"><Plus className="h-4 w-4" /> Nova area</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader><DialogTitle>{editArea ? "Editar area" : "Nova area"}</DialogTitle></DialogHeader>
            <AreaFormDialog area={editArea} onClose={() => setOpen(false)} />
          </DialogContent>
        </Dialog>
      </div>

      <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
        {(areas ?? []).map((area, i) => (
          <AreaCard key={area.id} area={area} index={i} />
        ))}
      </div>
    </div>
  );
}
