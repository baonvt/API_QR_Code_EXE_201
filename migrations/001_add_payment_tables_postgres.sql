-- ================================================================
-- MIGRATION: Add Payment Tables and SePay Integration
-- Date: 02/02/2026
-- Database: PostgreSQL
-- Description: Thêm các bảng và cột cho tích hợp SePay webhook
-- ================================================================

-- ================================================================
-- 1. UPDATE: Thêm cột vào bảng payment_settings
-- ================================================================

-- Thêm cột bank_code
ALTER TABLE payment_settings ADD COLUMN IF NOT EXISTS bank_code VARCHAR(20) NULL;

-- Thêm cột sepay_linked
ALTER TABLE payment_settings ADD COLUMN IF NOT EXISTS sepay_linked BOOLEAN DEFAULT FALSE;

-- Thêm cột sepay_bank_account_id
ALTER TABLE payment_settings ADD COLUMN IF NOT EXISTS sepay_bank_account_id VARCHAR(100) NULL;

-- Thêm cột sepay_linked_at
ALTER TABLE payment_settings ADD COLUMN IF NOT EXISTS sepay_linked_at TIMESTAMP NULL;

-- ================================================================
-- 2. UPDATE: Thêm cột vào bảng orders
-- ================================================================

-- Thêm cột payment_code
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_code VARCHAR(50) NULL;

-- Thêm index cho payment_code
CREATE INDEX IF NOT EXISTS idx_orders_payment_code ON orders(payment_code);

-- Thêm cột payment_expires_at
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_expires_at TIMESTAMP NULL;

-- ================================================================
-- 3. CREATE: Bảng package_subscriptions
-- ================================================================

CREATE TABLE IF NOT EXISTS package_subscriptions (
    id SERIAL PRIMARY KEY,
    
    -- Thông tin user (chưa tạo chính thức)
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NULL,
    restaurant_name VARCHAR(255) NOT NULL,
    
    -- Package info
    package_id INTEGER NOT NULL,
    billing_cycle VARCHAR(20) NOT NULL,
    amount DECIMAL(12,0) NOT NULL,
    
    -- Payment tracking
    payment_code VARCHAR(100) NOT NULL UNIQUE,
    payment_status VARCHAR(20) DEFAULT 'pending',
    qr_content VARCHAR(500) NULL,
    
    -- Kết quả sau khi thanh toán
    user_id INTEGER NULL,
    restaurant_id INTEGER NULL,
    
    expires_at TIMESTAMP NOT NULL,
    paid_at TIMESTAMP NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_package_subscriptions_package FOREIGN KEY (package_id) REFERENCES packages(id)
);

-- Tạo indexes
CREATE INDEX IF NOT EXISTS idx_package_subscriptions_code ON package_subscriptions(payment_code);
CREATE INDEX IF NOT EXISTS idx_package_subscriptions_email ON package_subscriptions(email);
CREATE INDEX IF NOT EXISTS idx_package_subscriptions_status ON package_subscriptions(payment_status);

-- ================================================================
-- 4. CREATE: Bảng payment_transactions
-- ================================================================

CREATE TABLE IF NOT EXISTS payment_transactions (
    id SERIAL PRIMARY KEY,
    
    -- Loại giao dịch
    transaction_type VARCHAR(20) NOT NULL,
    
    -- Reference
    reference_id INTEGER NOT NULL,
    reference_code VARCHAR(100) NOT NULL,
    
    -- Thông tin từ SePay webhook
    sepay_transaction_id BIGINT NULL,
    gateway VARCHAR(50) NULL,
    transaction_date TIMESTAMP NULL,
    account_number VARCHAR(50) NULL,
    sub_account VARCHAR(50) NULL,
    transfer_type VARCHAR(10) NULL,
    transfer_amount DECIMAL(12,0) NOT NULL,
    accumulated DECIMAL(12,0) NULL,
    code VARCHAR(500) NULL,
    transaction_content VARCHAR(500) NULL,
    reference_number VARCHAR(100) NULL,
    description VARCHAR(1000) NULL,
    
    -- Trạng thái xử lý
    status VARCHAR(20) DEFAULT 'pending',
    verified_at TIMESTAMP NULL,
    error_message VARCHAR(500) NULL,
    
    -- Raw data
    raw_webhook_data TEXT NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tạo indexes
CREATE INDEX IF NOT EXISTS idx_payment_transactions_reference ON payment_transactions(reference_code);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_sepay_id ON payment_transactions(sepay_transaction_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_type ON payment_transactions(transaction_type, status);

-- ================================================================
-- Thông báo hoàn thành
-- ================================================================
DO $$
BEGIN
    RAISE NOTICE '✅ Migration completed: Payment tables and SePay integration added!';
END $$;
