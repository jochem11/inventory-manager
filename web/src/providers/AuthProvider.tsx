import { createSignal, onCleanup, onMount, type ParentComponent } from "solid-js";
import * as authApi from "~/api/auth";
import type { AuthPayload, AuthUser } from "~/api/auth";
import { gqlRequest, hasCode } from "~/api/graphql";
import type { Access, Permission, Role } from "~/constants/access";
import { AuthContext, type AuthContextValue, type AuthStatus } from "~/context";

const asList = <T,>(value: T | T[] | undefined): T[] =>
  value === undefined ? [] : Array.isArray(value) ? value : [value];

/** Refresh this long before the access token expires. */
const REFRESH_MARGIN_MS = 60_000;
/** Another tab may have just rotated the refresh cookie: retry after this. */
const RACE_RETRY_MS = 500;

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

/**
 * Keeps the login. The access token lives only in memory (not in
 * localStorage, where any script could read it); the refresh token is an
 * httpOnly cookie the browser sends to the gateway. On page load, a refresh
 * restores the session from that cookie, and the token is refreshed again a
 * minute before it expires.
 */
export const AuthProvider: ParentComponent = (props) => {
  const [status, setStatus] = createSignal<AuthStatus>("loading");
  const [user, setUser] = createSignal<AuthUser>();
  const [roles, setRoles] = createSignal<string[]>([]);
  const [permissions, setPermissions] = createSignal<string[]>([]);
  let accessToken: string | undefined;
  let refreshTimer: ReturnType<typeof setTimeout> | undefined;
  let refreshing: Promise<boolean> | undefined;

  const scheduleRefresh = (expiresAt: string) => {
    clearTimeout(refreshTimer);
    const delay = Math.max(new Date(expiresAt).getTime() - Date.now() - REFRESH_MARGIN_MS, 5_000);
    refreshTimer = setTimeout(() => void refresh(), delay);
  };

  const startSession = (payload: AuthPayload) => {
    accessToken = payload.accessToken;
    setUser(payload.user ?? undefined);
    setRoles(payload.roles);
    setPermissions(payload.permissions);
    setStatus("authenticated");
    scheduleRefresh(payload.accessTokenExpiresAt);
  };

  const endSession = () => {
    accessToken = undefined;
    clearTimeout(refreshTimer);
    setUser(undefined);
    setRoles([]);
    setPermissions([]);
    setStatus("anonymous");
  };

  /**
   * Gets new tokens with the refresh cookie and reports whether there's a
   * session. Concurrent callers share one request: the cookie is rotated on
   * every refresh, so two at once would make one of them fail.
   */
  const refresh = (): Promise<boolean> => {
    refreshing ??= (async () => {
      for (let attempt = 1; ; attempt++) {
        try {
          startSession(await authApi.refreshToken());
          return true;
        } catch (error) {
          // Only an existing session can lose a race with another tab.
          if (hasCode(error, "UNAUTHENTICATED") && attempt === 1 && status() === "authenticated") {
            await sleep(RACE_RETRY_MS);
            continue;
          }
          // Logged out, or the gateway is unreachable: either way there's
          // no usable session right now.
          endSession();
          return false;
        }
      }
    })().finally(() => {
      refreshing = undefined;
    });
    return refreshing;
  };

  const hasRole = (...wanted: Role[]) => wanted.some((role) => roles().includes(role));
  const hasPermission = (...wanted: Permission[]) =>
    wanted.every((permission) => permissions().includes(permission));

  const value: AuthContextValue = {
    status,
    user,
    roles,
    permissions,
    hasRole,
    hasPermission,
    allows: (access?: Access) => {
      if (status() !== "authenticated") return false;
      const wantedRoles = asList(access?.role);
      return hasPermission(...asList(access?.permission)) && (wantedRoles.length === 0 || hasRole(...wantedRoles));
    },
    login: async (email, password) => {
      startSession(await authApi.login(email, password));
    },
    logout: async () => {
      try {
        await authApi.logout();
      } finally {
        endSession();
      }
    },
    request: async <T,>(query: string, variables?: Record<string, unknown>) => {
      try {
        return await gqlRequest<T>(query, variables, accessToken);
      } catch (error) {
        if (!hasCode(error, "UNAUTHENTICATED") || !(await refresh())) throw error;
        return gqlRequest<T>(query, variables, accessToken);
      }
    },
  };

  // Only in the browser: the server can't see the session.
  onMount(() => void refresh());
  onCleanup(() => clearTimeout(refreshTimer));

  return <AuthContext.Provider value={value}>{props.children}</AuthContext.Provider>;
};
