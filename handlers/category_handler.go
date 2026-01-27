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

// CreateCategoryInput request body cho tạo danh mục
type CreateCategoryInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Image       string `json:"image"`
	SortOrder   int    `json:"sort_order"`
}

// UpdateCategoryInput request body cho update danh mục
type UpdateCategoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	SortOrder   int    `json:"sort_order"`
	Status      string `json:"status"`
}

// ===============================
// HANDLERS
// ===============================

// GetCategories lấy danh sách danh mục của nhà hàng (Public)
// @Summary Lấy danh sách danh mục
// @Description Lấy tất cả danh mục của nhà hàng
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Param status query string false "Filter theo status" default(active)
// @Success 200 {object} map[string]interface{}
// @Router /restaurants/{id}/categories [get]
func GetCategories(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var categories []models.Category
	query := config.GetDB().Where("restaurant_id = ?", restaurantID)

	// Filter by status
	status := c.DefaultQuery("status", "active")
	if status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("sort_order ASC, id ASC").Find(&categories).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Lỗi khi lấy danh sách danh mục", "QUERY_ERROR", err.Error())
		return
	}

	var data []gin.H
	for _, cat := range categories {
		// Đếm số món trong danh mục
		var itemsCount int64
		config.GetDB().Model(&models.MenuItem{}).Where("category_id = ? AND status = ?", cat.ID, "active").Count(&itemsCount)

		data = append(data, gin.H{
			"id":          cat.ID,
			"name":        cat.Name,
			"description": cat.Description,
			"image":       cat.Image,
			"sort_order":  cat.SortOrder,
			"status":      cat.Status,
			"items_count": itemsCount,
		})
	}

	utils.SuccessResponse(c, http.StatusOK, data, "")
}

// CreateCategory tạo danh mục mới
// @Summary Tạo danh mục
// @Description Tạo danh mục món ăn mới
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Param category body CreateCategoryInput true "Thông tin danh mục"
// @Success 201 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/categories [post]
func CreateCategory(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || uint(restaurantID) != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền tạo danh mục cho nhà hàng này", "FORBIDDEN", "")
		return
	}

	var input CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	// Kiểm tra giới hạn danh mục theo package
	var restaurant models.Restaurant
	if err := config.GetDB().Preload("Package").First(&restaurant, restaurantID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy nhà hàng", "RESTAURANT_NOT_FOUND", "")
		return
	}

	var catCount int64
	config.GetDB().Model(&models.Category{}).Where("restaurant_id = ?", restaurantID).Count(&catCount)

	if restaurant.Package.MaxCategories != -1 && int(catCount) >= restaurant.Package.MaxCategories {
		utils.ErrorResponse(c, http.StatusForbidden, "Đã đạt giới hạn số danh mục của gói dịch vụ", "CATEGORY_LIMIT_EXCEEDED", "")
		return
	}

	category := models.Category{
		RestaurantID: uint(restaurantID),
		Name:         input.Name,
		Description:  &input.Description,
		Image:        &input.Image,
		SortOrder:    input.SortOrder,
		Status:       "active",
	}

	if err := config.GetDB().Create(&category).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo danh mục", "CREATE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"id":          category.ID,
		"name":        category.Name,
		"description": category.Description,
		"image":       category.Image,
		"sort_order":  category.SortOrder,
		"status":      category.Status,
	}, "Tạo danh mục thành công")
}

// UpdateCategory cập nhật danh mục
// @Summary Cập nhật danh mục
// @Description Cập nhật thông tin danh mục
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body UpdateCategoryInput true "Thông tin cập nhật"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /categories/{id} [put]
func UpdateCategory(c *gin.Context) {
	categoryID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var category models.Category
	if err := config.GetDB().First(&category, categoryID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy danh mục", "CATEGORY_NOT_FOUND", "")
		return
	}

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || category.RestaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền chỉnh sửa danh mục này", "FORBIDDEN", "")
		return
	}

	var input UpdateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	updates := make(map[string]interface{})
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if input.Image != "" {
		updates["image"] = input.Image
	}
	if input.SortOrder > 0 {
		updates["sort_order"] = input.SortOrder
	}
	if input.Status != "" {
		if input.Status != "active" && input.Status != "inactive" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Trạng thái không hợp lệ", "INVALID_STATUS", "")
			return
		}
		updates["status"] = input.Status
	}

	if err := config.GetDB().Model(&category).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể cập nhật danh mục", "UPDATE_ERROR", err.Error())
		return
	}

	config.GetDB().First(&category, categoryID)

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":          category.ID,
		"name":        category.Name,
		"description": category.Description,
		"image":       category.Image,
		"sort_order":  category.SortOrder,
		"status":      category.Status,
	}, "Cập nhật danh mục thành công")
}

// DeleteCategory xóa danh mục
// @Summary Xóa danh mục
// @Description Xóa danh mục (chỉ khi không có món)
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /categories/{id} [delete]
func DeleteCategory(c *gin.Context) {
	categoryID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var category models.Category
	if err := config.GetDB().First(&category, categoryID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy danh mục", "CATEGORY_NOT_FOUND", "")
		return
	}

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || category.RestaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền xóa danh mục này", "FORBIDDEN", "")
		return
	}

	// Kiểm tra có món ăn trong danh mục không
	var itemCount int64
	config.GetDB().Model(&models.MenuItem{}).Where("category_id = ?", categoryID).Count(&itemCount)
	if itemCount > 0 {
		utils.ErrorResponse(c, http.StatusConflict, "Không thể xóa danh mục có món ăn", "CATEGORY_HAS_ITEMS", "")
		return
	}

	if err := config.GetDB().Delete(&category).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể xóa danh mục", "DELETE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, nil, "Xóa danh mục thành công")
}

// GetCategoryItems lấy món theo danh mục (Public)
// @Summary Lấy món theo danh mục
// @Description Lấy danh sách món ăn trong một danh mục
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param status query string false "Filter theo status" default(active)
// @Success 200 {object} map[string]interface{}
// @Router /categories/{id}/items [get]
func GetCategoryItems(c *gin.Context) {
	categoryID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var category models.Category
	if err := config.GetDB().First(&category, categoryID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy danh mục", "CATEGORY_NOT_FOUND", "")
		return
	}

	var items []models.MenuItem
	query := config.GetDB().Where("category_id = ?", categoryID)

	status := c.DefaultQuery("status", "active")
	if status != "all" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("sort_order ASC, id ASC").Find(&items).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Lỗi khi lấy danh sách món", "QUERY_ERROR", err.Error())
		return
	}

	var data []gin.H
	for _, item := range items {
		data = append(data, gin.H{
			"id":            item.ID,
			"name":          item.Name,
			"description":   item.Description,
			"price":         item.Price,
			"image":         item.Image,
			"options":       item.Options,
			"tags":          item.Tags,
			"prep_location": item.PrepLocation,
			"prep_time":     item.PrepTime,
			"status":        item.Status,
		})
	}

	utils.SuccessResponse(c, http.StatusOK, data, "")
}
