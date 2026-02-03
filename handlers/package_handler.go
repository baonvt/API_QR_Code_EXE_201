package handlers

import (
	"net/http"
	"strconv"
	"time"

	"go-api/config"
	"go-api/models"
	"go-api/utils"

	"github.com/gin-gonic/gin"
)

// ===============================
// PACKAGES HANDLERS
// ===============================

// GetPackages lấy danh sách gói dịch vụ (Public)
// @Summary Lấy danh sách gói dịch vụ
// @Description Lấy tất cả gói dịch vụ đang active
// @Tags Packages
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /packages [get]
func GetPackages(c *gin.Context) {
	var packages []models.Package

	if err := config.GetDB().Where("is_active = ?", true).Order("sort_order ASC").Find(&packages).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Lỗi khi lấy danh sách gói", "QUERY_ERROR", err.Error())
		return
	}

	var data []gin.H
	for _, pkg := range packages {
		data = append(data, gin.H{
			"id":             pkg.ID,
			"name":           pkg.Name,
			"display_name":   pkg.DisplayName,
			"description":    pkg.Description,
			"monthly_price":  pkg.MonthlyPrice,
			"yearly_price":   pkg.YearlyPrice,
			"max_menu_items": pkg.MaxMenuItems,
			"max_tables":     pkg.MaxTables,
			"max_categories": pkg.MaxCategories,
			"features":       pkg.Features,
			"is_popular":     pkg.IsPopular,
		})
	}

	utils.SuccessResponse(c, http.StatusOK, data, "")
}

// CreatePackage tạo gói mới (Admin)
// @Summary Tạo gói dịch vụ
// @Description Admin tạo gói dịch vụ mới
// @Tags Admin
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{}
// @Security BearerAuth
// @Router /admin/packages [post]
func CreatePackage(c *gin.Context) {
	var input struct {
		Name          string  `json:"name" binding:"required"`
		DisplayName   string  `json:"display_name" binding:"required"`
		Description   string  `json:"description"`
		MonthlyPrice  float64 `json:"monthly_price" binding:"required"`
		YearlyPrice   float64 `json:"yearly_price" binding:"required"`
		MaxMenuItems  int     `json:"max_menu_items"`
		MaxTables     int     `json:"max_tables"`
		MaxCategories int     `json:"max_categories"`
		Features      string  `json:"features"`
		IsPopular     bool    `json:"is_popular"`
		SortOrder     int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	pkg := models.Package{
		Name:          input.Name,
		DisplayName:   input.DisplayName,
		Description:   &input.Description,
		MonthlyPrice:  input.MonthlyPrice,
		YearlyPrice:   input.YearlyPrice,
		MaxMenuItems:  input.MaxMenuItems,
		MaxTables:     input.MaxTables,
		MaxCategories: input.MaxCategories,
		Features:      &input.Features,
		IsPopular:     input.IsPopular,
		IsActive:      true,
		SortOrder:     input.SortOrder,
	}

	if err := config.GetDB().Create(&pkg).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo gói", "CREATE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"id":           pkg.ID,
		"name":         pkg.Name,
		"display_name": pkg.DisplayName,
	}, "Tạo gói thành công")
}

