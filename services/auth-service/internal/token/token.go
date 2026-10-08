// Package token issues access tokens: short-lived JWTs signed with Ed25519
// (alg EdDSA). Other services verify them with the public key from
// GetPublicKeys, so only the auth-service can create them.
package token

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jochem11/inventory-manager/shared/auth"
	"github.com/segmentio/ksuid"
)

// AccessTTL is short because an access token can't be revoked: logging out
// only stops the refresh token.
const AccessTTL = 15 * time.Minute

// Algorithm is the JWT "alg" of every token issued.
const Algorithm = auth.Algorithm

// Issuer signs access tokens.
type Issuer struct {
	key   ed25519.PrivateKey
	keyID string
	now   func() time.Time
}

// NewIssuer loads the signing key from pemKey, a PKCS#8 PEM block. An empty
// pemKey generates a fresh key (generated is true). That's fine in
// development: after a restart, older access tokens stop verifying and
// clients get new ones with their refresh token.
func NewIssuer(pemKey string) (issuer *Issuer, generated bool, err error) {
	var key ed25519.PrivateKey
	if pemKey == "" {
		if _, key, err = ed25519.GenerateKey(rand.Reader); err != nil {
			return nil, false, fmt.Errorf("generate signing key: %w", err)
		}
		generated = true
	} else {
		block, _ := pem.Decode([]byte(pemKey))
		if block == nil {
			return nil, false, errors.New("signing key: no PEM block")
		}
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, false, fmt.Errorf("signing key: %w", err)
		}
		var ok bool
		if key, ok = parsed.(ed25519.PrivateKey); !ok {
			return nil, false, errors.New("signing key: not an Ed25519 key")
		}
	}

	// The key id is derived from the public key, so it changes with the key
	// and verifiers know when to fetch the new one.
	sum := sha256.Sum256(key.Public().(ed25519.PublicKey))
	keyID := base64.RawURLEncoding.EncodeToString(sum[:])[:16]
	return &Issuer{key: key, keyID: keyID, now: time.Now}, generated, nil
}

// Issue signs an access token for a user's session.
func (i *Issuer) Issue(userID, sessionID, email string, roles, permissions []string) (string, time.Time, error) {
	now := i.now()
	expiresAt := now.Add(AccessTTL)
	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        ksuid.New().String(),
			Issuer:    auth.IssuerName,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{auth.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		SessionID:   sessionID,
		Email:       email,
		Roles:       roles,
		Permissions: permissions,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	t.Header["kid"] = i.keyID
	signed, err := t.SignedString(i.key)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, expiresAt, nil
}

// PublicKey returns the key id and the PKIX DER encoding of the public key.
func (i *Issuer) PublicKey() (keyID string, der []byte, err error) {
	der, err = x509.MarshalPKIXPublicKey(i.key.Public())
	if err != nil {
		return "", nil, fmt.Errorf("encode public key: %w", err)
	}
	return i.keyID, der, nil
}
