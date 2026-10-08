import { z } from "zod";

/** The editable profile fields, with the same rules the user-service checks. */
export const userSchema = z.object({
  firstName: z.string({ error: "Enter a first name" }).trim().min(1, "Enter a first name").max(100, "At most 100 characters"),
  lastName: z.string({ error: "Enter a last name" }).trim().min(1, "Enter a last name").max(100, "At most 100 characters"),
  phone: z
    .string()
    .trim()
    .regex(/^(\+[1-9][0-9]{1,14})?$/, "Use the international format, e.g. +31612345678")
    .optional()
    .transform((phone) => phone || undefined),
});

export type UserFormValues = z.output<typeof userSchema>;
