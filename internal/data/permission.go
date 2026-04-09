package data

import (
	"context"
	"server/internal/biz"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/go-kratos/kratos/v2/log"
)

type permissionRepo struct {
	data *Data
	log  *log.Helper
}

func NewPermissionRepo(data *Data, logger log.Logger) biz.PermissionRepo {
	return &permissionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *permissionRepo) CheckPermission(ctx context.Context, cp *biz.CheckPermissionRequest) (*biz.CheckPermissionResponse, error) {
	resp, err := r.data.SpiceDB.CheckPermission(ctx, &v1.CheckPermissionRequest{
		Resource: &v1.ObjectReference{
			ObjectType: cp.ResourceType,
			ObjectId:   cp.ResourceID,
		},
		Permission: cp.Relation,
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: cp.SubjectType,
				ObjectId:   cp.SubjectID,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &biz.CheckPermissionResponse{
		Allowed: resp.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION,
	}, nil
}

func (r *permissionRepo) WriteRelationship(ctx context.Context, rel *biz.WriteRelationshipRequest) (*biz.WriteRelationshipResponse, error) {
	_, err := r.data.SpiceDB.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{
			{
				Operation: v1.RelationshipUpdate_OPERATION_TOUCH,
				Relationship: &v1.Relationship{
					Resource: &v1.ObjectReference{
						ObjectType: rel.ResourceType,
						ObjectId:   rel.ResourceID,
					},
					Relation: rel.Relation,
					Subject: &v1.SubjectReference{
						Object: &v1.ObjectReference{
							ObjectType: rel.SubjectType,
							ObjectId:   rel.SubjectID,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &biz.WriteRelationshipResponse{
		Success: true,
	}, nil
}

func (r *permissionRepo) DeleteRelationship(ctx context.Context, rel *biz.DeleteRelationshipRequest) (*biz.DeleteRelationshipResponse, error) {
	_, err := r.data.SpiceDB.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{
			{
				Operation: v1.RelationshipUpdate_OPERATION_DELETE,
				Relationship: &v1.Relationship{
					Resource: &v1.ObjectReference{
						ObjectType: rel.ResourceType,
						ObjectId:   rel.ResourceID,
					},
					Relation: rel.Relation,
					Subject: &v1.SubjectReference{
						Object: &v1.ObjectReference{
							ObjectType: rel.SubjectType,
							ObjectId:   rel.SubjectID,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &biz.DeleteRelationshipResponse{
		Success: true,
	}, nil
}

func (r *permissionRepo) SwapRelationship(ctx context.Context, rel *biz.SwapRelationshipRequest) (*biz.SwapRelationshipResponse, error) {
	_, err := r.data.SpiceDB.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{
			{
				Operation: v1.RelationshipUpdate_OPERATION_DELETE,
				Relationship: &v1.Relationship{
					Resource: &v1.ObjectReference{
						ObjectType: rel.ResourceType,
						ObjectId:   rel.ResourceID,
					},
					Relation: rel.Relation,
					Subject: &v1.SubjectReference{
						Object: &v1.ObjectReference{
							ObjectType: rel.OldSubjectType,
							ObjectId:   rel.OldSubjectID,
						},
					},
				},
			},
			{
				Operation: v1.RelationshipUpdate_OPERATION_TOUCH,
				Relationship: &v1.Relationship{
					Resource: &v1.ObjectReference{
						ObjectType: rel.ResourceType,
						ObjectId:   rel.ResourceID,
					},
					Relation: rel.Relation,
					Subject: &v1.SubjectReference{
						Object: &v1.ObjectReference{
							ObjectType: rel.NewSubjectType,
							ObjectId:   rel.NewSubjectID,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &biz.SwapRelationshipResponse{
		Success: true,
	}, nil
}
