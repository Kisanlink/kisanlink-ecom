-- Initialize KisanLink E-commerce Database
-- This script sets up the basic database structure and initial data

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Create schemas
CREATE SCHEMA IF NOT EXISTS ecommerce;
CREATE SCHEMA IF NOT EXISTS audit;

-- Set default schema
SET search_path TO ecommerce, public;

-- Create basic tables (migrations will handle the full schema)
-- This is just to ensure the database is ready for the application

-- Create a simple health check table
CREATE TABLE IF NOT EXISTS health_check (
    id SERIAL PRIMARY KEY,
    status VARCHAR(10) DEFAULT 'ok',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert initial health check record
INSERT INTO health_check (status) VALUES ('ok') ON CONFLICT DO NOTHING;

-- Create initial admin user (for development only)
-- This will be replaced by proper user management
CREATE TABLE IF NOT EXISTS temp_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert development admin user
INSERT INTO temp_users (email, name, role)
VALUES ('admin@kisanlink.local', 'Development Admin', 'admin')
ON CONFLICT (email) DO NOTHING;

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA ecommerce TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA ecommerce TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA audit TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA audit TO postgres;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_temp_users_email ON temp_users(email);
CREATE INDEX IF NOT EXISTS idx_temp_users_role ON temp_users(role);

-- Log successful initialization
INSERT INTO health_check (status) VALUES ('initialized');

COMMIT;
