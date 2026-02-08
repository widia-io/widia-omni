import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api-client";
import type { JournalEntry } from "@/types/api";

export function useJournalEntries(params?: Record<string, string>) {
  return useQuery({
    queryKey: ["journal", params],
    queryFn: () => api<JournalEntry[]>("/api/v1/journal", { params }),
  });
}

export function useJournalEntry(date: string) {
  return useQuery({
    queryKey: ["journal", date],
    queryFn: () => api<JournalEntry>(`/api/v1/journal/${date}`),
    enabled: !!date,
    retry: false,
  });
}

export function useUpsertJournal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ date, ...data }: Partial<JournalEntry> & { date: string }) =>
      api<JournalEntry>(`/api/v1/journal/${date}`, { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["journal"] }),
  });
}

export function useDeleteJournal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (date: string) => api(`/api/v1/journal/${date}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["journal"] }),
  });
}
