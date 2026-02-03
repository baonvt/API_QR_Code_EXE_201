package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-api/config"
	"go-api/models"
	"go-api/services"
	"go-api/utils"

	"github.com/gin-gonic/gin"
)

// ===============================
// REQUEST STRUCTS
// ===============================

// OrderItemInput item trong đơn hàng
type OrderItemInput struct {
	MenuItemID      uint   `json:"menu_item_id" binding:"required"`
	Quantity        int    `json:"quantity" binding:"required,min=1"`
	SelectedOptions string `json:"selected_options"` // JSON string
	Notes           string `json:"notes"`
}

// CreateOrderInput request body cho tạo đơn hàng
// Khách order = thanh toán luôn
type CreateOrderInput struct {
	TableNumber   int              `json:"table_number" binding:"required"`
	PaymentMethod string           `json:"payment_method" binding:"required"` // cash, qr, momo, vnpay
	CustomerName  string           `json:"customer_name"`
	CustomerPhone string           `json:"customer_phone"`
	Notes         string           `json:"notes"`
	Items         []OrderItemInput `json:"items" binding:"required,min=1"`
}

// UpdateOrderStatusInput request body cho cập nhật trạng thái
type UpdateOrderStatusInput struct {
	Status string `json:"status" binding:"required"`
	Note   string `json:"note"`
}

// PayOrderInput request body cho thanh toán
type PayOrderInput struct {
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// AddOrderItemsInput request body cho thêm món
type AddOrderItemsInput struct {
	Items []OrderItemInput `json:"items" binding:"required,min=1"`
}

// ===============================
// HANDLERS
// ===============================

// GetOrders lấy danh sách đơn hàng của nhà hàng
// @Summary Lấy danh sách đơn hàng
// @Description Lấy danh sách đơn hàng của nhà hàng với phân trang và filter
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Param status query string false "Filter theo status" Enums(pending, confirmed, serving, completed, cancelled)
// @Param date query string false "Filter theo ngày (YYYY-MM-DD)"
// @Param table_id query int false "Filter theo bàn"
// @Param page query int false "Trang" default(1)
// @Param limit query int false "Số lượng/trang" default(20)
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/orders [get]
func GetOrders(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || uint(restaurantID) != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền xem đơn hàng của nhà hàng này", "FORBIDDEN", "")
		return
	}

	query := config.GetDB().Preload("Table").Where("restaurant_id = ?", restaurantID)

	// Filter by status
	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by date
	date := c.Query("date")
	if date != "" {
		query = query.Where("DATE(created_at) = ?", date)
	}

	// Filter by table
	tableID := c.Query("table_id")
	if tableID != "" {
		query = query.Where("table_id = ?", tableID)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Order{}).Count(&total)

	var orders []models.Order
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Lỗi khi lấy danh sách đơn hàng", "QUERY_ERROR", err.Error())
		return
	}

	var data []gin.H
	for _, order := range orders {
		// Đếm số món
		var itemsCount int64
		config.GetDB().Model(&models.OrderItem{}).Where("order_id = ?", order.ID).Count(&itemsCount)

		tableName := ""
		tableNumber := 0
		if order.Table != nil {
			if order.Table.Name != nil {
				tableName = *order.Table.Name
			}
			tableNumber = order.Table.TableNumber
		}

		data = append(data, gin.H{
			"id":             order.ID,
			"order_number":   order.OrderNumber,
			"table_id":       order.TableID,
			"table_number":   tableNumber,
			"table_name":     tableName,
			"status":         order.Status,
			"payment_status": order.PaymentStatus,
			"payment_timing": order.PaymentTiming,
			"subtotal":       order.Subtotal,
			"tax_amount":     order.TaxAmount,
			"service_charge": order.ServiceCharge,
			"total_amount":   order.TotalAmount,
			"items_count":    itemsCount,
			"created_at":     order.CreatedAt,
		})
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"orders": data,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		},
	})
}

