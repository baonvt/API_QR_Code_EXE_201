-- ================================================================
-- MIGRATION: Add Payment Tables and SePay Integration
-- Date: 02/02/2026
-- Description: Thêm các bảng và cột cho tích hợp SePay webhook
-- ================================================================

-- ================================================================
-- 1. UPDATE: Thêm cột vào bảng payment_settings
-- ================================================================
-- Thêm cột bank_code
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'payment_settings' AND COLUMN_NAME = 'bank_code')
BEGIN
    ALTER TABLE payment_settings ADD bank_code NVARCHAR(20) NULL;
END;
GO

-- Thêm cột sepay_linked
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'payment_settings' AND COLUMN_NAME = 'sepay_linked')
BEGIN
    ALTER TABLE payment_settings ADD sepay_linked BIT DEFAULT 0;
END;
GO

-- Thêm cột sepay_bank_account_id
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'payment_settings' AND COLUMN_NAME = 'sepay_bank_account_id')
BEGIN
    ALTER TABLE payment_settings ADD sepay_bank_account_id NVARCHAR(100) NULL;
END;
GO

-- Thêm cột sepay_linked_at
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'payment_settings' AND COLUMN_NAME = 'sepay_linked_at')
BEGIN
    ALTER TABLE payment_settings ADD sepay_linked_at DATETIME2 NULL;
END;
GO

-- ================================================================
-- 2. UPDATE: Thêm cột vào bảng orders
-- ================================================================
-- Thêm cột payment_code
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'orders' AND COLUMN_NAME = 'payment_code')
BEGIN
    ALTER TABLE orders ADD payment_code NVARCHAR(50) NULL;
    CREATE INDEX IX_orders_payment_code ON orders(payment_code);
END;
GO

-- Thêm cột payment_expires_at
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'orders' AND COLUMN_NAME = 'payment_expires_at')
BEGIN
    ALTER TABLE orders ADD payment_expires_at DATETIME2 NULL;
END;
GO

-- ================================================================
-- 3. CREATE: Bảng package_subscriptions
-- ================================================================
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'package_subscriptions')
BEGIN
    CREATE TABLE package_subscriptions (
        id INT IDENTITY(1,1) PRIMARY KEY,
        
        -- Thông tin user (chưa tạo chính thức)
        email NVARCHAR(255) NOT NULL,
        password_hash NVARCHAR(255) NOT NULL,
        name NVARCHAR(255) NOT NULL,
        phone NVARCHAR(20) NULL,
        restaurant_name NVARCHAR(255) NOT NULL,
        
        -- Package info
        package_id INT NOT NULL,
        billing_cycle NVARCHAR(20) NOT NULL,
        amount DECIMAL(12,0) NOT NULL,
        
        -- Payment tracking
        payment_code NVARCHAR(100) NOT NULL UNIQUE,
        payment_status NVARCHAR(20) DEFAULT 'pending',
        qr_content NVARCHAR(500) NULL,
        
        -- Kết quả sau khi thanh toán
        user_id INT NULL,
        restaurant_id INT NULL,
        
        expires_at DATETIME2 NOT NULL,
        paid_at DATETIME2 NULL,
        
        created_at DATETIME2 DEFAULT GETDATE(),
        updated_at DATETIME2 DEFAULT GETDATE(),
        
        CONSTRAINT FK_package_subscriptions_package FOREIGN KEY (package_id) REFERENCES packages(id)
    );
    
    CREATE INDEX IX_package_subscriptions_code ON package_subscriptions(payment_code);
    CREATE INDEX IX_package_subscriptions_email ON package_subscriptions(email);
    CREATE INDEX IX_package_subscriptions_status ON package_subscriptions(payment_status);
END;
GO

-- ================================================================
-- 4. CREATE: Bảng payment_transactions
-- ================================================================
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'payment_transactions')
BEGIN
    CREATE TABLE payment_transactions (
        id INT IDENTITY(1,1) PRIMARY KEY,
        
        -- Loại giao dịch
        transaction_type NVARCHAR(20) NOT NULL,
        
        -- Reference
        reference_id INT NOT NULL,
        reference_code NVARCHAR(100) NOT NULL,
        
        -- Thông tin từ SePay webhook
        sepay_transaction_id BIGINT NULL,
        gateway NVARCHAR(50) NULL,
        transaction_date DATETIME2 NULL,
        account_number NVARCHAR(50) NULL,
        sub_account NVARCHAR(50) NULL,
        transfer_type NVARCHAR(10) NULL,
        transfer_amount DECIMAL(12,0) NOT NULL,
        accumulated DECIMAL(12,0) NULL,
        code NVARCHAR(500) NULL,
        transaction_content NVARCHAR(500) NULL,
        reference_number NVARCHAR(100) NULL,
        description NVARCHAR(1000) NULL,
        
        -- Trạng thái xử lý
        status NVARCHAR(20) DEFAULT 'pending',
        verified_at DATETIME2 NULL,
        error_message NVARCHAR(500) NULL,
        
        -- Raw data
        raw_webhook_data NVARCHAR(MAX) NULL,
        
        created_at DATETIME2 DEFAULT GETDATE(),
        updated_at DATETIME2 DEFAULT GETDATE()
    );
    
    CREATE INDEX IX_payment_transactions_reference ON payment_transactions(reference_code);
    CREATE INDEX IX_payment_transactions_sepay_id ON payment_transactions(sepay_transaction_id);
    CREATE INDEX IX_payment_transactions_type ON payment_transactions(transaction_type, status);
END;
GO

PRINT '✅ Migration completed: Payment tables and SePay integration added!';
GO
