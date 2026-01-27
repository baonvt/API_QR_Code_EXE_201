-- ================================================================
-- HỆ THỐNG QUẢN LÝ NHÀ HÀNG - DATABASE SCHEMA V2
-- SQL Server
-- Ngày cập nhật: 27/01/2026
-- ================================================================

-- ================================================================
-- TỔNG QUAN HỆ THỐNG
-- ================================================================
/*
╔══════════════════════════════════════════════════════════════════╗
║                    3 VAI TRÒ CHÍNH                               ║
╠══════════════════════════════════════════════════════════════════╣
║ 1. ADMIN      │ Quản trị viên hệ thống (Bạn)                    ║
║               │ - Quản lý gói dịch vụ                            ║
║               │ - Quản lý nhà hàng đăng ký                       ║
║               │ - Xem thống kê toàn hệ thống                     ║
╠═══════════════╪══════════════════════════════════════════════════╣
║ 2. RESTAURANT │ Nhà hàng (Chủ nhà hàng)                         ║
║               │ - Đăng ký tài khoản + chọn gói                   ║
║               │ - Quản lý menu, bàn, đơn hàng                    ║
║               │ - Xem báo cáo doanh thu                          ║
║               │ - In bill sau khi khách thanh toán               ║
╠═══════════════╪══════════════════════════════════════════════════╣
║ 3. CUSTOMER   │ Khách hàng (Không cần đăng nhập)                ║
║               │ - Quét QR tại bàn → Xem menu                     ║
║               │ - Chọn món → Thanh toán → Tạo đơn hàng          ║
╚══════════════════════════════════════════════════════════════════╝

FLOW CHÍNH:
┌─────────┐    ┌─────────┐    ┌───────────┐    ┌─────────────┐
│ Quét QR │ -> │ Xem Menu│ -> │ Chọn món  │ -> │ Thanh toán  │
└─────────┘    └─────────┘    └───────────┘    └──────┬──────┘
                                                      │
                                                      v
┌─────────┐    ┌─────────┐    ┌───────────┐    ┌─────────────┐
│ In Bill │ <- │Hoàn thành│<- │ Phục vụ   │ <- │ Tạo đơn hàng│
└─────────┘    └─────────┘    └───────────┘    └─────────────┘
*/

-- ================================================================
-- XÓA BẢNG CŨ (NẾU CÓ) - THEO THỨ TỰ FOREIGN KEY
-- ================================================================
-- ================================================================
-- 1. BẢNG NGƯỜI DÙNG (users)
-- Ưu tiên: ★★★★★ (Quan trọng nhất - Core)
-- ================================================================
/*
Mô tả: Lưu thông tin người dùng hệ thống
- admin: Quản trị viên (1 người - bạn)
- restaurant: Chủ nhà hàng (đăng ký để sử dụng)
*/
CREATE TABLE users (
    id INT IDENTITY(1,1) PRIMARY KEY,
    email NVARCHAR(255) NOT NULL UNIQUE,
    password NVARCHAR(255) NOT NULL,             -- Mã hóa bcrypt
    name NVARCHAR(255) NOT NULL,
    role NVARCHAR(20) NOT NULL DEFAULT 'restaurant',
    avatar NVARCHAR(MAX) NULL,
    phone NVARCHAR(20) NULL,
    is_active BIT DEFAULT 1,
    last_login DATETIME2 NULL,
    created_at DATETIME2 DEFAULT GETDATE(),
    updated_at DATETIME2 DEFAULT GETDATE(),
    
    CONSTRAINT CK_users_role CHECK (role IN ('admin', 'restaurant'))
);
GO

-- Index cho tìm kiếm
CREATE INDEX IX_users_email ON users(email);
CREATE INDEX IX_users_role ON users(role);
GO

