package roles_test

import (
	"context"
	"fmt"
	"testing"

	"kisanlink-ecom/entities/models/roles"
	roleRepo "kisanlink-ecom/internal/repositories/roles"
	"kisanlink-ecom/tests/data"
	"kisanlink-ecom/tests/mocks"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrganizationRoleRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		orgRole   *roles.OrganizationRole
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
		errMsg    string
	}{
		{
			name:    "successful creation",
			orgRole: data.CreateTestOrganizationRole(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*roles.OrganizationRole")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "database error",
			orgRole: data.CreateTestOrganizationRole(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*roles.OrganizationRole")).Return(fmt.Errorf("database error"))
			},
			wantErr: true,
			errMsg:  "database error",
		},
		{
			name:    "nil organization role",
			orgRole: nil,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.Anything).Return(fmt.Errorf("cannot create nil entity"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewOrganizationRoleRepository(mockDB)
			err := repo.Create(context.Background(), tt.orgRole)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestOrganizationRoleRepository_GetByID(t *testing.T) {
	testOrgRole := data.CreateTestOrganizationRole()

	tests := []struct {
		name        string
		id          string
		setupMock   func(*mocks.MockDBManager)
		wantOrgRole *roles.OrganizationRole
		wantErr     bool
		errMsg      string
	}{
		{
			name: "successful retrieval",
			id:   testOrgRole.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{testOrgRole}
					}).
					Return(nil)
			},
			wantOrgRole: testOrgRole,
			wantErr:     false,
		},
		{
			name: "organization role not found",
			id:   "non-existent-id",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{}
					}).
					Return(nil)
			},
			wantOrgRole: nil,
			wantErr:     true,
			errMsg:      "not found",
		},
		{
			name: "database error",
			id:   testOrgRole.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Return(fmt.Errorf("database connection failed"))
			},
			wantOrgRole: nil,
			wantErr:     true,
			errMsg:      "failed to get organization role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewOrganizationRoleRepository(mockDB)
			gotOrgRole, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, gotOrgRole)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gotOrgRole)
				assert.Equal(t, tt.wantOrgRole.ID, gotOrgRole.ID)
				assert.Equal(t, tt.wantOrgRole.OrganizationID, gotOrgRole.OrganizationID)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestOrganizationRoleRepository_Update(t *testing.T) {
	testOrgRole := data.CreateTestOrganizationRole()

	tests := []struct {
		name      string
		orgRole   *roles.OrganizationRole
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name:    "successful update",
			orgRole: testOrgRole,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*roles.OrganizationRole")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "database error",
			orgRole: testOrgRole,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*roles.OrganizationRole")).
					Return(fmt.Errorf("update failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewOrganizationRoleRepository(mockDB)
			err := repo.Update(context.Background(), tt.orgRole)

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
func TestOrganizationRoleRepository_Delete(t *testing.T) {
	testID := "org-role-test-id-123"

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful deletion",
			id:   testID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*roles.OrganizationRole")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			id:   testID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*roles.OrganizationRole")).
					Return(fmt.Errorf("delete failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewOrganizationRoleRepository(mockDB)
			err := repo.Delete(context.Background(), tt.id)

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
func TestOrganizationRoleRepository_List(t *testing.T) {
	orgID := "org-test-id-123"
	testOrgRole1 := data.CreateTestOrganizationRoleWithOrgID(orgID)
	testOrgRole2 := data.CreateTestOrganizationRoleWithOrgID(orgID)
	testOrgRoles := []*roles.OrganizationRole{testOrgRole1, testOrgRole2}

	tests := []struct {
		name      string
		filter    *base.Filter
		limit     int
		offset    int
		setupMock func(*mocks.MockDBManager)
		wantCount int
		wantTotal int
		wantErr   bool
	}{
		{
			name:   "successful list with pagination",
			filter: nil,
			limit:  20,
			offset: 0,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = testOrgRoles
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*roles.OrganizationRole")).
					Return(int64(2), nil)
			},
			wantCount: 2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:   "empty result",
			filter: nil,
			limit:  20,
			offset: 0,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{}
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*roles.OrganizationRole")).
					Return(int64(0), nil)
			},
			wantCount: 0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:   "database error on list",
			filter: nil,
			limit:  20,
			offset: 0,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Return(fmt.Errorf("database error"))
			},
			wantCount: 0,
			wantTotal: 0,
			wantErr:   true,
		},
		{
			name:   "database error on count",
			filter: nil,
			limit:  20,
			offset: 0,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = testOrgRoles
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*roles.OrganizationRole")).
					Return(int64(0), fmt.Errorf("count failed"))
			},
			wantCount: 0,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewOrganizationRoleRepository(mockDB)
			orgRoles, total, err := repo.List(context.Background(), tt.filter, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(orgRoles))
				assert.Equal(t, tt.wantTotal, total)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestOrganizationRoleRepository_GetByOrganizationID(t *testing.T) {
	orgID := "org-test-id-123"
	testOrgRole1 := data.CreateTestOrganizationRoleWithOrgID(orgID)
	testOrgRole2 := data.CreateTestOrganizationRoleWithOrgID(orgID)
	testOrgRoles := []*roles.OrganizationRole{testOrgRole1, testOrgRole2}

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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = testOrgRoles
					}).
					Return(nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:  "no roles found",
			orgID: "non-existent-org",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{}
					}).
					Return(nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:  "database error",
			orgID: orgID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Return(fmt.Errorf("database error"))
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewOrganizationRoleRepository(mockDB)
			orgRoles, err := repo.GetByOrganizationID(context.Background(), tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(orgRoles))
				if tt.wantCount > 0 {
					assert.Equal(t, tt.orgID, orgRoles[0].OrganizationID)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestOrganizationRoleRepository_GetByAAARoleID(t *testing.T) {
	aaaRoleID := "aaa-role-test-id-123"
	testOrgRole1 := data.CreateTestOrganizationRole()
	testOrgRole1.AAARoleID = aaaRoleID
	testOrgRole2 := data.CreateTestOrganizationRole()
	testOrgRole2.AAARoleID = aaaRoleID
	testOrgRoles := []*roles.OrganizationRole{testOrgRole1, testOrgRole2}

	tests := []struct {
		name      string
		aaaRoleID string
		setupMock func(*mocks.MockDBManager)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "successful retrieval",
			aaaRoleID: aaaRoleID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = testOrgRoles
					}).
					Return(nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "no roles found",
			aaaRoleID: "non-existent-role",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{}
					}).
					Return(nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "database error",
			aaaRoleID: aaaRoleID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Return(fmt.Errorf("database error"))
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewOrganizationRoleRepository(mockDB)
			orgRoles, err := repo.GetByAAARoleID(context.Background(), tt.aaaRoleID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(orgRoles))
				if tt.wantCount > 0 {
					assert.Equal(t, tt.aaaRoleID, orgRoles[0].AAARoleID)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestOrganizationRoleRepository_GetDefaultRole(t *testing.T) {
	orgID := "org-test-id-123"
	testDefaultRole := data.CreateTestDefaultOrganizationRole()
	testDefaultRole.OrganizationID = orgID

	tests := []struct {
		name      string
		orgID     string
		setupMock func(*mocks.MockDBManager)
		wantFound bool
		wantErr   bool
	}{
		{
			name:  "successful retrieval",
			orgID: orgID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{testDefaultRole}
					}).
					Return(nil)
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name:  "not found",
			orgID: "non-existent-org",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{}
					}).
					Return(nil)
			},
			wantFound: false,
			wantErr:   false,
		},
		{
			name:  "database error",
			orgID: orgID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Return(fmt.Errorf("database error"))
			},
			wantFound: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewOrganizationRoleRepository(mockDB)
			orgRole, err := repo.GetDefaultRole(context.Background(), tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantFound {
					assert.NotNil(t, orgRole)
					assert.True(t, orgRole.IsDefault)
					assert.Equal(t, tt.orgID, orgRole.OrganizationID)
				} else {
					assert.Nil(t, orgRole)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestOrganizationRoleRepository_SetAsDefault(t *testing.T) {
	orgID := "org-test-id-123"
	roleID := "org-role-target-id"

	currentDefault := data.CreateTestDefaultOrganizationRole()
	currentDefault.OrganizationID = orgID

	targetRole := data.CreateTestOrganizationRoleWithOrgID(orgID)
	targetRole.ID = roleID
	targetRole.IsDefault = false

	tests := []struct {
		name      string
		orgID     string
		roleID    string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name:   "successful set as default",
			orgID:  orgID,
			roleID: roleID,
			setupMock: func(m *mocks.MockDBManager) {
				// First call to get all org roles
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{currentDefault, targetRole}
					}).
					Return(nil).
					Once()

				// Update to unset current default
				m.On("Update", mock.Anything, mock.AnythingOfType("*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						role := args.Get(1).(*roles.OrganizationRole)
						assert.False(t, role.IsDefault)
					}).
					Return(nil).
					Once()

				// GetByID for target role
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{targetRole}
					}).
					Return(nil).
					Once()

				// Update to set new default
				m.On("Update", mock.Anything, mock.AnythingOfType("*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						role := args.Get(1).(*roles.OrganizationRole)
						assert.True(t, role.IsDefault)
					}).
					Return(nil).
					Once()
			},
			wantErr: false,
		},
		{
			name:   "role not found",
			orgID:  orgID,
			roleID: "non-existent-role",
			setupMock: func(m *mocks.MockDBManager) {
				// First call to get all org roles
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{}
					}).
					Return(nil).
					Once()

				// GetByID for target role
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{}
					}).
					Return(nil).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewOrganizationRoleRepository(mockDB)
			err := repo.SetAsDefault(context.Background(), tt.orgID, tt.roleID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestOrganizationRoleRepository_GetByOrgAndAAARoleID(t *testing.T) {
	orgID := "org-test-id-123"
	aaaRoleID := "aaa-role-test-id-456"
	testOrgRole := data.CreateTestOrganizationRoleWithOrgID(orgID)
	testOrgRole.AAARoleID = aaaRoleID

	tests := []struct {
		name      string
		orgID     string
		aaaRoleID string
		setupMock func(*mocks.MockDBManager)
		wantFound bool
		wantErr   bool
	}{
		{
			name:      "successful retrieval",
			orgID:     orgID,
			aaaRoleID: aaaRoleID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{testOrgRole}
					}).
					Return(nil)
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "not found",
			orgID:     "non-existent-org",
			aaaRoleID: aaaRoleID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{}
					}).
					Return(nil)
			},
			wantFound: false,
			wantErr:   false,
		},
		{
			name:      "database error",
			orgID:     orgID,
			aaaRoleID: aaaRoleID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Return(fmt.Errorf("database error"))
			},
			wantFound: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewOrganizationRoleRepository(mockDB)
			orgRole, err := repo.GetByOrgAndAAARoleID(context.Background(), tt.orgID, tt.aaaRoleID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantFound {
					assert.NotNil(t, orgRole)
					assert.Equal(t, tt.orgID, orgRole.OrganizationID)
					assert.Equal(t, tt.aaaRoleID, orgRole.AAARoleID)
				} else {
					assert.Nil(t, orgRole)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestOrganizationRoleRepository_DeactivateOrgRole(t *testing.T) {
	testOrgRole := data.CreateTestOrganizationRole()

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful deactivation",
			id:   testOrgRole.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{testOrgRole}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						or := args.Get(1).(*roles.OrganizationRole)
						assert.False(t, or.IsActive)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "organization role not found",
			id:   "non-existent",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{}
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

			repo := roleRepo.NewOrganizationRoleRepository(mockDB)
			err := repo.DeactivateOrgRole(context.Background(), tt.id)

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
func TestOrganizationRoleRepository_ActivateOrgRole(t *testing.T) {
	testOrgRole := data.CreateTestOrganizationRole()
	testOrgRole.IsActive = false

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful activation",
			id:   testOrgRole.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{testOrgRole}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						or := args.Get(1).(*roles.OrganizationRole)
						assert.True(t, or.IsActive)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "organization role not found",
			id:   "non-existent",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{}
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

			repo := roleRepo.NewOrganizationRoleRepository(mockDB)
			err := repo.ActivateOrgRole(context.Background(), tt.id)

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
func TestOrganizationRoleRepository_GetOrgRoles_ActiveOnly(t *testing.T) {
	orgID := "org-test-id-123"
	activeRole := data.CreateTestOrganizationRoleWithOrgID(orgID)
	activeRole.IsActive = true
	inactiveRole := data.CreateTestOrganizationRoleWithOrgID(orgID)
	inactiveRole.IsActive = false

	tests := []struct {
		name       string
		orgID      string
		activeOnly bool
		setupMock  func(*mocks.MockDBManager)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "get all roles",
			orgID:      orgID,
			activeOnly: false,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						filter := args.Get(1).(*base.Filter)
						assert.Len(t, filter.Group.Conditions, 2) // organization_id + deleted_at
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{activeRole, inactiveRole}
					}).
					Return(nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "get active roles only",
			orgID:      orgID,
			activeOnly: true,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.OrganizationRole")).
					Run(func(args mock.Arguments) {
						filter := args.Get(1).(*base.Filter)
						assert.Len(t, filter.Group.Conditions, 3) // organization_id + is_active + deleted_at
						orgRoles := args.Get(2).(*[]*roles.OrganizationRole)
						*orgRoles = []*roles.OrganizationRole{activeRole}
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

			repo := roleRepo.NewOrganizationRoleRepository(mockDB)
			orgRoles, err := repo.GetOrgRoles(context.Background(), tt.orgID, tt.activeOnly)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(orgRoles))
			}

			mockDB.AssertExpectations(t)
		})
	}
}
