package customer

import (
	"context"
	"fmt"

	"kisanlink-ecom/internal/models/user"
	"kisanlink-ecom/internal/repositories/customer"
)

// CustomerServiceInterface defines the interface for customer service operations
type CustomerServiceInterface interface {
	CreateCustomer(ctx context.Context, aaaEntityID, customerCode string) (*user.Customer, error)
	GetCustomerByID(ctx context.Context, customerID string) (*user.Customer, error)
	GetCustomerByAAAEntityID(ctx context.Context, aaaEntityID string) (*user.Customer, error)
	GetCustomerByCustomerCode(ctx context.Context, customerCode string) (*user.Customer, error)
	UpdateCustomer(ctx context.Context, customerID string, updates map[string]interface{}) (*user.Customer, error)
	ListCustomers(ctx context.Context, limit, offset int, status string) ([]*user.Customer, error)
	DeleteCustomer(ctx context.Context, customerID string) error
}

// CustomerService handles customer-related business logic
type CustomerService struct {
	customerRepo *customer.CustomerRepository
}

// Ensure CustomerService implements CustomerServiceInterface
var _ CustomerServiceInterface = (*CustomerService)(nil)

// NewCustomerService creates a new customer service instance
func NewCustomerService(customerRepo *customer.CustomerRepository) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
	}
}

// CreateCustomer creates a new customer reference
func (s *CustomerService) CreateCustomer(ctx context.Context, aaaEntityID, customerCode string) (*user.Customer, error) {
	// Check if customer code already exists
	existing, err := s.customerRepo.GetByCustomerCode(ctx, customerCode)
	if err != nil {
		return nil, fmt.Errorf("failed to check customer code uniqueness: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("customer code %s already exists", customerCode)
	}

	// Create customer
	customer := user.NewCustomer(aaaEntityID, customerCode)

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

// GetCustomerByID retrieves a customer by ID
func (s *CustomerService) GetCustomerByID(ctx context.Context, customerID string) (*user.Customer, error) {
	customer := &user.Customer{}
	customer, err := s.customerRepo.GetByID(ctx, customerID, customer)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// GetCustomerByAAAEntityID retrieves a customer by aaa-service entity ID
func (s *CustomerService) GetCustomerByAAAEntityID(ctx context.Context, aaaEntityID string) (*user.Customer, error) {
	customer, err := s.customerRepo.GetByAAAEntityID(ctx, aaaEntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// GetCustomerByCustomerCode retrieves a customer by customer code
func (s *CustomerService) GetCustomerByCustomerCode(ctx context.Context, customerCode string) (*user.Customer, error) {
	customer, err := s.customerRepo.GetByCustomerCode(ctx, customerCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// UpdateCustomer updates an existing customer
func (s *CustomerService) UpdateCustomer(ctx context.Context, customerID string, updates map[string]interface{}) (*user.Customer, error) {
	// Get existing customer
	customer := &user.Customer{}
	customer, err := s.customerRepo.GetByID(ctx, customerID, customer)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Update fields
	if displayName, ok := updates["display_name"].(string); ok {
		customer.DisplayName = displayName
	}
	if email, ok := updates["email"].(string); ok {
		customer.Email = email
	}
	if phone, ok := updates["phone"].(string); ok {
		customer.Phone = phone
	}
	if status, ok := updates["status"].(string); ok {
		customer.Status = status
	}
	if isVerified, ok := updates["is_verified"].(bool); ok {
		customer.IsVerified = isVerified
	}

	// Update customer in repository
	if err := s.customerRepo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return customer, nil
}

// ListCustomers retrieves a list of customers
func (s *CustomerService) ListCustomers(ctx context.Context, limit, offset int, status string) ([]*user.Customer, error) {
	customers, err := s.customerRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	// Filter by status if specified
	if status != "" {
		var filtered []*user.Customer
		for _, customer := range customers {
			if customer.Status == status {
				filtered = append(filtered, customer)
			}
		}
		customers = filtered
	}

	return customers, nil
}

// DeleteCustomer soft deletes a customer
func (s *CustomerService) DeleteCustomer(ctx context.Context, customerID string) error {
	// Check if customer exists
	customer := &user.Customer{}
	_, err := s.customerRepo.GetByID(ctx, customerID, customer)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	// Soft delete customer
	if err := s.customerRepo.SoftDelete(ctx, customerID, "system"); err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	return nil
}