-- ================================================================
-- 2. BẢNG GÓI DỊCH VỤ (packages)
-- Ưu tiên: ★★★★★ (Quan trọng - Tính phí)
-- ================================================================
/*
Mô tả: Các gói dịch vụ cho nhà hàng
- Basic: Nhà hàng nhỏ (30 món, 10 bàn)
- Pro: Nhà hàng vừa (80 món, 25 bàn)  
- Premium: Nhà hàng lớn (Không giới hạn)
*/
CREATE TABLE packages (
    id INT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(50) NOT NULL UNIQUE,           -- 'Basic', 'Pro', 'Premium'
    display_name NVARCHAR(100) NOT NULL,
    description NVARCHAR(500) NULL,
    
    -- Giá (VND)
    monthly_price DECIMAL(12,0) NOT NULL,
    yearly_price DECIMAL(12,0) NOT NULL,
    
    -- Giới hạn (-1 = không giới hạn)
    max_menu_items INT DEFAULT 30,
    max_tables INT DEFAULT 10,
    max_categories INT DEFAULT 5,
    
    -- Tính năng (JSON array)
    features NVARCHAR(MAX) NULL,                 -- '["Tính năng 1", "Tính năng 2"]'
    
    is_popular BIT DEFAULT 0,                    -- Gói được đề xuất
    is_active BIT DEFAULT 1,
    sort_order INT DEFAULT 0,
    
    created_at DATETIME2 DEFAULT GETDATE()
);
GO

-- ================================================================
-- 3. BẢNG NHÀ HÀNG (restaurants)
-- Ưu tiên: ★★★★★ (Quan trọng - Core)
-- ================================================================
/*
Mô tả: Thông tin nhà hàng đăng ký sử dụng hệ thống
- Mỗi user (restaurant) sở hữu 1 nhà hàng
- Liên kết với gói dịch vụ đang sử dụng
*/
CREATE TABLE restaurants (
    id INT IDENTITY(1,1) PRIMARY KEY,
    owner_id INT NOT NULL,                       -- FK -> users.id
    package_id INT NOT NULL,                     -- FK -> packages.id
    
    -- Thông tin cơ bản
    name NVARCHAR(255) NOT NULL,
    slug NVARCHAR(100) NOT NULL UNIQUE,          -- URL-friendly: "nha-hang-abc"
    description NVARCHAR(1000) NULL,
    logo NVARCHAR(MAX) NULL,                     -- Base64 hoặc URL
    
    -- Liên hệ
    phone NVARCHAR(20) NULL,
    email NVARCHAR(255) NULL,
    address NVARCHAR(500) NULL,
    
    -- Cài đặt
    is_open BIT DEFAULT 1,                       -- Đang mở cửa?
    tax_rate DECIMAL(5,2) DEFAULT 10.00,         -- Thuế VAT (%)
    service_charge DECIMAL(5,2) DEFAULT 5.00,    -- Phí dịch vụ (%)
    currency NVARCHAR(10) DEFAULT 'VND',
    
    -- Gói dịch vụ
    package_start_date DATE NOT NULL,
    package_end_date DATE NOT NULL,
    package_status NVARCHAR(20) DEFAULT 'active', -- 'active', 'expired', 'cancelled'
    
    -- Trạng thái
    status NVARCHAR(20) DEFAULT 'active',        -- 'active', 'suspended', 'deleted'
    
    created_at DATETIME2 DEFAULT GETDATE(),
    updated_at DATETIME2 DEFAULT GETDATE(),
    
    CONSTRAINT FK_restaurants_owner FOREIGN KEY (owner_id) REFERENCES users(id),
    CONSTRAINT FK_restaurants_package FOREIGN KEY (package_id) REFERENCES packages(id),
    CONSTRAINT CK_restaurants_status CHECK (status IN ('active', 'suspended', 'deleted')),
    CONSTRAINT CK_restaurants_package_status CHECK (package_status IN ('active', 'expired', 'cancelled'))
);
GO

CREATE INDEX IX_restaurants_owner ON restaurants(owner_id);
CREATE INDEX IX_restaurants_slug ON restaurants(slug);
CREATE INDEX IX_restaurants_status ON restaurants(status);
GO

