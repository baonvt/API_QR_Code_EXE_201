package models

import (
	"time"

	"gorm.io/gorm"
)

// User model - Người dùng hệ thống
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"size:255;uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"size:255;not null"`
	Name      string         `json:"name" gorm:"size:255;not null"`
	Role      string         `json:"role" gorm:"size:20;default:'restaurant';not null"`
	Avatar    *string        `json:"avatar" gorm:"type:text"`
	Phone     *string        `json:"phone" gorm:"size:20"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Restaurant *Restaurant `json:"restaurant,omitempty" gorm:"foreignKey:OwnerID"`
}

func (User) TableName() string {
	return "users"
}

// Package model - Gói dịch vụ
type Package struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"size:50;uniqueIndex;not null"`
	DisplayName   string    `json:"display_name" gorm:"size:100;not null"`
	Description   *string   `json:"description" gorm:"size:500"`
	MonthlyPrice  float64   `json:"monthly_price" gorm:"type:decimal(12,0);not null"`
	YearlyPrice   float64   `json:"yearly_price" gorm:"type:decimal(12,0);not null"`
	MaxMenuItems  int       `json:"max_menu_items" gorm:"default:30"`
	MaxTables     int       `json:"max_tables" gorm:"default:10"`
	MaxCategories int       `json:"max_categories" gorm:"default:5"`
	Features      *string   `json:"features" gorm:"type:text"` // JSON array
	IsPopular     bool      `json:"is_popular" gorm:"default:false"`
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	SortOrder     int       `json:"sort_order" gorm:"default:0"`
	CreatedAt     time.Time `json:"created_at"`

	// Relationships
	Restaurants []Restaurant `json:"restaurants,omitempty" gorm:"foreignKey:PackageID"`
}

func (Package) TableName() string {
	return "packages"
}

// Restaurant model - Nhà hàng
type Restaurant struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	OwnerID     uint    `json:"owner_id" gorm:"not null;index"`
	PackageID   uint    `json:"package_id" gorm:"not null;index"`
	Name        string  `json:"name" gorm:"size:255;not null"`
	Slug        string  `json:"slug" gorm:"size:100;uniqueIndex;not null"`
	Description *string `json:"description" gorm:"size:1000"`
	Logo        *string `json:"logo" gorm:"type:text"`
	Phone       *string `json:"phone" gorm:"size:20"`
	Email       *string `json:"email" gorm:"size:255"`
	Address     *string `json:"address" gorm:"size:500"`

	IsOpen        bool    `json:"is_open" gorm:"default:true"`
	TaxRate       float64 `json:"tax_rate" gorm:"type:decimal(5,2);default:10.00"`
	ServiceCharge float64 `json:"service_charge" gorm:"type:decimal(5,2);default:5.00"`
	Currency      string  `json:"currency" gorm:"size:10;default:'VND'"`

	PackageStartDate time.Time `json:"package_start_date" gorm:"type:date;not null"`
	PackageEndDate   time.Time `json:"package_end_date" gorm:"type:date;not null"`
	PackageStatus    string    `json:"package_status" gorm:"size:20;default:'active'"`
	Status           string    `json:"status" gorm:"size:20;default:'active'"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Owner          *User           `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Package        *Package        `json:"package,omitempty" gorm:"foreignKey:PackageID"`
	PaymentSetting *PaymentSetting `json:"payment_setting,omitempty" gorm:"foreignKey:RestaurantID"`
	Tables         []Table         `json:"tables,omitempty" gorm:"foreignKey:RestaurantID"`
	Categories     []Category      `json:"categories,omitempty" gorm:"foreignKey:RestaurantID"`
	MenuItems      []MenuItem      `json:"menu_items,omitempty" gorm:"foreignKey:RestaurantID"`
	Orders         []Order         `json:"orders,omitempty" gorm:"foreignKey:RestaurantID"`
}

func (Restaurant) TableName() string {
	return "restaurants"
}

// PaymentSetting model - Cài đặt thanh toán
type PaymentSetting struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	RestaurantID  uint      `json:"restaurant_id" gorm:"uniqueIndex;not null"`
	BankName      *string   `json:"bank_name" gorm:"size:100"`
	AccountNumber *string   `json:"account_number" gorm:"size:50"`
	AccountName   *string   `json:"account_name" gorm:"size:255"`
	QRImage       *string   `json:"qr_image" gorm:"type:text"`
	AcceptCash    bool      `json:"accept_cash" gorm:"default:true"`
	AcceptQR      bool      `json:"accept_qr" gorm:"default:false"`
	AcceptMomo    bool      `json:"accept_momo" gorm:"default:false"`
	AcceptVNPay   bool      `json:"accept_vnpay" gorm:"default:false"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	Restaurant *Restaurant `json:"restaurant,omitempty" gorm:"foreignKey:RestaurantID"`
}

func (PaymentSetting) TableName() string {
	return "payment_settings"
}

// Table model - Bàn ăn
type Table struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RestaurantID uint      `json:"restaurant_id" gorm:"not null;index"`
	TableNumber  int       `json:"table_number" gorm:"not null"`
	Name         *string   `json:"name" gorm:"size:50"`
	Capacity     int       `json:"capacity" gorm:"default:4"`
	Status       string    `json:"status" gorm:"size:20;default:'available'"`
	QRCode       *string   `json:"qr_code" gorm:"type:text"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Restaurant *Restaurant `json:"restaurant,omitempty" gorm:"foreignKey:RestaurantID"`
	Orders     []Order     `json:"orders,omitempty" gorm:"foreignKey:TableID"`
}

