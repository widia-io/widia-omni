import { useAreas } from "@/hooks/use-areas";
import { useScoreHistory } from "@/hooks/use-scores";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/cn";
import { getAreaIcon, getAreaIconWithFallback, isRawAreaIcon } from "@/lib/icons";

const colorMap: Record<string, { bg: string; text: string; bar: string; border: string; shadow: string }> = {
  green: {
    bg: "bg-accent-green-soft", text: "text-accent-green",
    bar: "from-accent-green to-accent-sage", border: "border-accent-green-mid", shadow: "shadow-accent-green-soft",
  },
  orange: {
    bg: "bg-accent-orange-soft", text: "text-accent-orange",
    bar: "from-accent-orange to-accent-sand", border: "border-accent-orange-mid", shadow: "shadow-accent-orange-soft",
  },
  blue: {
    bg: "bg-accent-blue-soft", text: "text-accent-blue",
    bar: "from-accent-blue to-accent-sage", border: "border-accent-blue-mid", shadow: "shadow-accent-blue-soft",
  },
  rose: {
    bg: "bg-accent-rose-soft", text: "text-accent-rose",
    bar: "from-accent-rose to-accent-orange", border: "border-accent-rose-soft", shadow: "shadow-accent-rose-soft",
  },
  sand: {
    bg: "bg-accent-sand-soft", text: "text-accent-sand",
    bar: "from-accent-sand to-accent-blue", border: "border-accent-sand-soft", shadow: "shadow-accent-sand-soft",
  },
  sage: {
    bg: "bg-accent-sage-soft", text: "text-accent-sage",
    bar: "from-accent-sage to-accent-green", border: "border-accent-sage-soft", shadow: "shadow-accent-sage-soft",
  },
};

function getColorClasses(color: string) {
  return colorMap[color] ?? colorMap["orange"]!;
}

function getAreaDisplayName(name: string, slug: string) {
  const trimmed = name?.trim();
  if (trimmed) return trimmed;
  const normalizedSlug = slug?.trim();
  if (!normalizedSlug) return "Área sem nome";
  return normalizedSlug
    .split("-")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

export function AreasGrid() {
  const { data: areas, isLoading } = useAreas();
  const { data: history } = useScoreHistory(1);

  if (isLoading) {
    return (
      <div className="grid grid-cols-3 grid-rows-2 gap-3">
        {[1, 2, 3, 4, 5, 6].map((i) => <Skeleton key={i} className="h-28 rounded-[14px]" />)}
      </div>
    );
  }

  const areaScores = new Map(
    (history?.area_scores ?? []).map((s) => [s.area_id, s.score]),
  );

  return (
    <div className="grid grid-cols-2 gap-3 lg:grid-cols-3 lg:grid-rows-2">
      {(areas ?? []).slice(0, 6).map((area, i) => {
        const c = getColorClasses(area.color);
        const score = areaScores.get(area.id) ?? 0;
        const Icon = getAreaIcon(area.icon);
        const FallbackIcon = getAreaIconWithFallback(area);
        const rawIcon = area.icon?.trim();
        const showRawIcon = isRawAreaIcon(rawIcon);
        const displayName = getAreaDisplayName(area.name, area.slug);

        return (
          <div
            key={area.id}
            className={cn(
              "group relative cursor-pointer overflow-hidden rounded-[14px] border border-border bg-bg-card p-[18px_20px] transition-all duration-[250ms] hover:-translate-y-0.5 hover:border-transparent animate-in",
            )}
            style={{ animationDelay: `${0.15 + i * 0.05}s` }}
          >
            {/* Top accent line */}
            <div className={cn("absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r opacity-0 transition-opacity group-hover:opacity-100", c.bar)} />

            <div className="flex items-center justify-between mb-3">
              <div className={cn("flex h-[34px] w-[34px] items-center justify-center rounded-[9px]", c.bg)}>
                {Icon ? <Icon size={18} className={c.text} /> : showRawIcon ? <span className="text-lg">{rawIcon}</span> : <FallbackIcon size={18} className={c.text} />}
              </div>
              <span className={cn("font-mono text-xl font-bold", c.text)}>{score}</span>
            </div>
            <div className="text-sm font-semibold">{displayName}</div>
            <div className="h-[3px] mt-3 overflow-hidden rounded-[3px] bg-border">
              <div
                className={cn("h-full rounded-[3px] bg-gradient-to-r transition-all duration-[1.5s]", c.bar)}
                style={{ width: `${score}%` }}
              />
            </div>
          </div>
        );
      })}
    </div>
  );
}
