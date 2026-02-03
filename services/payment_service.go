package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go-api/config"
	"go-api/models"
	"go-api/utils"

	"golang.org/x/crypto/bcrypt"
)

// ===============================
// PACKAGE SUBSCRIPTION
// ===============================

// CreateSubscriptionInput input tạo đăng ký gói
type CreateSubscriptionInput struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	RestaurantName string `json:"restaurant_name"`
	PackageID      uint   `json:"package_id"`
	BillingCycle   string `json:"billing_cycle"` // monthly, yearly
}

// SubscriptionResult kết quả tạo đăng ký
type SubscriptionResult struct {
	SubscriptionID uint          `json:"subscription_id"`
	PaymentCode    string        `json:"payment_code"`
	Amount         float64       `json:"amount"`
	PackageName    string        `json:"package_name"`
	QRCode         *QRCodeResult `json:"qr_code"`
	ExpiresAt      time.Time     `json:"expires_at"`
	ExpiresInMins  int           `json:"expires_in_minutes"`
	IsFree         bool          `json:"is_free"` // true nếu gói miễn phí, đã tự động kích hoạt
}

// CreateSubscription tạo đăng ký gói mới (pending payment)
func CreateSubscription(input CreateSubscriptionInput) (*SubscriptionResult, error) {
	db := config.GetDB()

	// Kiểm tra email đã tồn tại trong users chưa
	var existingUser models.User
	if err := db.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("EMAIL_EXISTS: Email đã được sử dụng")
	}

	// Kiểm tra email đã có subscription pending chưa
	var existingSub models.PackageSubscription
	if err := db.Where("email = ? AND payment_status = ?", input.Email, "pending").First(&existingSub).Error; err == nil {
		// Nếu chưa hết hạn
		if existingSub.ExpiresAt.After(time.Now()) {
			// 🔥 FIX: Nếu user chọn gói KHÁC, cập nhật subscription thay vì trả về cũ
			if existingSub.PackageID != input.PackageID || existingSub.BillingCycle != input.BillingCycle {
				// Lấy thông tin gói mới
				var newPkg models.Package
				if err := db.First(&newPkg, input.PackageID).Error; err != nil {
					return nil, fmt.Errorf("INVALID_PACKAGE: Gói dịch vụ không tồn tại")
				}

				// Tính số tiền mới
				var newAmount float64
				billingCycle := input.BillingCycle
				if billingCycle == "yearly" {
					newAmount = newPkg.YearlyPrice
				} else {
					newAmount = newPkg.MonthlyPrice
					billingCycle = "monthly"
				}

				// Cập nhật subscription với gói mới
				newExpiresAt := time.Now().Add(30 * time.Minute)
				db.Model(&existingSub).Updates(map[string]interface{}{
					"package_id":    input.PackageID,
					"billing_cycle": billingCycle,
					"amount":        newAmount,
					"expires_at":    newExpiresAt,
				})

				// Tạo QR mới với số tiền mới
				qr := GenerateAdminQR(newAmount, existingSub.PaymentCode)

				log.Printf("🔄 Updated subscription %d: package %d -> %d, amount %.0f -> %.0f",
					existingSub.ID, existingSub.PackageID, input.PackageID, existingSub.Amount, newAmount)

				return &SubscriptionResult{
					SubscriptionID: existingSub.ID,
					PaymentCode:    existingSub.PaymentCode,
					Amount:         newAmount,
					PackageName:    newPkg.DisplayName,
					QRCode:         qr,
					ExpiresAt:      newExpiresAt,
					ExpiresInMins:  30,
				}, nil
			}

			// Nếu cùng gói, trả về subscription cũ
			var pkg models.Package
			db.First(&pkg, existingSub.PackageID)

			qr := GenerateAdminQR(existingSub.Amount, existingSub.PaymentCode)

			return &SubscriptionResult{
				SubscriptionID: existingSub.ID,
				PaymentCode:    existingSub.PaymentCode,
				Amount:         existingSub.Amount,
				PackageName:    pkg.DisplayName,
				QRCode:         qr,
				ExpiresAt:      existingSub.ExpiresAt,
				ExpiresInMins:  int(time.Until(existingSub.ExpiresAt).Minutes()),
			}, nil
		}
		// Hết hạn thì đánh dấu expired
		db.Model(&existingSub).Update("payment_status", "expired")
	}

	// Lấy thông tin gói
	var pkg models.Package
	if err := db.First(&pkg, input.PackageID).Error; err != nil {
		return nil, fmt.Errorf("INVALID_PACKAGE: Gói dịch vụ không tồn tại")
	}

	// Tính số tiền
	var amount float64
	if input.BillingCycle == "yearly" {
		amount = pkg.YearlyPrice
	} else {
		amount = pkg.MonthlyPrice
		input.BillingCycle = "monthly"
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("HASH_ERROR: Lỗi hệ thống")
	}

	// Tạo subscription
	expiresAt := time.Now().Add(30 * time.Minute) // QR hết hạn sau 30 phút

	subscription := models.PackageSubscription{
		Email:          input.Email,
		PasswordHash:   string(hashedPassword),
		Name:           input.Name,
		Phone:          &input.Phone,
		RestaurantName: input.RestaurantName,
		PackageID:      input.PackageID,
		BillingCycle:   input.BillingCycle,
		Amount:         amount,
		PaymentStatus:  "pending",
		ExpiresAt:      expiresAt,
	}

	// Lưu trước để lấy ID
	if err := db.Create(&subscription).Error; err != nil {
		return nil, fmt.Errorf("CREATE_ERROR: Không thể tạo đăng ký")
	}

	// Tạo payment code và cập nhật
	paymentCode := GeneratePackagePaymentCode(subscription.ID)
	qrContent := paymentCode

	db.Model(&subscription).Updates(map[string]interface{}{
		"payment_code": paymentCode,
		"qr_content":   qrContent,
	})

	// 🆓 Nếu gói MIỄN PHÍ (0đ) -> Tự động active ngay, không cần thanh toán
	if amount == 0 {
		log.Printf("🆓 Free package detected! Auto-activating subscription %d", subscription.ID)

		// Tạo fake transaction data
		freeTransaction := &SepayWebhookPayload{
			TransferAmount:     0,
			TransactionContent: paymentCode,
		}

		// Hoàn thành subscription ngay lập tức
		if err := CompleteSubscription(subscription.ID, freeTransaction); err != nil {
			log.Printf("❌ Failed to auto-complete free subscription: %v", err)
			return nil, fmt.Errorf("AUTO_ACTIVATE_ERROR: Không thể kích hoạt gói miễn phí")
		}

		// Trả về kết quả với status đã paid
		return &SubscriptionResult{
			SubscriptionID: subscription.ID,
			PaymentCode:    paymentCode,
			Amount:         0,
			PackageName:    pkg.DisplayName,
			QRCode:         nil, // Không cần QR
			ExpiresAt:      time.Now(),
			ExpiresInMins:  0,
			IsFree:         true, // Flag để frontend biết là gói free
		}, nil
	}

	// Tạo QR code cho gói có phí
	qr := GenerateAdminQR(amount, paymentCode)

	return &SubscriptionResult{
		SubscriptionID: subscription.ID,
		PaymentCode:    paymentCode,
		Amount:         amount,
		PackageName:    pkg.DisplayName,
		QRCode:         qr,
		ExpiresAt:      expiresAt,
		ExpiresInMins:  30,
	}, nil
}

