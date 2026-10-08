import "./userManagement.scss";
import * as api from "~/api/users";
import type { User } from "~/api/users";
import { DataTable, ErrorNotice, PageHeader } from "~/components/ui";
import { PERMISSIONS } from "~/constants/access";
import { useAuth } from "~/context";
import { createDeleteAction, createServerTable } from "~/hooks";
import { createUserEditor } from "./hooks";
import { USER_COLUMNS, userActions } from "./table";
import { fullName } from "./utils/user";

/**
 * Managing users: the paged users table, editing a user (in a dialog) and
 * deleting users. Needs users:read (guard it where it's rendered); without
 * users:write the table is read-only.
 */
export function UserManagement() {
  const auth = useAuth();
  const canWrite = () => auth.hasPermission(PERMISSIONS.usersWrite);
  const isMe = (user: User) => auth.user()?.id === user.id;

  const users = createServerTable({
    fetch: (variables) => api.fetchUsers(auth.request, variables),
    sortFields: api.USER_SORT_FIELDS,
  });
  const editor = createUserEditor(users);
  const remove = createDeleteAction<User>({
    source: users,
    noun: ["user", "users"],
    name: fullName,
    remove: (user) => api.deleteUser(auth.request, user.id),
    // Deleting yourself would end your own session mid-click.
    skip: (user) => (isMe(user) ? "Your own account can't be deleted here" : undefined),
    message: "Their profile and login are removed. This can't be undone.",
  });

  return (
    <>
      <PageHeader title="Users" description={users.summary("user", "users")} />

      <section class="user-management">
        <ErrorNotice message={remove.error()} />

        <DataTable
          source={users}
          columns={USER_COLUMNS}
          rowKey={(user) => user.id}
          rowLabel={fullName}
          actions={userActions({ canWrite, isMe, onEdit: editor.open, onDelete: remove.run })}
          // Row checkboxes and bulk delete only with users:write.
          onDelete={canWrite() ? remove.run : undefined}
          searchPlaceholder="Search name, email or phone…"
          storageKey="users-table"
          initialSort={{ id: "firstName", direction: "asc" }}
          emptyMessage="No users found."
        />
      </section>
    </>
  );
}
