import { GRAPHQL_URL } from "~/constants/api";

/** `extensions.code` values the gateway returns (plus NETWORK for no response). */
export type ErrorCode =
  | "BAD_USER_INPUT"
  | "UNAUTHENTICATED"
  | "FORBIDDEN"
  | "FAILED_PRECONDITION"
  | "NOT_FOUND"
  | "CONFLICT"
  | "INTERNAL"
  | "NETWORK";

/** The first error of a GraphQL response, or a failure to reach the gateway. */
export class GraphQLRequestError extends Error {
  constructor(
    message: string,
    readonly code?: ErrorCode,
    /** Per-field messages for BAD_USER_INPUT, keyed like the input fields. */
    readonly fields?: Record<string, string>,
  ) {
    super(message);
    this.name = "GraphQLRequestError";
  }
}

export const hasCode = (error: unknown, code: ErrorCode) =>
  error instanceof GraphQLRequestError && error.code === code;

/** A message to show a user for any error a request can throw. */
export const errorMessage = (error: unknown) => {
  if (hasCode(error, "NETWORK")) return "Can't reach the server. Try again in a moment.";
  if (hasCode(error, "INTERNAL")) return "Something went wrong on our side. Try again in a moment.";
  const message = error instanceof Error ? error.message : String(error);
  return message.charAt(0).toUpperCase() + message.slice(1);
};

type GraphQLResponse<T> = {
  data?: T | null;
  errors?: { message: string; extensions?: { code?: ErrorCode; fields?: Record<string, string> } }[];
};

/**
 * Sends one GraphQL operation to the gateway and returns its `data`, or
 * throws a `GraphQLRequestError` for the first error. Cookies are included,
 * so the gateway can read and set the refresh token cookie.
 */
export async function gqlRequest<T>(
  query: string,
  variables?: Record<string, unknown>,
  accessToken?: string,
): Promise<T> {
  let response: Response;
  try {
    response = await fetch(GRAPHQL_URL, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      },
      body: JSON.stringify({ query, variables }),
    });
  } catch {
    throw new GraphQLRequestError("can't reach the server", "NETWORK");
  }

  let body: GraphQLResponse<T>;
  try {
    body = await response.json();
  } catch {
    throw new GraphQLRequestError(`the server responded with ${response.status}`, "NETWORK");
  }

  const error = body.errors?.[0];
  if (error) {
    throw new GraphQLRequestError(error.message, error.extensions?.code, error.extensions?.fields);
  }
  if (!body.data) {
    throw new GraphQLRequestError(`the server responded with ${response.status}`, "INTERNAL");
  }
  return body.data;
}
