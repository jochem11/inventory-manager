import { createSignal } from "solid-js";
import { createStore, reconcile } from "solid-js/store";
import type { z } from "zod";
import type { FormContextValue } from "~/context/formContext";

/** Field key used for whole-form (schema-level, non-field) errors. */
export const FORM_LEVEL_KEY = "_form";

export type CreateFormOptions<Schema extends z.ZodType> = {
  /** Zod schema; its parsed *output* type is what `onSubmit` receives. */
  schema: Schema;
  /** Seed values. Anything omitted starts `undefined` and validates as such. */
  defaultValues?: Partial<z.input<Schema>>;
  /**
   * Called with validated, parsed data — only ever when the whole form is
   * valid. May be async; the form reports `submitting` while it runs.
   */
  onSubmit: (values: z.output<Schema>) => void | Promise<void>;
};

/**
 * Everything `createForm` hands back: the reactive `FormContextValue` the input
 * components read through, plus imperative helpers for driving the form yourself
 * (submit, reset, inspect validity).
 */
export type FormApi<Schema extends z.ZodType> = FormContextValue & {
  /** The reactive value store — read fields off it or pass it around. */
  values: z.input<Schema>;
  /** Full current error map (every field, regardless of touched state). */
  errors: () => Record<string, string>;
  /** Whether the current values pass the schema *right now*. */
  isValid: () => boolean;
  /** Reset values/errors/touched back to `defaultValues` (or the given ones). */
  reset: (next?: Partial<z.input<Schema>>) => void;
  /** Submit handler for `<form onSubmit>`: validates, then calls `onSubmit`. */
  handleSubmit: (event?: Event) => Promise<void>;
};

/**
 * Headless form primitive: owns the values, validates them against a Zod schema,
 * and only invokes `onSubmit` once the whole form is valid — so the handler
 * receives fully-typed, parsed data. Use it directly with a plain `<form>`, or
 * hand the result to `<Form form={…}>` to reuse the styled input components.
 */
export function createForm<Schema extends z.ZodType>(
  options: CreateFormOptions<Schema>,
): FormApi<Schema> {
  const seed = () => ({
    ...(options.defaultValues as Record<string, unknown> | undefined),
  });

  const [values, setValues] = createStore<Record<string, unknown>>(seed());
  const [errors, setErrors] = createSignal<Record<string, string>>({});
  const [touched, setTouched] = createSignal<Record<string, boolean>>({});
  const [submitting, setSubmitting] = createSignal(false);

  // Run the schema and fold Zod's issues into a `field -> message` map. Only
  // the first message per field is kept, which is what a user wants to see.
  const collectErrors = (): Record<string, string> => {
    const result = options.schema.safeParse(values);
    if (result.success) return {};
    const map: Record<string, string> = {};
    for (const issue of result.error.issues) {
      const key = issue.path.map(String).join(".") || FORM_LEVEL_KEY;
      if (!(key in map)) map[key] = issue.message;
    }
    return map;
  };

  const setValue: FormContextValue["setValue"] = (name, value) => {
    setValues(name, value);
    // Live-revalidate a field the user has already interacted with.
    if (touched()[name]) setErrors(collectErrors());
  };

  const markTouched: FormContextValue["setTouched"] = (name) => {
    setTouched((prev) => (prev[name] ? prev : { ...prev, [name]: true }));
    setErrors(collectErrors());
  };

  const reset: FormApi<Schema>["reset"] = (next) => {
    setValues(
      reconcile({
        ...((next ?? options.defaultValues) as
          | Record<string, unknown>
          | undefined),
      }),
    );
    setErrors({});
    setTouched({});
  };

  const handleSubmit = async (event?: Event) => {
    event?.preventDefault();
    const result = options.schema.safeParse(values);

    if (!result.success) {
      // Surface every error: touch each offending field so its message
      // shows, even ones the user never focused.
      const nextErrors: Record<string, string> = {};
      const nextTouched: Record<string, boolean> = { ...touched() };
      for (const issue of result.error.issues) {
        const key = issue.path.map(String).join(".") || FORM_LEVEL_KEY;
        if (!(key in nextErrors)) nextErrors[key] = issue.message;
        nextTouched[key] = true;
      }
      setErrors(nextErrors);
      setTouched(nextTouched);
      return;
    }

    setErrors({});
    try {
      setSubmitting(true);
      await options.onSubmit(result.data);
    } finally {
      setSubmitting(false);
    }
  };

  return {
    values: values as unknown as z.input<Schema>,
    value: (name) => values[name],
    // Errors stay hidden until the field is touched, so a pristine form
    // isn't a wall of red before anyone has typed.
    error: (name) => (touched()[name] ? errors()[name] : undefined),
    errors,
    touched: (name) => Boolean(touched()[name]),
    isValid: () => options.schema.safeParse(values).success,
    setValue,
    setTouched: markTouched,
    submitting,
    reset,
    handleSubmit,
  };
}
