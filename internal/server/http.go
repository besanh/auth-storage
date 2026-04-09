package server

import (
	authV1 "server/api/auth/v1"
	m2mV1 "server/api/m2m_auth/v1"
	permissionV1 "server/api/permission/v1"
	"server/internal/conf"
	"server/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, logger log.Logger, auth *service.AuthService, m2mAuth *service.M2MAuthService, permission *service.PermissionService) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	authV1.RegisterAuthServiceHTTPServer(srv, auth)
	m2mV1.RegisterM2MAuthServiceHTTPServer(srv, m2mAuth)
	permissionV1.RegisterPermissionServiceHTTPServer(srv, permission)
	return srv
}
