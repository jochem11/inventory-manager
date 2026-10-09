import "./formDialog.scss";
import { createSignal, type JSX } from "solid-js";
import type { z } from "zod";
import { Form, SubmitButton } from "~/components/form";
import { Button, ErrorNotice } from "~/components/ui";
import { defineDialog } from "~/factories/defineDialog";
import { createForm } from "~/hooks/createForm";
import { BaseDialog } from "../baseDialog/BaseDialog";

export type FormDialogProps<Schema extends z.ZodType = z.ZodType> = {
  title: string;
  description?: string;
  /** Validates the fields; `open()` resolves with its parsed output. */
  schema: Schema;
  defaultValues?: Partial<z.input<Schema>>;
  /**
   * The form's fields (`TextField`, `SelectField`, …). A function, so they're
   * created inside the dialog's form and find it through context.
   */
  fields: () => JSX.Element;
  submitLabel?: string;
  submittingLabel?: string;
  /**
   * Runs on submit with the valid values, while the dialog stays open and
   * shows progress, e.g. to save them. If it throws, the dialog stays open
   * and shows the error; otherwise it closes with the values.
   */
  onSubmit?: (values: z.output<Schema>) => unknown | Promise<unknown>;
  /** Turns an error from `onSubmit` into a message. Defaults to its message. */
  formatError?: (error: unknown) => string;
};

const defaultFormatError = (error: unknown) => (error instanceof Error ? error.message : String(error));

/**
 * A form in a dialog. `open()` resolves with the validated values on submit,
 * or `undefined` when cancelled. Open it with `createFormDialog()`, which keeps
 * the values typed by the schema.
 */
export const FormDialog = defineDialog<FormDialogProps, unknown>((props) => {
  const [error, setError] = createSignal<string>();

  const form = createForm({
    schema: props.schema,
    defaultValues: props.defaultValues,
    onSubmit: async (values) => {
      setError();
      try {
        await props.onSubmit?.(values);
        props.close(values);
      } catch (e) {
        setError((props.formatError ?? defaultFormatError)(e));
      }
    },
  });

  return (
    <BaseDialog title={props.title} description={props.description} onClose={() => props.close()}>
      <Form form={form} class="form-dialog__form">
        <ErrorNotice message={error()} />
        {props.fields()}
        <div class="base-dialog__actions">
          <Button variant="outline" onClick={() => props.close()}>
            Cancel
          </Button>
          <SubmitButton submittingText={props.submittingLabel ?? "Saving…"}>{props.submitLabel ?? "Save"}</SubmitButton>
        </div>
      </Form>
    </BaseDialog>
  );
});
