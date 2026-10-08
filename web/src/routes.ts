import { lazy } from "solid-js";
import type { RouteDefinition } from "@solidjs/router";

export const routes: RouteDefinition[] = [
  // Logged out (see PUBLIC_PATHS in features/auth).
  {
    path: "/login",
    component: lazy(() => import("./pages/Login")),
  },
  {
    path: "/register",
    component: lazy(() => import("./pages/Register")),
  },
  {
    path: "/verify-email",
    component: lazy(() => import("./pages/VerifyEmail")),
  },
  // Logged in.
  {
    path: "/",
    component: lazy(() => import("./pages/Home")),
  },
  {
    // Sections that exist in the navigation but aren't built yet.
    path: ["/categories", "/reports", "/settings/profile", "/settings/preferences"],
    component: lazy(() => import("./pages/ComingSoon")),
  },
  {
    path: "/locations",
    component: lazy(() => import("./pages/Locations")),
  },
  {
    path: "/users",
    component: lazy(() => import("./pages/Users")),
  },
  {
    path: "/about",
    component: lazy(() => import("./pages/About")),
  },
  {
    path: "*404",
    component: lazy(() => import("./pages/NotFound")),
  },
];
