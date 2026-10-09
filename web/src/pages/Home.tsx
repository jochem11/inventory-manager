import { PERMISSIONS } from "~/constants/access";
import { Forbidden, Guard } from "~/features/auth";
import { ItemManagement } from "~/features/items";

export default function Home() {
  return (
    <Guard permission={PERMISSIONS.itemsRead} fallback={<Forbidden />}>
      <ItemManagement />
    </Guard>
  );
}
