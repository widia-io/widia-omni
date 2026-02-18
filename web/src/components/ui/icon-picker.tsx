import { useState } from "react";
import { ChevronDown, Search } from "lucide-react";
import { areaIconMap } from "@/lib/icons";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { cn } from "@/lib/cn";

const entries = Object.entries(areaIconMap);

export function IconPicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (key: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");

  const filtered = search
    ? entries.filter(([key]) => key.includes(search.toLowerCase()))
    : entries;

  const CurrentIcon = areaIconMap[value];

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="inline-flex items-center gap-2 rounded-lg border border-border bg-bg-card px-3 py-2 text-sm transition-colors hover:border-text-muted"
        >
          {CurrentIcon ? (
            <CurrentIcon size={18} />
          ) : (
            <span className="text-lg leading-none">{value}</span>
          )}
          <ChevronDown size={14} className="text-text-muted" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-60 p-3">
        <div className="relative mb-2">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-text-muted" />
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar..."
            className="w-full rounded-md border border-border bg-bg-base py-1.5 pl-8 pr-2 text-xs outline-none placeholder:text-text-muted focus:border-text-muted"
          />
        </div>
        <div className="grid grid-cols-4 gap-1">
          {filtered.map(([key, Icon]) => (
            <button
              key={key}
              type="button"
              onClick={() => { onChange(key); setOpen(false); setSearch(""); }}
              className={cn(
                "flex h-9 w-full items-center justify-center rounded-md transition-colors hover:bg-bg-hover",
                value === key && "bg-accent-orange/10 text-accent-orange",
              )}
              title={key}
            >
              <Icon size={18} />
            </button>
          ))}
        </div>
        {filtered.length === 0 && (
          <p className="py-3 text-center text-xs text-text-muted">Nenhum ícone encontrado</p>
        )}
      </PopoverContent>
    </Popover>
  );
}
