package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/jochem11/inventory-manager/shared/errs"
)

type claimsKey struct{}

// errUnauthenticated is what a field that needs a login answers without one:
// the client should refresh its token and retry.
var errUnauthenticated = errs.New(errs.Unauthenticated, "not logged in, or the access token expired")

// ClaimsFrom returns the verified access token of the request, if any.
func ClaimsFrom(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsKey{}).(*Claims)
	return claims, ok
}

// Require returns the request's claims, or UNAUTHENTICATED when there is no
// valid access token: the client should refresh it and retry.
func Require(ctx context.Context) (*Claims, error) {
	claims, ok := ClaimsFrom(ctx)
	if !ok {
		return nil, errUnauthenticated
	}
	return claims, nil
}

// RequirePermission is Require plus a check for permission, e.g. "users:read".
func RequirePermission(ctx context.Context, permission string) error {
	claims, err := Require(ctx)
	if err != nil {
		return err
	}
	if !claims.HasPermission(permission) {
		return errs.New(errs.PermissionDenied, "missing permission "+permission)
	}
	return nil
}

// RequireRole is Require plus a check for at least one of roles. Prefer
// RequirePermission: roles are groups of permissions and may be regrouped.
func RequireRole(ctx context.Context, roles ...string) error {
	claims, err := Require(ctx)
	if err != nil {
		return err
	}
	if !claims.HasAnyRole(roles...) {
		return errs.New(errs.PermissionDenied, "needs role "+strings.Join(roles, " or "))
	}
	return nil
}

// refreshCookie is the httpOnly cookie that holds the refresh token. Its path
// is /graphql, so the browser sends it nowhere else.
const refreshCookie = "refresh_token"

// exchange is the HTTP request and response behind a GraphQL request, so
// resolvers can read and set the refresh cookie.
type exchange struct {
	w             http.ResponseWriter
	r             *http.Request
	secureCookies bool
}

type exchangeKey struct{}

func refreshTokenFrom(ctx context.Context) string {
	ex, ok := ctx.Value(exchangeKey{}).(*exchange)
	if !ok {
		return ""
	}
	cookie, err := ex.r.Cookie(refreshCookie)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func setRefreshCookie(ctx context.Context, token string, expires time.Time) {
	ex, ok := ctx.Value(exchangeKey{}).(*exchange)
	if !ok {
		return
	}
	http.SetCookie(ex.w, &http.Cookie{
		Name:     refreshCookie,
		Value:    token,
		Path:     "/graphql",
		Expires:  expires,
		HttpOnly: true, // JavaScript can't read it, so an XSS bug can't steal it
		Secure:   ex.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearRefreshCookie(ctx context.Context) {
	ex, ok := ctx.Value(exchangeKey{}).(*exchange)
	if !ok {
		return
	}
	http.SetCookie(ex.w, &http.Cookie{
		Name:     refreshCookie,
		Value:    "",
		Path:     "/graphql",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   ex.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}
