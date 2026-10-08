import "./home.scss";
import { createSignal, Show } from "solid-js";
import { z } from "zod";
import * as api from "~/api/items";
import type { Item } from "~/api/items";
import { ConfirmDialog, PromptDialog } from "~/components/dialogs";
import {
  CheckboxField,
  Form,
  SelectField,
  SubmitButton,
  TextAreaField,
  TextField,
} from "~/components/form";
import {
  Button,
  DataTable,
  ErrorNotice,
  PageHeader,
  type DataTableAction,
  type DataTableColumn,
} from "~/components/ui";
import { createDeleteAction, createDialog, createForm, createServerTable } from "~/hooks";

const CATEGORIES = [
  { value: "tools", label: "Tools" },
  { value: "hardware", label: "Hardware" },
  { value: "electrical", label: "Electrical" },
  { value: "other", label: "Other" },
];

const itemSchema = z.object({
  name: z.string().trim().min(1, "Give the item a name"),
  quantity: z
    .number({ error: "Enter a quantity" })
    .int("Whole numbers only")
    .min(0, "Can't be negative"),
  category: z.string({ error: "Pick a category" }),
  notes: z.string().optional(),
  lowStockAlert: z.boolean(),
});

const defaultValues = { quantity: 1, lowStockAlert: false };

const categoryLabel = (value: string) =>
  CATEGORIES.find((c) => c.value === value)?.label ?? value;

// Column ids double as the keys of ITEM_SORT_FIELDS and the filters itemFilters maps.
const COLUMNS: DataTableColumn<Item>[] = [
  {
    id: "name",
    header: "Name",
    value: (item) => item.name,
    cell: (item) => <span class="home__name">{item.name}</span>,
    width: 220,
    minWidth: 120,
    hideable: false,
  },
  {
    id: "category",
    header: "Category",
    value: (item) => item.category,
    text: (item) => categoryLabel(item.category),
    filter: { options: CATEGORIES },
    width: 140,
  },
  {
    id: "quantity",
    header: "Quantity",
    value: (item) => item.quantity,
    cell: (item) => (
      <span classList={{ "home__qty--empty": item.quantity === 0 }}>{item.quantity} pcs</span>
    ),
    align: "end",
    width: 120,
  },
  {
    id: "alert",
    header: "Low-stock alert",
    value: (item) => item.lowStockAlert,
    text: (item) => (item.lowStockAlert ? "On" : "Off"),
    cell: (item) => (
      <span class="home__badge" classList={{ "home__badge--on": item.lowStockAlert }}>
        {item.lowStockAlert ? "On" : "Off"}
      </span>
    ),
    filter: {
      options: [
        { value: "true", label: "On" },
        { value: "false", label: "Off" },
      ],
    },
    width: 150,
  },
  {
    id: "notes",
    header: "Location / notes",
    value: (item) => item.notes,
    cell: (item) => <span class="home__notes">{item.notes || "—"}</span>,
    width: 260,
    // Starts hidden: turn it on from the table's Columns menu.
    hidden: true,
  },
];

export default function Home() {
  const [showForm, setShowForm] = createSignal(false);

  const confirm = createDialog(ConfirmDialog);
  const prompt = createDialog(PromptDialog);

  // The table asks for a page; we fetch exactly that page from the API. The
  // search lives in `?q=`, which the topbar search box fills too.
  const items = createServerTable({
    fetch: api.fetchItems,
    sortFields: api.ITEM_SORT_FIELDS,
    filters: api.itemFilters,
  });

  // Built here (rather than via `<Form schema>`) so we can reset it after adding.
  const form = createForm({
    schema: itemSchema,
    defaultValues,
    onSubmit: async (input) => {
      await items.mutate(() => api.createItem(input));
      form.reset();
      setShowForm(false);
    },
  });

  const remove = createDeleteAction<Item>({
    source: items,
    noun: ["item", "items"],
    name: (item) => `"${item.name}"`,
    removeMany: (toRemove) => api.deleteItems(toRemove.map((i) => i.id)),
    message: "They will be removed from your inventory. This can't be undone.",
  });

  const update = (item: Item, patch: api.ItemPatch) =>
    items.mutate(() => api.updateItem(item.id, patch));

  const ACTIONS: DataTableAction<Item>[] = [
    {
      label: "Rename",
      icon: "pencil",
      onSelect: async (item) => {
        const name = await prompt.open({ title: "Rename item", label: "Name", initialValue: item.name });
        if (name) await update(item, { name });
      },
    },
    {
      label: "Duplicate",
      icon: "copy",
      onSelect: ({ id, ...rest }) =>
        items.mutate(() => api.createItem({ ...rest, name: `${rest.name} (copy)` })),
    },
    {
      label: "Turn alert on",
      icon: "info",
      hidden: (item) => item.lowStockAlert,
      onSelect: (item) => update(item, { lowStockAlert: true }),
    },
    {
      label: "Turn alert off",
      icon: "info",
      hidden: (item) => !item.lowStockAlert,
      onSelect: (item) => update(item, { lowStockAlert: false }),
    },
    { divider: true },
    { label: "Delete", icon: "trash", danger: true, onSelect: (item) => remove.run([item]) },
  ];

  const clearAll = async () => {
    const ok = await confirm.open({
      title: "Clear inventory?",
      message: "Every item will be removed.",
      confirmLabel: "Clear all",
      variant: "danger",
    });
    if (!ok) return;
    await items.mutate(() => api.deleteAllItems());
  };

  return (
    <>
      <PageHeader
        title="All items"
        description={items.summary("item", "items")}
        actions={
          <Button variant="outline" onClick={clearAll}>
            Clear all
          </Button>
        }
      />

      <section class="home">
        <ErrorNotice message={remove.error()} />

        <Show when={showForm()}>
          <Form form={form} class="home__form">
            <div class="home__form-header">
              <h2 class="home__heading">New item</h2>
              <Button
                size="sm"
                variant="ghost"
                onClick={() => {
                  form.reset();
                  setShowForm(false);
                }}>
                Cancel
              </Button>
            </div>
            <div class="home__row">
              <TextField name="name" label="Name" placeholder="e.g. Drill" />
              <TextField
                name="quantity"
                label="Quantity"
                type="number"
                min={0}
                variant="filled"
                suffix="pcs"
              />
            </div>
            <SelectField
              name="category"
              label="Category"
              placeholder="Pick a category"
              options={CATEGORIES}
            />
            <TextAreaField
              name="notes"
              label="Location / notes"
              placeholder="Where is it stored?"
              variant="underlined"
              rows={2}
              hint="Optional"
            />
            <CheckboxField name="lowStockAlert" label="Alert me when stock is low" variant="switch" />
            <SubmitButton submittingText="Adding…">Add item</SubmitButton>
          </Form>
        </Show>

        <DataTable
          source={items}
          columns={COLUMNS}
          rowKey={(item) => item.id}
          rowLabel={(item) => item.name}
          actions={ACTIONS}
          onAdd={() => setShowForm(true)}
          addLabel="Add item"
          onDelete={remove.run}
          searchPlaceholder="Search items…"
          storageKey="items-table"
          initialSort={{ id: "name", direction: "asc" }}
          emptyMessage="No items yet. Add your first one."
        />
      </section>
    </>
  );
}
