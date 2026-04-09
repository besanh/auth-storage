package server

import (
	authV1 "server/api/auth/v1"
	m2mV1 "server/api/m2m_auth/v1"
	permissionV1 "server/api/permission/v1"
	"server/internal/conf"
	"server/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, logger log.Logger, auth *service.AuthService, m2mAuth *service.M2MAuthService, permission *service.PermissionService) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	authV1.RegisterAuthServiceServer(srv, auth)
	m2mV1.RegisterAuthServiceServer(srv, m2mAuth)
	permissionV1.RegisterPermissionServiceServer(srv, permission)
	return srv
}
