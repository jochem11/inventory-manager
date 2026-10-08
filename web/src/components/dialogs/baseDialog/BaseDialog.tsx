import "./baseDialog.scss";
import { children, createUniqueId, onMount, Show, type JSX, type ParentProps } from "solid-js";

export type BaseDialogProps = ParentProps<{
  title: string;
  description?: string;
  /** Rendered in the footer, right-aligned (typically buttons). */
  actions?: JSX.Element;
  /** When set, shows a close (×) button in the header. */
  onClose?: () => void;
  class?: string;
}>;

/**
 * Presentational layout for dialog content (header, body, footer). It renders
 * *inside* the native <dialog> owned by the DialogProvider, so it only wires
 * up labelling on that element and leaves open/close state to the provider.
 */
export const BaseDialog = (props: BaseDialogProps) => {
  let el!: HTMLDivElement;
  const titleId = createUniqueId();
  const descriptionId = createUniqueId();
  const body = children(() => props.children);
  const actions = children(() => props.actions);

  onMount(() => {
    const dialog = el.closest("dialog");
    dialog?.setAttribute("aria-labelledby", titleId);
    if (props.description) dialog?.setAttribute("aria-describedby", descriptionId);
  });

  return (
    <div ref={el} class={`base-dialog ${props.class ?? ""}`}>
      <header class="base-dialog__header">
        <h2 id={titleId} class="base-dialog__title">
          {props.title}
        </h2>
        <Show when={props.onClose}>
          <button
            type="button"
            class="base-dialog__close"
            aria-label="Close"
            onClick={() => props.onClose?.()}>
            ×
          </button>
        </Show>
      </header>

      <Show when={props.description}>
        <p id={descriptionId} class="base-dialog__description">
          {props.description}
        </p>
      </Show>

      <Show when={body()}>
        <div class="base-dialog__body">{body()}</div>
      </Show>

      <Show when={actions()}>
        <footer class="base-dialog__actions">{actions()}</footer>
      </Show>
    </div>
  );
};
