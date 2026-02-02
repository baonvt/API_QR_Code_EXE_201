package services

import (
	"fmt"
	"net/url"
	"strings"

	"go-api/config"
)

// VietQR chuẩn QR code cho ngân hàng Việt Nam
// Sử dụng API của VietQR.io

// QRCodeInfo chứa thông tin tạo QR
type QRCodeInfo struct {
	BankCode      string  // Mã ngân hàng
	AccountNumber string  // Số tài khoản
	AccountName   string  // Tên tài khoản
	Amount        float64 // Số tiền
	Description   string  // Nội dung chuyển khoản
	Template      string  // Template: compact, compact2, qr_only, print
}

// QRCodeResult kết quả tạo QR
type QRCodeResult struct {
	QRURL       string `json:"qr_url"`       // URL ảnh QR
	QRContent   string `json:"qr_content"`   // Nội dung QR (cho client tự render)
	BankName    string `json:"bank_name"`    // Tên ngân hàng
	AccountNo   string `json:"account_no"`   // Số TK
	AccountName string `json:"account_name"` // Tên TK
	Amount      int64  `json:"amount"`       // Số tiền
	Description string `json:"description"`  // Nội dung
}

// GenerateVietQR tạo QR code thanh toán theo chuẩn VietQR
func GenerateVietQR(info QRCodeInfo) *QRCodeResult {
	// Sử dụng VietQR.io API (miễn phí)
	// Format: https://img.vietqr.io/image/{bankCode}-{accountNo}-{template}.png?amount={amount}&addInfo={description}&accountName={name}

	if info.Template == "" {
		info.Template = "compact2" // Template mặc định
	}

	// Build URL
	baseURL := fmt.Sprintf("https://img.vietqr.io/image/%s-%s-%s.png",
		info.BankCode,
		info.AccountNumber,
		info.Template,
	)

	// Add query params
	params := url.Values{}
	if info.Amount > 0 {
		params.Add("amount", fmt.Sprintf("%.0f", info.Amount))
	}
	if info.Description != "" {
		params.Add("addInfo", info.Description)
	}
	if info.AccountName != "" {
		params.Add("accountName", info.AccountName)
	}

	qrURL := baseURL
	if len(params) > 0 {
		qrURL = baseURL + "?" + params.Encode()
	}

	return &QRCodeResult{
		QRURL:       qrURL,
		QRContent:   info.Description,
		BankName:    config.BankCodeToName(info.BankCode),
		AccountNo:   info.AccountNumber,
		AccountName: info.AccountName,
		Amount:      int64(info.Amount),
		Description: info.Description,
	}
}

// GenerateAdminQR tạo QR thanh toán về tài khoản Admin (đăng ký gói)
func GenerateAdminQR(amount float64, description string) *QRCodeResult {
	cfg := config.GetSepayConfig()

	return GenerateVietQR(QRCodeInfo{
		BankCode:      cfg.BankCode,
		AccountNumber: cfg.AccountNumber,
		AccountName:   cfg.AccountName,
		Amount:        amount,
		Description:   description,
		Template:      "compact2",
	})
}

// GenerateRestaurantQR tạo QR thanh toán về tài khoản nhà hàng
func GenerateRestaurantQR(bankCode, accountNumber, accountName string, amount float64, description string) *QRCodeResult {
	return GenerateVietQR(QRCodeInfo{
		BankCode:      bankCode,
		AccountNumber: accountNumber,
		AccountName:   accountName,
		Amount:        amount,
		Description:   description,
		Template:      "compact2",
	})
}

// GeneratePaymentCode tạo mã thanh toán unique
// Format cho đăng ký gói: PKG{subscriptionID}{timestamp}
// Format cho đơn hàng: ORD{orderNumber} (không có dấu -)
func GeneratePackagePaymentCode(subscriptionID uint) string {
	// Format: PKG + ID + random suffix
	// Ví dụ: PKG123 (ngắn gọn để user dễ nhập)
	return fmt.Sprintf("PKG%d", subscriptionID)
}

// GenerateOrderPaymentCode tạo mã thanh toán cho đơn hàng
// Input: ORD-2026-0015 -> Output: ORD20260015
func GenerateOrderPaymentCode(orderNumber string) string {
	// Bỏ dấu -
	code := strings.ReplaceAll(orderNumber, "-", "")
	return code
}

// ParsePaymentCode phân tích mã thanh toán từ nội dung chuyển khoản
// Trả về loại (package/order), ID/code, và amount (nếu có)
func ParsePaymentCode(content string) (transactionType string, code string, found bool) {
	// Chuẩn hóa: uppercase, bỏ khoảng trắng thừa
	content = strings.ToUpper(strings.TrimSpace(content))

	// Tìm pattern PKG hoặc ORD
	if idx := strings.Index(content, "PKG"); idx != -1 {
		// Lấy phần sau PKG
		rest := content[idx+3:]
		// Lấy số liên tiếp
		code = extractNumbers(rest)
		if code != "" {
			return "package", "PKG" + code, true
		}
	}

	if idx := strings.Index(content, "ORD"); idx != -1 {
		// Lấy phần sau ORD
		rest := content[idx+3:]
		// Lấy số liên tiếp
		code = extractAlphanumeric(rest)
		if code != "" {
			return "order", "ORD" + code, true
		}
	}

	return "", "", false
}

// extractNumbers lấy các ký tự số liên tiếp từ đầu string
func extractNumbers(s string) string {
	var result strings.Builder
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result.WriteRune(c)
		} else if result.Len() > 0 {
			// Dừng khi gặp ký tự không phải số sau khi đã có số
			break
		}
	}
	return result.String()
}

// extractAlphanumeric lấy các ký tự số liên tiếp (cho order code)
func extractAlphanumeric(s string) string {
	var result strings.Builder
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result.WriteRune(c)
		} else if result.Len() > 0 {
			break
		}
	}
	return result.String()
}
