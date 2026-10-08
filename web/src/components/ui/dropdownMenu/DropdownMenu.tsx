import "../button/button.scss";
import "./dropdownMenu.scss";
import { A } from "@solidjs/router";
import {
  children,
  createSignal,
  createUniqueId,
  For,
  onCleanup,
  onMount,
  Show,
  type JSX,
} from "solid-js";
import { Portal } from "solid-js/web";
import { Icon, type IconName } from "../icon";

export type DropdownMenuItem =
  | {
      label: string;
      icon?: IconName;
      /** Renders a link… */
      href?: string;
      /** …or a button. Called after the menu closes (checkbox items keep it open). */
      onSelect?: () => void;
      danger?: boolean;
      /**
       * Makes this a checkbox item (e.g. "show column"): shows a check mark,
       * and selecting it toggles without closing the menu. An accessor, so the
       * items array itself can stay stable while the state changes.
       */
      checked?: () => boolean;
    }
  | { divider: true };

export type DropdownMenuProps = {
  /** Accessible name of the trigger button. */
  label: string;
  /** What the trigger button shows (avatar, icon, text…). */
  trigger: JSX.Element;
  /**
   * `pill`: rounded, with a chevron (account menus). `icon`: small square
   * (row actions). `button`: looks like a small ghost `Button`, with a chevron.
   */
  variant?: "pill" | "icon" | "button";
  /** Optional block above the items (e.g. account details). */
  header?: JSX.Element;
  items: DropdownMenuItem[];
  /** Which edge of the trigger the panel lines up with. */
  align?: "start" | "end";
  class?: string;
};

const isDivider = (item: DropdownMenuItem): item is { divider: true } => "divider" in item;

/**
 * A button that toggles a small panel of links/actions. The panel is portalled
 * to <body> and positioned against the trigger, so scrolling or clipping
 * containers (tables, cards) can't cut it off. Closes on selection, outside
 * click, Escape and Tab; arrow keys move between items.
 */
export function DropdownMenu(props: DropdownMenuProps) {
  const [open, setOpen] = createSignal(false);
  const panelId = createUniqueId();
  let root!: HTMLDivElement;
  let trigger!: HTMLButtonElement;
  let panel: HTMLDivElement | undefined;

  const close = (refocus = false) => {
    setOpen(false);
    if (refocus) trigger.focus();
  };

  onMount(() => {
    const onPointerDown = (e: PointerEvent) => {
      const target = e.target as Node;
      if (open() && !root.contains(target) && !panel?.contains(target)) close();
    };
    document.addEventListener("pointerdown", onPointerDown);
    onCleanup(() => document.removeEventListener("pointerdown", onPointerDown));
  });

  return (
    <div ref={root} class={`dropdown-menu dropdown-menu--${props.variant ?? "pill"} ${props.class ?? ""}`}>
      <button
        ref={trigger}
        type="button"
        class="dropdown-menu__trigger"
        classList={{ "button button--ghost button--sm": props.variant === "button" }}
        aria-haspopup="true"
        aria-expanded={open()}
        aria-controls={open() ? panelId : undefined}
        aria-label={props.label}
        title={props.variant === "icon" ? props.label : undefined}
        onClick={() => setOpen((o) => !o)}
        onKeyDown={(e) => {
          if (e.key === "ArrowDown" && !open()) {
            e.preventDefault();
            setOpen(true);
          }
        }}>
        {props.trigger}
        <Show when={props.variant !== "icon"}>
          <Icon name="chevron-down" class="dropdown-menu__chevron" />
        </Show>
      </button>

      <Show when={open()}>
        <Portal>
          <Panel
            {...props}
            id={panelId}
            anchor={trigger}
            ref={(el) => (panel = el)}
            onClose={close}
          />
        </Portal>
      </Show>
    </div>
  );
}

type PanelProps = DropdownMenuProps & {
  id: string;
  anchor: HTMLElement;
  ref: (el: HTMLDivElement) => void;
  onClose: (refocus?: boolean) => void;
};

