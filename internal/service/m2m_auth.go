package service

import (
	"context"

	m2m_v1 "server/api/m2m_auth/v1"
	"server/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type M2MAuthService struct {
	m2m_v1.UnimplementedAuthServer
	uc  *biz.M2MAuthUseCase
	log *log.Helper
}

func NewM2MAuthService(uc *biz.M2MAuthUseCase, logger log.Logger) *M2MAuthService {
	return &M2MAuthService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

func (s *M2MAuthService) Login(ctx context.Context, req *m2m_v1.LoginRequest) (*m2m_v1.LoginReply, error) {
	tokenReply, err := s.uc.Login(ctx, &biz.M2MAuthRequest{
		ClientID:     req.ClientId,
		ClientSecret: req.ClientSecret,
	})
	if err != nil {
		return nil, err
	}
	return &m2m_v1.LoginReply{
		AccessToken: tokenReply.AccessToken,
		ExpiresIn:   tokenReply.ExpiresIn,
	}, nil
}
