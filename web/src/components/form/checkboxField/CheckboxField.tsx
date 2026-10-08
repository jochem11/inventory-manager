import "./checkboxField.scss";
import { FieldShell, useField, type BaseFieldProps } from "../field";

export type CheckboxFieldProps = BaseFieldProps & {
  /** `checkbox` is a tick box; `switch` is an on/off toggle. */
  variant?: "checkbox" | "switch";
};

/** Boolean field bound to a `<Form>` field; stores `true` / `false`. */
export function CheckboxField(props: CheckboxFieldProps) {
  const field = useField(props);
  const variant = () => props.variant ?? "checkbox";

  return (
    <FieldShell
      {...props}
      field={field}
      modifiers={`field--check field--${variant()}`}
      hideLabel>
      <label class="field__check">
        <input
          id={field.id}
          name={props.name}
          type="checkbox"
          role={variant() === "switch" ? "switch" : undefined}
          class="field__check-input"
          disabled={props.disabled}
          checked={Boolean(field.value())}
          onChange={(e) => {
            field.setValue(e.currentTarget.checked);
            // Checkboxes have no "typing" phase, so validate on the first click.
            field.onBlur();
          }}
          {...field.aria()}
        />
        <span class="field__check-box" aria-hidden="true" />
        <span class="field__check-label">{props.label}</span>
      </label>
    </FieldShell>
  );
}
