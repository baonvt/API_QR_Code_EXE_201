package config

import (
	"go-api/models"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RunMigrations thực hiện auto migrate tất cả các model
func RunMigrations() error {
	db := GetDB()

	log.Println("🔄 Running database migrations...")

	// Migrate theo thứ tự để đảm bảo foreign key constraints
	err := db.AutoMigrate(
		&models.User{},                // 1. Users (base table)
		&models.Package{},             // 2. Packages (base table)
		&models.Restaurant{},          // 3. Restaurants (depends on users, packages)
		&models.PaymentSetting{},      // 4. Payment Settings (depends on restaurants)
		&models.Table{},               // 5. Tables (depends on restaurants)
		&models.Category{},            // 6. Categories (depends on restaurants)
		&models.MenuItem{},            // 7. Menu Items (depends on restaurants, categories)
		&models.Order{},               // 8. Orders (depends on restaurants, tables)
		&models.OrderItem{},           // 9. Order Items (depends on orders, menu_items)
		&models.PackageSubscription{}, // 10. Package Subscriptions (depends on packages)
		&models.PaymentTransaction{},  // 11. Payment Transactions (standalone)
		&models.Notification{},        // 12. Notifications (depends on restaurants)
		&models.ContactMessage{},      // 13. Contact Messages (standalone)
	)

	if err != nil {
		log.Printf("❌ Migration failed: %v", err)
		return err
	}

	log.Println("✅ Database migrations completed successfully!")
	return nil
}

// SeedPackages tạo dữ liệu mẫu cho packages
func SeedPackages() error {
	db := GetDB()

	// Kiểm tra xem đã có packages chưa
	var count int64
	db.Model(&models.Package{}).Count(&count)
	if count > 0 {
		log.Println("📦 Packages already seeded, skipping...")
		return nil
	}

	return createPackages(db)
}

// ReseedPackages xóa tất cả packages cũ và tạo lại packages mới
func ReseedPackages() error {
	db := GetDB()

	log.Println("� Updating packages with new data...")

	// Define packages với dữ liệu mới
	packagesData := []struct {
		Name          string
		DisplayName   string
		Description   string
		MonthlyPrice  float64
		YearlyPrice   float64
		MaxMenuItems  int
		MaxTables     int
		MaxCategories int
		Features      string
		IsPopular     bool
		SortOrder     int
	}{
		{
			Name:          "Starter",
			DisplayName:   "Starter",
			Description:   "Gói dùng thử miễn phí 7 ngày",
			MonthlyPrice:  0,
			YearlyPrice:   0,
			MaxMenuItems:  10,
			MaxTables:     3,
			MaxCategories: 3,
			Features:      `["Quản lý 10 món ăn", "Tối đa 3 bàn", "Đặt món qua QR", "Thanh toán tiền mặt"]`,
			IsPopular:     false,
			SortOrder:     0,
		},
		{
			Name:          "Basic",
			DisplayName:   "Gói Cơ Bản",
			Description:   "Dành cho quán nhỏ, phục vụ dưới 40 khách/lượt",
			MonthlyPrice:  229000,
			YearlyPrice:   2290000,
			MaxMenuItems:  30,
			MaxTables:     10,
			MaxCategories: 3,
			Features:      `["Tạo thực đơn (tối đa 30 món)", "Gọi món bằng mã QR", "Thống kê doanh thu cơ bản", "Quản lý tối đa 10 bàn", "3 danh mục món ăn (Món chính - Đồ uống - Tráng miệng)", "Hỗ trợ qua email"]`,
			IsPopular:     false,
			SortOrder:     1,
		},
		{
			Name:          "Pro",
			DisplayName:   "Gói Chuyên Nghiệp",
			Description:   "Dành cho quán cà phê và nhà hàng đang phát triển",
			MonthlyPrice:  270000,
			YearlyPrice:   2700000,
			MaxMenuItems:  80,
			MaxTables:     25,
			MaxCategories: 6,
			Features:      `["Bao gồm tất cả tính năng của Gói Cơ Bản", "Quản lý nhân viên phục vụ", "Lưu trữ đám mây", "Quản lý tối đa 25 bàn", "Tạo đến 80 món ăn/đồ uống", "6 danh mục món ăn (Món chính - Món phụ - Đồ nướng - Lẩu - Đồ uống - Tráng miệng)", "Báo cáo doanh thu chi tiết theo danh mục", "Hỗ trợ 24/7"]`,
			IsPopular:     true,
			SortOrder:     2,
		},
		{
			Name:          "Premium",
			DisplayName:   "Gói Cao Cấp",
			Description:   "Dành cho chuỗi hoặc nhà hàng có nhiều chi nhánh",
			MonthlyPrice:  279000,
			YearlyPrice:   2790000,
			MaxMenuItems:  -1,
			MaxTables:     -1,
			MaxCategories: -1,
			Features:      `["Bao gồm tất cả tính năng của Gói Chuyên Nghiệp", "Hỗ trợ kỹ thuật ưu tiên", "Kết nối nhiều chi nhánh", "Đánh giá & đặt chỗ của khách hàng", "Quản lý không giới hạn số bàn và món ăn", "Tạo danh mục tùy chỉnh linh hoạt", "Tích hợp thực đơn số đồng bộ giữa các chi nhánh", "API tích hợp", "Hỗ trợ ưu tiên 24/7", "Tùy chỉnh theo yêu cầu"]`,
			IsPopular:     false,
			SortOrder:     3,
		},
	}

	for _, p := range packagesData {
		// Tìm package theo name
		var pkg models.Package
		result := db.Where("name = ?", p.Name).First(&pkg)

		if result.Error != nil {
			// Package chưa tồn tại, tạo mới
			log.Printf("📦 Creating new package: %s", p.Name)
			newPkg := models.Package{
				Name:          p.Name,
				DisplayName:   p.DisplayName,
				Description:   &p.Description,
				MonthlyPrice:  p.MonthlyPrice,
				YearlyPrice:   p.YearlyPrice,
				MaxMenuItems:  p.MaxMenuItems,
				MaxTables:     p.MaxTables,
				MaxCategories: p.MaxCategories,
				Features:      &p.Features,
				IsPopular:     p.IsPopular,
				IsActive:      true,
				SortOrder:     p.SortOrder,
			}
			if err := db.Create(&newPkg).Error; err != nil {
				log.Printf("❌ Failed to create package %s: %v", p.Name, err)
				return err
			}
		} else {
			// Package đã tồn tại, cập nhật
			log.Printf("📦 Updating package: %s", p.Name)
			updates := map[string]interface{}{
				"display_name":   p.DisplayName,
				"description":    p.Description,
				"monthly_price":  p.MonthlyPrice,
				"yearly_price":   p.YearlyPrice,
				"max_menu_items": p.MaxMenuItems,
				"max_tables":     p.MaxTables,
				"max_categories": p.MaxCategories,
				"features":       p.Features,
				"is_popular":     p.IsPopular,
				"sort_order":     p.SortOrder,
			}
			if err := db.Model(&pkg).Updates(updates).Error; err != nil {
				log.Printf("❌ Failed to update package %s: %v", p.Name, err)
				return err
			}
		}
	}

	log.Println("✅ Packages updated successfully!")
	return nil
}

// createPackages tạo packages trong database
func createPackages(db *gorm.DB) error {
	log.Println("🌱 Seeding packages...")

	packages := []models.Package{
		{
			Name:          "Starter",
			DisplayName:   "Starter",
			Description:   stringPtr("Gói dùng thử miễn phí 7 ngày"),
			MonthlyPrice:  0,
			YearlyPrice:   0,
			MaxMenuItems:  10,
			MaxTables:     3,
			MaxCategories: 3,
			Features:      stringPtr(`["Quản lý 10 món ăn", "Tối đa 3 bàn", "Đặt món qua QR", "Thanh toán tiền mặt"]`),
			IsPopular:     false,
			IsActive:      true,
			SortOrder:     0,
		},
		{
			Name:          "Basic",
			DisplayName:   "Gói Cơ Bản",
			Description:   stringPtr("Dành cho quán nhỏ, phục vụ dưới 40 khách/lượt"),
			MonthlyPrice:  229000,
			YearlyPrice:   2290000,
			MaxMenuItems:  30,
			MaxTables:     10,
			MaxCategories: 3,
			Features:      stringPtr(`["Tạo thực đơn (tối đa 30 món)", "Gọi món bằng mã QR", "Thống kê doanh thu cơ bản", "Quản lý tối đa 10 bàn", "3 danh mục món ăn (Món chính - Đồ uống - Tráng miệng)", "Hỗ trợ qua email"]`),
			IsPopular:     false,
			IsActive:      true,
			SortOrder:     1,
		},
		{
			Name:          "Pro",
			DisplayName:   "Gói Chuyên Nghiệp",
			Description:   stringPtr("Dành cho quán cà phê và nhà hàng đang phát triển"),
			MonthlyPrice:  270000,
			YearlyPrice:   2700000,
			MaxMenuItems:  80,
			MaxTables:     25,
			MaxCategories: 6,
			Features:      stringPtr(`["Bao gồm tất cả tính năng của Gói Cơ Bản", "Quản lý nhân viên phục vụ", "Lưu trữ đám mây", "Quản lý tối đa 25 bàn", "Tạo đến 80 món ăn/đồ uống", "6 danh mục món ăn (Món chính - Món phụ - Đồ nướng - Lẩu - Đồ uống - Tráng miệng)", "Báo cáo doanh thu chi tiết theo danh mục", "Hỗ trợ 24/7"]`),
			IsPopular:     true,
			IsActive:      true,
			SortOrder:     2,
		},
		{
			Name:          "Premium",
			DisplayName:   "Gói Cao Cấp",
			Description:   stringPtr("Dành cho chuỗi hoặc nhà hàng có nhiều chi nhánh"),
			MonthlyPrice:  279000,
			YearlyPrice:   2790000,
			MaxMenuItems:  -1, // Unlimited
			MaxTables:     -1, // Unlimited
			MaxCategories: -1, // Unlimited
			Features:      stringPtr(`["Bao gồm tất cả tính năng của Gói Chuyên Nghiệp", "Hỗ trợ kỹ thuật ưu tiên", "Kết nối nhiều chi nhánh", "Đánh giá & đặt chỗ của khách hàng", "Quản lý không giới hạn số bàn và món ăn", "Tạo danh mục tùy chỉnh linh hoạt", "Tích hợp thực đơn số đồng bộ giữa các chi nhánh", "API tích hợp", "Hỗ trợ ưu tiên 24/7", "Tùy chỉnh theo yêu cầu"]`),
			IsPopular:     false,
			IsActive:      true,
			SortOrder:     3,
		},
	}

	for _, pkg := range packages {
		if err := db.Create(&pkg).Error; err != nil {
			log.Printf("❌ Failed to seed package %s: %v", pkg.Name, err)
			return err
		}
	}

	log.Println("✅ Packages seeded successfully!")
	return nil
}

// SeedAdminUser tạo tài khoản admin mặc định
func SeedAdminUser() error {
	db := GetDB()

	// Hash password với bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("❌ Failed to hash password: %v", err)
		return err
	}

	// Kiểm tra xem đã có admin chưa
	var existingAdmin models.User
	if err := db.Where("email = ?", "admin@fbmanager.com").First(&existingAdmin).Error; err == nil {
		// Admin đã tồn tại - cập nhật password
		db.Model(&existingAdmin).Update("password", string(hashedPassword))
		log.Println("👤 Admin user already exists, password updated!")
		return nil
	}

	log.Println("🌱 Seeding admin user...")

	admin := models.User{
		Email:    "admin@fbmanager.com",
		Password: string(hashedPassword),
		Name:     "Admin Hệ Thống",
		Role:     "admin",
		Phone:    stringPtr("0900000000"),
		IsActive: true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("❌ Failed to seed admin user: %v", err)
		return err
	}

	log.Println("✅ Admin user seeded successfully!")
	log.Println("   📧 Email: admin@fbmanager.com")
	log.Println("   🔑 Password: admin123")
	return nil
}

// RunSeeds chạy tất cả seed data
func RunSeeds() error {
	if err := SeedPackages(); err != nil {
		return err
	}
	if err := SeedAdminUser(); err != nil {
		return err
	}
	return nil
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
