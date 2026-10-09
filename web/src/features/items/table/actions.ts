import type { Item } from "~/api/items";
import type { DataTableAction } from "~/components/ui";

export type ItemActionOptions = {
  /** Whether the user may change items (items:write). */
  canWrite: () => boolean;
  onEdit: (item: Item) => void;
  onDuplicate: (item: Item) => void;
  onDelete: (items: Item[]) => void;
};

/** The "⋯" menu of a row; changes only show with items:write. */
export const itemActions = (options: ItemActionOptions): DataTableAction<Item>[] => [
  { label: "Edit", icon: "pencil", hidden: () => !options.canWrite(), onSelect: options.onEdit },
  { label: "Duplicate", icon: "copy", hidden: () => !options.canWrite(), onSelect: options.onDuplicate },
  { divider: true },
  {
    label: "Delete",
    icon: "trash",
    danger: true,
    hidden: () => !options.canWrite(),
    onSelect: (item) => options.onDelete([item]),
  },
];
