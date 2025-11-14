package models_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Kisanlink/kisanlink-ecom/entities/models"
)

// TestAllModelsRegistered validates that all models with base.BaseModel are registered in the registry
func TestAllModelsRegistered(t *testing.T) {
	// Get all registered models
	registeredModels := models.AllModels()

	// Create a map of registered model types for quick lookup
	registeredTypes := make(map[string]bool)
	for _, model := range registeredModels {
		modelType := reflect.TypeOf(model).Elem()
		fullName := fmt.Sprintf("%s.%s", modelType.PkgPath(), modelType.Name())
		registeredTypes[fullName] = true
		t.Logf("Registered: %s", fullName)
	}

	// Discover all models with base.BaseModel in the codebase
	modelsDir := "../../entities/models"
	discoveredModels := discoverModelsWithBaseModel(t, modelsDir)

	// Check if all discovered models are registered
	var missingModels []string
	for modelName := range discoveredModels {
		if !registeredTypes[modelName] {
			missingModels = append(missingModels, modelName)
		}
	}

	if len(missingModels) > 0 {
		t.Errorf("Found %d models with base.BaseModel that are NOT registered:", len(missingModels))
		for _, modelName := range missingModels {
			t.Errorf("  - %s", modelName)
		}
		t.Fatal("All models with base.BaseModel must be registered in entities/models/registry.go")
	}

	t.Logf("SUCCESS: All %d discovered models are properly registered", len(discoveredModels))
}

// TestNoDuplicateTableNames validates that no two models share the same table name
func TestNoDuplicateTableNames(t *testing.T) {
	registeredModels := models.AllModels()

	// Map of table names to model types
	tableNames := make(map[string][]string)

	for _, model := range registeredModels {
		modelType := reflect.TypeOf(model).Elem()
		fullName := fmt.Sprintf("%s.%s", modelType.PkgPath(), modelType.Name())

		// Try to get TableName method
		tableName := getTableName(model)
		if tableName != "" {
			tableNames[tableName] = append(tableNames[tableName], fullName)
		}
	}

	// Check for duplicates
	// Exception: catalog_items is intentionally shared by CatalogItem and its subtypes (STI pattern)
	var duplicates []string
	for tableName, models := range tableNames {
		if len(models) > 1 {
			// Allow catalog_items to be shared by CatalogItem subtypes (Single Table Inheritance)
			if tableName == "catalog_items" {
				expectedSTIModels := map[string]bool{
					"github.com/Kisanlink/kisanlink-ecom/entities/models/catalog.CatalogItem": true,
					"github.com/Kisanlink/kisanlink-ecom/entities/models/catalog.Product":     true,
					"github.com/Kisanlink/kisanlink-ecom/entities/models/catalog.Service":     true,
					"github.com/Kisanlink/kisanlink-ecom/entities/models/catalog.Labour":      true,
					"github.com/Kisanlink/kisanlink-ecom/entities/models/catalog.Contract":    true,
				}

				allValid := true
				for _, model := range models {
					if !expectedSTIModels[model] {
						allValid = false
						break
					}
				}

				if allValid {
					t.Logf("✓ Table '%s' correctly shared by STI models: %v", tableName, models)
					continue
				}
			}

			duplicates = append(duplicates, fmt.Sprintf("Table '%s' is used by: %v", tableName, models))
		}
	}

	if len(duplicates) > 0 {
		t.Errorf("Found %d table name conflicts:", len(duplicates))
		for _, dup := range duplicates {
			t.Errorf("  - %s", dup)
		}
		t.Fatal("Each model must have a unique table name")
	}

	t.Logf("SUCCESS: All %d models have unique table names", len(registeredModels))
}

// TestRegistryConsistency validates that helper functions return subsets of AllModels
func TestRegistryConsistency(t *testing.T) {
	allModels := models.AllModels()
	allModelTypes := make(map[string]bool)
	for _, model := range allModels {
		modelType := reflect.TypeOf(model).Elem()
		fullName := fmt.Sprintf("%s.%s", modelType.PkgPath(), modelType.Name())
		allModelTypes[fullName] = true
	}

	// Test each helper function
	helperFunctions := map[string]func() []interface{}{
		"CatalogModels":     models.CatalogModels,
		"MediaModels":       models.MediaModels,
		"PricingModels":     models.PricingModels,
		"ActorModels":       models.ActorModels,
		"OrderModels":       models.OrderModels,
		"TaxationModels":    models.TaxationModels,
		"DiscountModels":    models.DiscountModels,
		"MarketplaceModels": models.MarketplaceModels,
		"OutboxModels":      models.OutboxModels,
		"AuditModels":       models.AuditModels,
		"UserModels":        models.UserModels,
		"RoleModels":        models.RoleModels,
	}

	for funcName, helperFunc := range helperFunctions {
		helperModels := helperFunc()
		for _, model := range helperModels {
			modelType := reflect.TypeOf(model).Elem()
			fullName := fmt.Sprintf("%s.%s", modelType.PkgPath(), modelType.Name())

			if !allModelTypes[fullName] {
				t.Errorf("%s returns model %s which is NOT in AllModels()", funcName, fullName)
			}
		}
		t.Logf("%s: validated %d models", funcName, len(helperModels))
	}

	// Test ModelsByPriority
	priorityModels := models.ModelsByPriority()
	if len(priorityModels) != len(allModels) {
		t.Errorf("ModelsByPriority returns %d models, but AllModels returns %d", len(priorityModels), len(allModels))
	}

	for _, model := range priorityModels {
		modelType := reflect.TypeOf(model).Elem()
		fullName := fmt.Sprintf("%s.%s", modelType.PkgPath(), modelType.Name())

		if !allModelTypes[fullName] {
			t.Errorf("ModelsByPriority returns model %s which is NOT in AllModels()", fullName)
		}
	}

	t.Logf("SUCCESS: All helper functions are consistent with AllModels()")
}

