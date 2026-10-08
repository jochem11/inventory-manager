// ~/constants/theme.ts
export enum Theme {
  Light = "light",
  Dark = "dark",
}

export const THEME_STORAGE_KEY = "theme";
export const DEFAULT_THEME = Theme.Light;

// Type guard: checks whether a random string (e.g. from localStorage) is a valid Theme.
export const isTheme = (value: unknown): value is Theme =>
  Object.values(Theme).includes(value as Theme);