// GetOrder lấy chi tiết đơn hàng
// @Summary Lấy chi tiết đơn hàng
// @Description Lấy thông tin chi tiết của một đơn hàng (Public - cho tracking)
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Router /orders/{id} [get]
func GetOrder(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var order models.Order
	if err := config.GetDB().Preload("Table").Preload("OrderItems").Preload("OrderItems.MenuItem").First(&order, orderID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy đơn hàng", "ORDER_NOT_FOUND", "")
		return
	}

	// Build items response
	var items []gin.H
	for _, item := range order.OrderItems {
		items = append(items, gin.H{
			"id":               item.ID,
			"menu_item_id":     item.MenuItemID,
			"name":             item.ItemName,
			"price":            item.ItemPrice,
			"quantity":         item.Quantity,
			"selected_options": item.SelectedOptions,
			"notes":            item.Notes,
			"prep_status":      item.PrepStatus,
			"line_total":       item.LineTotal,
		})
	}

	tableName := ""
	if order.Table != nil && order.Table.Name != nil {
		tableName = *order.Table.Name
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":              order.ID,
		"order_number":    order.OrderNumber,
		"table_id":        order.TableID,
		"table_name":      tableName,
		"customer_name":   order.CustomerName,
		"customer_phone":  order.CustomerPhone,
		"status":          order.Status,
		"payment_status":  order.PaymentStatus,
		"payment_timing":  order.PaymentTiming,
		"payment_method":  order.PaymentMethod,
		"subtotal":        order.Subtotal,
		"tax_amount":      order.TaxAmount,
		"service_charge":  order.ServiceCharge,
		"discount_amount": order.DiscountAmount,
		"total_amount":    order.TotalAmount,
		"notes":           order.Notes,
		"items":           items,
		"created_at":      order.CreatedAt,
		"updated_at":      order.UpdatedAt,
	}, "")
}

// TrackOrder - Khách hàng theo dõi đơn hàng qua order number
// @Summary Theo dõi đơn hàng
// @Description Khách hàng xem trạng thái đơn hàng và thông tin thanh toán
// @Tags Public
// @Accept json
// @Produce json
// @Param orderNumber path string true "Mã đơn hàng"
// @Success 200 {object} map[string]interface{}
// @Router /public/orders/{orderNumber}/track [get]
func TrackOrder(c *gin.Context) {
	orderNumber := c.Param("orderNumber")

	var order models.Order
	if err := config.GetDB().
		Preload("Table").
		Preload("OrderItems").
		Preload("OrderItems.MenuItem").
		Where("order_number = ?", orderNumber).
		First(&order).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy đơn hàng", "ORDER_NOT_FOUND", "")
		return
	}

	// Lấy thông tin nhà hàng để lấy QR thanh toán
	var restaurant models.Restaurant
	config.GetDB().First(&restaurant, order.RestaurantID)

	// Build items response
	var items []gin.H
	for _, item := range order.OrderItems {
		items = append(items, gin.H{
			"id":       item.ID,
			"name":     item.ItemName,
			"price":    item.ItemPrice,
			"quantity": item.Quantity,
			"notes":    item.Notes,
		})
	}

	tableName := ""
	tableNumber := 0
	if order.Table != nil {
		if order.Table.Name != nil {
			tableName = *order.Table.Name
		}
		tableNumber = order.Table.TableNumber
	}

	// Response
	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":             order.ID,
		"order_number":   order.OrderNumber,
		"status":         order.Status,
		"payment_status": order.PaymentStatus,
		"payment_method": order.PaymentMethod,
		"total_amount":   order.TotalAmount,
		"table_name":     tableName,
		"table_number":   tableNumber,
		"items":          items,
		"created_at":     order.CreatedAt,
		"restaurant": gin.H{
			"name": restaurant.Name,
			"slug": restaurant.Slug,
		},
	}, "")
}

