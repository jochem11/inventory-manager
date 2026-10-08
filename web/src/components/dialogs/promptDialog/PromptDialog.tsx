import "./promptDialog.scss";
import { createSignal, createUniqueId } from "solid-js";
import { Button } from "~/components/ui";
import { defineDialog } from "~/factories/defineDialog";
import { BaseDialog } from "../baseDialog/BaseDialog";

export type PromptDialogProps = {
  title: string;
  label: string;
  message?: string;
  placeholder?: string;
  initialValue?: string;
  submitLabel?: string;
};

/**
 * Asks the user for a single line of text. `open()` resolves with the trimmed
 * value on submit, or `undefined` when cancelled / dismissed.
 */
export const PromptDialog = defineDialog<PromptDialogProps, string>((props) => {
  const inputId = createUniqueId();
  const [value, setValue] = createSignal(props.initialValue ?? "");

  return (
    <BaseDialog
      title={props.title}
      description={props.message}
      onClose={() => props.close()}>
      <form
        class="prompt-dialog__form"
        onSubmit={(e) => {
          e.preventDefault();
          const trimmed = value().trim();
          if (trimmed) props.close(trimmed);
        }}>
        <label for={inputId} class="prompt-dialog__label">
          {props.label}
        </label>
        <input
          id={inputId}
          class="prompt-dialog__input"
          placeholder={props.placeholder}
          value={value()}
          onInput={(e) => setValue(e.currentTarget.value)}
          autofocus
          required
        />
        <div class="base-dialog__actions">
          <Button variant="outline" onClick={() => props.close()}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={!value().trim()}>
            {props.submitLabel ?? "Save"}
          </Button>
        </div>
      </form>
    </BaseDialog>
  );
});
