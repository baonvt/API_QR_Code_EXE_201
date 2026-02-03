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
	IsActive *bool  `json:"is_active"` // Pointer để phân biệt false vs không gửi
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

	// Cập nhật is_active nếu được gửi
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
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
		"is_active":    table.IsActive,
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

	// Kiểm tra query param để xác định soft delete hay hard delete
	hardDelete := c.Query("hard") == "true"

	if hardDelete {
		// Hard delete - xóa hoàn toàn khỏi database
		if err := config.GetDB().Unscoped().Delete(&table).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể xóa hoàn toàn bàn", "DELETE_ERROR", err.Error())
			return
		}
		utils.SuccessResponse(c, http.StatusOK, nil, "Đã xóa hoàn toàn bàn")
	} else {
		// Soft delete - chỉ đánh dấu is_active = false
		if err := config.GetDB().Model(&table).Update("is_active", false).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể xóa bàn", "DELETE_ERROR", err.Error())
			return
		}
		utils.SuccessResponse(c, http.StatusOK, nil, "Đã vô hiệu hóa bàn")
	}
}

// GetTableDetail lấy chi tiết bàn với đơn hàng đang hoạt động
// @Summary Lấy chi tiết bàn
// @Description Lấy thông tin bàn kèm đơn hàng đang có và chi tiết từng món
// @Tags Tables
// @Accept json
// @Produce json
// @Param id path int true "Table ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /tables/{id}/detail [get]
func GetTableDetail(c *gin.Context) {
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
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền xem bàn này", "FORBIDDEN", "")
		return
	}

	// Lấy đơn hàng đang hoạt động của bàn (chưa completed, chưa cancelled)
	var activeOrders []models.Order
	config.GetDB().
		Preload("Items").
		Preload("OrderItems.MenuItem").
		Where("table_id = ? AND status NOT IN ?", tableID, []string{"completed", "cancelled"}).
		Order("created_at DESC").
		Find(&activeOrders)

	// Format orders với chi tiết từng món
	var ordersData []gin.H
	for _, order := range activeOrders {
		var itemsData []gin.H
		for _, item := range order.OrderItems {
			itemData := gin.H{
				"id":          item.ID,
				"quantity":    item.Quantity,
				"item_name":   item.ItemName,
				"item_price":  item.ItemPrice,
				"line_total":  item.LineTotal,
				"notes":       item.Notes,
				"prep_status": item.PrepStatus,
			}
			if item.MenuItem != nil {
				itemData["menu_item"] = gin.H{
					"id":    item.MenuItem.ID,
					"name":  item.MenuItem.Name,
					"image": item.MenuItem.Image,
					"price": item.MenuItem.Price,
				}
			}
			itemsData = append(itemsData, itemData)
		}

		ordersData = append(ordersData, gin.H{
			"id":             order.ID,
			"order_number":   order.OrderNumber,
			"status":         order.Status,
			"payment_status": order.PaymentStatus,
			"payment_method": order.PaymentMethod,
			"total_amount":   order.TotalAmount,
			"notes":          order.Notes,
			"created_at":     order.CreatedAt,
			"items":          itemsData,
			"items_count":    len(order.OrderItems),
		})
	}

	// Tính tổng tiền các order đang hoạt động
	var totalAmount float64
	for _, order := range activeOrders {
		totalAmount += order.TotalAmount
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"table": gin.H{
			"id":           table.ID,
			"table_number": table.TableNumber,
			"name":         table.Name,
			"capacity":     table.Capacity,
			"status":       table.Status,
			"is_active":    table.IsActive,
		},
		"active_orders":       ordersData,
		"active_orders_count": len(activeOrders),
		"total_amount":        totalAmount,
	}, "")
}

// GetTableBySlugAndNumber lấy thông tin bàn theo slug nhà hàng và số bàn (Public)
// @Summary Lấy bàn theo slug nhà hàng và số bàn
// @Description Lấy thông tin bàn theo slug nhà hàng và số bàn (cho khách quét QR)
// @Tags Public
// @Accept json
// @Produce json
// @Param slug path string true "Restaurant Slug"
// @Param tableNumber path int true "Table Number"
// @Success 200 {object} map[string]interface{}
// @Router /public/restaurants/{slug}/tables/{tableNumber} [get]
func GetTableBySlugAndNumber(c *gin.Context) {
	slug := c.Param("slug")
	tableNumberStr := c.Param("tableNumber")
	tableNumber, err := strconv.Atoi(tableNumberStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Số bàn không hợp lệ", "INVALID_TABLE_NUMBER", "")
		return
	}

	// Tìm nhà hàng theo slug
	var restaurant models.Restaurant
	if err := config.GetDB().Where("slug = ? AND status = ?", slug, "active").First(&restaurant).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy nhà hàng", "RESTAURANT_NOT_FOUND", "")
		return
	}

	// Tìm bàn theo restaurant_id và table_number
	var table models.Table
	if err := config.GetDB().Where("restaurant_id = ? AND table_number = ? AND is_active = ?", restaurant.ID, tableNumber, true).First(&table).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy bàn", "TABLE_NOT_FOUND", "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":            table.ID,
		"table_number":  table.TableNumber,
		"name":          table.Name,
		"capacity":      table.Capacity,
		"status":        table.Status,
		"restaurant_id": restaurant.ID,
		"restaurant": gin.H{
			"id":          restaurant.ID,
			"name":        restaurant.Name,
			"slug":        restaurant.Slug,
			"logo":        restaurant.Logo,
			"is_open":     restaurant.IsOpen,
			"description": restaurant.Description,
		},
	}, "")
}
