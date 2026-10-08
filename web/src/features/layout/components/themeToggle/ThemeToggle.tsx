import "./themeToggle.scss";
import { Button, Icon } from "~/components/ui";
import { useTheme } from "~/context";

/**
 * Light/dark switch. Both icons are always rendered and CSS picks one from
 * `<html data-theme>`, which the inline script in entry-server sets before
 * hydration — so the server (which can't know the theme) never mismatches.
 */
export function ThemeToggle() {
  const theme = useTheme();

  return (
    <Button
      variant="ghost"
      iconOnly
      class="theme-toggle"
      aria-label="Toggle dark mode"
      title="Toggle dark mode"
      onClick={() => theme.toggle()}>
      <Icon name="moon" class="theme-toggle__moon" />
      <Icon name="sun" class="theme-toggle__sun" />
    </Button>
  );
}
