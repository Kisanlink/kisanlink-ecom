package collaborator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/collaborator"
	collaboratorRequests "github.com/Kisanlink/kisanlink-ecom/entities/requests/collaborator"
	"github.com/Kisanlink/kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// CollaboratorRepository handles collaborator operations using the database manager
type CollaboratorRepository struct {
	*common.BaseRepository
	dbManager db.DBManager
}

// NewCollaboratorRepository creates a new collaborator repository
func NewCollaboratorRepository(dbManager db.DBManager) *CollaboratorRepository {
	return &CollaboratorRepository{
		BaseRepository: common.NewBaseRepository(dbManager),
		dbManager:      dbManager,
	}
}

// Create creates a new collaborator in the database
func (r *CollaboratorRepository) Create(ctx context.Context, collab *collaborator.Collaborator) error {
	return r.dbManager.Create(ctx, collab)
}

// GetByID retrieves a collaborator by ID from the database
// Respects soft delete filtering based on context
func (r *CollaboratorRepository) GetByID(ctx context.Context, id string) (*collaborator.Collaborator, error) {
	opts := common.QueryOptionsFromContext(ctx)

	if opts.IncludeDeleted {
		// Include deleted items
		var collab collaborator.Collaborator
		if err := r.dbManager.GetByID(ctx, id, &collab); err != nil {
			return nil, fmt.Errorf("failed to get collaborator by ID: %w", err)
		}
		return &collab, nil
	}

	// Filter out deleted items - use Find with filter
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "id",
			Operator: base.OpEqual,
			Value:    id,
		},
	}

	collaborators, err := r.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborator by ID: %w", err)
	}

	if len(collaborators) == 0 {
		return nil, fmt.Errorf("collaborator not found")
	}

	return collaborators[0], nil
}

// Update updates an existing collaborator in the database
func (r *CollaboratorRepository) Update(ctx context.Context, collab *collaborator.Collaborator) error {
	return r.dbManager.Update(ctx, collab)
}

// SoftDelete soft deletes a collaborator in the database
func (r *CollaboratorRepository) SoftDelete(ctx context.Context, id string, deletedBy string) error {
	collab, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get collaborator for soft delete: %w", err)
	}

	collab.SetDeletedBy(&deletedBy)
	return r.Update(ctx, collab)
}

// Find retrieves collaborators using filters
// Respects soft delete filtering based on context
func (r *CollaboratorRepository) Find(ctx context.Context, filter *base.Filter) ([]*collaborator.Collaborator, error) {
	var collaborators []*collaborator.Collaborator

	// Apply query options to filter
	enhancedFilter := r.ApplyQueryOptions(ctx, filter)

	if err := r.dbManager.List(ctx, enhancedFilter, &collaborators); err != nil {
		return nil, fmt.Errorf("failed to find collaborators: %w", err)
	}
	return collaborators, nil
}

// Count returns the count of collaborators matching the filter
// Respects soft delete filtering based on context
func (r *CollaboratorRepository) Count(ctx context.Context, filter *base.Filter) (int64, error) {
	var collab collaborator.Collaborator

	// Apply query options to filter
	enhancedFilter := r.ApplyQueryOptions(ctx, filter)

	return r.dbManager.Count(ctx, enhancedFilter, &collab)
}

// GetByUserID retrieves a collaborator by user ID
func (r *CollaboratorRepository) GetByUserID(ctx context.Context, userID string) (*collaborator.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "user_id",
			Operator: base.OpEqual,
			Value:    userID,
		},
	}

	collaborators, err := r.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborator by user ID: %w", err)
	}

	if len(collaborators) == 0 {
		return nil, nil
	}

	return collaborators[0], nil
}

// GetByUsername retrieves a collaborator by username
func (r *CollaboratorRepository) GetByUsername(ctx context.Context, username string) (*collaborator.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "username",
			Operator: base.OpEqual,
			Value:    username,
		},
	}

	collaborators, err := r.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborator by username: %w", err)
	}

	if len(collaborators) == 0 {
		return nil, nil
	}

	return collaborators[0], nil
}

// GetByEmail retrieves a collaborator by email
func (r *CollaboratorRepository) GetByEmail(ctx context.Context, email string) (*collaborator.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "email",
			Operator: base.OpEqual,
			Value:    email,
		},
	}

	collaborators, err := r.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborator by email: %w", err)
	}

	if len(collaborators) == 0 {
		return nil, nil
	}

	return collaborators[0], nil
}

