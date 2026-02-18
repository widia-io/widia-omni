export const TASK_FILTER_QUERY_KEYS = [
  "is_completed",
  "area_id",
  "goal_id",
  "project_id",
  "section_id",
  "label_id",
  "due_from",
  "due_to",
] as const;

export type TaskFilterQueryKey = (typeof TASK_FILTER_QUERY_KEYS)[number];

function normalizeTaskFilter(filter: Record<string, string>) {
  const normalized: Record<string, string> = {};

  for (const key of TASK_FILTER_QUERY_KEYS) {
    const value = filter[key];
    if (value) normalized[key] = value;
  }

  // Section only makes sense when an area is selected in the current UI model.
  if (!normalized.area_id) {
    delete normalized.section_id;
  }

  return normalized;
}

export function parseTaskFilterFromSearchParams(searchParams: URLSearchParams): Record<string, string> {
  const raw: Record<string, string> = {};

  for (const key of TASK_FILTER_QUERY_KEYS) {
    const value = searchParams.get(key);
    if (value) raw[key] = value;
  }

  return normalizeTaskFilter(raw);
}

export function buildSearchParamsWithTaskFilter(
  searchParams: URLSearchParams,
  filter: Record<string, string>,
): URLSearchParams {
  const normalized = normalizeTaskFilter(filter);
  const next = new URLSearchParams(searchParams);

  for (const key of TASK_FILTER_QUERY_KEYS) {
    next.delete(key);
  }

  for (const key of TASK_FILTER_QUERY_KEYS) {
    const value = normalized[key];
    if (value) next.set(key, value);
  }

  return next;
}

export function areTaskFiltersEqual(a: Record<string, string>, b: Record<string, string>) {
  const left = normalizeTaskFilter(a);
  const right = normalizeTaskFilter(b);

  for (const key of TASK_FILTER_QUERY_KEYS) {
    if ((left[key] ?? "") !== (right[key] ?? "")) return false;
  }

  return true;
}
