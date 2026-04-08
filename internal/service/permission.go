package service

import (
	"context"

	v1 "server/api/permission/v1"
	"server/internal/biz"
)

type PermissionService struct {
	v1.UnimplementedPermissionServer
	uc *biz.PermissionUseCase
}

func NewPermissionService(uc *biz.PermissionUseCase) *PermissionService {
	return &PermissionService{
		uc: uc,
	}
}

func (s *PermissionService) CheckPermission(ctx context.Context, req *v1.CheckPermissionRequest) (*v1.CheckPermissionReply, error) {
	resp, err := s.uc.CheckPermission(ctx, &biz.CheckPermissionRequest{
		ResourceType: req.GetResourceType(),
		ResourceID:   req.GetResourceId(),
		Relation:     req.GetRelation(),
		SubjectType:  req.GetSubjectType(),
		SubjectID:    req.GetSubjectId(),
	})
	if err != nil {
		return nil, err
	}

	return &v1.CheckPermissionReply{
		Allowed: resp.Allowed,
	}, nil
}

func (s *PermissionService) WriteRelationship(ctx context.Context, req *v1.WriteRelationshipRequest) (*v1.WriteRelationshipReply, error) {
	resp, err := s.uc.WriteRelationship(ctx, &biz.WriteRelationshipRequest{
		ResourceType: req.GetResourceType(),
		ResourceID:   req.GetResourceId(),
		Relation:     req.GetRelation(),
		SubjectType:  req.GetSubjectType(),
		SubjectID:    req.GetSubjectId(),
	})
	if err != nil {
		return nil, err
	}

	return &v1.WriteRelationshipReply{
		Success: resp.Success,
	}, nil
}

func (s *PermissionService) DeleteRelationship(ctx context.Context, req *v1.DeleteRelationshipRequest) (*v1.DeleteRelationshipReply, error) {
	resp, err := s.uc.DeleteRelationship(ctx, &biz.DeleteRelationshipRequest{
		ResourceType: req.GetResourceType(),
		ResourceID:   req.GetResourceId(),
		Relation:     req.GetRelation(),
		SubjectType:  req.GetSubjectType(),
		SubjectID:    req.GetSubjectId(),
	})
	if err != nil {
		return nil, err
	}

	return &v1.DeleteRelationshipReply{
		Success: resp.Success,
	}, nil
}
