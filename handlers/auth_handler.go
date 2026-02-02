package handlers

import (
	"fmt"
	"net/http"
	"time"

	"go-api/config"
	"go-api/middleware"
	"go-api/models"
	"go-api/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// ===============================
// REQUEST STRUCTS
// ===============================

// LoginInput request body cho login
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterInput request body cho register
type RegisterInput struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	Name           string `json:"name" binding:"required"`
	Phone          string `json:"phone"`
	RestaurantName string `json:"restaurant_name" binding:"required"`
	PackageID      uint   `json:"package_id" binding:"required"`
}

// ===============================
// HANDLERS
// ===============================

// Login đăng nhập
// @Summary Đăng nhập
// @Description Đăng nhập bằng email và password để lấy JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body LoginInput true "Thông tin đăng nhập"
// @Success 200 {object} map[string]interface{}
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	// Tìm user theo email
	var user models.User
	if err := config.GetDB().Where("email = ?", input.Email).First(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Email hoặc mật khẩu không đúng", "INVALID_CREDENTIALS", "")
		return
	}

	// Kiểm tra user có active không
	if !user.IsActive {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Tài khoản đã bị vô hiệu hóa", "ACCOUNT_DISABLED", "")
		return
	}

	// Kiểm tra password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Email hoặc mật khẩu không đúng", "INVALID_CREDENTIALS", "")
		return
	}

	// Lấy restaurant_id nếu là restaurant
	var restaurantID *uint
	if user.Role == "restaurant" {
		var restaurant models.Restaurant
		if err := config.GetDB().Where("owner_id = ?", user.ID).First(&restaurant).Error; err == nil {
			restaurantID = &restaurant.ID
		}
	}

	// Tạo JWT token
	token, err := middleware.GenerateToken(user.ID, user.Email, user.Role, restaurantID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo token", "TOKEN_ERROR", err.Error())
		return
	}

	// Cập nhật last_login
	now := time.Now()
	config.GetDB().Model(&user).Update("last_login", now)

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"user": gin.H{
			"id":     user.ID,
			"email":  user.Email,
			"name":   user.Name,
			"role":   user.Role,
			"avatar": user.Avatar,
		},
		"restaurant_id": restaurantID,
		"access_token":  token,
		"expires_in":    86400,
	}, "Đăng nhập thành công")
}

// Register đăng ký nhà hàng mới
// @Summary Đăng ký nhà hàng
// @Description Đăng ký tài khoản và tạo nhà hàng mới
// @Tags Auth
// @Accept json
// @Produce json
// @Param register body RegisterInput true "Thông tin đăng ký"
// @Success 201 {object} map[string]interface{}
// @Router /auth/register [post]
func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	db := config.GetDB()

	// Kiểm tra email đã tồn tại chưa
	var existingUser models.User
	if err := db.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Email đã được sử dụng", "EMAIL_EXISTS", "")
		return
	}

	// Kiểm tra package có tồn tại không
	var pkg models.Package
	if err := db.First(&pkg, input.PackageID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Gói dịch vụ không tồn tại", "INVALID_PACKAGE", "")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Lỗi hệ thống", "HASH_ERROR", err.Error())
		return
	}

	// Bắt đầu transaction
	tx := db.Begin()

	// Tạo user
	user := models.User{
		Email:    input.Email,
		Password: string(hashedPassword),
		Name:     input.Name,
		Role:     "restaurant",
		Phone:    &input.Phone,
		IsActive: true,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo tài khoản", "CREATE_USER_ERROR", err.Error())
		return
	}

	// Tạo slug cho restaurant
	slug := utils.GenerateSlug(input.RestaurantName)
	// Kiểm tra slug trùng và thêm số nếu cần
	var count int64
	tx.Model(&models.Restaurant{}).Where("slug LIKE ?", slug+"%").Count(&count)
	if count > 0 {
		slug = fmt.Sprintf("%s-%d", slug, count+1)
	}

	// Tạo restaurant
	restaurant := models.Restaurant{
		OwnerID:          user.ID,
		PackageID:        input.PackageID,
		Name:             input.RestaurantName,
		Slug:             slug,
		IsOpen:           true,
		TaxRate:          10.0,
		ServiceCharge:    5.0,
		Currency:         "VND",
		PackageStartDate: time.Now(),
		PackageEndDate:   time.Now().AddDate(0, 1, 0), // 1 tháng trial
		PackageStatus:    "active",
		Status:           "active",
	}

	if err := tx.Create(&restaurant).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo nhà hàng", "CREATE_RESTAURANT_ERROR", err.Error())
		return
	}

	// Commit transaction
	tx.Commit()

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
		"restaurant": gin.H{
			"id":   restaurant.ID,
			"name": restaurant.Name,
			"slug": restaurant.Slug,
		},
	}, "Đăng ký thành công")
}

