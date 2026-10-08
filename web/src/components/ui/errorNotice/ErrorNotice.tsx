import "./errorNotice.scss";
import { Show } from "solid-js";

export type ErrorNoticeProps = {
  message?: string;
  onRetry?: () => void;
};

export function ErrorNotice(props: ErrorNoticeProps) {
  return (
    <Show when={props.message}>
      <p class="error-notice" role="alert">
        {props.message}.{" "}
        <Show when={props.onRetry}>
          <button type="button" onClick={() => props.onRetry?.()}>
            Try again
          </button>
        </Show>
      </p>
    </Show>
  );
}
