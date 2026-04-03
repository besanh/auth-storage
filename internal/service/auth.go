package service

import (
	"context"
	v1 "server/api/auth/v1"
	"server/internal/biz"
)

type AuthService struct {
	v1.UnimplementedAuthServer
	uc *biz.AuthUseCase
}

func NewAuthService(uc *biz.AuthUseCase) *AuthService {
	return &AuthService{
		uc: uc,
	}
}

func (s *AuthService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterReply, error) {
	userId, err := s.uc.Register(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &v1.RegisterReply{
		UserId: userId,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	accessToken, refreshToken, expiresIn, userId, err := s.uc.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &v1.LoginReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		UserId:       userId,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *v1.RefreshTokenRequest) (*v1.RefreshTokenReply, error) {
	accessToken, refreshToken, expiresIn, userId, err := s.uc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &v1.RefreshTokenReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		UserId:       userId,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *v1.LogoutRequest) (*v1.LogoutReply, error) {
	err := s.uc.Logout(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &v1.LogoutReply{}, nil
}

func (s *AuthService) CheckPermission(ctx context.Context, req *v1.CheckPermissionRequest) (*v1.CheckPermissionReply, error) {
	allowed, err := s.uc.CheckPermission(ctx, req.SubjectType, req.SubjectId, req.Relation, req.ObjectType, req.ObjectId)
	if err != nil {
		return nil, err
	}

	return &v1.CheckPermissionReply{
		Allowed: allowed,
	}, nil
}
