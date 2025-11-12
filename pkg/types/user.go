package types

import (
	"context"
	"time"
)

// User represents a user entity in the system.
// It contains profile information, credentials, status, and associations.
type User struct {
	// ID is the unique identifier for the user.
	ID string `json:"id" gorm:"primaryKey"`
	// Username is the unique name used for logging in.
	Username string `json:"username" gorm:"uniqueIndex;not null"`
	// Email is the user's email address, also used for communication and recovery.
	Email string `json:"email" gorm:"uniqueIndex"`
	// Phone is the user's phone number.
	Phone string `json:"phone,omitempty" gorm:"index"`
	// Password is the hashed password of the user. It is not exposed in API responses.
	Password string `json:"-" gorm:"not null"`
	// Status indicates the current state of the user's account.
	Status UserStatus `json:"status" gorm:"not null"`
	// Attributes stores custom user profile information as a JSON object.
	Attributes map[string]interface{} `json:"attributes,omitempty" gorm:"type:jsonb"`
	// Groups lists the groups the user is a member of.
	Groups []UserGroup `json:"groups,omitempty" gorm:"many2many:user_group_memberships;"`
	// CreatedAt is the timestamp when the user was created.
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	// UpdatedAt is the timestamp of the last update.
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
	// LastLoginAt records the timestamp of the user's last successful login.
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
}

// UserStatus defines the possible states of a user account.
type UserStatus string

// Possible user account statuses.
const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
	UserStatusPending  UserStatus = "pending_verification"
)

// UserGroup represents a collection of users, used for assigning permissions or managing policies collectively.
type UserGroup struct {
	// ID is the unique identifier for the group.
	ID string `json:"id" gorm:"primaryKey"`
	// Name is the unique name of the group.
	Name string `json:"name" gorm:"uniqueIndex;not null"`
	// Description provides a human-readable explanation of the group's purpose.
	Description string `json:"description,omitempty"`
	// ParentID allows for creating hierarchical group structures.
	ParentID *string `json:"parentId,omitempty" gorm:"index"`
	// Metadata stores custom information about the group.
	Metadata map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	// Users lists the members of this group.
	Users []User `json:"users,omitempty" gorm:"many2many:user_group_memberships;"`
	// CreatedAt is the timestamp when the group was created.
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	// UpdatedAt is the timestamp of the last update.
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// UserSession represents an active user session, tracking their login state.
type UserSession struct {
	// ID is the unique identifier for the session.
	ID string `json:"id"`
	// UserID is the ID of the user this session belongs to.
	UserID string `json:"userId"`
	// IPAddress is the IP address from which the session was initiated.
	IPAddress string `json:"ipAddress"`
	// UserAgent is the user agent string of the client.
	UserAgent string `json:"userAgent"`
	// CreatedAt is the timestamp when the session was created.
	CreatedAt time.Time `json:"createdAt"`
	// ExpiresAt is the timestamp when the session will expire.
	ExpiresAt time.Time `json:"expiresAt"`
}

// UserLifecycleState defines the lifecycle state of a user, typically for provisioning workflows.
type UserLifecycleState string

// Possible user lifecycle states.
const (
	LifecycleStateProvisioned   UserLifecycleState = "provisioned"
	LifecycleStateActive        UserLifecycleState = "active"
	LifecycleStateSuspended     UserLifecycleState = "suspended"
	LifecycleStateDeprovisioned UserLifecycleState = "deprovisioned"
)

// UserType distinguishes between different kinds of users in the system.
type UserType string

// Supported user types.
const (
	UserTypeHuman     UserType = "human"
	UserTypeService   UserType = "service_account"
	UserTypeFederated UserType = "federated"
)

// UserRepository defines the interface for user persistence.
type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	ListUsers(ctx context.Context) ([]*User, error)
}
