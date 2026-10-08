import { createSignal, Match, Switch } from "solid-js";
import { resendVerification } from "~/api/auth";
import { errorMessage } from "~/api/graphql";
import { Button } from "~/components/ui";

/** "Send a new activation link" for an account that isn't verified yet. */
export function ResendVerification(props: { email: string }) {
  const [state, setState] = createSignal<"idle" | "sending" | "sent">("idle");
  const [error, setError] = createSignal<string>();

  const send = async () => {
    setState("sending");
    setError();
    try {
      await resendVerification(props.email);
      setState("sent");
    } catch (e) {
      setError(errorMessage(e));
      setState("idle");
    }
  };

  return (
    <Switch>
      <Match when={state() === "sent"}>
        <p>A new link is on its way to {props.email}.</p>
      </Match>
      <Match when={true}>
        <Button variant="link" size="sm" loading={state() === "sending"} onClick={send}>
          Send a new activation link
        </Button>
        {error() && <p>{error()}</p>}
      </Match>
    </Switch>
  );
}
