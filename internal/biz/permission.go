package biz

import (
	"context"
	"server/internal/conf"
)

type PermissionRepo interface {
	CheckPermission(ctx context.Context, cp *CheckPermissionRequest) (*CheckPermissionResponse, error)
	WriteRelationship(ctx context.Context, rel *WriteRelationshipRequest) (*WriteRelationshipResponse, error)
}

type PermissionUseCase struct {
	permissionRepo PermissionRepo
	conf           *conf.Server
}

func NewPermissionUseCase(permissionRepo PermissionRepo, conf *conf.Server) *PermissionUseCase {
	return &PermissionUseCase{permissionRepo: permissionRepo, conf: conf}
}

func (p *PermissionUseCase) CheckPermission(ctx context.Context, cp *CheckPermissionRequest) (*CheckPermissionResponse, error) {
	return p.permissionRepo.CheckPermission(ctx, cp)
}

func (p *PermissionUseCase) WriteRelationship(ctx context.Context, rel *WriteRelationshipRequest) (*WriteRelationshipResponse, error) {
	return p.permissionRepo.WriteRelationship(ctx, rel)
}