// GetSubscriptionStatus kiểm tra trạng thái đăng ký
func GetSubscriptionStatus(paymentCode string) (*models.PackageSubscription, error) {
	db := config.GetDB()

	var subscription models.PackageSubscription
	if err := db.Preload("Package").Where("payment_code = ?", paymentCode).First(&subscription).Error; err != nil {
		return nil, fmt.Errorf("NOT_FOUND: Không tìm thấy đăng ký")
	}

	// Kiểm tra hết hạn
	if subscription.PaymentStatus == "pending" && subscription.ExpiresAt.Before(time.Now()) {
		db.Model(&subscription).Update("payment_status", "expired")
		subscription.PaymentStatus = "expired"
	}

	return &subscription, nil
}

// CompleteSubscription hoàn thành đăng ký sau khi thanh toán
func CompleteSubscription(subscriptionID uint, transactionData *SepayWebhookPayload) error {
	db := config.GetDB()

	var subscription models.PackageSubscription
	if err := db.First(&subscription, subscriptionID).Error; err != nil {
		return fmt.Errorf("NOT_FOUND: Không tìm thấy đăng ký")
	}

	if subscription.PaymentStatus == "paid" {
		return fmt.Errorf("ALREADY_PAID: Đăng ký đã được thanh toán")
	}

	// Bắt đầu transaction
	tx := db.Begin()

	// 1. Tạo User
	user := models.User{
		Email:    subscription.Email,
		Password: subscription.PasswordHash,
		Name:     subscription.Name,
		Phone:    subscription.Phone,
		Role:     "restaurant",
		IsActive: true,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("CREATE_USER_ERROR: %v", err)
	}

	// 2. Tạo Restaurant
	slug := utils.GenerateSlug(subscription.RestaurantName)
	var count int64
	tx.Model(&models.Restaurant{}).Where("slug LIKE ?", slug+"%").Count(&count)
	if count > 0 {
		slug = fmt.Sprintf("%s-%d", slug, count+1)
	}

	// Tính thời hạn gói
	var packageEndDate time.Time
	if subscription.BillingCycle == "yearly" {
		packageEndDate = time.Now().AddDate(1, 0, 0)
	} else {
		packageEndDate = time.Now().AddDate(0, 1, 0)
	}

	restaurant := models.Restaurant{
		OwnerID:          user.ID,
		PackageID:        subscription.PackageID,
		Name:             subscription.RestaurantName,
		Slug:             slug,
		IsOpen:           true,
		TaxRate:          10.0,
		ServiceCharge:    5.0,
		Currency:         "VND",
		PackageStartDate: time.Now(),
		PackageEndDate:   packageEndDate,
		PackageStatus:    "active",
		Status:           "active",
	}

	if err := tx.Create(&restaurant).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("CREATE_RESTAURANT_ERROR: %v", err)
	}

	// 3. Cập nhật subscription
	now := time.Now()
	tx.Model(&subscription).Updates(map[string]interface{}{
		"payment_status": "paid",
		"paid_at":        now,
		"user_id":        user.ID,
		"restaurant_id":  restaurant.ID,
	})

	// 4. Lưu transaction record
	rawData, _ := json.Marshal(transactionData)
	transaction := models.PaymentTransaction{
		TransactionType:    "package",
		ReferenceID:        subscription.ID,
		ReferenceCode:      subscription.PaymentCode,
		SepayTransactionID: &transactionData.ID,
		Gateway:            &transactionData.Gateway,
		AccountNumber:      &transactionData.AccountNumber,
		TransferType:       &transactionData.TransferType,
		TransferAmount:     transactionData.TransferAmount,
		Accumulated:        &transactionData.Accumulated,
		Code:               transactionData.Code,
		TransactionContent: &transactionData.TransactionContent,
		ReferenceNumber:    &transactionData.ReferenceNumber,
		Description:        &transactionData.Description,
		Status:             "completed",
		VerifiedAt:         &now,
		RawWebhookData:     stringPtr(string(rawData)),
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("CREATE_TRANSACTION_ERROR: %v", err)
	}

	tx.Commit()

	log.Printf("✅ Subscription completed: ID=%d, User=%d, Restaurant=%d",
		subscription.ID, user.ID, restaurant.ID)

	return nil
}