// CreateOrder tạo đơn hàng mới (Customer - Public)
// @Summary Tạo đơn hàng mới
// @Description Khách hàng tạo đơn và thanh toán luôn
// @Tags Public
// @Accept json
// @Produce json
// @Param slug path string true "Restaurant Slug"
// @Param order body CreateOrderInput true "Thông tin đơn hàng"
// @Success 201 {object} map[string]interface{}
// @Router /public/restaurants/{slug}/orders [post]
func CreateOrder(c *gin.Context) {
	slug := c.Param("slug")

	// Tìm nhà hàng
	var restaurant models.Restaurant
	if err := config.GetDB().Where("slug = ? AND status = ?", slug, "active").First(&restaurant).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy nhà hàng", "RESTAURANT_NOT_FOUND", "")
		return
	}

	// Kiểm tra nhà hàng có mở cửa không
	if !restaurant.IsOpen {
		utils.ErrorResponse(c, http.StatusBadRequest, "Nhà hàng hiện đang đóng cửa", "RESTAURANT_CLOSED", "")
		return
	}

	var input CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	// Tìm bàn theo table_number
	var table models.Table
	if err := config.GetDB().Where("restaurant_id = ? AND table_number = ? AND is_active = ?", restaurant.ID, input.TableNumber, true).First(&table).Error; err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bàn không tồn tại", "TABLE_NOT_FOUND", "")
		return
	}

	db := config.GetDB()
	tx := db.Begin()

	// Tạo order number
	var orderCount int64
	tx.Model(&models.Order{}).Where("restaurant_id = ?", restaurant.ID).Count(&orderCount)
	orderNumber := fmt.Sprintf("ORD-%d-%04d", time.Now().Year(), orderCount+1)

	// Validate payment method
	validMethods := []string{"cash", "qr", "momo", "vnpay"}
	isValidPayment := false
	for _, m := range validMethods {
		if m == input.PaymentMethod {
			isValidPayment = true
			break
		}
	}
	if !isValidPayment {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusBadRequest, "Phương thức thanh toán không hợp lệ", "INVALID_PAYMENT_METHOD", "")
		return
	}

	// Tạo order - Khách order = chờ thanh toán (LUỒNG MỚI)
	// Status: pending (chờ xác nhận)
	// Payment: unpaid (chưa thanh toán)
	// Bàn: KHÔNG chiếm (vẫn available cho đến khi nhà hàng xác nhận thanh toán)
	paymentMethod := input.PaymentMethod
	order := models.Order{
		RestaurantID:  restaurant.ID,
		TableID:       table.ID,
		OrderNumber:   orderNumber,
		CustomerName:  &input.CustomerName,
		CustomerPhone: &input.CustomerPhone,
		Status:        "pending", // Chờ xác nhận
		PaymentTiming: "before",  // Thanh toán trước
		PaymentStatus: "unpaid",  // Chưa thanh toán
		PaymentMethod: &paymentMethod,
		Notes:         &input.Notes,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo đơn hàng", "CREATE_ERROR", err.Error())
		return
	}

	// Tạo order items
	var subtotal float64 = 0
	for _, itemInput := range input.Items {
		// Lấy thông tin món
		var menuItem models.MenuItem
		if err := tx.Where("id = ? AND restaurant_id = ? AND status = ?", itemInput.MenuItemID, restaurant.ID, "active").First(&menuItem).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusBadRequest, "Món không tồn tại hoặc đã ngừng bán", "INVALID_MENU_ITEM", "")
			return
		}

		lineTotal := menuItem.Price * float64(itemInput.Quantity)
		subtotal += lineTotal

		orderItem := models.OrderItem{
			OrderID:         order.ID,
			MenuItemID:      menuItem.ID,
			ItemName:        menuItem.Name,
			ItemPrice:       menuItem.Price,
			Quantity:        itemInput.Quantity,
			SelectedOptions: &itemInput.SelectedOptions,
			Notes:           &itemInput.Notes,
			PrepStatus:      "pending", // Chờ xác nhận
			PrepLocation:    menuItem.PrepLocation,
			LineTotal:       lineTotal,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo chi tiết đơn hàng", "CREATE_ITEM_ERROR", err.Error())
			return
		}
	}

	// Tính toán tổng tiền
	taxAmount := subtotal * restaurant.TaxRate / 100
	serviceCharge := subtotal * restaurant.ServiceCharge / 100
	totalAmount := subtotal + taxAmount + serviceCharge

	// Cập nhật order với tổng tiền
	tx.Model(&order).Updates(map[string]interface{}{
		"subtotal":       subtotal,
		"tax_amount":     taxAmount,
		"service_charge": serviceCharge,
		"total_amount":   totalAmount,
	})

	// KHÔNG cập nhật trạng thái bàn - bàn vẫn trống cho đến khi xác nhận thanh toán
	// tx.Model(&table).Update("status", "occupied") -- BỎ DÒNG NÀY

	tx.Commit()

	// Tạo thông báo cho nhà hàng
	tableName := "Bàn " + strconv.Itoa(table.TableNumber)
	if table.Name != nil {
		tableName = *table.Name
	}
	CreateOrderNotification(restaurant.ID, order.ID, orderNumber, tableName, totalAmount)

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"id":             order.ID,
		"order_number":   orderNumber,
		"status":         order.Status,
		"payment_status": order.PaymentStatus,
		"total_amount":   totalAmount,
		"tracking_url":   "/" + slug + "/order/" + strconv.Itoa(int(order.ID)),
		"message":        "Vui lòng thanh toán để hoàn tất đơn hàng",
	}, "Đơn hàng đã được tạo. Vui lòng thanh toán!")
}