/**
 * Only mounted while open, so the `header` (and items) are created lazily —
 * building them up front would create elements the server never rendered.
 */
function Panel(props: PanelProps) {
  const header = children(() => props.header);
  const [position, setPosition] = createSignal<JSX.CSSProperties>({ visibility: "hidden" });
  let el!: HTMLDivElement;

  const GAP = 6;

  // Below the trigger, or above it when there's no room below.
  const place = () => {
    const anchor = props.anchor.getBoundingClientRect();
    const height = el.offsetHeight;
    const fitsBelow = anchor.bottom + GAP + height <= window.innerHeight - GAP;
    const top = fitsBelow ? anchor.bottom + GAP : Math.max(GAP, anchor.top - GAP - height);
    const horizontal =
      props.align === "start"
        ? { left: `${anchor.left}px` }
        : { right: `${document.documentElement.clientWidth - anchor.right}px` };
    setPosition({ top: `${top}px`, ...horizontal });
  };

  const items = () => [...el.querySelectorAll<HTMLElement>(".dropdown-menu__item")];

  const onKeyDown = (e: KeyboardEvent) => {
    const list = items();
    const index = list.indexOf(document.activeElement as HTMLElement);
    if (e.key === "ArrowDown" || e.key === "ArrowUp") {
      e.preventDefault();
      const step = e.key === "ArrowDown" ? 1 : -1;
      list[(index + step + list.length) % list.length]?.focus();
    } else if (e.key === "Home" || e.key === "End") {
      e.preventDefault();
      (e.key === "Home" ? list[0] : list.at(-1))?.focus();
    } else if (e.key === "Escape") {
      e.preventDefault();
      props.onClose(true);
    } else if (e.key === "Tab") {
      // The panel lives at the end of <body>; tabbing on would jump far
      // away, so close and hand focus back to the trigger instead.
      e.preventDefault();
      props.onClose(true);
    }
  };

  onMount(() => {
    place();
    items()[0]?.focus({ preventScroll: true });
    window.addEventListener("scroll", place, true);
    window.addEventListener("resize", place);
    onCleanup(() => {
      window.removeEventListener("scroll", place, true);
      window.removeEventListener("resize", place);
    });
  });

  return (
    <div
      ref={(node) => {
        el = node;
        props.ref(node);
      }}
      id={props.id}
      class="dropdown-menu__panel"
      classList={{ "dropdown-menu__panel--start": props.align === "start" }}
      style={position()}
      onKeyDown={onKeyDown}>
      <Show when={header()}>
        <div class="dropdown-menu__header">{header()}</div>
      </Show>

      <ul class="dropdown-menu__list">
        <For each={props.items}>
          {(item) => {
            if (isDivider(item)) return <li class="dropdown-menu__divider" role="separator" />;

            const content = (
              <>
                <Show when={item.icon}>{(icon) => <Icon name={icon()} />}</Show>
                {item.label}
              </>
            );
            const select = () => {
              if (!item.checked) props.onClose(true);
              item.onSelect?.();
            };

            return (
              <li>
                <Show
                  when={item.href}
                  fallback={
                    <button
                      type="button"
                      class="dropdown-menu__item"
                      classList={{ "dropdown-menu__item--danger": item.danger }}
                      aria-pressed={item.checked ? item.checked() : undefined}
                      onClick={select}>
                      <Show when={item.checked}>
                        {(checked) => (
                          <span
                            class="dropdown-menu__check"
                            classList={{ "dropdown-menu__check--on": checked()() }}>
                            <Icon name="check" />
                          </span>
                        )}
                      </Show>
                      {content}
                    </button>
                  }>
                  {(href) => (
                    <A
                      href={href()}
                      class="dropdown-menu__item"
                      classList={{ "dropdown-menu__item--danger": item.danger }}
                      onClick={() => props.onClose()}>
                      {content}
                    </A>
                  )}
                </Show>
              </li>
            );
          }}
        </For>
      </ul>
    </div>
  );
}
