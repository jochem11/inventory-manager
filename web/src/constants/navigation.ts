import type { Crumb } from "~/components/ui";
import { PERMISSIONS, type Access } from "~/constants/access";
import { isNavGroup, type NavItem } from "~/types/navigation";

/** The sidebar, top to bottom. Breadcrumbs and page titles are derived from it too. */
export const NAVIGATION: NavItem[] = [
  {
    label: "Inventory",
    icon: "package",
    children: [
      { label: "All items", href: "/", end: true },
      { label: "Categories", href: "/categories" },
      { label: "Locations", href: "/locations" },
    ],
  },
  { label: "Reports", href: "/reports", icon: "bar-chart" },
  { label: "Users", href: "/users", icon: "user", access: { permission: PERMISSIONS.usersRead } },
  {
    label: "Settings",
    icon: "sliders",
    children: [
      { label: "Profile", href: "/settings/profile" },
      { label: "Preferences", href: "/settings/preferences" },
    ],
  },
  { label: "About", href: "/about", icon: "info" },
];

/**
 * The navigation without the links that `allows` rejects, and without groups
 * left empty. Pass `useAuth().allows`.
 */
export function visibleNavigation(allows: (access?: Access) => boolean, items = NAVIGATION): NavItem[] {
  return items.flatMap((item): NavItem[] => {
    if (!isNavGroup(item)) return !item.access || allows(item.access) ? [item] : [];
    const children = item.children.filter((child) => !child.access || allows(child.access));
    return children.length > 0 ? [{ ...item, children }] : [];
  });
}

const normalize = (path: string) => (path.length > 1 ? path.replace(/\/+$/, "") : path);

/**
 * The chain of nav entries leading to `pathname` — e.g. `/categories` gives
 * `[Inventory, Categories]`. Empty when the path isn't in the navigation.
 */
export function findNavTrail(pathname: string): Crumb[] {
  const path = normalize(pathname);

  for (const item of NAVIGATION) {
    if (!isNavGroup(item)) {
      if (item.href === path) return [{ label: item.label, href: item.href }];
      continue;
    }
    const child = item.children.find((c) => c.href === path);
    if (child) return [{ label: item.label }, { label: child.label, href: child.href }];
  }

  return [];
}