// UpdateOrderStatus cập nhật trạng thái đơn hàng
// @Summary Cập nhật trạng thái đơn hàng
// @Description Nhà hàng cập nhật trạng thái: confirmed -> serving -> completed
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param status body UpdateOrderStatusInput true "Trạng thái mới"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /orders/{id}/status [put]
func UpdateOrderStatus(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var order models.Order
	if err := config.GetDB().First(&order, orderID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy đơn hàng", "ORDER_NOT_FOUND", "")
		return
	}

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || order.RestaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền cập nhật đơn hàng này", "FORBIDDEN", "")
		return
	}

	var input UpdateOrderStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	// Validate status transitions - Đơn giản hóa: không có bếp
	// Flow: confirmed -> serving -> completed
	validTransitions := map[string][]string{
		"pending":   {"confirmed", "cancelled"},
		"confirmed": {"serving", "cancelled"},
		"serving":   {"completed", "cancelled"},
	}

	allowed, exists := validTransitions[order.Status]
	if !exists {
		utils.ErrorResponse(c, http.StatusBadRequest, "Không thể thay đổi trạng thái từ trạng thái hiện tại", "INVALID_TRANSITION", "")
		return
	}

	isValid := false
	for _, s := range allowed {
		if s == input.Status {
			isValid = true
			break
		}
	}

	if !isValid {
		utils.ErrorResponse(c, http.StatusBadRequest, "Chuyển trạng thái không hợp lệ", "INVALID_TRANSITION", fmt.Sprintf("Cannot transition from %s to %s", order.Status, input.Status))
		return
	}

	updates := map[string]interface{}{
		"status": input.Status,
	}

	if input.Status == "completed" {
		now := time.Now()
		updates["completed_at"] = now
	}

	if input.Status == "cancelled" && input.Note != "" {
		updates["cancel_reason"] = input.Note
	}

	if err := config.GetDB().Model(&order).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể cập nhật trạng thái", "UPDATE_ERROR", err.Error())
		return
	}

	// Nếu completed hoặc cancelled, giải phóng bàn
	if input.Status == "completed" || input.Status == "cancelled" {
		config.GetDB().Model(&models.Table{}).Where("id = ?", order.TableID).Update("status", "available")
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":     order.ID,
		"status": input.Status,
	}, "Cập nhật trạng thái thành công")
}

// PayOrder thanh toán đơn hàng
// @Summary Thanh toán đơn hàng
// @Description Xác nhận thanh toán cho đơn hàng (dùng khi khách chưa thanh toán trước)
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param payment body PayOrderInput true "Phương thức thanh toán"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /orders/{id}/pay [put]
func PayOrder(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var order models.Order
	if err := config.GetDB().First(&order, orderID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy đơn hàng", "ORDER_NOT_FOUND", "")
		return
	}

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || order.RestaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền thanh toán đơn hàng này", "FORBIDDEN", "")
		return
	}

	if order.PaymentStatus == "paid" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Đơn hàng đã được thanh toán", "ALREADY_PAID", "")
		return
	}

	var input PayOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	// Validate payment method
	validMethods := []string{"cash", "qr", "momo", "vnpay"}
	isValid := false
	for _, m := range validMethods {
		if m == input.PaymentMethod {
			isValid = true
			break
		}
	}
	if !isValid {
		utils.ErrorResponse(c, http.StatusBadRequest, "Phương thức thanh toán không hợp lệ", "INVALID_PAYMENT_METHOD", "")
		return
	}

	now := time.Now()
	if err := config.GetDB().Model(&order).Updates(map[string]interface{}{
		"payment_status": "paid",
		"payment_method": input.PaymentMethod,
		"paid_at":        now,
	}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể cập nhật thanh toán", "UPDATE_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"order_id":       order.ID,
		"payment_status": "paid",
		"payment_method": input.PaymentMethod,
		"paid_at":        now,
		"total_amount":   order.TotalAmount,
	}, "Thanh toán thành công!")
}

