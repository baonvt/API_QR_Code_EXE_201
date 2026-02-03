package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go-api/config"
	"go-api/models"
	"go-api/services"
	"go-api/utils"

	"github.com/gin-gonic/gin"
)

// ===============================
// SEPAY WEBHOOK HANDLER
// ===============================

// SepayWebhookPayload payload từ SePay webhook
type SepayWebhookPayload struct {
	ID                 int64   `json:"id"`
	Gateway            string  `json:"gateway"`
	TransactionDate    string  `json:"transactionDate"`
	AccountNumber      string  `json:"accountNumber"`
	SubAccount         *string `json:"subAccount"`
	TransferType       string  `json:"transferType"`
	TransferAmount     float64 `json:"transferAmount"`
	Accumulated        float64 `json:"accumulated"`
	Code               *string `json:"code"`
	TransactionContent string  `json:"content"`       // SePay gửi "content" không phải "transactionContent"
	ReferenceNumber    string  `json:"referenceCode"` // SePay gửi "referenceCode" không phải "referenceNumber"
	Description        string  `json:"description"`
}

// HandleSepayWebhook xử lý webhook từ SePay
// @Summary Webhook SePay
// @Description Nhận thông báo giao dịch từ SePay
// @Tags Webhooks
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /webhooks/sepay [post]
func HandleSepayWebhook(c *gin.Context) {
	var payload SepayWebhookPayload

	// Parse JSON payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("❌ SePay webhook parse error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid payload"})
		return
	}

	// Log webhook nhận được
	rawJSON, _ := json.Marshal(payload)
	log.Printf("📥 SePay Webhook received: %s", string(rawJSON))

	// Chỉ xử lý giao dịch tiền VÀO
	if payload.TransferType != "in" {
		log.Printf("⏭️ Skipping outgoing transaction")
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Skipped outgoing transaction"})
		return
	}

	// Kiểm tra giao dịch đã xử lý chưa (tránh duplicate)
	db := config.GetDB()
	var existingTx models.PaymentTransaction
	if err := db.Where("sepay_transaction_id = ?", payload.ID).First(&existingTx).Error; err == nil {
		log.Printf("⏭️ Transaction already processed: %d", payload.ID)
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Already processed"})
		return
	}

	// Parse payment code từ nội dung chuyển khoản
	transactionType, code, found := services.ParsePaymentCode(payload.TransactionContent)
	if !found {
		log.Printf("⚠️ No payment code found in: %s", payload.TransactionContent)
		// Vẫn lưu transaction để tracking
		saveUnmatchedTransaction(&payload)
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "No payment code found"})
		return
	}

	log.Printf("🔍 Found payment code: type=%s, code=%s", transactionType, code)

	// Chuyển đổi payload sang service format
	servicePayload := &services.SepayWebhookPayload{
		ID:                 payload.ID,
		Gateway:            payload.Gateway,
		TransactionDate:    payload.TransactionDate,
		AccountNumber:      payload.AccountNumber,
		SubAccount:         payload.SubAccount,
		TransferType:       payload.TransferType,
		TransferAmount:     payload.TransferAmount,
		Accumulated:        payload.Accumulated,
		Code:               payload.Code,
		TransactionContent: payload.TransactionContent,
		ReferenceNumber:    payload.ReferenceNumber,
		Description:        payload.Description,
	}

	// Xử lý theo loại giao dịch
	var err error
	switch transactionType {
	case "package":
		err = handlePackagePayment(code, servicePayload)
	case "order":
		err = handleOrderPayment(code, servicePayload)
	default:
		log.Printf("⚠️ Unknown transaction type: %s", transactionType)
	}

	if err != nil {
		log.Printf("❌ Payment processing error: %v", err)
		c.JSON(http.StatusOK, gin.H{"success": true, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Payment processed"})
}

// handlePackagePayment xử lý thanh toán đăng ký gói
func handlePackagePayment(paymentCode string, payload *services.SepayWebhookPayload) error {
	db := config.GetDB()

	// Tìm subscription
	var subscription models.PackageSubscription
	if err := db.Where("payment_code = ?", paymentCode).First(&subscription).Error; err != nil {
		log.Printf("❌ Subscription not found: %s", paymentCode)
		return err
	}

	// Kiểm tra đã thanh toán chưa
	if subscription.PaymentStatus == "paid" {
		log.Printf("⏭️ Subscription already paid: %s", paymentCode)
		return nil
	}

	// Kiểm tra số tiền
	if payload.TransferAmount < subscription.Amount {
		log.Printf("⚠️ Amount mismatch: expected %.0f, got %.0f", subscription.Amount, payload.TransferAmount)
		// Vẫn có thể chấp nhận nếu số tiền >= yêu cầu
	}

	// Hoàn thành subscription
	return services.CompleteSubscription(subscription.ID, payload)
}

// handleOrderPayment xử lý thanh toán đơn hàng
func handleOrderPayment(paymentCode string, payload *services.SepayWebhookPayload) error {
	return services.CompleteOrderPayment(paymentCode, payload)
}

// saveUnmatchedTransaction lưu giao dịch không khớp code
func saveUnmatchedTransaction(payload *SepayWebhookPayload) {
	db := config.GetDB()

	rawJSON, _ := json.Marshal(payload)
	tx := models.PaymentTransaction{
		TransactionType:    "unknown",
		ReferenceID:        0,
		ReferenceCode:      "UNMATCHED",
		SepayTransactionID: &payload.ID,
		Gateway:            &payload.Gateway,
		AccountNumber:      &payload.AccountNumber,
		TransferType:       &payload.TransferType,
		TransferAmount:     payload.TransferAmount,
		Accumulated:        &payload.Accumulated,
		Code:               payload.Code,
		TransactionContent: &payload.TransactionContent,
		ReferenceNumber:    &payload.ReferenceNumber,
		Description:        &payload.Description,
		Status:             "unmatched",
		RawWebhookData:     stringPtr(string(rawJSON)),
	}

	db.Create(&tx)
}

// ===============================
// PAYMENT HANDLERS
// ===============================

// CreateSubscriptionInput input đăng ký gói
type CreateSubscriptionInput struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	Name           string `json:"name" binding:"required"`
	Phone          string `json:"phone"`
	RestaurantName string `json:"restaurant_name" binding:"required"`
	PackageID      uint   `json:"package_id" binding:"required"`
	BillingCycle   string `json:"billing_cycle"` // monthly, yearly
}

// CreateSubscription tạo đăng ký gói mới
// @Summary Tạo đăng ký gói
// @Description Đăng ký gói dịch vụ mới, nhận QR thanh toán
// @Tags Payment
// @Accept json
// @Produce json
// @Param subscription body CreateSubscriptionInput true "Thông tin đăng ký"
// @Success 201 {object} map[string]interface{}
// @Router /payment/subscribe [post]
func CreateSubscription(c *gin.Context) {
	var input CreateSubscriptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	// Gọi service
	result, err := services.CreateSubscription(services.CreateSubscriptionInput{
		Email:          input.Email,
		Password:       input.Password,
		Name:           input.Name,
		Phone:          input.Phone,
		RestaurantName: input.RestaurantName,
		PackageID:      input.PackageID,
		BillingCycle:   input.BillingCycle,
	})

	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), "SUBSCRIPTION_ERROR", "")
		return
	}

	// Nếu là gói miễn phí - trả về response khác
	if result.IsFree {
		utils.SuccessResponse(c, http.StatusCreated, gin.H{
			"subscription_id": result.SubscriptionID,
			"payment_code":    result.PaymentCode,
			"amount":          0,
			"package":         result.PackageName,
			"is_free":         true,
		}, "Đăng ký gói miễn phí thành công! Tài khoản đã được kích hoạt.")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"subscription_id": result.SubscriptionID,
		"payment_code":    result.PaymentCode,
		"amount":          result.Amount,
		"package":         result.PackageName,
		"qr_url":          result.QRCode.QRURL,
		"qr_content":      result.QRCode.QRContent,
		"bank_info": gin.H{
			"bank_name":      result.QRCode.BankName,
			"account_number": result.QRCode.AccountNo,
			"account_name":   result.QRCode.AccountName,
		},
		"expires_at":         result.ExpiresAt,
		"expires_in_minutes": result.ExpiresInMins,
		"is_free":            false,
	}, "Vui lòng chuyển khoản để hoàn tất đăng ký")
}

