package collaborator

// UserContext contains authenticated user information from JWT interceptor
type UserContext struct {
	UserID   string
	FpoID    uint64
	Roles    []string
	TenantID string
	Email    string
}

// CreateSagaResult contains the result of create collaborator saga
type CreateSagaResult struct {
	CollaboratorID uint64
	AddressID      string
}
