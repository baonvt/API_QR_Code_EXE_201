package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go-api/config"
	"go-api/models"
	"go-api/utils"

	"github.com/gin-gonic/gin"
)

// ===============================
// NOTIFICATION HANDLERS
// ===============================

// GetNotifications lấy danh sách thông báo của nhà hàng
// @Summary Lấy danh sách thông báo
// @Description Lấy tất cả thông báo của nhà hàng với phân trang
// @Tags Notifications
// @Accept json
// @Produce json
// @Param unread_only query bool false "Chỉ lấy thông báo chưa đọc"
// @Param limit query int false "Số lượng" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/notifications [get]
func GetNotifications(c *gin.Context) {
	restaurantID, _ := c.Get("restaurant_id")
	if restaurantID == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Không có quyền truy cập", "FORBIDDEN", "")
		return
	}

	rid := *restaurantID.(*uint)

	// Parse query params
	unreadOnly := c.Query("unread_only") == "true"
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 50 {
		limit = 50
	}

	db := config.GetDB()
	query := db.Where("restaurant_id = ?", rid)

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	// Đếm tổng
	var total int64
	query.Model(&models.Notification{}).Count(&total)

	// Đếm chưa đọc
	var unreadCount int64
	db.Model(&models.Notification{}).Where("restaurant_id = ? AND is_read = ?", rid, false).Count(&unreadCount)

	// Lấy danh sách
	var notifications []models.Notification
	query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&notifications)

	var data []gin.H
	for _, n := range notifications {
		item := gin.H{
			"id":         n.ID,
			"type":       n.Type,
			"title":      n.Title,
			"message":    n.Message,
			"is_read":    n.IsRead,
			"read_at":    n.ReadAt,
			"created_at": n.CreatedAt,
		}

		// Parse data JSON nếu có
		if n.Data != nil {
			var dataObj map[string]interface{}
			if err := json.Unmarshal([]byte(*n.Data), &dataObj); err == nil {
				item["data"] = dataObj
			}
		}

		data = append(data, item)
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"notifications": data,
		"total":         total,
		"unread_count":  unreadCount,
		"limit":         limit,
		"offset":        offset,
	}, "")
}

// GetUnreadNotificationCount lấy số lượng thông báo chưa đọc
// @Summary Lấy số lượng thông báo chưa đọc
// @Description Lấy nhanh số lượng thông báo chưa đọc (dùng cho badge)
// @Tags Notifications
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/notifications/unread-count [get]
func GetUnreadNotificationCount(c *gin.Context) {
	restaurantID, _ := c.Get("restaurant_id")
	if restaurantID == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Không có quyền truy cập", "FORBIDDEN", "")
		return
	}

	rid := *restaurantID.(*uint)

	var unreadCount int64
	config.GetDB().Model(&models.Notification{}).Where("restaurant_id = ? AND is_read = ?", rid, false).Count(&unreadCount)

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"unread_count": unreadCount,
	}, "")
}

// MarkNotificationRead đánh dấu thông báo đã đọc
// @Summary Đánh dấu đã đọc
// @Description Đánh dấu một thông báo là đã đọc
// @Tags Notifications
// @Accept json
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /notifications/{id}/read [put]
func MarkNotificationRead(c *gin.Context) {
	notificationID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	restaurantID, _ := c.Get("restaurant_id")
	if restaurantID == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Không có quyền truy cập", "FORBIDDEN", "")
		return
	}

	rid := *restaurantID.(*uint)

	var notification models.Notification
	if err := config.GetDB().Where("id = ? AND restaurant_id = ?", notificationID, rid).First(&notification).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy thông báo", "NOT_FOUND", "")
		return
	}

	now := time.Now()
	config.GetDB().Model(&notification).Updates(map[string]interface{}{
		"is_read": true,
		"read_at": now,
	})

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"id":      notification.ID,
		"is_read": true,
		"read_at": now,
	}, "Đã đánh dấu đã đọc")
}

