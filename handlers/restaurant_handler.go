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
// REQUEST STRUCTS
// ===============================

// UpdateRestaurantInput request body cho update restaurant
type UpdateRestaurantInput struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Logo          string  `json:"logo"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	Address       string  `json:"address"`
	IsOpen        *bool   `json:"is_open"`
	TaxRate       float64 `json:"tax_rate"`
	ServiceCharge float64 `json:"service_charge"`
}

// ===============================
// HANDLERS
// ===============================

// GetRestaurantBySlug lấy thông tin nhà hàng theo slug (Public)
// @Summary Lấy nhà hàng theo slug
// @Description Lấy thông tin nhà hàng theo slug (cho khách)
// @Tags Public
// @Accept json
// @Produce json
// @Param slug path string true "Restaurant Slug"
// @Success 200 {object} map[string]interface{}
// @Router /public/restaurants/{slug} [get]
func GetRestaurantBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var restaurant models.Restaurant
	if err := config.GetDB().Where("slug = ? AND status = ?", slug, "active").First(&restaurant).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy nhà hàng", "RESTAURANT_NOT_FOUND", "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":             restaurant.ID,
		"name":           restaurant.Name,
		"slug":           restaurant.Slug,
		"description":    restaurant.Description,
		"logo":           restaurant.Logo,
		"phone":          restaurant.Phone,
		"address":        restaurant.Address,
		"is_open":        restaurant.IsOpen,
		"tax_rate":       restaurant.TaxRate,
		"service_charge": restaurant.ServiceCharge,
	}, "")
}

// GetMyRestaurant lấy thông tin nhà hàng của user hiện tại
// @Summary Lấy nhà hàng của tôi
// @Description Lấy thông tin nhà hàng của user đang đăng nhập
// @Tags Restaurants
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/me [get]
func GetMyRestaurant(c *gin.Context) {
	restaurantID, exists := c.Get("restaurant_id")
	if !exists || restaurantID == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Bạn chưa có nhà hàng", "NO_RESTAURANT", "")
		return
	}

	// Lấy user_id từ context
	userID, _ := c.Get("user_id")

	var restaurant models.Restaurant
	if err := config.GetDB().Preload("Package").First(&restaurant, restaurantID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy nhà hàng", "RESTAURANT_NOT_FOUND", "")
		return
	}

	// Lấy thông tin owner
	var owner models.User
	config.GetDB().First(&owner, userID)

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":               restaurant.ID,
		"name":             owner.Name,
		"email":            owner.Email,
		"phone":            owner.Phone,
		"avatar":           owner.Avatar,
		"role":             owner.Role,
		"restaurantId":     restaurant.ID,
		"restaurantName":   restaurant.Name,
		"slug":             restaurant.Slug,
		"description":      restaurant.Description,
		"logo":             restaurant.Logo,
		"address":          restaurant.Address,
		"is_open":          restaurant.IsOpen,
		"tax_rate":         restaurant.TaxRate,
		"service_charge":   restaurant.ServiceCharge,
		"currency":         restaurant.Currency,
		"package_status":   restaurant.PackageStatus,
		"package_end_date": restaurant.PackageEndDate,
		"status":           restaurant.Status,
		"package": gin.H{
			"id":           restaurant.Package.ID,
			"name":         restaurant.Package.Name,
			"display_name": restaurant.Package.DisplayName,
		},
	}, "")
}

// UpdateRestaurant cập nhật thông tin nhà hàng
// @Summary Cập nhật nhà hàng
// @Description Cập nhật thông tin nhà hàng
// @Tags Restaurants
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Param restaurant body UpdateRestaurantInput true "Thông tin cập nhật"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id} [put]
func UpdateRestaurant(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Kiểm tra quyền sở hữu
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || uint(restaurantID) != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền chỉnh sửa nhà hàng này", "FORBIDDEN", "")
		return
	}

	var restaurant models.Restaurant
	if err := config.GetDB().First(&restaurant, restaurantID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy nhà hàng", "RESTAURANT_NOT_FOUND", "")
		return
	}

	var input UpdateRestaurantInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	// Cập nhật các trường
	updates := make(map[string]interface{})
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if input.Logo != "" {
		updates["logo"] = input.Logo
	}
	if input.Phone != "" {
		updates["phone"] = input.Phone
	}
	if input.Email != "" {
		updates["email"] = input.Email
	}
	if input.Address != "" {
		updates["address"] = input.Address
	}
	if input.IsOpen != nil {
		updates["is_open"] = *input.IsOpen
	}
	if input.TaxRate > 0 {
		updates["tax_rate"] = input.TaxRate
	}
	if input.ServiceCharge >= 0 {
		updates["service_charge"] = input.ServiceCharge
	}

	if err := config.GetDB().Model(&restaurant).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể cập nhật nhà hàng", "UPDATE_ERROR", err.Error())
		return
	}

	// Lấy lại thông tin mới
	config.GetDB().First(&restaurant, restaurantID)

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":             restaurant.ID,
		"name":           restaurant.Name,
		"slug":           restaurant.Slug,
		"description":    restaurant.Description,
		"logo":           restaurant.Logo,
		"phone":          restaurant.Phone,
		"email":          restaurant.Email,
		"address":        restaurant.Address,
		"is_open":        restaurant.IsOpen,
		"tax_rate":       restaurant.TaxRate,
		"service_charge": restaurant.ServiceCharge,
	}, "Cập nhật nhà hàng thành công")
}

// ===============================
// ADMIN HANDLERS
// ===============================

// GetAllRestaurants lấy danh sách tất cả nhà hàng (Admin)
// @Summary Lấy tất cả nhà hàng
// @Description Admin lấy danh sách tất cả nhà hàng
// @Tags Admin
// @Accept json
// @Produce json
// @Param status query string false "Filter theo status"
// @Param package_status query string false "Filter theo package_status"
// @Param search query string false "Tìm kiếm theo tên"
// @Param page query int false "Trang" default(1)
// @Param limit query int false "Số lượng/trang" default(20)
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /admin/restaurants [get]
func GetAllRestaurants(c *gin.Context) {
	var restaurants []models.Restaurant

	query := config.GetDB().Preload("Owner").Preload("Package")

	// Filter by status
	status := c.Query("status")
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	// Filter by package_status
	packageStatus := c.Query("package_status")
	if packageStatus != "" {
		query = query.Where("package_status = ?", packageStatus)
	}

	// Search
	search := c.Query("search")
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Restaurant{}).Count(&total)

	if err := query.Offset(offset).Limit(limit).Find(&restaurants).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Lỗi khi lấy danh sách nhà hàng", "QUERY_ERROR", err.Error())
		return
	}

	// Build response
	var data []gin.H
	for _, r := range restaurants {
		data = append(data, gin.H{
			"id":             r.ID,
			"name":           r.Name,
			"slug":           r.Slug,
			"status":         r.Status,
			"package_status": r.PackageStatus,
			"owner": gin.H{
				"id":    r.Owner.ID,
				"name":  r.Owner.Name,
				"email": r.Owner.Email,
			},
			"package": gin.H{
				"id":           r.Package.ID,
				"name":         r.Package.Name,
				"display_name": r.Package.DisplayName,
			},
			"created_at": r.CreatedAt,
		})
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	utils.PaginatedResponse(c, http.StatusOK, data, utils.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, "")
}

// UpdateRestaurantStatus cập nhật trạng thái nhà hàng (Admin)
// @Summary Cập nhật trạng thái nhà hàng
// @Description Admin cập nhật trạng thái nhà hàng (active/suspended/deleted)
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /admin/restaurants/{id}/status [put]
func UpdateRestaurantStatus(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var restaurant models.Restaurant
	if err := config.GetDB().First(&restaurant, restaurantID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy nhà hàng", "RESTAURANT_NOT_FOUND", "")
		return
	}

	var input struct {
		Status string `json:"status" binding:"required,oneof=active suspended deleted"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	if err := config.GetDB().Model(&restaurant).Update("status", input.Status).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể cập nhật trạng thái", "UPDATE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":     restaurant.ID,
		"status": input.Status,
	}, "Cập nhật trạng thái thành công")
}
