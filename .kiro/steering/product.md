---
inclusion: always
---

# KisanLink E-Commerce API

A modern, modular e-commerce API built for the agricultural sector, focusing on products, services, and labor marketplace functionality.

## Domain Context

This is an agricultural e-commerce platform with three core catalog types:

- **Products**: Physical agricultural goods (seeds, fertilizers, equipment)
- **Services**: Professional agricultural services (consultation, maintenance)
- **Labor**: Agricultural workforce marketplace

## Code Conventions

### Naming Patterns

- Use descriptive, domain-specific names: `CatalogItem`, `InventoryLot`, `OrderStatus`
- Prefix interfaces with `I`: `ICatalogService`, `IOrderRepository`
- Use agricultural terminology where appropriate: `Harvest`, `Season`, `Crop`

### API Design Principles

- RESTful endpoints following `/api/v1/{domain}/{resource}` pattern
- Consistent response format using `utils.SuccessResponse()` and `utils.ErrorResponse()`
- Pagination support for all list endpoints using `common.PaginationRequest`
- Include Swagger annotations for all public endpoints

### Business Logic Rules

- All catalog items must have valid pricing and availability
- Orders require authentication and proper authorization
- Inventory tracking is mandatory for physical products
- Service bookings require time slot validation
- Labor contracts need skill verification

### Error Handling Standards

- Use structured error responses with appropriate HTTP status codes
- Include validation details for 400 Bad Request responses
- Log errors with context using logrus structured logging
- Return user-friendly messages while logging technical details

### Database Patterns

- Use kisanlink-db package abstractions for multi-backend support
- All models extend `base.BaseModel` for consistent timestamps and IDs
- Implement soft deletes for audit trails
- Use transactions for multi-table operations

### Authentication & Authorization

- JWT tokens validated through external AAA service
- Role-based access control with agricultural-specific roles
- Mock authentication in development when AAA service unavailable
- Always validate permissions before business operations