// UpdatePackage cập nhật gói (Admin)
// @Summary Cập nhật gói dịch vụ
// @Description Admin cập nhật gói dịch vụ
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "Package ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /admin/packages/{id} [put]
func UpdatePackage(c *gin.Context) {
	packageID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var pkg models.Package
	if err := config.GetDB().First(&pkg, packageID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy gói", "PACKAGE_NOT_FOUND", "")
		return
	}

	var input struct {
		DisplayName   string  `json:"display_name"`
		Description   string  `json:"description"`
		MonthlyPrice  float64 `json:"monthly_price"`
		YearlyPrice   float64 `json:"yearly_price"`
		MaxMenuItems  int     `json:"max_menu_items"`
		MaxTables     int     `json:"max_tables"`
		MaxCategories int     `json:"max_categories"`
		Features      string  `json:"features"`
		IsPopular     *bool   `json:"is_popular"`
		IsActive      *bool   `json:"is_active"`
		SortOrder     int     `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	updates := make(map[string]interface{})
	if input.DisplayName != "" {
		updates["display_name"] = input.DisplayName
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if input.MonthlyPrice > 0 {
		updates["monthly_price"] = input.MonthlyPrice
	}
	if input.YearlyPrice > 0 {
		updates["yearly_price"] = input.YearlyPrice
	}
	if input.MaxMenuItems != 0 {
		updates["max_menu_items"] = input.MaxMenuItems
	}
	if input.MaxTables != 0 {
		updates["max_tables"] = input.MaxTables
	}
	if input.MaxCategories != 0 {
		updates["max_categories"] = input.MaxCategories
	}
	if input.Features != "" {
		updates["features"] = input.Features
	}
	if input.IsPopular != nil {
		updates["is_popular"] = *input.IsPopular
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if input.SortOrder > 0 {
		updates["sort_order"] = input.SortOrder
	}

	if err := config.GetDB().Model(&pkg).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể cập nhật gói", "UPDATE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":           pkg.ID,
		"name":         pkg.Name,
		"display_name": pkg.DisplayName,
	}, "Cập nhật gói thành công")
}

// ReseedPackages xóa và tạo lại tất cả packages (Admin)
// @Summary Reseed packages
// @Description Admin xóa tất cả packages cũ và tạo lại packages mới
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /admin/packages/reseed [post]
func ReseedPackages(c *gin.Context) {
	if err := config.ReseedPackages(); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể reseed packages", "RESEED_ERROR", err.Error())
		return
	}

	// Lấy danh sách packages mới
	var packages []models.Package
	config.GetDB().Where("is_active = ?", true).Order("sort_order ASC").Find(&packages)

	var data []gin.H
	for _, pkg := range packages {
		data = append(data, gin.H{
			"id":             pkg.ID,
			"name":           pkg.Name,
			"display_name":   pkg.DisplayName,
			"monthly_price":  pkg.MonthlyPrice,
			"yearly_price":   pkg.YearlyPrice,
			"max_menu_items": pkg.MaxMenuItems,
			"max_tables":     pkg.MaxTables,
			"max_categories": pkg.MaxCategories,
		})
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message":  "Đã reseed packages thành công",
		"packages": data,
	}, "Reseed packages thành công")
}

// ===============================
// PAYMENT SETTINGS HANDLERS
// ===============================

// GetPaymentSettings lấy cài đặt thanh toán
// @Summary Lấy cài đặt thanh toán
// @Description Lấy cài đặt phương thức thanh toán của nhà hàng
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/payment-settings [get]
func GetPaymentSettings(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || uint(restaurantID) != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền xem cài đặt thanh toán này", "FORBIDDEN", "")
		return
	}

	var settings models.PaymentSetting
	if err := config.GetDB().Where("restaurant_id = ?", restaurantID).First(&settings).Error; err != nil {
		// Tạo mới nếu chưa có
		settings = models.PaymentSetting{
			RestaurantID: uint(restaurantID),
			AcceptCash:   true,
		}
		config.GetDB().Create(&settings)
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":             settings.ID,
		"bank_name":      settings.BankName,
		"account_number": settings.AccountNumber,
		"account_name":   settings.AccountName,
		"qr_image":       settings.QRImage,
		"accept_cash":    settings.AcceptCash,
		"accept_qr":      settings.AcceptQR,
		"accept_momo":    settings.AcceptMomo,
		"accept_vnpay":   settings.AcceptVNPay,
	}, "")
}

// UpdatePaymentSettings cập nhật cài đặt thanh toán
// @Summary Cập nhật cài đặt thanh toán
// @Description Cập nhật cài đặt phương thức thanh toán
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/payment-settings [put]
func UpdatePaymentSettings(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || uint(restaurantID) != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền cập nhật cài đặt thanh toán này", "FORBIDDEN", "")
		return
	}

	var input struct {
		BankName      string `json:"bank_name"`
		AccountNumber string `json:"account_number"`
		AccountName   string `json:"account_name"`
		QRImage       string `json:"qr_image"`
		AcceptCash    *bool  `json:"accept_cash"`
		AcceptQR      *bool  `json:"accept_qr"`
		AcceptMomo    *bool  `json:"accept_momo"`
		AcceptVNPay   *bool  `json:"accept_vnpay"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	var settings models.PaymentSetting
	if err := config.GetDB().Where("restaurant_id = ?", restaurantID).First(&settings).Error; err != nil {
		// Tạo mới
		settings = models.PaymentSetting{
			RestaurantID: uint(restaurantID),
		}
		config.GetDB().Create(&settings)
	}

	updates := make(map[string]interface{})
	if input.BankName != "" {
		updates["bank_name"] = input.BankName
	}
	if input.AccountNumber != "" {
		updates["account_number"] = input.AccountNumber
	}
	if input.AccountName != "" {
		updates["account_name"] = input.AccountName
	}
	if input.QRImage != "" {
		updates["qr_image"] = input.QRImage
	}
	if input.AcceptCash != nil {
		updates["accept_cash"] = *input.AcceptCash
	}
	if input.AcceptQR != nil {
		updates["accept_qr"] = *input.AcceptQR
	}
	if input.AcceptMomo != nil {
		updates["accept_momo"] = *input.AcceptMomo
	}
	if input.AcceptVNPay != nil {
		updates["accept_vnpay"] = *input.AcceptVNPay
	}

	if err := config.GetDB().Model(&settings).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể cập nhật cài đặt", "UPDATE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Cập nhật cài đặt thanh toán thành công",
	}, "Cập nhật thành công")
}

