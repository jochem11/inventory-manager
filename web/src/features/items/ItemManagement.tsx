import "./itemManagement.scss";
import * as api from "~/api/items";
import type { Item } from "~/api/items";
import { DataTable, ErrorNotice, PageHeader } from "~/components/ui";
import { PERMISSIONS } from "~/constants/access";
import { useAuth } from "~/context";
import { createDeleteAction, createServerTable } from "~/hooks";
import { createItemEditor, createItemOptions } from "./hooks";
import { itemActions, itemColumns } from "./table";

/**
 * The inventory: the paged items table with add, edit, duplicate and delete.
 * Needs items:read (guard it where it's rendered); without items:write it's
 * read-only.
 */
export function ItemManagement() {
  const auth = useAuth();
  const canWrite = () => auth.hasPermission(PERMISSIONS.itemsWrite);

  // The table asks for a page; we fetch exactly that page from the API. The
  // search lives in `?q=`, which the topbar search box fills too.
  const items = createServerTable({
    fetch: (variables) => api.fetchItems(auth.request, variables),
    sortFields: api.ITEM_SORT_FIELDS,
    filters: api.itemFilters,
  });
  const options = createItemOptions();
  const editor = createItemEditor(items, options);
  const remove = createDeleteAction<Item>({
    source: items,
    noun: ["item", "items"],
    name: (item) => `"${item.name}"`,
    remove: (item) => api.deleteItem(auth.request, item.id),
    message: "They will be removed from your inventory. This can't be undone.",
  });

  return (
    <>
      <PageHeader title="All items" description={items.summary("item", "items")} />

      <section class="item-management">
        <ErrorNotice message={remove.error()} />

        <DataTable
          source={items}
          columns={itemColumns(options.categories(), options.statuses())}
          rowKey={(item) => item.id}
          rowLabel={(item) => item.name}
          actions={itemActions({ canWrite, onEdit: editor.edit, onDuplicate: editor.duplicate, onDelete: remove.run })}
          onAdd={canWrite() ? editor.add : undefined}
          addLabel="Add item"
          onDelete={canWrite() ? remove.run : undefined}
          searchPlaceholder="Search items…"
          storageKey="items-table"
          initialSort={{ id: "name", direction: "asc" }}
          emptyMessage="No items yet. Add your first one."
        />
      </section>
    </>
  );
}
