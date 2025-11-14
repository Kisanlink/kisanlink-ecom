package discounts_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/discounts"
	discRepo "github.com/Kisanlink/kisanlink-ecom/internal/repositories/discounts"
	"github.com/Kisanlink/kisanlink-ecom/tests/data"
	"github.com/Kisanlink/kisanlink-ecom/tests/mocks"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDiscountRuleRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		rule      *discounts.DiscountRule
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful creation",
			rule: data.CreateTestDiscountRule(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*discounts.DiscountRule")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			rule: data.CreateTestDiscountRule(),
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*discounts.DiscountRule")).Return(fmt.Errorf("database error"))
			},
			wantErr: true,
			errMsg:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := discRepo.NewDiscountRuleRepository(mockDB)
			err := repo.Create(context.Background(), tt.rule)

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

func TestDiscountRuleRepository_GetByID(t *testing.T) {
	testRule := data.CreateTestDiscountRuleWithID("rule-test-id-123")

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantRule  *discounts.DiscountRule
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful retrieval",
			id:   testRule.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						rules := args.Get(2).(*[]*discounts.DiscountRule)
						*rules = []*discounts.DiscountRule{testRule}
					}).
					Return(nil)
			},
			wantRule: testRule,
			wantErr:  false,
		},
		{
			name: "not found",
			id:   "non-existent-id",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						rules := args.Get(2).(*[]*discounts.DiscountRule)
						*rules = []*discounts.DiscountRule{}
					}).
					Return(nil)
			},
			wantRule: nil,
			wantErr:  true,
			errMsg:   "not found",
		},
		{
			name: "database error",
			id:   testRule.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*discounts.DiscountRule")).
					Return(fmt.Errorf("database connection failed"))
			},
			wantRule: nil,
			wantErr:  true,
			errMsg:   "failed to get discount rule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := discRepo.NewDiscountRuleRepository(mockDB)
			gotRule, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, gotRule)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gotRule)
				assert.Equal(t, tt.wantRule.ID, gotRule.ID)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDiscountRuleRepository_Update(t *testing.T) {
	testRule := data.CreateTestDiscountRule()

	tests := []struct {
		name      string
		rule      *discounts.DiscountRule
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful update",
			rule: testRule,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*discounts.DiscountRule")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			rule: testRule,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*discounts.DiscountRule")).
					Return(fmt.Errorf("update failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := discRepo.NewDiscountRuleRepository(mockDB)
			err := repo.Update(context.Background(), tt.rule)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDiscountRuleRepository_Delete(t *testing.T) {
	testID := "rule-test-id-123"

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
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*discounts.DiscountRule")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			id:   testID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("Delete", mock.Anything, testID, mock.AnythingOfType("*discounts.DiscountRule")).
					Return(fmt.Errorf("delete failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := discRepo.NewDiscountRuleRepository(mockDB)
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

func TestDiscountRuleRepository_List(t *testing.T) {
	testRules := data.CreateTestDiscountRulesArray(5, "org-test-id-123")

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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						rules := args.Get(2).(*[]*discounts.DiscountRule)
						*rules = testRules
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*discounts.DiscountRule")).
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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						rules := args.Get(2).(*[]*discounts.DiscountRule)
						*rules = []*discounts.DiscountRule{}
					}).
					Return(nil)
				m.On("Count", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*discounts.DiscountRule")).
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

			repo := discRepo.NewDiscountRuleRepository(mockDB)
			rules, total, err := repo.List(context.Background(), tt.filter, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(rules))
				assert.Equal(t, tt.wantTotal, total)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDiscountRuleRepository_GetByOrganizationID(t *testing.T) {
	orgID := "org-test-id-123"
	testRules := data.CreateTestDiscountRulesArray(3, orgID)

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
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						rules := args.Get(2).(*[]*discounts.DiscountRule)
						*rules = testRules
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

			repo := discRepo.NewDiscountRuleRepository(mockDB)
			rules, err := repo.GetByOrganizationID(context.Background(), tt.orgID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(rules))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDiscountRuleRepository_GetByRuleType(t *testing.T) {
	orgID := "org-test-id-123"
	stackableRules := data.CreateTestDiscountRulesArrayByType(2, orgID, "stackable")

	tests := []struct {
		name       string
		orgID      string
		ruleType   string
		activeOnly bool
		setupMock  func(*mocks.MockDBManager)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "get stackable rules",
			orgID:      orgID,
			ruleType:   "stackable",
			activeOnly: true,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						rules := args.Get(2).(*[]*discounts.DiscountRule)
						*rules = stackableRules
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

			repo := discRepo.NewDiscountRuleRepository(mockDB)
			rules, err := repo.GetByRuleType(context.Background(), tt.orgID, tt.ruleType, tt.activeOnly)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(rules))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDiscountRuleRepository_GetByPriority(t *testing.T) {
	orgID := "org-test-id-123"
	testRules := data.CreateTestDiscountRulesArray(3, orgID)

	tests := []struct {
		name       string
		orgID      string
		activeOnly bool
		setupMock  func(*mocks.MockDBManager)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "get rules by priority",
			orgID:      orgID,
			activeOnly: true,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						filter := args.Get(1).(*base.Filter)
						assert.Len(t, filter.Sort, 1)
						assert.Equal(t, "priority", filter.Sort[0].Field)
						assert.Equal(t, "DESC", filter.Sort[0].Direction)

						rules := args.Get(2).(*[]*discounts.DiscountRule)
						*rules = testRules
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

			repo := discRepo.NewDiscountRuleRepository(mockDB)
			rules, err := repo.GetByPriority(context.Background(), tt.orgID, tt.activeOnly)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(rules))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDiscountRuleRepository_GetByName(t *testing.T) {
	orgID := "org-test-id-123"
	testRule := data.CreateTestDiscountRuleWithName(orgID, "Test Rule")

	tests := []struct {
		name      string
		orgID     string
		ruleName  string
		setupMock func(*mocks.MockDBManager)
		wantFound bool
		wantErr   bool
	}{
		{
			name:     "successful retrieval",
			orgID:    orgID,
			ruleName: "Test Rule",
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						rules := args.Get(2).(*[]*discounts.DiscountRule)
						*rules = []*discounts.DiscountRule{testRule}
					}).
					Return(nil)
			},
			wantFound: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewMockDBManager()
			tt.setupMock(mockDB)

			repo := discRepo.NewDiscountRuleRepository(mockDB)
			rule, err := repo.GetByName(context.Background(), tt.orgID, tt.ruleName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantFound {
					assert.NotNil(t, rule)
					assert.Equal(t, tt.ruleName, rule.Name)
				} else {
					assert.Nil(t, rule)
				}
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDiscountRuleRepository_UpdatePriority(t *testing.T) {
	testRule := data.CreateTestDiscountRule()
	testRule.ID = "rule-test-id-123"

	tests := []struct {
		name      string
		id        string
		priority  int
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name:     "successful priority update",
			id:       testRule.ID,
			priority: 10,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						rules := args.Get(2).(*[]*discounts.DiscountRule)
						*rules = []*discounts.DiscountRule{testRule}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						rule := args.Get(1).(*discounts.DiscountRule)
						assert.Equal(t, 10, rule.Priority)
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

			repo := discRepo.NewDiscountRuleRepository(mockDB)
			err := repo.UpdatePriority(context.Background(), tt.id, tt.priority)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDiscountRuleRepository_DeactivateRule(t *testing.T) {
	testRule := data.CreateTestDiscountRule()
	testRule.ID = "rule-test-id-123"

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)
		wantErr   bool
	}{
		{
			name: "successful deactivation",
			id:   testRule.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						rules := args.Get(2).(*[]*discounts.DiscountRule)
						*rules = []*discounts.DiscountRule{testRule}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						rule := args.Get(1).(*discounts.DiscountRule)
						assert.False(t, rule.IsActive)
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

			repo := discRepo.NewDiscountRuleRepository(mockDB)
			err := repo.DeactivateRule(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDiscountRuleRepository_ActivateRule(t *testing.T) {
	testRule := data.CreateTestInactiveDiscountRule()
	testRule.ID = "rule-test-id-123"

	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockDBManager)

		wantErr bool
	}{
		{
			name: "successful activation",
			id:   testRule.ID,
			setupMock: func(m *mocks.MockDBManager) {
				m.On("List", mock.Anything, mock.AnythingOfType("*base.Filter"), mock.AnythingOfType("*[]*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						rules := args.Get(2).(*[]*discounts.DiscountRule)
						*rules = []*discounts.DiscountRule{testRule}
					}).
					Return(nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*discounts.DiscountRule")).
					Run(func(args mock.Arguments) {
						rule := args.Get(1).(*discounts.DiscountRule)
						assert.True(t, rule.IsActive)
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

			repo := discRepo.NewDiscountRuleRepository(mockDB)
			err := repo.ActivateRule(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
