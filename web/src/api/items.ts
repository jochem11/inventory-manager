import type { Connection, ListVariables } from "~/api/list";
import seedItems from "~/data/items.json";

// ---------------------------------------------------------------------------
// Items API — a stand-in for the GraphQL API.
//
// Each exported function mirrors the operation it will become, so pages can be
// written against the real shapes today. When the API exists, replace the
// function bodies with GraphQL requests; the signatures stay the same:
//
//   query Items($offset: Int!, $limit: Int!, $orderBy: ItemOrder, $filter: ItemFilter) {
//     items(offset: $offset, limit: $limit, orderBy: $orderBy, filter: $filter) {
//       totalCount
//       nodes { id name quantity category notes lowStockAlert }
//     }
//   }
//   mutation CreateItem($input: ItemInput!)          { createItem(input: $input) { id … } }
//   mutation UpdateItem($id: ID!, $input: ItemPatch!) { updateItem(id: $id, input: $input) { id … } }
//   mutation DeleteItems($ids: [ID!]!)               { deleteItems(ids: $ids) }
// ---------------------------------------------------------------------------

export type Item = {
  id: string;
  name: string;
  quantity: number;
  category: string;
  notes?: string;
  lowStockAlert: boolean;
};

export type ItemInput = Omit<Item, "id">;
export type ItemPatch = Partial<ItemInput>;

export type ItemSortField = "name" | "quantity" | "category" | "notes" | "lowStockAlert";

/** The `filter` of the items query. */
export type ItemFilter = {
  search?: string;
  name?: string;
  category?: string;
  quantity?: string;
  notes?: string;
  lowStockAlert?: boolean;
};

/** DataTable column id → API sort field. */
export const ITEM_SORT_FIELDS: Record<string, ItemSortField> = {
  name: "name",
  category: "category",
  quantity: "quantity",
  notes: "notes",
  alert: "lowStockAlert",
};

/** The column filters as API filter fields: the "alert" column is the boolean lowStockAlert. */
export const itemFilters = ({ alert, ...rest }: Record<string, string>): Omit<ItemFilter, "search"> => ({
  ...rest,
  lowStockAlert: alert === undefined ? undefined : alert === "true",
});

// ---------------------------------------------------------------------------
// Mock server below — delete when the real API is wired up.
// ---------------------------------------------------------------------------

let db: Item[] = structuredClone(seedItems);
const LATENCY_MS = 250;
const delay = <T>(value: T) => new Promise<T>((resolve) => setTimeout(() => resolve(value), LATENCY_MS));
const includes = (text: string | undefined, part: string | undefined) =>
  !part || (text ?? "").toLowerCase().includes(part.toLowerCase());

/** `query Items` */
export function fetchItems(variables: ListVariables<ItemSortField>): Promise<Connection<Item>> {
  const { orderBy } = variables;
  const filter = variables.filter as ItemFilter;
  let rows = db.filter(
    (item) =>
      (includes(item.name, filter.search) ||
        includes(item.notes, filter.search) ||
        includes(item.category, filter.search)) &&
      includes(item.name, filter.name) &&
      includes(item.notes, filter.notes) &&
      includes(String(item.quantity), filter.quantity) &&
      (!filter.category || item.category === filter.category) &&
      (filter.lowStockAlert === undefined || item.lowStockAlert === filter.lowStockAlert),
  );

  if (orderBy) {
    const dir = orderBy.direction === "ASC" ? 1 : -1;
    rows = [...rows].sort((a, b) => {
      const va = a[orderBy.field];
      const vb = b[orderBy.field];
      if (va === undefined || vb === undefined) return Number(va === undefined) - Number(vb === undefined);
      if (typeof va === "number" && typeof vb === "number") return (va - vb) * dir;
      return String(va).localeCompare(String(vb), undefined, { numeric: true }) * dir;
    });
  }

  return delay({
    totalCount: rows.length,
    nodes: structuredClone(rows.slice(variables.offset, variables.offset + variables.limit)),
  });
}

/** `mutation CreateItem` */
export function createItem(input: ItemInput): Promise<Item> {
  const item = { ...input, id: crypto.randomUUID() };
  db = [...db, item];
  return delay(structuredClone(item));
}

/** `mutation UpdateItem` */
export function updateItem(id: string, patch: ItemPatch): Promise<Item> {
  db = db.map((item) => (item.id === id ? { ...item, ...patch } : item));
  return delay(structuredClone(db.find((item) => item.id === id)!));
}

/** `mutation DeleteItems` — returns how many were removed. */
export function deleteItems(ids: string[]): Promise<number> {
  const remove = new Set(ids);
  const before = db.length;
  db = db.filter((item) => !remove.has(item.id));
  return delay(before - db.length);
}

/** `mutation DeleteAllItems` */
export function deleteAllItems(): Promise<number> {
  const count = db.length;
  db = [];
  return delay(count);
}
