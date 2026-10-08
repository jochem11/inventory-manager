import "./selectField.scss";
import { For } from "solid-js";
import {
  FieldShell,
  useField,
  type BaseFieldProps,
  type FieldSize,
  type FieldVariant,
} from "../field";

export type SelectOption = {
  value: string;
  label: string;
  disabled?: boolean;
};

export type SelectFieldProps = BaseFieldProps & {
  options: SelectOption[];
  /** Shown as an unselectable first option while no value is picked. */
  placeholder?: string;
  variant?: FieldVariant;
  size?: FieldSize;
};

/**
 * Native `<select>` bound to a `<Form>` field. Picking the placeholder stores
 * `undefined`, so a required schema field reports it as missing.
 */
export function SelectField(props: SelectFieldProps) {
  const field = useField(props);
  // Selection is driven per-option rather than via `<select value>`: the
  // select's value is applied before its options exist, so the browser would
  // fall back to the first enabled option while the form still holds nothing.
  const current = () => (field.value() as string | undefined) ?? "";

  return (
    <FieldShell
      {...props}
      field={field}
      modifiers={`field--${props.variant ?? "outlined"} field--${props.size ?? "md"}`}>
      <div class="field__control field__control--select">
        <select
          id={field.id}
          name={props.name}
          class="field__input field__input--select"
          disabled={props.disabled}
          onChange={(e) => field.setValue(e.currentTarget.value || undefined)}
          onBlur={field.onBlur}
          {...field.aria()}>
          {/* Always present so an empty value has something to select; hidden
              from the dropdown when there's no placeholder text. */}
          <option
            value=""
            disabled
            hidden={!props.placeholder}
            selected={current() === ""}>
            {props.placeholder ?? ""}
          </option>
          <For each={props.options}>
            {(option) => (
              <option
                value={option.value}
                disabled={option.disabled}
                selected={current() === option.value}>
                {option.label}
              </option>
            )}
          </For>
        </select>
      </div>
    </FieldShell>
  );
}
