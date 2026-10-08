// Package auth is the contract for access tokens between the auth-service,
// which issues them, and everything that verifies them.
package auth

import (
	"slices"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// IssuerName and Audience are checked by everyone who verifies a token.
	IssuerName = "auth-service"
	Audience   = "inventory-manager"
	// Algorithm is EdDSA (Ed25519): verifiers only need the public key.
	Algorithm = "EdDSA"
)

// Claims are an access token's contents. Subject (sub) is the user id.
type Claims struct {
	jwt.RegisteredClaims
	// SessionID is the session the token was issued for.
	SessionID   string   `json:"sid"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// HasPermission reports whether the token grants permission, e.g. "users:write".
func (c *Claims) HasPermission(permission string) bool {
	return slices.Contains(c.Permissions, permission)
}

// HasAnyRole reports whether the token has at least one of roles.
func (c *Claims) HasAnyRole(roles ...string) bool {
	return slices.ContainsFunc(roles, func(role string) bool { return slices.Contains(c.Roles, role) })
}
