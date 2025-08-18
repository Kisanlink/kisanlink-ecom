# KisanLink E-commerce Extensible Models

This package provides extensible models for ecommerce functionality that can be imported and used by other services in the KisanLink ecosystem.

## Overview

The models are designed to be:
- **Extensible**: Easy to extend and customize for different use cases
- **Database-agnostic**: Work with both PostgreSQL and DynamoDB
- **Consistent**: Use the same base models and ID generation patterns
- **Reusable**: Can be imported by multiple services

## Available Models

### Catalog Models (`pkg/models/catalog/`)
- **CatalogItem**: Base catalog item with common fields
- **Product**: Physical products (seeds, fertilizers, tools)
- **ServiceOffering**: Services (drone spraying, tractor, harvester)
- **LabourOffering**: Manual labor services
- **InventoryLot**: Inventory tracking
- **ServiceSlot**: Service availability slots
- **LabourPool**: Available labor pool

### Pricing Models (`pkg/models/pricing/`)
- **Price**: Base pricing with validity periods
- **PriceTier**: Quantity-based pricing tiers
- **PriceRule**: Complex pricing rules and conditions

### Taxation Models (`pkg/models/taxation/`)
- **TaxRate**: Tax rate configurations (GST, VAT, etc.)
- **TaxRule**: Complex tax rules and exemptions
- **TaxExemption**: Tax exemptions and reductions
- **TaxCalculation**: Tax calculation results

### Discount Models (`pkg/models/discounts/`)
- **Discount**: Discount configurations and codes
- **DiscountRule**: Complex discount rules
- **DiscountUsage**: Usage tracking for discounts
- **DiscountCalculation**: Discount calculation results

## Usage

### Importing Models

```go
import (
    "github.com/Kisanlink/kisanlink-ecom/pkg/models/catalog"
    "github.com/Kisanlink/kisanlink-ecom/pkg/models/pricing"
    "github.com/Kisanlink/kisanlink-ecom/pkg/models/taxation"
    "github.com/Kisanlink/kisanlink-ecom/pkg/models/discounts"
)
```

### Creating New Instances

```go
// Create a new product
product := catalog.NewProduct(
    "org_123",
    "SEED001",
    "Organic Tomato Seeds",
    "Seeds",
    "kg",
    299.99,
)

// Create a new price
price := pricing.NewPrice(
    "org_123",
    "product",
    "org_123",
    pricing.PriceTypeBase,
    299.99,
    pricing.CurrencyINR,
)

// Create a new tax rate
taxRate := taxation.NewTaxRate(
    "org_123",
    taxation.TaxTypeGST,
    18.0,
)

// Create a new discount
discount := discounts.NewDiscount(
    "org_123",
    "SAVE20",
    "20% Off Seeds",
    discounts.DiscountTypePercentage,
    20.0,
)
```

### ID Generation

All models use the `kisanlink-db` package for consistent ID generation:

```go
// Models automatically generate IDs using the hash package
// Example IDs: CUST_12345678, CAT_87654321, PRICE_11223344
```

## Database Technology Decisions

### PostgreSQL (Relational)
- **Structured data** with complex relationships
- **ACID compliance** for transactions
- **Complex queries** and aggregations
- **Data integrity** and constraints

**Models**: Catalog, Pricing, Taxation, Discounts, Orders, Users

### DynamoDB (NoSQL)
- **High-frequency access** patterns
- **Simple key-value** storage
- **Temporary data** (sessions, cart)
- **High-volume writes** (audit logs)

**Models**: Sessions, Cart, Preferences, Cache, Audit Logs

## Extending Models

### Adding New Fields

```go
type CustomProduct struct {
    catalog.Product
    Brand        string  `json:"brand"`
    Weight       float64 `json:"weight"`
    IsOrganic    bool    `json:"is_organic"`
}
```

### Adding New Methods

```go
func (p *CustomProduct) CalculateShipping() float64 {
    if p.Weight > 5.0 {
        return 100.0
    }
    return 50.0
}
```

### Custom Validation

```go
func (p *CustomProduct) BeforeCreate() error {
    if err := p.Product.BeforeCreate(); err != nil {
        return err
    }
    
    if p.Brand == "" {
        return fmt.Errorf("brand is required")
    }
    
    return nil
}
```

## Best Practices

1. **Always use the provided constructors** (e.g., `NewProduct`, `NewPrice`)
2. **Extend models by embedding** rather than copying fields
3. **Use the validation methods** (`BeforeCreate`, `BeforeUpdate`, etc.)
4. **Follow the ID generation patterns** for consistency
5. **Consider database technology** when designing new models

## Dependencies

- `github.com/Kisanlink/kisanlink-db`: Base models and database management
- `github.com/Kisanlink/kisanlink-db/pkg/core/hash`: ID generation utilities

## Contributing

When adding new models:
1. Follow the existing patterns and conventions
2. Use appropriate hash sizes for ID generation
3. Include proper validation methods
4. Add database mapping configuration
5. Update this documentation

## License

This package is part of the KisanLink ecommerce system and follows the same licensing terms.
