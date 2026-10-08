import { Show, type JSX } from "solid-js";
import {
  FieldShell,
  useField,
  type BaseFieldProps,
  type FieldSize,
  type FieldVariant,
} from "../field";

export type TextFieldProps = BaseFieldProps & {
  type?: "text" | "email" | "password" | "number" | "search" | "tel" | "url";
  placeholder?: string;
  autocomplete?: string;
  variant?: FieldVariant;
  size?: FieldSize;
  /** Rendered inside the input box, before the text (icon, "€", …). */
  prefix?: JSX.Element;
  /** Rendered inside the input box, after the text (unit, icon, …). */
  suffix?: JSX.Element;
  min?: number;
  max?: number;
  step?: number | "any";
};

/**
 * Single-line input bound to a `<Form>` field. `type="number"` stores a
 * `number` (or `undefined` when empty) so the schema can use `z.number()`.
 */
export function TextField(props: TextFieldProps) {
  const field = useField(props);

  const onInput = (el: HTMLInputElement) => {
    if (props.type === "number") {
      const n = el.valueAsNumber;
      field.setValue(Number.isNaN(n) ? undefined : n);
    } else {
      field.setValue(el.value);
    }
  };

  return (
    <FieldShell
      {...props}
      field={field}
      modifiers={`field--${props.variant ?? "outlined"} field--${props.size ?? "md"}`}>
      <div class="field__control">
        <Show when={props.prefix}>
          <span class="field__adornment">{props.prefix}</span>
        </Show>
        <input
          id={field.id}
          name={props.name}
          class="field__input"
          type={props.type ?? "text"}
          placeholder={props.placeholder}
          autocomplete={props.autocomplete}
          disabled={props.disabled}
          min={props.min}
          max={props.max}
          step={props.step}
          value={(field.value() as string | number | undefined) ?? ""}
          onInput={(e) => onInput(e.currentTarget)}
          onBlur={field.onBlur}
          {...field.aria()}
        />
        <Show when={props.suffix}>
          <span class="field__adornment">{props.suffix}</span>
        </Show>
      </div>
    </FieldShell>
  );
}
