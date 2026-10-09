import { itemStatusesApi } from "~/api/named";
import { PERMISSIONS } from "~/constants/access";
import { Forbidden, Guard } from "~/features/auth";
import { NamedRecordManagement } from "~/features/namedRecords";

export default function ItemStatuses() {
  return (
    <Guard permission={PERMISSIONS.itemsRead} fallback={<Forbidden />}>
      <NamedRecordManagement title="Statuses" noun={["status", "statuses"]} api={itemStatusesApi} storageKey="statuses-table" />
    </Guard>
  );
}
