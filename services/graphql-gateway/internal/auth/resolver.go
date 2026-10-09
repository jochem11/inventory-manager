package auth

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/graph/model"
	authpb "github.com/jochem11/inventory-manager/services/graphql-gateway/pkg/pb/auth"
	"github.com/jochem11/inventory-manager/shared/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Resolver resolves the auth mutations by calling the auth-service. Like the
// other resolvers it returns gRPC errors as they are; gqlerr.Present maps them.
type Resolver struct {
	client authpb.AuthServiceClient
}

func NewResolver(client authpb.AuthServiceClient) *Resolver {
	return &Resolver{client: client}
}

// Register returns the new user's id.
func (r *Resolver) Register(ctx context.Context, input model.RegisterInput) (string, error) {
	resp, err := r.client.Register(ctx, &authpb.RegisterRequest{
		Email:     input.Email,
		Password:  input.Password,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Phone:     input.Phone,
	})
	if err != nil {
		return "", err
	}
	return resp.GetUserId(), nil
}

func (r *Resolver) VerifyEmail(ctx context.Context, token string) (bool, error) {
	if _, err := r.client.VerifyEmail(ctx, &authpb.VerifyEmailRequest{Token: token}); err != nil {
		return false, err
	}
	return true, nil
}

func (r *Resolver) ResendVerification(ctx context.Context, email string) (bool, error) {
	if _, err := r.client.ResendVerification(ctx, &authpb.ResendVerificationRequest{Email: email}); err != nil {
		return false, err
	}
	return true, nil
}

func (r *Resolver) Login(ctx context.Context, email, password string) (*model.AuthPayload, error) {
	resp, err := r.client.Login(ctx, &authpb.LoginRequest{Email: email, Password: password})
	if err != nil {
		return nil, err
	}
	return payload(ctx, resp.GetTokens())
}

// Refresh uses the refresh token cookie. When the session is over, the
// cookie is cleared, so the browser stops sending a dead token.
func (r *Resolver) Refresh(ctx context.Context) (*model.AuthPayload, error) {
	token := refreshTokenFrom(ctx)
	if token == "" {
		return nil, errs.New(errs.Unauthenticated, "not logged in")
	}
	resp, err := r.client.Refresh(ctx, &authpb.RefreshRequest{RefreshToken: token})
	if status.Code(err) == codes.Unauthenticated {
		clearRefreshCookie(ctx)
	}
	if err != nil {
		return nil, err
	}
	return payload(ctx, resp.GetTokens())
}

// Logout always succeeds and clears the cookie, even if the session was
// already gone.
func (r *Resolver) Logout(ctx context.Context) (bool, error) {
	if token := refreshTokenFrom(ctx); token != "" {
		if _, err := r.client.Logout(ctx, &authpb.LogoutRequest{RefreshToken: token}); err != nil {
			return false, err
		}
	}
	clearRefreshCookie(ctx)
	return true, nil
}

// payload sets the refresh cookie and returns the rest to the client, with
// the roles and permissions from the access token, so the web app can show
// only what the user may do.
func payload(ctx context.Context, t *authpb.Tokens) (*model.AuthPayload, error) {
	// The token comes straight from the auth-service, so it needn't be
	// verified here; every request that sends it back is verified.
	var claims Claims
	if _, _, err := jwt.NewParser().ParseUnverified(t.GetAccessToken(), &claims); err != nil {
		return nil, fmt.Errorf("read access token: %w", err)
	}
	setRefreshCookie(ctx, t.GetRefreshToken(), t.GetRefreshTokenExpiresAt().AsTime())
	return &model.AuthPayload{
		AccessToken:          t.GetAccessToken(),
		AccessTokenExpiresAt: t.GetAccessTokenExpiresAt().AsTime(),
		UserID:               t.GetUserId(),
		Roles:                nonNil(claims.Roles),
		Permissions:          nonNil(claims.Permissions),
	}, nil
}

// nonNil turns nil into an empty list: the schema's lists are non-null.
func nonNil(list []string) []string {
	if list == nil {
		return []string{}
	}
	return list
}
