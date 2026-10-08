import { createSignal } from "solid-js";
import type { DataTableQuery } from "~/components/ui";

export type TablePage<T> = { rows: T[]; total: number };

/**
 * Connects a server-mode `DataTable` to an async fetcher (e.g. a GraphQL
 * query). Pass the result's `rows`/`total`/`loading`/`onQueryChange` to the
 * table, and run changes through `mutate()` so the table shows progress and
 * refreshes afterwards.
 *
 * Responses that arrive out of order are ignored, so a slow request for an old
 * page can't overwrite a newer one.
 */
export function createTableSource<T>(fetchPage: (query: DataTableQuery) => Promise<TablePage<T>>) {
  const [rows, setRows] = createSignal<T[]>([]);
  const [total, setTotal] = createSignal(0);
  // Starts true: the first request fires once the table mounts.
  const [loading, setLoading] = createSignal(true);
  const [error, setError] = createSignal<unknown>();

  let lastQuery: DataTableQuery | undefined;
  let latestRequest = 0;

  const load = async (query: DataTableQuery) => {
    lastQuery = query;
    const request = ++latestRequest;
    setLoading(true);
    setError(undefined);
    try {
      const page = await fetchPage(query);
      if (request !== latestRequest) return;
      setRows(page.rows);
      setTotal(page.total);
    } catch (e) {
      if (request === latestRequest) setError(e);
    } finally {
      if (request === latestRequest) setLoading(false);
    }
  };

  const refetch = () => (lastQuery ? load(lastQuery) : Promise.resolve());

  return {
    rows,
    total,
    loading,
    error,
    /** Pass to `<DataTable onQueryChange>`. */
    onQueryChange: load,
    /** Re-run the current query. */
    refetch,
    /**
     * Run a change (a create/update/delete mutation) with the table in its
     * loading state from the first moment, then refetch the current page.
     */
    mutate: async <R>(change: () => Promise<R>): Promise<R> => {
      setLoading(true);
      try {
        return await change();
      } finally {
        await refetch();
      }
    },
  };
}

/** What `createTableSource` returns, to pass a table's data around. */
export type TableSource<T> = ReturnType<typeof createTableSource<T>>;
