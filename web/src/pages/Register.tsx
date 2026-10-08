import { A } from "@solidjs/router";
import { createSignal, Show } from "solid-js";
import { z } from "zod";
import { register } from "~/api/auth";
import { errorMessage, GraphQLRequestError, hasCode } from "~/api/graphql";
import { Form, SubmitButton, TextField } from "~/components/form";
import { AuthCard, AuthNotice, ResendVerification } from "~/features/auth";

// The same rules the auth-service checks, so most mistakes show up while typing.
const schema = z.object({
  firstName: z.string({ error: "Enter your first name" }).trim().min(1, "Enter your first name").max(100, "At most 100 characters"),
  lastName: z.string({ error: "Enter your last name" }).trim().min(1, "Enter your last name").max(100, "At most 100 characters"),
  email: z.string({ error: "Enter your email address" }).trim().pipe(z.email("Enter a valid email address")),
  password: z
    .string({ error: "Choose a password" })
    .min(8, "Use at least 8 characters")
    .max(72, "Use at most 72 characters"),
  phone: z
    .string()
    .trim()
    .regex(/^(\+[1-9][0-9]{1,14})?$/, "Use the international format, e.g. +31612345678")
    .optional()
    .transform((phone) => phone || undefined),
});

export default function Register() {
  const [registeredEmail, setRegisteredEmail] = createSignal<string>();
  const [error, setError] = createSignal<string>();

  const submit = async (values: z.output<typeof schema>) => {
    setError();
    try {
      await register(values);
      setRegisteredEmail(values.email);
    } catch (e) {
      if (hasCode(e, "CONFLICT")) {
        setError("An account with this email already exists");
      } else if (hasCode(e, "BAD_USER_INPUT") && e instanceof GraphQLRequestError && e.fields) {
        setError(Object.values(e.fields).join(". "));
      } else {
        setError(errorMessage(e));
      }
    }
  };

  return (
    <Show
      when={registeredEmail()}
      fallback={
        <AuthCard
          title="Create an account"
          description="We'll email you a link to activate it."
          footer={
            <>
              Already have an account? <A href="/login">Log in</A>
            </>
          }>
          <Form schema={schema} onSubmit={submit}>
            <Show when={error()}>{(message) => <AuthNotice tone="error">{message()}.</AuthNotice>}</Show>
            <div class="auth-form-row">
              <TextField name="firstName" label="First name" autocomplete="given-name" />
              <TextField name="lastName" label="Last name" autocomplete="family-name" />
            </div>
            <TextField name="email" type="email" label="Email" autocomplete="email" />
            <TextField
              name="password"
              type="password"
              label="Password"
              autocomplete="new-password"
              hint="At least 8 characters."
            />
            <TextField
              name="phone"
              type="tel"
              label="Phone (optional)"
              autocomplete="tel"
              placeholder="+31612345678"
            />
            <SubmitButton block submittingText="Creating account…">
              Create account
            </SubmitButton>
          </Form>
        </AuthCard>
      }>
      {(email) => (
        <AuthCard
          title="Check your email"
          description={<>We sent an activation link to {email()}. Open it to activate your account, then log in.</>}
          footer={<A href="/login">Back to log in</A>}>
          <AuthNotice>
            <p>No email after a few minutes? Check your spam folder, or:</p>
            <ResendVerification email={email()} />
          </AuthNotice>
        </AuthCard>
      )}
    </Show>
  );
}
