-- KisanLink E-Commerce Database Initialization Script
-- This script sets up the initial database structure and sample data

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create additional schemas if needed
-- CREATE SCHEMA IF NOT EXISTS analytics;
-- CREATE SCHEMA IF NOT EXISTS audit;

-- Create sample data for development (optional)
-- You can uncomment these if you want some initial test data

/*
-- Sample Users
INSERT INTO users (id, username, email, full_name, password_hash, role, status, created_at, updated_at, created_by, updated_by) VALUES
('usr_1704067200_12345678', 'admin', 'admin@kisanlink.local', 'System Administrator', '$2a$10$example_hash_here', 'admin', 'active', NOW(), NOW(), 'system', 'system'),
('usr_1704067260_23456789', 'customer1', 'customer1@example.com', 'John Doe', '$2a$10$example_hash_here', 'customer', 'active', NOW(), NOW(), 'system', 'system'),
('usr_1704067320_34567890', 'vendor1', 'vendor1@example.com', 'Jane Smith', '$2a$10$example_hash_here', 'vendor', 'active', NOW(), NOW(), 'system', 'system');

-- Sample Products
INSERT INTO products (id, name, description, price, currency, category, stock, status, created_at, updated_at, created_by, updated_by) VALUES
('prd_1704067400_12345678', 'Organic Tomatoes', 'Fresh organic tomatoes from local farms', 4.99, 'USD', 'Vegetables', 100, 'active', NOW(), NOW(), 'system', 'system'),
('prd_1704067460_23456789', 'Fresh Milk', 'Farm fresh whole milk, 1 gallon', 3.49, 'USD', 'Dairy', 50, 'active', NOW(), NOW(), 'system', 'system'),
('prd_1704067520_34567890', 'Brown Rice', 'Organic brown rice, 5 lb bag', 12.99, 'USD', 'Grains', 25, 'active', NOW(), NOW(), 'system', 'system');

-- Sample Orders
INSERT INTO orders (id, user_id, items, total_price, currency, status, created_at, updated_at, created_by, updated_by) VALUES
('ord_1704067600_12345678', 'usr_1704067260_23456789', '[{"product_id":"prd_1704067400_12345678","quantity":2,"price":4.99}]', 9.98, 'USD', 'pending', NOW(), NOW(), 'system', 'system'),
('ord_1704067660_23456789', 'usr_1704067260_23456789', '[{"product_id":"prd_1704067460_23456789","quantity":1,"price":3.49},{"product_id":"prd_1704067520_34567890","quantity":1,"price":12.99}]', 16.48, 'USD', 'confirmed', NOW(), NOW(), 'system', 'system');
*/

-- Create indexes for better performance
-- These will be created automatically by GORM, but you can add custom ones here

-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email_active ON users(email) WHERE status = 'active' AND deleted_at IS NULL;
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_category_active ON products(category) WHERE status = 'active' AND deleted_at IS NULL;
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_user_date ON orders(user_id, created_at) WHERE deleted_at IS NULL;

-- Grant permissions (if using specific database users)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO kisanlink_app;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO kisanlink_app;

-- Log initialization completion
DO $$
BEGIN
    RAISE NOTICE 'KisanLink E-Commerce database initialized successfully!';
    RAISE NOTICE 'Database: kisanlink_ecom';
    RAISE NOTICE 'Extensions: uuid-ossp';
    RAISE NOTICE 'Ready for GORM auto-migration';
END $$; 