import { JSX, Component } from "solid-js";

declare const DIALOG_BRAND: unique symbol;

export type DialogProps<R = void> = {
  /** Close the dialog, optionally resolving `open()` with a result. */
  close: (result?: R) => void;
};

export type DialogInstance<P extends Record<string, unknown>, R = void> = {
  readonly [DIALOG_BRAND]: true;
  readonly component: Component<P & DialogProps<R>>;
};

export type InferProps<D> =
  D extends DialogInstance<infer P, unknown> ? P : never;
export type InferResult<D> =
  D extends DialogInstance<Record<string, unknown>, infer R> ? R : never;

/**
 * A live dialog in the provider's stack. The entry stores a render thunk rather
 * than a component + props pair, so the JSX is built where the caller's types
 * are known — no casts are needed to bridge the branded instance to the stack.
 */
export type DialogEntry = {
  id: string;
  render: (close: (result?: unknown) => void) => JSX.Element;
  onClose?: (result: unknown) => void;
};
