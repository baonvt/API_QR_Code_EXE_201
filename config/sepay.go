package config

import (
	"log"
	"os"
)

// SepayConfig chứa cấu hình SePay
type SepayConfig struct {
	APIKey        string // API Key để xác thực
	APIToken      string // API Token để tra cứu giao dịch
	BankCode      string // Mã ngân hàng: MB, VCB, TCB, ACB...
	AccountNumber string // Số tài khoản
	AccountName   string // Tên tài khoản
	WebhookURL    string // URL webhook
}

var sepayConfig *SepayConfig

// GetSepayConfig trả về cấu hình SePay
func GetSepayConfig() *SepayConfig {
	if sepayConfig == nil {
		LoadSepayConfig()
	}
	return sepayConfig
}

// LoadSepayConfig load cấu hình SePay từ environment
func LoadSepayConfig() {
	sepayConfig = &SepayConfig{
		APIKey:        os.Getenv("SEPAY_API_KEY"),
		APIToken:      os.Getenv("SEPAY_API_TOKEN"),
		BankCode:      os.Getenv("SEPAY_BANK_CODE"),
		AccountNumber: os.Getenv("SEPAY_ACCOUNT_NUMBER"),
		AccountName:   os.Getenv("SEPAY_ACCOUNT_NAME"),
		WebhookURL:    os.Getenv("SEPAY_WEBHOOK_URL"),
	}

	// Log cấu hình (ẩn sensitive data)
	if sepayConfig.APIKey != "" {
		log.Printf("✅ SePay configured: Bank=%s, Account=****%s",
			sepayConfig.BankCode,
			getLastChars(sepayConfig.AccountNumber, 4))
	} else {
		log.Println("⚠️ SePay not configured (SEPAY_API_KEY not set)")
	}
}

// IsSepayConfigured kiểm tra SePay đã được cấu hình chưa
func IsSepayConfigured() bool {
	cfg := GetSepayConfig()
	return cfg.APIKey != "" && cfg.BankCode != "" && cfg.AccountNumber != ""
}

// getLastChars lấy n ký tự cuối của string
func getLastChars(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

// BankCodeToName chuyển mã ngân hàng sang tên đầy đủ
func BankCodeToName(code string) string {
	banks := map[string]string{
		"MB":       "MB Bank",
		"VCB":      "Vietcombank",
		"TCB":      "Techcombank",
		"ACB":      "ACB",
		"TPB":      "TPBank",
		"VPB":      "VPBank",
		"BIDV":     "BIDV",
		"VTB":      "VietinBank",
		"STB":      "Sacombank",
		"HDB":      "HDBank",
		"MSB":      "MSB",
		"SHB":      "SHB",
		"EIB":      "Eximbank",
		"OCB":      "OCB",
		"NAB":      "Nam A Bank",
		"CAKE":     "CAKE by VPBank",
		"UBANK":    "Ubank by VPBank",
		"TIMO":     "Timo by Ban Viet",
		"VIETBANK": "Viet Bank",
	}

	if name, ok := banks[code]; ok {
		return name
	}
	return code
}
