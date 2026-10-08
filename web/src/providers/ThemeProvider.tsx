import { createEffect, createSignal, type ParentComponent } from "solid-js";
import { isServer } from "solid-js/web";
import {
  DEFAULT_THEME,
  isTheme,
  Theme,
  THEME_STORAGE_KEY,
} from "~/constants/theme";
import { ThemeContext, type ThemeContextValue } from "~/context";

const getInitialTheme = (): Theme => {
  if (isServer) return DEFAULT_THEME;
  const saved = localStorage.getItem(THEME_STORAGE_KEY);
  if (isTheme(saved)) return saved;
  return matchMedia("(prefers-color-scheme: dark)").matches
    ? Theme.Dark
    : Theme.Light;
};

export const ThemeProvider: ParentComponent = (props) => {
  const [theme, setTheme] = createSignal<Theme>(getInitialTheme());

  createEffect(() => {
    document.documentElement.dataset.theme = theme();
    localStorage.setItem(THEME_STORAGE_KEY, theme());
  });

  const value: ThemeContextValue = {
    theme,
    setTheme,
    toggle: () =>
      setTheme((t) => (t === Theme.Light ? Theme.Dark : Theme.Light)),
  };

  return (
    <ThemeContext.Provider value={value}>
      {props.children}
    </ThemeContext.Provider>
  );
};
