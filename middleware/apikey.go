package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// APIKeyMiddleware middleware xác thực API Key
func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		expectedAPIKey := os.Getenv("API_KEY")

		// Nếu không cấu hình API_KEY, bỏ qua kiểm tra
		if expectedAPIKey == "" {
			c.Next()
			return
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "API Key không được cung cấp",
				"error": gin.H{
					"code":    "MISSING_API_KEY",
					"details": "X-API-Key header is required",
				},
			})
			c.Abort()
			return
		}

		if apiKey != expectedAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "API Key không hợp lệ",
				"error": gin.H{
					"code":    "INVALID_API_KEY",
					"details": "The provided API Key is invalid",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAPIKeyMiddleware middleware kiểm tra API Key nếu có
func OptionalAPIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		expectedAPIKey := os.Getenv("API_KEY")

		if apiKey != "" && expectedAPIKey != "" && apiKey != expectedAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "API Key không hợp lệ",
				"error": gin.H{
					"code":    "INVALID_API_KEY",
					"details": "The provided API Key is invalid",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
