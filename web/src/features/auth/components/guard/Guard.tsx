import { A } from "@solidjs/router";
import { Show, type JSX, type ParentProps } from "solid-js";
import { PageHeader } from "~/components/ui";
import type { Access } from "~/constants/access";
import { useAuth } from "~/context";

export type GuardProps = ParentProps<
  Access & {
    /** Shown instead of the children without access. Nothing by default. */
    fallback?: JSX.Element;
  }
>;

/**
 * Renders its children only when the user has the access: every listed
 * permission, and one of the listed roles.
 *
 *   <Guard permission="users:write"><Button>Delete</Button></Guard>
 *   <Guard role="admin" fallback={<Forbidden />}>…</Guard>
 *
 * It only hides things: the gateway checks every request again.
 */
export function Guard(props: GuardProps) {
  const auth = useAuth();
  return (
    <Show when={auth.allows({ permission: props.permission, role: props.role })} fallback={props.fallback}>
      {props.children}
    </Show>
  );
}

/** A page's content for users without access to it (like a 403). */
export function Forbidden() {
  return (
    <>
      <PageHeader title="No access" description="Your account doesn't have permission to see this page." />
      <A href="/">Back to your items</A>
    </>
  );
}
