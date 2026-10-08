import "./topbar.scss";
import type { JSX } from "solid-js";

export type TopbarProps = {
  /** Left side: menu button, breadcrumbs, title… */
  start?: JSX.Element;
  /** Right side: search, toggles, account menu… */
  end?: JSX.Element;
  class?: string;
};

/** Sticky, translucent bar across the top of the content area. */
export function Topbar(props: TopbarProps) {
  return (
    <header class={`topbar ${props.class ?? ""}`}>
      <div class="topbar__start">{props.start}</div>
      <div class="topbar__end">{props.end}</div>
    </header>
  );
}
