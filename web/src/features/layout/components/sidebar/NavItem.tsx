import "./navItem.scss";
import { A, useLocation } from "@solidjs/router";
import { createEffect, createSignal, createUniqueId, For, Show } from "solid-js";
import { Icon } from "~/components/ui";
import type { NavGroup, NavLink } from "~/types/navigation";

type NavLinkItemProps = {
  item: NavLink;
  /** Rendered inside a group: smaller, no icon column. */
  nested?: boolean;
  /** Sidebar is in icon-only mode, so expose the label as a tooltip. */
  collapsed?: boolean;
};

export function NavLinkItem(props: NavLinkItemProps) {
  return (
    <li>
      <A
        href={props.item.href}
        end={props.item.end}
        class="nav-item"
        classList={{ "nav-item--nested": props.nested }}
        activeClass="nav-item--active"
        title={props.collapsed ? props.item.label : undefined}>
        <Show when={props.item.icon}>
          {(icon) => <Icon name={icon()} class="nav-item__icon" />}
        </Show>
        <span class="nav-item__label">{props.item.label}</span>
      </A>
    </li>
  );
}

type NavGroupItemProps = {
  item: NavGroup;
  collapsed?: boolean;
  /** Called when the group is clicked while the sidebar is icon-only. */
  onExpandSidebar?: () => void;
};

/** A disclosure in the sidebar; opens itself when one of its pages is active. */
export function NavGroupItem(props: NavGroupItemProps) {
  const location = useLocation();
  const panelId = createUniqueId();

  const containsActive = () =>
    props.item.children.some((c) =>
      c.end
        ? location.pathname === c.href
        : location.pathname === c.href || location.pathname.startsWith(`${c.href}/`),
    );

  const [open, setOpen] = createSignal(containsActive());

  // Navigating into the group (e.g. via a breadcrumb or link) reveals it.
  createEffect(() => {
    if (containsActive()) setOpen(true);
  });

  const toggle = () => {
    if (props.collapsed && props.onExpandSidebar) {
      props.onExpandSidebar();
      setOpen(true);
    } else {
      setOpen((o) => !o);
    }
  };

  return (
    <li
      class="nav-group"
      classList={{ "nav-group--open": open(), "nav-group--active": containsActive() }}>
      <button
        type="button"
        class="nav-item nav-group__trigger"
        aria-expanded={open()}
        aria-controls={panelId}
        title={props.collapsed ? props.item.label : undefined}
        onClick={toggle}>
        <Icon name={props.item.icon} class="nav-item__icon" />
        <span class="nav-item__label">{props.item.label}</span>
        <Icon name="chevron-down" class="nav-group__chevron" />
      </button>

      {/* `inert` keeps links in a closed group out of the tab order. */}
      <div id={panelId} class="nav-group__panel" inert={!open() || undefined}>
        <div class="nav-group__inner">
          <ul class="nav-group__list">
            <For each={props.item.children}>
              {(child) => <NavLinkItem item={child} nested />}
            </For>
          </ul>
        </div>
      </div>
    </li>
  );
}
