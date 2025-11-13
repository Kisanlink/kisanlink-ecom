-- Initialize Test Database for KisanLink E-commerce
-- This script sets up the test database with minimal data

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Create schemas
CREATE SCHEMA IF NOT EXISTS ecommerce;
CREATE SCHEMA IF NOT EXISTS audit;

-- Set default schema
SET search_path TO ecommerce, public;

-- Create health check table
CREATE TABLE IF NOT EXISTS health_check (
    id SERIAL PRIMARY KEY,
    status VARCHAR(10) DEFAULT 'ok',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert test health check record
INSERT INTO health_check (status) VALUES ('test_ready');

-- Create test users table
CREATE TABLE IF NOT EXISTS temp_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert test users
INSERT INTO temp_users (email, name, role) VALUES
('test@kisanlink.local', 'Test User', 'user'),
('admin@kisanlink.local', 'Test Admin', 'admin'),
('collaborator@kisanlink.local', 'Test Collaborator', 'collaborator')
ON CONFLICT (email) DO NOTHING;

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA ecommerce TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA ecommerce TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA audit TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA audit TO postgres;

COMMIT;
