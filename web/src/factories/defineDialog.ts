import { Component } from "solid-js";
import { DialogInstance, DialogProps } from "~/types/dialog";

export const defineDialog = <P extends Record<string, unknown>, R = void>(
  component: Component<P & DialogProps<R>>,
): DialogInstance<P, R> => {
  // Single controlled cast to attach the phantom brand; unavoidable and local.
  return { component } as unknown as DialogInstance<P, R>;
};
 