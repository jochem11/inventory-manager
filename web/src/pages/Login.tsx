import { A, useNavigate, useSearchParams } from "@solidjs/router";
import { createEffect, createSignal, Show } from "solid-js";
import { z } from "zod";
import { errorMessage, hasCode } from "~/api/graphql";
import { Form, SubmitButton, TextField } from "~/components/form";
import { useAuth } from "~/context";
import { AuthCard, AuthNotice, ResendVerification, safeRedirect } from "~/features/auth";

const schema = z.object({
  email: z.string({ error: "Enter your email address" }).trim().pipe(z.email("Enter a valid email address")),
  password: z.string({ error: "Enter your password" }).min(1, "Enter your password"),
});

export default function Login() {
  const auth = useAuth();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const [error, setError] = createSignal<string>();
  const [unverifiedEmail, setUnverifiedEmail] = createSignal<string>();

  // Logged in (now, or already when the page opened): go where you were going.
  createEffect(() => {
    if (auth.status() === "authenticated") navigate(safeRedirect(params.redirect), { replace: true });
  });

  const submit = async (values: z.output<typeof schema>) => {
    setError();
    setUnverifiedEmail();
    try {
      await auth.login(values.email, values.password);
    } catch (e) {
      if (hasCode(e, "FAILED_PRECONDITION")) setUnverifiedEmail(values.email);
      else setError(errorMessage(e));
    }
  };

  return (
    <AuthCard
      title="Log in"
      description="Welcome back. Log in to manage your inventory."
      footer={
        <>
          No account yet? <A href="/register">Create one</A>
        </>
      }>
      <Form schema={schema} onSubmit={submit}>
        <Show when={error()}>{(message) => <AuthNotice tone="error">{message()}.</AuthNotice>}</Show>
        <Show when={unverifiedEmail()}>
          {(email) => (
            <AuthNotice>
              <p>Your email isn't verified yet. Open the activation link we sent to {email()}.</p>
              <ResendVerification email={email()} />
            </AuthNotice>
          )}
        </Show>
        <TextField name="email" type="email" label="Email" autocomplete="email" />
        <TextField name="password" type="password" label="Password" autocomplete="current-password" />
        <SubmitButton block submittingText="Logging in…">
          Log in
        </SubmitButton>
      </Form>
    </AuthCard>
  );
}