// discoverModelsWithBaseModel parses Go files to find structs with base.BaseModel
func discoverModelsWithBaseModel(t *testing.T, dir string) map[string]bool {
	models := make(map[string]bool)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip test files and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Parse the Go file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Logf("Warning: failed to parse %s: %v", path, err)
			return nil
		}

		// Inspect the AST for struct declarations
		ast.Inspect(node, func(n ast.Node) bool {
			// Look for type declarations
			typeSpec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}

			// Check if it's a struct
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return true
			}

			// Check if the struct has base.BaseModel as a field
			hasBaseModel := false
			for _, field := range structType.Fields.List {
				// Check for embedded BaseModel
				if ident, ok := field.Type.(*ast.SelectorExpr); ok {
					if x, ok := ident.X.(*ast.Ident); ok {
						if x.Name == "base" && ident.Sel.Name == "BaseModel" {
							hasBaseModel = true
							break
						}
					}
				}
			}

			if hasBaseModel {
				// Construct the full model name
				pkgName := node.Name.Name
				fullName := fmt.Sprintf("github.com/Kisanlink/kisanlink-ecom/entities/models/%s.%s", pkgName, typeSpec.Name.Name)
				models[fullName] = true
				t.Logf("Discovered model with BaseModel: %s", fullName)
			}

			return true
		})

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk models directory: %v", err)
	}

	return models
}

// getTableName attempts to call the TableName method on a model
func getTableName(model interface{}) string {
	// Use reflection to call TableName() method if it exists
	value := reflect.ValueOf(model)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Create a new instance of the struct type
	instance := reflect.New(value.Type())

	// Try to call TableName method
	method := instance.MethodByName("TableName")
	if method.IsValid() {
		results := method.Call([]reflect.Value{})
		if len(results) > 0 {
			if tableName, ok := results[0].Interface().(string); ok {
				return tableName
			}
		}
	}

	return ""
}

// TestBaseModelIntegration validates that all models properly integrate with base.BaseModel
func TestBaseModelIntegration(t *testing.T) {
	registeredModels := models.AllModels()

	for _, model := range registeredModels {
		modelType := reflect.TypeOf(model).Elem()
		fullName := fmt.Sprintf("%s.%s", modelType.PkgPath(), modelType.Name())

		// Check if model has base.BaseModel field (embedded or named)
		hasBaseModel := false
		for i := 0; i < modelType.NumField(); i++ {
			field := modelType.Field(i)
			fieldType := field.Type

			// Check for embedded BaseModel
			if field.Anonymous && fieldType.String() == "base.BaseModel" {
				hasBaseModel = true
				break
			}

			// Check for named BaseModel field
			if fieldType.String() == "base.BaseModel" {
				hasBaseModel = true
				break
			}
		}

		// Special cases:
		// 1. Models with custom primary keys (no BaseModel)
		// 2. Catalog subtypes (Product, Service, Labour, Contract) inherit through CatalogItem
		if strings.Contains(fullName, "/user.User") ||
			strings.Contains(fullName, "/roles.UserRole") ||
			strings.Contains(fullName, "/roles.OrganizationRole") ||
			strings.Contains(fullName, "/roles.EcommerceRole") {
			t.Logf("✓ %s uses custom primary key (no BaseModel required)", fullName)
			continue
		}

		if strings.Contains(fullName, "/catalog.Product") ||
			strings.Contains(fullName, "/catalog.Service") ||
			strings.Contains(fullName, "/catalog.Labour") ||
			strings.Contains(fullName, "/catalog.Contract") {
			t.Logf("✓ %s inherits BaseModel through CatalogItem", fullName)
			continue
		}

		if !hasBaseModel {
			t.Errorf("✗ %s is registered but does NOT have base.BaseModel", fullName)
		} else {
			t.Logf("✓ %s has base.BaseModel", fullName)
		}
	}
}