// GetSubscriptionStatus kiểm tra trạng thái đăng ký
// @Summary Kiểm tra trạng thái đăng ký
// @Description Kiểm tra đăng ký đã thanh toán chưa
// @Tags Payment
// @Accept json
// @Produce json
// @Param code path string true "Payment Code"
// @Success 200 {object} map[string]interface{}
// @Router /payment/subscribe/{code}/status [get]
func GetSubscriptionStatus(c *gin.Context) {
	code := c.Param("code")

	subscription, err := services.GetSubscriptionStatus(code)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), "NOT_FOUND", "")
		return
	}

	response := gin.H{
		"status":     subscription.PaymentStatus,
		"amount":     subscription.Amount,
		"expires_at": subscription.ExpiresAt,
	}

	if subscription.PaymentStatus == "paid" {
		response["paid_at"] = subscription.PaidAt
		response["user_id"] = subscription.UserID
		response["restaurant_id"] = subscription.RestaurantID
	}

	if subscription.Package != nil {
		response["package"] = subscription.Package.DisplayName
	}

	utils.SuccessResponse(c, http.StatusOK, response, "")
}

// GetSubscriptionQR lấy lại QR code
// @Summary Lấy QR code đăng ký
// @Description Lấy lại QR code cho đăng ký pending
// @Tags Payment
// @Accept json
// @Produce json
// @Param code path string true "Payment Code"
// @Success 200 {object} map[string]interface{}
// @Router /payment/subscribe/{code}/qr [get]
func GetSubscriptionQR(c *gin.Context) {
	code := c.Param("code")

	subscription, err := services.GetSubscriptionStatus(code)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), "NOT_FOUND", "")
		return
	}

	if subscription.PaymentStatus != "pending" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Đăng ký không còn pending", "NOT_PENDING", "")
		return
	}

	// Generate QR
	qr := services.GenerateAdminQR(subscription.Amount, subscription.PaymentCode)

	// Tính thời gian còn lại
	expiresInMins := int(time.Until(subscription.ExpiresAt).Minutes())
	if expiresInMins < 0 {
		expiresInMins = 0
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"qr_url":       qr.QRURL,
		"qr_content":   qr.QRContent,
		"amount":       subscription.Amount,
		"payment_code": subscription.PaymentCode,
		"bank_info": gin.H{
			"bank_name":      qr.BankName,
			"account_number": qr.AccountNo,
			"account_name":   qr.AccountName,
		},
		"expires_at":         subscription.ExpiresAt,
		"expires_in_minutes": expiresInMins,
	}, "")
}

