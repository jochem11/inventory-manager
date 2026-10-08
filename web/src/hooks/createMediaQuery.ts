import { createSignal, onCleanup, onMount, type Accessor } from "solid-js";

/**
 * Whether a CSS media query currently matches, kept up to date. Always `false`
 * during SSR and the first client render, so hydration stays consistent.
 */
export function createMediaQuery(query: string): Accessor<boolean> {
  const [matches, setMatches] = createSignal(false);

  onMount(() => {
    const mql = matchMedia(query);
    setMatches(mql.matches);
    const onChange = (e: MediaQueryListEvent) => setMatches(e.matches);
    mql.addEventListener("change", onChange);
    onCleanup(() => mql.removeEventListener("change", onChange));
  });

  return matches;
}
