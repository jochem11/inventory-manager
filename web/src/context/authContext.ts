import { createContext, useContext, type Accessor } from "solid-js";
import type { AuthUser } from "~/api/auth";
import type { Access, Permission, Role } from "~/constants/access";
import { ContextError } from "~/errors/contextError";

/**
 * `loading` until the first refresh attempt answers whether there's a
 * session (and always on the server, which can't see the cookie's session).
 */
export type AuthStatus = "loading" | "authenticated" | "anonymous";

export type AuthContextValue = {
  status: Accessor<AuthStatus>;
  /** The logged-in user's profile; undefined when logged out or not created yet. */
  user: Accessor<AuthUser | undefined>;
  /** The roles and permissions in the access token; empty when logged out. */
  roles: Accessor<string[]>;
  permissions: Accessor<string[]>;
  /** Whether the user has at least one of roles. */
  hasRole: (...roles: Role[]) => boolean;
  /** Whether the user has every one of permissions. */
  hasPermission: (...permissions: Permission[]) => boolean;
  /** Whether the user meets access: all its permissions and one of its roles. */
  allows: (access?: Access) => boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  /**
   * A GraphQL request as the logged-in user. When the access token turns out
   * to be expired, it refreshes once and retries.
   */
  request: <T>(query: string, variables?: Record<string, unknown>) => Promise<T>;
};

export const AuthContext = createContext<AuthContextValue>();

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new ContextError("useAuth must be used inside <AuthProvider>");
  return ctx;
};
