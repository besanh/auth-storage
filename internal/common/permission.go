package common

// The entities that exist in spicedb graph
type ResourceType string

const (
	TypeUser   ResourceType = "user"
	TypeFolder ResourceType = "folder"
	TypeFile   ResourceType = "file"
)

func (r ResourceType) String() string {
	return string(r)
}

// The actions or roles a subject can have on a resource
type Relation string

const (
	// Standard Access Roles
	RelationOwner  Relation = "owner"  // Ultimate control, can delete
	RelationWriter Relation = "writer" // Can edit files, upload to folders
	RelationViewer Relation = "viewer" // Read-only access

	// Structural Links
	RelationParent Relation = "parent" // Links a file/folder to its parent folder
)

func (r Relation) String() string {
	return string(r)
}

// The system-level permissions granted to microservices
type SystemScope string

const (
	ScopeFGAWrite SystemScope = "auth:fga:write" // Allows service to mutate SpiceDB
	ScopeFGARead  SystemScope = "auth:fga:read"  // Allows service to query SpiceDB
)

func (s SystemScope) String() string {
	return string(s)
}
