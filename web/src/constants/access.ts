/**
 * Roles and permissions, as in shared/auth/permissions.go. Prefer checking
 * permissions: a role is a named set of permissions, and roles may be
 * regrouped without touching the checks.
 *
 * These checks only decide what to show. The gateway checks every request
 * again (@hasPermission / @hasRole) and answers 403 FORBIDDEN.
 */
export const ROLES = {
  user: "user",
  admin: "admin",
} as const;

export const PERMISSIONS = {
  usersRead: "users:read",
  usersWrite: "users:write",
  itemsRead: "items:read",
  itemsWrite: "items:write",
} as const;

export type Role = (typeof ROLES)[keyof typeof ROLES];
export type Permission = (typeof PERMISSIONS)[keyof typeof PERMISSIONS];

/**
 * What something needs. Every listed permission is required; of the roles,
 * one is enough. Leave both out for "any logged-in user".
 */
export type Access = {
  permission?: Permission | Permission[];
  role?: Role | Role[];
};