// GetByOrganizationID retrieves collaborators by organization ID
func (r *CollaboratorRepository) GetByOrganizationID(ctx context.Context, orgID string, limit, offset int) ([]*collaborator.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "organization_id",
			Operator: base.OpEqual,
			Value:    orgID,
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}

// GetByType retrieves collaborators by type
func (r *CollaboratorRepository) GetByType(ctx context.Context, collaboratorType collaborator.CollaboratorType, limit, offset int) ([]*collaborator.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "collaborator_type",
			Operator: base.OpEqual,
			Value:    string(collaboratorType),
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}

// GetByStatus retrieves collaborators by status
func (r *CollaboratorRepository) GetByStatus(ctx context.Context, status collaborator.CollaboratorStatus, limit, offset int) ([]*collaborator.Collaborator, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(status),
		},
	}
	filter.Limit = limit
	filter.Offset = offset

	return r.Find(ctx, filter)
}

// ListCollaborators retrieves collaborators with filtering and pagination
func (r *CollaboratorRepository) ListCollaborators(ctx context.Context, filter *collaboratorRequests.CollaboratorFilter, offset, limit int) ([]*collaborator.Collaborator, int, error) {
	dbFilter := base.NewFilter()

	// Organization scoping
	if filter.OrganizationID != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "organization_id",
			Operator: base.OpEqual,
			Value:    *filter.OrganizationID,
		})
	}

	// Collaborator type
	if filter.CollaboratorType != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "collaborator_type",
			Operator: base.OpEqual,
			Value:    string(*filter.CollaboratorType),
		})
	}

	// Status
	if filter.Status != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "status",
			Operator: base.OpEqual,
			Value:    string(*filter.Status),
		})
	}

	// Verification status
	if filter.IsVerified != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "is_verified",
			Operator: base.OpEqual,
			Value:    *filter.IsVerified,
		})
	}

	// Onboarding completion
	if filter.OnboardingCompleted != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "onboarding_completed",
			Operator: base.OpEqual,
			Value:    *filter.OnboardingCompleted,
		})
	}

	// Location
	if filter.Location != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "location",
			Operator: base.OpContains,
			Value:    *filter.Location,
		})
	}

	// Business type
	if filter.BusinessType != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "business_type",
			Operator: base.OpEqual,
			Value:    *filter.BusinessType,
		})
	}

	// Trust score range
	if filter.MinTrustScore != nil && filter.MaxTrustScore != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "trust_score",
			Operator: base.OpDateBetween,
			Value:    *filter.MinTrustScore,
			Value2:   *filter.MaxTrustScore,
		})
	} else if filter.MinTrustScore != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "trust_score",
			Operator: base.OpGreaterEqual,
			Value:    *filter.MinTrustScore,
		})
	} else if filter.MaxTrustScore != nil {
		dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
			Field:    "trust_score",
			Operator: base.OpLessEqual,
			Value:    *filter.MaxTrustScore,
		})
	}

	// Date range filters
	if filter.CreatedAfter != nil {
		if createdAfter, err := time.Parse(time.RFC3339, *filter.CreatedAfter); err == nil {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "created_at",
				Operator: base.OpGreaterEqual,
				Value:    createdAfter,
			})
		}
	}

	if filter.CreatedBefore != nil {
		if createdBefore, err := time.Parse(time.RFC3339, *filter.CreatedBefore); err == nil {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "created_at",
				Operator: base.OpLessEqual,
				Value:    createdBefore,
			})
		}
	}

	if filter.LastActiveAfter != nil {
		if lastActiveAfter, err := time.Parse(time.RFC3339, *filter.LastActiveAfter); err == nil {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "last_activity_at",
				Operator: base.OpGreaterEqual,
				Value:    lastActiveAfter,
			})
		}
	}

	if filter.LastActiveBefore != nil {
		if lastActiveBefore, err := time.Parse(time.RFC3339, *filter.LastActiveBefore); err == nil {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "last_activity_at",
				Operator: base.OpLessEqual,
				Value:    lastActiveBefore,
			})
		}
	}

	// Tags filter (contains any)
	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			dbFilter.Group.Conditions = append(dbFilter.Group.Conditions, base.FilterCondition{
				Field:    "tags",
				Operator: base.OpContains,
				Value:    tag,
			})
		}
	}

	// Search across multiple fields
	if filter.Search != nil && *filter.Search != "" {
		search := *filter.Search
		dbFilter.Group.Groups = append(dbFilter.Group.Groups, base.FilterGroup{
			Logic: base.LogicOr,
			Conditions: []base.FilterCondition{
				{Field: "first_name", Operator: base.OpContains, Value: search},
				{Field: "last_name", Operator: base.OpContains, Value: search},
				{Field: "username", Operator: base.OpContains, Value: search},
				{Field: "email", Operator: base.OpContains, Value: search},
				{Field: "business_name", Operator: base.OpContains, Value: search},
			},
		})
	}

	dbFilter.Limit = limit
	dbFilter.Offset = offset

	collaborators, err := r.Find(ctx, dbFilter)
	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination
	total, err := r.Count(ctx, dbFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return collaborators, int(total), nil
}

