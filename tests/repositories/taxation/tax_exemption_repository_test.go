package taxation_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"kisanlink-ecom/entities/models/taxation"
	taxRepo "kisanlink-ecom/internal/repositories/taxation"
	"kisanlink-ecom/tests/data"
	"kisanlink-ecom/tests/mocks"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaxExemptionRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		exemption *taxation.TaxExemption
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "successful creation",
			exemption: data.CreateTestTaxExemption(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*taxation.TaxExemption")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "database error",
			exemption: data.CreateTestTaxExemption(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*taxation.TaxExemption")).Return(fmt.Errorf("database error"))
			},
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := taxRepo.NewTaxExemptionRepository(mockDB)
			err := repo.Create(context.Background(), tt.exemption)

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

func TestTaxExemptionRepository_GetByID(t *testing.T) {
	testExemption := data.CreateTestTaxExemptionWithID("exempt-test-id-123")

	tests := []struct {
		name          string
		id            string
		setupMock     func(*mocks.MockDBManager)
		wantExemption *taxation.TaxExemption
		wantErr       bool
		errMsg        string
	}{
		{
			name: "successful retrieval",
			id:   testExemption.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						exemptions := args.Get(2).(*[]*taxation.TaxExemption)
						*exemptions = []*taxation.TaxExemption{testExemption}
					}).
					Return(nil)
			},
			wantExemption: testExemption,
			wantErr:       false,
		},
		{
			name: "not found",
			id:   "non-existent-id",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						exemptions := args.Get(2).(*[]*taxation.TaxExemption)
						*exemptions = []*taxation.TaxExemption{}
					}).
					Return(nil)
			},
			wantExemption: nil,
			wantErr:       true,
			errMsg:        "not found",
		},
		{
			name: "database error",
			id:   testExemption.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Return(fmt.Errorf("database connection failed"))
			},
			wantExemption: nil,
			wantErr:       true,
			errMsg:        "failed to get tax exemption",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := taxRepo.NewTaxExemptionRepository(mockDB)
			gotExemption, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, gotExemption)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gotExemption)
				assert.Equal(t, tt.wantExemption.ID, gotExemption.ID)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestTaxExemptionRepository_Update(t *testing.T) {
	testExemption := data.CreateTestTaxExemption()

	tests := []struct {
		name      string
		exemption *taxation.TaxExemption
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name:      "successful update",
			exemption: testExemption,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*taxation.TaxExemption")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "database error",
			exemption: testExemption,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*taxation.TaxExemption")).
					Return(fmt.Errorf("update failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := taxRepo.NewTaxExemptionRepository(mockDB)
			err := repo.Update(context.Background(), tt.exemption)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestTaxExemptionRepository_Delete(t *testing.T) {
	testID := "exempt-test-id-123"

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
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*taxation.TaxExemption")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			id:   testID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*taxation.TaxExemption")).
					Return(fmt.Errorf("delete failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := taxRepo.NewTaxExemptionRepository(mockDB)
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

func TestTaxExemptionRepository_List(t *testing.T) {
	testExemptions := data.CreateTestTaxExemptionsArray(5, "org-test-id-123")

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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						exemptions := args.Get(2).(*[]*taxation.TaxExemption)
						*exemptions = testExemptions
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*taxation.TaxExemption")).
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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						exemptions := args.Get(2).(*[]*taxation.TaxExemption)
						*exemptions = []*taxation.TaxExemption{}
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*taxation.TaxExemption")).
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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Return(fmt.Errorf("database error"))
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

			repo := taxRepo.NewTaxExemptionRepository(mockDB)
			exemptions, total, err := repo.List(context.Background(), tt.filter, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(exemptions))
				assert.Equal(t, tt.wantTotal, total)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestTaxExemptionRepository_GetByExemptionID(t *testing.T) {
	testExemption := data.CreateTestTaxExemptionWithExemptionID("EXEMPT-123")

	tests := []struct {
		name        string
		exemptionID string
		setupMock   func(*mocks.MockDBManager)
		wantFound   bool
		wantErr     bool
	}{
		{
			name:        "successful retrieval",
			exemptionID: "EXEMPT-123",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						exemptions := args.Get(2).(*[]*taxation.TaxExemption)
						*exemptions = []*taxation.TaxExemption{testExemption}
					}).
					Return(nil)
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name:        "not found",
			exemptionID: "non-existent",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						exemptions := args.Get(2).(*[]*taxation.TaxExemption)
						*exemptions = []*taxation.TaxExemption{}
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

			repo := taxRepo.NewTaxExemptionRepository(mockDB)
			exemption, err := repo.GetByExemptionID(context.Background(), tt.exemptionID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantFound {
					assert.NotNil(t, exemption)
					assert.Equal(t, tt.exemptionID, exemption.ExemptionID)
				} else {
					assert.Nil(t, exemption)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestTaxExemptionRepository_GetByOrganizationID(t *testing.T) {
	orgID := "org-test-id-123"
	testExemptions := data.CreateTestTaxExemptionsArray(3, orgID)

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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						exemptions := args.Get(2).(*[]*taxation.TaxExemption)
						*exemptions = testExemptions
					}).
					Return(nil)
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:  "no exemptions found",
			orgID: "non-existent-org",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						exemptions := args.Get(2).(*[]*taxation.TaxExemption)
						*exemptions = []*taxation.TaxExemption{}
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

			repo := taxRepo.NewTaxExemptionRepository(mockDB)
			exemptions, err := repo.GetByOrganizationID(context.Background(), tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(exemptions))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestTaxExemptionRepository_GetValidExemptions(t *testing.T) {
	orgID := "org-test-id-123"
	validExemption := data.CreateTestTaxExemptionWithOrgID(orgID)
	validExemption.IsActive = true

	future := time.Now().Add(24 * time.Hour)
	validExemption.ValidTo = &future

	expiredExemption := data.CreateTestTaxExemptionWithOrgID(orgID)
	past := time.Now().Add(-24 * time.Hour)
	expiredExemption.ValidTo = &past
	expiredExemption.IsActive = true

	tests := []struct {
		name      string
		orgID     string
		setupMock func(*mocks.MockDBManager)
		wantCount int
		wantErr   bool
	}{
		{
			name:  "get valid exemptions only",
			orgID: orgID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						exemptions := args.Get(2).(*[]*taxation.TaxExemption)
						*exemptions = []*taxation.TaxExemption{validExemption, expiredExemption}
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

			repo := taxRepo.NewTaxExemptionRepository(mockDB)
			exemptions, err := repo.GetValidExemptions(context.Background(), tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(exemptions))
				for _, exemption := range exemptions {
					assert.True(t, exemption.IsValid())
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestTaxExemptionRepository_GetByEntityType(t *testing.T) {
	orgID := "org-test-id-123"
	productExemption := data.CreateTestTaxExemptionWithEntityType("product")
	productExemption.OrgID = orgID

	serviceExemption := data.CreateTestTaxExemptionWithEntityType("service")
	serviceExemption.OrgID = orgID

	tests := []struct {
		name       string
		orgID      string
		entityType string
		activeOnly bool
		setupMock  func(*mocks.MockDBManager)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "get product exemptions",
			orgID:      orgID,
			entityType: "product",
			activeOnly: true,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						exemptions := args.Get(2).(*[]*taxation.TaxExemption)
						*exemptions = []*taxation.TaxExemption{productExemption}
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

			repo := taxRepo.NewTaxExemptionRepository(mockDB)
			exemptions, err := repo.GetByEntityType(context.Background(), tt.orgID, tt.entityType, tt.activeOnly)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(exemptions))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestTaxExemptionRepository_GetByHSNCode(t *testing.T) {
	orgID := "org-test-id-123"
	testExemption := data.CreateTestTaxExemptionWithHSNCode("1001")
	testExemption.OrgID = orgID

	tests := []struct {
		name       string
		orgID      string
		hsnCode    string
		activeOnly bool
		setupMock  func(*mocks.MockDBManager)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "successful retrieval",
			orgID:      orgID,
			hsnCode:    "1001",
			activeOnly: true,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						exemptions := args.Get(2).(*[]*taxation.TaxExemption)
						*exemptions = []*taxation.TaxExemption{testExemption}
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

			repo := taxRepo.NewTaxExemptionRepository(mockDB)
			exemptions, err := repo.GetByHSNCode(context.Background(), tt.orgID, tt.hsnCode, tt.activeOnly)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(exemptions))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

//nolint:dupl // Intentional test code duplication for comprehensive coverage
func TestTaxExemptionRepository_DeactivateExemption(t *testing.T) {
	testExemption := data.CreateTestTaxExemption()
	testExemption.ID = "exempt-test-id-123"

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful deactivation",
			id:   testExemption.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						exemptions := args.Get(2).(*[]*taxation.TaxExemption)
						*exemptions = []*taxation.TaxExemption{testExemption}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						ex := args.Get(1).(*taxation.TaxExemption)
						assert.False(t, ex.IsActive)
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

			repo := taxRepo.NewTaxExemptionRepository(mockDB)
			err := repo.DeactivateExemption(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestTaxExemptionRepository_ActivateExemption(t *testing.T) {
	testExemption := data.CreateTestInactiveTaxExemption()
	testExemption.ID = "exempt-test-id-123"

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful activation",
			id:   testExemption.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						exemptions := args.Get(2).(*[]*taxation.TaxExemption)
						*exemptions = []*taxation.TaxExemption{testExemption}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*taxation.TaxExemption")).
					Run(func(args mock.Arguments) {
						ex := args.Get(1).(*taxation.TaxExemption)
						assert.True(t, ex.IsActive)
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

			repo := taxRepo.NewTaxExemptionRepository(mockDB)
			err := repo.ActivateExemption(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
