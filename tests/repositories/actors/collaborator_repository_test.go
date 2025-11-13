package actors_test

import (
	"context"
	"fmt"
	"testing"

	"kisanlink-ecom/entities/models/actors"
	actorRepo "kisanlink-ecom/internal/repositories/actors"
	"kisanlink-ecom/tests/data"
	"kisanlink-ecom/tests/mocks"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCollaboratorRepository_Create(t *testing.T) {
	tests := []struct {
		name         string
		collaborator *actors.Collaborator
		setupMock    func(*mocks.MockDBManager)
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "successful creation",
			collaborator: data.CreateTestCollaborator(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*actors.Collaborator")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "database error",
			collaborator: data.CreateTestCollaborator(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*actors.Collaborator")).Return(fmt.Errorf("database error"))
			},
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := actorRepo.NewCollaboratorRepository(mockDB)
			err := repo.Create(context.Background(), tt.collaborator)

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

func TestCollaboratorRepository_GetByID(t *testing.T) {
	testCollab := data.CreateTestCollaboratorWithID("collab-test-id-123")

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful retrieval",
			id:   testCollab.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collabs := args.Get(2).(*[]*actors.Collaborator)
						*collabs = []*actors.Collaborator{testCollab}
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   "non-existent-id",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collabs := args.Get(2).(*[]*actors.Collaborator)
						*collabs = []*actors.Collaborator{}
					}).
					Return(nil)
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "database error",
			id:   testCollab.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Return(fmt.Errorf("database connection failed"))
			},
			wantErr: true,
			errMsg:  "failed to get organization collaborator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := actorRepo.NewCollaboratorRepository(mockDB)
			gotCollab, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, gotCollab)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gotCollab)
				assert.Equal(t, testCollab.ID, gotCollab.ID)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestCollaboratorRepository_Update(t *testing.T) {
	testCollab := data.CreateTestCollaborator()

	tests := []struct {
		name         string
		collaborator *actors.Collaborator
		setupMock    func(*mocks.MockDBManager)
		wantErr      bool
	}{
		{
			name:         "successful update",
			collaborator: testCollab,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*actors.Collaborator")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "database error",
			collaborator: testCollab,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*actors.Collaborator")).
					Return(fmt.Errorf("update failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := actorRepo.NewCollaboratorRepository(mockDB)
			err := repo.Update(context.Background(), tt.collaborator)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestCollaboratorRepository_Delete(t *testing.T) {
	testID := "collab-test-id-123"

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
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*actors.Collaborator")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			id:   testID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*actors.Collaborator")).
					Return(fmt.Errorf("delete failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := actorRepo.NewCollaboratorRepository(mockDB)
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

func TestCollaboratorRepository_List(t *testing.T) {
	testCollabs := data.CreateTestCollaboratorsArray(5, "org-test-id-123")

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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collabs := args.Get(2).(*[]*actors.Collaborator)
						*collabs = testCollabs
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*actors.Collaborator")).
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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collabs := args.Get(2).(*[]*actors.Collaborator)
						*collabs = []*actors.Collaborator{}
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*actors.Collaborator")).
					Return(int64(0), nil)
			},
			wantCount: 0,
			wantTotal: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := actorRepo.NewCollaboratorRepository(mockDB)
			collabs, total, err := repo.List(context.Background(), tt.filter, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(collabs))
				assert.Equal(t, tt.wantTotal, total)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestCollaboratorRepository_GetByContextOrganizationID(t *testing.T) {
	contextOrgID := "org-test-id-123"
	testCollabs := data.CreateTestCollaboratorsArray(3, contextOrgID)

	tests := []struct {
		name         string
		contextOrgID string
		setupMock    func(*mocks.MockDBManager)
		wantCount    int
		wantErr      bool
	}{
		{
			name:         "successful retrieval",
			contextOrgID: contextOrgID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collabs := args.Get(2).(*[]*actors.Collaborator)
						*collabs = testCollabs
					}).
					Return(nil)
			},
			wantCount: 3,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := actorRepo.NewCollaboratorRepository(mockDB)
			collabs, err := repo.GetByContextOrganizationID(context.Background(), tt.contextOrgID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(collabs))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestCollaboratorRepository_GetByRole(t *testing.T) {
	contextOrgID := "org-test-id-123"
	adminCollabs := data.CreateTestCollaboratorsArrayWithRole(2, contextOrgID, actors.CollaboratorRoleAdmin)

	tests := []struct {
		name         string
		contextOrgID string
		role         actors.CollaboratorRole
		setupMock    func(*mocks.MockDBManager)
		wantCount    int
		wantErr      bool
	}{
		{
			name:         "get admin collaborators",
			contextOrgID: contextOrgID,
			role:         actors.CollaboratorRoleAdmin,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collabs := args.Get(2).(*[]*actors.Collaborator)
						*collabs = adminCollabs
					}).
					Return(nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := actorRepo.NewCollaboratorRepository(mockDB)
			collabs, err := repo.GetByRole(context.Background(), tt.contextOrgID, tt.role, nil)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(collabs))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestCollaboratorRepository_GetActiveCollaborators(t *testing.T) {
	contextOrgID := "org-test-id-123"
	activeCollabs := data.CreateTestCollaboratorsArrayWithStatus(2, contextOrgID, actors.CollaboratorStatusActive)

	tests := []struct {
		name         string
		contextOrgID string
		setupMock    func(*mocks.MockDBManager)
		wantCount    int
		wantErr      bool
	}{
		{
			name:         "get active collaborators",
			contextOrgID: contextOrgID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collabs := args.Get(2).(*[]*actors.Collaborator)
						*collabs = activeCollabs
					}).
					Return(nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := actorRepo.NewCollaboratorRepository(mockDB)
			collabs, err := repo.GetActiveCollaborators(context.Background(), tt.contextOrgID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(collabs))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestCollaboratorRepository_GetByEmployeeID(t *testing.T) {
	testCollab := data.CreateTestCollaboratorWithEmployeeID("EMP-123")

	tests := []struct {
		name       string
		employeeID string
		setupMock  func(*mocks.MockDBManager)
		wantFound  bool
		wantErr    bool
	}{
		{
			name:       "successful retrieval",
			employeeID: "EMP-123",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collabs := args.Get(2).(*[]*actors.Collaborator)
						*collabs = []*actors.Collaborator{testCollab}
					}).
					Return(nil)
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name:       "not found",
			employeeID: "NON-EXISTENT",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collabs := args.Get(2).(*[]*actors.Collaborator)
						*collabs = []*actors.Collaborator{}
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

			repo := actorRepo.NewCollaboratorRepository(mockDB)
			collab, err := repo.GetByEmployeeID(context.Background(), tt.employeeID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantFound {
					assert.NotNil(t, collab)
					assert.Equal(t, tt.employeeID, collab.EmployeeID)
				} else {
					assert.Nil(t, collab)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestCollaboratorRepository_ActivateCollaborator(t *testing.T) {
	testCollab := data.CreateTestInactiveCollaborator()
	testCollab.ID = "collab-test-id-123"

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful activation",
			id:   testCollab.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collabs := args.Get(2).(*[]*actors.Collaborator)
						*collabs = []*actors.Collaborator{testCollab}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collab := args.Get(1).(*actors.Collaborator)
						assert.Equal(t, actors.CollaboratorStatusActive, collab.Status)
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

			repo := actorRepo.NewCollaboratorRepository(mockDB)
			err := repo.ActivateCollaborator(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestCollaboratorRepository_DeactivateCollaborator(t *testing.T) {
	testCollab := data.CreateTestCollaborator()
	testCollab.ID = "collab-test-id-123"

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful deactivation",
			id:   testCollab.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collabs := args.Get(2).(*[]*actors.Collaborator)
						*collabs = []*actors.Collaborator{testCollab}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collab := args.Get(1).(*actors.Collaborator)
						assert.Equal(t, actors.CollaboratorStatusInactive, collab.Status)
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

			repo := actorRepo.NewCollaboratorRepository(mockDB)
			err := repo.DeactivateCollaborator(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestCollaboratorRepository_UpdateAccessLevel(t *testing.T) {
	testCollab := data.CreateTestCollaborator()
	testCollab.ID = "collab-test-id-123"

	tests := []struct {
		name        string
		id          string
		accessLevel int
		setupMock   func(*mocks.MockDBManager)
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "successful access level update",
			id:          testCollab.ID,
			accessLevel: 8,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collabs := args.Get(2).(*[]*actors.Collaborator)
						*collabs = []*actors.Collaborator{testCollab}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collab := args.Get(1).(*actors.Collaborator)
						assert.Equal(t, 8, collab.AccessLevel)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "invalid access level - too low",
			id:          testCollab.ID,
			accessLevel: 0,
			setupMock:   func(_ *mocks.MockDBManager) {},
			wantErr:     true,
			errMsg:      "invalid access level",
		},
		{
			name:        "invalid access level - too high",
			id:          testCollab.ID,
			accessLevel: 11,
			setupMock:   func(_ *mocks.MockDBManager) {},
			wantErr:     true,
			errMsg:      "invalid access level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := actorRepo.NewCollaboratorRepository(mockDB)
			err := repo.UpdateAccessLevel(context.Background(), tt.id, tt.accessLevel)

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

func TestCollaboratorRepository_UpdateRole(t *testing.T) {
	testCollab := data.CreateTestCollaborator()
	testCollab.ID = "collab-test-id-123"

	tests := []struct {
		name      string
		id        string
		role      actors.CollaboratorRole
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful role update",
			id:   testCollab.ID,
			role: actors.CollaboratorRoleManager,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collabs := args.Get(2).(*[]*actors.Collaborator)
						*collabs = []*actors.Collaborator{testCollab}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*actors.Collaborator")).
					Run(func(args mock.Arguments) {
						collab := args.Get(1).(*actors.Collaborator)
						assert.Equal(t, actors.CollaboratorRoleManager, collab.Role)
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

			repo := actorRepo.NewCollaboratorRepository(mockDB)
			err := repo.UpdateRole(context.Background(), tt.id, tt.role)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
