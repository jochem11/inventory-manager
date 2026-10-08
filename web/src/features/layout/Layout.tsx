import "./layout.scss";
import { useLocation, useNavigate, useSearchParams } from "@solidjs/router";
import {
  createEffect,
  createSignal,
  on,
  onCleanup,
  onMount,
  Suspense,
  type ParentProps,
} from "solid-js";
import { Avatar, Breadcrumbs, Button, DropdownMenu, Icon, SearchBox } from "~/components/ui";
import { ConfirmDialog } from "~/components/dialogs";
import { APP_NAME, APP_VERSION } from "~/constants/app";
import { findNavTrail, visibleNavigation } from "~/constants/navigation";
import { useAuth } from "~/context";
import { createDialog, createMediaQuery } from "~/hooks";
import { Sidebar } from "./components/sidebar";
import { ThemeToggle } from "./components/themeToggle";
import { Topbar } from "./components/topbar";

const SIDEBAR_ID = "app-sidebar";
const COLLAPSED_STORAGE_KEY = "sidebar-collapsed";

/**
 * The app shell every route renders inside (it's the router's `root`). It
 * wires the generic navigation components to this app: its nav config,
 * search, account and theme. On narrow screens the sidebar becomes a drawer.
 */
export function Layout(props: ParentProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const confirm = createDialog(ConfirmDialog);
  const auth = useAuth();
  // The profile can be missing for a moment right after registering.
  const userName = () => {
    const user = auth.user();
    return user ? `${user.firstName} ${user.lastName}` : "Account";
  };

  // Keep in sync with $drawer-breakpoint in layout.scss.
  const isDrawer = createMediaQuery("(width < 56rem)");
  const [collapsed, setCollapsed] = createSignal(false);
  const [navOpen, setNavOpen] = createSignal(false);
  // Transitions stay off until the saved sidebar state is applied, so a
  // collapsed sidebar doesn't visibly animate shut on every page load.
  const [ready, setReady] = createSignal(false);

  onMount(() => {
    try {
      setCollapsed(localStorage.getItem(COLLAPSED_STORAGE_KEY) === "true");
    } catch {
      // Storage unavailable (private mode etc.) — keep the default.
    }
    requestAnimationFrame(() => setReady(true));

    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setNavOpen(false);
    };
    document.addEventListener("keydown", onKeyDown);
    onCleanup(() => document.removeEventListener("keydown", onKeyDown));
  });

  const toggleCollapsed = () => {
    const next = !collapsed();
    setCollapsed(next);
    try {
      localStorage.setItem(COLLAPSED_STORAGE_KEY, String(next));
    } catch {
      // Not persisted; fine.
    }
  };

  // Close the mobile drawer once a link in it has been followed.
  createEffect(on(() => location.pathname, () => setNavOpen(false), { defer: true }));

  const search = (query: string) =>
    navigate(query ? `/?q=${encodeURIComponent(query)}` : "/");

  const logout = async () => {
    const ok = await confirm.open({
      title: "Log out?",
      message: "You'll need to sign in again to manage your inventory.",
      confirmLabel: "Log out",
    });
    if (!ok) return;
    await auth.logout();
    navigate("/login");
  };

  return (
    <div
      class="layout"
      classList={{
        "layout--collapsed": collapsed(),
        "layout--nav-open": navOpen(),
        "layout--ready": ready(),
      }}>
      <a href="#main" class="layout__skip">
        Skip to content
      </a>

      <div class="layout__sidebar">
        <Sidebar
          id={SIDEBAR_ID}
          brand={{ label: APP_NAME, href: "/", icon: "package" }}
          items={visibleNavigation(auth.allows)}
          footer={`v${APP_VERSION}`}
          // The drawer is always full width; collapsing only applies on desktop.
          collapsed={collapsed() && !isDrawer()}
          onToggleCollapsed={isDrawer() ? undefined : toggleCollapsed}
          onClose={isDrawer() ? () => setNavOpen(false) : undefined}
        />
      </div>
      <div class="layout__backdrop" aria-hidden="true" onClick={() => setNavOpen(false)} />

      <div class="layout__body">
        <Topbar
          start={
            <>
              <Button
                variant="ghost"
                iconOnly
                class="layout__menu"
                aria-label="Open navigation"
                aria-controls={SIDEBAR_ID}
                aria-expanded={navOpen()}
                onClick={() => setNavOpen(true)}>
                <Icon name="menu" />
              </Button>
              <Breadcrumbs items={findNavTrail(location.pathname)} />
            </>
          }
          end={
            <>
              <SearchBox
                class="layout__search"
                placeholder="Search items…"
                value={typeof params.q === "string" ? params.q : ""}
                onSearch={search}
                shortcut
              />
              <ThemeToggle />
              <DropdownMenu
                class="layout__account"
                label={`Account menu for ${userName()}`}
                trigger={
                  <>
                    <Avatar name={userName()} />
                    <span class="layout__account-name">{userName()}</span>
                  </>
                }
                header={
                  <div class="layout__account-header">
                    <Avatar name={userName()} size="lg" />
                    <div class="layout__account-who">
                      <strong>{userName()}</strong>
                      <span>{auth.user()?.email}</span>
                    </div>
                  </div>
                }
                items={[
                  { label: "Profile", icon: "user", href: "/settings/profile" },
                  { label: "Preferences", icon: "sliders", href: "/settings/preferences" },
                  { divider: true },
                  { label: "Log out", icon: "log-out", onSelect: logout, danger: true },
                ]}
              />
            </>
          }
        />
        <main id="main" class="layout__main" tabindex="-1">
          <Suspense>{props.children}</Suspense>
        </main>
      </div>
    </div>
  );
}
