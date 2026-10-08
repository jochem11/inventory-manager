import { useLocation, useNavigate } from "@solidjs/router";
import { createEffect, Show, type ParentProps } from "solid-js";
import { useAuth } from "~/context";

/**
 * Renders its children only when logged in. While the session is being
 * restored it renders nothing; without a session it sends you to /login,
 * which brings you back here afterwards.
 */
export function RequireAuth(props: ParentProps) {
  const auth = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  createEffect(() => {
    if (auth.status() !== "anonymous") return;
    const here = location.pathname + location.search;
    navigate(here === "/" ? "/login" : `/login?redirect=${encodeURIComponent(here)}`, {
      replace: true,
    });
  });

  return <Show when={auth.status() === "authenticated"}>{props.children}</Show>;
}

/** Where to go after logging in: only paths on this site, never "//evil.com". */
export const safeRedirect = (value: unknown) =>
  typeof value === "string" && value.startsWith("/") && !value.startsWith("//") ? value : "/";
