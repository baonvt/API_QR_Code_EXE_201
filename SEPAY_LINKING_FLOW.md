# SePay Restaurant Linking Flow - Quy Trình Liên Kết SePay

## 📋 Tổng Quan

Quy trình này cho phép **chủ nhà hàng** liên kết tài khoản ngân hàng của họ với SePay để tự động nhận tiền từ khách hàng và nhận thông báo giao dịch qua webhook.

---

## 🔄 Quy Trình Chi Tiết

### STEP 1: Nhà hàng Gửi Thông Tin Ngân Hàng

**Endpoint:** `POST /api/v1/restaurants/{id}/sepay/link`

**Request:**
```json
{
  "bank_code": "MB",           // Mã ngân hàng (MB, VCB, TCB, ACB...)
  "account_number": "0393531965",  // Số tài khoản
  "account_name": "NGUYEN VAN A"   // Tên chủ tài khoản
}
```

**Response:**
```json
{
  "success": true,
  "message": "Phiên kết nối SePay đã tạo, vui lòng quét QR hoặc nhấn link để xác thực",
  "data": {
    "session_id": "sess_123456789",           // ID của phiên kết nối
    "bank_code": "MB",
    "bank_name": "MB Bank",
    "account_no": "****1965",                 // Che giấu số TK
    "account_name": "NGUYEN VAN A",
    "qr_code": "https://qr.sepay.vn/...",    // QR code để quét
    "link_url": "https://auth.sepay.vn/...", // Link để nhấp vào
    "expires_in_seconds": 300,                 // Hết hạn sau 5 phút
    "expires_at": "2026-02-02T21:51:32Z"
  }
}
```

**Quá trình:**
1. Server Golang gọi `sepayService.CreateLinkingSession()`
2. SePay API trả về `session_id` + `qr_code` + `link_url`
3. Server lưu thông tin bank vào database
4. Trả QR/Link cho nhà hàng để quét

---

### STEP 2: Chủ Nhà Hàng Xác Thực Qua App Ngân Hàng

Chủ nhà hàng có 2 lựa chọn:

**Option A - Quét QR:**
- Mở ứng dụng ngân hàng
- Chọn "Quét QR"
- Quét mã QR nhận được từ server
- Xác thực qua OTP hoặc Face ID

**Option B - Nhấn Link:**
- Nhấp vào `link_url` 
- Đăng nhập vào ứng dụng SePay
- Xác thực quyền truy cập tài khoản ngân hàng

**Kết Quả:** SePay xác nhận quyền truy cập tài khoản và lưu liên kết.

---

### STEP 3: Kiểm Tra Trạng Thái Liên Kết

**Endpoint:** `GET /api/v1/restaurants/{id}/sepay/link/check?session_id=sess_123456789`

**Response (Chưa Xác Thực):**
```json
{
  "success": true,
  "data": {
    "linked": false,
    "session_id": "sess_123456789",
    "message": "Chưa hoàn tất xác thực"
  }
}
```

**Response (Đã Xác Thực):**
```json
{
  "success": true,
  "data": {
    "linked": true,
    "session_id": "sess_123456789",
    "account_id": "acc_987654321",  // ID tài khoản từ SePay
    "bank_code": "MB",
    "account_name": "NGUYEN VAN A",
    "linked_at": "2026-02-02T21:52:00Z",
    "message": "Đã liên kết SePay thành công"
  }
}
```

**Quá trình:**
1. Frontend gọi API kiểm tra status dùng `session_id`
2. Server gọi `sepayService.GetLinkingStatus(session_id)`
3. SePay API trả về trạng thái liên kết
4. Nếu `linked: true`, server cập nhật DB:
   - `sepay_linked = true`
   - `sepay_bank_account_id = acc_987654321`
   - `sepay_linked_at = now()`

---

### STEP 4: Nhà Hàng Được Liên Kết

**Endpoint:** `GET /api/v1/restaurants/{id}/sepay/status`

**Response:**
```json
{
  "success": true,
  "data": {
    "linked": true,
    "linked_at": "2026-02-02T21:52:00Z",
    "bank_name": "MB Bank",
    "account_no": "****1965",
    "account_name": "NGUYEN VAN A",
    "accept_cash": true,
    "accept_qr": true,
    "accept_momo": false,
    "accept_vnpay": false
  }
}
```

---

## 💰 Sau Khi Liên Kết Thành Công

### Khách Hàng Thanh Toán Qua Nhà Hàng

1. **Khách tạo đơn hàng** → Server tạo payment_code (ví dụ: `ORD20260001`)

