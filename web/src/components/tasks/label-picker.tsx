import { Tag, Check, Settings2 } from "lucide-react";
import { useLabels } from "@/hooks/use-labels";
import { Popover, PopoverTrigger, PopoverContent } from "@/components/ui/popover";
import { cn } from "@/lib/cn";
import type { Label } from "@/types/api";

const MAX_DOTS = 3;

const colorDot: Record<string, string> = {
  orange: "bg-accent-orange",
  blue: "bg-accent-blue",
  green: "bg-accent-green",
  rose: "bg-accent-rose",
  sand: "bg-accent-sand",
  sage: "bg-accent-sage",
};

interface LabelPickerProps {
  selectedIds: string[];
  onChange: (ids: string[]) => void;
  onManageClick: () => void;
}

export function LabelPicker({ selectedIds, onChange, onManageClick }: LabelPickerProps) {
  const { data: labels } = useLabels();

  function toggle(id: string) {
    onChange(
      selectedIds.includes(id)
        ? selectedIds.filter((s) => s !== id)
        : [...selectedIds, id],
    );
  }

  const selected = (labels ?? []).filter((l) => selectedIds.includes(l.id));
  const overflow = selected.length - MAX_DOTS;

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          className={cn(
            "flex items-center gap-1.5 rounded-full border border-border px-2.5 py-1 text-xs transition-colors hover:border-text-muted hover:bg-bg-elevated focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent-orange/50",
            selected.length === 0 && "text-text-muted",
          )}
        >
          <Tag className="h-3 w-3 shrink-0 text-text-muted" />
          {selected.length > 0 ? (
            <span className="flex items-center gap-1">
              {selected.slice(0, MAX_DOTS).map((l) => (
                <span
                  key={l.id}
                  className={cn("h-2 w-2 rounded-full", colorDot[l.color] ?? "bg-text-muted")}
                />
              ))}
              {overflow > 0 && (
                <span className="ml-0.5 text-text-muted">+{overflow}</span>
              )}
            </span>
          ) : (
            <span>Etiquetas</span>
          )}
        </button>
      </PopoverTrigger>

      <PopoverContent align="start" className="w-56 p-0">
        <div className="max-h-56 overflow-y-auto py-1">
          {(labels ?? []).length === 0 ? (
            <p className="py-6 text-center text-xs text-text-muted">Nenhuma etiqueta</p>
          ) : (
            (labels ?? []).map((label) => (
              <LabelRow
                key={label.id}
                label={label}
                checked={selectedIds.includes(label.id)}
                onToggle={() => toggle(label.id)}
              />
            ))
          )}
        </div>

        <div className="border-t border-border px-1 py-1">
          <button
            type="button"
            onClick={onManageClick}
            className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-xs text-text-muted transition-colors hover:bg-bg-elevated hover:text-text-secondary"
          >
            <Settings2 className="h-3 w-3" />
            Gerenciar etiquetas
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
}

function LabelRow({ label, checked, onToggle }: { label: Label; checked: boolean; onToggle: () => void }) {
  return (
    <button
      type="button"
      onClick={onToggle}
      className="flex w-full items-center gap-2.5 px-2.5 py-1.5 text-left transition-colors hover:bg-bg-elevated"
    >
      <span className={cn("h-2 w-2 shrink-0 rounded-full", colorDot[label.color] ?? "bg-text-muted")} />
      <span className="flex-1 text-sm text-text-primary">{label.name}</span>
      <Check
        className={cn(
          "h-3.5 w-3.5 text-text-muted transition-opacity",
          checked ? "opacity-100" : "opacity-0",
        )}
      />
    </button>
  );
}
