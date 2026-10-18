package domain

import "time"

// Role is the hierarchical application role carried in the JWT.
type Role string

const (
	RoleSoporte  Role = "soporte"
	RoleSysadmin Role = "sysadmin"
	RoleAdmin    Role = "admin"
)

// Valid reports whether r is a known role.
func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleSysadmin, RoleSoporte:
		return true
	default:
		return false
	}
}

var roleRank = map[Role]int{
	RoleSoporte:  1,
	RoleSysadmin: 2,
	RoleAdmin:    3,
}

// AtLeast reports whether r is equal to or higher than min in the hierarchy
// admin > sysadmin > soporte.
func (r Role) AtLeast(min Role) bool {
	return roleRank[r] >= roleRank[min]
}

// User is the domain representation of an application user. It carries the
// password hash so the service can verify it; the hash is never serialized.
type User struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	Role         Role
	Active       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserUpdate describes a partial update of a user.
type UserUpdate struct {
	Name   *string
	Email  *string
	Role   *Role
	Active *bool
}
