import "./field.scss";
import { createUniqueId, Show, type JSX } from "solid-js";
import { useForm } from "~/context/formContext";

/** Visual treatment shared by every text-like input. */
export type FieldVariant = "outlined" | "filled" | "underlined";
export type FieldSize = "sm" | "md" | "lg";

export type BaseFieldProps = {
  /** Key into the form values (and the schema). */
  name: string;
  label?: string;
  /** Helper text under the input; replaced by the error when there is one. */
  hint?: string;
  class?: string;
  disabled?: boolean;
};

/**
 * Wires a field up to the surrounding `<Form>`: stable ids for labelling, the
 * live value/error, and the aria attributes the input itself should spread.
 */
export function useField(props: BaseFieldProps) {
  const form = useForm();
  const id = createUniqueId();
  const hintId = `${id}-hint`;

  const error = () => form.error(props.name);

  return {
    form,
    id,
    hintId,
    error,
    value: () => form.value(props.name),
    setValue: (value: unknown) => form.setValue(props.name, value),
    onBlur: () => form.setTouched(props.name),
    aria: () => ({
      "aria-invalid": error() ? true : undefined,
      "aria-describedby": error() || props.hint ? hintId : undefined,
    }),
  };
}

type FieldShellProps = BaseFieldProps & {
  field: ReturnType<typeof useField>;
  /** Extra modifier classes, e.g. `field--filled field--lg`. */
  modifiers?: string;
  /** Checkboxes put the label beside the control, so they render their own. */
  hideLabel?: boolean;
  children: JSX.Element;
};

/** Label + control + hint/error layout every field component renders into. */
export function FieldShell(props: FieldShellProps) {
  return (
    <div
      class={`field ${props.modifiers ?? ""} ${props.class ?? ""}`}
      classList={{
        "field--invalid": !!props.field.error(),
        "field--disabled": !!props.disabled,
      }}>
      <Show when={props.label && !props.hideLabel}>
        <label for={props.field.id} class="field__label">
          {props.label}
        </label>
      </Show>

      {props.children}

      <Show when={props.field.error() ?? props.hint}>
        {(message) => (
          <p
            id={props.field.hintId}
            class="field__hint"
            role={props.field.error() ? "alert" : undefined}>
            {message()}
          </p>
        )}
      </Show>
    </div>
  );
}
