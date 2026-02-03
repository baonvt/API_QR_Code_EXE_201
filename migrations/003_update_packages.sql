-- Migration: Update packages with new pricing and features
-- Date: 2026-02-03
-- Description: Xóa packages cũ và thêm packages mới với giá và tính năng đã cập nhật

-- =============================================
-- STEP 1: Xóa tất cả packages cũ
-- =============================================
DELETE FROM packages;

-- =============================================
-- STEP 2: Thêm packages mới
-- =============================================

-- Gói Starter (Dùng thử miễn phí)
INSERT INTO packages (name, display_name, description, monthly_price, yearly_price, max_menu_items, max_tables, max_categories, features, is_popular, is_active, sort_order, created_at, updated_at)
VALUES (
    'Starter',
    'Starter',
    'Gói dùng thử miễn phí 7 ngày',
    0,
    0,
    10,
    3,
    3,
    '["Quản lý 10 món ăn", "Tối đa 3 bàn", "Đặt món qua QR", "Thanh toán tiền mặt"]',
    false,
    true,
    0,
    NOW(),
    NOW()
);

-- Gói Basic (Cơ Bản)
INSERT INTO packages (name, display_name, description, monthly_price, yearly_price, max_menu_items, max_tables, max_categories, features, is_popular, is_active, sort_order, created_at, updated_at)
VALUES (
    'Basic',
    'Gói Cơ Bản',
    'Dành cho quán nhỏ, phục vụ dưới 40 khách/lượt',
    229000,
    2290000,
    30,
    10,
    3,
    '["Tạo thực đơn (tối đa 30 món)", "Gọi món bằng mã QR", "Thống kê doanh thu cơ bản", "Quản lý tối đa 10 bàn", "3 danh mục món ăn (Món chính - Đồ uống - Tráng miệng)", "Hỗ trợ qua email"]',
    false,
    true,
    1,
    NOW(),
    NOW()
);

-- Gói Pro (Chuyên Nghiệp)
INSERT INTO packages (name, display_name, description, monthly_price, yearly_price, max_menu_items, max_tables, max_categories, features, is_popular, is_active, sort_order, created_at, updated_at)
VALUES (
    'Pro',
    'Gói Chuyên Nghiệp',
    'Dành cho quán cà phê và nhà hàng đang phát triển',
    270000,
    2700000,
    80,
    25,
    6,
    '["Bao gồm tất cả tính năng của Gói Cơ Bản", "Quản lý nhân viên phục vụ", "Lưu trữ đám mây", "Quản lý tối đa 25 bàn", "Tạo đến 80 món ăn/đồ uống", "6 danh mục món ăn (Món chính - Món phụ - Đồ nướng - Lẩu - Đồ uống - Tráng miệng)", "Báo cáo doanh thu chi tiết theo danh mục", "Hỗ trợ 24/7"]',
    true,
    true,
    2,
    NOW(),
    NOW()
);

-- Gói Premium (Cao Cấp)
INSERT INTO packages (name, display_name, description, monthly_price, yearly_price, max_menu_items, max_tables, max_categories, features, is_popular, is_active, sort_order, created_at, updated_at)
VALUES (
    'Premium',
    'Gói Cao Cấp',
    'Dành cho chuỗi hoặc nhà hàng có nhiều chi nhánh',
    279000,
    2790000,
    -1,
    -1,
    -1,
    '["Bao gồm tất cả tính năng của Gói Chuyên Nghiệp", "Hỗ trợ kỹ thuật ưu tiên", "Kết nối nhiều chi nhánh", "Đánh giá & đặt chỗ của khách hàng", "Quản lý không giới hạn số bàn và món ăn", "Tạo danh mục tùy chỉnh linh hoạt (theo chi nhánh, theo loại hình phục vụ)", "Tích hợp thực đơn số đồng bộ giữa các chi nhánh", "API tích hợp", "Hỗ trợ ưu tiên 24/7", "Tùy chỉnh theo yêu cầu"]',
    false,
    true,
    3,
    NOW(),
    NOW()
);

-- =============================================
-- VERIFICATION: Kiểm tra kết quả
-- =============================================
-- SELECT id, name, display_name, monthly_price, yearly_price, max_menu_items, max_tables, max_categories, is_popular FROM packages ORDER BY sort_order;
