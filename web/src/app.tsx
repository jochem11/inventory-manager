import { Router, useLocation } from "@solidjs/router";
import { Show, type ParentProps } from "solid-js";
import { routes } from "./routes";
import "./app.scss";
import { AuthLayout, PUBLIC_PATHS, RequireAuth } from "./features/auth";
import { Layout } from "./features/layout";
import { AuthProvider, DialogProvider, ThemeProvider } from "./providers";

/**
 * Picks the shell for the current page: the logged-out pages get the plain
 * AuthLayout, everything else needs a login and gets the app's Layout.
 */
function Root(props: ParentProps) {
  const location = useLocation();
  return (
    <Show
      when={PUBLIC_PATHS.includes(location.pathname)}
      fallback={
        <RequireAuth>
          <Layout>{props.children}</Layout>
        </RequireAuth>
      }>
      <AuthLayout>{props.children}</AuthLayout>
    </Show>
  );
}

export default function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <DialogProvider>
          <Router root={Root}>{routes}</Router>
        </DialogProvider>
      </AuthProvider>
    </ThemeProvider>
  );
}
