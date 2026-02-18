import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { AreaFormDialog } from "@/components/areas/area-form-dialog";

const mocked = vi.hoisted(() => ({
  createMutate: vi.fn(),
  updateMutate: vi.fn(),
  toastSuccess: vi.fn(),
  toastError: vi.fn(),
  state: {
    createPending: false,
    updatePending: false,
    usageData: {
      counters: { areas_count: 1 },
      limits: { max_areas: 4 },
    } as {
      counters: { areas_count: number };
      limits: { max_areas: number };
    } | undefined,
  },
}));

vi.mock("@/hooks/use-areas", () => ({
  useCreateArea: () => ({
    mutate: mocked.createMutate,
    isPending: mocked.state.createPending,
  }),
  useUpdateArea: () => ({
    mutate: mocked.updateMutate,
    isPending: mocked.state.updatePending,
  }),
}));

vi.mock("@/hooks/use-settings", () => ({
  useWorkspaceUsage: () => ({
    data: mocked.state.usageData,
  }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: mocked.toastSuccess,
    error: mocked.toastError,
  },
}));

describe("AreaFormDialog", () => {
  beforeEach(() => {
    mocked.createMutate.mockReset();
    mocked.updateMutate.mockReset();
    mocked.toastSuccess.mockReset();
    mocked.toastError.mockReset();
    mocked.state.createPending = false;
    mocked.state.updatePending = false;
    mocked.state.usageData = {
      counters: { areas_count: 1 },
      limits: { max_areas: 4 },
    };
  });

  it("submits create on Enter when form is valid", async () => {
    const user = userEvent.setup();

    render(<AreaFormDialog onClose={vi.fn()} />);

    const nameInput = screen.getByPlaceholderText("Ex.: Saúde e energia");
    await user.type(nameInput, "Saude");
    await user.keyboard("{Enter}");

    expect(mocked.createMutate).toHaveBeenCalledTimes(1);
    expect(mocked.createMutate.mock.calls[0]?.[0]).toMatchObject({
      name: "Saude",
      slug: "saude",
      icon: "heart",
      color: "orange",
      weight: 1,
      is_active: true,
    });
  });

  it("closes on Escape when mutation is not pending", () => {
    const onClose = vi.fn();

    render(<AreaFormDialog onClose={onClose} />);
    fireEvent.keyDown(document, { key: "Escape" });

    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it("blocks submit when area limit is reached", async () => {
    const user = userEvent.setup();
    mocked.state.usageData = {
      counters: { areas_count: 4 },
      limits: { max_areas: 4 },
    };

    render(<AreaFormDialog onClose={vi.fn()} />);

    const nameInput = screen.getByPlaceholderText("Ex.: Saúde e energia");
    await user.type(nameInput, "Nova Area");
    const submitButton = screen.getByRole("button", { name: "Criar área" });
    expect(submitButton).toBeDisabled();

    expect(mocked.createMutate).not.toHaveBeenCalled();
  });
});
