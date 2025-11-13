# E-commerce RBAC Migrations

This folder contains migrations for seeding Role-Based Access Control (RBAC) data to the AAA service.

## Files

- `001_seed_ecommerce_rbac.go` - Seeds e-commerce business logic level RBAC data

## Usage

To seed the e-commerce RBAC data:

```bash
go run cmd/migrate/main.go --seed-rbac
```

## What Gets Seeded

### Business Actions (scoped to ecommerce module)

- `ecommerce.buy` - Purchase items from catalog
- `ecommerce.sell` - Sell items through catalog
- `ecommerce.manage_catalog` - Manage product catalog
- `ecommerce.manage_orders` - Manage customer orders
- `ecommerce.view_analytics` - View e-commerce analytics
- `ecommerce.approve_orders` - Approve customer orders
- `ecommerce.cancel_orders` - Cancel customer orders
- `ecommerce.refund_orders` - Process order refunds
- `ecommerce.manage_inventory` - Manage product inventory
- `ecommerce.process_payments` - Process payment transactions
- `ecommerce.manage_shipping` - Manage shipping and logistics
- `ecommerce.view_catalog` - View product catalog

### Business Resources (scoped to ecommerce module)

- `ecommerce.product` - Product catalog items
- `ecommerce.service` - Service catalog items
- `ecommerce.labor` - Labor catalog items
- `ecommerce.order` - Customer orders
- `ecommerce.customer` - Customer accounts
- `ecommerce.seller` - Seller accounts
- `ecommerce.organization` - E-commerce organizations
- `ecommerce.catalog` - Product catalog
- `ecommerce.inventory` - Inventory management
- `ecommerce.payment` - Payment processing
- `ecommerce.shipping` - Shipping and logistics
- `ecommerce.analytics` - E-commerce analytics and reporting

### Business Roles

- `ecommerce.buyer` - Basic purchasing permissions
- `ecommerce.seller` - Product and order management
- `ecommerce.admin` - Full business permissions
- `ecommerce.customer_support` - Order and customer management
- `ecommerce.inventory_manager` - Inventory and catalog management

## Module Scoping

All resources, actions, and permissions are properly scoped to the `ecommerce` module to avoid conflicts with other services in the AAA system.