// CreateOrderPaymentQR tạo QR thanh toán đơn hàng
// @Summary Tạo QR thanh toán đơn hàng
// @Description Tạo mã QR để khách thanh toán đơn hàng
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Router /payment/orders/{id}/qr [post]
func CreateOrderPaymentQR(c *gin.Context) {
	var orderID uint
	if _, err := parseUint(c.Param("id"), &orderID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Order ID không hợp lệ", "INVALID_ID", "")
		return
	}

	qr, err := services.CreateOrderPaymentQR(orderID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), "QR_ERROR", "")
		return
	}

	// Lấy order để trả về thêm thông tin
	db := config.GetDB()
	var order models.Order
	db.First(&order, orderID)

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"order_id":     orderID,
		"order_number": order.OrderNumber,
		"amount":       order.TotalAmount,
		"payment_code": order.PaymentCode,
		"qr_url":       qr.QRURL,
		"bank_info": gin.H{
			"bank_name":      qr.BankName,
			"account_number": qr.AccountNo,
			"account_name":   qr.AccountName,
		},
		"expires_at":         order.PaymentExpiresAt,
		"expires_in_minutes": 15,
	}, "Quét mã QR để thanh toán")
}

// GetOrderPaymentStatus kiểm tra trạng thái thanh toán đơn hàng
// @Summary Kiểm tra thanh toán đơn hàng
// @Description Kiểm tra đơn hàng đã thanh toán chưa
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Router /payment/orders/{id}/status [get]
func GetOrderPaymentStatus(c *gin.Context) {
	var orderID uint
	if _, err := parseUint(c.Param("id"), &orderID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Order ID không hợp lệ", "INVALID_ID", "")
		return
	}

	db := config.GetDB()
	var order models.Order
	if err := db.First(&order, orderID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy đơn hàng", "NOT_FOUND", "")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"order_id":       orderID,
		"payment_status": order.PaymentStatus,
		"payment_method": order.PaymentMethod,
		"paid_at":        order.PaidAt,
		"total_amount":   order.TotalAmount,
	}, "")
}

