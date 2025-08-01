package repositories

import (
	"context"
	"fmt"

	"kisanlink-ecom/internal/models/user"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// UserRepository extends BaseFilterableRepository with User-specific methods
type UserRepository struct {
	*base.BaseFilterableRepository[*user.User]
}

// NewUserRepository creates a new user repository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*user.User](),
	}
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	// Create a filter to find by email
	filter := base.NewFilter()
	filter.Group.Conditions = []base.FilterCondition{
		{
			Field:    "email",
			Operator: base.OpEqual,
			Value:    email,
		},
	}

	users, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user with email %s not found", email)
	}

	return users[0], nil
}

// Create overrides BaseRepository.Create to add email uniqueness check
func (r *UserRepository) Create(ctx context.Context, user *user.User) error {
	// Check email uniqueness
	existingUser, err := r.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	return r.BaseFilterableRepository.Create(ctx, user)
}
