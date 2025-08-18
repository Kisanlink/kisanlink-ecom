package config

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeDynamoDB   DatabaseType = "dynamodb"
)

// MultiDatabaseConfig defines database configuration for different model types
type MultiDatabaseConfig struct {
	// PostgreSQL configuration for relational data
	PostgreSQL PostgreSQLConfig `yaml:"postgresql" json:"postgresql"`

	// DynamoDB configuration for NoSQL data
	DynamoDB DynamoDBConfig `yaml:"dynamodb" json:"dynamodb"`
}

// PostgreSQLConfig defines PostgreSQL connection settings
type PostgreSQLConfig struct {
	Host            string `yaml:"host" json:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port            int    `yaml:"port" json:"port" env:"POSTGRES_PORT" env-default:"5432"`
	Database        string `yaml:"database" json:"database" env:"POSTGRES_DB" env-default:"kisanlink_ecom"`
	Username        string `yaml:"username" json:"username" env:"POSTGRES_USER" env-default:"postgres"`
	Password        string `yaml:"password" json:"password" env:"POSTGRES_PASSWORD" env-default:""`
	SSLMode         string `yaml:"ssl_mode" json:"ssl_mode" env:"POSTGRES_SSL_MODE" env-default:"disable"`
	MaxOpenConns    int    `yaml:"max_open_conns" json:"max_open_conns" env:"POSTGRES_MAX_OPEN_CONNS" env-default:"25"`
	MaxIdleConns    int    `yaml:"max_idle_conns" json:"max_idle_conns" env:"POSTGRES_MAX_IDLE_CONNS" env-default:"5"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" json:"conn_max_lifetime" env:"POSTGRES_CONN_MAX_LIFETIME" env-default:"300"`
}

// DynamoDBConfig defines DynamoDB connection settings
type DynamoDBConfig struct {
	Region          string `yaml:"region" json:"region" env:"DYNAMODB_REGION" env-default:"us-east-1"`
	Endpoint        string `yaml:"endpoint" json:"endpoint" env:"DYNAMODB_ENDPOINT" env-default:""`
	AccessKeyID     string `yaml:"access_key_id" json:"access_key_id" env:"DYNAMODB_ACCESS_KEY_ID" env-default:""`
	SecretAccessKey string `yaml:"secret_access_key" json:"secret_access_key" env:"DYNAMODB_SECRET_ACCESS_KEY" env-default:""`
	DisableSSL      bool   `yaml:"disable_ssl" json:"disable_ssl" env:"DYNAMODB_DISABLE_SSL" env-default:"false"`
	Table           string `yaml:"table" json:"table" env:"DYNAMODB_TABLE" env-default:"kisanlink_ecom"`
}

// ModelDatabaseMapping defines which database each model should use
var ModelDatabaseMapping = map[string]DatabaseType{
	// PostgreSQL Models - Relational data with complex queries and relationships

	// User/Organization references (structured, relational)
	"Customer":     DatabaseTypePostgreSQL,
	"Collaborator": DatabaseTypePostgreSQL,
	"Vendor":       DatabaseTypePostgreSQL,

	// Catalog system (structured, relational, complex queries)
	"CatalogItem":     DatabaseTypePostgreSQL,
	"Product":         DatabaseTypePostgreSQL,
	"ServiceOffering": DatabaseTypePostgreSQL,
	"LabourOffering":  DatabaseTypePostgreSQL,
	"InventoryLot":    DatabaseTypePostgreSQL,
	"ServiceSlot":     DatabaseTypePostgreSQL,
	"LabourPool":      DatabaseTypePostgreSQL,

	// Order system (transactional, ACID requirements)
	"Order":     DatabaseTypePostgreSQL,
	"OrderItem": DatabaseTypePostgreSQL,

	// Pricing system (structured, relational)
	"Price":     DatabaseTypePostgreSQL,
	"PriceTier": DatabaseTypePostgreSQL,
	"PriceRule": DatabaseTypePostgreSQL,

	// Taxation system (structured, relational)
	"TaxRate":      DatabaseTypePostgreSQL,
	"TaxRule":      DatabaseTypePostgreSQL,
	"TaxExemption": DatabaseTypePostgreSQL,

	// Discount system (structured, relational)
	"Discount":      DatabaseTypePostgreSQL,
	"DiscountRule":  DatabaseTypePostgreSQL,
	"DiscountUsage": DatabaseTypePostgreSQL,

	// DynamoDB Models - NoSQL data with simple access patterns

	// Session/Authentication data (high-frequency, simple key-value)
	"UserSession":  DatabaseTypeDynamoDB,
	"AuthToken":    DatabaseTypeDynamoDB,
	"RefreshToken": DatabaseTypeDynamoDB,

	// Shopping cart (temporary, high-frequency access)
	"ShoppingCart": DatabaseTypeDynamoDB,
	"CartItem":     DatabaseTypeDynamoDB,

	// User preferences/settings (simple key-value)
	"UserPreference": DatabaseTypeDynamoDB,
	"UserSetting":    DatabaseTypeDynamoDB,

	// Audit logs (high-volume write, infrequent complex queries)
	"AuditLog":    DatabaseTypeDynamoDB,
	"ActivityLog": DatabaseTypeDynamoDB,

	// Cache data (temporary, high-frequency)
	"CacheEntry": DatabaseTypeDynamoDB,
	"RateLimit":  DatabaseTypeDynamoDB,
}

// GetDatabaseTypeForModel returns the database type for a given model
func GetDatabaseTypeForModel(modelName string) DatabaseType {
	if dbType, exists := ModelDatabaseMapping[modelName]; exists {
		return dbType
	}
	// Default to PostgreSQL for unknown models
	return DatabaseTypePostgreSQL
}

// IsPostgreSQLModel checks if a model should use PostgreSQL
func IsPostgreSQLModel(modelName string) bool {
	return GetDatabaseTypeForModel(modelName) == DatabaseTypePostgreSQL
}

// IsDynamoDBModel checks if a model should use DynamoDB
func IsDynamoDBModel(modelName string) bool {
	return GetDatabaseTypeForModel(modelName) == DatabaseTypeDynamoDB
}