// ===============================
// RESTAURANT SEPAY LINKING
// ===============================

// LinkSepayInput input liên kết SePay cho nhà hàng
type LinkSepayInput struct {
	BankCode      string `json:"bank_code" binding:"required"`      // MB, VCB, TCB...
	AccountNumber string `json:"account_number" binding:"required"` // Số TK
	AccountName   string `json:"account_name" binding:"required"`   // Tên TK
}

// LinkSepayAccount liên kết tài khoản SePay cho nhà hàng
// @Summary Liên kết SePay cho nhà hàng
// @Description Nhà hàng cấu hình TK ngân hàng để nhận tiền từ khách
// @Tags Restaurant Payment
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Param input body LinkSepayInput true "Thông tin TK ngân hàng"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/sepay/link [post]
func LinkSepayAccount(c *gin.Context) {
	restaurantID, _ := parseUintParam(c.Param("id"))

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || restaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền", "FORBIDDEN", "")
		return
	}

	var input LinkSepayInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", "VALIDATION_ERROR", err.Error())
		return
	}

	db := config.GetDB()

	// Tìm hoặc tạo payment settings
	var settings models.PaymentSetting
	if err := db.Where("restaurant_id = ?", restaurantID).First(&settings).Error; err != nil {
		settings = models.PaymentSetting{
			RestaurantID: restaurantID,
		}
		db.Create(&settings)
	}

	// Cập nhật thông tin bank trước tiên
	now := time.Now()
	bankName := config.BankCodeToName(input.BankCode)

	updates := map[string]interface{}{
		"bank_code":       input.BankCode,
		"bank_name":       bankName,
		"account_number":  input.AccountNumber,
		"account_name":    input.AccountName,
		"sepay_linked":    true,
		"sepay_linked_at": now,
		"accept_qr":       true, // Tự động bật thanh toán QR
	}

	if err := db.Model(&settings).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Không thể cập nhật thông tin bank", "UPDATE_ERROR", err.Error())
		return
	}

	log.Printf("✅ Restaurant %d linked bank account: %s ****%s", restaurantID, bankName, input.AccountNumber[len(input.AccountNumber)-4:])

	// Tạo QR mẫu cho nhà hàng
	qr := services.GenerateRestaurantQR(input.BankCode, input.AccountNumber, input.AccountName, 0, "")

	// Trả về thông tin đã lưu
	response := gin.H{
		"message":      "Liên kết tài khoản ngân hàng thành công",
		"linked":       true,
		"linked_at":    now,
		"bank_code":    input.BankCode,
		"bank_name":    bankName,
		"account_no":   maskAccountNumber(input.AccountNumber),
		"account_name": input.AccountName,
		"qr_sample":    qr.QRURL,
	}

	utils.SuccessResponse(c, http.StatusOK, response, "Đã lưu thông tin tài khoản ngân hàng. Khách hàng có thể thanh toán qua QR.")
}

// CheckSepayLinkingSession kiểm tra trạng thái phiên kết nối SePay
// @Summary Kiểm tra trạng thái liên kết SePay
// @Description Kiểm tra xem nhà hàng đã hoàn tất xác thực SePay chưa
// @Tags Restaurant Payment
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Param session_id query string true "Session ID từ phiên tạo liên kết"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/sepay/link/check [get]
func CheckSepayLinkingSession(c *gin.Context) {
	restaurantID, _ := parseUintParam(c.Param("id"))
	sessionID := c.Query("session_id")

	if sessionID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Thiếu session_id", "MISSING_SESSION", "")
		return
	}

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || restaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền", "FORBIDDEN", "")
		return
	}

	// STEP 1: Gọi SePay API kiểm tra trạng thái
	sepayService := services.NewSepayService()
	statusResp, err := sepayService.GetLinkingStatus(sessionID)
	if err != nil {
		log.Printf("❌ Failed to get SePay linking status: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Không thể kiểm tra trạng thái: "+err.Error(), "SEPAY_ERROR", "")
		return
	}

	log.Printf("📊 SePay linking status: linked=%v", statusResp.Linked)

	// STEP 2: Nếu đã liên kết, cập nhật DB
	if statusResp.Linked {
		db := config.GetDB()
		now := time.Now()

		updates := map[string]interface{}{
			"sepay_linked":          true,
			"sepay_bank_account_id": statusResp.AccountID,
			"sepay_linked_at":       now,
		}

		if err := db.Model(&models.PaymentSetting{}).
			Where("restaurant_id = ?", restaurantID).
			Updates(updates).Error; err != nil {
			log.Printf("❌ Failed to update payment settings: %v", err)
		}

		log.Printf("✅ Restaurant %d linked to SePay successfully", restaurantID)
	}

	// STEP 3: Trả về trạng thái
	response := gin.H{
		"linked":       statusResp.Linked,
		"session_id":   sessionID,
		"account_id":   statusResp.AccountID,
		"bank_code":    statusResp.BankCode,
		"account_name": statusResp.AccountName,
	}

	message := "Chưa hoàn tất xác thực"
	if statusResp.Linked {
		response["linked_at"] = statusResp.LinkedAt
		message = "Đã liên kết SePay thành công"
	}

	utils.SuccessResponse(c, http.StatusOK, response, message)
}

