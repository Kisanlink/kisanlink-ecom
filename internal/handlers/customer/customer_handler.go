package customer

import (
	"net/http"
	"strconv"

	"kisanlink-ecom/internal/models/common"
	"kisanlink-ecom/internal/services/customer"

	"github.com/gin-gonic/gin"
)

// CustomerHandler handles HTTP requests for customer operations
type CustomerHandler struct {
	customerService customer.CustomerServiceInterface
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(customerService customer.CustomerServiceInterface) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
	}
}

// CreateCustomerRequest represents the request to create a customer
type CreateCustomerRequest struct {
	AAAEntityID  string `json:"aaa_entity_id" binding:"required" example:"user_123"`
	CustomerCode string `json:"customer_code" binding:"required" example:"CUST001"`
}

// UpdateCustomerRequest represents the request to update a customer
type UpdateCustomerRequest struct {
	DisplayName *string `json:"display_name,omitempty" example:"John Doe"`
	Email       *string `json:"email,omitempty" example:"john@example.com"`
	Phone       *string `json:"phone,omitempty" example:"+1234567890"`
	Status      *string `json:"status,omitempty" example:"active"`
	IsVerified  *bool   `json:"is_verified,omitempty" example:"true"`
}

// CreateCustomer creates a new customer
// @Summary Create a new customer
// @Description Create a new customer reference to aaa-service user
// @Tags customers
// @Accept json
// @Produce json
// @Param customer body CreateCustomerRequest true "Customer creation request"
// @Success 201 {object} common.APIResponse
// @Failure 400 {object} common.APIResponse
// @Failure 409 {object} common.APIResponse
// @Failure 500 {object} common.APIResponse
// @Router /api/v1/customers [post]
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.APIResponse{
			Success: false,
			Message: "Invalid request",
			Error: &common.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid request",
				Details: err.Error(),
			},
		})
		return
	}

	customer, err := h.customerService.CreateCustomer(c.Request.Context(), req.AAAEntityID, req.CustomerCode)
	if err != nil {
		if err.Error() == "customer code already exists" {
			c.JSON(http.StatusConflict, common.APIResponse{
				Success: false,
				Message: "Conflict",
				Error: &common.APIError{
					Code:    "CONFLICT",
					Message: "Customer code already exists",
					Details: err.Error(),
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, common.APIResponse{
			Success: false,
			Message: "Internal server error",
			Error: &common.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, common.APIResponse{
		Success: true,
		Message: "Customer created successfully",
		Data:    customer,
	})
}

// GetCustomer retrieves a customer by ID
// @Summary Get customer by ID
// @Description Retrieve a customer by their ID
// @Tags customers
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} common.APIResponse
// @Failure 404 {object} common.APIResponse
// @Failure 500 {object} common.APIResponse
// @Router /api/v1/customers/{id} [get]
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	customerID := c.Param("id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, common.APIResponse{
			Success: false,
			Message: "Bad request",
			Error: &common.APIError{
				Code:    "BAD_REQUEST",
				Message: "Customer ID is required",
			},
		})
		return
	}

	customer, err := h.customerService.GetCustomerByID(c.Request.Context(), customerID)
	if err != nil {
		if err.Error() == "customer not found" {
			c.JSON(http.StatusNotFound, common.APIResponse{
				Success: false,
				Message: "Not found",
				Error: &common.APIError{
					Code:    "NOT_FOUND",
					Message: "Customer not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, common.APIResponse{
			Success: false,
			Message: "Internal server error",
			Error: &common.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, common.APIResponse{
		Success: true,
		Message: "Customer retrieved successfully",
		Data:    customer,
	})
}

// GetCustomerByAAAEntityID retrieves a customer by aaa-service entity ID
// @Summary Get customer by AAA entity ID
// @Description Retrieve a customer by their aaa-service entity ID
// @Tags customers
// @Produce json
// @Param aaa_entity_id query string true "AAA Service Entity ID"
// @Success 200 {object} common.APIResponse
// @Failure 404 {object} common.APIResponse
// @Failure 500 {object} common.APIResponse
// @Router /api/v1/customers/aaa-entity [get]
func (h *CustomerHandler) GetCustomerByAAAEntityID(c *gin.Context) {
	aaaEntityID := c.Query("aaa_entity_id")
	if aaaEntityID == "" {
		c.JSON(http.StatusBadRequest, common.APIResponse{
			Success: false,
			Message: "Bad request",
			Error: &common.APIError{
				Code:    "BAD_REQUEST",
				Message: "AAA entity ID is required",
			},
		})
		return
	}

	customer, err := h.customerService.GetCustomerByAAAEntityID(c.Request.Context(), aaaEntityID)
	if err != nil {
		if err.Error() == "customer not found" {
			c.JSON(http.StatusNotFound, common.APIResponse{
				Success: false,
				Message: "Not found",
				Error: &common.APIError{
					Code:    "NOT_FOUND",
					Message: "Customer not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, common.APIResponse{
			Success: false,
			Message: "Internal server error",
			Error: &common.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, common.APIResponse{
		Success: true,
		Message: "Customer retrieved successfully",
		Data:    customer,
	})
}

// UpdateCustomer updates an existing customer
// @Summary Update customer
// @Description Update an existing customer's information
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param customer body UpdateCustomerRequest true "Customer update request"
// @Success 200 {object} common.APIResponse
// @Failure 400 {object} common.APIResponse
// @Failure 404 {object} common.APIResponse
// @Failure 500 {object} common.APIResponse
// @Router /api/v1/customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	customerID := c.Param("id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, common.APIResponse{
			Success: false,
			Message: "Bad request",
			Error: &common.APIError{
				Code:    "BAD_REQUEST",
				Message: "Customer ID is required",
			},
		})
		return
	}

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.APIResponse{
			Success: false,
			Message: "Invalid request",
			Error: &common.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid request",
				Details: err.Error(),
			},
		})
		return
	}

	// Convert request to updates map
	updates := make(map[string]interface{})
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.IsVerified != nil {
		updates["is_verified"] = *req.IsVerified
	}

	customer, err := h.customerService.UpdateCustomer(c.Request.Context(), customerID, updates)
	if err != nil {
		if err.Error() == "customer not found" {
			c.JSON(http.StatusNotFound, common.APIResponse{
				Success: false,
				Message: "Not found",
				Error: &common.APIError{
					Code:    "NOT_FOUND",
					Message: "Customer not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, common.APIResponse{
			Success: false,
			Message: "Internal server error",
			Error: &common.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, common.APIResponse{
		Success: true,
		Message: "Customer updated successfully",
		Data:    customer,
	})
}

// ListCustomers retrieves a list of customers
// @Summary List customers
// @Description Retrieve a list of customers with optional filtering
// @Tags customers
// @Produce json
// @Param limit query int false "Number of customers to return (default: 10, max: 100)"
// @Param offset query int false "Number of customers to skip (default: 0)"
// @Param status query string false "Filter by status"
// @Success 200 {object} common.APIResponse
// @Failure 500 {object} common.APIResponse
// @Router /api/v1/customers [get]
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	status := c.Query("status")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	customers, err := h.customerService.ListCustomers(c.Request.Context(), limit, offset, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.APIResponse{
			Success: false,
			Message: "Internal server error",
			Error: &common.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, common.APIResponse{
		Success: true,
		Message: "Customers retrieved successfully",
		Data:    customers,
	})
}

// DeleteCustomer deletes a customer
// @Summary Delete customer
// @Description Soft delete a customer
// @Tags customers
// @Produce json
// @Param id path string true "Customer ID"
// @Success 204 "No content"
// @Failure 400 {object} common.APIResponse
// @Failure 404 {object} common.APIResponse
// @Failure 500 {object} common.APIResponse
// @Router /api/v1/customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	customerID := c.Param("id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, common.APIResponse{
			Success: false,
			Message: "Bad request",
			Error: &common.APIError{
				Code:    "BAD_REQUEST",
				Message: "Customer ID is required",
			},
		})
		return
	}

	err := h.customerService.DeleteCustomer(c.Request.Context(), customerID)
	if err != nil {
		if err.Error() == "customer not found" {
			c.JSON(http.StatusNotFound, common.APIResponse{
				Success: false,
				Message: "Not found",
				Error: &common.APIError{
					Code:    "NOT_FOUND",
					Message: "Customer not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, common.APIResponse{
			Success: false,
			Message: "Internal server error",
			Error: &common.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
				Details: err.Error(),
			},
		})
		return
	}

	c.Status(http.StatusNoContent)
}
