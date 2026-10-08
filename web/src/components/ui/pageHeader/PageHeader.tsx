import "./pageHeader.scss";
import { children, createEffect, Show, type JSX } from "solid-js";
import { APP_NAME } from "~/constants/app";

export type PageHeaderProps = {
  title: string;
  description?: string;
  /** Right-aligned controls (typically buttons). */
  actions?: JSX.Element;
};

/** Title block at the top of a page; also sets the browser tab title. */
export function PageHeader(props: PageHeaderProps) {
  const actions = children(() => props.actions);

  createEffect(() => {
    document.title = `${props.title} · ${APP_NAME}`;
  });

  return (
    <header class="page-header">
      <div class="page-header__text">
        <h1 class="page-header__title">{props.title}</h1>
        <Show when={props.description}>
          <p class="page-header__description">{props.description}</p>
        </Show>
      </div>
      <Show when={actions()}>
        <div class="page-header__actions">{actions()}</div>
      </Show>
    </header>
  );
}