// Logout đăng xuất
// @Summary Đăng xuất
// @Description Đăng xuất khỏi hệ thống
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /auth/logout [post]
func Logout(c *gin.Context) {
	// Với JWT stateless, logout chỉ cần xóa token ở client
	// Có thể implement blacklist token nếu cần
	utils.SuccessResponse(c, http.StatusOK, nil, "Đăng xuất thành công")
}

// GetMe lấy thông tin user hiện tại
// @Summary Lấy thông tin user
// @Description Lấy thông tin user đang đăng nhập
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /auth/me [get]
func GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := config.GetDB().First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy người dùng", "USER_NOT_FOUND", "")
		return
	}

	response := gin.H{
		"id":     user.ID,
		"email":  user.Email,
		"name":   user.Name,
		"role":   user.Role,
		"avatar": user.Avatar,
		"phone":  user.Phone,
	}

	// Nếu là restaurant, thêm thông tin nhà hàng
	if user.Role == "restaurant" {
		var restaurant models.Restaurant
		if err := config.GetDB().Where("owner_id = ?", user.ID).First(&restaurant).Error; err == nil {
			response["restaurant"] = gin.H{
				"id":   restaurant.ID,
				"name": restaurant.Name,
				"slug": restaurant.Slug,
			}
		}
	}

	utils.SuccessResponse(c, http.StatusOK, response, "")
}

// CheckEmailInput request body cho check email
type CheckEmailInput struct {
	Email string `json:"email" binding:"required,email"`
}

// CheckEmail kiểm tra email đã tồn tại chưa
// @Summary Kiểm tra email
// @Description Kiểm tra email đã được đăng ký chưa
// @Tags Auth
// @Accept json
// @Produce json
// @Param check body CheckEmailInput true "Email cần kiểm tra"
// @Success 200 {object} map[string]interface{}
// @Router /auth/check-email [post]
func CheckEmail(c *gin.Context) {
	var input CheckEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Email không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	db := config.GetDB()

	// Kiểm tra email đã tồn tại trong users chưa
	var existingUser models.User
	if err := db.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Email đã được sử dụng", "EMAIL_EXISTS", "")
		return
	}

	// Kiểm tra email có pending subscription không (đã đăng ký nhưng chưa thanh toán)
	var existingSub models.PackageSubscription
	if err := db.Where("email = ? AND payment_status = ?", input.Email, "pending").First(&existingSub).Error; err == nil {
		// Nếu chưa hết hạn, vẫn cho phép (user có thể quay lại thanh toán)
		// Chỉ báo lỗi nếu đã được sử dụng bởi user thực
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"available": true,
		"email":     input.Email,
	}, "Email có thể sử dụng")
}

// RefreshToken làm mới token
// @Summary Làm mới token
// @Description Lấy token mới từ token hiện tại
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /auth/refresh [post]
func RefreshToken(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Token không hợp lệ", "INVALID_TOKEN", "")
		return
	}

	claimsData := claims.(*middleware.Claims)

	// Tạo token mới
	token, err := middleware.GenerateToken(claimsData.UserID, claimsData.Email, claimsData.Role, claimsData.RestaurantID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo token", "TOKEN_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"access_token": token,
		"expires_in":   86400,
	}, "Token đã được làm mới")
}
