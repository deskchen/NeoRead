// @vitest-environment jsdom
import { act } from "react";
import { createRoot } from "react-dom/client";
import { expect, it, vi } from "vitest";
import { api } from "@neoread/api";
import type { JournalRankResult } from "@neoread/core";
import { JournalRankBadges } from "./JournalRankBadges";

vi.mock("@neoread/api", () => ({ api: { documentJournalRank: vi.fn() } }));

it("shows every selected subject and omits unselected indicators", async () => {
  vi.stubGlobal("IS_REACT_ACT_ENVIRONMENT", true);
  vi.mocked(api.documentJournalRank).mockResolvedValue({
    status: "found",
    badges: [
      { key: "sci", name: "SCI/JCR", value: "Q1" },
      { key: "sciUpSmall", name: "中科院小类", value: "自动化与控制系统4区" },
      { key: "sciUpSmall", name: "中科院小类", value: "计算机：人工智能4区" },
      { key: "jci", name: "JCI", value: "0.94" },
    ],
  });
  const host = document.createElement("div");
  const root = createRoot(host);
  await act(async () =>
    root.render(
      <JournalRankBadges
        id="paper"
        venue="Journal of Machine Learning Research"
        config={{
          enabled: true,
          hasKey: true,
          filterMode: "include",
          selected: ["sci", "sciUpSmall"],
          available: [],
        }}
        refreshRevision={0}
      />,
    ),
  );
  expect(host.textContent).toContain("SCI/JCR Q1");
  expect(host.textContent).toContain("自动化与控制系统4区");
  expect(host.textContent).toContain("计算机：人工智能4区");
  expect(host.textContent).not.toContain("JCI");
  await act(async () => root.unmount());
});

it("hides selected indicators in exclusion mode", async () => {
  vi.mocked(api.documentJournalRank).mockResolvedValue({
    status: "found",
    badges: [
      { key: "sci", name: "SCI/JCR", value: "Q1" },
      { key: "jci", name: "JCI", value: "0.94" },
    ],
  });
  const host = document.createElement("div");
  const root = createRoot(host);
  await act(async () =>
    root.render(
      <JournalRankBadges
        id="paper"
        venue="Nature"
        config={{
          enabled: true,
          hasKey: true,
          filterMode: "exclude",
          selected: ["sci"],
          available: [],
        }}
        refreshRevision={0}
      />,
    ),
  );
  expect(host.textContent).not.toContain("SCI/JCR");
  expect(host.textContent).toContain("JCI 0.94");
  await act(async () => root.unmount());
});

it("omits whole badges when the title column is too narrow", async () => {
  let badgeWidth = 40;
  const oldWidth = Object.getOwnPropertyDescriptor(
    HTMLElement.prototype,
    "offsetWidth",
  );
  const oldClientWidth = Object.getOwnPropertyDescriptor(
    HTMLElement.prototype,
    "clientWidth",
  );
  Object.defineProperty(HTMLElement.prototype, "offsetWidth", {
    configurable: true,
    get: () => badgeWidth,
  });
  Object.defineProperty(HTMLElement.prototype, "clientWidth", {
    configurable: true,
    get: () => 60,
  });
  vi.mocked(api.documentJournalRank).mockResolvedValue({
    status: "found",
    badges: [
      { key: "sci", name: "SCI/JCR", value: "Q1" },
      { key: "jci", name: "JCI", value: "0.94" },
    ],
  });
  const host = document.createElement("div");
  const root = createRoot(host);
  try {
    await act(async () =>
      root.render(
        <JournalRankBadges
          id="paper"
          venue="Nature"
          config={{
            enabled: true,
            hasKey: true,
            filterMode: "include",
            selected: ["sci", "jci"],
            available: [],
          }}
          refreshRevision={0}
        />,
      ),
    );
    expect(
      host.querySelectorAll('.journal-rank-badge[data-hidden="true"]'),
    ).toHaveLength(1);
    expect(
      host.querySelectorAll(".journal-rank-badge:not([data-hidden])"),
    ).toHaveLength(1);
    badgeWidth = 80;
    await act(async () => window.dispatchEvent(new Event("resize")));
    expect(
      host.querySelectorAll('.journal-rank-badge[data-hidden="true"]'),
    ).toHaveLength(2);
  } finally {
    await act(async () => root.unmount());
    if (oldWidth)
      Object.defineProperty(HTMLElement.prototype, "offsetWidth", oldWidth);
    else delete (HTMLElement.prototype as { offsetWidth?: number }).offsetWidth;
    if (oldClientWidth)
      Object.defineProperty(
        HTMLElement.prototype,
        "clientWidth",
        oldClientWidth,
      );
    else delete (HTMLElement.prototype as { clientWidth?: number }).clientWidth;
  }
});

it("does not query while the feature is off", async () => {
  vi.mocked(api.documentJournalRank).mockClear();
  const host = document.createElement("div");
  const root = createRoot(host);
  await act(async () =>
    root.render(
      <JournalRankBadges
        id="paper"
        venue="Nature"
        config={{
          enabled: false,
          hasKey: true,
          filterMode: "include",
          selected: ["sci"],
          available: [],
        }}
        refreshRevision={0}
      />,
    ),
  );
  expect(api.documentJournalRank).not.toHaveBeenCalled();
  expect(host.textContent).toBe("");
  await act(async () => root.unmount());
});

