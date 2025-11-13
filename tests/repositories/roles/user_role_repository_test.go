package roles_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"kisanlink-ecom/entities/models/roles"
	roleRepo "kisanlink-ecom/internal/repositories/roles"
	"kisanlink-ecom/tests/data"
	"kisanlink-ecom/tests/mocks"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserRoleRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		userRole  *roles.UserRole
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "successful creation",
			userRole: data.CreateTestUserRole(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*roles.UserRole")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "database error",
			userRole: data.CreateTestUserRole(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*roles.UserRole")).Return(fmt.Errorf("database error"))
			},
			wantErr: true,
			errMsg:  "database error",
		},
		{
			name:     "nil user role",
			userRole: nil,
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

			repo := roleRepo.NewUserRoleRepository(mockDB)
			err := repo.Create(context.Background(), tt.userRole)

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
func TestUserRoleRepository_GetByID(t *testing.T) {
	testUserRole := data.CreateTestUserRole()

	tests := []struct {
		name         string
		id           string
		setupMock    func(*mocks.MockDBManager)
		wantUserRole *roles.UserRole
		wantErr      bool
		errMsg       string
	}{
		{
			name: "successful retrieval",
			id:   testUserRole.Id,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{testUserRole}
					}).
					Return(nil)
			},
			wantUserRole: testUserRole,
			wantErr:      false,
		},
		{
			name: "user role not found",
			id:   "non-existent-id",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{}
					}).
					Return(nil)
			},
			wantUserRole: nil,
			wantErr:      true,
			errMsg:       "not found",
		},
		{
			name: "database error",
			id:   testUserRole.Id,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Return(fmt.Errorf("database connection failed"))
			},
			wantUserRole: nil,
			wantErr:      true,
			errMsg:       "failed to get user role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewUserRoleRepository(mockDB)
			gotUserRole, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, gotUserRole)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gotUserRole)
				assert.Equal(t, tt.wantUserRole.Id, gotUserRole.Id)
				assert.Equal(t, tt.wantUserRole.UserId, gotUserRole.UserId)
				assert.Equal(t, tt.wantUserRole.RoleId, gotUserRole.RoleId)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestUserRoleRepository_Update(t *testing.T) {
	testUserRole := data.CreateTestUserRole()

	tests := []struct {
		name      string
		userRole  *roles.UserRole
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name:     "successful update",
			userRole: testUserRole,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*roles.UserRole")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "database error",
			userRole: testUserRole,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*roles.UserRole")).
					Return(fmt.Errorf("update failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewUserRoleRepository(mockDB)
			err := repo.Update(context.Background(), tt.userRole)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestUserRoleRepository_Delete(t *testing.T) {
	testID := "user-role-test-id-123"

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
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*roles.UserRole")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			id:   testID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*roles.UserRole")).
					Return(fmt.Errorf("delete failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := roleRepo.NewUserRoleRepository(mockDB)
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
func TestUserRoleRepository_List(t *testing.T) {
	testUserRoles := data.CreateTestUserRolesArray(5, "user-test-id-123")

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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = testUserRoles
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*roles.UserRole")).
					Return(int64(5), nil)
			},
			wantCount: 5,
			wantTotal: 5,
			wantErr:   false,
		},
		{
			name:   "empty result",
			filter: nil,
			limit:  20,
			offset: 0,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{}
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*roles.UserRole")).
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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = testUserRoles
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*roles.UserRole")).
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

			repo := roleRepo.NewUserRoleRepository(mockDB)
			userRoles, total, err := repo.List(context.Background(), tt.filter, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(userRoles))
				assert.Equal(t, tt.wantTotal, total)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestUserRoleRepository_GetByUserID(t *testing.T) {
	userID := "user-test-id-123"
	testUserRoles := data.CreateTestUserRolesArray(3, userID)

	tests := []struct {
		name      string
		userID    string
		setupMock func(*mocks.MockDBManager)
		wantCount int
		wantErr   bool
	}{
		{
			name:   "successful retrieval",
			userID: userID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = testUserRoles
					}).
					Return(nil)
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:   "no roles found",
			userID: "non-existent-user",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{}
					}).
					Return(nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "database error",
			userID: userID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
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

			repo := roleRepo.NewUserRoleRepository(mockDB)
			userRoles, err := repo.GetByUserID(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(userRoles))
				if tt.wantCount > 0 {
					assert.Equal(t, tt.userID, userRoles[0].UserId)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestUserRoleRepository_GetByRoleID(t *testing.T) {
	roleID := "role-test-id-123"
	testUserRole1 := data.CreateTestUserRoleWithIDs("user-1", roleID)
	testUserRole2 := data.CreateTestUserRoleWithIDs("user-2", roleID)
	testUserRoles := []*roles.UserRole{testUserRole1, testUserRole2}

	tests := []struct {
		name      string
		roleID    string
		setupMock func(*mocks.MockDBManager)
		wantCount int
		wantErr   bool
	}{
		{
			name:   "successful retrieval",
			roleID: roleID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = testUserRoles
					}).
					Return(nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:   "no users found",
			roleID: "non-existent-role",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{}
					}).
					Return(nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "database error",
			roleID: roleID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
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

			repo := roleRepo.NewUserRoleRepository(mockDB)
			userRoles, err := repo.GetByRoleID(context.Background(), tt.roleID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(userRoles))
				if tt.wantCount > 0 {
					assert.Equal(t, tt.roleID, userRoles[0].RoleId)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestUserRoleRepository_GetActiveUserRole(t *testing.T) {
	userID := "user-test-id-123"
	roleID := "role-test-id-456"
	testUserRole := data.CreateTestUserRoleWithIDs(userID, roleID)
	testUserRole.IsActive = true

	tests := []struct {
		name      string
		userID    string
		roleID    string
		setupMock func(*mocks.MockDBManager)
		wantFound bool
		wantErr   bool
	}{
		{
			name:   "successful retrieval",
			userID: userID,
			roleID: roleID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{testUserRole}
					}).
					Return(nil)
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name:   "not found",
			userID: userID,
			roleID: "non-existent-role",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{}
					}).
					Return(nil)
			},
			wantFound: false,
			wantErr:   false,
		},
		{
			name:   "database error",
			userID: userID,
			roleID: roleID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
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

			repo := roleRepo.NewUserRoleRepository(mockDB)
			userRole, err := repo.GetActiveUserRole(context.Background(), tt.userID, tt.roleID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantFound {
					assert.NotNil(t, userRole)
					assert.Equal(t, tt.userID, userRole.UserId)
					assert.Equal(t, tt.roleID, userRole.RoleId)
					assert.True(t, userRole.IsActive)
				} else {
					assert.Nil(t, userRole)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestUserRoleRepository_DeactivateUserRole(t *testing.T) {
	testUserRole := data.CreateTestUserRole()

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful deactivation",
			id:   testUserRole.Id,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{testUserRole}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*roles.UserRole")).
					Run(func(args mock.Arguments) {
						ur := args.Get(1).(*roles.UserRole)
						assert.False(t, ur.IsActive)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "user role not found",
			id:   "non-existent",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{}
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

			repo := roleRepo.NewUserRoleRepository(mockDB)
			err := repo.DeactivateUserRole(context.Background(), tt.id)

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
func TestUserRoleRepository_ActivateUserRole(t *testing.T) {
	testUserRole := data.CreateTestInactiveUserRole()

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful activation",
			id:   testUserRole.Id,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{testUserRole}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*roles.UserRole")).
					Run(func(args mock.Arguments) {
						ur := args.Get(1).(*roles.UserRole)
						assert.True(t, ur.IsActive)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "user role not found",
			id:   "non-existent",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{}
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

			repo := roleRepo.NewUserRoleRepository(mockDB)
			err := repo.ActivateUserRole(context.Background(), tt.id)

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
func TestUserRoleRepository_GetUserRoles_ActiveOnly(t *testing.T) {
	userID := "user-test-id-123"
	activeRole := data.CreateTestUserRoleWithIDs(userID, "role-1")
	activeRole.IsActive = true
	inactiveRole := data.CreateTestUserRoleWithIDs(userID, "role-2")
	inactiveRole.IsActive = false

	tests := []struct {
		name       string
		userID     string
		activeOnly bool
		setupMock  func(*mocks.MockDBManager)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "get all roles",
			userID:     userID,
			activeOnly: false,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						filter := args.Get(1).(*base.Filter)
						assert.Len(t, filter.Group.Conditions, 2) // user_id + deleted_at
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{activeRole, inactiveRole}
					}).
					Return(nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "get active roles only",
			userID:     userID,
			activeOnly: true,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						filter := args.Get(1).(*base.Filter)
						assert.Len(t, filter.Group.Conditions, 3) // user_id + is_active + deleted_at
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{activeRole}
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

			repo := roleRepo.NewUserRoleRepository(mockDB)
			userRoles, err := repo.GetUserRoles(context.Background(), tt.userID, tt.activeOnly)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(userRoles))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestUserRoleRepository_GetRoleUsers_ActiveOnly(t *testing.T) {
	roleID := "role-test-id-123"
	activeUserRole := data.CreateTestUserRoleWithIDs("user-1", roleID)
	activeUserRole.IsActive = true
	inactiveUserRole := data.CreateTestUserRoleWithIDs("user-2", roleID)
	inactiveUserRole.IsActive = false

	tests := []struct {
		name       string
		roleID     string
		activeOnly bool
		setupMock  func(*mocks.MockDBManager)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "get all users",
			roleID:     roleID,
			activeOnly: false,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						filter := args.Get(1).(*base.Filter)
						assert.Len(t, filter.Group.Conditions, 2) // role_id + deleted_at
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{activeUserRole, inactiveUserRole}
					}).
					Return(nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "get active users only",
			roleID:     roleID,
			activeOnly: true,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*roles.UserRole")).
					Run(func(args mock.Arguments) {
						filter := args.Get(1).(*base.Filter)
						assert.Len(t, filter.Group.Conditions, 3) // role_id + is_active + deleted_at
						userRoles := args.Get(2).(*[]*roles.UserRole)
						*userRoles = []*roles.UserRole{activeUserRole}
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

			repo := roleRepo.NewUserRoleRepository(mockDB)
			userRoles, err := repo.GetRoleUsers(context.Background(), tt.roleID, tt.activeOnly)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(userRoles))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestUserRole_IsExpired(t *testing.T) {
	tests := []struct {
		name       string
		userRole   *roles.UserRole
		wantResult bool
	}{
		{
			name:       "no expiration",
			userRole:   data.CreateTestUserRole(),
			wantResult: false,
		},
		{
			name:       "future expiration",
			userRole:   data.CreateTestUserRoleWithExpiration(24 * time.Hour),
			wantResult: false,
		},
		{
			name:       "past expiration",
			userRole:   data.CreateTestUserRoleWithExpiration(-24 * time.Hour),
			wantResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.userRole.IsExpired()
			assert.Equal(t, tt.wantResult, result)
		})
	}
}

func TestUserRole_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		userRole   *roles.UserRole
		wantResult bool
	}{
		{
			name:       "active without expiration",
			userRole:   data.CreateTestUserRole(),
			wantResult: true,
		},
		{
			name:       "inactive without expiration",
			userRole:   data.CreateTestInactiveUserRole(),
			wantResult: false,
		},
		{
			name: "active with future expiration",
			userRole: func() *roles.UserRole {
				ur := data.CreateTestUserRole()
				future := time.Now().Add(24 * time.Hour)
				ur.ExpiresAt = &future
				return ur
			}(),
			wantResult: true,
		},
		{
			name: "active with past expiration",
			userRole: func() *roles.UserRole {
				ur := data.CreateTestUserRole()
				past := time.Now().Add(-24 * time.Hour)
				ur.ExpiresAt = &past
				return ur
			}(),
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.userRole.IsValid()
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