// ===============================
// ORDER PAYMENT
// ===============================

// CreateOrderPaymentQR tạo QR thanh toán cho đơn hàng
func CreateOrderPaymentQR(orderID uint) (*QRCodeResult, error) {
	db := config.GetDB()

	var order models.Order
	if err := db.Preload("Restaurant").First(&order, orderID).Error; err != nil {
		return nil, fmt.Errorf("ORDER_NOT_FOUND: Không tìm thấy đơn hàng")
	}

	if order.PaymentStatus == "paid" {
		return nil, fmt.Errorf("ALREADY_PAID: Đơn hàng đã được thanh toán")
	}

	// Lấy payment settings của nhà hàng
	var settings models.PaymentSetting
	if err := db.Where("restaurant_id = ?", order.RestaurantID).First(&settings).Error; err != nil {
		return nil, fmt.Errorf("NO_PAYMENT_SETTINGS: Nhà hàng chưa cấu hình thanh toán")
	}

	// Kiểm tra nhà hàng đã cấu hình bank chưa
	if settings.AccountNumber == nil || settings.BankCode == nil {
		return nil, fmt.Errorf("NO_BANK_CONFIG: Nhà hàng chưa cấu hình tài khoản ngân hàng")
	}

	// Tạo payment code
	paymentCode := GenerateOrderPaymentCode(order.OrderNumber)
	expiresAt := time.Now().Add(15 * time.Minute)

	// Cập nhật order
	db.Model(&order).Updates(map[string]interface{}{
		"payment_code":       paymentCode,
		"payment_expires_at": expiresAt,
		"payment_status":     "pending",
	})

	// Tạo QR
	qr := GenerateRestaurantQR(
		*settings.BankCode,
		*settings.AccountNumber,
		*settings.AccountName,
		order.TotalAmount,
		paymentCode,
	)

	return qr, nil
}