// ===============================
// ADMIN STATS HANDLERS
// ===============================

// GetAdminStats thống kê toàn hệ thống (Admin)
// GET /admin/stats
func GetAdminStats(c *gin.Context) {
	db := config.GetDB()

	// Tổng số nhà hàng
	var totalRestaurants int64
	db.Model(&models.Restaurant{}).Count(&totalRestaurants)

	// Nhà hàng active
	var activeRestaurants int64
	db.Model(&models.Restaurant{}).Where("status = ?", "active").Count(&activeRestaurants)

	// Nhà hàng mới trong tháng
	var newThisMonth int64
	db.Model(&models.Restaurant{}).Where("EXTRACT(MONTH FROM created_at) = EXTRACT(MONTH FROM NOW()) AND EXTRACT(YEAR FROM created_at) = EXTRACT(YEAR FROM NOW())").Count(&newThisMonth)

	// Thống kê theo gói
	var packageStats []struct {
		PackageName string
		Count       int64
	}
	db.Model(&models.Restaurant{}).
		Select("packages.name as package_name, count(*) as count").
		Joins("JOIN packages ON packages.id = restaurants.package_id").
		Group("packages.name").
		Scan(&packageStats)

	var byPackage []gin.H
	for _, ps := range packageStats {
		byPackage = append(byPackage, gin.H{
			"package": ps.PackageName,
			"count":   ps.Count,
		})
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"total_restaurants":  totalRestaurants,
		"active_restaurants": activeRestaurants,
		"new_this_month":     newThisMonth,
		"by_package":         byPackage,
	}, "")
}

// ===============================
// PACKAGE UPGRADE HANDLERS
// ===============================

// UpgradePackageInput input cho upgrade package
type UpgradePackageInput struct {
	PackageID    uint   `json:"package_id" binding:"required"`
	BillingCycle string `json:"billing_cycle"` // monthly or yearly
}

