package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// getJWTSecret lấy secret key, đọc mỗi lần để đảm bảo .env đã được load
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-super-secret-key-change-in-production"
	}
	return []byte(secret)
}

// Claims cấu trúc JWT claims
type Claims struct {
	UserID       uint   `json:"user_id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	RestaurantID *uint  `json:"restaurant_id,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken tạo JWT token
func GenerateToken(userID uint, email, role string, restaurantID *uint) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:       userID,
		Email:        email,
		Role:         role,
		RestaurantID: restaurantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "go-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// ValidateToken xác thực JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// AuthMiddleware middleware xác thực JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Vui lòng đăng nhập",
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"details": "Missing Authorization header",
				},
			})
			c.Abort()
			return
		}

		// Kiểm tra format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token không hợp lệ",
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"details": "Invalid Authorization header format",
				},
			})
			c.Abort()
			return
		}

		claims, err := ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token hết hạn hoặc không hợp lệ",
				"error": gin.H{
					"code":    "TOKEN_EXPIRED",
					"details": err.Error(),
				},
			})
			c.Abort()
			return
		}

		// Lưu thông tin user vào context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("restaurant_id", claims.RestaurantID)
		c.Set("claims", claims)

		c.Next()
	}
}

// RoleMiddleware middleware kiểm tra role
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Vui lòng đăng nhập",
				"error": gin.H{
					"code": "UNAUTHORIZED",
				},
			})
			c.Abort()
			return
		}

		roleStr := role.(string)
		allowed := false
		for _, r := range allowedRoles {
			if r == roleStr {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Bạn không có quyền truy cập",
				"error": gin.H{
					"code":    "FORBIDDEN",
					"details": "Role not allowed",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOnly middleware chỉ cho phép admin
func AdminOnly() gin.HandlerFunc {
	return RoleMiddleware("admin")
}

// RestaurantOnly middleware chỉ cho phép restaurant
func RestaurantOnly() gin.HandlerFunc {
	return RoleMiddleware("restaurant")
}

// RestaurantOrAdmin middleware cho phép restaurant hoặc admin
func RestaurantOrAdmin() gin.HandlerFunc {
	return RoleMiddleware("restaurant", "admin")
}
