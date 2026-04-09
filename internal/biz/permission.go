package biz

import (
	"context"
	"server/internal/conf"
)

type PermissionRepo interface {
	CheckPermission(ctx context.Context, input CheckPermissionRequest) (*CheckPermissionResponse, error)
	WriteRelationship(ctx context.Context, input WriteRelationshipRequest) (*WriteRelationshipResponse, error)
	DeleteRelationship(ctx context.Context, input DeleteRelationshipRequest) (*DeleteRelationshipResponse, error)
	SwapRelationship(ctx context.Context, input SwapRelationshipRequest) (*SwapRelationshipResponse, error)
}

type PermissionUseCase struct {
	permissionRepo PermissionRepo
	conf           *conf.Server
}

func NewPermissionUseCase(permissionRepo PermissionRepo, conf *conf.Server) *PermissionUseCase {
	return &PermissionUseCase{
		permissionRepo: permissionRepo,
		conf:           conf,
	}
}

func (p *PermissionUseCase) CheckPermission(ctx context.Context, input CheckPermissionRequest) (*CheckPermissionResponse, error) {
	return p.permissionRepo.CheckPermission(ctx, input)
}

func (p *PermissionUseCase) WriteRelationship(ctx context.Context, input WriteRelationshipRequest) (*WriteRelationshipResponse, error) {
	return p.permissionRepo.WriteRelationship(ctx, input)
}

func (p *PermissionUseCase) DeleteRelationship(ctx context.Context, input DeleteRelationshipRequest) (*DeleteRelationshipResponse, error) {
	return p.permissionRepo.DeleteRelationship(ctx, input)
}

func (p *PermissionUseCase) SwapRelationship(ctx context.Context, input SwapRelationshipRequest) (*SwapRelationshipResponse, error) {
	return p.permissionRepo.SwapRelationship(ctx, input)
}
