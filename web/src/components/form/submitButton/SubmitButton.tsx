import { splitProps, type JSX } from "solid-js";
import { useForm } from "~/context/formContext";
import { Button, type ButtonProps } from "~/components/ui";

export type SubmitButtonProps = Omit<ButtonProps, "type" | "loading"> & {
  /** Label swapped in while `onSubmit` is running (defaults to `children`). */
  submittingText?: JSX.Element;
};

/**
 * `Button` that submits the surrounding `<Form>`. It shows a spinner and
 * disables itself while the async `onSubmit` is in flight, so a double click
 * can't submit twice.
 */
export function SubmitButton(props: SubmitButtonProps) {
  const form = useForm();
  const [local, rest] = splitProps(props, ["submittingText", "children"]);

  return (
    <Button variant="primary" {...rest} type="submit" loading={form.submitting()}>
      {form.submitting() && local.submittingText ? local.submittingText : local.children}
    </Button>
  );
}
