package models

import (
	"kisanlink-ecom/entities/models/actors"
	"kisanlink-ecom/entities/models/catalog"
	"kisanlink-ecom/entities/models/common"
	"kisanlink-ecom/entities/models/media"
	"kisanlink-ecom/entities/models/pricing"
)

// AllModels returns a slice of all GORM models for migration
// This ensures all models are registered with GORM AutoMigrate
func AllModels() []interface{} {
	return []interface{}{
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

		// Common models
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
	}
}

// AuditModels returns audit-specific models
func AuditModels() []interface{} {
	return []interface{}{
		&common.AuditLog{},
	}
}

// ModelsByPriority returns models in migration order (dependencies first)
func ModelsByPriority() []interface{} {
	return []interface{}{
		// Base models first (no dependencies)
		&actors.Vendor{},
		&actors.Customer{},
		&actors.Collaborator{},
		&catalog.Category{},

		// Models with foreign key dependencies
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

		// Pricing models
		&pricing.Price{},
		&pricing.PriceTier{},
		&pricing.PriceRule{},

		// Audit models last (references all other models)
		&common.AuditLog{},
	}
}