2. **Nhà hàng tạo QR thanh toán:**
   ```
   POST /api/v1/payment/orders/{id}/qr
   ```
   - Server gọi `GenerateRestaurantQR()` → Tạo QR chỉ tới **tài khoản ngân hàng của nhà hàng**
   - QR chứa nội dung: `ORD20260001` (payment code)

3. **Khách quét QR và thanh toán:**
   - Chuyển khoản đến số TK của nhà hàng
   - Nội dung: `ORD20260001`
   - Số tiền: [Tổng tiền đơn hàng]

4. **SePay Webhook Xác Nhận:**
   ```
   POST /api/v1/webhooks/sepay
   ```
   - SePay gửi webhook tự động khi có giao dịch
   - Server kiểm tra:
     - `transfer_type == "in"` (tiền vào)
     - `transaction_content` chứa `ORD20260001`
     - `transfer_amount` khớp với order total
   - **Tự động đánh dấu đơn hàng là "PAID"** ✅

---

## 🗄️ Cấu Trúc Database

### PaymentSetting Table (Cập nhật)
```sql
ALTER TABLE payment_settings ADD COLUMN (
  sepay_linked BOOLEAN DEFAULT FALSE,
  sepay_bank_account_id VARCHAR(100),
  sepay_linked_at TIMESTAMP NULL
);
```

| Column | Type | Mô Tả |
|--------|------|-------|
| `sepay_linked` | BOOLEAN | Đã liên kết SePay? |
| `sepay_bank_account_id` | VARCHAR(100) | ID tài khoản từ SePay API |
| `sepay_linked_at` | TIMESTAMP | Thời gian liên kết |

---

## 🔐 Security & Validation

### 1. **Verification Steps:**
- ✅ Bank account quét QR → Xác thực qua OTP/Face ID
- ✅ SePay API xác nhận quyền truy cập
- ✅ Server lưu `account_id` từ SePay làm chứng thực

### 2. **Account Number Masking:**
```go
maskAccountNumber("0393531965") → "****1965"
```
- Che giấu số TK trong response
- Chỉ hiển thị 4 chữ số cuối

### 3. **Webhook Signature Verification:**
- SePay webhook cần có signature xác minh
- Các giao dịch phải kiểm tra:
  - `transfer_type == "in"`
  - Tồn tại `payment_code` trong nội dung
  - Số tiền khớp với order/subscription

---

## 🔄 Webhook Flow Chi Tiết

### Khi Khách Thanh Toán:
```
Khách quét QR → Ngân hàng xác thực → SePay nhận giao dịch
→ SePay gửi Webhook tới server
→ Server kiểm tra transaction_content = "ORD20260001"
→ Tìm Order với payment_code = "ORD20260001"
→ Kiểm tra số tiền & trạng thái thanh toán
→ Cập nhật Order: payment_status = "PAID"
→ Trả response HTTP 200 OK
```

### Webhook Payload:
```json
{
  "id": 1234567890,
  "gateway": "MB",
  "transferType": "in",
  "transferAmount": 150000,
  "accountNumber": "0393531965",
  "transactionContent": "ORD20260001",
  "transactionDate": "2026-02-02T21:55:00Z",
  "referenceNumber": "123456789"
}
```

---

## 📱 API Endpoints Summary

### Restaurant SePay Linking

| Method | Endpoint | Mô Tả |
|--------|----------|-------|
| POST | `/api/v1/restaurants/{id}/sepay/link` | Tạo phiên kết nối SePay |
| GET | `/api/v1/restaurants/{id}/sepay/link/check` | Kiểm tra trạng thái liên kết |
| GET | `/api/v1/restaurants/{id}/sepay/status` | Lấy trạng thái SePay của nhà hàng |
| DELETE | `/api/v1/restaurants/{id}/sepay/unlink` | Hủy liên kết SePay |

### Payment

| Method | Endpoint | Mô Tả |
|--------|----------|-------|
| POST | `/api/v1/payment/orders/{id}/qr` | Tạo QR thanh toán đơn hàng |
| GET | `/api/v1/payment/orders/{id}/status` | Kiểm tra trạng thái thanh toán |
| POST | `/api/v1/webhooks/sepay` | Webhook SePay (Callback) |

---

## ⚙️ Environment Variables

```env
SEPAY_API_KEY=your_api_key
SEPAY_API_TOKEN=your_api_token
SEPAY_BANK_CODE=MB
SEPAY_ACCOUNT_NUMBER=0393531965
SEPAY_ACCOUNT_NAME=DUONG MANH HUY
WEBHOOK_URL=https://apiqrcodeexe201-production.up.railway.app/api/v1/webhooks/sepay
```

