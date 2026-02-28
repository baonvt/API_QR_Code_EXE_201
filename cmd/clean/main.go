package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://postgres:xONvpdQBoVghOadodkdMaSOQBdyIEOSb@gondola.proxy.rlwy.net:18128/railway"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("🧹 Cleaning all data...")

	// Xóa theo thứ tự
	tables := []string{
		"DELETE FROM order_items",
		"DELETE FROM orders",
		"DELETE FROM tables",
		"DELETE FROM menu_items",
		"DELETE FROM categories",
		"DELETE FROM notifications",
	}

	for _, sql := range tables {
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("❌ Failed: %s - %v", sql, err)
		} else {
			fmt.Printf("✅ %s\n", sql)
		}
	}

	fmt.Println("🧹 Done!")
}
