package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

// Middleware verifies the request's `Authorization: Bearer` token and puts
// its claims in the context. An invalid or expired token counts as no token:
// whatever needs a user then returns UNAUTHENTICATED, and the client
// refreshes. secureCookies sets Secure on the refresh cookie (needs HTTPS).
func Middleware(verifier *Verifier, secureCookies bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), exchangeKey{}, &exchange{w: w, r: r, secureCookies: secureCookies})

			if raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok && raw != "" {
				claims, err := verifier.Verify(ctx, raw)
				if err != nil {
					slog.DebugContext(ctx, "ignoring access token", "error", err)
				} else {
					ctx = context.WithValue(ctx, claimsKey{}, claims)
				}
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
