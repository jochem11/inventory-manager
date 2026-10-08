import "./userFormFields.scss";
import { TextField } from "~/components/form";

/** The editable profile fields; use inside a form with `userSchema`. */
export function UserFormFields() {
  return (
    <>
      <div class="user-form-fields__row">
        <TextField name="firstName" label="First name" />
        <TextField name="lastName" label="Last name" />
      </div>
      <TextField name="phone" type="tel" label="Phone" placeholder="+31612345678" hint="Optional" />
    </>
  );
}