---

## 🧪 Testing Flow

### 1. Test Linking:
```bash
curl -X POST http://localhost:8080/api/v1/restaurants/1/sepay/link \
  -H "Authorization: Bearer {jwt_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "bank_code": "MB",
    "account_number": "0393531965",
    "account_name": "NGUYEN VAN A"
  }'
```

### 2. Check Linking Status:
```bash
curl -X GET "http://localhost:8080/api/v1/restaurants/1/sepay/link/check?session_id=sess_123456789" \
  -H "Authorization: Bearer {jwt_token}"
```

### 3. Create Order Payment QR:
```bash
curl -X POST http://localhost:8080/api/v1/payment/orders/1/qr \
  -H "Authorization: Bearer {jwt_token}"
```

### 4. Webhook Test (Simulate Payment):
```bash
curl -X POST http://localhost:8080/api/v1/webhooks/sepay \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1234567890,
    "gateway": "MB",
    "transferType": "in",
    "transferAmount": 150000,
    "accountNumber": "0393531965",
    "transactionContent": "ORD20260001",
    "transactionDate": "2026-02-02T21:55:00Z",
    "referenceNumber": "123456789"
  }'
```

---

## ✨ Key Features

✅ **Quy trình xác thực an toàn:** Chủ shop xác thực qua ngân hàng  
✅ **Tự động thanh toán:** SePay webhook tự động xác nhận giao dịch  
✅ **Quản lý quyền:** Chỉ chủ shop có quyền liên kết/hủy liên kết  
✅ **Che giấu thông tin:** Số TK che giấu trong API response  
✅ **Tracking giao dịch:** Lưu tất cả webhook vào database  
✅ **Hỗ trợ Restful:** RESTful API dễ tích hợp  

---

## 🚀 Deployment

Sau khi test thành công:

1. **Push code lên Railway:**
   ```bash
   git add .
   git commit -m "SePay linking integration"
   git push origin main
   ```

2. **Cập nhật Railway env vars:**
   - `SEPAY_API_KEY`
   - `SEPAY_API_TOKEN`
   - `WEBHOOK_URL` (Railway public URL)

3. **Cấu hình SePay Dashboard:**
   - Đăng nhập SePay Admin
   - Webhook Settings → Cập nhật URL: `https://your-domain.com/api/v1/webhooks/sepay`

4. **Test với giao dịch thực:**
   - Chủ shop liên kết STK thực
   - Khách thanh toán
   - Kiểm tra webhook log

---

## 📚 Architecture Diagram

```
┌─────────────────────┐
│   Restaurant App    │
└──────────┬──────────┘
           │ 1. POST /sepay/link
           │    (bank_code, account_number, account_name)
           ▼
┌─────────────────────┐
│   Go API Server     │
│                     │
│ ┌─────────────────┐ │
│ │ sepayService    │ │
│ └────────┬────────┘ │
└──────────┼──────────┘
           │ 2. CreateLinkingSession()
           ▼
┌─────────────────────┐
│   SePay API         │
│ CreateLinking       │
└──────────┬──────────┘
           │ 3. Return: session_id, qr_code, link_url
           ▼
┌─────────────────────┐
│   Restaurant App    │
│  Display QR/Link    │
│  for scan+verify    │
└─────────────────────┘
           │ 4. Scan QR / Click Link
           ▼
┌─────────────────────┐
│   Bank App          │
│   Verify + Confirm  │
└─────────────────────┘
           │ 5. SePay Link Success
           ▼
┌─────────────────────┐
│   Go API Server     │
│ GetLinkingStatus()  │
│ Update DB           │
└─────────────────────┘

[Later when customer pays]

┌──────────────────────┐
│  Customer Bank App   │
│  Transfer Money      │
└──────────┬───────────┘
           │ QR Content: ORD20260001
           ▼
┌──────────────────────┐
│   SePay Webhook      │
│   Sends to Server    │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│   Go API Server      │
│ - Verify signature   │
│ - Check payment_code │
│ - Update Order: PAID │
└──────────────────────┘
```

---

## 🔧 Troubleshooting

| Vấn Đề | Nguyên Nhân | Giải Pháp |
|--------|-----------|----------|
| "Invalid session_id" | SePay session hết hạn | Tạo phiên mới (5 phút timeout) |
| Webhook không nhận | URL sai hoặc firewall | Cập nhật webhook URL trên SePay |
| Không tìm order | Payment code sai format | Kiểm tra format `ORD{id}` |
| Không link được | SePay API token sai | Verify token từ SePay Dashboard |

