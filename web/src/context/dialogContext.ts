import { createContext, useContext } from "solid-js";
import { ContextError } from "~/errors/contextError";
import { DialogEntry } from "~/types/dialog";

export type DialogContextType = {
  mount(entry: Omit<DialogEntry, "id">): string;
  /** Close a dialog by id, resolving its `open()` promise with `result`. */
  close(id: string, result?: unknown): void;
};

export const DialogContext = createContext<DialogContextType>();

export const useDialogContext = () => {
  const ctx = useContext(DialogContext);
  if (!ctx)
    throw new ContextError("useDialogContext must be used within a DialogProvider");
  return ctx;
};
