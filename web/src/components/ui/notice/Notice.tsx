import "./notice.scss";
import { Show, type JSX } from "solid-js";

export type NoticeProps = {
  /** Default "info". Errors are announced to screen readers right away. */
  tone?: "error" | "success" | "info";
  /** Shows a "Try again" button after the content. */
  onRetry?: () => void;
  children: JSX.Element;
};

/** A short message in a tinted box: an error, a confirmation or a hint. */
export function Notice(props: NoticeProps) {
  return (
    <div class={`notice notice--${props.tone ?? "info"}`} role={props.tone === "error" ? "alert" : "status"}>
      {props.children}
      <Show when={props.onRetry}>
        {" "}
        <button type="button" onClick={() => props.onRetry?.()}>
          Try again
        </button>
      </Show>
    </div>
  );
}
