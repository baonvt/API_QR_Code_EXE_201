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
// STATISTICS HANDLERS
// ===============================

// GetStatsOverview thống kê tổng quan
// @Summary Thống kê tổng quan
// @Description Lấy thống kê tổng quan của nhà hàng (hôm nay, tháng này, bàn, đơn hàng)
// @Tags Statistics
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/stats/overview [get]
func GetStatsOverview(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || uint(restaurantID) != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền xem thống kê này", "FORBIDDEN", "")
		return
	}

	db := config.GetDB()
	today := time.Now().Format("2006-01-02")

	// Thống kê hôm nay
	var todayOrders int64
	var todayRevenue float64
	db.Model(&models.Order{}).
		Where("restaurant_id = ? AND DATE(created_at) = ? AND payment_status = ?", restaurantID, today, "paid").
		Count(&todayOrders)
	db.Model(&models.Order{}).
		Where("restaurant_id = ? AND DATE(created_at) = ? AND payment_status = ?", restaurantID, today, "paid").
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&todayRevenue)

	avgTodayOrder := float64(0)
	if todayOrders > 0 {
		avgTodayOrder = todayRevenue / float64(todayOrders)
	}

	// Thống kê tháng này
	var monthOrders int64
	var monthRevenue float64
	db.Model(&models.Order{}).
		Where("restaurant_id = ? AND EXTRACT(MONTH FROM created_at) = EXTRACT(MONTH FROM NOW()) AND EXTRACT(YEAR FROM created_at) = EXTRACT(YEAR FROM NOW()) AND payment_status = ?", restaurantID, "paid").
		Count(&monthOrders)
	db.Model(&models.Order{}).
		Where("restaurant_id = ? AND EXTRACT(MONTH FROM created_at) = EXTRACT(MONTH FROM NOW()) AND EXTRACT(YEAR FROM created_at) = EXTRACT(YEAR FROM NOW()) AND payment_status = ?", restaurantID, "paid").
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&monthRevenue)

	avgMonthOrder := float64(0)
	if monthOrders > 0 {
		avgMonthOrder = monthRevenue / float64(monthOrders)
	}

	// Thống kê bàn
	var totalTables int64
	var availableTables int64
	var occupiedTables int64
	db.Model(&models.Table{}).Where("restaurant_id = ? AND is_active = ?", restaurantID, true).Count(&totalTables)
	db.Model(&models.Table{}).Where("restaurant_id = ? AND is_active = ? AND status = ?", restaurantID, true, "available").Count(&availableTables)
	db.Model(&models.Table{}).Where("restaurant_id = ? AND is_active = ? AND status = ?", restaurantID, true, "occupied").Count(&occupiedTables)

	// Thống kê đơn hàng theo trạng thái
	ordersByStatus := make(map[string]int64)
	statuses := []string{"pending", "confirmed", "preparing", "ready", "serving"}
	for _, status := range statuses {
		var count int64
		db.Model(&models.Order{}).Where("restaurant_id = ? AND status = ?", restaurantID, status).Count(&count)
		ordersByStatus[status] = count
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"today": gin.H{
			"orders":          todayOrders,
			"revenue":         todayRevenue,
			"avg_order_value": avgTodayOrder,
		},
		"this_month": gin.H{
			"orders":          monthOrders,
			"revenue":         monthRevenue,
			"avg_order_value": avgMonthOrder,
		},
		"tables": gin.H{
			"total":     totalTables,
			"available": availableTables,
			"occupied":  occupiedTables,
		},
		"orders_by_status": ordersByStatus,
	}, "")
}

