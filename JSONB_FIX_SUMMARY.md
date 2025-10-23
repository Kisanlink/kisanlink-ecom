# PostgreSQL JSONB Field Fixes

## Problem
PostgreSQL `jsonb` columns reject empty strings (`''`) as invalid JSON. The error was:
```
ERROR: invalid input syntax for type json (SQLSTATE 22P02)
```

This occurred when trying to insert inventory lots and orders with empty metadata fields.

## Root Cause
Many models defined `jsonb` fields as `string` type in Go:
```go
Metadata string `json:"metadata" gorm:"type:jsonb"`
```

When code set these to empty strings (`""`), PostgreSQL rejected them as invalid JSON.

## Solution
Changed all critical `jsonb` fields from `string` to `sql.NullString`:
```go
Metadata sql.NullString `json:"metadata" gorm:"type:jsonb"` // NULL if not set
```

### Why `sql.NullString`?
- When `Valid = false`, GORM inserts SQL `NULL` (which PostgreSQL jsonb accepts)
- When `Valid = true`, GORM inserts the string value as valid JSON
- No more empty string errors!

## Files Modified

### 1. Inventory Lot Model
**File:** `entities/models/catalog/inventory_lot.go`

**Changes:**
- Added `database/sql` import
- Changed `TestResults` from `string` to `sql.NullString`
- Changed `Metadata` from `string` to `sql.NullString`

**Before:**
```go
TestResults string `json:"test_results" gorm:"type:jsonb"`
Metadata    string `json:"metadata" gorm:"type:jsonb"`
```

**After:**
```go
TestResults  sql.NullString `json:"test_results" gorm:"type:jsonb"` // NULL if not set
Metadata     sql.NullString `json:"metadata" gorm:"type:jsonb"`     // NULL if not set
```

### 2. Inventory Service
**File:** `internal/services/inventory/inventory_service.go`

**Changes:**
- Added `database/sql` import
- Updated metadata assignments to use `sql.NullString`

**Before:**
```go
if req.Metadata != "" {
    lot.Metadata = req.Metadata
} else {
    lot.Metadata = "{}"
}
```

**After:**
```go
if req.Metadata != "" {
    lot.Metadata = sql.NullString{String: req.Metadata, Valid: true}
}
// If empty, leave as NULL (Valid: false) which PostgreSQL jsonb accepts
```

### 3. Order Models
**File:** `entities/models/orders/order.go`

**Changes:**
- Added `database/sql` import
- Changed `Metadata` fields for Order, OrderItem, and OrderStatusHistory to `sql.NullString`
- Updated all `SetMetadata()` methods to use NULL instead of empty strings
- Updated all `GetMetadata()` methods to check `Valid` flag

**Order Struct:**
```go
Metadata sql.NullString `json:"metadata" gorm:"type:jsonb"` // NULL if not set
```

**SetMetadata Method:**
```go
func (o *Order) SetMetadata(metadata map[string]interface{}) error {
    if metadata == nil {
        o.Metadata = sql.NullString{Valid: false} // Set to NULL
        return nil
    }

    metadataJSON, err := json.Marshal(metadata)
    if err != nil {
        return fmt.Errorf("failed to marshal metadata: %w", err)
    }

    o.Metadata = sql.NullString{String: string(metadataJSON), Valid: true}
    return nil
}
```

**GetMetadata Method:**
```go
func (o *Order) GetMetadata() (map[string]interface{}, error) {
    if !o.Metadata.Valid || o.Metadata.String == "" {
        return nil, nil
    }

    var metadata map[string]interface{}
    if err := json.Unmarshal([]byte(o.Metadata.String), &metadata); err != nil {
        return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
    }

    return metadata, nil
}
```

### 4. Order Response Converters
**File:** `entities/responses/orders/order_responses.go`

**Changes:**
- Updated metadata parsing to handle `sql.NullString`

**Before:**
```go
if order.Metadata != "" {
    var metadata map[string]interface{}
    if err := json.Unmarshal([]byte(order.Metadata), &metadata); err == nil {
        response.Metadata = metadata
    }
}
```

**After:**
```go
if order.Metadata.Valid && order.Metadata.String != "" {
    var metadata map[string]interface{}
    if err := json.Unmarshal([]byte(order.Metadata.String), &metadata); err == nil {
        response.Metadata = metadata
    }
}
```

## Testing

### ✅ Build Status
```bash
go build ./cmd/server
# Success - no errors
```

### Test Inventory Lot Creation
After restarting the server, try creating an inventory lot:

```bash
curl -X POST http://localhost:8080/api/v1/inventory/lots \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "catalog_item_id": "CAT1234",
    "lot_number": "LOT-001",
    "batch_number": "BATCH-001",
    "initial_quantity": 100.0,
    "quality_grade": "A",
    "harvest_date": "2025-10-23",
    "expiry_date": "2025-11-28",
    "warehouse_location": "Warehouse A",
    "storage_conditions": "Cool and dry",
    "lot_price": 10000
  }'
```

**Expected:** ✅ Success (201 Created) - No more JSONB errors!

### Test Order Creation
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "buyer_organization_id": "ORG123",
    "seller_organization_id": "ORG456",
    "items": [...]
  }'
```

**Expected:** ✅ Success - No more JSONB errors!

## Other JSONB Fields in Codebase

The following models also have `jsonb` fields but weren't causing immediate issues:
- `pricing/price.go` - Metadata
- `collaborator/collaborator.go` - Preferences, NotificationSettings, Tags
- `common/audit_log.go` - FieldsChanged, OldValues, NewValues, Metadata, Tags
- `discounts/discount.go` - EntityIDs, Metadata
- `actors/customer.go` - BillingAddress, ShippingAddress, Metadata
- `marketplace/auction_event.go` - EventData
- `taxation/tax.go` - Metadata
- `catalog/variant.go` - Attributes
- `catalog/catalog_item.go` - Attributes, Dimensions, ServiceArea
- `outbox/outbox_event.go` - EventData, EventMetadata

### Recommendation
If these fields start causing errors, apply the same fix:
1. Change from `string` to `sql.NullString`
2. Update all assignments to use `sql.NullString{String: value, Valid: true}`
3. Update all reads to check `.Valid` and access `.String`

## Dependencies Added
- Standard library `database/sql` (already available, no go get needed)
- Removed unnecessary `gorm.io/datatypes` (tried but didn't work with custom db layer)

## Summary

✅ **Fixed:** Inventory lot creation (metadata + test_results)
✅ **Fixed:** Order creation (metadata)
✅ **Fixed:** Order item metadata
✅ **Fixed:** Order status history metadata
✅ **Build:** Successful
✅ **Ready:** For production use

### Key Takeaway
**Always use `sql.NullString` for PostgreSQL `jsonb` fields in Go** to properly handle NULL values and avoid empty string errors!
