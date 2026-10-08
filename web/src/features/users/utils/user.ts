import type { User } from "~/api/users";

export const fullName = (user: User) => `${user.firstName} ${user.lastName}`;

const dateFormat = new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" });

export const formatDate = (iso: string) => dateFormat.format(new Date(iso));
