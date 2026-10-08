import "./textAreaField.scss";
import {
  FieldShell,
  useField,
  type BaseFieldProps,
  type FieldSize,
  type FieldVariant,
} from "../field";

export type TextAreaFieldProps = BaseFieldProps & {
  placeholder?: string;
  rows?: number;
  variant?: FieldVariant;
  size?: FieldSize;
};

/** Multi-line text input bound to a `<Form>` field. */
export function TextAreaField(props: TextAreaFieldProps) {
  const field = useField(props);

  return (
    <FieldShell
      {...props}
      field={field}
      modifiers={`field--${props.variant ?? "outlined"} field--${props.size ?? "md"}`}>
      <div class="field__control">
        <textarea
          id={field.id}
          name={props.name}
          class="field__input field__input--textarea"
          placeholder={props.placeholder}
          rows={props.rows ?? 4}
          disabled={props.disabled}
          value={(field.value() as string | undefined) ?? ""}
          onInput={(e) => field.setValue(e.currentTarget.value)}
          onBlur={field.onBlur}
          {...field.aria()}
        />
      </div>
    </FieldShell>
  );
}
