import { Button } from "~/components/ui";
import { defineDialog } from "~/factories/defineDialog";
import { BaseDialog } from "../baseDialog/BaseDialog";

export type ConfirmDialogProps = {
  title: string;
  message?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  /** `danger` styles the confirm button for destructive actions. */
  variant?: "default" | "danger";
};

/**
 * Asks the user to confirm an action. `open()` resolves with `true` when
 * confirmed, `false` when cancelled, and `undefined` when dismissed via
 * Escape / backdrop — so `if (await confirm.open(...))` covers all cases.
 */
export const ConfirmDialog = defineDialog<ConfirmDialogProps, boolean>(
  (props) => (
    <BaseDialog
      title={props.title}
      description={props.message}
      onClose={() => props.close(false)}
      actions={
        <>
          <Button variant="outline" onClick={() => props.close(false)}>
            {props.cancelLabel ?? "Cancel"}
          </Button>
          <Button
            variant={props.variant === "danger" ? "danger" : "primary"}
            autofocus
            onClick={() => props.close(true)}>
            {props.confirmLabel ?? "Confirm"}
          </Button>
        </>
      }
    />
  ),
);
