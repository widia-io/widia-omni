import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { LifeScore, ScoreHistory } from "@/types/api";

export function useCurrentScore() {
  return useQuery({
    queryKey: ["scores", "current"],
    queryFn: () => api<LifeScore>("/api/v1/scores/current"),
  });
}

export function useScoreHistory(weeks?: number) {
  return useQuery({
    queryKey: ["scores", "history", weeks],
    queryFn: () => api<ScoreHistory>("/api/v1/scores/history", { params: weeks ? { weeks: String(weeks) } : {} }),
  });
}