-- ================================================================
-- 4. BẢNG CÀI ĐẶT THANH TOÁN (payment_settings)
-- Ưu tiên: ★★★★☆ (Quan trọng - Thanh toán QR)
-- ================================================================
/*
Mô tả: Cài đặt thanh toán QR cho mỗi nhà hàng
*/
CREATE TABLE payment_settings (
    id INT IDENTITY(1,1) PRIMARY KEY,
    restaurant_id INT NOT NULL UNIQUE,           -- FK -> restaurants.id
    
    -- Thông tin ngân hàng
    bank_name NVARCHAR(100) NULL,
    account_number NVARCHAR(50) NULL,
    account_name NVARCHAR(255) NULL,
    
    -- QR Code
    qr_image NVARCHAR(MAX) NULL,                 -- Base64 image
    
    -- Phương thức thanh toán được bật
    accept_cash BIT DEFAULT 1,
    accept_qr BIT DEFAULT 0,
    accept_momo BIT DEFAULT 0,
    accept_vnpay BIT DEFAULT 0,
    
    created_at DATETIME2 DEFAULT GETDATE(),
    updated_at DATETIME2 DEFAULT GETDATE(),
    
    CONSTRAINT FK_payment_settings_restaurant FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE
);
GO

-- ================================================================
-- 5. BẢNG BÀN ĂN (tables)
-- Ưu tiên: ★★★★★ (Quan trọng - Core)
-- ================================================================
/*
Mô tả: Danh sách bàn của nhà hàng
- Mỗi bàn có mã QR riêng để khách quét
- Trạng thái: available (trống), occupied (có khách), reserved (đã đặt)
*/
CREATE TABLE tables (
    id INT IDENTITY(1,1) PRIMARY KEY,
    restaurant_id INT NOT NULL,                  -- FK -> restaurants.id
    
    table_number INT NOT NULL,                   -- Số bàn: 1, 2, 3...
    name NVARCHAR(50) NULL,                      -- Tên hiển thị: "Bàn VIP 1"
    capacity INT DEFAULT 4,                      -- Sức chứa (người)
    
    status NVARCHAR(20) DEFAULT 'available',     -- 'available', 'occupied', 'reserved'
    
    -- QR Code URL: /{slug}/menu/{table_number}
    qr_code NVARCHAR(MAX) NULL,                  -- Base64 QR image (optional)
    
    is_active BIT DEFAULT 1,
    created_at DATETIME2 DEFAULT GETDATE(),
    updated_at DATETIME2 DEFAULT GETDATE(),
    
    CONSTRAINT FK_tables_restaurant FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
    CONSTRAINT UQ_tables_number UNIQUE (restaurant_id, table_number),
    CONSTRAINT CK_tables_status CHECK (status IN ('available', 'occupied', 'reserved'))
);
GO

CREATE INDEX IX_tables_restaurant ON tables(restaurant_id);
CREATE INDEX IX_tables_status ON tables(restaurant_id, status);
GO

-- ================================================================
-- 6. BẢNG DANH MỤC (categories)
-- Ưu tiên: ★★★★☆ (Quan trọng - Menu)
-- ================================================================
/*
Mô tả: Danh mục món ăn (Khai vị, Món chính, Đồ uống...)
*/
CREATE TABLE categories (
    id INT IDENTITY(1,1) PRIMARY KEY,
    restaurant_id INT NOT NULL,                  -- FK -> restaurants.id
    
    name NVARCHAR(100) NOT NULL,
    description NVARCHAR(500) NULL,
    image NVARCHAR(MAX) NULL,
    
    sort_order INT DEFAULT 0,                    -- Thứ tự hiển thị
    status NVARCHAR(20) DEFAULT 'active',        -- 'active', 'inactive'
    
    created_at DATETIME2 DEFAULT GETDATE(),
    updated_at DATETIME2 DEFAULT GETDATE(),
    
    CONSTRAINT FK_categories_restaurant FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
    CONSTRAINT CK_categories_status CHECK (status IN ('active', 'inactive'))
);
GO

CREATE INDEX IX_categories_restaurant ON categories(restaurant_id);
GO

