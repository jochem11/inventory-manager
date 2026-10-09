import { z } from "zod";
import { nameField, optionalPhoneField } from "~/schemas/fields";

/** The editable profile fields, with the same rules the user-service checks. */
export const userSchema = z.object({
  firstName: nameField("Enter a first name"),
  lastName: nameField("Enter a last name"),
  phone: optionalPhoneField(),
});

export type UserFormValues = z.output<typeof userSchema>;
