import { A, useSearchParams } from "@solidjs/router";
import { createSignal, Match, onMount, Switch } from "solid-js";
import { z } from "zod";
import { resendVerification, verifyEmail } from "~/api/auth";
import { errorMessage, hasCode } from "~/api/graphql";
import { Form, SubmitButton, TextField } from "~/components/form";
import { AuthCard } from "~/features/auth";
import { Notice } from "~/components/ui";
import { emailField } from "~/schemas/fields";

const resendSchema = z.object({
  email: emailField(),
});

/** The page the activation link opens: /verify-email?token=… */
export default function VerifyEmail() {
  const [params] = useSearchParams();
  const [state, setState] = createSignal<"verifying" | "verified" | "invalid" | "error">("verifying");
  const [error, setError] = createSignal<string>();
  const [resentTo, setResentTo] = createSignal<string>();

  // In the browser only: a link must not be used up by server rendering.
  onMount(async () => {
    const token = typeof params.token === "string" ? params.token : "";
    if (!token) return setState("invalid");
    try {
      await verifyEmail(token);
      setState("verified");
    } catch (e) {
      if (hasCode(e, "UNAUTHENTICATED")) setState("invalid");
      else {
        setError(errorMessage(e));
        setState("error");
      }
    }
  });

  const resend = async (values: z.output<typeof resendSchema>) => {
    await resendVerification(values.email);
    setResentTo(values.email);
  };

  return (
    <Switch>
      <Match when={state() === "verifying"}>
        <AuthCard title="Verifying your email…" description="One moment.">
          <></>
        </AuthCard>
      </Match>
      <Match when={state() === "verified"}>
        <AuthCard title="Email verified" description="Your account is active. You can log in now.">
          <A href="/login" class="button button--primary button--md button--block">
            Log in
          </A>
        </AuthCard>
      </Match>
      <Match when={state() === "error"}>
        <AuthCard title="Couldn't verify your email" footer={<A href="/login">Back to log in</A>}>
          <Notice tone="error">{error()}.</Notice>
        </AuthCard>
      </Match>
      <Match when={state() === "invalid"}>
        <AuthCard
          title="This link doesn't work"
          description="It has expired, was already used, or a newer link replaced it. Get a new one:"
          footer={<A href="/login">Back to log in</A>}>
          <Switch>
            <Match when={resentTo()}>
              {(email) => (
                <Notice tone="success">
                  If {email()} has an account that isn't verified yet, a new link is on its way.
                </Notice>
              )}
            </Match>
            <Match when={true}>
              <Form schema={resendSchema} onSubmit={resend}>
                <TextField name="email" type="email" label="Email" autocomplete="email" />
                <SubmitButton block submittingText="Sending…">
                  Send a new link
                </SubmitButton>
              </Form>
            </Match>
          </Switch>
        </AuthCard>
      </Match>
    </Switch>
  );
}
