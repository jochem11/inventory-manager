import "./button.scss";
import { children, Show, splitProps, type JSX } from "solid-js";

export type ButtonVariant =
  | "primary"
  | "secondary"
  | "accent"
  | "outline"
  | "ghost"
  | "danger"
  | "link";
export type ButtonSize = "sm" | "md" | "lg";

export type ButtonProps = Omit<JSX.ButtonHTMLAttributes<HTMLButtonElement>, "class"> & {
  variant?: ButtonVariant;
  size?: ButtonSize;
  /** Stretch to the full width of the container. */
  block?: boolean;
  /** Shows a spinner and disables the button (e.g. while a request runs). */
  loading?: boolean;
  /** Square button for a lone icon; pass an `aria-label` alongside it. */
  iconOnly?: boolean;
  /** Rendered before the label (hidden while `loading`, the spinner takes its place). */
  startIcon?: JSX.Element;
  /** Rendered after the label. */
  endIcon?: JSX.Element;
  class?: string;
};

/**
 * The app's shared button. Defaults to `type="button"` so it never submits a
 * form by accident — pass `type="submit"` (or use `SubmitButton`) for that.
 */
export function Button(props: ButtonProps) {
  const [local, rest] = splitProps(props, [
    "variant",
    "size",
    "block",
    "loading",
    "iconOnly",
    "startIcon",
    "endIcon",
    "class",
    "children",
    "disabled",
    "type",
  ]);
  // Resolve once: reading `children` twice (check + render) would create the
  // elements twice, which breaks hydration.
  const label = children(() => local.children);
  const startIcon = children(() => local.startIcon);
  const endIcon = children(() => local.endIcon);

  return (
    <button
      {...rest}
      type={local.type ?? "button"}
      class={`button button--${local.variant ?? "secondary"} button--${local.size ?? "md"} ${local.class ?? ""}`}
      classList={{
        "button--block": local.block,
        "button--icon-only": local.iconOnly,
        "button--loading": local.loading,
      }}
      disabled={local.disabled || local.loading}
      aria-busy={local.loading || undefined}>
      <Show
        when={local.loading}
        fallback={
          <Show when={startIcon()}>
            <span class="button__icon">{startIcon()}</span>
          </Show>
        }>
        <span class="button__spinner" aria-hidden="true" />
      </Show>
      <Show when={label()}>
        <span class="button__label">{label()}</span>
      </Show>
      <Show when={endIcon()}>
        <span class="button__icon">{endIcon()}</span>
      </Show>
    </button>
  );
}
