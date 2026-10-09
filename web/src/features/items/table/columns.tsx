import "./itemsTable.scss";
import type { Item } from "~/api/items";
import type { SelectOption } from "~/components/form/selectField";
import type { DataTableColumn } from "~/components/ui";

const dateFormat = new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" });
const formatDate = (iso: string) => dateFormat.format(new Date(iso));

/**
 * The items table's columns. The category and status columns filter on the
 * given options (by id); their ids map to the API's filter in itemFilters.
 */
export const itemColumns = (categories: SelectOption[], statuses: SelectOption[]): DataTableColumn<Item>[] => [
  {
    id: "name",
    header: "Name",
    value: (item) => item.name,
    cell: (item) => <span class="items-table__name">{item.name}</span>,
    width: 240,
    minWidth: 120,
    hideable: false,
  },
  {
    id: "category",
    header: "Category",
    value: (item) => item.category.name,
    filter: { options: categories },
    // The API sorts by name, date or id; not by a related table's name.
    sortable: false,
    width: 160,
  },
  {
    id: "status",
    header: "Status",
    value: (item) => item.status.name,
    cell: (item) => <span class="items-table__badge">{item.status.name}</span>,
    filter: { options: statuses },
    sortable: false,
    width: 150,
  },
  {
    id: "description",
    header: "Description",
    value: (item) => item.description,
    cell: (item) => <span class="items-table__muted">{item.description || "—"}</span>,
    // Search covers it; the API has no separate description filter.
    filter: false,
    sortable: false,
    width: 280,
  },
  {
    id: "createdAt",
    header: "Added",
    value: (item) => new Date(item.createdAt),
    text: (item) => formatDate(item.createdAt),
    filter: false,
    width: 180,
    hidden: true,
  },
  {
    id: "updatedAt",
    header: "Updated",
    value: (item) => new Date(item.updatedAt),
    text: (item) => formatDate(item.updatedAt),
    filter: false,
    width: 180,
    hidden: true,
  },
];
