import type { z } from "zod";
import { FormDialog, type FormDialogProps } from "~/components/dialogs/formDialog";
import { createDialog } from "./createDialog";

/**
 * Opens `FormDialog`s. `open()` takes the form (schema, start values, fields)
 * and resolves with the validated values, typed by the schema, or `undefined`
 * when cancelled.
 *
 *   const formDialog = createFormDialog();
 *   const values = await formDialog.open({
 *     title: "New category",
 *     schema: z.object({ name: z.string().min(1) }),
 *     fields: () => <TextField name="name" label="Name" />,
 *   });
 */
export function createFormDialog() {
  const dialog = createDialog(FormDialog);
  return {
    open: <Schema extends z.ZodType>(props: FormDialogProps<Schema>) =>
      // FormDialog is defined once for any schema; this is where the
      // schema's type is put back on the result.
      dialog.open(props as unknown as FormDialogProps) as Promise<z.output<Schema> | undefined>,
    close: dialog.close,
  };
}
