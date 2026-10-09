import { z } from "zod";

// Field rules shared by the forms. They match what the services check
// (shared/validation tags), so most mistakes show up while typing; the
// services check everything again.

/** A first or last name: trimmed, 1–100 characters. */
export const nameField = (required: string) =>
  z.string({ error: required }).trim().min(1, required).max(100, "At most 100 characters");

/** An email address, trimmed. */
export const emailField = (required = "Enter your email address") =>
  z.string({ error: required }).trim().pipe(z.email("Enter a valid email address"));

/** A new password: 8–72 characters (bcrypt ignores anything after 72). */
export const newPasswordField = () =>
  z
    .string({ error: "Choose a password" })
    .min(8, "Use at least 8 characters")
    .max(72, "Use at most 72 characters");

/** An optional phone number in international (E.164) format; empty is undefined. */
export const optionalPhoneField = () =>
  z
    .string()
    .trim()
    .regex(/^(\+[1-9][0-9]{1,14})?$/, "Use the international format, e.g. +31612345678")
    .optional()
    .transform((phone) => phone || undefined);
