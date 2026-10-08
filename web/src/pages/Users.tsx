import { PERMISSIONS } from "~/constants/access";
import { Forbidden, Guard } from "~/features/auth";
import { UserManagement } from "~/features/users";

export default function Users() {
  return (
    <Guard permission={PERMISSIONS.usersRead} fallback={<Forbidden />}>
      <UserManagement />
    </Guard>
  );
}
