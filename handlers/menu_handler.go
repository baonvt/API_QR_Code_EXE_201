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

// CreateMenuItemInput request body cho tạo món
type CreateMenuItemInput struct {
	CategoryID   uint    `json:"category_id" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description"`
	Price        float64 `json:"price" binding:"required"`
	Image        string  `json:"image"`
	Options      string  `json:"options"`       // JSON string
	Tags         string  `json:"tags"`          // JSON string
	PrepLocation string  `json:"prep_location"` // kitchen, bar
	PrepTime     int     `json:"prep_time"`
	SortOrder    int     `json:"sort_order"`
}

// UpdateMenuItemInput request body cho update món
type UpdateMenuItemInput struct {
	CategoryID   uint    `json:"category_id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Image        string  `json:"image"`
	Options      string  `json:"options"`
	Tags         string  `json:"tags"`
	PrepLocation string  `json:"prep_location"`
	PrepTime     int     `json:"prep_time"`
	SortOrder    int     `json:"sort_order"`
	Status       string  `json:"status"`
}

// ===============================
// HANDLERS
// ===============================

// GetMenu lấy toàn bộ menu của nhà hàng (Public)
// @Summary Lấy menu nhà hàng
// @Description Lấy danh sách tất cả món ăn của nhà hàng
// @Tags Menu
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Param category_id query int false "Filter theo danh mục"
// @Param status query string false "Filter theo status" default(active)
// @Success 200 {object} map[string]interface{}
// @Router /restaurants/{id}/menu [get]
func GetMenu(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var items []models.MenuItem
	query := config.GetDB().Preload("Category").Where("restaurant_id = ?", restaurantID)

	// Filter by category
	categoryID := c.Query("category_id")
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	// Filter by status
	status := c.DefaultQuery("status", "active")
	if status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Lỗi khi lấy menu", "QUERY_ERROR", err.Error())
		return
	}

	var data []gin.H
	for _, item := range items {
		categoryName := ""
		if item.Category != nil {
			categoryName = item.Category.Name
		}

		data = append(data, gin.H{
			"id":            item.ID,
			"name":          item.Name,
			"description":   item.Description,
			"price":         item.Price,
			"image":         item.Image,
			"category_id":   item.CategoryID,
			"category_name": categoryName,
			"options":       item.Options,
			"tags":          item.Tags,
			"prep_location": item.PrepLocation,
			"prep_time":     item.PrepTime,
			"status":        item.Status,
		})
	}

	utils.SuccessResponse(c, http.StatusOK, data, "")
}

// GetMenuBySlug lấy menu theo slug nhà hàng (Public - cho khách)
// @Summary Lấy menu theo slug
// @Description Lấy menu đầy đủ của nhà hàng theo slug (cho khách hàng)
// @Tags Public
// @Accept json
// @Produce json
// @Param slug path string true "Restaurant Slug"
// @Success 200 {object} map[string]interface{}
// @Router /public/restaurants/{slug}/menu [get]
func GetMenuBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var restaurant models.Restaurant
	if err := config.GetDB().Where("slug = ? AND status = ?", slug, "active").First(&restaurant).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy nhà hàng", "RESTAURANT_NOT_FOUND", "")
		return
	}

	// Lấy danh mục kèm món
	var categories []models.Category
	if err := config.GetDB().Where("restaurant_id = ? AND status = ?", restaurant.ID, "active").
		Order("sort_order ASC").Find(&categories).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Lỗi khi lấy danh mục", "QUERY_ERROR", err.Error())
		return
	}

	var menuData []gin.H
	for _, cat := range categories {
		var items []models.MenuItem
		config.GetDB().Where("category_id = ? AND status = ?", cat.ID, "active").
			Order("sort_order ASC").Find(&items)

		var itemsData []gin.H
		for _, item := range items {
			itemsData = append(itemsData, gin.H{
				"id":          item.ID,
				"name":        item.Name,
				"description": item.Description,
				"price":       item.Price,
				"image":       item.Image,
				"options":     item.Options,
				"tags":        item.Tags,
			})
		}

		menuData = append(menuData, gin.H{
			"id":          cat.ID,
			"name":        cat.Name,
			"description": cat.Description,
			"image":       cat.Image,
			"items":       itemsData,
		})
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"restaurant": gin.H{
			"id":             restaurant.ID,
			"name":           restaurant.Name,
			"slug":           restaurant.Slug,
			"is_open":        restaurant.IsOpen,
			"tax_rate":       restaurant.TaxRate,
			"service_charge": restaurant.ServiceCharge,
		},
		"menu": menuData,
	}, "")
}

// CreateMenuItem tạo món mới
// @Summary Tạo món mới
// @Description Tạo một món ăn mới cho nhà hàng
// @Tags Menu
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Param menu body CreateMenuItemInput true "Thông tin món ăn"
// @Success 201 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/menu [post]
func CreateMenuItem(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || uint(restaurantID) != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền tạo món cho nhà hàng này", "FORBIDDEN", "")
		return
	}

	var input CreateMenuItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	// Kiểm tra giới hạn món theo package
	var restaurant models.Restaurant
	if err := config.GetDB().Preload("Package").First(&restaurant, restaurantID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy nhà hàng", "RESTAURANT_NOT_FOUND", "")
		return
	}

	var itemCount int64
	config.GetDB().Model(&models.MenuItem{}).Where("restaurant_id = ?", restaurantID).Count(&itemCount)

	if restaurant.Package.MaxMenuItems != -1 && int(itemCount) >= restaurant.Package.MaxMenuItems {
		utils.ErrorResponse(c, http.StatusForbidden, "Đã đạt giới hạn số món của gói dịch vụ", "MENU_LIMIT_EXCEEDED", "")
		return
	}

	// Kiểm tra category thuộc về nhà hàng
	var category models.Category
	if err := config.GetDB().Where("id = ? AND restaurant_id = ?", input.CategoryID, restaurantID).First(&category).Error; err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Danh mục không hợp lệ", "INVALID_CATEGORY", "")
		return
	}

	prepLocation := "kitchen"
	if input.PrepLocation != "" {
		prepLocation = input.PrepLocation
	}

	prepTime := 15
	if input.PrepTime > 0 {
		prepTime = input.PrepTime
	}

	item := models.MenuItem{
		RestaurantID: uint(restaurantID),
		CategoryID:   input.CategoryID,
		Name:         input.Name,
		Description:  &input.Description,
		Price:        input.Price,
		Image:        &input.Image,
		Options:      &input.Options,
		Tags:         &input.Tags,
		PrepLocation: prepLocation,
		PrepTime:     prepTime,
		SortOrder:    input.SortOrder,
		Status:       "active",
	}

	if err := config.GetDB().Create(&item).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo món", "CREATE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"id":            item.ID,
		"name":          item.Name,
		"description":   item.Description,
		"price":         item.Price,
		"image":         item.Image,
		"category_id":   item.CategoryID,
		"options":       item.Options,
		"tags":          item.Tags,
		"prep_location": item.PrepLocation,
		"prep_time":     item.PrepTime,
		"status":        item.Status,
	}, "Tạo món thành công")
}

// UpdateMenuItem cập nhật món
// @Summary Cập nhật món
// @Description Cập nhật thông tin món ăn
// @Tags Menu
// @Accept json
// @Produce json
// @Param id path int true "Menu Item ID"
// @Param menu body UpdateMenuItemInput true "Thông tin cập nhật"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /menu/{id} [put]
func UpdateMenuItem(c *gin.Context) {
	itemID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var item models.MenuItem
	if err := config.GetDB().First(&item, itemID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy món", "MENU_ITEM_NOT_FOUND", "")
		return
	}

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || item.RestaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền chỉnh sửa món này", "FORBIDDEN", "")
		return
	}

	var input UpdateMenuItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	updates := make(map[string]interface{})

	if input.CategoryID > 0 {
		// Kiểm tra category thuộc về nhà hàng
		var category models.Category
		if err := config.GetDB().Where("id = ? AND restaurant_id = ?", input.CategoryID, item.RestaurantID).First(&category).Error; err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Danh mục không hợp lệ", "INVALID_CATEGORY", "")
			return
		}
		updates["category_id"] = input.CategoryID
	}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if input.Price > 0 {
		updates["price"] = input.Price
	}
	if input.Image != "" {
		updates["image"] = input.Image
	}
	if input.Options != "" {
		updates["options"] = input.Options
	}
	if input.Tags != "" {
		updates["tags"] = input.Tags
	}
	if input.PrepLocation != "" {
		updates["prep_location"] = input.PrepLocation
	}
	if input.PrepTime > 0 {
		updates["prep_time"] = input.PrepTime
	}
	if input.SortOrder > 0 {
		updates["sort_order"] = input.SortOrder
	}
	if input.Status != "" {
		validStatuses := []string{"active", "inactive", "out_of_stock"}
		isValid := false
		for _, s := range validStatuses {
			if s == input.Status {
				isValid = true
				break
			}
		}
		if !isValid {
			utils.ErrorResponse(c, http.StatusBadRequest, "Trạng thái không hợp lệ", "INVALID_STATUS", "")
			return
		}
		updates["status"] = input.Status
	}

	if err := config.GetDB().Model(&item).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể cập nhật món", "UPDATE_ERROR", err.Error())
		return
	}

	config.GetDB().First(&item, itemID)

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":            item.ID,
		"name":          item.Name,
		"description":   item.Description,
		"price":         item.Price,
		"image":         item.Image,
		"category_id":   item.CategoryID,
		"options":       item.Options,
		"tags":          item.Tags,
		"prep_location": item.PrepLocation,
		"prep_time":     item.PrepTime,
		"status":        item.Status,
	}, "Cập nhật món thành công")
}

// DeleteMenuItem xóa món
// @Summary Xóa món
// @Description Xóa (vô hiệu hóa) món ăn
// @Tags Menu
// @Accept json
// @Produce json
// @Param id path int true "Menu Item ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /menu/{id} [delete]
func DeleteMenuItem(c *gin.Context) {
	itemID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var item models.MenuItem
	if err := config.GetDB().First(&item, itemID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy món", "MENU_ITEM_NOT_FOUND", "")
		return
	}

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || item.RestaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền xóa món này", "FORBIDDEN", "")
		return
	}

	// Soft delete - đánh dấu status = inactive thay vì xóa hẳn
	if err := config.GetDB().Model(&item).Update("status", "inactive").Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể xóa món", "DELETE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, nil, "Xóa món thành công")
}
