import { fieldErrorMessage } from "~/api/graphql";
import * as api from "~/api/users";
import type { User } from "~/api/users";
import { useAuth } from "~/context";
import { createFormDialog, type TableSource } from "~/hooks";
import { UserFormFields } from "../components/userFormFields";
import { userSchema } from "../utils/schema";
import { fullName } from "../utils/user";

/**
 * Editing a user in a dialog. Saving runs through the table's `mutate`, so
 * the table refreshes; a failed save keeps the dialog open with the error.
 */
export function createUserEditor(users: TableSource<User>) {
  const auth = useAuth();
  const formDialog = createFormDialog();

  const open = (user: User) =>
    formDialog.open({
      title: `Edit ${fullName(user)}`,
      description: user.email,
      schema: userSchema,
      defaultValues: { firstName: user.firstName, lastName: user.lastName, phone: user.phone ?? "" },
      fields: () => <UserFormFields />,
      submitLabel: "Save changes",
      // Email and avatar aren't editable here, but an update replaces every
      // field, so send the current ones along. The email stays as is because
      // the login's email lives in the auth-service.
      onSubmit: (values) =>
        users.mutate(() =>
          api.updateUser(auth.request, user.id, { ...values, email: user.email, avatarUrl: user.avatarUrl ?? undefined }),
        ),
      formatError: fieldErrorMessage,
    });

  return { open };
}
