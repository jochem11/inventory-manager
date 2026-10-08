import type { Connection, ListVariables } from "~/api/list";
import type { AuthContextValue } from "~/context/authContext";

// ---------------------------------------------------------------------------
// Users API: the gateway's `users` queries and mutations (schema/user.graphqls).
// They need a login, so every function takes `request` from `useAuth()`.
// ---------------------------------------------------------------------------

type Request = AuthContextValue["request"];

export type User = {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
  phone?: string | null;
  avatarUrl?: string | null;
  /** ISO timestamp. */
  createdAt: string;
  /** ISO timestamp. */
  updatedAt: string;
};

/** Every field is replaced on update: leaving phone or avatarUrl out clears it. */
export type UserInput = {
  firstName: string;
  lastName: string;
  email: string;
  phone?: string;
  avatarUrl?: string;
};

export type UserSortField = "FIRST_NAME" | "LAST_NAME" | "EMAIL" | "PHONE" | "CREATED_AT" | "UPDATED_AT";

/** DataTable column id → API sort field. The filter fields match the column ids. */
export const USER_SORT_FIELDS: Record<string, UserSortField> = {
  firstName: "FIRST_NAME",
  lastName: "LAST_NAME",
  email: "EMAIL",
  phone: "PHONE",
  createdAt: "CREATED_AT",
  updatedAt: "UPDATED_AT",
};

const USER_FIELDS = "id firstName lastName email phone avatarUrl createdAt updatedAt";

export const fetchUsers = async (request: Request, variables: ListVariables<UserSortField>) =>
  (
    await request<{ users: Connection<User> }>(
      `query Users($offset: Int!, $limit: Int!, $orderBy: UserOrder, $filter: UserFilter) {
        users(offset: $offset, limit: $limit, orderBy: $orderBy, filter: $filter) {
          totalCount
          nodes { ${USER_FIELDS} }
        }
      }`,
      variables,
    )
  ).users;

export const updateUser = async (request: Request, id: string, input: UserInput) =>
  (
    await request<{ updateUser: User }>(
      `mutation UpdateUser($id: ID!, $input: UserInput!) {
        updateUser(id: $id, input: $input) { ${USER_FIELDS} }
      }`,
      { id, input },
    )
  ).updateUser;

/**
 * Deletes the profile. The user-service then publishes UserDeleted, and the
 * auth-service removes the login and its sessions.
 */
export const deleteUser = async (request: Request, id: string) =>
  (
    await request<{ deleteUser: string }>(`mutation DeleteUser($id: ID!) { deleteUser(id: $id) }`, {
      id,
    })
  ).deleteUser;
