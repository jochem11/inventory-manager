import { gqlRequest } from "./graphql";

// ---------------------------------------------------------------------------
// Auth API. The access token is returned here and kept in memory by
// AuthProvider; the refresh token never reaches JavaScript: the gateway keeps
// it in an httpOnly cookie, which `gqlRequest` sends along.
// ---------------------------------------------------------------------------

export type AuthUser = {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
  avatarUrl?: string | null;
};

export type AuthPayload = {
  accessToken: string;
  /** ISO timestamp; refresh before it. */
  accessTokenExpiresAt: string;
  /** Null if the profile doesn't exist (yet). */
  user: AuthUser | null;
  /** From the access token, e.g. ["user", "admin"]. */
  roles: string[];
  /** From the access token, e.g. ["users:read"]. */
  permissions: string[];
};

export type RegisterInput = {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  phone?: string;
};

const USER_FIELDS = "id firstName lastName email avatarUrl";
const AUTH_PAYLOAD_FIELDS = `accessToken accessTokenExpiresAt roles permissions user { ${USER_FIELDS} }`;

export const login = async (email: string, password: string) =>
  (
    await gqlRequest<{ login: AuthPayload }>(
      `mutation Login($email: String!, $password: String!) {
        login(email: $email, password: $password) { ${AUTH_PAYLOAD_FIELDS} }
      }`,
      { email, password },
    )
  ).login;

/** New tokens from the refresh token cookie; UNAUTHENTICATED when logged out. */
export const refreshToken = async () =>
  (
    await gqlRequest<{ refreshToken: AuthPayload }>(
      `mutation RefreshToken { refreshToken { ${AUTH_PAYLOAD_FIELDS} } }`,
    )
  ).refreshToken;

export const logout = async () => {
  await gqlRequest<{ logout: boolean }>(`mutation Logout { logout }`);
};

/** Returns the new user's id. The account works once the email is verified. */
export const register = async (input: RegisterInput) =>
  (
    await gqlRequest<{ register: string }>(
      `mutation Register($input: RegisterInput!) { register(input: $input) }`,
      { input },
    )
  ).register;

export const verifyEmail = async (token: string) => {
  await gqlRequest<{ verifyEmail: boolean }>(
    `mutation VerifyEmail($token: String!) { verifyEmail(token: $token) }`,
    { token },
  );
};

export const resendVerification = async (email: string) => {
  await gqlRequest<{ resendVerification: boolean }>(
    `mutation ResendVerification($email: String!) { resendVerification(email: $email) }`,
    { email },
  );
};