-- ================================================================
-- 7. BẢNG MÓN ĂN (menu_items)
-- Ưu tiên: ★★★★★ (Quan trọng - Core)
-- ================================================================
/*
Mô tả: Danh sách món ăn/đồ uống của nhà hàng
*/
CREATE TABLE menu_items (
    id INT IDENTITY(1,1) PRIMARY KEY,
    restaurant_id INT NOT NULL,                  -- FK -> restaurants.id
    category_id INT NOT NULL,                    -- FK -> categories.id
    
    name NVARCHAR(255) NOT NULL,
    description NVARCHAR(1000) NULL,
    price DECIMAL(12,0) NOT NULL,                -- Giá (VND)
    image NVARCHAR(MAX) NULL,
    
    -- Tùy chọn
    options NVARCHAR(MAX) NULL,                  -- JSON: [{"name": "Size", "choices": ["S", "M", "L"]}]
    tags NVARCHAR(MAX) NULL,                     -- JSON: ["Bán chạy", "Mới"]
    
    -- Vị trí chuẩn bị
    prep_location NVARCHAR(20) DEFAULT 'kitchen', -- 'kitchen', 'bar'
    prep_time INT DEFAULT 15,                    -- Thời gian chuẩn bị (phút)
    
    sort_order INT DEFAULT 0,
    status NVARCHAR(20) DEFAULT 'active',        -- 'active', 'inactive', 'out_of_stock'
    
    created_at DATETIME2 DEFAULT GETDATE(),
    updated_at DATETIME2 DEFAULT GETDATE(),
    
    CONSTRAINT FK_menu_items_restaurant FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE,
    CONSTRAINT FK_menu_items_category FOREIGN KEY (category_id) REFERENCES categories(id),
    CONSTRAINT CK_menu_items_status CHECK (status IN ('active', 'inactive', 'out_of_stock'))
);
GO

CREATE INDEX IX_menu_items_restaurant ON menu_items(restaurant_id);
CREATE INDEX IX_menu_items_category ON menu_items(category_id);
CREATE INDEX IX_menu_items_status ON menu_items(restaurant_id, status);
GO

-- ================================================================
-- 8. BẢNG ĐƠN HÀNG (orders)
-- Ưu tiên: ★★★★★ (Quan trọng nhất - Core)
-- ================================================================
/*
Mô tả: Đơn hàng của khách
Flow trạng thái:
  pending -> confirmed -> preparing -> ready -> serving -> completed
                      \-> cancelled

Thanh toán:
  - payment_timing: 'before' (trước), 'after' (sau khi ăn)
  - payment_status: 'unpaid', 'paid'
  - payment_method: 'cash', 'qr', 'momo', 'vnpay'
*/
CREATE TABLE orders (
    id INT IDENTITY(1,1) PRIMARY KEY,
    restaurant_id INT NOT NULL,                  -- FK -> restaurants.id
    table_id INT NOT NULL,                       -- FK -> tables.id
    
    -- Mã đơn hàng: ORD-2026-0001
    order_number NVARCHAR(50) NOT NULL,
    
    -- Thông tin khách (optional)
    customer_name NVARCHAR(255) NULL,
    customer_phone NVARCHAR(20) NULL,
    
    -- Trạng thái đơn hàng
    status NVARCHAR(20) DEFAULT 'pending',
    -- 'pending', 'confirmed', 'preparing', 'ready', 'serving', 'completed', 'cancelled'
    
    -- Thanh toán
    payment_timing NVARCHAR(10) DEFAULT 'after', -- 'before', 'after'
    payment_method NVARCHAR(20) NULL,            -- 'cash', 'qr', 'momo', 'vnpay'
    payment_status NVARCHAR(20) DEFAULT 'unpaid', -- 'unpaid', 'paid'
    paid_at DATETIME2 NULL,
    
    -- Tính tiền
    subtotal DECIMAL(12,0) DEFAULT 0,            -- Tạm tính
    tax_amount DECIMAL(12,0) DEFAULT 0,          -- Thuế VAT
    service_charge DECIMAL(12,0) DEFAULT 0,      -- Phí dịch vụ
    discount_amount DECIMAL(12,0) DEFAULT 0,     -- Giảm giá
    total_amount DECIMAL(12,0) DEFAULT 0,        -- Tổng cộng
    
    -- Ghi chú
    notes NVARCHAR(1000) NULL,
    cancel_reason NVARCHAR(500) NULL,
    
    created_at DATETIME2 DEFAULT GETDATE(),
    updated_at DATETIME2 DEFAULT GETDATE(),
    completed_at DATETIME2 NULL,
    
    CONSTRAINT FK_orders_restaurant FOREIGN KEY (restaurant_id) REFERENCES restaurants(id),
    CONSTRAINT FK_orders_table FOREIGN KEY (table_id) REFERENCES tables(id),
    CONSTRAINT CK_orders_status CHECK (status IN ('pending', 'confirmed', 'preparing', 'ready', 'serving', 'completed', 'cancelled')),
    CONSTRAINT CK_orders_payment_timing CHECK (payment_timing IN ('before', 'after')),
    CONSTRAINT CK_orders_payment_status CHECK (payment_status IN ('unpaid', 'paid'))
);
GO

