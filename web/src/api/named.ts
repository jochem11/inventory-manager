import type { Connection, ListVariables } from "~/api/list";
import type { AuthContextValue } from "~/context/authContext";

// ---------------------------------------------------------------------------
// Records that are just a name: categories and item statuses. Their GraphQL
// operations have the same shape, so one factory builds both APIs.
// ---------------------------------------------------------------------------

type Request = AuthContextValue["request"];

export type NamedRecord = { id: string; name: string };

export type NamedRecordDetails = NamedRecord & {
  /** ISO timestamp. */
  createdAt: string;
  /** ISO timestamp. */
  updatedAt: string;
};

export type NameSortField = "NAME" | "CREATED_AT" | "UPDATED_AT";

/** DataTable column id → API sort field. The filter fields match the column ids. */
export const NAME_SORT_FIELDS: Record<string, NameSortField> = {
  name: "NAME",
  createdAt: "CREATED_AT",
  updatedAt: "UPDATED_AT",
};

export type NamedRecordApi = ReturnType<typeof namedRecordApi>;

const FIELDS = "id name createdAt updatedAt";

/**
 * The list, create, update and delete calls of one kind of named record, e.g.
 * `namedRecordApi({ list: "categories", type: "Category" })` uses
 * `categories`, `createCategory`, `updateCategory` and `deleteCategory`.
 */
export function namedRecordApi(names: { list: string; type: string }) {
  const { list, type } = names;
  return {
    fetch: async (request: Request, variables: ListVariables<NameSortField>) =>
      (
        await request<Record<string, Connection<NamedRecordDetails>>>(
          `query List($offset: Int!, $limit: Int!, $orderBy: NameOrder, $filter: NameFilter) {
            ${list}(offset: $offset, limit: $limit, orderBy: $orderBy, filter: $filter) {
              totalCount
              nodes { ${FIELDS} }
            }
          }`,
          variables,
        )
      )[list]!,

    create: async (request: Request, name: string) =>
      (
        await request<Record<string, NamedRecordDetails>>(
          `mutation Create($name: String!) { create${type}(name: $name) { ${FIELDS} } }`,
          { name },
        )
      )[`create${type}`]!,

    update: async (request: Request, id: string, name: string) =>
      (
        await request<Record<string, NamedRecordDetails>>(
          `mutation Update($id: ID!, $name: String!) { update${type}(id: $id, name: $name) { ${FIELDS} } }`,
          { id, name },
        )
      )[`update${type}`]!,

    remove: async (request: Request, id: string) =>
      (await request<Record<string, string>>(`mutation Delete($id: ID!) { delete${type}(id: $id) }`, { id }))[
        `delete${type}`
      ]!,

    /** All records (up to 100), sorted by name, e.g. for a select field's options. */
    all: async (request: Request) =>
      (
        await request<Record<string, Connection<NamedRecord>>>(
          `query All { ${list}(limit: 100, orderBy: { field: NAME }) { nodes { id name } } }`,
        )
      )[list]!.nodes,
  };
}

export const categoriesApi = namedRecordApi({ list: "categories", type: "Category" });
export const itemStatusesApi = namedRecordApi({ list: "itemStatuses", type: "ItemStatus" });
