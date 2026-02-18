import {
  areTaskFiltersEqual,
  buildSearchParamsWithTaskFilter,
  parseTaskFilterFromSearchParams,
} from "@/lib/task-filters-url";

describe("task-filters-url", () => {
  it("parses known task filters from URL", () => {
    const searchParams = new URLSearchParams(
      "area_id=area-1&goal_id=goal-1&is_completed=false&due_from=2026-02-01&due_to=2026-02-28&foo=bar",
    );

    const filter = parseTaskFilterFromSearchParams(searchParams);

    expect(filter).toEqual({
      area_id: "area-1",
      goal_id: "goal-1",
      is_completed: "false",
      due_from: "2026-02-01",
      due_to: "2026-02-28",
    });
  });

  it("drops section filter when area is missing", () => {
    const searchParams = new URLSearchParams("section_id=section-1");

    const filter = parseTaskFilterFromSearchParams(searchParams);

    expect(filter).toEqual({});
  });

  it("merges task filters into URL and preserves unrelated params", () => {
    const current = new URLSearchParams("tab=focus&foo=bar&area_id=old");
    const next = buildSearchParamsWithTaskFilter(current, {
      area_id: "area-1",
      section_id: "section-1",
      is_completed: "false",
      goal_id: "goal-1",
      due_from: "2026-02-01",
      due_to: "2026-02-28",
    });

    expect(next.toString()).toBe(
      "tab=focus&foo=bar&is_completed=false&area_id=area-1&goal_id=goal-1&section_id=section-1&due_from=2026-02-01&due_to=2026-02-28",
    );
  });

  it("treats filters as equal regardless of section when area is not set", () => {
    const left = { is_completed: "false", section_id: "section-1" };
    const right = { is_completed: "false" };

    expect(areTaskFiltersEqual(left, right)).toBe(true);
  });

  it("keeps advanced filters independent from area", () => {
    const searchParams = new URLSearchParams("goal_id=goal-1&due_from=2026-02-01&due_to=2026-02-28");

    const filter = parseTaskFilterFromSearchParams(searchParams);

    expect(filter).toEqual({
      goal_id: "goal-1",
      due_from: "2026-02-01",
      due_to: "2026-02-28",
    });
  });
});
