import "./namedRecordManagement.scss";
import { z } from "zod";
import { fieldErrorMessage } from "~/api/graphql";
import { NAME_SORT_FIELDS, type NamedRecordApi, type NamedRecordDetails } from "~/api/named";
import { TextField } from "~/components/form";
import { DataTable, ErrorNotice, PageHeader, type DataTableAction, type DataTableColumn } from "~/components/ui";
import { PERMISSIONS } from "~/constants/access";
import { useAuth } from "~/context";
import { createDeleteAction, createFormDialog, createServerTable } from "~/hooks";

export type NamedRecordManagementProps = {
  title: string;
  description?: string;
  /** E.g. `["category", "categories"]`. */
  noun: [singular: string, plural: string];
  api: NamedRecordApi;
  /** Remembers the table's columns under this key. */
  storageKey: string;
};

const dateFormat = new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" });
const formatDate = (iso: string) => dateFormat.format(new Date(iso));

const COLUMNS: DataTableColumn<NamedRecordDetails>[] = [
  {
    id: "name",
    header: "Name",
    value: (row) => row.name,
    cell: (row) => <span class="named-records__name">{row.name}</span>,
    width: 260,
    hideable: false,
  },
  {
    id: "createdAt",
    header: "Created",
    value: (row) => new Date(row.createdAt),
    text: (row) => formatDate(row.createdAt),
    filter: false,
    width: 200,
  },
  {
    id: "updatedAt",
    header: "Updated",
    value: (row) => new Date(row.updatedAt),
    text: (row) => formatDate(row.updatedAt),
    filter: false,
    width: 200,
    hidden: true,
  },
];

const nameSchema = z.object({
  name: z.string({ error: "Enter a name" }).trim().min(1, "Enter a name").max(100, "At most 100 characters"),
});

/**
 * A page managing records that are just a name, like categories and item
 * statuses: a server-side table with add, rename and delete. Needs
 * items:read; without items:write it's read-only.
 */
export function NamedRecordManagement(props: NamedRecordManagementProps) {
  const auth = useAuth();
  const canWrite = () => auth.hasPermission(PERMISSIONS.itemsWrite);
  const [singular, plural] = props.noun;

  const records = createServerTable({
    fetch: (variables) => props.api.fetch(auth.request, variables),
    sortFields: NAME_SORT_FIELDS,
  });
  const formDialog = createFormDialog();
  const remove = createDeleteAction<NamedRecordDetails>({
    source: records,
    noun: props.noun,
    name: (row) => `"${row.name}"`,
    remove: (row) => props.api.remove(auth.request, row.id),
    message: `Only ${plural} without items can be deleted.`,
  });

  const openForm = (row?: NamedRecordDetails) =>
    formDialog.open({
      title: row ? `Rename ${singular}` : `New ${singular}`,
      schema: nameSchema,
      defaultValues: { name: row?.name ?? "" },
      fields: () => <TextField name="name" label="Name" />,
      submitLabel: row ? "Save" : `Add ${singular}`,
      onSubmit: ({ name }) =>
        records.mutate(() => (row ? props.api.update(auth.request, row.id, name) : props.api.create(auth.request, name))),
      formatError: fieldErrorMessage,
    });

  const actions: DataTableAction<NamedRecordDetails>[] = [
    { label: "Rename", icon: "pencil", hidden: () => !canWrite(), onSelect: (row) => openForm(row) },
    { divider: true },
    { label: "Delete", icon: "trash", danger: true, hidden: () => !canWrite(), onSelect: (row) => remove.run([row]) },
  ];

  return (
    <>
      <PageHeader title={props.title} description={props.description ?? records.summary(singular, plural)} />

      <section class="named-records">
        <ErrorNotice message={remove.error()} />

        <DataTable
          source={records}
          columns={COLUMNS}
          rowKey={(row) => row.id}
          rowLabel={(row) => row.name}
          actions={actions}
          onAdd={canWrite() ? () => openForm() : undefined}
          addLabel={`Add ${singular}`}
          onDelete={canWrite() ? remove.run : undefined}
          searchPlaceholder={`Search ${plural}…`}
          storageKey={props.storageKey}
          initialSort={{ id: "name", direction: "asc" }}
          emptyMessage={`No ${plural} yet.`}
        />
      </section>
    </>
  );
}
