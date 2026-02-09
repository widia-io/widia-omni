import { useState } from "react";
import { Plus } from "lucide-react";
import { areaIconMap } from "@/lib/icons";
import { useAreas, useCreateArea, useUpdateArea, useDeleteArea } from "@/hooks/use-areas";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { cn } from "@/lib/cn";
import type { LifeArea } from "@/types/api";

function AreaFormDialog({ area, onClose }: { area?: LifeArea; onClose: () => void }) {
  const create = useCreateArea();
  const update = useUpdateArea();
  const [name, setName] = useState(area?.name ?? "");
  const [icon, setIcon] = useState(area?.icon ?? "🎯");
  const [color, setColor] = useState(area?.color ?? "orange");
  const [weight, setWeight] = useState(String(area?.weight ?? 1));

  const colors = ["green", "orange", "blue", "rose", "sand", "sage"];

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const slug = name.toLowerCase().replace(/\s+/g, "-").replace(/[^a-z0-9-]/g, "");
    const data = { name, slug, icon, color, weight: Number(weight) };
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
      <Button type="submit" className="w-full" disabled={create.isPending || update.isPending}>
        {area ? "Salvar" : "Criar area"}
      </Button>
    </form>
  );
}

export function Component() {
  const { data: areas, isLoading } = useAreas();
  const deleteArea = useDeleteArea();
  const [editArea, setEditArea] = useState<LifeArea | undefined>();
  const [open, setOpen] = useState(false);

  if (isLoading) {
    return <div className="grid grid-cols-3 gap-4">{[1,2,3,4,5,6].map(i => <Skeleton key={i} className="h-32 rounded-[14px]" />)}</div>;
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

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {(areas ?? []).map((area) => (
          <Card key={area.id} className="cursor-pointer transition-all hover:-translate-y-0.5" onClick={() => { setEditArea(area); setOpen(true); }}>
            <div className="flex items-center justify-between mb-2">
              <span className="text-2xl">{(() => { const Icon = areaIconMap[area.icon]; return Icon ? <Icon size={24} /> : area.icon; })()}</span>
              <button
                onClick={(e) => { e.stopPropagation(); deleteArea.mutate(area.id); }}
                className="text-xs text-text-muted hover:text-accent-rose transition-colors"
              >
                Excluir
              </button>
            </div>
            <div className="font-semibold">{area.name}</div>
            <div className="mt-1 text-xs text-text-muted">Peso: {area.weight}</div>
          </Card>
        ))}
      </div>
    </div>
  );
}