// MarkAllNotificationsRead đánh dấu tất cả thông báo đã đọc
// @Summary Đánh dấu tất cả đã đọc
// @Description Đánh dấu tất cả thông báo của nhà hàng là đã đọc
// @Tags Notifications
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /notifications/read-all [put]
func MarkAllNotificationsRead(c *gin.Context) {
	restaurantID, _ := c.Get("restaurant_id")
	if restaurantID == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Không có quyền truy cập", "FORBIDDEN", "")
		return
	}

	rid := *restaurantID.(*uint)
	now := time.Now()

	result := config.GetDB().Model(&models.Notification{}).
		Where("restaurant_id = ? AND is_read = ?", rid, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"updated_count": result.RowsAffected,
	}, "Đã đánh dấu tất cả đã đọc")
}

// DeleteNotification xóa thông báo
// @Summary Xóa thông báo
// @Description Xóa một thông báo
// @Tags Notifications
// @Accept json
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /notifications/{id} [delete]
func DeleteNotification(c *gin.Context) {
	notificationID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	restaurantID, _ := c.Get("restaurant_id")
	if restaurantID == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Không có quyền truy cập", "FORBIDDEN", "")
		return
	}

	rid := *restaurantID.(*uint)

	result := config.GetDB().Where("id = ? AND restaurant_id = ?", notificationID, rid).Delete(&models.Notification{})
	if result.RowsAffected == 0 {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy thông báo", "NOT_FOUND", "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, nil, "Đã xóa thông báo")
}

// ClearAllNotifications xóa tất cả thông báo
// @Summary Xóa tất cả thông báo
// @Description Xóa tất cả thông báo của nhà hàng
// @Tags Notifications
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /notifications/clear [delete]
func ClearAllNotifications(c *gin.Context) {
	restaurantID, _ := c.Get("restaurant_id")
	if restaurantID == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Không có quyền truy cập", "FORBIDDEN", "")
		return
	}

	rid := *restaurantID.(*uint)

	result := config.GetDB().Where("restaurant_id = ?", rid).Delete(&models.Notification{})

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"deleted_count": result.RowsAffected,
	}, "Đã xóa tất cả thông báo")
}

// ===============================
// HELPER FUNCTIONS
// ===============================

// CreateNotification tạo thông báo mới (internal helper)
func CreateNotification(restaurantID uint, notifType, title, message string, data map[string]interface{}) error {
	var dataStr *string
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err == nil {
			s := string(jsonData)
			dataStr = &s
		}
	}

	notification := models.Notification{
		RestaurantID: restaurantID,
		Type:         notifType,
		Title:        title,
		Message:      message,
		Data:         dataStr,
		IsRead:       false,
	}

	return config.GetDB().Create(&notification).Error
}

// CreateOrderNotification tạo thông báo đơn hàng mới
func CreateOrderNotification(restaurantID uint, orderID uint, orderNumber string, tableName string, totalAmount float64) error {
	return CreateNotification(
		restaurantID,
		"new_order",
		"Đơn hàng mới #"+orderNumber,
		tableName+" • "+formatCurrency(totalAmount),
		map[string]interface{}{
			"order_id":     orderID,
			"order_number": orderNumber,
			"table_name":   tableName,
			"total_amount": totalAmount,
		},
	)
}

// CreatePaymentPendingNotification tạo thông báo chờ thanh toán
func CreatePaymentPendingNotification(restaurantID uint, orderID uint, orderNumber string, totalAmount float64) error {
	return CreateNotification(
		restaurantID,
		"payment_pending",
		"Chờ xác nhận thanh toán #"+orderNumber,
		"Số tiền: "+formatCurrency(totalAmount),
		map[string]interface{}{
			"order_id":     orderID,
			"order_number": orderNumber,
			"total_amount": totalAmount,
		},
	)
}

// CreateSystemNotification tạo thông báo hệ thống
func CreateSystemNotification(restaurantID uint, notifType, title, message string) error {
	return CreateNotification(restaurantID, notifType, title, message, nil)
}

// formatCurrency format tiền VND
func formatCurrency(amount float64) string {
	// Simple formatting - can be improved
	return strconv.FormatFloat(amount, 'f', 0, 64) + "đ"
}
