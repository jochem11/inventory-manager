// Package auth handles logins in the gateway: it verifies access tokens on
// every request, keeps the refresh token in a cookie, and resolves the auth
// queries and mutations by calling the auth-service.
package auth

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	authpb "github.com/jochem11/inventory-manager/services/graphql-gateway/pkg/pb/auth"
	"github.com/jochem11/inventory-manager/shared/auth"
)

// Claims are a verified access token's contents.
type Claims = auth.Claims

// keyRefetchInterval limits how often an unknown key id makes the verifier
// ask the auth-service for its keys, so made-up tokens can't flood it.
const keyRefetchInterval = 10 * time.Second

// Verifier checks access tokens with the auth-service's public keys. It
// fetches the keys when it sees a key id it doesn't know, which also picks up
// a new key after the auth-service restarts or rotates.
type Verifier struct {
	client authpb.AuthServiceClient

	mu        sync.Mutex
	keys      map[string]ed25519.PublicKey
	lastFetch time.Time
}

func NewVerifier(client authpb.AuthServiceClient) *Verifier {
	return &Verifier{client: client, keys: map[string]ed25519.PublicKey{}}
}

// Verify checks the signature, issuer, audience and expiry of raw.
func (v *Verifier) Verify(ctx context.Context, raw string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(raw, claims,
		func(t *jwt.Token) (any, error) {
			keyID, _ := t.Header["kid"].(string)
			return v.key(ctx, keyID)
		},
		jwt.WithValidMethods([]string{auth.Algorithm}),
		jwt.WithIssuer(auth.IssuerName),
		jwt.WithAudience(auth.Audience),
		jwt.WithExpirationRequired(),
		// Reject base64 variants of a token (e.g. different padding bits in
		// the last character), so each token has exactly one valid spelling.
		jwt.WithStrictDecoding(),
	)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func (v *Verifier) key(ctx context.Context, keyID string) (ed25519.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if key, ok := v.keys[keyID]; ok {
		return key, nil
	}
	if time.Since(v.lastFetch) < keyRefetchInterval {
		return nil, fmt.Errorf("unknown signing key %q", keyID)
	}
	v.lastFetch = time.Now()

	resp, err := v.client.GetPublicKeys(ctx, &authpb.GetPublicKeysRequest{})
	if err != nil {
		return nil, fmt.Errorf("fetch signing keys: %w", err)
	}
	keys := map[string]ed25519.PublicKey{}
	for _, k := range resp.GetKeys() {
		parsed, err := x509.ParsePKIXPublicKey(k.GetPublicKey())
		if err != nil {
			return nil, fmt.Errorf("parse signing key %q: %w", k.GetKeyId(), err)
		}
		if key, ok := parsed.(ed25519.PublicKey); ok && k.GetAlgorithm() == auth.Algorithm {
			keys[k.GetKeyId()] = key
		}
	}
	v.keys = keys

	if key, ok := v.keys[keyID]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("unknown signing key %q", keyID)
}
