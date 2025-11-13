package catalog

import (
	"testing"
	"time"

	catalogModels "kisanlink-ecom/entities/models/catalog"
	catalogService "kisanlink-ecom/internal/services/catalog"

	"github.com/stretchr/testify/assert"
)

func TestETagService(t *testing.T) {
	etagService := catalogService.NewETagService(nil)

	// Create test catalog item
	testItem := &catalogModels.CatalogItem{
		ItemType: catalogModels.CatalogItemTypeProduct,
		Name:     "Test Product",
		Version:  1,
	}
	testItem.ID = "test-id-123"
	testItem.UpdatedAt = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("ETag generation should be consistent", func(t *testing.T) {
		etag1 := etagService.GenerateETag(testItem)
		etag2 := etagService.GenerateETag(testItem)

		assert.Equal(t, etag1, etag2)
		assert.NotEmpty(t, etag1)
		assert.True(t, len(etag1) > 2) // Should be quoted
		assert.Contains(t, etag1, `"`) // Should contain quotes
	})

	t.Run("ETag should change when version changes", func(t *testing.T) {
		etag1 := etagService.GenerateETag(testItem)

		testItem.Version = 2
		etag2 := etagService.GenerateETag(testItem)

		assert.NotEqual(t, etag1, etag2)
		assert.NotEmpty(t, etag1)
		assert.NotEmpty(t, etag2)
	})

	t.Run("ETag should change when updated time changes", func(t *testing.T) {
		// Reset version for consistent test
		testItem.Version = 1
		etag1 := etagService.GenerateETag(testItem)

		testItem.UpdatedAt = time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
		etag2 := etagService.GenerateETag(testItem)

		assert.NotEqual(t, etag1, etag2)
		assert.NotEmpty(t, etag1)
		assert.NotEmpty(t, etag2)
	})

	t.Run("ETag should be empty for nil item", func(t *testing.T) {
		etag := etagService.GenerateETag(nil)
		assert.Empty(t, etag)
	})

	t.Run("ETag comparison should work correctly", func(t *testing.T) {
		etag1 := etagService.GenerateETag(testItem)
		etag2 := etagService.GenerateETag(testItem)

		// Same ETags should be equal
		assert.True(t, etagService.CompareETags(etag1, etag2))

		// Different ETags should not be equal
		testItem.Version = 999
		etag3 := etagService.GenerateETag(testItem)
		assert.False(t, etagService.CompareETags(etag1, etag3))

		// Empty ETags should not be equal
		assert.False(t, etagService.CompareETags("", etag1))
		assert.False(t, etagService.CompareETags(etag1, ""))
		assert.False(t, etagService.CompareETags("", ""))
	})

	t.Run("ETag validation should work", func(t *testing.T) {
		// Valid quoted ETag
		err := etagService.ValidateETagFormat(`"valid-etag"`)
		assert.NoError(t, err)

		// Empty ETag should be valid (no caching)
		err = etagService.ValidateETagFormat("")
		assert.NoError(t, err)

		// Unquoted ETag should be invalid
		err = etagService.ValidateETagFormat("unquoted-etag")
		assert.Error(t, err)

		// Empty quoted ETag should be invalid
		err = etagService.ValidateETagFormat(`""`)
		assert.Error(t, err)
	})

	t.Run("parseIfNoneMatch should handle various formats", func(t *testing.T) {
		// Test wildcard
		etags := etagService.ParseIfNoneMatch("*")
		assert.Equal(t, []string{"*"}, etags)

		// Test single ETag
		etags = etagService.ParseIfNoneMatch(`"etag1"`)
		assert.Equal(t, []string{`"etag1"`}, etags)

		// Test multiple ETags
		etags = etagService.ParseIfNoneMatch(`"etag1", "etag2", "etag3"`)
		expected := []string{`"etag1"`, `"etag2"`, `"etag3"`}
		assert.Equal(t, expected, etags)

		// Test with spaces and unquoted
		etags = etagService.ParseIfNoneMatch(`etag1, etag2`)
		expected = []string{`"etag1"`, `"etag2"`}
		assert.Equal(t, expected, etags)
	})
}
