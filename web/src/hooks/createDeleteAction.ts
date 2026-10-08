import { createSignal } from "solid-js";
import { errorMessage } from "~/api/graphql";
import { ConfirmDialog } from "~/components/dialogs";
import { createDialog } from "./createDialog";

export type DeleteActionOptions<T> = {
  /** The table to refresh afterwards (a `createServerTable` or `createTableSource`). */
  source: { mutate: <R>(change: () => Promise<R>) => Promise<R> };
  /** E.g. `["user", "users"]`, for "Delete 3 users?". */
  noun: [singular: string, plural: string];
  /** The row's name, for "Delete Ada Lovelace?". */
  name: (row: T) => string;
  /** Deletes one row; used per row unless `removeMany` is given. */
  remove?: (row: T) => Promise<unknown>;
  /** Deletes all rows in one call, when the API can. */
  removeMany?: (rows: T[]) => Promise<unknown>;
  /** Why a row can't be deleted (it's then left out), or undefined when it can. */
  skip?: (row: T) => string | undefined;
  /** The confirmation's text, e.g. "This can't be undone." */
  message?: string;
};

/**
 * Deleting rows after a confirmation. `run(rows)` fits `<DataTable onDelete>`
 * and a row action's `onSelect`; `error()` holds what went wrong, for a notice.
 */
export function createDeleteAction<T>(options: DeleteActionOptions<T>) {
  const confirm = createDialog(ConfirmDialog);
  const [error, setError] = createSignal<string>();
  const [singular, plural] = options.noun;

  const run = async (selected: T[]) => {
    const reasons = selected.map((row) => options.skip?.(row));
    const toRemove = selected.filter((_, i) => !reasons[i]);
    const skipped = reasons.find(Boolean);
    if (toRemove.length === 0) {
      setError(skipped ?? `Nothing to delete`);
      return;
    }

    const ok = await confirm.open({
      title:
        toRemove.length === 1 ? `Delete ${options.name(toRemove[0]!)}?` : `Delete ${toRemove.length} ${plural}?`,
      message: [skipped && `${skipped}.`, options.message].filter(Boolean).join(" ") || undefined,
      confirmLabel: "Delete",
      variant: "danger",
    });
    if (!ok) return;

    setError();
    try {
      await options.source.mutate(() =>
        options.removeMany
          ? options.removeMany(toRemove)
          : Promise.all(toRemove.map((row) => options.remove!(row))),
      );
    } catch (e) {
      setError(`Couldn't delete the ${toRemove.length === 1 ? singular : plural}: ${errorMessage(e)}`);
    }
  };

  return { run, error, clearError: () => setError() };
}
