import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Component as AreaDetailPage } from "@/pages/area-detail";

const mocked = vi.hoisted(() => ({
  navigate: vi.fn(),
  deleteMutate: vi.fn((_: string, options?: { onError?: () => void }) => {
    options?.onError?.();
  }),
  toastError: vi.fn(),
  summary: {
    area: {
      id: "area-1",
      name: "Saude",
      slug: "saude",
      icon: "heart",
      color: "orange",
      weight: 1,
      sort_order: 1,
      is_active: true,
      workspace_id: "ws-1",
      created_at: "2026-02-18T00:00:00.000Z",
      updated_at: "2026-02-18T00:00:00.000Z",
    },
    stats: {
      goals_active: 1,
      goals_completed: 0,
      projects_active: 0,
      projects_completed: 0,
      tasks_pending: 2,
      tasks_completed_this_week: 0,
      habits_active: 0,
      current_streak_avg: 0,
      area_score: 68,
    },
  },
}));

vi.mock("react-router", async () => {
  const actual = await vi.importActual<typeof import("react-router")>("react-router");
  return {
    ...actual,
    useParams: () => ({ id: "area-1" }),
    useNavigate: () => mocked.navigate,
    Link: ({ to, children, ...props }: { to: string; children: unknown }) => (
      <a href={to} {...props}>
        {children}
      </a>
    ),
  };
});

vi.mock("@/hooks/use-areas", () => ({
  useAreaSummary: () => ({
    data: mocked.summary,
    isLoading: false,
    isError: false,
  }),
  useDeleteArea: () => ({
    mutate: mocked.deleteMutate,
    isPending: false,
  }),
}));

vi.mock("@/hooks/use-settings", () => ({
  useWorkspaceUsage: () => ({
    data: {
      counters: { areas_count: 1 },
      limits: { max_areas: 4 },
    },
  }),
}));

vi.mock("@/hooks/use-goals", () => ({
  useGoals: () => ({ data: [] }),
  useCreateGoal: () => ({
    mutate: vi.fn(),
    isPending: false,
  }),
}));

vi.mock("@/hooks/use-tasks", () => ({
  useTasks: () => ({ data: [] }),
  useCreateTask: () => ({
    mutate: vi.fn(),
    isPending: false,
  }),
}));

vi.mock("@/hooks/use-projects", () => ({
  useProjects: () => ({ data: [] }),
  useCreateProject: () => ({
    mutate: vi.fn(),
    isPending: false,
  }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: mocked.toastError,
  },
}));

describe("AreaDetail delete flow", () => {
  beforeEach(() => {
    mocked.navigate.mockReset();
    mocked.deleteMutate.mockClear();
    mocked.toastError.mockReset();
  });

  it("keeps confirm dialog open when delete mutation fails", async () => {
    const user = userEvent.setup();

    render(<AreaDetailPage />);

    await user.click(screen.getByRole("button", { name: "Excluir área" }));
    await user.click(screen.getByRole("button", { name: "Excluir" }));

    expect(mocked.deleteMutate).toHaveBeenCalledTimes(1);
    expect(mocked.toastError).toHaveBeenCalledWith("Erro ao excluir área");
    expect(screen.getByText(/deseja excluir\\?/i)).toBeInTheDocument();
  });
});
