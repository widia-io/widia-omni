import { useState } from "react";
import { Trash2 } from "lucide-react";
import { useLabels, useCreateLabel, useUpdateLabel, useDeleteLabel } from "@/hooks/use-labels";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { cn } from "@/lib/cn";
import type { Label } from "@/types/api";

const COLORS = ["green", "orange", "blue", "rose", "sand", "sage"] as const;

const colorDot: Record<string, string> = {
  orange: "bg-accent-orange",
  blue: "bg-accent-blue",
  green: "bg-accent-green",
  rose: "bg-accent-rose",
  sand: "bg-accent-sand",
  sage: "bg-accent-sage",
};

function ColorDots({ value, onChange }: { value: string; onChange: (c: string) => void }) {
  return (
    <div className="flex items-center gap-1">
      {COLORS.map((c) => (
        <button
          key={c}
          type="button"
          onClick={() => onChange(c)}
          className={cn(
            "h-5 w-5 rounded-full transition-all",
            value === c
              ? "ring-2 ring-text-primary ring-offset-1 ring-offset-bg-card scale-110"
              : "hover:scale-105",
            colorDot[c],
          )}
        />
      ))}
    </div>
  );
}

interface LabelManagerDialogProps {
  open: boolean;
  onOpenChange: (v: boolean) => void;
}

export function LabelManagerDialog({ open, onOpenChange }: LabelManagerDialogProps) {
  const { data: labels } = useLabels();
  const createLabel = useCreateLabel();
  const [name, setName] = useState("");
  const [color, setColor] = useState("orange");

  function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    createLabel.mutate({ name: name.trim(), color }, {
      onSuccess: () => { setName(""); setColor("orange"); },
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Gerenciar etiquetas</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleCreate} className="flex items-center gap-2">
          <input
            placeholder="Nova etiqueta..."
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="flex-1 bg-transparent text-sm text-text-primary placeholder:text-text-muted outline-none border-b border-border py-1.5 focus:border-text-muted transition-colors"
          />
          <ColorDots value={color} onChange={setColor} />
          <button
            type="submit"
            disabled={!name.trim() || createLabel.isPending}
            className="shrink-0 text-xs font-medium text-accent-orange transition-opacity hover:opacity-80 disabled:opacity-30"
          >
            Criar
          </button>
        </form>

        <div className="mt-1 max-h-64 overflow-y-auto -mx-1">
          {(labels ?? []).length === 0 ? (
            <p className="py-8 text-center text-xs text-text-muted">Nenhuma etiqueta</p>
          ) : (
            (labels ?? []).map((label) => (
              <LabelRow key={label.id} label={label} />
            ))
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

function LabelRow({ label }: { label: Label }) {
  const updateLabel = useUpdateLabel();
  const deleteLabel = useDeleteLabel();
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(label.name);
  const [color, setColor] = useState(label.color);

  function handleSave() {
    if (!name.trim()) return;
    updateLabel.mutate({ id: label.id, name: name.trim(), color }, {
      onSuccess: () => setEditing(false),
    });
  }

  function handleCancel() {
    setName(label.name);
    setColor(label.color);
    setEditing(false);
  }

  if (editing) {
    return (
      <div className="space-y-2 rounded-lg bg-bg-elevated px-3 py-2.5 mx-1">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          autoFocus
          onKeyDown={(e) => {
            if (e.key === "Enter") handleSave();
            if (e.key === "Escape") handleCancel();
          }}
          className="w-full bg-transparent text-sm text-text-primary outline-none border-b border-border pb-1 focus:border-text-muted transition-colors"
        />
        <div className="flex items-center justify-between">
          <ColorDots value={color} onChange={setColor} />
          <div className="flex gap-2">
            <button
              type="button"
              onClick={handleCancel}
              className="text-xs text-text-muted hover:text-text-secondary transition-colors"
            >
              Cancelar
            </button>
            <button
              type="button"
              onClick={handleSave}
              disabled={!name.trim() || updateLabel.isPending}
              className="text-xs font-medium text-accent-orange hover:opacity-80 disabled:opacity-30 transition-opacity"
            >
              Salvar
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="group flex items-center gap-2.5 rounded-lg px-3 py-2 mx-1 transition-colors hover:bg-bg-elevated">
      <span className={cn("h-2.5 w-2.5 shrink-0 rounded-full", colorDot[label.color] ?? "bg-text-muted")} />
      <button
        type="button"
        onClick={() => setEditing(true)}
        className="flex-1 text-left text-sm text-text-primary"
      >
        {label.name}
      </button>
      <button
        type="button"
        onClick={() => deleteLabel.mutate(label.id)}
        className="text-text-muted opacity-0 transition-all hover:text-accent-rose group-hover:opacity-100"
      >
        <Trash2 className="h-3.5 w-3.5" />
      </button>
    </div>
  );
}
