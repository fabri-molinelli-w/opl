package auth

import (
	"sort"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// Operation identifiers passed by oapi-codegen (strict mode) to the
// authorization middleware. They match the generated handler names, which are
// the operationIds of api/openapi.yaml with an initial capital letter. They are
// the single key the middleware uses to decide the minimum role of a request.
const (
	OperationGetHealth      = "GetHealth"
	OperationLogin          = "Login"
	OperationGetCurrentUser = "GetCurrentUser"
	OperationChangePassword = "ChangeOwnPassword"
	OperationListUsers      = "ListUsers"
	OperationCreateUser     = "CreateUser"
	OperationUpdateUser     = "UpdateUser"

	OperationListServers   = "ListServers"
	OperationCreateServer  = "CreateServer"
	OperationGetServer     = "GetServer"
	OperationUpdateServer  = "UpdateServer"
	OperationDeleteServer  = "DeleteServer"
	OperationListServices  = "ListServices"
	OperationCreateService = "CreateService"
	OperationGetService    = "GetService"
	OperationUpdateService = "UpdateService"
	OperationDeleteService = "DeleteService"

	OperationListDocuments        = "ListDocuments"
	OperationCreateDocument       = "CreateDocument"
	OperationListDocumentTags     = "ListDocumentTags"
	OperationGetDocument          = "GetDocument"
	OperationUpdateDocument       = "UpdateDocument"
	OperationDeleteDocument       = "DeleteDocument"
	OperationListDocumentVersions = "ListDocumentVersions"
	OperationGetDocumentVersion   = "GetDocumentVersion"

	OperationChatWithOpsbot = "ChatWithOpsbot"
)

// Policy maps every operationId of the contract to the minimum role required
// (the x-min-role extension) plus the public allowlist. Centralizing it here
// means adding an endpoint without an entry results in a default deny.
type Policy struct {
	public  map[string]bool
	minRole map[string]domain.Role
}

func NewPolicy() Policy {
	return Policy{
		public: map[string]bool{
			OperationGetHealth: true,
			OperationLogin:     true,
		},
		minRole: map[string]domain.Role{
			OperationGetCurrentUser: domain.RoleSoporte,
			OperationChangePassword: domain.RoleSoporte,
			OperationListUsers:      domain.RoleAdmin,
			OperationCreateUser:     domain.RoleAdmin,
			OperationUpdateUser:     domain.RoleAdmin,

			OperationListServers:   domain.RoleSoporte,
			OperationCreateServer:  domain.RoleSysadmin,
			OperationGetServer:     domain.RoleSoporte,
			OperationUpdateServer:  domain.RoleSysadmin,
			OperationDeleteServer:  domain.RoleSysadmin,
			OperationListServices:  domain.RoleSoporte,
			OperationCreateService: domain.RoleSysadmin,
			OperationGetService:    domain.RoleSoporte,
			OperationUpdateService: domain.RoleSysadmin,
			OperationDeleteService: domain.RoleSysadmin,

			OperationListDocuments:        domain.RoleSoporte,
			OperationCreateDocument:       domain.RoleSoporte,
			OperationListDocumentTags:     domain.RoleSoporte,
			OperationGetDocument:          domain.RoleSoporte,
			OperationUpdateDocument:       domain.RoleSoporte,
			OperationDeleteDocument:       domain.RoleSysadmin,
			OperationListDocumentVersions: domain.RoleSoporte,
			OperationGetDocumentVersion:   domain.RoleSoporte,

			OperationChatWithOpsbot: domain.RoleSoporte,
		},
	}
}

// IsPublic reports whether the operation is accessible without authentication.
func (p Policy) IsPublic(operationID string) bool {
	return p.public[operationID]
}

// MinRole returns the minimum role for an operation. The second value is false
// for unknown operations (the caller must deny).
func (p Policy) MinRole(operationID string) (domain.Role, bool) {
	role, ok := p.minRole[operationID]
	return role, ok
}

// ProtectedOperations lists the operationIds that require authentication, sorted.
func (p Policy) ProtectedOperations() []string {
	ops := make([]string, 0, len(p.minRole))
	for op := range p.minRole {
		ops = append(ops, op)
	}
	sort.Strings(ops)
	return ops
}
