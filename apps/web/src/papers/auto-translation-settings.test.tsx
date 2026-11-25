// @vitest-environment jsdom
import { act, useState } from "react";
import { createRoot } from "react-dom/client";
import { expect, it, vi } from "vitest";
import { defaultTheme } from "@neoread/core";
import { AutoTranslationSettings } from "./AutoTranslationSettings";

it("turns automatic translation on and off independently of translation features", async () => {
  vi.stubGlobal("IS_REACT_ACT_ENVIRONMENT", true);
  vi.stubGlobal("PointerEvent", MouseEvent);
  const host = document.createElement("div");
  document.body.append(host);
  const root = createRoot(host);
  function Harness() {
    const [theme, setTheme] = useState(defaultTheme);
    return (
      <AutoTranslationSettings
        theme={theme}
        setTheme={(patch) =>
          setTheme((previous) => ({ ...previous, ...patch }))
        }
      />
    );
  }
  try {
    await act(async () => root.render(<Harness />));
    const slider = host.querySelector<HTMLInputElement>('input[type="range"]')!;
    expect(slider.value).toBe("0");
    expect(host.querySelector("p")?.textContent).toBe(
      "Papers translate only on request.",
    );
    const [features, automatic] = Array.from(
      host.querySelectorAll<HTMLButtonElement>('[role="switch"]'),
    );
    await act(async () => automatic.click());
    expect(slider.value).toBe("10");
    await act(async () => automatic.click());
    expect(slider.value).toBe("0");
    await act(async () => {
      slider.dispatchEvent(
        new KeyboardEvent("keydown", { key: "ArrowRight", bubbles: true }),
      );
    });
    expect(host.querySelector("p span")?.textContent).toBe("10");
    expect(slider.getAttribute("aria-valuetext")).toBe("Up to 10 pages");
    await act(async () => {
      slider.dispatchEvent(
        new KeyboardEvent("keydown", { key: "Home", bubbles: true }),
      );
    });
    expect(host.querySelector("p")?.textContent).toBe(
      "Papers translate only on request.",
    );
    expect(slider.value).toBe("0");
    await act(async () => {
      slider.dispatchEvent(
        new KeyboardEvent("keydown", { key: "End", bubbles: true }),
      );
    });
    expect(host.querySelector("p span")?.textContent).toBe("150");
    await act(async () => features.click());
    expect(host.querySelector('input[type="range"]')).toBeNull();
    await act(async () => features.click());
    expect(host.querySelector("p span")?.textContent).toBe("150");
  } finally {
    await act(async () => root.unmount());
    host.remove();
    vi.unstubAllGlobals();
  }
});
