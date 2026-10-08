import { createContext, useContext } from "solid-js";

/** A field's current validation message, or `undefined` when it's valid. */
export type FieldError = string | undefined;

/**
 * The shape every input reads/writes through. Deliberately untyped over the
 * field values (`unknown`) so the context can be created once at module load;
 * the `<Form>` component is what carries the schema's static types.
 */
export type FormContextValue = {
  /** Current value of a field (`undefined` until it's first set). */
  value: (name: string) => unknown;
  /** The error to *display* for a field — only surfaces once touched. */
  error: (name: string) => FieldError;
  /** Whether the field has been edited or blurred yet. */
  touched: (name: string) => boolean;
  /** Update a field's value (revalidates it once it's been touched). */
  setValue: (name: string, value: unknown) => void;
  /** Mark a field touched (typically on blur) so its error can show. */
  setTouched: (name: string) => void;
  /** True while the async `onSubmit` handler is in flight. */
  submitting: () => boolean;
};

export const FormContext = createContext<FormContextValue>();

export const useForm = () => {
  const ctx = useContext(FormContext);
  if (!ctx) throw new Error("useForm must be used within a <Form>");
  return ctx;
};