CREATE INDEX IX_orders_restaurant ON orders(restaurant_id);
CREATE INDEX IX_orders_table ON orders(table_id);
CREATE INDEX IX_orders_status ON orders(restaurant_id, status);
CREATE INDEX IX_orders_date ON orders(restaurant_id, created_at);
CREATE INDEX IX_orders_number ON orders(order_number);
GO

-- ================================================================
-- 9. BẢNG CHI TIẾT ĐƠN HÀNG (order_items)
-- Ưu tiên: ★★★★★ (Quan trọng - Core)
-- ================================================================
/*
Mô tả: Chi tiết món trong đơn hàng
*/
CREATE TABLE order_items (
    id INT IDENTITY(1,1) PRIMARY KEY,
    order_id INT NOT NULL,                       -- FK -> orders.id
    menu_item_id INT NOT NULL,                   -- FK -> menu_items.id
    
    -- Snapshot thông tin món (phòng trường hợp món bị sửa sau)
    item_name NVARCHAR(255) NOT NULL,
    item_price DECIMAL(12,0) NOT NULL,
    
    quantity INT NOT NULL DEFAULT 1,
    
    -- Tùy chọn đã chọn
    selected_options NVARCHAR(MAX) NULL,         -- JSON: {"Size": "L", "Đá": "Ít"}
    
    -- Ghi chú cho món
    notes NVARCHAR(500) NULL,
    
    -- Trạng thái chuẩn bị món
    prep_status NVARCHAR(20) DEFAULT 'pending',  -- 'pending', 'preparing', 'completed'
    prep_location NVARCHAR(20) DEFAULT 'kitchen',
    
    -- Tính tiền
    line_total DECIMAL(12,0) NOT NULL,           -- item_price * quantity
    
    created_at DATETIME2 DEFAULT GETDATE(),
    
    CONSTRAINT FK_order_items_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    CONSTRAINT FK_order_items_menu FOREIGN KEY (menu_item_id) REFERENCES menu_items(id),
    CONSTRAINT CK_order_items_prep_status CHECK (prep_status IN ('pending', 'preparing', 'completed'))
);
GO

CREATE INDEX IX_order_items_order ON order_items(order_id);
CREATE INDEX IX_order_items_prep ON order_items(order_id, prep_status);
GO


-- ================================================================
-- DỮ LIỆU MẪU
-- ================================================================

-- 1. Tạo Admin
INSERT INTO users (email, password, name, role, phone)
VALUES ('admin@fbmanager.com', '$2a$10$xxxxxxxxxxx', 'Admin Hệ Thống', 'admin', '0900000000');

-- 2. Tạo các gói dịch vụ
INSERT INTO packages (name, display_name, description, monthly_price, yearly_price, max_menu_items, max_tables, max_categories, features, is_popular, sort_order)
VALUES 
('Basic', 'Gói Cơ Bản', 'Phù hợp cho nhà hàng nhỏ, quán ăn gia đình', 199000, 1990000, 30, 10, 5, 
 '["Quản lý 30 món ăn", "Tối đa 10 bàn", "Đặt món qua QR", "Thanh toán tiền mặt", "Báo cáo cơ bản"]', 0, 1),

('Pro', 'Gói Chuyên Nghiệp', 'Phù hợp cho nhà hàng vừa, có đội ngũ phục vụ', 499000, 4990000, 80, 25, 10, 
 '["Quản lý 80 món ăn", "Tối đa 25 bàn", "Đặt món qua QR", "Thanh toán QR/MoMo/VNPay", "Báo cáo chi tiết", "Marketing cơ bản"]', 1, 2),

