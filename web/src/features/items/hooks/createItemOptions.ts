import { createResource } from "solid-js";
import { categoriesApi, itemStatusesApi } from "~/api/named";
import type { SelectOption } from "~/components/form/selectField";
import { useAuth } from "~/context";

/**
 * The categories and statuses, as options for select fields and table
 * filters. `refetch` reloads them, e.g. after one was added elsewhere.
 */
export function createItemOptions() {
  const auth = useAuth();
  const toOptions = (records: { id: string; name: string }[]): SelectOption[] =>
    records.map((r) => ({ value: r.id, label: r.name }));

  const [categories, categoriesControl] = createResource(async () => toOptions(await categoriesApi.all(auth.request)), {
    initialValue: [],
  });
  const [statuses, statusesControl] = createResource(async () => toOptions(await itemStatusesApi.all(auth.request)), {
    initialValue: [],
  });

  return {
    categories,
    statuses,
    refetch: () => {
      categoriesControl.refetch();
      statusesControl.refetch();
    },
  };
}

export type ItemOptions = ReturnType<typeof createItemOptions>;
