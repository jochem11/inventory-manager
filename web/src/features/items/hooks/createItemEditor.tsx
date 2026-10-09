import { fieldErrorMessage } from "~/api/graphql";
import * as api from "~/api/items";
import type { Item } from "~/api/items";
import { useAuth } from "~/context";
import { createFormDialog, type TableSource } from "~/hooks";
import { ItemFormFields } from "../components/itemFormFields";
import { itemSchema } from "../utils/schema";
import type { ItemOptions } from "./createItemOptions";

/**
 * Adding and editing items in a dialog. Saving runs through the table's
 * `mutate`, so the table refreshes; a failed save keeps the dialog open with
 * the error.
 */
export function createItemEditor(items: TableSource<Item>, options: ItemOptions) {
  const auth = useAuth();
  const formDialog = createFormDialog();

  const open = (item?: Item) =>
    formDialog.open({
      title: item ? `Edit ${item.name}` : "New item",
      schema: itemSchema,
      defaultValues: item
        ? {
            name: item.name,
            description: item.description ?? "",
            imageUrl: item.imageUrl ?? "",
            categoryId: item.category.id,
            statusId: item.status.id,
          }
        : { statusId: options.statuses()[0]?.value },
      fields: () => <ItemFormFields categories={options.categories()} statuses={options.statuses()} />,
      submitLabel: item ? "Save changes" : "Add item",
      onSubmit: (input) =>
        items.mutate(() => (item ? api.updateItem(auth.request, item.id, input) : api.createItem(auth.request, input))),
      formatError: fieldErrorMessage,
    });

  return {
    add: () => open(),
    edit: (item: Item) => open(item),
    duplicate: (item: Item) =>
      items.mutate(() => api.createItem(auth.request, { ...api.itemToInput(item), name: `${item.name} (copy)` })),
  };
}