// ConfirmOrderPayment xác nhận thanh toán đơn hàng (nhà hàng bấm tích)
// @Summary Xác nhận đã nhận thanh toán
// @Description Nhà hàng xác nhận đã nhận được tiền từ khách, đơn chuyển sang confirmed và bàn được chiếm
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /orders/{id}/confirm-payment [put]
func ConfirmOrderPayment(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var order models.Order
	if err := config.GetDB().Preload("Table").First(&order, orderID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy đơn hàng", "ORDER_NOT_FOUND", "")
		return
	}

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || order.RestaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền xác nhận đơn hàng này", "FORBIDDEN", "")
		return
	}

	if order.PaymentStatus == "paid" && order.Status != "pending" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Đơn hàng đã được xác nhận thanh toán", "ALREADY_CONFIRMED", "")
		return
	}

	db := config.GetDB()
	tx := db.Begin()

	now := time.Now()

	// Cập nhật order: payment_status = paid, status = confirmed
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"payment_status": "paid",
		"status":         "confirmed",
		"paid_at":        now,
	}).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể xác nhận thanh toán", "UPDATE_ERROR", err.Error())
		return
	}

	// Cập nhật trạng thái order items
	tx.Model(&models.OrderItem{}).Where("order_id = ?", order.ID).Update("prep_status", "confirmed")

	// Cập nhật trạng thái bàn thành occupied
	if order.Table != nil {
		tx.Model(&models.Table{}).Where("id = ?", order.TableID).Update("status", "occupied")
	}

	tx.Commit()

	// Tạo thông báo thành công
	CreateSystemNotification(order.RestaurantID, "system_success", "Thanh toán đã xác nhận", "Đơn #"+order.OrderNumber+" đã được xác nhận thanh toán")

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"order_id":       order.ID,
		"order_number":   order.OrderNumber,
		"status":         "confirmed",
		"payment_status": "paid",
		"paid_at":        now,
		"total_amount":   order.TotalAmount,
		"table_status":   "occupied",
	}, "Xác nhận thanh toán thành công! Đơn hàng đã được xác nhận.")
}

// AddOrderItems thêm món vào đơn hàng hiện tại
// @Summary Thêm món vào đơn hàng
// @Description Khách thêm món vào đơn hàng đang phục vụ (Public)
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param items body AddOrderItemsInput true "Danh sách món thêm"
// @Success 200 {object} map[string]interface{}
// @Router /orders/{id}/items [post]
func AddOrderItems(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var order models.Order
	if err := config.GetDB().First(&order, orderID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy đơn hàng", "ORDER_NOT_FOUND", "")
		return
	}

	// Kiểm tra đơn hàng còn có thể thêm món không
	if order.Status == "completed" || order.Status == "cancelled" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Không thể thêm món vào đơn hàng đã hoàn thành hoặc đã hủy", "ORDER_CLOSED", "")
		return
	}

	var input AddOrderItemsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	// Lấy thông tin nhà hàng
	var restaurant models.Restaurant
	config.GetDB().First(&restaurant, order.RestaurantID)

	db := config.GetDB()
	tx := db.Begin()

	var addedSubtotal float64 = 0
	for _, itemInput := range input.Items {
		var menuItem models.MenuItem
		if err := tx.Where("id = ? AND restaurant_id = ? AND status = ?", itemInput.MenuItemID, order.RestaurantID, "active").First(&menuItem).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusBadRequest, "Món không tồn tại hoặc đã ngừng bán", "INVALID_MENU_ITEM", "")
			return
		}

		lineTotal := menuItem.Price * float64(itemInput.Quantity)
		addedSubtotal += lineTotal

		orderItem := models.OrderItem{
			OrderID:         order.ID,
			MenuItemID:      menuItem.ID,
			ItemName:        menuItem.Name,
			ItemPrice:       menuItem.Price,
			Quantity:        itemInput.Quantity,
			SelectedOptions: &itemInput.SelectedOptions,
			Notes:           &itemInput.Notes,
			PrepStatus:      "confirmed", // Không cần bếp - xác nhận luôn
			PrepLocation:    "service",   // Phục vụ trực tiếp
			LineTotal:       lineTotal,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể thêm món", "CREATE_ITEM_ERROR", err.Error())
			return
		}
	}

	// Cập nhật tổng tiền
	newSubtotal := order.Subtotal + addedSubtotal
	taxAmount := newSubtotal * restaurant.TaxRate / 100
	serviceCharge := newSubtotal * restaurant.ServiceCharge / 100
	totalAmount := newSubtotal + taxAmount + serviceCharge

	tx.Model(&order).Updates(map[string]interface{}{
		"subtotal":       newSubtotal,
		"tax_amount":     taxAmount,
		"service_charge": serviceCharge,
		"total_amount":   totalAmount,
	})

	tx.Commit()

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"order_id":     order.ID,
		"total_amount": totalAmount,
		"items_added":  len(input.Items),
	}, "Thêm món thành công!")
}

