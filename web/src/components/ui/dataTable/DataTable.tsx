import "./dataTable.scss";
import {
  createEffect,
  createMemo,
  createSignal,
  For,
  Index,
  on,
  onCleanup,
  onMount,
  Show,
  untrack,
} from "solid-js";
import { createStore, reconcile } from "solid-js/store";
import { Button } from "../button";
import { DropdownMenu, type DropdownMenuItem } from "../dropdownMenu";
import { Icon } from "../icon";
import type {
  CellValue,
  DataTableColumn,
  DataTableProps,
  DataTableQuery,
  FilterOption,
  SortState,
} from "./types";

type Key = string | number;

const DEFAULT_WIDTH = 160;
const DEFAULT_MIN_WIDTH = 64;
const SELECT_COL_WIDTH = 44;
const ACTIONS_COL_WIDTH = 56;
const SERVER_SEARCH_DEBOUNCE_MS = 300;

const formatValue = (value: CellValue): string => {
  if (value === null || value === undefined) return "";
  if (value instanceof Date) return value.toLocaleDateString();
  if (typeof value === "boolean") return value ? "Yes" : "No";
  return String(value);
};

const errorText = (error: unknown) => (error instanceof Error ? error.message : "something went wrong");

const isEmpty = (value: CellValue) => value === null || value === undefined || value === "";

// Numbers numerically, dates by time, text naturally ("Item 2" < "Item 10").
const compareValues = (a: CellValue, b: CellValue): number => {
  if (a instanceof Date && b instanceof Date) return a.getTime() - b.getTime();
  if (typeof a === "number" && typeof b === "number") return a - b;
  if (typeof a === "boolean" && typeof b === "boolean") return Number(a) - Number(b);
  return formatValue(a).localeCompare(formatValue(b), undefined, { numeric: true, sensitivity: "base" });
};

const optionValue = (o: FilterOption) => (typeof o === "string" ? o : o.value);
const optionLabel = (o: FilterOption) => (typeof o === "string" ? o : o.label);

type Saved = { hidden?: string[]; widths?: Record<string, number> };

/**
 * A table with search, per-column filters, sorting, paging, resizable and
 * hideable columns, row selection with bulk delete, and a per-row action menu.
 *
 * In client mode it processes `rows` itself. In server mode (for API-backed
 * data) it only renders the page it's given and reports what to fetch through
 * `onQueryChange`. Either way it never changes `rows`; adding and removing are
 * reported through `onAdd` / `onDelete`.
 */
