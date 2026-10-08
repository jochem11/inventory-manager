import { createContext, useContext, type Accessor } from "solid-js";
import { ContextError } from "~/errors/contextError";
import type { Theme } from "~/constants/theme";

export type ThemeContextValue = {
  theme: Accessor<Theme>;
  setTheme: (theme: Theme) => void;
  toggle: () => void;
};

export const ThemeContext = createContext<ThemeContextValue>();

export const useTheme = () => {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new ContextError("useTheme must be used inside <ThemeProvider>");
  return ctx;
};
