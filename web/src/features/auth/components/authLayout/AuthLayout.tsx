import "./authLayout.scss";
import { A } from "@solidjs/router";
import { children, Show, Suspense, type JSX, type ParentProps } from "solid-js";
import { Icon } from "~/components/ui";
import { APP_NAME } from "~/constants/app";
import { ThemeToggle } from "~/features/layout/components/themeToggle";

/** The shell of the pages you see logged out: login, register, verify email. */
export function AuthLayout(props: ParentProps) {
  return (
    <div class="auth-layout">
      <header class="auth-layout__top">
        <A href="/login" class="auth-layout__brand">
          <Icon name="package" />
          <span>{APP_NAME}</span>
        </A>
        <ThemeToggle />
      </header>
      <main class="auth-layout__main">
        <Suspense>{props.children}</Suspense>
      </main>
    </div>
  );
}

export type AuthCardProps = {
  title: string;
  description?: JSX.Element;
  children: JSX.Element;
  /** Under the card, e.g. "No account yet? Create one". */
  footer?: JSX.Element;
};

/** The card each logged-out page puts its form in. */
export function AuthCard(props: AuthCardProps) {
  // Resolved once: reading a JSX prop twice would render it twice.
  const description = children(() => props.description);
  const footer = children(() => props.footer);

  return (
    <section class="auth-card" aria-labelledby="auth-card-title">
      <h1 id="auth-card-title" class="auth-card__title">
        {props.title}
      </h1>
      <Show when={description()}>
        <p class="auth-card__description">{description()}</p>
      </Show>
      {props.children}
      <Show when={footer()}>
        <p class="auth-card__footer">{footer()}</p>
      </Show>
    </section>
  );
}

