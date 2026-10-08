import "./sidebar.scss";
import { A } from "@solidjs/router";
import { children, For, Show, type JSX } from "solid-js";
import { Button, Icon } from "~/components/ui";
import { NavGroupItem, NavLinkItem } from "./NavItem";
import { isNavGroup, type NavItem, type SidebarBrand } from "~/types/navigation";

export type SidebarProps = {
  id?: string;
  /** Accessible name of the navigation landmark. */
  label?: string;
  brand: SidebarBrand;
  items: NavItem[];
  /** Extra content in the footer, hidden while collapsed (version, status…). */
  footer?: JSX.Element;
  /** Icon-only mode: labels, subnavs and the footer content are hidden. */
  collapsed?: boolean;
  /** Shows the collapse toggle in the footer when given. */
  onToggleCollapsed?: () => void;
  /** Shows a close (×) button next to the brand when given (drawer use). */
  onClose?: () => void;
  class?: string;
};

/**
 * Vertical app navigation: brand, nav links with collapsible groups, and a
 * footer. It only renders itself — where it sits (sticky column, drawer) is
 * up to the parent.
 */
export function Sidebar(props: SidebarProps) {
  const footer = children(() => props.footer);

  return (
    <aside
      id={props.id}
      class={`sidebar ${props.class ?? ""}`}
      classList={{ "sidebar--collapsed": props.collapsed }}
      aria-label={props.label ?? "Main navigation"}>
      <div class="sidebar__brand">
        <A
          href={props.brand.href}
          class="sidebar__logo"
          title={props.collapsed ? props.brand.label : undefined}>
          <span class="sidebar__logo-mark">
            <Icon name={props.brand.icon} />
          </span>
          <span class="sidebar__logo-text">{props.brand.label}</span>
        </A>
        <Show when={props.onClose}>
          <Button
            variant="ghost"
            iconOnly
            aria-label="Close navigation"
            onClick={() => props.onClose?.()}>
            <Icon name="x" />
          </Button>
        </Show>
      </div>

      <nav class="sidebar__nav">
        <ul class="sidebar__list">
          <For each={props.items}>
            {(item) =>
              isNavGroup(item) ? (
                <NavGroupItem
                  item={item}
                  collapsed={props.collapsed}
                  onExpandSidebar={props.onToggleCollapsed}
                />
              ) : (
                <NavLinkItem item={item} collapsed={props.collapsed} />
              )
            }
          </For>
        </ul>
      </nav>

      <Show when={props.onToggleCollapsed || footer()}>
        <div class="sidebar__footer">
          <Show when={props.onToggleCollapsed}>
            <Button
              variant="ghost"
              iconOnly
              aria-label={props.collapsed ? "Expand sidebar" : "Collapse sidebar"}
              aria-pressed={!!props.collapsed}
              title={props.collapsed ? "Expand sidebar" : "Collapse sidebar"}
              onClick={() => props.onToggleCollapsed?.()}>
              <Icon name="panel-left" />
            </Button>
          </Show>
          <div class="sidebar__footer-content">{footer()}</div>
        </div>
      </Show>
    </aside>
  );
}