// GetSepayStatus lấy trạng thái SePay của nhà hàng
// @Summary Lấy trạng thái SePay
// @Description Kiểm tra nhà hàng đã liên kết SePay chưa
// @Tags Restaurant Payment
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/sepay/status [get]
func GetSepayStatus(c *gin.Context) {
	restaurantID, _ := parseUintParam(c.Param("id"))

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || restaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền", "FORBIDDEN", "")
		return
	}

	db := config.GetDB()
	var settings models.PaymentSetting
	if err := db.Where("restaurant_id = ?", restaurantID).First(&settings).Error; err != nil {
		utils.SuccessResponse(c, http.StatusOK, gin.H{
			"linked":    false,
			"message":   "Chưa liên kết tài khoản ngân hàng",
			"accept_qr": false,
		}, "")
		return
	}

	response := gin.H{
		"linked":       settings.SepayLinked,
		"linked_at":    settings.SepayLinkedAt,
		"accept_cash":  settings.AcceptCash,
		"accept_qr":    settings.AcceptQR,
		"accept_momo":  settings.AcceptMomo,
		"accept_vnpay": settings.AcceptVNPay,
	}

	if settings.SepayLinked && settings.AccountNumber != nil {
		response["bank_name"] = settings.BankName
		response["account_no"] = maskAccountNumber(*settings.AccountNumber)
		response["account_name"] = settings.AccountName
	}

	utils.SuccessResponse(c, http.StatusOK, response, "")
}

// UnlinkSepayAccount hủy liên kết SePay
// @Summary Hủy liên kết SePay
// @Description Hủy liên kết tài khoản ngân hàng
// @Tags Restaurant Payment
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID"
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /restaurants/{id}/sepay/unlink [delete]
func UnlinkSepayAccount(c *gin.Context) {
	restaurantID, _ := parseUintParam(c.Param("id"))

	// Kiểm tra quyền
	currentRestaurantID, _ := c.Get("restaurant_id")
	role, _ := c.Get("role")

	if role != "admin" && (currentRestaurantID == nil || restaurantID != *currentRestaurantID.(*uint)) {
		utils.ErrorResponse(c, http.StatusForbidden, "Bạn không có quyền", "FORBIDDEN", "")
		return
	}

	db := config.GetDB()

	updates := map[string]interface{}{
		"sepay_linked":          false,
		"sepay_bank_account_id": nil,
		"sepay_linked_at":       nil,
		"accept_qr":             false,
	}

	db.Model(&models.PaymentSetting{}).Where("restaurant_id = ?", restaurantID).Updates(updates)

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Đã hủy liên kết",
		"linked":  false,
	}, "Hủy liên kết thành công")
}

// ===============================
// HELPER FUNCTIONS
// ===============================

func stringPtr(s string) *string {
	return &s
}

func parseUint(s string, result *uint) (bool, error) {
	var val uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, nil
		}
		val = val*10 + uint64(c-'0')
	}
	*result = uint(val)
	return true, nil
}

func parseUintParam(s string) (uint, error) {
	var val uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		val = val*10 + uint64(c-'0')
	}
	return uint(val), nil
}

// maskAccountNumber che bớt số tài khoản: 0393531965 -> ****1965
func maskAccountNumber(acc string) string {
	if len(acc) <= 4 {
		return acc
	}
	return "****" + acc[len(acc)-4:]
}
