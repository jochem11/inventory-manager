// Package grpc exposes the auth service over gRPC. The handlers only convert
// between protobuf messages and the service's types; the service does the
// work.
package grpc

import (
	"context"

	"github.com/jochem11/inventory-manager/services/auth-service/internal/domain"
	authpb "github.com/jochem11/inventory-manager/services/auth-service/pkg/pb/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	auth domain.AuthService
}

func NewAuthHandler(auth domain.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	userID, err := h.auth.Register(ctx, domain.RegisterInput{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
		Phone:     req.Phone,
	})
	if err != nil {
		return nil, toStatus(ctx, err)
	}
	return &authpb.RegisterResponse{UserId: userID}, nil
}

func (h *AuthHandler) VerifyEmail(ctx context.Context, req *authpb.VerifyEmailRequest) (*authpb.VerifyEmailResponse, error) {
	if err := h.auth.VerifyEmail(ctx, req.GetToken()); err != nil {
		return nil, toStatus(ctx, err)
	}
	return &authpb.VerifyEmailResponse{}, nil
}

func (h *AuthHandler) ResendVerification(ctx context.Context, req *authpb.ResendVerificationRequest) (*authpb.ResendVerificationResponse, error) {
	if err := h.auth.ResendVerification(ctx, req.GetEmail()); err != nil {
		return nil, toStatus(ctx, err)
	}
	return &authpb.ResendVerificationResponse{}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	tokens, err := h.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, toStatus(ctx, err)
	}
	return &authpb.LoginResponse{Tokens: toProtoTokens(tokens)}, nil
}

func (h *AuthHandler) Refresh(ctx context.Context, req *authpb.RefreshRequest) (*authpb.RefreshResponse, error) {
	tokens, err := h.auth.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, toStatus(ctx, err)
	}
	return &authpb.RefreshResponse{Tokens: toProtoTokens(tokens)}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
	if err := h.auth.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, toStatus(ctx, err)
	}
	return &authpb.LogoutResponse{}, nil
}

func (h *AuthHandler) GetPublicKeys(ctx context.Context, _ *authpb.GetPublicKeysRequest) (*authpb.GetPublicKeysResponse, error) {
	keys, err := h.auth.PublicKeys()
	if err != nil {
		return nil, toStatus(ctx, err)
	}
	resp := &authpb.GetPublicKeysResponse{}
	for _, k := range keys {
		resp.Keys = append(resp.Keys, &authpb.PublicKey{KeyId: k.KeyID, Algorithm: k.Algorithm, PublicKey: k.DER})
	}
	return resp, nil
}

func toProtoTokens(t *domain.Tokens) *authpb.Tokens {
	return &authpb.Tokens{
		UserId:                t.UserID,
		AccessToken:           t.AccessToken,
		AccessTokenExpiresAt:  timestamppb.New(t.AccessTokenExpiresAt),
		RefreshToken:          t.RefreshToken,
		RefreshTokenExpiresAt: timestamppb.New(t.RefreshTokenExpiresAt),
	}
}
