package handlers

import (
	"net/http"
	"strconv"

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
