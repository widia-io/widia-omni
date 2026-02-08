import { useState } from "react";
import { format, addDays, subDays } from "date-fns";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useJournalEntry, useUpsertJournal } from "@/hooks/use-journal";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/cn";

const moods = [
  { emoji: "😫", value: 1 },
  { emoji: "😐", value: 2 },
  { emoji: "😊", value: 3 },
  { emoji: "🔥", value: 4 },
  { emoji: "🚀", value: 5 },
];

export function Component() {
  const [date, setDate] = useState(new Date());
  const dateStr = format(date, "yyyy-MM-dd");
  const { data: entry, isLoading } = useJournalEntry(dateStr);
  const upsert = useUpsertJournal();

  const [mood, setMood] = useState<number | null>(null);
  const [energy, setEnergy] = useState<number | null>(null);
  const [notes, setNotes] = useState("");
  const [wins, setWins] = useState("");
  const [challenges, setChallenges] = useState("");
  const [gratitude, setGratitude] = useState("");

  // Sync state when entry loads
  const currentMood = mood ?? entry?.mood ?? null;
  const currentEnergy = energy ?? entry?.energy ?? null;

  function handleSave() {
    upsert.mutate({
      date: dateStr,
      mood: currentMood,
      energy: currentEnergy,
      notes: notes || entry?.notes || null,
      wins: wins ? wins.split("\n").filter(Boolean) : entry?.wins ?? [],
      challenges: challenges ? challenges.split("\n").filter(Boolean) : entry?.challenges ?? [],
      gratitude: gratitude ? gratitude.split("\n").filter(Boolean) : entry?.gratitude ?? [],
    });
  }

  if (isLoading) return <Skeleton className="h-96 rounded-[14px]" />;

  return (
    <div>
      <div className="mb-6 flex items-center gap-4">
        <h1 className="text-2xl font-bold">Journal</h1>
        <div className="flex items-center gap-2 rounded-lg border border-border bg-bg-card px-3 py-1.5">
          <button onClick={() => { setDate(subDays(date, 1)); setMood(null); setEnergy(null); setNotes(""); setWins(""); setChallenges(""); setGratitude(""); }}>
            <ChevronLeft className="h-4 w-4 text-text-muted hover:text-text-primary" />
          </button>
          <span className="font-mono text-sm text-text-secondary min-w-[120px] text-center">
            {format(date, "dd/MM/yyyy")}
          </span>
          <button onClick={() => { setDate(addDays(date, 1)); setMood(null); setEnergy(null); setNotes(""); setWins(""); setChallenges(""); setGratitude(""); }}>
            <ChevronRight className="h-4 w-4 text-text-muted hover:text-text-primary" />
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Mood + Energy */}
        <div className="rounded-[14px] border border-border bg-bg-card p-6 space-y-4">
          <div>
            <Label className="mb-2 block">Humor</Label>
            <div className="flex gap-2">
              {moods.map(({ emoji, value }) => (
                <button key={value} onClick={() => setMood(value)} className={cn("flex h-10 w-10 items-center justify-center rounded-lg border text-xl transition-all hover:scale-110", currentMood === value ? "border-accent-orange bg-accent-orange-soft" : "border-border")}>
                  {emoji}
                </button>
              ))}
            </div>
          </div>
          <div>
            <Label className="mb-2 block">Energia (1-5)</Label>
            <div className="flex gap-2">
              {[1, 2, 3, 4, 5].map((v) => (
                <button key={v} onClick={() => setEnergy(v)} className={cn("flex h-10 w-10 items-center justify-center rounded-lg border font-mono text-sm transition-all", currentEnergy === v ? "border-accent-blue bg-accent-blue-soft text-accent-blue" : "border-border text-text-muted")}>
                  {v}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Notes */}
        <div className="rounded-[14px] border border-border bg-bg-card p-6">
          <Label className="mb-2 block">Notas</Label>
          <Textarea
            value={notes || entry?.notes || ""}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Como foi seu dia..."
            className="min-h-[120px]"
          />
        </div>

        {/* Wins */}
        <div className="rounded-[14px] border border-border bg-bg-card p-6">
          <Label className="mb-2 block">Vitorias</Label>
          <Textarea
            value={wins || (entry?.wins ?? []).join("\n")}
            onChange={(e) => setWins(e.target.value)}
            placeholder="Uma por linha..."
            className="min-h-[80px]"
          />
        </div>

        {/* Challenges + Gratitude */}
        <div className="space-y-4">
          <div className="rounded-[14px] border border-border bg-bg-card p-6">
            <Label className="mb-2 block">Desafios</Label>
            <Textarea
              value={challenges || (entry?.challenges ?? []).join("\n")}
              onChange={(e) => setChallenges(e.target.value)}
              placeholder="Uma por linha..."
              className="min-h-[60px]"
            />
          </div>
          <div className="rounded-[14px] border border-border bg-bg-card p-6">
            <Label className="mb-2 block">Gratidao</Label>
            <Textarea
              value={gratitude || (entry?.gratitude ?? []).join("\n")}
              onChange={(e) => setGratitude(e.target.value)}
              placeholder="Uma por linha..."
              className="min-h-[60px]"
            />
          </div>
        </div>
      </div>

      <div className="mt-6 flex justify-end">
        <Button onClick={handleSave} disabled={upsert.isPending}>
          {upsert.isPending ? "Salvando..." : "Salvar"}
        </Button>
      </div>
    </div>
  );
}
