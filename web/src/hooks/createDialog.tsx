import { Dynamic } from "solid-js/web";
import { useDialogContext } from "~/context";
import { DialogInstance } from "~/types/dialog";

type OpenFn<P, R> = keyof P extends never
  ? () => Promise<R | undefined>
  : (props: P) => Promise<R | undefined>;

export const createDialog = <P extends Record<string, unknown>, R>(
  instance: DialogInstance<P, R>,
): {
  /** Open the dialog; resolves when it closes (with the result, if any). */
  open: OpenFn<P, R>;
  /** Programmatically close the current instance, resolving with `undefined`. */
  close: () => void;
} => {
  const ctx = useDialogContext();
  let currentId: string | null = null;

  const open = ((props?: P) => {
    if (currentId !== null) ctx.close(currentId);

    return new Promise<R | undefined>((resolve) => {
      currentId = ctx.mount({
        // Build the JSX here, where P and R are known — no casts needed
        // to reconcile the branded instance with the untyped stack.
        // <Dynamic> is Solid's primitive for a component held in a value.
        render: (close) => (
          <Dynamic
            component={instance.component}
            {...(props ?? ({} as P))}
            close={close as (result?: R) => void}
          />
        ),
        onClose: (result) => {
          currentId = null;
          resolve(result as R | undefined);
        },
      });
    });
  }) as OpenFn<P, R>;

  const close = () => {
    if (currentId !== null) ctx.close(currentId);
  };

  return { open, close };
};
