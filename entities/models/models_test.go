package models

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestModelsCompilation(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Test AutoMigrate with all models
	err = db.AutoMigrate(AllModels()...)
	if err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	t.Log("All models migrated successfully")
}

func TestModelsByPriority(t *testing.T) {
	models := ModelsByPriority()
	if len(models) == 0 {
		t.Fatal("No models returned by ModelsByPriority")
	}

	// Create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Test AutoMigrate with models in priority order
	err = db.AutoMigrate(models...)
	if err != nil {
		t.Fatalf("Failed to migrate models in priority order: %v", err)
	}

	t.Logf("Successfully migrated %d models in priority order", len(models))
}

func TestCatalogModels(t *testing.T) {
	models := CatalogModels()
	if len(models) == 0 {
		t.Fatal("No catalog models returned")
	}

	// Create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Test AutoMigrate with catalog models
	err = db.AutoMigrate(models...)
	if err != nil {
		t.Fatalf("Failed to migrate catalog models: %v", err)
	}

	t.Logf("Successfully migrated %d catalog models", len(models))
}
