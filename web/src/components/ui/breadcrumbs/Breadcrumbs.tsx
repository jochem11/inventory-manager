import "./breadcrumbs.scss";
import { A } from "@solidjs/router";
import { For, Show } from "solid-js";
import { Icon } from "../icon";

export type Crumb = {
  label: string;
  /** Omit for non-navigable steps (e.g. a nav group). */
  href?: string;
};

export type BreadcrumbsProps = {
  /** Outermost first; the last one is the current page. */
  items: Crumb[];
  class?: string;
};

/** A trail of links ending at the current page. Renders nothing when empty. */
export function Breadcrumbs(props: BreadcrumbsProps) {
  return (
    <Show when={props.items.length > 0}>
      <nav class={`breadcrumbs ${props.class ?? ""}`} aria-label="Breadcrumb">
        <ol class="breadcrumbs__list">
          <For each={props.items}>
            {(crumb, i) => {
              const isLast = () => i() === props.items.length - 1;
              return (
                <li class="breadcrumbs__item" classList={{ "breadcrumbs__item--current": isLast() }}>
                  <Show when={i() > 0}>
                    <Icon name="chevron-right" class="breadcrumbs__separator" />
                  </Show>
                  <Show
                    when={crumb.href && !isLast()}
                    fallback={
                      <span aria-current={isLast() ? "page" : undefined}>{crumb.label}</span>
                    }>
                    <A href={crumb.href!} class="breadcrumbs__link">
                      {crumb.label}
                    </A>
                  </Show>
                </li>
              );
            }}
          </For>
        </ol>
      </nav>
    </Show>
  );
}
