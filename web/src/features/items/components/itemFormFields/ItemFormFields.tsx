import "./itemFormFields.scss";
import { SelectField, TextAreaField, TextField } from "~/components/form";
import type { SelectOption } from "~/components/form/selectField";

export type ItemFormFieldsProps = {
  categories: SelectOption[];
  statuses: SelectOption[];
};

/** An item's fields; use inside a form with `itemSchema`. */
export function ItemFormFields(props: ItemFormFieldsProps) {
  return (
    <>
      <TextField name="name" label="Name" placeholder="e.g. ThinkPad T14" />
      <div class="item-form-fields__row">
        <SelectField name="categoryId" label="Category" placeholder="Pick a category" options={props.categories} />
        <SelectField name="statusId" label="Status" placeholder="Pick a status" options={props.statuses} />
      </div>
      <TextAreaField name="description" label="Description" rows={3} hint="Optional" />
      <TextField name="imageUrl" type="url" label="Image URL" placeholder="https://…" hint="Optional" />
    </>
  );
}
