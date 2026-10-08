package model

import "time"

// AuthPayload is bound in gqlgen.yml instead of generated: it carries the
// user id, so the schema's user field gets its own resolver.
type AuthPayload struct {
	AccessToken          string    `json:"accessToken"`
	AccessTokenExpiresAt time.Time `json:"accessTokenExpiresAt"`
	UserID               string    `json:"-"`
	Roles                []string  `json:"roles"`
	Permissions          []string  `json:"permissions"`
}
