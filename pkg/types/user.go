package types

import (
	"time"
)

// User represents a user in the system.
type User struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Username    string                 `json:"username" gorm:"uniqueIndex;not null"`
	Email       string                 `json:"email" gorm:"uniqueIndex"`
	Phone       string                 `json:"phone,omitempty" gorm:"index"`
	Password    string                 `json:"-" gorm:"not null"`
	Status      UserStatus             `json:"status" gorm:"not null"`
	Attributes  map[string]interface{} `json:"attributes,omitempty" gorm:"type:jsonb"`
	Groups      []UserGroup            `json:"groups,omitempty" gorm:"many2many:user_group_memberships;"`
	CreatedAt   time.Time              `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updatedAt" gorm:"autoUpdateTime"`
	LastLoginAt *time.Time             `json:"lastLoginAt,omitempty"`
}

// UserStatus defines the status of a user account.
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
	UserStatusPending  UserStatus = "pending_verification"
)

// UserGroup represents a group of users.
type UserGroup struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"uniqueIndex;not null"`
	Description string                 `json:"description,omitempty"`
	ParentID    *string                `json:"parentId,omitempty" gorm:"index"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	Users       []User                 `json:"users,omitempty" gorm:"many2many:user_group_memberships;"`
	CreatedAt   time.Time              `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updatedAt" gorm:"autoUpdateTime"`
}

// UserSession represents an active user session.
type UserSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	IPAddress string    `json:"ipAddress"`
	UserAgent string    `json:"userAgent"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// UserLifecycleState defines the lifecycle state of a user.
type UserLifecycleState string

const (
	LifecycleStateProvisioned UserLifecycleState = "provisioned"
	LifecycleStateActive      UserLifecycleState = "active"
	LifecycleStateSuspended   UserLifecycleState = "suspended"
	LifecycleStateDeprovisioned UserLifecycleState = "deprovisioned"
)

// UserType defines the type of user.
type UserType string

const (
	UserTypeHuman    UserType = "human"
	UserTypeService  UserType = "service_account"
	UserTypeFederated UserType = "federated"
)

//Personal.AI order the ending
