/**
 * The GraphQL gateway. The browser calls it directly (CORS with credentials),
 * so it can store the httpOnly refresh token cookie the gateway sets.
 */
export const GRAPHQL_URL: string =
  import.meta.env.VITE_GRAPHQL_URL ?? "http://localhost:4000/graphql";
