import type { DataTableQuery } from "~/components/ui";

// ---------------------------------------------------------------------------
// The gateway's list convention, shared by every list query:
//
//   things(offset: Int!, limit: Int!, orderBy: ThingOrder, filter: ThingFilter): ThingConnection!
//   type ThingConnection { totalCount: Int!  nodes: [Thing!]! }
// ---------------------------------------------------------------------------

/** One page of a list, plus how many items match across all pages. */
export type Connection<T> = { totalCount: number; nodes: T[] };

export type ListVariables<SortField extends string = string> = {
  offset: number;
  limit: number;
  orderBy?: { field: SortField; direction: "ASC" | "DESC" };
  filter: Record<string, unknown>;
};

export type ListOptions<SortField extends string> = {
  /** DataTable column id → the API's sort field. Columns not listed don't sort on the server. */
  sortFields: Record<string, SortField>;
  /**
   * The column filters (column id → value) as the API's filter fields. By
   * default they're passed on as they are, so name columns after the fields.
   */
  filters?: (filters: Record<string, string>) => Record<string, unknown>;
};

/** Translates a DataTable's state into a list query's variables. */
export function toListVariables<SortField extends string>(
  query: DataTableQuery,
  options: ListOptions<SortField>,
): ListVariables<SortField> {
  const field = query.sort && options.sortFields[query.sort.id];
  return {
    offset: query.page * query.pageSize,
    limit: query.pageSize,
    orderBy: field ? { field, direction: query.sort!.direction === "asc" ? "ASC" : "DESC" } : undefined,
    filter: {
      search: query.search || undefined,
      ...(options.filters ? options.filters(query.filters) : query.filters),
    },
  };
}