('Premium', 'Gói Cao Cấp', 'Phù hợp cho nhà hàng lớn, chuỗi nhà hàng', 999000, 9990000, -1, -1, -1, 
 '["Không giới hạn món ăn", "Không giới hạn bàn", "Đặt món qua QR", "Tất cả phương thức thanh toán", "Báo cáo nâng cao", "Marketing đầy đủ", "Hỗ trợ ưu tiên 24/7"]', 0, 3);

GO

-- ================================================================
-- VIEWS HỮU ÍCH
-- ================================================================

-- View: Thống kê đơn hàng theo ngày
CREATE VIEW vw_daily_order_stats AS
SELECT 
    restaurant_id,
    CAST(created_at AS DATE) AS order_date,
    COUNT(*) AS total_orders,
    SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) AS completed_orders,
    SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) AS cancelled_orders,
    SUM(CASE WHEN payment_status = 'paid' THEN total_amount ELSE 0 END) AS total_revenue
FROM orders
GROUP BY restaurant_id, CAST(created_at AS DATE);
GO

-- View: Thống kê bàn
CREATE VIEW vw_table_stats AS
SELECT 
    restaurant_id,
    COUNT(*) AS total_tables,
    SUM(CASE WHEN status = 'available' THEN 1 ELSE 0 END) AS available_tables,
    SUM(CASE WHEN status = 'occupied' THEN 1 ELSE 0 END) AS occupied_tables
FROM tables
WHERE is_active = 1
GROUP BY restaurant_id;
GO

-- ================================================================
-- STORED PROCEDURES
-- ================================================================

-- SP: Tạo đơn hàng mới
CREATE PROCEDURE sp_create_order
    @restaurant_id INT,
    @table_id INT,
    @payment_timing NVARCHAR(10) = 'after',
    @customer_name NVARCHAR(255) = NULL,
    @notes NVARCHAR(1000) = NULL
AS
BEGIN
    DECLARE @order_number NVARCHAR(50);
    DECLARE @order_count INT;
    
    -- Tạo mã đơn hàng
    SELECT @order_count = COUNT(*) + 1 FROM orders WHERE restaurant_id = @restaurant_id;
    SET @order_number = 'ORD-' + CAST(YEAR(GETDATE()) AS NVARCHAR) + '-' + RIGHT('0000' + CAST(@order_count AS NVARCHAR), 4);
    
    -- Insert đơn hàng
    INSERT INTO orders (restaurant_id, table_id, order_number, payment_timing, customer_name, notes)
    VALUES (@restaurant_id, @table_id, @order_number, @payment_timing, @customer_name, @notes);
    
    -- Cập nhật trạng thái bàn
    UPDATE tables SET status = 'occupied', updated_at = GETDATE() WHERE id = @table_id;
    
    SELECT SCOPE_IDENTITY() AS order_id, @order_number AS order_number;
END;
GO

-- SP: Cập nhật trạng thái đơn hàng
CREATE PROCEDURE sp_update_order_status
    @order_id INT,
    @new_status NVARCHAR(20)
AS
BEGIN
    DECLARE @table_id INT;
    
    UPDATE orders 
    SET status = @new_status, 
        updated_at = GETDATE(),
        completed_at = CASE WHEN @new_status = 'completed' THEN GETDATE() ELSE completed_at END
    WHERE id = @order_id;
    
    -- Nếu hoàn thành hoặc hủy, giải phóng bàn
    IF @new_status IN ('completed', 'cancelled')
    BEGIN
        SELECT @table_id = table_id FROM orders WHERE id = @order_id;
        UPDATE tables SET status = 'available', updated_at = GETDATE() WHERE id = @table_id;
    END
END;
GO

-- SP: Thanh toán đơn hàng
CREATE PROCEDURE sp_pay_order
    @order_id INT,
    @payment_method NVARCHAR(20)
AS
BEGIN
    UPDATE orders 
    SET payment_status = 'paid',
        payment_method = @payment_method,
        paid_at = GETDATE(),
        updated_at = GETDATE()
    WHERE id = @order_id;
END;
GO

PRINT 'Database schema created successfully!';
GO
