import "./usersTable.scss";
import type { User } from "~/api/users";
import { Avatar, type DataTableColumn } from "~/components/ui";
import { formatDate, fullName } from "../utils/user";

// Column ids double as the API filter fields and the keys of USER_SORT_FIELDS.
export const USER_COLUMNS: DataTableColumn<User>[] = [
  {
    id: "firstName",
    header: "First name",
    value: (user) => user.firstName,
    cell: (user) => (
      <span class="users-table__name">
        <Avatar name={fullName(user)} size="sm" />
        {user.firstName}
      </span>
    ),
    width: 200,
    minWidth: 120,
    hideable: false,
  },
  {
    id: "lastName",
    header: "Last name",
    value: (user) => user.lastName,
    width: 180,
  },
  {
    id: "email",
    header: "Email",
    value: (user) => user.email,
    width: 240,
  },
  {
    id: "phone",
    header: "Phone",
    value: (user) => user.phone,
    cell: (user) => <span class="users-table__muted">{user.phone || "—"}</span>,
    width: 160,
  },
  {
    id: "createdAt",
    header: "Created",
    value: (user) => new Date(user.createdAt),
    text: (user) => formatDate(user.createdAt),
    // The API filters on text fields only.
    filter: false,
    width: 180,
  },
  {
    id: "updatedAt",
    header: "Updated",
    value: (user) => new Date(user.updatedAt),
    text: (user) => formatDate(user.updatedAt),
    filter: false,
    width: 180,
    hidden: true,
  },
];
