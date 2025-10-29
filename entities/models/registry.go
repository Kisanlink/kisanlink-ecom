package models

import (
	"kisanlink-ecom/entities/models/actors"
	"kisanlink-ecom/entities/models/catalog"
	"kisanlink-ecom/entities/models/collaborator"
	"kisanlink-ecom/entities/models/common"
	"kisanlink-ecom/entities/models/discounts"
	"kisanlink-ecom/entities/models/marketplace"
	"kisanlink-ecom/entities/models/media"
	"kisanlink-ecom/entities/models/orders"
	"kisanlink-ecom/entities/models/outbox"
	"kisanlink-ecom/entities/models/pricing"
	"kisanlink-ecom/entities/models/roles"
	"kisanlink-ecom/entities/models/taxation"
	"kisanlink-ecom/entities/models/user"
)

// AllModels returns a slice of all GORM models for migration
// This ensures all models are registered with GORM AutoMigrate
func AllModels() []interface{} {
	return []interface{}{
		// User and Role models (no dependencies)
		&user.User{},
		&roles.EcommerceRole{},
		&roles.UserRole{},
		&roles.OrganizationRole{},

		// Catalog models
		&catalog.CatalogItem{},
		&catalog.Product{},
		&catalog.Service{},
		&catalog.Labour{},
		&catalog.Contract{},
		&catalog.Category{},
		&catalog.Variant{},
		&catalog.Availability{},
		&catalog.InventoryLot{},
		&catalog.InventoryAuditLog{},
		&catalog.SLA{},

		// Media models
		&media.Media{},

		// Pricing models
		&pricing.Price{},
		&pricing.PriceTier{},
		&pricing.PriceRule{},

		// Actor models
		&actors.Vendor{},
		&actors.Customer{},
		&actors.Collaborator{},
		&collaborator.Collaborator{},

		// Order models
		&orders.Order{},
		&orders.OrderItem{},
		&orders.OrderStatusHistory{},

		// Taxation models
		&taxation.TaxRate{},
		&taxation.TaxRule{},
		&taxation.TaxExemption{},

		// Discount models
		&discounts.Discount{},
		&discounts.DiscountUsage{},

		// Marketplace models
		&marketplace.Listing{},
		&marketplace.Bid{},
		&marketplace.AuctionEvent{},

		// Outbox models
		&outbox.OutboxEvent{},

		// Common models
		&common.SequenceCounter{},
		&common.AuditLog{},
	}
}

// CatalogModels returns catalog-specific models
func CatalogModels() []interface{} {
	return []interface{}{
		&catalog.CatalogItem{},
		&catalog.Product{},
		&catalog.Service{},
		&catalog.Labour{},
		&catalog.Contract{},
		&catalog.Category{},
		&catalog.Variant{},
		&catalog.Availability{},
		&catalog.InventoryLot{},
		&catalog.SLA{},
	}
}

// MediaModels returns media-specific models
func MediaModels() []interface{} {
	return []interface{}{
		&media.Media{},
	}
}

// PricingModels returns pricing-specific models
func PricingModels() []interface{} {
	return []interface{}{
		&pricing.Price{},
		&pricing.PriceTier{},
		&pricing.PriceRule{},
	}
}

// ActorModels returns actor-specific models
func ActorModels() []interface{} {
	return []interface{}{
		&actors.Vendor{},
		&actors.Customer{},
		&actors.Collaborator{},
		&collaborator.Collaborator{},
	}
}

// OrderModels returns order-specific models
func OrderModels() []interface{} {
	return []interface{}{
		&orders.Order{},
		&orders.OrderItem{},
		&orders.OrderStatusHistory{},
	}
}

// TaxationModels returns taxation-specific models
func TaxationModels() []interface{} {
	return []interface{}{
		&taxation.TaxRate{},
		&taxation.TaxRule{},
		&taxation.TaxExemption{},
	}
}

// DiscountModels returns discount-specific models
func DiscountModels() []interface{} {
	return []interface{}{
		&discounts.Discount{},
		&discounts.DiscountUsage{},
	}
}

// MarketplaceModels returns marketplace-specific models
func MarketplaceModels() []interface{} {
	return []interface{}{
		&marketplace.Listing{},
		&marketplace.Bid{},
		&marketplace.AuctionEvent{},
	}
}

// OutboxModels returns outbox-specific models
func OutboxModels() []interface{} {
	return []interface{}{
		&outbox.OutboxEvent{},
	}
}

// AuditModels returns audit-specific models
func AuditModels() []interface{} {
	return []interface{}{
		&common.SequenceCounter{},
		&common.AuditLog{},
		&catalog.InventoryAuditLog{},
	}
}

// UserModels returns user-related models
func UserModels() []interface{} {
	return []interface{}{
		&user.User{},
	}
}

// RoleModels returns role-related models
func RoleModels() []interface{} {
	return []interface{}{
		&roles.EcommerceRole{},
		&roles.UserRole{},
		&roles.OrganizationRole{},
	}
}

// ModelsByPriority returns models in migration order (dependencies first)
func ModelsByPriority() []interface{} {
	return []interface{}{
		// Base models first (no dependencies)
		&common.SequenceCounter{},

		// User and role models (no dependencies)
		&user.User{},
		&roles.EcommerceRole{},
		&roles.UserRole{},
		&roles.OrganizationRole{},

		// Actor models
		&actors.Vendor{},
		&actors.Customer{},
		&actors.Collaborator{},
		&collaborator.Collaborator{},
		&catalog.Category{},

		// Catalog models with foreign key dependencies
		&catalog.CatalogItem{},
		&catalog.Product{},
		&catalog.Service{},
		&catalog.Labour{},
		&catalog.Contract{},
		&catalog.Variant{},
		&catalog.Availability{},
		&catalog.InventoryLot{},
		&catalog.SLA{},
		&media.Media{},

		// Pricing models (depend on catalog)
		&pricing.Price{},
		&pricing.PriceTier{},
		&pricing.PriceRule{},

		// Taxation models (depend on catalog)
		&taxation.TaxRate{},
		&taxation.TaxRule{},
		&taxation.TaxExemption{},

		// Discount models (depend on catalog/pricing)
		&discounts.Discount{},
		&discounts.DiscountUsage{},

		// Order models (depend on catalog)
		&orders.Order{},
		&orders.OrderItem{},
		&orders.OrderStatusHistory{},

		// Marketplace models (depend on products)
		&marketplace.Listing{},
		&marketplace.Bid{},
		&marketplace.AuctionEvent{},

		// Outbox models (event processing)
		&outbox.OutboxEvent{},

		// Audit models last (references all other models)
		&catalog.InventoryAuditLog{},
		&common.AuditLog{},
	}
}
