//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	"kisanlink-ecom/internal/database"
	catalogRepo "kisanlink-ecom/internal/repositories/catalog"
	"kisanlink-ecom/internal/repositories/inventory"
	inventoryService "kisanlink-ecom/internal/services/inventory"
	"kisanlink-ecom/tests/testutils"

	"github.com/Kisanlink/kisanlink-db/pkg/db"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInventoryIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database from environment variables
	cfg := testutils.LoadTestDatabaseConfig()

	dbManager, err := database.NewDatabaseManager(cfg)
	if err != nil {
		t.Skipf("Database not available for integration test: %v", err)
	}
	defer dbManager.Close()

	// Initialize repositories and services
	catalogRepository := catalogRepo.NewCatalogRepository(dbManager.GetManager(db.BackendGorm))
	inventoryRepo := inventory.NewInventoryRepository(dbManager.GetManager(db.BackendGorm))
	inventorySvc := inventoryService.NewInventoryService(inventoryRepo, catalogRepository)

	ctx := context.Background()
	userID := "test-user-123"
	orgID := "test-org-456"

	t.Run("CompleteInventoryWorkflow", func(t *testing.T) {
		// Create a test product
		product := &catalogModels.CatalogItem{
			OrganizationID: orgID,
			ItemType:       catalogModels.CatalogItemTypeProduct,
			Name:           "Test Product",
			Description:    "A test product for inventory testing",
			BasePrice:      decimal.NewFromFloat(25.50),
			IsActive:       true,
			Visibility:     catalogModels.VisibilityPrivate,
		}

		err := catalogRepository.Create(ctx, product)
		require.NoError(t, err)

		// Create inventory lot
		createReq := &inventoryService.CreateInventoryLotRequest{
			CatalogItemID:     product.ID,
			LotNumber:         "LOT-001",
			BatchNumber:       "BATCH-001",
			InitialQuantity:   decimal.NewFromInt(100),
			QualityGrade:      "A",
			HarvestDate:       &time.Time{},
			ExpiryDate:        func() *time.Time { t := time.Now().AddDate(0, 6, 0); return &t }(),
			WarehouseLocation: "Warehouse A",
			StorageConditions: "Cool, dry place",
			LotPrice:          func() *decimal.Decimal { p := decimal.NewFromFloat(24.00); return &p }(),
			Metadata:          `{"source": "test"}`,
		}

		lot, err := inventorySvc.CreateInventoryLot(ctx, createReq, userID, orgID)
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(100), lot.AvailableQty)
		assert.Equal(t, decimal.Zero, lot.ReservedQty)
		assert.Equal(t, catalogModels.LotStatusActive, lot.Status)

		// Check available quantity
		available, err := inventorySvc.GetAvailableQuantity(ctx, product.ID, userID, orgID)
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(100), available)

		// Reserve some inventory
		reservedLots, err := inventorySvc.ReserveInventory(ctx, product.ID, decimal.NewFromInt(30), userID, orgID)
		require.NoError(t, err)
		assert.Len(t, reservedLots, 1)
		assert.Equal(t, decimal.NewFromInt(70), reservedLots[0].AvailableQty)
		assert.Equal(t, decimal.NewFromInt(30), reservedLots[0].ReservedQty)

		// Check available quantity after reservation
		available, err = inventorySvc.GetAvailableQuantity(ctx, product.ID, userID, orgID)
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(70), available)

		// Sell some reserved inventory
		err = inventorySvc.SellInventory(ctx, product.ID, decimal.NewFromInt(20), userID, orgID)
		require.NoError(t, err)

		// Check quantities after sale
		updatedLot, err := inventorySvc.GetInventoryLot(ctx, lot.ID, userID, orgID)
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(70), updatedLot.AvailableQty)
		assert.Equal(t, decimal.NewFromInt(10), updatedLot.ReservedQty)

		// Release remaining reserved inventory
		err = inventorySvc.ReleaseInventory(ctx, product.ID, decimal.NewFromInt(10), userID, orgID)
		require.NoError(t, err)

		// Check final quantities
		finalLot, err := inventorySvc.GetInventoryLot(ctx, lot.ID, userID, orgID)
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(80), finalLot.AvailableQty)
		assert.Equal(t, decimal.Zero, finalLot.ReservedQty)
		assert.Equal(t, catalogModels.LotStatusActive, finalLot.Status)

		// Test inventory adjustment
		adjustedLot, err := inventorySvc.AdjustInventory(ctx, lot.ID, decimal.NewFromInt(20), "Stock adjustment", userID, orgID)
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(100), adjustedLot.AvailableQty)
		assert.Equal(t, decimal.NewFromInt(120), adjustedLot.Quantity)

		// Test audit trail
		auditResponse, err := inventorySvc.GetAuditTrail(ctx, lot.ID, 0, 10, userID, orgID)
		require.NoError(t, err)
		assert.Greater(t, len(auditResponse.Logs), 0)
		assert.Greater(t, auditResponse.Total, int64(0))
	})

	t.Run("InventoryValidationErrors", func(t *testing.T) {
		// Test creating inventory for non-existent product
		createReq := &inventoryService.CreateInventoryLotRequest{
			CatalogItemID:   "non-existent-id",
			LotNumber:       "LOT-002",
			InitialQuantity: decimal.NewFromInt(50),
		}

		_, err := inventorySvc.CreateInventoryLot(ctx, createReq, userID, orgID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "catalog item not found")

		// Test reserving more than available
		product := &catalogModels.CatalogItem{
			OrganizationID: orgID,
			ItemType:       catalogModels.CatalogItemTypeProduct,
			Name:           "Limited Product",
			BasePrice:      decimal.NewFromFloat(10.00),
			IsActive:       true,
			Visibility:     catalogModels.VisibilityPrivate,
		}

		err = catalogRepository.Create(ctx, product)
		require.NoError(t, err)

		createReq = &inventoryService.CreateInventoryLotRequest{
			CatalogItemID:   product.ID,
			LotNumber:       "LOT-LIMITED",
			InitialQuantity: decimal.NewFromInt(10),
		}

		_, err = inventorySvc.CreateInventoryLot(ctx, createReq, userID, orgID)
		require.NoError(t, err)

		// Try to reserve more than available
		_, err = inventorySvc.ReserveInventory(ctx, product.ID, decimal.NewFromInt(15), userID, orgID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient inventory")

		// Try to release more than reserved
		_, err = inventorySvc.ReserveInventory(ctx, product.ID, decimal.NewFromInt(5), userID, orgID)
		require.NoError(t, err)

		err = inventorySvc.ReleaseInventory(ctx, product.ID, decimal.NewFromInt(10), userID, orgID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient reserved inventory")

		// Try to sell more than reserved
		err = inventorySvc.SellInventory(ctx, product.ID, decimal.NewFromInt(10), userID, orgID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient reserved inventory")
	})

	t.Run("InventoryForNonProducts", func(t *testing.T) {
		// Create a service (not a product)
		service := &catalogModels.CatalogItem{
			OrganizationID: orgID,
			ItemType:       catalogModels.CatalogItemTypeService,
			Name:           "Test Service",
			BasePrice:      decimal.NewFromFloat(100.00),
			IsActive:       true,
			Visibility:     catalogModels.VisibilityPrivate,
		}

		err := catalogRepository.Create(ctx, service)
		require.NoError(t, err)

		// Try to create inventory for service
		createReq := &inventoryService.CreateInventoryLotRequest{
			CatalogItemID:   service.ID,
			LotNumber:       "LOT-SERVICE",
			InitialQuantity: decimal.NewFromInt(10),
		}

		_, err = inventorySvc.CreateInventoryLot(ctx, createReq, userID, orgID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "inventory lots can only be created for products")

		// Try to get available quantity for service
		_, err = inventorySvc.GetAvailableQuantity(ctx, service.ID, userID, orgID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "inventory is only available for products")
	})

	t.Run("ExpirationHandling", func(t *testing.T) {
		// Create a product with expiring inventory
		product := &catalogModels.CatalogItem{
			OrganizationID: orgID,
			ItemType:       catalogModels.CatalogItemTypeProduct,
			Name:           "Perishable Product",
			BasePrice:      decimal.NewFromFloat(15.00),
			IsActive:       true,
			Visibility:     catalogModels.VisibilityPrivate,
		}

		err := catalogRepository.Create(ctx, product)
		require.NoError(t, err)

		// Create inventory lot that expires yesterday
		yesterday := time.Now().AddDate(0, 0, -1)
		createReq := &inventoryService.CreateInventoryLotRequest{
			CatalogItemID:   product.ID,
			LotNumber:       "LOT-EXPIRED",
			InitialQuantity: decimal.NewFromInt(50),
			ExpiryDate:      &yesterday,
		}

		_, err = inventorySvc.CreateInventoryLot(ctx, createReq, userID, orgID)
		require.NoError(t, err)

		// Process expiring lots
		expiredLots, err := inventorySvc.ProcessExpiringLots(ctx, orgID)
		require.NoError(t, err)
		assert.Greater(t, len(expiredLots), 0)
		// Find our expired lot
		var foundExpiredLot *catalogModels.InventoryLot
		for _, expiredLot := range expiredLots {
			if expiredLot.LotNumber == "LOT-EXPIRED" {
				foundExpiredLot = expiredLot
				break
			}
		}
		require.NotNil(t, foundExpiredLot)
		assert.Equal(t, catalogModels.LotStatusExpired, foundExpiredLot.Status)

		// Manually mark a lot as expired
		createReq = &inventoryService.CreateInventoryLotRequest{
			CatalogItemID:   product.ID,
			LotNumber:       "LOT-MANUAL-EXPIRE",
			InitialQuantity: decimal.NewFromInt(25),
		}

		lot2, err := inventorySvc.CreateInventoryLot(ctx, createReq, userID, orgID)
		require.NoError(t, err)

		err = inventorySvc.MarkLotExpired(ctx, lot2.ID, userID, orgID)
		require.NoError(t, err)

		// Verify lot is marked as expired
		expiredLot, err := inventorySvc.GetInventoryLot(ctx, lot2.ID, userID, orgID)
		require.NoError(t, err)
		assert.Equal(t, catalogModels.LotStatusExpired, expiredLot.Status)
	})

	t.Run("OrganizationIsolation", func(t *testing.T) {
		// Create product in different organization
		otherOrgID := "other-org-789"
		product := &catalogModels.CatalogItem{
			OrganizationID: otherOrgID,
			ItemType:       catalogModels.CatalogItemTypeProduct,
			Name:           "Other Org Product",
			BasePrice:      decimal.NewFromFloat(20.00),
			IsActive:       true,
			Visibility:     catalogModels.VisibilityPrivate,
		}

		err := catalogRepository.Create(ctx, product)
		require.NoError(t, err)

		// Try to create inventory lot for product in different org
		createReq := &inventoryService.CreateInventoryLotRequest{
			CatalogItemID:   product.ID,
			LotNumber:       "LOT-OTHER-ORG",
			InitialQuantity: decimal.NewFromInt(30),
		}

		_, err = inventorySvc.CreateInventoryLot(ctx, createReq, userID, orgID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "catalog item does not belong to organization")

		// Create inventory in correct org
		_, err = inventorySvc.CreateInventoryLot(ctx, createReq, userID, otherOrgID)
		require.NoError(t, err)

		// Try to access inventory from wrong org
		_, err = inventorySvc.GetAvailableQuantity(ctx, product.ID, userID, orgID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "catalog item does not belong to organization")
	})
}
