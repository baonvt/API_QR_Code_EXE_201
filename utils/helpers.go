package utils

import (
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// GenerateSlug tạo slug từ tên
func GenerateSlug(name string) string {
	// Chuyển về lowercase
	slug := strings.ToLower(name)

	// Loại bỏ dấu tiếng Việt
	slug = removeVietnameseAccents(slug)

	// Thay thế các ký tự đặc biệt bằng dấu gạch ngang
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Loại bỏ dấu gạch ngang đầu và cuối
	slug = strings.Trim(slug, "-")

	return slug
}

// removeVietnameseAccents loại bỏ dấu tiếng Việt
func removeVietnameseAccents(s string) string {
	// Mapping cho các ký tự đặc biệt tiếng Việt
	replacements := map[string]string{
		"đ": "d", "Đ": "D",
	}

	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}

	// Sử dụng unicode normalization để loại bỏ dấu
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)

	return result
}

// StringPtr trả về pointer của string
func StringPtr(s string) *string {
	return &s
}

// UintPtr trả về pointer của uint
func UintPtr(u uint) *uint {
	return &u
}

// GenerateRandomCode tạo mã ngẫu nhiên
func GenerateRandomCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())

	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

// GenerateSepayQRURL tạo URL QR code VietQR qua SePay
func GenerateSepayQRURL(bankCode, accountNumber, accountName string, amount float64, description string) string {
	// SePay QR URL format: https://qr.sepay.vn/img?bank={bank}&acc={account}&template=compact&amount={amount}&des={description}
	baseURL := "https://qr.sepay.vn/img"

	params := url.Values{}
	params.Add("bank", bankCode)
	params.Add("acc", accountNumber)
	params.Add("template", "compact")
	params.Add("amount", fmt.Sprintf("%.0f", amount))
	params.Add("des", description)

	return baseURL + "?" + params.Encode()
}