// CreateUpgradeSubscription tạo yêu cầu nâng cấp gói với thanh toán QR
// @Summary Nâng cấp gói dịch vụ
// @Description Tạo yêu cầu nâng cấp gói và nhận QR thanh toán
// @Tags Packages
// @Accept json
// @Produce json
// @Param body body UpgradePackageInput true "Thông tin nâng cấp"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/upgrade [post]
func CreateUpgradeSubscription(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	if currentRestaurantID == nil || uint(restaurantID) != *currentRestaurantID.(*uint) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền", "FORBIDDEN", "")
		return
	}

	var input UpgradePackageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	// Lấy thông tin nhà hàng
	var restaurant models.Restaurant
	if err := config.GetDB().Preload("Package").First(&restaurant, restaurantID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy nhà hàng", "NOT_FOUND", "")
		return
	}

	// Lấy gói muốn upgrade
	var newPackage models.Package
	if err := config.GetDB().First(&newPackage, input.PackageID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy gói dịch vụ", "PACKAGE_NOT_FOUND", "")
		return
	}

	// Xác định giá
	billingCycle := input.BillingCycle
	if billingCycle == "" {
		billingCycle = "monthly"
	}

	var amount float64
	if billingCycle == "yearly" {
		amount = newPackage.YearlyPrice
	} else {
		amount = newPackage.MonthlyPrice
	}

	// Tạo mã đăng ký unique
	subscriptionCode := utils.GenerateRandomCode(8)
	expiresAt := time.Now().Add(24 * time.Hour) // Hết hạn sau 24h

	// Lưu vào PackageSubscription
	subscription := models.PackageSubscription{
		Email:          getStringValue(restaurant.Email),
		Phone:          restaurant.Phone,
		Name:           restaurant.Name,
		RestaurantName: restaurant.Name,
		PackageID:      newPackage.ID,
		BillingCycle:   billingCycle,
		Amount:         amount,
		PaymentCode:    subscriptionCode,
		PaymentStatus:  "pending",
		RestaurantID:   &restaurant.ID,
		ExpiresAt:      expiresAt,
	}

	if err := config.GetDB().Create(&subscription).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo đăng ký", "CREATE_ERROR", err.Error())
		return
	}

	// Lấy thông tin SePay để tạo QR
	sepayConfig := config.GetSepayConfig()
	if sepayConfig.BankCode == "" {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Chưa cấu hình thanh toán", "SEPAY_NOT_CONFIGURED", "")
		return
	}

	// Tạo QR URL
	description := "UPGRADE " + subscriptionCode
	qrURL := utils.GenerateSepayQRURL(
		sepayConfig.BankCode,
		sepayConfig.AccountNumber,
		sepayConfig.AccountName,
		amount,
		description,
	)

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"subscription_code": subscriptionCode,
		"package":           newPackage.DisplayName,
		"billing_cycle":     billingCycle,
		"amount":            amount,
		"qr_url":            qrURL,
		"bank_info": gin.H{
			"bank_code":      sepayConfig.BankCode,
			"account_number": sepayConfig.AccountNumber,
			"account_name":   sepayConfig.AccountName,
		},
		"message": "Quét mã QR để thanh toán. Gói sẽ được kích hoạt sau khi thanh toán thành công.",
	}, "Tạo yêu cầu nâng cấp thành công")
}

// GetUpgradeStatus kiểm tra trạng thái nâng cấp
// @Summary Kiểm tra trạng thái nâng cấp
// @Description Kiểm tra xem thanh toán nâng cấp đã hoàn tất chưa
// @Tags Packages
// @Produce json
// @Param code path string true "Subscription Code"
// @Success 200 {object} map[string]interface{}
// @Router /upgrade/{code}/status [get]
func GetUpgradeStatus(c *gin.Context) {
	code := c.Param("code")

	var subscription models.PackageSubscription
	if err := config.GetDB().Where("payment_code = ?", code).First(&subscription).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy đăng ký", "NOT_FOUND", "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"code":          subscription.PaymentCode,
		"status":        subscription.PaymentStatus,
		"package_id":    subscription.PackageID,
		"amount":        subscription.Amount,
		"billing_cycle": subscription.BillingCycle,
		"paid_at":       subscription.PaidAt,
	}, "")
}

// Helper function
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
