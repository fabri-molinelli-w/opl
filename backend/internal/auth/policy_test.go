package auth

import (
	"testing"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

func TestPolicy_PublicOperations(t *testing.T) {
	policy := NewPolicy()
	if !policy.IsPublic(OperationGetHealth) {
		t.Error("getHealth must be public")
	}
	if !policy.IsPublic(OperationLogin) {
		t.Error("login must be public")
	}
	if policy.IsPublic(OperationListUsers) {
		t.Error("listUsers must not be public")
	}
}

func TestPolicy_MinRoles(t *testing.T) {
	policy := NewPolicy()

	tests := []struct {
		op   string
		want domain.Role
	}{
		{OperationGetCurrentUser, domain.RoleSoporte},
		{OperationChangePassword, domain.RoleSoporte},
		{OperationListUsers, domain.RoleAdmin},
		{OperationCreateUser, domain.RoleAdmin},
		{OperationUpdateUser, domain.RoleAdmin},
	}
	for _, tt := range tests {
		got, ok := policy.MinRole(tt.op)
		if !ok {
			t.Fatalf("operation %q has no minimum role", tt.op)
		}
		if got != tt.want {
			t.Errorf("operation %q: expected min role %q, got %q", tt.op, tt.want, got)
		}
	}
}

func TestPolicy_SoporteAllowedOnSoporteOperation(t *testing.T) {
	policy := NewPolicy()
	minRole, ok := policy.MinRole(OperationGetCurrentUser)
	if !ok {
		t.Fatal("getCurrentUser has no minimum role")
	}
	if !domain.RoleSoporte.AtLeast(minRole) {
		t.Error("soporte role must satisfy a soporte-minimum operation")
	}
}

func TestPolicy_DefaultDeny(t *testing.T) {
	policy := NewPolicy()
	if _, ok := policy.MinRole("unknownOperation"); ok {
		t.Error("unknown operations must not have a minimum role (default deny)")
	}
}
