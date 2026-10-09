import type { Connection, ListVariables } from "~/api/list";
import type { AuthContextValue } from "~/context/authContext";
import type { NamedRecord } from "~/api/named";

// ---------------------------------------------------------------------------
// Items API: the gateway's items queries and mutations (schema/item.graphqls).
// They need a login with items:read / items:write, so every function takes
// `request` from `useAuth()`.
// ---------------------------------------------------------------------------

type Request = AuthContextValue["request"];

export type Item = {
  id: string;
  name: string;
  description?: string | null;
  imageUrl?: string | null;
  category: NamedRecord;
  status: NamedRecord;
  /** ISO timestamp. */
  createdAt: string;
  /** ISO timestamp. */
  updatedAt: string;
};

/** Every field is replaced on update: leaving description or imageUrl out clears it. */
export type ItemInput = {
  name: string;
  description?: string;
  imageUrl?: string;
  categoryId: string;
  statusId: string;
};

export type ItemSortField = "NAME" | "CREATED_AT" | "UPDATED_AT";

/** DataTable column id → API sort field. */
export const ITEM_SORT_FIELDS: Record<string, ItemSortField> = {
  name: "NAME",
  createdAt: "CREATED_AT",
  updatedAt: "UPDATED_AT",
};

/** The column filters as API filter fields: the category and status columns filter by id. */
export const itemFilters = ({ category, status, ...rest }: Record<string, string>) => ({
  ...rest,
  categoryId: category,
  statusId: status,
});

const ITEM_FIELDS = "id name description imageUrl category { id name } status { id name } createdAt updatedAt";

export const fetchItems = async (request: Request, variables: ListVariables<ItemSortField>) =>
  (
    await request<{ items: Connection<Item> }>(
      `query Items($offset: Int!, $limit: Int!, $orderBy: ItemOrder, $filter: ItemFilter) {
        items(offset: $offset, limit: $limit, orderBy: $orderBy, filter: $filter) {
          totalCount
          nodes { ${ITEM_FIELDS} }
        }
      }`,
      variables,
    )
  ).items;

export const createItem = async (request: Request, input: ItemInput) =>
  (
    await request<{ createItem: Item }>(
      `mutation CreateItem($input: ItemInput!) { createItem(input: $input) { ${ITEM_FIELDS} } }`,
      { input },
    )
  ).createItem;

export const updateItem = async (request: Request, id: string, input: ItemInput) =>
  (
    await request<{ updateItem: Item }>(
      `mutation UpdateItem($id: ID!, $input: ItemInput!) { updateItem(id: $id, input: $input) { ${ITEM_FIELDS} } }`,
      { id, input },
    )
  ).updateItem;

export const deleteItem = async (request: Request, id: string) =>
  (await request<{ deleteItem: string }>(`mutation DeleteItem($id: ID!) { deleteItem(id: $id) }`, { id })).deleteItem;

/** The input that recreates item, e.g. to edit or duplicate it. */
export const itemToInput = (item: Item): ItemInput => ({
  name: item.name,
  description: item.description ?? undefined,
  imageUrl: item.imageUrl ?? undefined,
  categoryId: item.category.id,
  statusId: item.status.id,
});