it("forces only the requested refresh, then reuses cache for another venue", async () => {
  vi.mocked(api.documentJournalRank).mockClear();
  vi.mocked(api.documentJournalRank).mockResolvedValue({
    status: "notFound",
    badges: [],
  });
  const host = document.createElement("div");
  const root = createRoot(host);
  const config = {
    enabled: true,
    hasKey: true,
    filterMode: "include" as const,
    selected: ["sci"],
    available: [],
  };
  await act(async () =>
    root.render(
      <JournalRankBadges
        id="paper"
        venue="Nature"
        config={config}
        refreshRevision={0}
      />,
    ),
  );
  await act(async () =>
    root.render(
      <JournalRankBadges
        id="paper"
        venue="Nature"
        config={config}
        refreshRevision={1}
      />,
    ),
  );
  await act(async () =>
    root.render(
      <JournalRankBadges
        id="paper"
        venue="Science"
        config={config}
        refreshRevision={1}
      />,
    ),
  );
  expect(
    vi.mocked(api.documentJournalRank).mock.calls.map((call) => call[1]),
  ).toEqual([false, true, false]);
  await act(async () => root.unmount());
});

it("reloads cached ratings when only the venue abbreviation changes and ignores an older response", async () => {
  vi.stubGlobal("IS_REACT_ACT_ENVIRONMENT", true);
  vi.mocked(api.documentJournalRank).mockReset();
  let resolveOldRequest!: (value: JournalRankResult) => void;
  const oldRequest = new Promise<JournalRankResult>((resolve) => {
    resolveOldRequest = resolve;
  });
  vi.mocked(api.documentJournalRank)
    .mockReturnValueOnce(oldRequest)
    .mockResolvedValueOnce({
      status: "found",
      badges: [{ key: "ccf", name: "CCF", value: "A" }],
    });
  const host = document.createElement("div");
  const root = createRoot(host);
  const config = {
    enabled: true,
    hasKey: true,
    filterMode: "include" as const,
    selected: ["ccf"],
    available: [],
  };
  await act(async () =>
    root.render(
      <JournalRankBadges
        id="paper"
        venue="International Conference on Long-Tail Engineering"
        venueShort="OLD"
        config={config}
        refreshRevision={0}
      />,
    ),
  );
  await act(async () =>
    root.render(
      <JournalRankBadges
        id="paper"
        venue="International Conference on Long-Tail Engineering"
        venueShort="ICLTE"
        config={config}
        refreshRevision={0}
      />,
    ),
  );
  expect(api.documentJournalRank).toHaveBeenCalledTimes(2);
  expect(
    vi.mocked(api.documentJournalRank).mock.calls.map((call) => call[1]),
  ).toEqual([false, false]);
  expect(host.textContent).toContain("CCF A");
  await act(async () =>
    resolveOldRequest({
      status: "found",
      badges: [{ key: "ccf", name: "CCF", value: "C" }],
    }),
  );
  expect(host.textContent).toContain("CCF A");
  expect(host.textContent).not.toContain("CCF C");
  await act(async () => root.unmount());
});

it("shows the Non-CCF label for an arXiv preprint without an API key", async () => {
  vi.mocked(api.documentJournalRank).mockResolvedValue({
    status: "found",
    badges: [{ key: "ccf", name: "Non-CCF", value: "" }],
  });
  const host = document.createElement("div");
  const root = createRoot(host);
  await act(async () =>
    root.render(
      <JournalRankBadges
        id="paper"
        venue="arXiv"
        config={{
          enabled: true,
          hasKey: false,
          filterMode: "include",
          selected: ["ccf"],
          available: [],
        }}
        refreshRevision={0}
      />,
    ),
  );
  expect(host.querySelector(".journal-rank-badge")?.textContent).toBe(
    "Non-CCF",
  );
  await act(async () => root.unmount());
});

it.each([
  {
    badges: [
      { key: "thcpl", name: "THCPL", value: "A", year: "2019" },
      { key: "core", name: "CORE", value: "A*", year: "2026" },
      { key: "ccf", name: "CCF", value: "A", subject: "Systems" },
    ],
    labels: ["CCF A", "CORE A*", "THCPL A"],
  },
  {
    badges: [
      { key: "thcpl", name: "Non-THCPL", value: "", year: "2019" },
      { key: "core", name: "Non-CORE", value: "", year: "2026" },
      { key: "ccf", name: "Non-CCF", value: "" },
    ],
    labels: ["Non-CCF", "Non-CORE", "Non-THCPL"],
  },
])(
  "orders and compacts the venue badges as $labels",
  async ({ badges, labels }) => {
    vi.mocked(api.documentJournalRank).mockResolvedValue({
      status: "found",
      badges,
    });
    const host = document.createElement("div");
    const root = createRoot(host);
    await act(async () =>
      root.render(
        <JournalRankBadges
          id="paper"
          venue="Example Venue"
          config={{
            enabled: true,
            hasKey: true,
            filterMode: "include",
            selected: ["thcpl", "core", "ccf"],
            available: [],
          }}
          refreshRevision={0}
        />,
      ),
    );
    const elements = [
      ...host.querySelectorAll<HTMLElement>(".journal-rank-badge"),
    ];
    expect(elements.map((element) => element.textContent)).toEqual(labels);
    expect(elements[1].title).toContain("2026");
    expect(elements[2].title).toContain("2019");
    await act(async () => root.unmount());
  },
);