// GetStatsRevenue thống kê doanh thu
// @Summary Thống kê doanh thu
// @Description Lấy thống kê doanh thu theo thời gian
// @Tags Statistics
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Param period query string false "Kỳ thống kê" default(day) Enums(day, week, month)
// @Param start_date query string false "Ngày bắt đầu (YYYY-MM-DD)"
// @Param end_date query string false "Ngày kết thúc (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/stats/revenue [get]
func GetStatsRevenue(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || uint(restaurantID) != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền xem thống kê này", "FORBIDDEN", "")
		return
	}

	db := config.GetDB()

	// Lấy tham số
	period := c.DefaultQuery("period", "day")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Mặc định 30 ngày gần nhất
	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	// Tổng doanh thu và đơn hàng trong khoảng thời gian
	var totalRevenue float64
	var totalOrders int64
	db.Model(&models.Order{}).
		Where("restaurant_id = ? AND DATE(created_at) >= ? AND DATE(created_at) <= ? AND payment_status = ?",
			restaurantID, startDate, endDate, "paid").
		Count(&totalOrders)
	db.Model(&models.Order{}).
		Where("restaurant_id = ? AND DATE(created_at) >= ? AND DATE(created_at) <= ? AND payment_status = ?",
			restaurantID, startDate, endDate, "paid").
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalRevenue)

	avgOrderValue := float64(0)
	if totalOrders > 0 {
		avgOrderValue = totalRevenue / float64(totalOrders)
	}

	// Lấy dữ liệu biểu đồ
	var chartData []gin.H

	if period == "day" {
		var results []struct {
			Date    string
			Revenue float64
			Orders  int64
		}
		db.Model(&models.Order{}).
			Select("DATE(created_at) as date, COALESCE(SUM(total_amount), 0) as revenue, COUNT(*) as orders").
			Where("restaurant_id = ? AND DATE(created_at) >= ? AND DATE(created_at) <= ? AND payment_status = ?",
				restaurantID, startDate, endDate, "paid").
			Group("DATE(created_at)").
			Order("date ASC").
			Scan(&results)

		for _, r := range results {
			chartData = append(chartData, gin.H{
				"date":    r.Date,
				"revenue": r.Revenue,
				"orders":  r.Orders,
			})
		}
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"total_revenue":   totalRevenue,
		"total_orders":    totalOrders,
		"avg_order_value": avgOrderValue,
		"chart_data":      chartData,
	}, "")
}

// GetStatsMenu thống kê món bán chạy
// @Summary Thống kê món bán chạy
// @Description Lấy top món bán chạy và thống kê theo danh mục
// @Tags Statistics
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/stats/menu [get]
func GetStatsMenu(c *gin.Context) {
	restaurantID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || uint(restaurantID) != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền xem thống kê này", "FORBIDDEN", "")
		return
	}

	db := config.GetDB()

	// Top món bán chạy
	var topItems []struct {
		ID           uint
		Name         string
		QuantitySold int64
		Revenue      float64
	}
	db.Model(&models.OrderItem{}).
		Select("menu_items.id, menu_items.name, SUM(order_items.quantity) as quantity_sold, SUM(order_items.line_total) as revenue").
		Joins("JOIN menu_items ON menu_items.id = order_items.menu_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.restaurant_id = ? AND orders.payment_status = ?", restaurantID, "paid").
		Group("menu_items.id, menu_items.name").
		Order("quantity_sold DESC").
		Limit(10).
		Scan(&topItems)

	var topItemsData []gin.H
	for _, item := range topItems {
		topItemsData = append(topItemsData, gin.H{
			"id":            item.ID,
			"name":          item.Name,
			"quantity_sold": item.QuantitySold,
			"revenue":       item.Revenue,
		})
	}

	// Thống kê theo danh mục
	var byCategory []struct {
		Category     string
		QuantitySold int64
		Revenue      float64
	}
	db.Model(&models.OrderItem{}).
		Select("categories.name as category, SUM(order_items.quantity) as quantity_sold, SUM(order_items.line_total) as revenue").
		Joins("JOIN menu_items ON menu_items.id = order_items.menu_item_id").
		Joins("JOIN categories ON categories.id = menu_items.category_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.restaurant_id = ? AND orders.payment_status = ?", restaurantID, "paid").
		Group("categories.name").
		Order("revenue DESC").
		Scan(&byCategory)

	var byCategoryData []gin.H
	for _, cat := range byCategory {
		byCategoryData = append(byCategoryData, gin.H{
			"category":      cat.Category,
			"quantity_sold": cat.QuantitySold,
			"revenue":       cat.Revenue,
		})
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"top_items":   topItemsData,
		"by_category": byCategoryData,
	}, "")
}