// CompleteOrderPayment hoàn thành thanh toán đơn hàng
func CompleteOrderPayment(paymentCode string, transactionData *SepayWebhookPayload) error {
	db := config.GetDB()

	// Tìm order theo payment_code
	var order models.Order
	if err := db.Where("payment_code = ?", paymentCode).First(&order).Error; err != nil {
		return fmt.Errorf("ORDER_NOT_FOUND: Không tìm thấy đơn hàng với mã %s", paymentCode)
	}

	if order.PaymentStatus == "paid" {
		return fmt.Errorf("ALREADY_PAID: Đơn hàng đã được thanh toán")
	}

	// Kiểm tra số tiền
	if transactionData.TransferAmount < order.TotalAmount {
		return fmt.Errorf("AMOUNT_MISMATCH: Số tiền không khớp. Cần %.0f, nhận %.0f",
			order.TotalAmount, transactionData.TransferAmount)
	}

	now := time.Now()
	paymentMethod := "qr"

	// Cập nhật order
	db.Model(&order).Updates(map[string]interface{}{
		"payment_status": "paid",
		"payment_method": paymentMethod,
		"paid_at":        now,
	})

	// Lưu transaction record
	rawData, _ := json.Marshal(transactionData)
	transaction := models.PaymentTransaction{
		TransactionType:    "order",
		ReferenceID:        order.ID,
		ReferenceCode:      paymentCode,
		SepayTransactionID: &transactionData.ID,
		Gateway:            &transactionData.Gateway,
		AccountNumber:      &transactionData.AccountNumber,
		TransferType:       &transactionData.TransferType,
		TransferAmount:     transactionData.TransferAmount,
		Accumulated:        &transactionData.Accumulated,
		Code:               transactionData.Code,
		TransactionContent: &transactionData.TransactionContent,
		ReferenceNumber:    &transactionData.ReferenceNumber,
		Description:        &transactionData.Description,
		Status:             "completed",
		VerifiedAt:         &now,
		RawWebhookData:     stringPtr(string(rawData)),
	}

	db.Create(&transaction)

	log.Printf("✅ Order payment completed: OrderID=%d, Code=%s, Amount=%.0f",
		order.ID, paymentCode, transactionData.TransferAmount)

	return nil
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
