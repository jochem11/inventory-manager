import { A } from "@solidjs/router";
import { createSignal, Show } from "solid-js";
import { z } from "zod";
import { register } from "~/api/auth";
import { fieldErrorMessage, hasCode } from "~/api/graphql";
import { Form, SubmitButton, TextField } from "~/components/form";
import { AuthCard, ResendVerification } from "~/features/auth";
import { Notice } from "~/components/ui";
import { emailField, nameField, newPasswordField, optionalPhoneField } from "~/schemas/fields";

// The same rules the auth-service checks, so most mistakes show up while typing.
const schema = z.object({
  firstName: nameField("Enter your first name"),
  lastName: nameField("Enter your last name"),
  email: emailField(),
  password: newPasswordField(),
  phone: optionalPhoneField(),
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
      } else {
        setError(fieldErrorMessage(e));
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
            <Show when={error()}>{(message) => <Notice tone="error">{message()}.</Notice>}</Show>
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
          <Notice>
            <p>No email after a few minutes? Check your spam folder, or:</p>
            <ResendVerification email={email()} />
          </Notice>
        </AuthCard>
      )}
    </Show>
  );
}
