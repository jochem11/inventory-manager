import type { IconName } from "~/components/ui";
import type { Access } from "~/constants/access";

export type NavLink = {
  label: string;
  href: string;
  icon?: IconName;
  /** Only mark active on an exact match (needed for `/`). */
  end?: boolean;
  /** Hide the link from users without this access. */
  access?: Access;
};

/** A collapsible section holding its own links. */
export type NavGroup = {
  label: string;
  icon: IconName;
  children: NavLink[];
};

export type NavItem = NavLink | NavGroup;

export const isNavGroup = (item: NavItem): item is NavGroup => "children" in item;

export type SidebarBrand = {
  label: string;
  href: string;
  icon: IconName;
};
