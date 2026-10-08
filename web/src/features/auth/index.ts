export * from "./components/authLayout";
export * from "./components/guard";
export * from "./components/requireAuth";
export * from "./components/resendVerification";

/** Pages you can see without logging in; they use AuthLayout. */
export const PUBLIC_PATHS = ["/login", "/register", "/verify-email"];
