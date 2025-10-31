package roles_test

import (
	"context"
	"fmt"
	"testing"

	"kisanlink-ecom/entities/models/roles"
	roleRepo "kisanlink-ecom/internal/repositories/roles"
	"kisanlink-ecom/tests/data"
	"kisanlink-ecom/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEcommerceRoleRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		role      *roles.EcommerceRole
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful creation",
			role: data.CreateTestEcommerceRole(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*roles.EcommerceRole")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			role: data.CreateTestEcommerceRole(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*roles.EcommerceRole")).Return(fmt.Errorf("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewEcommerceRoleRepository(mockDB)
			err := repo.Create(context.Background(), tt.role)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestEcommerceRoleRepository_GetByID(t *testing.T) {
	testRole := data.CreateTestEcommerceRole()

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantRole  *roles.EcommerceRole
		wantErr   bool
	}{
		{
			name: "successful retrieval",
			id:   testRole.AAARoleID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{testRole}
					}).
					Return(nil)
			},
			wantRole: testRole,
			wantErr:  false,
		},
		{
			name: "role not found",
			id:   "non-existent-id",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{}
					}).
					Return(nil)
			},
			wantRole: nil,
			wantErr:  true,
		},
		{
			name: "database error",
			id:   testRole.AAARoleID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Return(fmt.Errorf("database error"))
			},
			wantRole: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewEcommerceRoleRepository(mockDB)
			gotRole, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, gotRole)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gotRole)
				assert.Equal(t, tt.wantRole.AAARoleID, gotRole.AAARoleID)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestEcommerceRoleRepository_GetByRoleName(t *testing.T) {
	testRole := data.CreateTestEcommerceRole()

	tests := []struct {
		name      string
		roleName  string
		setupMock func(*mocks.MockDBManager)
		wantFound bool
		wantErr   bool
	}{
		{
			name:     "successful retrieval",
			roleName: testRole.RoleName,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{testRole}
					}).
					Return(nil)
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name:     "not found",
			roleName: "Non Existent Role",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{}
					}).
					Return(nil)
			},
			wantFound: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewEcommerceRoleRepository(mockDB)
			role, err := repo.GetByRoleName(context.Background(), tt.roleName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantFound {
					assert.NotNil(t, role)
					assert.Equal(t, tt.roleName, role.RoleName)
				} else {
					assert.Nil(t, role)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestEcommerceRoleRepository_GetByOrganizationID(t *testing.T) {
	orgID := "org-test-id-123"
	testRole := data.CreateTestEcommerceRoleWithOrgScope(orgID)

	tests := []struct {
		name      string
		orgID     string
		setupMock func(*mocks.MockDBManager)
		wantCount int
		wantErr   bool
	}{
		{
			name:  "successful retrieval",
			orgID: orgID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{testRole}
					}).
					Return(nil)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:  "no roles found",
			orgID: "non-existent-org",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{}
					}).
					Return(nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewEcommerceRoleRepository(mockDB)
			roles, err := repo.GetByOrganizationID(context.Background(), tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(roles))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestEcommerceRoleRepository_GetPlatformRoles(t *testing.T) {
	platformRole := data.CreateTestEcommerceRole()
	platformRole.OrganizationID = nil
	activeRole := data.CreateTestEcommerceRole()
	activeRole.OrganizationID = nil
	activeRole.IsActive = true
	inactiveRole := data.CreateTestEcommerceRole()
	inactiveRole.OrganizationID = nil
	inactiveRole.IsActive = false

	tests := []struct {
		name       string
		activeOnly bool
		setupMock  func(*mocks.MockDBManager)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "get all platform roles",
			activeOnly: false,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{activeRole, inactiveRole}
					}).
					Return(nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "get active platform roles only",
			activeOnly: true,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{activeRole}
					}).
					Return(nil)
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewEcommerceRoleRepository(mockDB)
			roles, err := repo.GetPlatformRoles(context.Background(), tt.activeOnly)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(roles))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestEcommerceRoleRepository_GetRolesWithPermission(t *testing.T) {
	roleWithCatalog := data.CreateTestEcommerceRole()
	roleWithCatalog.CanManageCatalog = true

	roleWithoutCatalog := data.CreateTestEcommerceRole()
	roleWithoutCatalog.CanManageCatalog = false

	tests := []struct {
		name       string
		permission string
		activeOnly bool
		setupMock  func(*mocks.MockDBManager)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "find roles with catalog permission",
			permission: "catalog.manage",
			activeOnly: false,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{roleWithCatalog, roleWithoutCatalog}
					}).
					Return(nil)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "no roles with permission",
			permission: "nonexistent.permission",
			activeOnly: false,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{roleWithCatalog}
					}).
					Return(nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewEcommerceRoleRepository(mockDB)
			roles, err := repo.GetRolesWithPermission(context.Background(), tt.permission, tt.activeOnly)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(roles))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestEcommerceRoleRepository_DeactivateRole(t *testing.T) {
	testRole := data.CreateTestEcommerceRole()

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful deactivation",
			id:   testRole.AAARoleID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{testRole}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						r := args.Get(1).(*roles.EcommerceRole)
						assert.False(t, r.IsActive)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "role not found",
			id:   "non-existent",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{}
					}).
					Return(nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewEcommerceRoleRepository(mockDB)
			err := repo.DeactivateRole(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestEcommerceRoleRepository_ActivateRole(t *testing.T) {
	testRole := data.CreateTestEcommerceRole()
	testRole.IsActive = false

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful activation",
			id:   testRole.AAARoleID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{testRole}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						r := args.Get(1).(*roles.EcommerceRole)
						assert.True(t, r.IsActive)
					}).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewEcommerceRoleRepository(mockDB)
			err := repo.ActivateRole(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestEcommerceRoleRepository_SyncFromAAA(t *testing.T) {
	existingRole := data.CreateTestEcommerceRole()

	tests := []struct {
		name        string
		aaaRoleID   string
		roleName    string
		description string
		setupMock   func(*mocks.MockDBManager)
		wantErr     bool
	}{
		{
			name:        "create new role",
			aaaRoleID:   "new-role-id",
			roleName:    "New Role",
			description: "New role description",
			setupMock: func(m *mocks.MockDBManager) {
				// GetByID returns not found
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{}
					}).
					Return(nil)
				// Create new role
				m.On("Create", mock.Anything, mock.AnythingOfType("*roles.EcommerceRole")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "update existing role",
			aaaRoleID:   existingRole.AAARoleID,
			roleName:    "Updated Role Name",
			description: "Updated description",
			setupMock: func(m *mocks.MockDBManager) {
				// GetByID returns existing role
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.EcommerceRole")).
					Run(func(args mock.Arguments) {
						ecomRoles := args.Get(2).(*[]*roles.EcommerceRole)
						*ecomRoles = []*roles.EcommerceRole{existingRole}
					}).
					Return(nil)
				// Update existing role
				m.On("Update", mock.Anything, mock.AnythingOfType("*roles.EcommerceRole")).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewEcommerceRoleRepository(mockDB)
			err := repo.SyncFromAAA(context.Background(), tt.aaaRoleID, tt.roleName, tt.description)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestEcommerceRole_HasPermission(t *testing.T) {
	role := data.CreateTestEcommerceRole()
	role.CanManageCatalog = true
	role.CanManageOrders = false

	tests := []struct {
		name       string
		permission string
		want       bool
	}{
		{
			name:       "has catalog permission",
			permission: "catalog.manage",
			want:       true,
		},
		{
			name:       "does not have orders permission",
			permission: "orders.manage",
			want:       false,
		},
		{
			name:       "unknown permission",
			permission: "unknown.permission",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := role.HasPermission(tt.permission)
			assert.Equal(t, tt.want, result)
		})
	}
}
