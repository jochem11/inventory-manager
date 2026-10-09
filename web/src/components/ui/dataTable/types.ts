import type { JSX } from "solid-js";
import type { IconName } from "../icon";

export type CellValue = string | number | boolean | Date | null | undefined;

/** A dropdown filter choice. `value` is what's matched / sent to the server. */
export type FilterOption = string | { value: string; label: string };

export type DataTableColumn<T> = {
  /** Stable key for sorting, filtering, widths and visibility (and the server query). */
  id: string;
  header: string;
  /** The raw value: used for sorting, and (formatted) for search and display. */
  value: (row: T) => CellValue;
  /** Text used for search, text filters and the default cell. Defaults to a formatted `value`. */
  text?: (row: T) => string;
  /** Custom cell content; defaults to `text`. */
  cell?: (row: T) => JSX.Element;
  /** Default `true`. */
  sortable?: boolean;
  /**
   * Per-column filter shown in the filter row. `"text"` (default) matches
   * substrings; `{ options }` is an exact-match dropdown; `false` disables it.
   */
  filter?: "text" | { options: FilterOption[] } | false;
  /** Start hidden; users can show it from the Columns menu. */
  hidden?: boolean;
  /** Whether users may hide it (default `true`). */
  hideable?: boolean;
  /** Initial width in px (default 160). Users can drag to resize. */
  width?: number;
  /** Smallest width a resize can reach, in px (default 64). */
  minWidth?: number;
  align?: "start" | "center" | "end";
};

export type DataTableAction<T> =
  | {
      label: string;
      icon?: IconName;
      danger?: boolean;
      onSelect: (row: T) => void;
      /** Leave the action out for rows where it doesn't apply. */
      hidden?: (row: T) => boolean;
    }
  | { divider: true };

export type SortState = { id: string; direction: "asc" | "desc" };

/** Everything the table asks of the data in server mode; map it to API variables. */
export type DataTableQuery = {
  /** Zero-based. */
  page: number;
  pageSize: number;
  sort?: SortState;
  /** Trimmed, debounced global search. */
  search: string;
  /** Column id → filter value; only non-empty filters on visible columns. */
  filters: Record<string, string>;
};

/**
 * Server-side data for a table in one object: what `createServerTable`
 * returns. Pass it as `source` instead of rows/totalRows/loading/onQueryChange.
 */
export type DataTableSource<T> = {
  /** The current page. */
  rows: () => T[];
  /** Matching rows across all pages. */
  total: () => number;
  loading: () => boolean;
  /** The last load's error; the table shows it with a "Try again" button. */
  error: () => unknown;
  refetch: () => unknown;
  onQueryChange: (query: DataTableQuery) => void;
  /** Controlled search (e.g. from the URL); without it the table keeps its own. */
  search?: () => string;
  setSearch?: (search: string) => void;
};

export type DataTableProps<T> = {
  /**
   * Server-side data (see `createServerTable`). Puts the table in server mode
   * and replaces rows, totalRows, loading, onQueryChange, search and
   * onSearchChange.
   */
  source?: DataTableSource<T>;
  /** Client mode: all rows. Server mode without `source`: just the current page. */
  rows?: T[];
  columns: DataTableColumn<T>[];
  rowKey: (row: T) => string | number;
  /** Row name for accessible labels, e.g. "Actions for ThinkPad T14". */
  rowLabel?: (row: T) => string;

  /**
   * `"client"` (default): the table searches, filters, sorts and pages `rows`
   * itself. `"server"`: `rows` is already the requested page; the table reports
   * what to fetch via `onQueryChange` and shows `totalRows` in the pager.
   */
  mode?: "client" | "server";
  /** Server mode: total matching rows across all pages. */
  totalRows?: number;
  /** Server mode: fires on mount and whenever paging, sorting, search or filters change. */
  onQueryChange?: (query: DataTableQuery) => void;
  /** Shows a progress bar and dims the rows (e.g. while a request is in flight). */
  loading?: boolean;

  /** Per-row "⋯" menu. */
  actions?: DataTableAction<T>[];
  /** Shows an Add button in the toolbar. */
  onAdd?: () => void;
  addLabel?: string;
  /** Enables row checkboxes and a bulk Delete for the selection (kept across pages). */
  onDelete?: (rows: T[]) => void;

  /** Global search. Pass `search` + `onSearchChange` to control it (e.g. from the URL). */
  searchable?: boolean;
  search?: string;
  onSearchChange?: (query: string) => void;
  searchPlaceholder?: string;

  /** Remember column visibility and widths in localStorage under this key. */
  storageKey?: string;
  initialSort?: SortState;
  pageSize?: number;
  pageSizes?: number[];
  emptyMessage?: string;
  class?: string;
};
