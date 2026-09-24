import { useEffect, useState, type ReactNode } from "react";
import { ThemeContext } from "./themeContextStore";

export type Theme = "dark" | "light" | "system";

const storageKey = "forgeai-theme";

function getStoredTheme(): Theme {
  const stored = window.localStorage.getItem(storageKey);
  return stored === "light" || stored === "system" || stored === "dark"
    ? stored
    : "dark";
}

function getSystemTheme(): "dark" | "light" {
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

function resolveTheme(theme: Theme): "dark" | "light" {
  return theme === "system" ? getSystemTheme() : theme;
}

function applyTheme(theme: Theme) {
  const resolved = resolveTheme(theme);
  document.documentElement.dataset.themeChoice = theme;
  document.documentElement.dataset.theme = resolved;
  document.documentElement.style.colorScheme = resolved;
  return resolved;
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(getStoredTheme);
  const [resolvedTheme, setResolvedTheme] = useState<"dark" | "light">(() => resolveTheme(theme));

  useEffect(() => {
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const syncTheme = () => {
      setResolvedTheme(applyTheme(theme));
    };

    syncTheme();
    const onChange = () => {
      if (theme === "system") syncTheme();
    };
    if (typeof media.addEventListener === "function") {
      media.addEventListener("change", onChange);
      return () => media.removeEventListener("change", onChange);
    }
    media.addListener(onChange);
    return () => media.removeListener(onChange);
  }, [theme]);

  function setTheme(nextTheme: Theme) {
    window.localStorage.setItem(storageKey, nextTheme);
    setResolvedTheme(applyTheme(nextTheme));
    setThemeState(nextTheme);
  }

  return (
    <ThemeContext.Provider value={{ theme, resolvedTheme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

