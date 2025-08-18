package routes

import (
	"kisanlink-ecom/internal/handlers/customer"

	"github.com/gin-gonic/gin"
)

// SetupCustomerRoutes sets up customer-related routes
func SetupCustomerRoutes(router *gin.Engine, customerHandler *customer.CustomerHandler) {
	// Customer routes group
	customerGroup := router.Group("/api/v1/customers")
	{
		// Create customer
		customerGroup.POST("", customerHandler.CreateCustomer)

		// Get customer by ID
		customerGroup.GET("/:id", customerHandler.GetCustomer)

		// Get customer by AAA entity ID
		customerGroup.GET("/aaa-entity", customerHandler.GetCustomerByAAAEntityID)

		// Update customer
		customerGroup.PUT("/:id", customerHandler.UpdateCustomer)

		// List customers
		customerGroup.GET("", customerHandler.ListCustomers)

		// Delete customer
		customerGroup.DELETE("/:id", customerHandler.DeleteCustomer)
	}
}
