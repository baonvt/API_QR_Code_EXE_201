package services

import (
	"fmt"
	"net/url"
	"strings"
)

// VietQR Service - Generate VietQR payment QR code
// Documentation: https://vietqr.io/

// BankBinMap - Mapping bank code to bank bin
var BankBinMap = map[string]string{
	"VCB":        "970436", // Vietcombank
	"TCB":        "970407", // Techcombank
	"MB":         "970422", // MB Bank
	"ACB":        "970416", // ACB
	"VTB":        "970415", // Vietinbank
	"BIDV":       "970418", // BIDV
	"AGR":        "970405", // Agribank
	"OCB":        "970448", // OCB
	"TPB":        "970423", // TPBank
	"SCB":        "970429", // Sacombank
	"VPB":        "970432", // VPBank
	"SHB":        "970443", // SHB
	"EIB":        "970431", // Eximbank
	"MSB":        "970426", // MSB
	"CAKE":       "546034", // CAKE
	"UBANK":      "546035", // Ubank
	"TIMO":       "963388", // Timo
	"VIB":        "970441", // VIB
	"SGB":        "970400", // Saigonbank
	"BVB":        "970438", // BaoVietBank
	"SEAB":       "970440", // SeABank
	"COOPBANK":   "970446", // Co-op Bank
	"LPB":        "970449", // LienVietPostBank
	"KLB":        "970452", // KienLongBank
	"KBank":      "668888", // KBank (Kasikornbank)
	"KEBHANAHCM": "970463", // KEB Hana
	"KEBHANASG":  "970466", // KEB Hana SG
	"MAFC":       "977777", // MAFC
	"CITIBANK":   "533948", // Citibank
	"KBHCM":      "970462", // KB Bank
	"VBSP":       "999888", // Ngân hàng Chính sách Xã hội
	"WVN":        "970457", // Woori
	"VRB":        "970421", // VRB
	"UOB":        "970458", // UOB
	"SCVN":       "970410", // Standard Chartered
	"PBVN":       "970439", // PublicBank
	"NHB":        "970419", // NCB
	"IVB":        "970434", // IndovinaBank
	"IBK":        "970456", // IBK
	"HSBC":       "458761", // HSBC
	"HLBVN":      "970442", // HongLeong Bank
	"GPB":        "970408", // GPBank
	"DOB":        "970406", // DongA Bank
	"DBS":        "796500", // DBS Bank
	"CIMB":       "422589", // CIMB
	"CBB":        "970444", // CBBank
	"ABBANK":     "970425", // ABBANK
	"SHBVN":      "970443", // ShinhanBank
	"VAB":        "970427", // VietA Bank
	"NAB":        "970428", // NamA Bank
	"BAB":        "970409", // BacA Bank
	"PGB":        "970430", // PGBank
	"SGICB":      "970400", // Saigon Bank
}

// GenerateVietQRURL generates VietQR payment URL
func GenerateVietQRURL(bankCode, accountNumber, accountName string, amount float64, description string) (string, error) {
	// Get bank bin from bank code
	bankBin, ok := BankBinMap[strings.ToUpper(bankCode)]
	if !ok {
		return "", fmt.Errorf("unsupported bank code: %s", bankCode)
	}

	// Validate account number
	if accountNumber == "" {
		return "", fmt.Errorf("account number is required")
	}

	// Clean account name (remove special characters)
	cleanAccountName := strings.ToUpper(accountName)
	cleanAccountName = strings.ReplaceAll(cleanAccountName, "Đ", "D")
	cleanAccountName = url.QueryEscape(cleanAccountName)

	// Clean description
	cleanDescription := url.QueryEscape(description)

	// Build VietQR URL
	// Format: https://img.vietqr.io/image/{BANK_BIN}-{ACCOUNT_NUMBER}-{TEMPLATE}.png?amount={AMOUNT}&addInfo={DESCRIPTION}&accountName={ACCOUNT_NAME}
	template := "compact2" // compact, compact2, qr_only, print

	qrURL := fmt.Sprintf(
		"https://img.vietqr.io/image/%s-%s-%s.png?amount=%.0f&addInfo=%s&accountName=%s",
		bankBin,
		accountNumber,
		template,
		amount,
		cleanDescription,
		cleanAccountName,
	)

	return qrURL, nil
}

// GetBankBin returns bank bin from bank code
func GetBankBin(bankCode string) (string, bool) {
	bin, ok := BankBinMap[strings.ToUpper(bankCode)]
	return bin, ok
}
