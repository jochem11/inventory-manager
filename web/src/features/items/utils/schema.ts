import { z } from "zod";

const optionalText = (max: number, message: string) =>
  z
    .string()
    .trim()
    .max(max, message)
    .optional()
    .transform((value) => value || undefined);

/** An item's fields, with the rules the item-service checks. */
export const itemSchema = z.object({
  name: z.string({ error: "Give the item a name" }).trim().min(1, "Give the item a name").max(200, "At most 200 characters"),
  description: optionalText(2000, "At most 2000 characters"),
  imageUrl: optionalText(2048, "At most 2048 characters").refine(
    (url) => !url || /^https?:\/\/\S+$/.test(url),
    "Use a web address starting with http:// or https://",
  ),
  categoryId: z.string({ error: "Pick a category" }).min(1, "Pick a category"),
  statusId: z.string({ error: "Pick a status" }).min(1, "Pick a status"),
});
