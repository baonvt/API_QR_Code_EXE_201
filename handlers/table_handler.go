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

// CreateTableInput request body cho tạo bàn
type CreateTableInput struct {
	TableNumber int    `json:"table_number" binding:"required"`
	Name        string `json:"name"`
	Capacity    int    `json:"capacity"`
}

// UpdateTableInput request body cho update bàn
type UpdateTableInput struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Status   string `json:"status"`
}

// ===============================
// HANDLERS
// ===============================

// GetTables lấy danh sách bàn của nhà hàng
// @Summary Lấy danh sách bàn
// @Description Lấy tất cả bàn của nhà hàng
// @Tags Tables
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Param status query string false "Filter theo status" Enums(available, occupied, reserved)
// @Param is_active query bool false "Filter theo is_active"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/tables [get]
func GetTables(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || uint(restaurantID) != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền xem bàn của nhà hàng này", "FORBIDDEN", "")
		return
	}

	var tables []models.Table
	query := config.GetDB().Where("restaurant_id = ?", restaurantID)

	// Filter by status
	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by active
	isActive := c.Query("is_active")
	if isActive == "true" {
		query = query.Where("is_active = ?", true)
	} else if isActive == "false" {
		query = query.Where("is_active = ?", false)
	}

	if err := query.Order("table_number ASC").Find(&tables).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Lỗi khi lấy danh sách bàn", "QUERY_ERROR", err.Error())
		return
	}

	// Lấy restaurant để tạo QR URL
	var restaurant models.Restaurant
	config.GetDB().First(&restaurant, restaurantID)

	var data []gin.H
	for _, t := range tables {
		data = append(data, gin.H{
			"id":           t.ID,
			"table_number": t.TableNumber,
			"name":         t.Name,
			"capacity":     t.Capacity,
			"status":       t.Status,
			"is_active":    t.IsActive,
			"qr_url":       "/" + restaurant.Slug + "/menu/" + strconv.Itoa(t.TableNumber),
		})
	}

	utils.SuccessResponse(c, http.StatusOK, data, "")
}

// CreateTable tạo bàn mới
// @Summary Tạo bàn mới
// @Description Tạo bàn mới cho nhà hàng
// @Tags Tables
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Param table body CreateTableInput true "Thông tin bàn"
// @Success 201 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/tables [post]
func CreateTable(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || uint(restaurantID) != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền tạo bàn cho nhà hàng này", "FORBIDDEN", "")
		return
	}

	var input CreateTableInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	// Kiểm tra giới hạn bàn theo package
	var restaurant models.Restaurant
	if err := config.GetDB().Preload("Package").First(&restaurant, restaurantID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy nhà hàng", "RESTAURANT_NOT_FOUND", "")
		return
	}

	var tableCount int64
	config.GetDB().Model(&models.Table{}).Where("restaurant_id = ?", restaurantID).Count(&tableCount)

	if restaurant.Package.MaxTables != -1 && int(tableCount) >= restaurant.Package.MaxTables {
		utils.ErrorResponse(c, http.StatusForbidden, "Đã đạt giới hạn số bàn của gói dịch vụ", "TABLE_LIMIT_EXCEEDED", "")
		return
	}

	// Kiểm tra table_number đã tồn tại chưa
	var existingTable models.Table
	if err := config.GetDB().Where("restaurant_id = ? AND table_number = ?", restaurantID, input.TableNumber).First(&existingTable).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Số bàn đã tồn tại", "TABLE_NUMBER_EXISTS", "")
		return
	}

	// Tạo bàn mới
	capacity := 4
	if input.Capacity > 0 {
		capacity = input.Capacity
	}

	table := models.Table{
		RestaurantID: uint(restaurantID),
		TableNumber:  input.TableNumber,
		Name:         &input.Name,
		Capacity:     capacity,
		Status:       "available",
		IsActive:     true,
	}

	if err := config.GetDB().Create(&table).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo bàn", "CREATE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"id":           table.ID,
		"table_number": table.TableNumber,
		"name":         table.Name,
		"capacity":     table.Capacity,
		"status":       table.Status,
		"qr_url":       "/" + restaurant.Slug + "/menu/" + strconv.Itoa(table.TableNumber),
	}, "Tạo bàn thành công")
}

// UpdateTable cập nhật bàn
// @Summary Cập nhật bàn
// @Description Cập nhật thông tin bàn
// @Tags Tables
// @Accept json
// @Produce json
// @Param id path int true "Table ID"
// @Param table body UpdateTableInput true "Thông tin cập nhật"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /tables/{id} [put]
func UpdateTable(c *gin.Context) {
	tableID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var table models.Table
	if err := config.GetDB().First(&table, tableID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy bàn", "TABLE_NOT_FOUND", "")
		return
	}

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || table.RestaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền chỉnh sửa bàn này", "FORBIDDEN", "")
		return
	}

	var input UpdateTableInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	updates := make(map[string]interface{})
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Capacity > 0 {
		updates["capacity"] = input.Capacity
	}
	if input.Status != "" {
		// Validate status
		validStatuses := []string{"available", "occupied", "reserved"}
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

	if err := config.GetDB().Model(&table).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể cập nhật bàn", "UPDATE_ERROR", err.Error())
		return
	}

	config.GetDB().First(&table, tableID)

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":           table.ID,
		"table_number": table.TableNumber,
		"name":         table.Name,
		"capacity":     table.Capacity,
		"status":       table.Status,
	}, "Cập nhật bàn thành công")
}

// DeleteTable xóa bàn
// @Summary Xóa bàn
// @Description Xóa (vô hiệu hóa) bàn
// @Tags Tables
// @Accept json
// @Produce json
// @Param id path int true "Table ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /tables/{id} [delete]
func DeleteTable(c *gin.Context) {
	tableID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var table models.Table
	if err := config.GetDB().First(&table, tableID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy bàn", "TABLE_NOT_FOUND", "")
		return
	}

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || table.RestaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền xóa bàn này", "FORBIDDEN", "")
		return
	}

	// Kiểm tra có đơn hàng đang xử lý không
	var orderCount int64
	config.GetDB().Model(&models.Order{}).Where("table_id = ? AND status NOT IN ?", tableID, []string{"completed", "cancelled"}).Count(&orderCount)
	if orderCount > 0 {
		utils.ErrorResponse(c, http.StatusConflict, "Không thể xóa bàn đang có đơn hàng", "TABLE_HAS_ORDERS", "")
		return
	}

	// Soft delete - chỉ đánh dấu is_active = false
	if err := config.GetDB().Model(&table).Update("is_active", false).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể xóa bàn", "DELETE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, nil, "Xóa bàn thành công")
}