// GetCollaboratorStats retrieves statistics about collaborators
func (r *CollaboratorRepository) GetCollaboratorStats(ctx context.Context, orgID *string) (*CollaboratorStats, error) {
	baseFilter := base.NewFilter()
	if orgID != nil {
		baseFilter.Group.Conditions = append(baseFilter.Group.Conditions, base.FilterCondition{
			Field:    "organization_id",
			Operator: base.OpEqual,
			Value:    *orgID,
		})
	}

	var collab collaborator.Collaborator

	// Total collaborators
	total, err := r.dbManager.Count(ctx, baseFilter, &collab)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Active collaborators
	activeFilter := *baseFilter
	activeFilter.Group.Conditions = append(activeFilter.Group.Conditions, base.FilterCondition{
		Field:    "status",
		Operator: base.OpEqual,
		Value:    string(collaborator.CollaboratorStatusActive),
	})
	active, err := r.dbManager.Count(ctx, &activeFilter, &collab)
	if err != nil {
		return nil, fmt.Errorf("failed to get active count: %w", err)
	}

	// Pending collaborators
	pendingFilter := *baseFilter
	pendingFilter.Group.Conditions = append(pendingFilter.Group.Conditions, base.FilterCondition{
		Field:    "status",
		Operator: base.OpEqual,
		Value:    string(collaborator.CollaboratorStatusPending),
	})
	pending, err := r.dbManager.Count(ctx, &pendingFilter, &collab)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending count: %w", err)
	}

	// Verified collaborators
	verifiedFilter := *baseFilter
	verifiedFilter.Group.Conditions = append(verifiedFilter.Group.Conditions, base.FilterCondition{
		Field:    "is_verified",
		Operator: base.OpEqual,
		Value:    true,
	})
	verified, err := r.dbManager.Count(ctx, &verifiedFilter, &collab)
	if err != nil {
		return nil, fmt.Errorf("failed to get verified count: %w", err)
	}

	// Recent registrations (last 7 days)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	recentFilter := *baseFilter
	recentFilter.Group.Conditions = append(recentFilter.Group.Conditions, base.FilterCondition{
		Field:    "created_at",
		Operator: base.OpGreaterEqual,
		Value:    sevenDaysAgo,
	})
	recent, err := r.dbManager.Count(ctx, &recentFilter, &collab)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent count: %w", err)
	}

	// Active in last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	activeRecentFilter := *baseFilter
	activeRecentFilter.Group.Conditions = append(activeRecentFilter.Group.Conditions, base.FilterCondition{
		Field:    "last_activity_at",
		Operator: base.OpGreaterEqual,
		Value:    thirtyDaysAgo,
	})
	activeRecent, err := r.dbManager.Count(ctx, &activeRecentFilter, &collab)
	if err != nil {
		return nil, fmt.Errorf("failed to get active recent count: %w", err)
	}

	// TODO: Implement more detailed statistics like collaborators by type, status, etc.
	// This would require more complex queries or multiple database calls

	stats := &CollaboratorStats{
		TotalCollaborators:    int(total),
		ActiveCollaborators:   int(active),
		PendingCollaborators:  int(pending),
		VerifiedCollaborators: int(verified),
		RecentRegistrations:   int(recent),
		ActiveInLast30Days:    int(activeRecent),
	}

	// Calculate onboarding completion rate
	if total > 0 {
		onboardingFilter := *baseFilter
		onboardingFilter.Group.Conditions = append(onboardingFilter.Group.Conditions, base.FilterCondition{
			Field:    "onboarding_completed",
			Operator: base.OpEqual,
			Value:    true,
		})
		onboardingCompleted, err := r.dbManager.Count(ctx, &onboardingFilter, &collab)
		if err == nil {
			stats.OnboardingCompletion = float64(onboardingCompleted) / float64(total) * 100
		}
	}

	return stats, nil
}