func (Table) TableName() string {
	return "tables"
}

// Category model - Danh mục món ăn
type Category struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RestaurantID uint      `json:"restaurant_id" gorm:"not null;index"`
	Name         string    `json:"name" gorm:"size:100;not null"`
	Description  *string   `json:"description" gorm:"size:500"`
	Image        *string   `json:"image" gorm:"type:text"`
	SortOrder    int       `json:"sort_order" gorm:"default:0"`
	Status       string    `json:"status" gorm:"size:20;default:'active'"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Restaurant *Restaurant `json:"restaurant,omitempty" gorm:"foreignKey:RestaurantID"`
	MenuItems  []MenuItem  `json:"menu_items,omitempty" gorm:"foreignKey:CategoryID"`
}

func (Category) TableName() string {
	return "categories"
}

// MenuItem model - Món ăn
type MenuItem struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	RestaurantID uint      `json:"restaurant_id" gorm:"not null;index"`
	CategoryID   uint      `json:"category_id" gorm:"not null;index"`
	Name         string    `json:"name" gorm:"size:255;not null"`
	Description  *string   `json:"description" gorm:"size:1000"`
	Price        float64   `json:"price" gorm:"type:decimal(12,0);not null"`
	Image        *string   `json:"image" gorm:"type:text"`
	Options      *string   `json:"options" gorm:"type:text"` // JSON
	Tags         *string   `json:"tags" gorm:"type:text"`    // JSON
	PrepLocation string    `json:"prep_location" gorm:"size:20;default:'kitchen'"`
	PrepTime     int       `json:"prep_time" gorm:"default:15"`
	SortOrder    int       `json:"sort_order" gorm:"default:0"`
	Status       string    `json:"status" gorm:"size:20;default:'active'"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Restaurant *Restaurant `json:"restaurant,omitempty" gorm:"foreignKey:RestaurantID"`
	Category   *Category   `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	OrderItems []OrderItem `json:"order_items,omitempty" gorm:"foreignKey:MenuItemID"`
}

func (MenuItem) TableName() string {
	return "menu_items"
}

// Order model - Đơn hàng
type Order struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	RestaurantID   uint       `json:"restaurant_id" gorm:"not null;index"`
	TableID        uint       `json:"table_id" gorm:"not null;index"`
	OrderNumber    string     `json:"order_number" gorm:"size:50;not null;index"`
	CustomerName   *string    `json:"customer_name" gorm:"size:255"`
	CustomerPhone  *string    `json:"customer_phone" gorm:"size:20"`
	Status         string     `json:"status" gorm:"size:20;default:'pending'"`
	PaymentTiming  string     `json:"payment_timing" gorm:"size:10;default:'after'"`
	PaymentMethod  *string    `json:"payment_method" gorm:"size:20"`
	PaymentStatus  string     `json:"payment_status" gorm:"size:20;default:'unpaid'"`
	PaidAt         *time.Time `json:"paid_at"`
	Subtotal       float64    `json:"subtotal" gorm:"type:decimal(12,0);default:0"`
	TaxAmount      float64    `json:"tax_amount" gorm:"type:decimal(12,0);default:0"`
	ServiceCharge  float64    `json:"service_charge" gorm:"type:decimal(12,0);default:0"`
	DiscountAmount float64    `json:"discount_amount" gorm:"type:decimal(12,0);default:0"`
	TotalAmount    float64    `json:"total_amount" gorm:"type:decimal(12,0);default:0"`
	Notes          *string    `json:"notes" gorm:"size:1000"`
	CancelReason   *string    `json:"cancel_reason" gorm:"size:500"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at"`

	// Relationships
	Restaurant *Restaurant `json:"restaurant,omitempty" gorm:"foreignKey:RestaurantID"`
	Table      *Table      `json:"table,omitempty" gorm:"foreignKey:TableID"`
	OrderItems []OrderItem `json:"order_items,omitempty" gorm:"foreignKey:OrderID"`
}

func (Order) TableName() string {
	return "orders"
}

// OrderItem model - Chi tiết đơn hàng
type OrderItem struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	OrderID         uint      `json:"order_id" gorm:"not null;index"`
	MenuItemID      uint      `json:"menu_item_id" gorm:"not null;index"`
	ItemName        string    `json:"item_name" gorm:"size:255;not null"`
	ItemPrice       float64   `json:"item_price" gorm:"type:decimal(12,0);not null"`
	Quantity        int       `json:"quantity" gorm:"default:1;not null"`
	SelectedOptions *string   `json:"selected_options" gorm:"type:text"` // JSON
	Notes           *string   `json:"notes" gorm:"size:500"`
	PrepStatus      string    `json:"prep_status" gorm:"size:20;default:'pending'"`
	PrepLocation    string    `json:"prep_location" gorm:"size:20;default:'kitchen'"`
	LineTotal       float64   `json:"line_total" gorm:"type:decimal(12,0);not null"`
	CreatedAt       time.Time `json:"created_at"`

	// Relationships
	Order    *Order    `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	MenuItem *MenuItem `json:"menu_item,omitempty" gorm:"foreignKey:MenuItemID"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
