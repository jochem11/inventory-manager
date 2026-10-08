import {
  createEffect,
  createUniqueId,
  For,
  onCleanup,
  onMount,
  ParentProps,
} from "solid-js";
import { createStore, produce } from "solid-js/store";
import { isServer, Portal } from "solid-js/web";
import { DialogContext } from "~/context";
import { DialogEntry } from "~/types/dialog";

/**
 * Wraps one stack entry in a native <dialog>, which gives us focus trapping,
 * top-layer stacking and background inertness for free. Escape and backdrop
 * clicks are routed through `requestClose` so state stays the single source of
 * truth (we never let the element close itself out from under the signal).
 */
const DialogShell = (props: {
  entry: DialogEntry;
  onClose: (id: string, result?: unknown) => void;
}) => {
  let el!: HTMLDialogElement;
  const requestClose = (result?: unknown) =>
    props.onClose(props.entry.id, result);

  onMount(() => {
    const previouslyFocused = document.activeElement as HTMLElement | null;
    el.showModal();
    onCleanup(() => previouslyFocused?.focus?.());
  });

  return (
    <dialog
      ref={el}
      onCancel={(e) => {
        e.preventDefault(); // Escape: close via state, not the element
        requestClose();
      }}
      onClick={(e) => {
        if (e.target === el) requestClose(); // click on ::backdrop
      }}>
      {props.entry.render(requestClose)}
    </dialog>
  );
};

export const DialogProvider = (props: ParentProps) => {
  // A store (rather than a signal of an array) gives the stack fine-grained
  // updates: mounting/unmounting one dialog doesn't recreate the others.
  const [dialogs, setDialogs] = createStore<DialogEntry[]>([]);

  const mount = (entry: Omit<DialogEntry, "id">): string => {
    const id = createUniqueId(); // SSR-safe unique id
    setDialogs(dialogs.length, { ...entry, id });
    return id;
  };

  const close = (id: string, result?: unknown) => {
    const entry = dialogs.find((d) => d.id === id);
    if (!entry) return; // idempotent: only resolve once
    const { onClose } = entry; // capture before the proxy is detached
    setDialogs(
      produce((list) => {
        const i = list.findIndex((d) => d.id === id);
        if (i >= 0) list.splice(i, 1);
      }),
    );
    onClose?.(result);
  };

  // Lock body scroll while any dialog is open.
  createEffect(() => {
    if (isServer) return;
    document.body.style.overflow = dialogs.length > 0 ? "hidden" : "";
  });

  return (
    <DialogContext.Provider value={{ mount, close }}>
      {props.children}
      <Portal>
        <For each={dialogs}>
          {(entry) => <DialogShell entry={entry} onClose={close} />}
        </For>
      </Portal>
    </DialogContext.Provider>
  );
};
