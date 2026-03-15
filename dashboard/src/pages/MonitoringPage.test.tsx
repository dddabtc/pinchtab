import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MonitoringPage from "./MonitoringPage";
import { useAppStore } from "../stores/useAppStore";
import type { Instance, InstanceTab, Settings } from "../generated/types";

const apiMock = vi.hoisted(() => ({
  fetchTabScreenshot: vi.fn(),
  stopInstance: vi.fn(),
}));

vi.mock("../services/api", () => apiMock);

vi.mock("../components/molecules", () => ({
  TabsChart: () => <div data-testid="tabs-chart" />,
  InstanceListItem: ({
    instance,
    onClick,
    selected,
  }: {
    instance: Instance;
    onClick: () => void;
    selected: boolean;
  }) => (
    <button onClick={onClick} aria-pressed={selected}>
      {instance.profileName}
    </button>
  ),
  TabItem: ({ tab }: { tab: InstanceTab }) => <div>{tab.title}</div>,
}));

const baseSettings: Settings = {
  screencast: { fps: 1, quality: 30, maxWidth: 800 },
  stealth: "light",
  browser: { blockImages: false, blockMedia: false, noAnimations: false },
  monitoring: { memoryMetrics: false, pollInterval: 30 },
};

const instances: Instance[] = [
  {
    id: "inst_headless",
    profileId: "prof_headless",
    profileName: "Headless worker",
    port: "9988",
    headless: true,
    status: "running",
    startTime: "2026-03-06T10:00:00Z",
    attached: false,
  },
  {
    id: "inst_empty",
    profileId: "prof_empty",
    profileName: "No tabs",
    port: "9989",
    headless: false,
    status: "running",
    startTime: "2026-03-06T11:00:00Z",
    attached: false,
  },
];

const tabsByInstance: Record<string, InstanceTab[]> = {
  inst_headless: [
    {
      id: "tab_primary",
      instanceId: "inst_headless",
      title: "PinchTab Dashboard",
      url: "https://pinchtab.dev/dashboard",
    },
  ],
  inst_empty: [],
};

describe("MonitoringPage", () => {
  const createObjectURL = vi.fn();
  const revokeObjectURL = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    createObjectURL.mockReturnValue("blob:preview-1");
    useAppStore.setState({
      instances,
      tabsChartData: [],
      memoryChartData: [],
      serverChartData: [],
      currentTabs: tabsByInstance,
      currentMemory: {},
      settings: baseSettings,
    });
    vi.stubGlobal("URL", {
      ...URL,
      createObjectURL,
      revokeObjectURL,
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("captures and shows a manual screenshot preview for the selected instance", async () => {
    apiMock.fetchTabScreenshot.mockResolvedValue(
      new Blob(["png"], { type: "image/png" }),
    );

    render(<MonitoringPage />);

    const captureButton = screen.getByRole("button", {
      name: "Capture Screenshot",
    });
    expect(captureButton).toBeEnabled();
    expect(
      screen.getByText("Capture a manual screenshot of the first open tab."),
    ).toBeInTheDocument();

    await userEvent.click(captureButton);

    await waitFor(() =>
      expect(apiMock.fetchTabScreenshot).toHaveBeenCalledWith("tab_primary"),
    );

    const preview = await screen.findByRole("img", {
      name: "Screenshot preview for Headless worker",
    });
    expect(preview).toHaveAttribute("src", "blob:preview-1");
    expect(
      screen.getByText("Last captured from tab tab_primary"),
    ).toBeVisible();
    expect(
      screen.getByRole("button", { name: "Refresh Screenshot" }),
    ).toBeVisible();
  });

  it("shows an error message when screenshot capture fails", async () => {
    apiMock.fetchTabScreenshot.mockRejectedValue(
      new Error("backend unavailable"),
    );

    render(<MonitoringPage />);

    await userEvent.click(
      screen.getByRole("button", { name: "Capture Screenshot" }),
    );

    expect(
      await screen.findByText("Screenshot failed: backend unavailable"),
    ).toBeVisible();
  });

  it("disables screenshot capture when the selected instance has no open tabs", async () => {
    render(<MonitoringPage />);

    await userEvent.click(screen.getByRole("button", { name: "No tabs" }));

    expect(
      screen.getByRole("button", { name: "Capture Screenshot" }),
    ).toBeDisabled();
    expect(
      screen.getByText("Open a tab to capture a screenshot."),
    ).toBeInTheDocument();
  });
});
