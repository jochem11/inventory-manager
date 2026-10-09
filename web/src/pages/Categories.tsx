import { categoriesApi } from "~/api/named";
import { PERMISSIONS } from "~/constants/access";
import { Forbidden, Guard } from "~/features/auth";
import { NamedRecordManagement } from "~/features/namedRecords";

export default function Categories() {
  return (
    <Guard permission={PERMISSIONS.itemsRead} fallback={<Forbidden />}>
      <NamedRecordManagement title="Categories" noun={["category", "categories"]} api={categoriesApi} storageKey="categories-table" />
    </Guard>
  );
}