// GetOrderBill lấy thông tin in bill
// @Summary Lấy thông tin hóa đơn
// @Description Lấy thông tin để in bill cho đơn hàng
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /orders/{id}/bill [get]
func GetOrderBill(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var order models.Order
	if err := config.GetDB().Preload("Table").Preload("OrderItems").Preload("Restaurant").First(&order, orderID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy đơn hàng", "ORDER_NOT_FOUND", "")
		return
	}

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || order.RestaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền xem bill này", "FORBIDDEN", "")
		return
	}

	// Build items
	var items []gin.H
	for _, item := range order.OrderItems {
		items = append(items, gin.H{
			"name":     item.ItemName,
			"quantity": item.Quantity,
			"price":    item.ItemPrice,
			"total":    item.LineTotal,
		})
	}

	tableName := ""
	if order.Table != nil && order.Table.Name != nil {
		tableName = *order.Table.Name
	}

	restaurantAddress := ""
	restaurantPhone := ""
	if order.Restaurant != nil {
		if order.Restaurant.Address != nil {
			restaurantAddress = *order.Restaurant.Address
		}
		if order.Restaurant.Phone != nil {
			restaurantPhone = *order.Restaurant.Phone
		}
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"restaurant": gin.H{
			"name":    order.Restaurant.Name,
			"address": restaurantAddress,
			"phone":   restaurantPhone,
		},
		"order": gin.H{
			"order_number": order.OrderNumber,
			"table_name":   tableName,
			"created_at":   order.CreatedAt,
			"completed_at": order.CompletedAt,
		},
		"items": items,
		"summary": gin.H{
			"subtotal":        order.Subtotal,
			"tax_amount":      order.TaxAmount,
			"service_charge":  order.ServiceCharge,
			"discount_amount": order.DiscountAmount,
			"total_amount":    order.TotalAmount,
		},
		"payment": gin.H{
			"method":  order.PaymentMethod,
			"status":  order.PaymentStatus,
			"paid_at": order.PaidAt,
		},
	}, "")
}

// GetOrderPaymentQR - Lấy VietQR code cho thanh toán đơn hàng
// @Summary Lấy QR thanh toán
// @Description Tạo VietQR code để khách thanh toán đơn hàng
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Router /orders/{id}/payment-qr [get]
func GetOrderPaymentQR(c *gin.Context) {
	orderID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var order models.Order
	if err := config.GetDB().
		Preload("Restaurant").
		Preload("Restaurant.PaymentSetting").
		First(&order, orderID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy đơn hàng", "ORDER_NOT_FOUND", "")
		return
	}

	// Check if order is already paid
	if order.PaymentStatus == "paid" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Đơn hàng đã được thanh toán", "ALREADY_PAID", "")
		return
	}

	// Get payment settings
	paymentSetting := order.Restaurant.PaymentSetting
	if paymentSetting == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Nhà hàng chưa cấu hình thông tin thanh toán", "PAYMENT_SETTINGS_NOT_FOUND", "")
		return
	}

	// Validate bank info
	if paymentSetting.BankCode == nil || *paymentSetting.BankCode == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Chưa cấu hình mã ngân hàng", "BANK_CODE_MISSING", "")
		return
	}
	if paymentSetting.AccountNumber == nil || *paymentSetting.AccountNumber == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Chưa cấu hình số tài khoản", "ACCOUNT_NUMBER_MISSING", "")
		return
	}

	accountName := "Nha Hang"
	if paymentSetting.AccountName != nil {
		accountName = *paymentSetting.AccountName
	}

	// Generate description
	description := fmt.Sprintf("Thanh toan don %s", order.OrderNumber)

	// Generate VietQR URL
	qrURL, err := services.GenerateVietQRURL(
		*paymentSetting.BankCode,
		*paymentSetting.AccountNumber,
		accountName,
		order.TotalAmount,
		description,
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể tạo mã QR", "QR_GENERATION_ERROR", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"order_number":   order.OrderNumber,
		"total_amount":   order.TotalAmount,
		"qr_url":         qrURL,
		"bank_name":      paymentSetting.BankName,
		"bank_code":      paymentSetting.BankCode,
		"account_number": paymentSetting.AccountNumber,
		"account_name":   paymentSetting.AccountName,
		"description":    description,
	}, "")
}
