# 📚 API DOCUMENTATION - HỆ THỐNG QUẢN LÝ NHÀ HÀNG

> **Version:** 2.0  
> **Base URL:** `https://api.fbmanager.com/v1`  
> **Updated:** 27/01/2026

---

## 📋 MỤC LỤC

1. [Tổng quan](#tổng-quan)
2. [Authentication](#authentication)
3. [API theo mức độ ưu tiên](#api-theo-mức-độ-ưu-tiên)
   - [★★★★★ Ưu tiên 1: Core APIs](#ưu-tiên-1-core-apis)
   - [★★★★☆ Ưu tiên 2: Business APIs](#ưu-tiên-2-business-apis)
   - [★★★☆☆ Ưu tiên 3: Feature APIs](#ưu-tiên-3-feature-apis)
   - [★★☆☆☆ Ưu tiên 4: Enhancement APIs](#ưu-tiên-4-enhancement-apis)

---

## 🎯 TỔNG QUAN

### Roles & Permissions

| Role | Mô tả | Cần đăng nhập |
|------|-------|---------------|
| `admin` | Quản trị viên hệ thống | ✅ Có |
| `restaurant` | Chủ nhà hàng | ✅ Có |
| `customer` | Khách hàng | ❌ Không |

### Response Format

```json
{
  "success": true,
  "data": { ... },
  "message": "Thành công",
  "error": null
}
```

### Error Format

```json
{
  "success": false,
  "data": null,
  "message": "Lỗi",
  "error": {
    "code": "ERROR_CODE",
    "details": "Chi tiết lỗi"
  }
}
```

### HTTP Status Codes

| Code | Mô tả |
|------|-------|
| 200 | Thành công |
| 201 | Tạo mới thành công |
| 400 | Bad Request |
| 401 | Chưa đăng nhập |
| 403 | Không có quyền |
| 404 | Không tìm thấy |
| 500 | Lỗi server |

---

## 🔐 AUTHENTICATION

### Headers

```
Authorization: Bearer <access_token>
Content-Type: application/json
```

---

# 📊 API THEO MỨC ĐỘ ƯU TIÊN

---

## ★★★★★ ƯU TIÊN 1: CORE APIs
> **Phải có ngay từ đầu - Hệ thống không thể hoạt động nếu thiếu**

---

### 1.1 AUTH - Xác thực

#### `POST /auth/login`
> **Đăng nhập** (Admin & Restaurant)

**Request:**
```json
{
  "email": "restaurant@example.com",
  "password": "123456"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "email": "restaurant@example.com",
      "name": "Nguyễn Văn A",
      "role": "restaurant",
      "avatar": null
    },
    "access_token": "eyJhbGc...",
    "expires_in": 86400
  }
}
```

---

#### `POST /auth/register`
> **Đăng ký nhà hàng mới**

**Request:**
```json
{
  "email": "newrestaurant@example.com",
  "password": "123456",
  "name": "Nguyễn Văn B",
  "phone": "0901234567",
  "restaurant_name": "Nhà hàng ABC",
  "package_id": 1
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 2,
      "email": "newrestaurant@example.com",
      "name": "Nguyễn Văn B",
      "role": "restaurant"
    },
    "restaurant": {
      "id": 1,
      "name": "Nhà hàng ABC",
      "slug": "nha-hang-abc"
    }
  }
}
```

---

#### `POST /auth/logout`
> **Đăng xuất** | 🔒 Yêu cầu đăng nhập

**Response:**
```json
{
  "success": true,
  "message": "Đăng xuất thành công"
}
```

---

#### `GET /auth/me`
> **Lấy thông tin user hiện tại** | 🔒 Yêu cầu đăng nhập

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "restaurant@example.com",
    "name": "Nguyễn Văn A",
    "role": "restaurant",
    "avatar": null,
    "phone": "0901234567",
    "restaurant": {
      "id": 1,
      "name": "Nhà hàng ABC",
      "slug": "nha-hang-abc"
    }
  }
}
```

---

### 1.2 RESTAURANTS - Nhà hàng

#### `GET /restaurants/:slug`
> **Lấy thông tin nhà hàng theo slug** | 🌐 Public (cho khách xem)

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Nhà hàng ABC",
    "slug": "nha-hang-abc",
    "description": "Nhà hàng ẩm thực Việt",
    "logo": "https://...",
    "phone": "0901234567",
    "address": "123 Nguyễn Huệ, Q1, HCM",
    "is_open": true,
    "tax_rate": 10,
    "service_charge": 5
  }
}
```

---

#### `PUT /restaurants/:id`
> **Cập nhật thông tin nhà hàng** | 🔒 Restaurant

**Request:**
```json
{
  "name": "Nhà hàng ABC Updated",
  "description": "Mô tả mới",
  "phone": "0909999999",
  "address": "456 Lê Lợi, Q1, HCM",
  "is_open": true,
  "tax_rate": 10,
  "service_charge": 5
}
```

---

### 1.3 TABLES - Bàn ăn

#### `GET /restaurants/:id/tables`
> **Lấy danh sách bàn** | 🔒 Restaurant

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "table_number": 1,
      "name": "Bàn 1",
      "capacity": 4,
      "status": "available",
      "qr_url": "/nha-hang-abc/menu/1"
    },
    {
      "id": 2,
      "table_number": 2,
      "name": "Bàn VIP",
      "capacity": 8,
      "status": "occupied"
    }
  ]
}
```

---

#### `POST /restaurants/:id/tables`
> **Tạo bàn mới** | 🔒 Restaurant

**Request:**
```json
{
  "table_number": 3,
  "name": "Bàn 3",
  "capacity": 4
}
```

---

#### `PUT /tables/:id`
> **Cập nhật bàn** | 🔒 Restaurant

**Request:**
```json
{
  "name": "Bàn VIP 1",
  "capacity": 6,
  "status": "available"
}
```

---

#### `DELETE /tables/:id`
> **Xóa bàn** | 🔒 Restaurant

---

### 1.4 CATEGORIES - Danh mục

#### `GET /restaurants/:id/categories`
> **Lấy danh sách danh mục** | 🌐 Public

**Query params:**
- `status`: `active` | `inactive` | `all` (default: `active`)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Khai vị",
      "description": "Các món khai vị",
      "image": null,
      "sort_order": 1,
      "status": "active",
      "items_count": 5
    }
  ]
}
```

---

#### `POST /restaurants/:id/categories`
> **Tạo danh mục** | 🔒 Restaurant

**Request:**
```json
{
  "name": "Món chính",
  "description": "Các món ăn chính",
  "image": "base64...",
  "sort_order": 2
}
```

---

#### `PUT /categories/:id`
> **Cập nhật danh mục** | 🔒 Restaurant

---

#### `DELETE /categories/:id`
> **Xóa danh mục** | 🔒 Restaurant

---

### 1.5 MENU ITEMS - Món ăn

#### `GET /restaurants/:id/menu`
> **Lấy toàn bộ menu** | 🌐 Public

**Query params:**
- `category_id`: Lọc theo danh mục
- `status`: `active` | `inactive` | `out_of_stock`

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Phở bò",
      "description": "Phở bò truyền thống",
      "price": 45000,
      "image": "https://...",
      "category_id": 1,
      "category_name": "Món chính",
      "options": [
        {"name": "Size", "choices": ["Nhỏ", "Lớn"]}
      ],
      "tags": ["Bán chạy"],
      "status": "active"
    }
  ]
}
```

---

#### `GET /categories/:id/items`
> **Lấy món theo danh mục** | 🌐 Public

---

#### `POST /restaurants/:id/menu`
> **Tạo món mới** | 🔒 Restaurant

**Request:**
```json
{
  "category_id": 1,
  "name": "Bún bò Huế",
  "description": "Bún bò cay nồng đặc trưng",
  "price": 50000,
  "image": "base64...",
  "options": [
    {"name": "Cay", "choices": ["Không cay", "Ít cay", "Cay vừa", "Siêu cay"]}
  ],
  "tags": ["Mới"],
  "prep_location": "kitchen",
  "prep_time": 20
}
```

---

#### `PUT /menu/:id`
> **Cập nhật món** | 🔒 Restaurant

---

#### `DELETE /menu/:id`
> **Xóa món** | 🔒 Restaurant

---

### 1.6 ORDERS - Đơn hàng (QUAN TRỌNG NHẤT)

#### `GET /restaurants/:id/orders`
> **Lấy danh sách đơn hàng** | 🔒 Restaurant

**Query params:**
- `status`: `pending` | `confirmed` | `preparing` | `ready` | `serving` | `completed` | `cancelled`
- `date`: `2026-01-27` (lọc theo ngày)
- `table_id`: Lọc theo bàn
- `page`: Trang (default: 1)
- `limit`: Số lượng (default: 20)

**Response:**
```json
{
  "success": true,
  "data": {
    "orders": [
      {
        "id": 1,
        "order_number": "ORD-2026-0001",
        "table_id": 1,
        "table_number": 1,
        "table_name": "Bàn 1",
        "status": "pending",
        "payment_status": "unpaid",
        "payment_timing": "after",
        "subtotal": 150000,
        "tax_amount": 15000,
        "service_charge": 7500,
        "total_amount": 172500,
        "items_count": 3,
        "created_at": "2026-01-27T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "total_pages": 3
    }
  }
}
```

---

#### `GET /orders/:id`
> **Lấy chi tiết đơn hàng** | 🔒 Restaurant / 🌐 Customer (bằng order_number)

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "order_number": "ORD-2026-0001",
    "table_id": 1,
    "table_name": "Bàn 1",
    "status": "preparing",
    "payment_status": "unpaid",
    "payment_timing": "after",
    "payment_method": null,
    "subtotal": 150000,
    "tax_amount": 15000,
    "service_charge": 7500,
    "discount_amount": 0,
    "total_amount": 172500,
    "notes": "Không hành",
    "items": [
      {
        "id": 1,
        "menu_item_id": 1,
        "name": "Phở bò",
        "price": 45000,
        "quantity": 2,
        "selected_options": {"Size": "Lớn"},
        "notes": "Ít bánh",
        "prep_status": "preparing",
        "line_total": 90000
      },
      {
        "id": 2,
        "menu_item_id": 5,
        "name": "Nước cam",
        "price": 30000,
        "quantity": 2,
        "selected_options": {"Đá": "Nhiều đá"},
        "notes": null,
        "prep_status": "completed",
        "line_total": 60000
      }
    ],
    "created_at": "2026-01-27T10:30:00Z",
    "updated_at": "2026-01-27T10:35:00Z"
  }
}
```

---

#### `POST /restaurants/:slug/orders`
> **Tạo đơn hàng mới** | 🌐 Customer (Khách đặt món)

**Request:**
```json
{
  "table_number": 1,
  "payment_timing": "after",
  "customer_name": "Khách vãng lai",
  "notes": "Không hành",
  "items": [
    {
      "menu_item_id": 1,
      "quantity": 2,
      "selected_options": {"Size": "Lớn"},
      "notes": "Ít bánh"
    },
    {
      "menu_item_id": 5,
      "quantity": 2,
      "selected_options": {"Đá": "Nhiều đá"}
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "order_number": "ORD-2026-0001",
    "status": "pending",
    "total_amount": 172500,
    "tracking_url": "/nha-hang-abc/order/1"
  },
  "message": "Đơn hàng đã được gửi thành công!"
}
```

---

#### `PUT /orders/:id/status`
> **Cập nhật trạng thái đơn hàng** | 🔒 Restaurant

**Request:**
```json
{
  "status": "confirmed",
  "note": "Đã xác nhận đơn hàng"
}
```

**Valid transitions:**
```
pending -> confirmed, cancelled
confirmed -> preparing, cancelled
preparing -> ready, cancelled
ready -> serving
serving -> completed
```

---

#### `POST /orders/:id/items`
> **Thêm món vào đơn hàng hiện tại** | 🌐 Customer

**Request:**
```json
{
  "items": [
    {
      "menu_item_id": 3,
      "quantity": 1,
      "notes": "Thêm món"
    }
  ]
}
```

---

#### `PUT /orders/:id/pay`
> **Thanh toán đơn hàng** | 🔒 Restaurant

**Request:**
```json
{
  "payment_method": "cash"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "order_id": 1,
    "payment_status": "paid",
    "payment_method": "cash",
    "paid_at": "2026-01-27T11:00:00Z",
    "total_amount": 172500
  },
  "message": "Thanh toán thành công!"
}
```

---

#### `GET /orders/:id/bill`
> **Lấy thông tin in bill** | 🔒 Restaurant

**Response:**
```json
{
  "success": true,
  "data": {
    "restaurant": {
      "name": "Nhà hàng ABC",
      "address": "123 Nguyễn Huệ, Q1",
      "phone": "0901234567"
    },
    "order": {
      "order_number": "ORD-2026-0001",
      "table_name": "Bàn 1",
      "created_at": "2026-01-27T10:30:00Z",
      "completed_at": "2026-01-27T11:00:00Z"
    },
    "items": [
      {"name": "Phở bò (Lớn)", "quantity": 2, "price": 45000, "total": 90000},
      {"name": "Nước cam", "quantity": 2, "price": 30000, "total": 60000}
    ],
    "summary": {
      "subtotal": 150000,
      "tax_amount": 15000,
      "service_charge": 7500,
      "discount_amount": 0,
      "total_amount": 172500
    },
    "payment": {
      "method": "cash",
      "status": "paid",
      "paid_at": "2026-01-27T11:00:00Z"
    }
  }
}
```

---

## ★★★★☆ ƯU TIÊN 2: BUSINESS APIs
> **Cần thiết cho nghiệp vụ - Xây dựng sau khi có Core**

---

### 2.1 PACKAGES - Gói dịch vụ

#### `GET /packages`
> **Lấy danh sách gói dịch vụ** | 🌐 Public

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Basic",
      "display_name": "Gói Cơ Bản",
      "description": "Phù hợp cho nhà hàng nhỏ",
      "monthly_price": 199000,
      "yearly_price": 1990000,
      "max_menu_items": 30,
      "max_tables": 10,
      "max_categories": 5,
      "features": ["Quản lý 30 món ăn", "Tối đa 10 bàn", ...],
      "is_popular": false
    }
  ]
}
```

---

#### `POST /packages`
> **Tạo gói mới** | 🔒 Admin

---

#### `PUT /packages/:id`
> **Cập nhật gói** | 🔒 Admin

---

### 2.2 PAYMENT SETTINGS - Cài đặt thanh toán

#### `GET /restaurants/:id/payment-settings`
> **Lấy cài đặt thanh toán** | 🔒 Restaurant

---

#### `PUT /restaurants/:id/payment-settings`
> **Cập nhật cài đặt thanh toán** | 🔒 Restaurant

**Request:**
```json
{
  "bank_name": "Vietcombank",
  "account_number": "1234567890",
  "account_name": "NGUYEN VAN A",
  "qr_image": "base64...",
  "accept_cash": true,
  "accept_qr": true,
  "accept_momo": false,
  "accept_vnpay": false
}
```

---

### 2.3 STATISTICS - Thống kê

#### `GET /restaurants/:id/stats/overview`
> **Thống kê tổng quan** | 🔒 Restaurant

**Response:**
```json
{
  "success": true,
  "data": {
    "today": {
      "orders": 25,
      "revenue": 5000000,
      "avg_order_value": 200000
    },
    "this_month": {
      "orders": 500,
      "revenue": 100000000,
      "avg_order_value": 200000
    },
    "tables": {
      "total": 15,
      "available": 10,
      "occupied": 5
    },
    "orders_by_status": {
      "pending": 3,
      "confirmed": 2,
      "preparing": 5,
      "ready": 1,
      "serving": 2
    }
  }
}
```

---

#### `GET /restaurants/:id/stats/revenue`
> **Thống kê doanh thu** | 🔒 Restaurant

**Query params:**
- `period`: `day` | `week` | `month` | `year`
- `start_date`: `2026-01-01`
- `end_date`: `2026-01-31`

**Response:**
```json
{
  "success": true,
  "data": {
    "total_revenue": 100000000,
    "total_orders": 500,
    "avg_order_value": 200000,
    "chart_data": [
      {"date": "2026-01-01", "revenue": 3000000, "orders": 15},
      {"date": "2026-01-02", "revenue": 4500000, "orders": 22}
    ]
  }
}
```

---

#### `GET /restaurants/:id/stats/menu`
> **Thống kê món bán chạy** | 🔒 Restaurant

**Response:**
```json
{
  "success": true,
  "data": {
    "top_items": [
      {"id": 1, "name": "Phở bò", "quantity_sold": 150, "revenue": 6750000},
      {"id": 5, "name": "Nước cam", "quantity_sold": 200, "revenue": 6000000}
    ],
    "by_category": [
      {"category": "Món chính", "quantity_sold": 300, "revenue": 15000000},
      {"category": "Đồ uống", "quantity_sold": 400, "revenue": 12000000}
    ]
  }
}
```

---

## ★★★☆☆ ƯU TIÊN 3: FEATURE APIs
> **Tính năng bổ sung - Nâng cao trải nghiệm**

---

### 3.1 ADMIN - Quản lý hệ thống

#### `GET /admin/restaurants`
> **Lấy danh sách nhà hàng** | 🔒 Admin

**Query params:**
- `status`: `active` | `suspended` | `all`
- `package_status`: `active` | `expired`
- `search`: Tìm kiếm theo tên
- `page`, `limit`

---

#### `PUT /admin/restaurants/:id/status`
> **Cập nhật trạng thái nhà hàng** | 🔒 Admin

**Request:**
```json
{
  "status": "suspended",
  "reason": "Vi phạm điều khoản"
}
```

---

#### `GET /admin/stats`
> **Thống kê toàn hệ thống** | 🔒 Admin

**Response:**
```json
{
  "success": true,
  "data": {
    "total_restaurants": 150,
    "active_restaurants": 120,
    "new_this_month": 25,
    "total_revenue": 50000000,
    "by_package": [
      {"package": "Basic", "count": 80},
      {"package": "Pro", "count": 50},
      {"package": "Premium", "count": 20}
    ]
  }
}
```

---

### 3.2 USER PROFILE

#### `PUT /users/profile`
> **Cập nhật thông tin cá nhân** | 🔒 Logged in

**Request:**
```json
{
  "name": "Nguyễn Văn A",
  "phone": "0909999999",
  "avatar": "base64..."
}
```

---

#### `PUT /users/password`
> **Đổi mật khẩu** | 🔒 Logged in

**Request:**
```json
{
  "old_password": "123456",
  "new_password": "654321"
}
```

---

### 3.3 TABLE QR

#### `GET /tables/:id/qr`
> **Lấy QR code cho bàn** | 🔒 Restaurant

**Response:**
```json
{
  "success": true,
  "data": {
    "table_id": 1,
    "table_number": 1,
    "qr_url": "https://fbmanager.com/nha-hang-abc/menu/1",
    "qr_image": "base64..."
  }
}
```

---

#### `GET /restaurants/:id/qr-all`
> **Tải tất cả QR code** | 🔒 Restaurant

**Response:** ZIP file chứa tất cả QR code của các bàn

---

## ★★☆☆☆ ƯU TIÊN 4: ENHANCEMENT APIs
> **Mở rộng - Thêm sau khi hệ thống ổn định**

---

### 4.1 NOTIFICATIONS

#### `GET /notifications`
> **Lấy thông báo** | 🔒 Logged in

---

#### `PUT /notifications/:id/read`
> **Đánh dấu đã đọc** | 🔒 Logged in

---

### 4.2 MARKETING (Tương lai)

#### `GET /restaurants/:id/promotions`
> **Danh sách khuyến mãi**

#### `POST /restaurants/:id/promotions`
> **Tạo khuyến mãi**

---

### 4.3 REVIEWS (Tương lai)

#### `GET /restaurants/:slug/reviews`
> **Đánh giá nhà hàng**

#### `POST /restaurants/:slug/reviews`
> **Gửi đánh giá**

---

## 📱 CUSTOMER FLOW APIs (PUBLIC)

> **Không cần đăng nhập - Khách hàng sử dụng**

### Flow hoàn chỉnh:

```
1. GET  /restaurants/:slug                    # Xem thông tin nhà hàng
2. GET  /restaurants/:slug/categories         # Lấy danh mục
3. GET  /restaurants/:slug/menu               # Lấy menu
4. POST /restaurants/:slug/orders             # Đặt món (tạo đơn)
5. GET  /orders/:order_number                 # Theo dõi đơn hàng
6. POST /orders/:id/items                     # Gọi thêm món
```

---

## 🔑 API KEYS & WEBHOOKS (Tương lai)

### Webhooks cho tích hợp

```
POST /webhooks/order-created      # Khi có đơn mới
POST /webhooks/order-completed    # Khi đơn hoàn thành
POST /webhooks/payment-received   # Khi nhận thanh toán
```

---

## 📋 TÓM TẮT SỐ LƯỢNG APIs

| Ưu tiên | Nhóm | Số API | Trạng thái |
|---------|------|--------|------------|
| ★★★★★ | Auth | 4 | Bắt buộc |
| ★★★★★ | Restaurants | 2 | Bắt buộc |
| ★★★★★ | Tables | 4 | Bắt buộc |
| ★★★★★ | Categories | 4 | Bắt buộc |
| ★★★★★ | Menu Items | 4 | Bắt buộc |
| ★★★★★ | Orders | 7 | Bắt buộc |
| ★★★★☆ | Packages | 3 | Quan trọng |
| ★★★★☆ | Payment Settings | 2 | Quan trọng |
| ★★★★☆ | Statistics | 3 | Quan trọng |
| ★★★☆☆ | Admin | 3 | Cần thiết |
| ★★★☆☆ | User Profile | 2 | Cần thiết |
| ★★★☆☆ | Table QR | 2 | Cần thiết |
| ★★☆☆☆ | Notifications | 2 | Mở rộng |
| ★★☆☆☆ | Marketing | 2 | Mở rộng |
| ★★☆☆☆ | Reviews | 2 | Mở rộng |

**Tổng cộng: ~46 APIs**
- Ưu tiên 1 (Bắt buộc): **25 APIs**
- Ưu tiên 2 (Quan trọng): **8 APIs**
- Ưu tiên 3 (Cần thiết): **7 APIs**
- Ưu tiên 4 (Mở rộng): **6 APIs**

---

## 🚀 THỨ TỰ TRIỂN KHAI ĐỀ XUẤT

### Phase 1: MVP (2-3 tuần)
- Auth APIs (4)
- Restaurant basic (2)
- Tables CRUD (4)
- Categories CRUD (4)
- Menu CRUD (4)
- Orders CRUD (7)

### Phase 2: Business (1-2 tuần)
- Packages (3)
- Payment Settings (2)
- Statistics (3)

### Phase 3: Enhancement (1-2 tuần)
- Admin management (3)
- User profile (2)
- Table QR (2)

### Phase 4: Future
- Notifications
- Marketing
- Reviews
- Webhooks
