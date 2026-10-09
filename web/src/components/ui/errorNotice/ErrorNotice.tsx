import { Show } from "solid-js";
import { Notice } from "../notice";

export type ErrorNoticeProps = {
  /** Nothing is rendered without a message. */
  message?: string;
  /** Shows a "Try again" button. */
  onRetry?: () => void;
};

/** An error Notice for a message that may be empty, e.g. a failed action's. */
export function ErrorNotice(props: ErrorNoticeProps) {
  return (
    <Show when={props.message}>
      <Notice tone="error" onRetry={props.onRetry}>
        {props.message}.
      </Notice>
    </Show>
  );
}
