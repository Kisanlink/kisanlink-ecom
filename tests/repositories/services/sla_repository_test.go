package services_test

import (
	"context"
	"fmt"
	"testing"

	"kisanlink-ecom/entities/models/services"
	svcRepo "kisanlink-ecom/internal/repositories/services"
	"kisanlink-ecom/tests/data"
	"kisanlink-ecom/tests/mocks"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSLARepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		sla       *services.SLA
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful creation",
			sla:  data.CreateTestSLA(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*services.SLA")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			sla:  data.CreateTestSLA(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*services.SLA")).Return(fmt.Errorf("database error"))
			},
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := svcRepo.NewSLARepository(mockDB)
			err := repo.Create(context.Background(), tt.sla)

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

func TestSLARepository_GetByID(t *testing.T) {
	testSLA := data.CreateTestSLAWithID("sla-test-id-123")

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantSLA   *services.SLA
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful retrieval",
			id:   testSLA.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Run(func(args mock.Arguments) {
						slas := args.Get(2).(*[]*services.SLA)
						*slas = []*services.SLA{testSLA}
					}).
					Return(nil)
			},
			wantSLA: testSLA,
			wantErr: false,
		},
		{
			name: "not found",
			id:   "non-existent-id",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Run(func(args mock.Arguments) {
						slas := args.Get(2).(*[]*services.SLA)
						*slas = []*services.SLA{}
					}).
					Return(nil)
			},
			wantSLA: nil,
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "database error",
			id:   testSLA.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Return(fmt.Errorf("database connection failed"))
			},
			wantSLA: nil,
			wantErr: true,
			errMsg:  "failed to get service SLA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := svcRepo.NewSLARepository(mockDB)
			gotSLA, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, gotSLA)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gotSLA)
				assert.Equal(t, tt.wantSLA.ID, gotSLA.ID)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestSLARepository_Update(t *testing.T) {
	testSLA := data.CreateTestSLA()

	tests := []struct {
		name      string
		sla       *services.SLA
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful update",
			sla:  testSLA,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*services.SLA")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			sla:  testSLA,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*services.SLA")).
					Return(fmt.Errorf("update failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := svcRepo.NewSLARepository(mockDB)
			err := repo.Update(context.Background(), tt.sla)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestSLARepository_Delete(t *testing.T) {
	testID := "sla-test-id-123"

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
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*services.SLA")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			id:   testID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*services.SLA")).
					Return(fmt.Errorf("delete failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := svcRepo.NewSLARepository(mockDB)
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

func TestSLARepository_List(t *testing.T) {
	testSLAs := data.CreateTestSLAsArray(5, "org-test-id-123")

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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Run(func(args mock.Arguments) {
						slas := args.Get(2).(*[]*services.SLA)
						*slas = testSLAs
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*services.SLA")).
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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Run(func(args mock.Arguments) {
						slas := args.Get(2).(*[]*services.SLA)
						*slas = []*services.SLA{}
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*services.SLA")).
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

			repo := svcRepo.NewSLARepository(mockDB)
			slas, total, err := repo.List(context.Background(), tt.filter, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(slas))
				assert.Equal(t, tt.wantTotal, total)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestSLARepository_GetByOrganizationID(t *testing.T) {
	orgID := "org-test-id-123"
	testSLAs := data.CreateTestSLAsArray(3, orgID)

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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Run(func(args mock.Arguments) {
						slas := args.Get(2).(*[]*services.SLA)
						*slas = testSLAs
					}).
					Return(nil)
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:  "no SLAs found",
			orgID: "non-existent-org",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Run(func(args mock.Arguments) {
						slas := args.Get(2).(*[]*services.SLA)
						*slas = []*services.SLA{}
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

			repo := svcRepo.NewSLARepository(mockDB)
			slas, err := repo.GetByOrganizationID(context.Background(), tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(slas))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestSLARepository_GetByCatalogItemID(t *testing.T) {
	catalogItemID := "catalog-item-test-id-123"
	testSLAs := data.CreateTestSLAsArrayForCatalogItem(2, catalogItemID)

	tests := []struct {
		name          string
		catalogItemID string
		setupMock     func(*mocks.MockDBManager)
		wantCount     int
		wantErr       bool
	}{
		{
			name:          "successful retrieval",
			catalogItemID: catalogItemID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Run(func(args mock.Arguments) {
						slas := args.Get(2).(*[]*services.SLA)
						*slas = testSLAs
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

			repo := svcRepo.NewSLARepository(mockDB)
			slas, err := repo.GetByCatalogItemID(context.Background(), tt.catalogItemID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(slas))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestSLARepository_GetBySLAType(t *testing.T) {
	orgID := "org-test-id-123"
	responseSLA := data.CreateTestResponseTimeSLA()
	responseSLA.OrganizationID = orgID

	tests := []struct {
		name       string
		orgID      string
		slaType    services.SLAType
		activeOnly bool
		setupMock  func(*mocks.MockDBManager)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "get response time SLAs",
			orgID:      orgID,
			slaType:    services.SLATypeResponse,
			activeOnly: true,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Run(func(args mock.Arguments) {
						slas := args.Get(2).(*[]*services.SLA)
						*slas = []*services.SLA{responseSLA}
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

			repo := svcRepo.NewSLARepository(mockDB)
			slas, err := repo.GetBySLAType(context.Background(), tt.orgID, tt.slaType, tt.activeOnly)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(slas))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestSLARepository_GetByName(t *testing.T) {
	orgID := "org-test-id-123"
	testSLA := data.CreateTestSLAWithName(orgID, "Test SLA")

	tests := []struct {
		name      string
		orgID     string
		slaName   string
		setupMock func(*mocks.MockDBManager)
		wantFound bool
		wantErr   bool
	}{
		{
			name:    "successful retrieval",
			orgID:   orgID,
			slaName: "Test SLA",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Run(func(args mock.Arguments) {
						slas := args.Get(2).(*[]*services.SLA)
						*slas = []*services.SLA{testSLA}
					}).
					Return(nil)
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name:    "not found",
			orgID:   orgID,
			slaName: "Non-existent SLA",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Run(func(args mock.Arguments) {
						slas := args.Get(2).(*[]*services.SLA)
						*slas = []*services.SLA{}
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

			repo := svcRepo.NewSLARepository(mockDB)
			sla, err := repo.GetByName(context.Background(), tt.orgID, tt.slaName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantFound {
					assert.NotNil(t, sla)
					assert.Equal(t, tt.slaName, sla.Name)
				} else {
					assert.Nil(t, sla)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestSLARepository_DeactivateSLA(t *testing.T) {
	testSLA := data.CreateTestSLA()
	testSLA.ID = "sla-test-id-123"

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful deactivation",
			id:   testSLA.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Run(func(args mock.Arguments) {
						slas := args.Get(2).(*[]*services.SLA)
						*slas = []*services.SLA{testSLA}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*services.SLA")).
					Run(func(args mock.Arguments) {
						sla := args.Get(1).(*services.SLA)
						assert.False(t, sla.IsActive)
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

			repo := svcRepo.NewSLARepository(mockDB)
			err := repo.DeactivateSLA(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestSLARepository_ActivateSLA(t *testing.T) {
	testSLA := data.CreateTestInactiveSLA()
	testSLA.ID = "sla-test-id-123"

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful activation",
			id:   testSLA.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Run(func(args mock.Arguments) {
						slas := args.Get(2).(*[]*services.SLA)
						*slas = []*services.SLA{testSLA}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*services.SLA")).
					Run(func(args mock.Arguments) {
						sla := args.Get(1).(*services.SLA)
						assert.True(t, sla.IsActive)
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

			repo := svcRepo.NewSLARepository(mockDB)
			err := repo.ActivateSLA(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				//nolint:dupl // Intentional test code duplication for comprehensive coverage
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestSLARepository_UpdateThresholds(t *testing.T) {
	testSLA := data.CreateTestSLA()
	testSLA.ID = "sla-test-id-123"
	warning := 70.0
	critical := 90.0

	tests := []struct {
		name      string
		id        string
		warning   *float64
		critical  *float64
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name:     "successful threshold update",
			id:       testSLA.ID,
			warning:  &warning,
			critical: &critical,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*services.SLA")).
					Run(func(args mock.Arguments) {
						slas := args.Get(2).(*[]*services.SLA)
						*slas = []*services.SLA{testSLA}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*services.SLA")).
					Run(func(args mock.Arguments) {
						sla := args.Get(1).(*services.SLA)
						assert.NotNil(t, sla.WarningThreshold)
						assert.NotNil(t, sla.CriticalThreshold)
						assert.Equal(t, warning, *sla.WarningThreshold)
						assert.Equal(t, critical, *sla.CriticalThreshold)
					}).
					Return(nil)
			},
			wantErr: false,
		},
	}

	//nolint:dupl // Intentional test code duplication for comprehensive coverage
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := svcRepo.NewSLARepository(mockDB)
			err := repo.UpdateThresholds(context.Background(), tt.id, tt.warning, tt.critical)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
