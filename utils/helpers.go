package utils

import (
	"regexp"
	"strings"
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
