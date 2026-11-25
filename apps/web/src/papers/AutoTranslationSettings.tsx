import type { ReaderTheme } from "@neoread/core";
import { Slider } from "@neoread/ui/components/slider";
import { Switch } from "@neoread/ui/components/switch";

export function AutoTranslationSettings({
  theme,
  setTheme,
  showAutomatic = true,
}: {
  theme: ReaderTheme;
  showAutomatic?: boolean;
  setTheme: (patch: Partial<ReaderTheme>) => void;
}) {
  const limit = theme.autoTranslateMaxPages ?? 0;
  const enabled = theme.translationEnabled !== false;
  const showLimit = enabled && showAutomatic;
  return (
    <div className="mt-5 space-y-3 text-sm">
      <label className="flex items-center justify-between gap-3">
        <span className="font-medium">Translation features</span>
        <Switch
          aria-label="Translation features"
          checked={enabled}
          onCheckedChange={(checked) =>
            setTheme({ translationEnabled: checked })
          }
        />
      </label>
      {showLimit && (
        <label className="flex items-center justify-between gap-3">
          <span className="font-medium">Automatically translate papers</span>
          <Switch
            checked={limit > 0}
            onCheckedChange={(checked) =>
              setTheme({ autoTranslateMaxPages: checked ? 10 : 0 })
            }
          />
        </label>
      )}
      {showLimit && (
        <>
          <p>
            {limit === 0 ? (
              "Papers translate only on request."
            ) : (
              <>
                Automatically translate papers up to{" "}
                <span className="rounded bg-muted px-2 py-0.5 tabular-nums">
                  {limit}
                </span>{" "}
                pages.
              </>
            )}
          </p>
          <Slider
            thumbLabel="Automatic translation page limit"
            // Position directly from the value before the settings panel has measurable dimensions.
            thumbAlignment="center"
            min={0}
            max={150}
            step={10}
            value={[limit]}
            getAriaValueText={(_, value) =>
              value === 0 ? "Manual translation only" : `Up to ${value} pages`
            }
            onValueChange={(value) =>
              setTheme({
                autoTranslateMaxPages: Array.isArray(value) ? value[0] : value,
              })
            }
          />
        </>
      )}
    </div>
  );
}
