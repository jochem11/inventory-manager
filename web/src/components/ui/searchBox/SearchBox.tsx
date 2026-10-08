import "./searchBox.scss";
import { createEffect, createSignal, onCleanup, onMount, Show } from "solid-js";
import { Icon } from "../icon";

export type SearchBoxProps = {
  /** Current query (e.g. from the URL); the box follows it when it changes. */
  value?: string;
  /** Called with the trimmed query on Enter. */
  onSearch: (query: string) => void;
  placeholder?: string;
  /** Accessible name; defaults to the placeholder. */
  label?: string;
  /** Focus the box with ⌘K / Ctrl+K from anywhere on the page. */
  shortcut?: boolean;
  class?: string;
};

/** Pill-shaped search input that reports the query on submit. */
export function SearchBox(props: SearchBoxProps) {
  const [value, setValue] = createSignal(props.value ?? "");
  const [shortcutLabel, setShortcutLabel] = createSignal("Ctrl K");
  let input!: HTMLInputElement;

  createEffect(() => setValue(props.value ?? ""));

  onMount(() => {
    if (!props.shortcut) return;
    if (/Mac|iPhone|iPad/.test(navigator.platform)) setShortcutLabel("⌘K");

    const onKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        input.focus();
        input.select();
      }
    };
    document.addEventListener("keydown", onKeyDown);
    onCleanup(() => document.removeEventListener("keydown", onKeyDown));
  });

  const placeholder = () => props.placeholder ?? "Search…";

  return (
    <form
      role="search"
      class={`search-box ${props.class ?? ""}`}
      onSubmit={(e) => {
        e.preventDefault();
        props.onSearch(value().trim());
      }}>
      <Icon name="search" class="search-box__icon" />
      <input
        ref={input}
        type="search"
        name="q"
        class="search-box__input"
        placeholder={placeholder()}
        aria-label={props.label ?? placeholder()}
        value={value()}
        onInput={(e) => setValue(e.currentTarget.value)}
        onKeyDown={(e) => {
          if (e.key === "Escape") e.currentTarget.blur();
        }}
      />
      <Show when={props.shortcut}>
        <kbd class="search-box__kbd" aria-hidden="true">
          {shortcutLabel()}
        </kbd>
      </Show>
    </form>
  );
}
