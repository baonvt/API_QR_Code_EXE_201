package config

import (
	"go-api/models"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// RunMigrations thực hiện auto migrate tất cả các model
func RunMigrations() error {
	db := GetDB()

	log.Println("🔄 Running database migrations...")

	// Migrate theo thứ tự để đảm bảo foreign key constraints
	err := db.AutoMigrate(
		&models.User{},           // 1. Users (base table)
		&models.Package{},        // 2. Packages (base table)
		&models.Restaurant{},     // 3. Restaurants (depends on users, packages)
		&models.PaymentSetting{}, // 4. Payment Settings (depends on restaurants)
		&models.Table{},          // 5. Tables (depends on restaurants)
		&models.Category{},       // 6. Categories (depends on restaurants)
		&models.MenuItem{},       // 7. Menu Items (depends on restaurants, categories)
		&models.Order{},          // 8. Orders (depends on restaurants, tables)
		&models.OrderItem{},      // 9. Order Items (depends on orders, menu_items)
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

	log.Println("🌱 Seeding packages...")

	packages := []models.Package{
		{
			Name:          "Basic",
			DisplayName:   "Gói Cơ Bản",
			Description:   stringPtr("Phù hợp cho nhà hàng nhỏ, quán ăn gia đình"),
			MonthlyPrice:  199000,
			YearlyPrice:   1990000,
			MaxMenuItems:  30,
			MaxTables:     10,
			MaxCategories: 5,
			Features:      stringPtr(`["Quản lý 30 món ăn", "Tối đa 10 bàn", "Đặt món qua QR", "Thanh toán tiền mặt", "Báo cáo cơ bản"]`),
			IsPopular:     false,
			IsActive:      true,
			SortOrder:     1,
		},
		{
			Name:          "Pro",
			DisplayName:   "Gói Chuyên Nghiệp",
			Description:   stringPtr("Phù hợp cho nhà hàng vừa, có đội ngũ phục vụ"),
			MonthlyPrice:  499000,
			YearlyPrice:   4990000,
			MaxMenuItems:  80,
			MaxTables:     25,
			MaxCategories: 10,
			Features:      stringPtr(`["Quản lý 80 món ăn", "Tối đa 25 bàn", "Đặt món qua QR", "Thanh toán QR/MoMo/VNPay", "Báo cáo chi tiết", "Marketing cơ bản"]`),
			IsPopular:     true,
			IsActive:      true,
			SortOrder:     2,
		},
		{
			Name:          "Premium",
			DisplayName:   "Gói Cao Cấp",
			Description:   stringPtr("Phù hợp cho nhà hàng lớn, chuỗi nhà hàng"),
			MonthlyPrice:  999000,
			YearlyPrice:   9990000,
			MaxMenuItems:  -1, // Unlimited
			MaxTables:     -1, // Unlimited
			MaxCategories: -1, // Unlimited
			Features:      stringPtr(`["Không giới hạn món ăn", "Không giới hạn bàn", "Đặt món qua QR", "Tất cả phương thức thanh toán", "Báo cáo nâng cao", "Marketing đầy đủ", "Hỗ trợ ưu tiên 24/7"]`),
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