export function DataTable<T>(props: DataTableProps<T>) {
  const serverMode = () => !!props.source || props.mode === "server";
  // Data, from `source` when given, otherwise from the individual props.
  const rows = () => (props.source ? props.source.rows() : (props.rows ?? []));
  const totalRows = () => (props.source ? props.source.total() : props.totalRows);
  const loading = () => (props.source ? props.source.loading() : !!props.loading);
  const loadError = () => props.source?.error();
  const onQueryChange = (query: DataTableQuery) =>
    (props.source?.onQueryChange ?? props.onQueryChange)?.(query);

  // --- State ---------------------------------------------------------------
  const [sort, setSort] = createSignal<SortState | undefined>(untrack(() => props.initialSort));
  const [localSearch, setLocalSearch] = createSignal("");
  const [filters, setFilters] = createStore<Record<string, string>>({});
  const [showFilters, setShowFilters] = createSignal(false);
  const [page, setPage] = createSignal(0);
  const [pageSize, setPageSize] = createSignal(untrack(() => props.pageSize ?? 10));
  // Selected rows by key. Rows (not just keys) are kept so a selection can
  // span pages in server mode, where other pages' rows aren't in `rows`.
  const [selected, setSelected] = createSignal<ReadonlyMap<Key, T>>(new Map());
  const [widths, setWidths] = createStore<Record<string, number>>({});
  const [hidden, setHidden] = createSignal<ReadonlySet<string>>(
    new Set(untrack(() => props.columns.filter((c) => c.hidden).map((c) => c.id))),
  );
  const [compact, setCompact] = createSignal(false);
  const [resizing, setResizing] = createSignal<string>();

  const search = () => props.source?.search?.() ?? props.search ?? localSearch();
  const setSearch = (query: string) =>
    (props.source?.setSearch ?? props.onSearchChange ?? setLocalSearch)(query);

  const selectable = () => !!props.onDelete;
  const hasActions = () => (props.actions?.length ?? 0) > 0;
  const pageSizes = () => props.pageSizes ?? [10, 25, 50];

  const columns = createMemo(() => props.columns.filter((c) => !hidden().has(c.id)));
  const columnById = (id: string) => columns().find((c) => c.id === id);
  const textOf = (col: DataTableColumn<T>, row: T) =>
    col.text ? col.text(row) : formatValue(col.value(row));
  const filterOf = (col: DataTableColumn<T>) => col.filter ?? "text";
  const widthOf = (col: DataTableColumn<T>) => widths[col.id] ?? col.width ?? DEFAULT_WIDTH;

  // Filters on hidden columns are kept (they come back with the column) but
  // don't apply while it's hidden.
  const activeFilters = () =>
    Object.fromEntries(Object.entries(filters).filter(([id, v]) => v && columnById(id)));
  const activeFilterCount = () => Object.keys(activeFilters()).length;
  const hasFilterableColumns = () => columns().some((c) => filterOf(c) !== false);
  const clearColumnFilters = () =>
    setFilters(reconcile(Object.fromEntries(Object.keys(filters).map((k) => [k, ""]))));

  // --- Remember visibility + widths ----------------------------------------
  // Read after mount so the server-rendered HTML (which can't see
  // localStorage) and the first client render agree.
  onMount(() => {
    if (!props.storageKey) return;
    try {
      const saved = JSON.parse(localStorage.getItem(props.storageKey) ?? "{}") as Saved;
      if (saved.hidden) setHidden(new Set(saved.hidden));
      if (saved.widths) setWidths(saved.widths);
    } catch {
      // Unavailable or corrupt storage — keep the defaults.
    }
  });
  createEffect(
    on(
      [hidden, () => JSON.stringify(widths)],
      () => {
        if (!props.storageKey) return;
        const saved: Saved = { hidden: [...hidden()], widths: { ...widths } };
        try {
          localStorage.setItem(props.storageKey, JSON.stringify(saved));
        } catch {
          // Not persisted; fine.
        }
      },
      { defer: true },
    ),
  );

  // --- Client-side pipeline: search → filters → sort → page ----------------
  const filtered = createMemo(() => {
    if (serverMode()) return rows();
    const query = search().trim().toLowerCase();
    const active = Object.entries(activeFilters());
    return rows().filter((row) => {
      if (query && !columns().some((c) => textOf(c, row).toLowerCase().includes(query))) {
        return false;
      }
      return active.every(([id, wanted]) => {
        const col = columnById(id)!;
        return typeof filterOf(col) === "object"
          ? String(col.value(row)) === wanted || textOf(col, row) === wanted
          : textOf(col, row).toLowerCase().includes(wanted.toLowerCase());
      });
    });
  });

  const sorted = createMemo(() => {
    const state = sort();
    const col = state && columnById(state.id);
    if (serverMode() || !state || !col) return filtered();
    const direction = state.direction === "asc" ? 1 : -1;
    return [...filtered()].sort((a, b) => {
      const va = col.value(a);
      const vb = col.value(b);
      // Empty cells always sink to the bottom, whichever the direction.
      if (isEmpty(va) || isEmpty(vb)) return Number(isEmpty(va)) - Number(isEmpty(vb));
      return compareValues(va, vb) * direction;
    });
  });

  const total = () => (serverMode() ? (totalRows() ?? rows().length) : sorted().length);
  const pageCount = () => Math.max(1, Math.ceil(total() / pageSize()));
  const pageRows = createMemo(() =>
    serverMode() ? rows() : sorted().slice(page() * pageSize(), (page() + 1) * pageSize()),
  );
  const rangeStart = () => (total() === 0 ? 0 : page() * pageSize() + 1);
  const rangeEnd = () => Math.min(total(), page() * pageSize() + pageRows().length);

  // --- Server mode: report the query ----------------------------------------
  // Search is debounced so typing doesn't fire a request per keystroke.
  const [debouncedSearch, setDebouncedSearch] = createSignal(untrack(search).trim());
  createEffect(
    on(
      search,
      (value) => {
        if (!serverMode()) return;
        const timer = setTimeout(() => setDebouncedSearch(value.trim()), SERVER_SEARCH_DEBOUNCE_MS);
        onCleanup(() => clearTimeout(timer));
      },
      { defer: true },
    ),
  );

  const query = createMemo(
    (): DataTableQuery => ({
      page: page(),
      pageSize: pageSize(),
      sort: sort(),
      search: debouncedSearch(),
      filters: activeFilters(),
    }),
    undefined,
    // Only a real change should trigger a new request.
    { equals: (a, b) => JSON.stringify(a) === JSON.stringify(b) },
  );
  createEffect(() => {
    if (serverMode()) onQueryChange(query());
  });

  // --- Keeping page/selection valid -----------------------------------------
  // New search/filters/sort/page size: start from the first page again.
  createEffect(
    on(
      [serverMode() ? debouncedSearch : search, () => JSON.stringify(activeFilters()), pageSize, sort],
      () => setPage(0),
      { defer: true },
    ),
  );
  // Rows removed from under the last page: step back to one that exists.
  createEffect(() => {
    if (page() > pageCount() - 1) setPage(pageCount() - 1);
  });
  // Client mode: forget selections for rows that no longer exist.
  createEffect(() => {
    if (serverMode()) return;
    const keys = new Set(rows().map(props.rowKey));
    const current = untrack(selected);
    if ([...current.keys()].some((k) => !keys.has(k))) {
      setSelected(new Map([...current].filter(([k]) => keys.has(k))));
    }
  });

  // --- Sorting -------------------------------------------------------------
  // Click cycles: ascending → descending → off.
  const toggleSort = (col: DataTableColumn<T>) => {
    const current = sort();
    if (current?.id !== col.id) setSort({ id: col.id, direction: "asc" });
    else if (current.direction === "asc") setSort({ id: col.id, direction: "desc" });
    else setSort(undefined);
  };
  const sortDirection = (col: DataTableColumn<T>) =>
    sort()?.id === col.id ? sort()!.direction : undefined;

  // --- Selection -----------------------------------------------------------
  const isSelected = (row: T) => selected().has(props.rowKey(row));
  const pageSelectedCount = () => pageRows().filter(isSelected).length;
  const allOnPageSelected = () => pageRows().length > 0 && pageSelectedCount() === pageRows().length;

  const toggleRow = (row: T) => {
    const key = props.rowKey(row);
    const next = new Map(selected());
    if (!next.delete(key)) next.set(key, row);
    setSelected(next);
  };
  const togglePage = () => {
    const next = new Map(selected());
    const all = allOnPageSelected();
    for (const row of pageRows()) {
      if (all) next.delete(props.rowKey(row));
      else next.set(props.rowKey(row), row);
    }
    setSelected(next);
  };
  const deleteSelected = () => {
    props.onDelete?.([...selected().values()]);
    setSelected(new Map());
  };

  // --- Column visibility ----------------------------------------------------
  const toggleColumn = (col: DataTableColumn<T>) => {
    const next = new Set(hidden());
    if (next.has(col.id)) next.delete(col.id);
    // Never hide the last visible column.
    else if (columns().length > 1) next.add(col.id);
    setHidden(next);
    if (sort()?.id === col.id && next.has(col.id)) setSort(undefined);
  };

  // Stable item objects (only rebuilt if the column list changes); `checked`
  // is an accessor so ticking one doesn't re-create — and unfocus — the list.
  const columnMenuItems = createMemo<DropdownMenuItem[]>(() =>
    props.columns
      .filter((c) => c.hideable !== false)
      .map((col) => ({
        label: col.header,
        checked: () => !hidden().has(col.id),
        onSelect: () => toggleColumn(col),
      })),
  );

  // --- Column resizing -----------------------------------------------------
  const resizeTo = (col: DataTableColumn<T>, width: number) =>
    setWidths(col.id, Math.max(col.minWidth ?? DEFAULT_MIN_WIDTH, Math.round(width)));

  const startResize = (col: DataTableColumn<T>, e: PointerEvent) => {
    e.preventDefault();
    const handle = e.currentTarget as HTMLElement;
    const startX = e.clientX;
    const startWidth = widthOf(col);
    handle.setPointerCapture(e.pointerId);
    setResizing(col.id);

    const move = (ev: PointerEvent) => resizeTo(col, startWidth + ev.clientX - startX);
    const end = () => {
      handle.removeEventListener("pointermove", move);
      handle.removeEventListener("pointerup", end);
      handle.removeEventListener("pointercancel", end);
      setResizing(undefined);
    };
    handle.addEventListener("pointermove", move);
    handle.addEventListener("pointerup", end);
    handle.addEventListener("pointercancel", end);
  };

  const resizeWithKeys = (col: DataTableColumn<T>, e: KeyboardEvent) => {
    const step = e.shiftKey ? 40 : 10;
    if (e.key === "ArrowRight") resizeTo(col, widthOf(col) + step);
    else if (e.key === "ArrowLeft") resizeTo(col, widthOf(col) - step);
    else return;
    e.preventDefault();
  };

  const resetWidth = (col: DataTableColumn<T>) => setWidths(col.id, undefined!);

  // The table is at least as wide as its columns; a filler column soaks up
  // any extra space so resizing one column never stretches the others.
  const minTableWidth = () =>
    columns().reduce((sum, c) => sum + widthOf(c), 0) +
    (selectable() ? SELECT_COL_WIDTH : 0) +
    (hasActions() ? ACTIONS_COL_WIDTH : 0);
  const columnCount = () => columns().length + 1 + (selectable() ? 1 : 0) + (hasActions() ? 1 : 0);

  const menuItems = (row: T): DropdownMenuItem[] =>
    (props.actions ?? [])
      .filter((a) => "divider" in a || !a.hidden?.(row))
      .map((a) =>
        "divider" in a
          ? a
          : { label: a.label, icon: a.icon, danger: a.danger, onSelect: () => a.onSelect(row) },
      );

  const clearAll = () => {
    clearColumnFilters();
    setSearch("");
  };

  return (
    <div
      class={`data-table ${props.class ?? ""}`}
      classList={{
        "data-table--compact": compact(),
        "data-table--resizing": !!resizing(),
        "data-table--loading": loading(),
      }}
      aria-busy={loading() || undefined}>
      {/* --- Toolbar ------------------------------------------------------ */}
      <div class="data-table__toolbar">
        <Show when={props.searchable !== false}>
          <label class="data-table__search">
            <Icon name="search" />
            <input
              type="search"
              name="search"
              placeholder={props.searchPlaceholder ?? "Search…"}
              aria-label={props.searchPlaceholder ?? "Search table"}
              value={search()}
              onInput={(e) => setSearch(e.currentTarget.value)}
            />
          </label>
        </Show>

        <Show when={hasFilterableColumns()}>
          <Button
            size="sm"
            variant={showFilters() ? "secondary" : "ghost"}
            startIcon={<Icon name="filter" />}
            aria-pressed={showFilters()}
            onClick={() => setShowFilters((v) => !v)}>
            Filters
            <Show when={activeFilterCount() > 0}>
              <span class="data-table__count">{activeFilterCount()}</span>
            </Show>
          </Button>
        </Show>

        <Show when={columnMenuItems().length > 0}>
          <DropdownMenu
            variant="button"
            align="start"
            label="Choose columns"
            trigger={
              <>
                <Icon name="columns" />
                Columns
              </>
            }
            items={columnMenuItems()}
          />
        </Show>

        <Button
          size="sm"
          variant={compact() ? "secondary" : "ghost"}
          startIcon={<Icon name="rows" />}
          aria-pressed={compact()}
          onClick={() => setCompact((v) => !v)}>
          Compact
        </Button>

        <div class="data-table__toolbar-end">
          <Show when={selected().size > 0}>
            <span class="data-table__selection" aria-live="polite">
              {selected().size} selected
            </span>
            <Button size="sm" variant="ghost" onClick={() => setSelected(new Map())}>
              Clear
            </Button>
            <Button size="sm" variant="danger" startIcon={<Icon name="trash" />} onClick={deleteSelected}>
              Delete
            </Button>
          </Show>
          <Show when={props.onAdd}>
            <Button
              size="sm"
              variant="primary"
              startIcon={<Icon name="plus" />}
              onClick={() => props.onAdd?.()}>
              {props.addLabel ?? "Add"}
            </Button>
          </Show>
        </div>
      </div>

      {/* --- Table -------------------------------------------------------- */}
      <div class="data-table__scroll">
        <div class="data-table__progress" aria-hidden="true" />
        <table class="data-table__table" style={{ "min-width": `${minTableWidth()}px` }}>
          <colgroup>
            <Show when={selectable()}>
              <col style={{ width: `${SELECT_COL_WIDTH}px` }} />
            </Show>
            <For each={columns()}>{(col) => <col style={{ width: `${widthOf(col)}px` }} />}</For>
            <col />
            <Show when={hasActions()}>
              <col style={{ width: `${ACTIONS_COL_WIDTH}px` }} />
            </Show>
          </colgroup>

          <thead>
            <tr>
              <Show when={selectable()}>
                <th class="data-table__select">
                  <input
                    type="checkbox"
                    name="select-page"
                    aria-label="Select all rows on this page"
                    checked={allOnPageSelected()}
                    ref={(el) =>
                      createEffect(() => {
                        el.indeterminate = pageSelectedCount() > 0 && !allOnPageSelected();
                      })
                    }
                    onChange={togglePage}
                  />
                </th>
              </Show>

              <For each={columns()}>
                {(col) => (
                  <th
                    scope="col"
                    class={`data-table__th data-table__cell--${col.align ?? "start"}`}
                    aria-sort={
                      sortDirection(col) === "asc"
                        ? "ascending"
                        : sortDirection(col) === "desc"
                          ? "descending"
                          : undefined
                    }>
                    <Show
                      when={col.sortable !== false}
                      fallback={<span class="data-table__th-label">{col.header}</span>}>
                      <button
                        type="button"
                        class="data-table__sort"
                        classList={{ "data-table__sort--active": !!sortDirection(col) }}
                        onClick={() => toggleSort(col)}>
                        <span class="data-table__th-label">{col.header}</span>
                        <Icon
                          class="data-table__sort-icon"
                          name={
                            sortDirection(col) === "asc"
                              ? "arrow-up"
                              : sortDirection(col) === "desc"
                                ? "arrow-down"
                                : "chevrons-up-down"
                          }
                        />
                      </button>
                    </Show>
                    <span
                      class="data-table__resizer"
                      classList={{ "data-table__resizer--active": resizing() === col.id }}
                      role="separator"
                      aria-orientation="vertical"
                      aria-label={`Resize ${col.header} column`}
                      aria-valuenow={widthOf(col)}
                      aria-valuemin={col.minWidth ?? DEFAULT_MIN_WIDTH}
                      tabindex="0"
                      title="Drag to resize · double-click to reset"
                      onPointerDown={(e) => startResize(col, e)}
                      onKeyDown={(e) => resizeWithKeys(col, e)}
                      onDblClick={() => resetWidth(col)}
                    />
                  </th>
                )}
              </For>
              <th class="data-table__filler" aria-hidden="true" />
              <Show when={hasActions()}>
                <th class="data-table__actions">
                  <span class="data-table__sr-only">Actions</span>
                </th>
              </Show>
            </tr>

            <Show when={showFilters()}>
              <tr class="data-table__filters">
                <Show when={selectable()}>
                  <td />
                </Show>
                <For each={columns()}>
                  {(col) => {
                    const filter = filterOf(col);
                    return (
                      <td>
                        <Show when={filter !== false}>
                          <Show
                            when={typeof filter === "object" && filter}
                            fallback={
                              <input
                                class="data-table__filter"
                                type="text"
                                name={`filter-${col.id}`}
                                placeholder="Filter…"
                                aria-label={`Filter ${col.header}`}
                                value={filters[col.id] ?? ""}
                                onInput={(e) => setFilters(col.id, e.currentTarget.value)}
                              />
                            }>
                            {(f) => (
                              <select
                                class="data-table__filter"
                                name={`filter-${col.id}`}
                                aria-label={`Filter ${col.header}`}
                                onChange={(e) => setFilters(col.id, e.currentTarget.value)}>
                                {/* Per-option `selected`: a <select value> is set before its options exist. */}
                                <option value="" selected={!filters[col.id]}>
                                  All
                                </option>
                                <For each={f().options}>
                                  {(option) => (
                                    <option
                                      value={optionValue(option)}
                                      selected={filters[col.id] === optionValue(option)}>
                                      {optionLabel(option)}
                                    </option>
                                  )}
                                </For>
                              </select>
                            )}
                          </Show>
                        </Show>
                      </td>
                    );
                  }}
                </For>
                <td />
                <Show when={hasActions()}>
                  <td class="data-table__actions">
                    <Show when={activeFilterCount() > 0}>
                      <Button
                        size="sm"
                        variant="ghost"
                        iconOnly
                        aria-label="Clear filters"
                        title="Clear filters"
                        onClick={clearColumnFilters}>
                        <Icon name="x" />
                      </Button>
                    </Show>
                  </td>
                </Show>
              </tr>
            </Show>
          </thead>

          <tbody>
            <Show
              when={pageRows().length > 0 && !loadError()}
              fallback={
                <tr>
                  <td class="data-table__empty" colSpan={columnCount()}>
                    <Show
                      when={!loadError()}
                      fallback={
                        <span class="data-table__error" role="alert">
                          Couldn't load the data: {errorText(loadError())}.{" "}
                          <button type="button" class="data-table__link" onClick={() => props.source?.refetch()}>
                            Try again
                          </button>
                        </span>
                      }>
                    <Show
                      when={!loading()}
                      fallback="Loading…">
                      <Show
                        when={search() || activeFilterCount() > 0}
                        fallback={props.emptyMessage ?? "Nothing here yet."}>
                        No rows match your search or filters.{" "}
                        <button type="button" class="data-table__link" onClick={clearAll}>
                          Clear all filters
                        </button>
                      </Show>
                    </Show>
                    </Show>
                  </td>
                </tr>
              }>
              <For each={pageRows()}>
                {(row) => (
                  <tr classList={{ "data-table__row--selected": isSelected(row) }}>
                    <Show when={selectable()}>
                      <td class="data-table__select">
                        <input
                          type="checkbox"
                          name="select-row"
                          value={props.rowKey(row)}
                          aria-label={`Select ${props.rowLabel?.(row) ?? "row"}`}
                          checked={isSelected(row)}
                          onChange={() => toggleRow(row)}
                        />
                      </td>
                    </Show>
                    <Index each={columns()}>
                      {(col) => (
                        <td class={`data-table__cell--${col().align ?? "start"}`}>
                          {col().cell ? col().cell!(row) : textOf(col(), row)}
                        </td>
                      )}
                    </Index>
                    <td class="data-table__filler" aria-hidden="true" />
                    <Show when={hasActions()}>
                      <td class="data-table__actions">
                        <DropdownMenu
                          variant="icon"
                          align="end"
                          label={`Actions for ${props.rowLabel?.(row) ?? "row"}`}
                          trigger={<Icon name="more-horizontal" />}
                          items={menuItems(row)}
                        />
                      </td>
                    </Show>
                  </tr>
                )}
              </For>
            </Show>
          </tbody>
        </table>
      </div>

      {/* --- Footer / paging ----------------------------------------------- */}
      <div class="data-table__footer">
        <label class="data-table__page-size">
          Rows per page
          <select name="page-size" onChange={(e) => setPageSize(Number(e.currentTarget.value))}>
            <For each={pageSizes()}>
              {(size) => (
                <option value={size} selected={size === pageSize()}>
                  {size}
                </option>
              )}
            </For>
          </select>
        </label>

        <span class="data-table__range">
          {rangeStart()}–{rangeEnd()} of {total()}
        </span>

        <div class="data-table__pager">
          <Button size="sm" variant="ghost" iconOnly aria-label="First page" disabled={page() === 0} onClick={() => setPage(0)}>
            <Icon name="chevrons-left" />
          </Button>
          <Button size="sm" variant="ghost" iconOnly aria-label="Previous page" disabled={page() === 0} onClick={() => setPage((p) => p - 1)}>
            <Icon name="chevron-left" />
          </Button>
          <span class="data-table__page">
            Page {page() + 1} of {pageCount()}
          </span>
          <Button size="sm" variant="ghost" iconOnly aria-label="Next page" disabled={page() >= pageCount() - 1} onClick={() => setPage((p) => p + 1)}>
            <Icon name="chevron-right" />
          </Button>
          <Button size="sm" variant="ghost" iconOnly aria-label="Last page" disabled={page() >= pageCount() - 1} onClick={() => setPage(pageCount() - 1)}>
            <Icon name="chevrons-right" />
          </Button>
        </div>
      </div>
    </div>
  );
}
