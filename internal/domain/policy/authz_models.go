package policy

import (
	"time"
)

// Role represents a collection of permissions.
type Role struct {
	ID          uint          `gorm:"primaryKey"`
	Code        string        `gorm:"unique;not null;size:50"`
	Description string        `gorm:"size:255"`
	Permissions []*Permission `gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Permission represents the ability to perform an action on a resource.
type Permission struct {
	ID          uint      `gorm:"primaryKey"`
	Resource    string    `gorm:"uniqueIndex:idx_resource_action;not null;size:100"`
	Action      string    `gorm:"uniqueIndex:idx_resource_action;not null;size:50"`
	Description string    `gorm:"size:255"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserRole assigns a role to a user.
type UserRole struct {
	UserID    string    `gorm:"primaryKey"`
	RoleID    uint      `gorm:"primaryKey"`
	Role      Role
	CreatedAt time.Time
}
