import { useSearchParams } from "@solidjs/router";
import { createSignal } from "solid-js";
import { toListVariables, type Connection, type ListOptions, type ListVariables } from "~/api/list";
import { createTableSource } from "./createTableSource";

export type ServerTableOptions<T, SortField extends string> = ListOptions<SortField> & {
  /** Loads one page, e.g. `(variables) => api.fetchUsers(auth.request, variables)`. */
  fetch: (variables: ListVariables<SortField>) => Promise<Connection<T>>;
  /**
   * The URL parameter that holds the search, so it survives a reload and the
   * topbar search box can fill it. Default `"q"`; `false` keeps it local.
   */
  searchParam?: string | false;
};

/**
 * Everything a server-side `DataTable` needs, for a list that follows the
 * gateway's convention (see api/list.ts). Pass it as `<DataTable source={…}>`.
 *
 *   const users = createServerTable({
 *     fetch: (variables) => api.fetchUsers(auth.request, variables),
 *     sortFields: USER_SORT_FIELDS,
 *   });
 *   <DataTable source={users} columns={USER_COLUMNS} rowKey={(u) => u.id} />
 *
 * Run changes through `mutate()` so the table shows progress and refreshes.
 */
export function createServerTable<T, SortField extends string>(options: ServerTableOptions<T, SortField>) {
  const source = createTableSource<T>(async (query) => {
    const page = await options.fetch(toListVariables(query, options));
    return { rows: page.nodes, total: page.totalCount };
  });

  const param = options.searchParam ?? "q";
  const [params, setParams] = useSearchParams();
  const [localSearch, setLocalSearch] = createSignal("");

  const search = () => {
    if (!param) return localSearch();
    const value = params[param];
    return typeof value === "string" ? value : "";
  };
  const setSearch = (value: string) =>
    param ? setParams({ [param]: value || undefined }, { replace: true }) : setLocalSearch(value);

  return {
    ...source,
    search,
    setSearch,
    /** For a page header: "Loading…", "3 users", "1 user matches your search". */
    summary: (singular: string, plural: string) => {
      if (source.loading() && source.total() === 0) return "Loading…";
      const count = `${source.total()} ${source.total() === 1 ? singular : plural}`;
      if (!search()) return count;
      return `${count} ${source.total() === 1 ? "matches" : "match"} your search`;
    },
  };
}

export type ServerTable<T> = ReturnType<typeof createServerTable<T, string>>;
