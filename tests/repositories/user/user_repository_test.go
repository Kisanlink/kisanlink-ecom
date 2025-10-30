package user_test

import (
	"context"
	"fmt"
	"testing"

	"kisanlink-ecom/entities/models/user"
	userRepo "kisanlink-ecom/internal/repositories/user"
	"kisanlink-ecom/tests/data"
	"kisanlink-ecom/tests/mocks"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestUserRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		user      *user.User
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful creation",
			user: data.CreateTestUser(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			user: data.CreateTestUser(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(fmt.Errorf("database error"))
			},
			wantErr: true,
			errMsg:  "database error",
		},
		{
			name: "nil user",
			user: nil,
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

			repo := userRepo.NewUserRepository(mockDB)
			err := repo.Create(context.Background(), tt.user)

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

func TestUserRepository_GetByID(t *testing.T) {
	testUser := data.CreateTestUser()

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantUser  *user.User
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful retrieval",
			id:   testUser.Id,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("GetByID", mock.Anything, testUser.Id, mock.AnythingOfType("*user.User")).
					Run(func(args mock.Arguments) {
						u := args.Get(2).(*user.User)
						*u = *testUser
					}).
					Return(nil)
			},
			wantUser: testUser,
			wantErr:  false,
		},
		{
			name: "user not found",
			id:   "non-existent-id",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("GetByID", mock.Anything, "non-existent-id", mock.AnythingOfType("*user.User")).
					Return(gorm.ErrRecordNotFound)
			},
			wantUser: nil,
			wantErr:  true,
			errMsg:   "failed to get user by ID",
		},
		{
			name: "database error",
			id:   testUser.Id,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("GetByID", mock.Anything, testUser.Id, mock.AnythingOfType("*user.User")).
					Return(fmt.Errorf("database connection failed"))
			},
			wantUser: nil,
			wantErr:  true,
			errMsg:   "failed to get user by ID",
		},
		{
			name: "empty ID",
			id:   "",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("GetByID", mock.Anything, "", mock.AnythingOfType("*user.User")).
					Return(fmt.Errorf("invalid ID"))
			},
			wantUser: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := userRepo.NewUserRepository(mockDB)
			gotUser, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, gotUser)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gotUser)
				assert.Equal(t, tt.wantUser.Id, gotUser.Id)
				assert.Equal(t, tt.wantUser.Username, gotUser.Username)
				assert.Equal(t, tt.wantUser.Email, gotUser.Email)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestUserRepository_GetByUsername(t *testing.T) {
	testUser := data.CreateTestUser()

	tests := []struct {
		name      string
		username  string
		setupMock func(*mocks.MockDBManager)
		wantUser  *user.User
		wantErr   bool
	}{
		{
			name:     "successful retrieval",
			username: testUser.Username,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*user.User")).
					Run(func(args mock.Arguments) {
						users := args.Get(2).(*[]*user.User)
						*users = []*user.User{testUser}
					}).
					Return(nil)
			},
			wantUser: testUser,
			wantErr:  false,
		},
		{
			name:     "user not found",
			username: "nonexistent",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*user.User")).
					Run(func(args mock.Arguments) {
						users := args.Get(2).(*[]*user.User)
						*users = []*user.User{}
					}).
					Return(nil)
			},
			wantUser: nil,
			wantErr:  false,
		},
		{
			name:     "database error",
			username: testUser.Username,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*user.User")).
					Return(fmt.Errorf("database error"))
			},
			wantUser: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := userRepo.NewUserRepository(mockDB)
			gotUser, err := repo.GetByUsername(context.Background(), tt.username)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, gotUser)
			} else {
				assert.NoError(t, err)
				if tt.wantUser != nil {
					assert.NotNil(t, gotUser)
					assert.Equal(t, tt.wantUser.Username, gotUser.Username)
				} else {
					assert.Nil(t, gotUser)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestUserRepository_GetByEmail(t *testing.T) {
	testUser := data.CreateTestUser()

	tests := []struct {
		name      string
		email     string
		setupMock func(*mocks.MockDBManager)
		wantUser  *user.User
		wantErr   bool
	}{
		{
			name:  "successful retrieval",
			email: testUser.Email,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*user.User")).
					Run(func(args mock.Arguments) {
						users := args.Get(2).(*[]*user.User)
						*users = []*user.User{testUser}
					}).
					Return(nil)
			},
			wantUser: testUser,
			wantErr:  false,
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*user.User")).
					Run(func(args mock.Arguments) {
						users := args.Get(2).(*[]*user.User)
						*users = []*user.User{}
					}).
					Return(nil)
			},
			wantUser: nil,
			wantErr:  false,
		},
		{
			name:  "database error",
			email: testUser.Email,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*user.User")).
					Return(fmt.Errorf("database error"))
			},
			wantUser: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := userRepo.NewUserRepository(mockDB)
			gotUser, err := repo.GetByEmail(context.Background(), tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, gotUser)
			} else {
				assert.NoError(t, err)
				if tt.wantUser != nil {
					assert.NotNil(t, gotUser)
					assert.Equal(t, tt.wantUser.Email, gotUser.Email)
				} else {
					assert.Nil(t, gotUser)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	testUser := data.CreateTestUser()

	tests := []struct {
		name      string
		user      *user.User
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful update",
			user: testUser,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			user: testUser,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).
					Return(fmt.Errorf("update failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := userRepo.NewUserRepository(mockDB)
			err := repo.Update(context.Background(), tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	testID := "user-test-id-123"

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
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*user.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			id:   testID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*user.User")).
					Return(fmt.Errorf("delete failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := userRepo.NewUserRepository(mockDB)
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

func TestUserRepository_List(t *testing.T) {
	testUsers := data.CreateTestUsersArray(5)

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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*user.User")).
					Run(func(args mock.Arguments) {
						users := args.Get(2).(*[]*user.User)
						*users = testUsers
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*user.User")).
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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*user.User")).
					Run(func(args mock.Arguments) {
						users := args.Get(2).(*[]*user.User)
						*users = []*user.User{}
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*user.User")).
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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*user.User")).
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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*user.User")).
					Run(func(args mock.Arguments) {
						users := args.Get(2).(*[]*user.User)
						*users = testUsers
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*user.User")).
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

			repo := userRepo.NewUserRepository(mockDB)
			users, total, err := repo.List(context.Background(), tt.filter, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(users))
				assert.Equal(t, tt.wantTotal, total)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestUserRepository_ActivateUser(t *testing.T) {
	testUser := data.CreateTestInactiveUser()

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful activation",
			id:   testUser.Id,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("GetByID", mock.Anything, testUser.Id, mock.AnythingOfType("*user.User")).
					Run(func(args mock.Arguments) {
						u := args.Get(2).(*user.User)
						*u = *testUser
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).
					Run(func(args mock.Arguments) {
						u := args.Get(1).(*user.User)
						assert.True(t, u.IsActive)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   "non-existent",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("GetByID", mock.Anything, "non-existent", mock.AnythingOfType("*user.User")).
					Return(gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := userRepo.NewUserRepository(mockDB)
			err := repo.ActivateUser(context.Background(), tt.id)

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
func TestUserRepository_DeactivateUser(t *testing.T) {
	testUser := data.CreateTestUser()

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful deactivation",
			id:   testUser.Id,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("GetByID", mock.Anything, testUser.Id, mock.AnythingOfType("*user.User")).
					Run(func(args mock.Arguments) {
						u := args.Get(2).(*user.User)
						*u = *testUser
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).
					Run(func(args mock.Arguments) {
						u := args.Get(1).(*user.User)
						assert.False(t, u.IsActive)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   "non-existent",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("GetByID", mock.Anything, "non-existent", mock.AnythingOfType("*user.User")).
					Return(gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := userRepo.NewUserRepository(mockDB)
			err := repo.DeactivateUser(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
