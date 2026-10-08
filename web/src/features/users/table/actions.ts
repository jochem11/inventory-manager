import type { User } from "~/api/users";
import type { DataTableAction } from "~/components/ui";

export type UserActionOptions = {
  /** Whether the user may edit and delete (users:write). */
  canWrite: () => boolean;
  /** Whether a row is the logged-in user, who can't delete themselves. */
  isMe: (user: User) => boolean;
  onEdit: (user: User) => void;
  onDelete: (users: User[]) => void;
};

/** The "⋯" menu of a row. Edit and Delete only show with users:write. */
export const userActions = (options: UserActionOptions): DataTableAction<User>[] => [
  { label: "Edit", icon: "pencil", hidden: () => !options.canWrite(), onSelect: options.onEdit },
  {
    label: "Copy email",
    icon: "copy",
    onSelect: (user) => void navigator.clipboard?.writeText(user.email),
  },
  { divider: true },
  {
    label: "Delete",
    icon: "trash",
    danger: true,
    hidden: (user) => !options.canWrite() || options.isMe(user),
    onSelect: (user) => options.onDelete([user]),
  },
];
