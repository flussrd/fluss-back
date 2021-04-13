package models

// ResourceType resource type
type ResourceType string

const (
	// ResourceTypeModule module
	ResourceTypeModule ResourceType = "module"
	// ResourceTypeUser user
	ResourceTypeUser ResourceType = "user"
)

// ActionType action type
type ActionType string

const (
	// ActionTypeRead read
	ActionTypeRead ActionType = "read"
	// ActionTypeWrite write
	ActionTypeWrite ActionType = "write"
	// ActionTypeUpdate update
	ActionTypeUpdate ActionType = "update"
	// ActionTypeDelete delete
	ActionTypeDelete ActionType = "delete"
	// ActionTypeAll all
	ActionTypeAll ActionType = "*"
)

// Role represents a role where permissions are defined
type Role struct {
	// Name is the name of the role
	Name string `json:"roleName" bson:"roleName"`
	// Permissions the list of permissions linked to the role
	Permissions []Permission `json:"permissions" bson:"permissions"`
}

// Permission represents a permission related to a role
type Permission struct {
	// Resource resource to be accesible on the permission
	Resource ResourceType `json:"resource" bson:"resource"`
	// Action action to be done or performed on the resource
	Action ActionType `json:"action" bson:"action"`
}