// UpdateLastActivity updates the last activity timestamp for a collaborator
func (r *CollaboratorRepository) UpdateLastActivity(ctx context.Context, id string) error {
	collab, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get collaborator: %w", err)
	}

	collab.UpdateLastActivity()
	return r.Update(ctx, collab)
}

// UpdateLastLogin updates the last login timestamp and increments login count
func (r *CollaboratorRepository) UpdateLastLogin(ctx context.Context, id string) error {
	collab, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get collaborator: %w", err)
	}

	collab.UpdateLastLogin()
	return r.Update(ctx, collab)
}

// BulkUpdateStatus updates the status of multiple collaborators
func (r *CollaboratorRepository) BulkUpdateStatus(ctx context.Context, ids []string, status collaborator.CollaboratorStatus, updatedBy string) (int, []string, error) {
	successful := 0
	failed := []string{}

	for _, id := range ids {
		collab, err := r.GetByID(ctx, id)
		if err != nil {
			failed = append(failed, id)
			continue
		}

		collab.Status = status
		collab.SetUpdatedBy(updatedBy)

		if err := r.Update(ctx, collab); err != nil {
			failed = append(failed, id)
			continue
		}

		successful++
	}

	return successful, failed, nil
}

// GetCollaboratorsByIDs retrieves multiple collaborators by their IDs
func (r *CollaboratorRepository) GetCollaboratorsByIDs(ctx context.Context, ids []string) ([]*collaborator.Collaborator, error) {
	if len(ids) == 0 {
		return []*collaborator.Collaborator{}, nil
	}

	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "id",
			Operator: base.OpIn,
			Value:    ids,
		},
	}

	return r.Find(ctx, filter)
}

// SearchCollaborators performs a full-text search on collaborators
func (r *CollaboratorRepository) SearchCollaborators(ctx context.Context, query string, limit, offset int) ([]*collaborator.Collaborator, int, error) {
	filter := base.NewFilter()

	// Search across multiple fields
	searchTerms := strings.Fields(strings.ToLower(query))
	if len(searchTerms) > 0 {
		searchGroup := base.FilterGroup{
			Logic:      base.LogicOr,
			Conditions: []base.FilterCondition{},
		}

		for _, term := range searchTerms {
			searchGroup.Conditions = append(searchGroup.Conditions,
				base.FilterCondition{Field: "first_name", Operator: base.OpContains, Value: term},
				base.FilterCondition{Field: "last_name", Operator: base.OpContains, Value: term},
				base.FilterCondition{Field: "username", Operator: base.OpContains, Value: term},
				base.FilterCondition{Field: "email", Operator: base.OpContains, Value: term},
				base.FilterCondition{Field: "business_name", Operator: base.OpContains, Value: term},
				base.FilterCondition{Field: "location", Operator: base.OpContains, Value: term},
			)
		}

		filter.Group.Groups = append(filter.Group.Groups, searchGroup)
	}

	filter.Limit = limit
	filter.Offset = offset

	collaborators, err := r.Find(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Get total count for pagination
	var collab collaborator.Collaborator
	total, err := r.dbManager.Count(ctx, filter, &collab)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return collaborators, int(total), nil
}

// Helper struct for statistics (this should be moved to a proper place)
type CollaboratorStats struct {
	TotalCollaborators    int     `json:"total_collaborators"`
	ActiveCollaborators   int     `json:"active_collaborators"`
	PendingCollaborators  int     `json:"pending_collaborators"`
	VerifiedCollaborators int     `json:"verified_collaborators"`
	RecentRegistrations   int     `json:"recent_registrations_7_days"`
	ActiveInLast30Days    int     `json:"active_in_last_30_days"`
	OnboardingCompletion  float64 `json:"onboarding_completion_rate"`
}
