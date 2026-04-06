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
	resp, err := s.uc.Register(ctx, &biz.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &v1.RegisterReply{
		UserId: resp.UserID,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	resp, err := s.uc.Login(ctx, &biz.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &v1.LoginReply{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
		UserId:       resp.UserID,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *v1.RefreshTokenRequest) (*v1.RefreshTokenReply, error) {
	resp, err := s.uc.RefreshToken(ctx, &biz.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, err
	}
	return &v1.RefreshTokenReply{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
		UserId:       resp.UserID,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *v1.LogoutRequest) (*v1.LogoutReply, error) {
	_, err := s.uc.Logout(ctx, &biz.LogoutRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, err
	}
	return &v1.LogoutReply{}, nil
}

func (s *AuthService) CheckPermission(ctx context.Context, req *v1.CheckPermissionRequest) (*v1.CheckPermissionReply, error) {
	resp, err := s.uc.CheckPermission(ctx, &biz.CheckPermissionRequest{
		SubjectType:  req.SubjectType,
		SubjectID:    req.SubjectId,
		Relation:     req.Relation,
		ResourceType: req.ObjectType,
		ResourceID:   req.ObjectId,
	})
	if err != nil {
		return nil, err
	}

	return &v1.CheckPermissionReply{
		Allowed: resp.Allowed,
	}, nil
}
