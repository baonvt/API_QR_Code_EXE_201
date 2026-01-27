package utils

import "github.com/gin-gonic/gin"

// Response chuẩn API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo thông tin lỗi
type ErrorInfo struct {
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// Pagination thông tin phân trang
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// SuccessResponse trả về response thành công
func SuccessResponse(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// ErrorResponse trả về response lỗi
func ErrorResponse(c *gin.Context, statusCode int, message string, code string, details string) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    code,
			Details: details,
		},
	})
}

// PaginatedResponse trả về response có phân trang
func PaginatedResponse(c *gin.Context, statusCode int, data interface{}, pagination Pagination, message string) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"data":    data,
		"pagination": gin.H{
			"page":        pagination.Page,
			"limit":       pagination.Limit,
			"total":       pagination.Total,
			"total_pages": pagination.TotalPages,
		},
		"message": message,
	})
}
