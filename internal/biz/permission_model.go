package biz

type CheckPermissionRequest struct {
	ResourceType string
	ResourceID   string
	Relation     string
	SubjectType  string
	SubjectID    string
}

type CheckPermissionResponse struct {
	Allowed bool
}

type WriteRelationshipRequest struct {
	ResourceType string
	ResourceID   string
	Relation     string
	SubjectType  string
	SubjectID    string
}

type WriteRelationshipResponse struct {
	Success bool
}

type DeleteRelationshipRequest struct {
	ResourceType string
	ResourceID   string
	Relation     string
	SubjectType  string
	SubjectID    string
}

type DeleteRelationshipResponse struct {
	Success bool
}
