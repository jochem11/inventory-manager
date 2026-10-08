import { type JSX } from "solid-js";
import type { z } from "zod";
import { FormContext } from "~/context/formContext";
import { createForm, type FormApi } from "~/hooks/createForm";
import "./form.scss";

export { FORM_LEVEL_KEY } from "~/hooks/createForm";

type BaseFormProps = {
  class?: string;
  children: JSX.Element;
};

export type FormProps<Schema extends z.ZodType> = BaseFormProps &
  (
    | {
        /** A pre-built instance from `createForm`. */
        form: FormApi<Schema>;
        schema?: never;
        defaultValues?: never;
        onSubmit?: never;
      }
    | {
        form?: never;
        /** Zod schema; its parsed *output* type is what `onSubmit` gets. */
        schema: Schema;
        /** Seed values; anything omitted starts `undefined`. */
        defaultValues?: Partial<z.input<Schema>>;
        /** Called with validated, parsed data once the form is valid. */
        onSubmit: (values: z.output<Schema>) => void | Promise<void>;
      }
  );

/**
 * A schema-driven form provider. Give it a `schema` + `onSubmit` (it builds the
 * form for you) or a `form` from `createForm` (when you need to drive validity,
 * reset, etc. from outside). Either way it exposes the form through context, so
 * the inputs in this folder (`TextField`, `SelectField`, …) wire themselves up
 * via `useForm()`. `onSubmit` only fires when every field passes; touched fields
 * re-validate live as the user types.
 */
export function Form<Schema extends z.ZodType>(props: FormProps<Schema>) {
  // Solid component bodies run once, so this creates a single stable instance.
  // `onSubmit` is read live so a changing prop is always respected.
  const form =
    props.form ??
    createForm<Schema>({
      schema: props.schema!,
      defaultValues: props.defaultValues,
      onSubmit: (values) => props.onSubmit!(values),
    });

  return (
    <FormContext.Provider value={form}>
      <form
        class={`form${props.class ? ` ${props.class}` : ""}`}
        novalidate
        onSubmit={form.handleSubmit}>
        {props.children}
      </form>
    </FormContext.Provider>
  );
}
