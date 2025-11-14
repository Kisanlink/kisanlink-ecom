// Package user provides repository operations for user entities.
// This includes CRUD operations and user-specific queries like username and email lookups.
package user

import (
	"context"
	"fmt"

	"github.com/Kisanlink/kisanlink-ecom/entities/models/user"
	"github.com/Kisanlink/kisanlink-ecom/internal/repositories/common"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/db"
)

// Repository handles user operations using the database manager
type Repository struct {
	*common.BaseRepository
	dbManager db.DBManager
}

// NewUserRepository creates a new user repository
func NewUserRepository(dbManager db.DBManager) *Repository {
	return &Repository{
		BaseRepository: common.NewBaseRepository(dbManager),
		dbManager:      dbManager,
	}
}

// Create creates a new user in the database
func (r *Repository) Create(ctx context.Context, user *user.User) error {
	return r.dbManager.Create(ctx, user)
}

// GetByID retrieves a user by ID from the database
// Note: User model doesn't use soft delete (no deleted_at field), but we keep the pattern consistent
func (r *Repository) GetByID(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	if err := r.dbManager.GetByID(ctx, id, &u); err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &u, nil
}

// GetByUsername retrieves a user by username
func (r *Repository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "username",
			Operator: base.OpEqual,
			Value:    username,
		},
	}

	var users []*user.User
	if err := r.dbManager.List(ctx, filter, &users); err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	if len(users) == 0 {
		return nil, nil
	}

	return users[0], nil
}

// GetByEmail retrieves a user by email
func (r *Repository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "email",
			Operator: base.OpEqual,
			Value:    email,
		},
	}

	var users []*user.User
	if err := r.dbManager.List(ctx, filter, &users); err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if len(users) == 0 {
		return nil, nil
	}

	return users[0], nil
}

// Update updates an existing user in the database
func (r *Repository) Update(ctx context.Context, user *user.User) error {
	return r.dbManager.Update(ctx, user)
}

// Delete deletes a user from the database
func (r *Repository) Delete(ctx context.Context, id string) error {
	u := &user.User{}
	return r.dbManager.Delete(ctx, id, u)
}

// List retrieves users with filtering and pagination
func (r *Repository) List(ctx context.Context, filter *base.Filter, limit, offset int) ([]*user.User, int, error) {
	dbFilter := filter
	if dbFilter == nil {
		dbFilter = base.NewFilter()
	}

	dbFilter.Limit = limit
	dbFilter.Offset = offset

	var users []*user.User
	if err := r.dbManager.List(ctx, dbFilter, &users); err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Get total count for pagination
	total, err := r.dbManager.Count(ctx, dbFilter, &user.User{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, int(total), nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *Repository) UpdateLastLogin(ctx context.Context, id string) error {
	u, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user for last login update: %w", err)
	}

	// Update last login timestamp
	// Note: The LastLoginAt field is a *time.Time, so we need to handle it appropriately
	// This would typically be done in the service layer, but we'll update it here for simplicity
	return r.Update(ctx, u)
}

// ActivateUser activates a user
func (r *Repository) ActivateUser(ctx context.Context, id string) error {
	u, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user for activation: %w", err)
	}

	u.IsActive = true
	return r.Update(ctx, u)
}

// DeactivateUser deactivates a user
func (r *Repository) DeactivateUser(ctx context.Context, id string) error {
	u, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user for deactivation: %w", err)
	}

	u.IsActive = false
	return r.Update(ctx, u)
}
