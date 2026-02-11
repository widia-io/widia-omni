import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { format } from "date-fns";
import { api } from "@/lib/api-client";
import { cn } from "@/lib/cn";
import type { JournalEntry } from "@/types/api";

const moods = [
  { emoji: "😫", value: 1 },
  { emoji: "😐", value: 2 },
  { emoji: "😊", value: 3 },
  { emoji: "🔥", value: 4 },
  { emoji: "🚀", value: 5 },
];

export function MoodSelector() {
  const [selected, setSelected] = useState<number | null>(null);
  const qc = useQueryClient();

  const mutation = useMutation({
    mutationFn: (mood: number) => {
      const today = format(new Date(), "yyyy-MM-dd");
      return api<JournalEntry>(`/api/v1/journal/${today}`, {
        method: "PUT",
        body: JSON.stringify({ mood }),
      });
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["journal"] }),
  });

  function handleSelect(value: number) {
    setSelected(value);
    mutation.mutate(value);
  }

  return (
    <div className="mt-4 rounded-[10px] border border-border bg-bg-secondary p-3.5">
      <div className="mb-2.5 text-[11px] font-semibold uppercase tracking-wider text-text-muted">
        Como você está hoje?
      </div>
      <div className="flex gap-2">
        {moods.map(({ emoji, value }) => (
          <button
            key={value}
            onClick={() => handleSelect(value)}
            className={cn(
              "flex h-9 w-9 items-center justify-center rounded-lg border text-lg transition-all hover:scale-110 hover:bg-bg-card",
              selected === value
                ? "border-accent-orange bg-accent-orange-soft"
                : "border-border bg-transparent",
            )}
          >
            {emoji}
          </button>
        ))}
      </div>
    </div>
  );
}
