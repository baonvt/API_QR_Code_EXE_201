package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase kết nối đến PostgreSQL database
func ConnectDatabase() {
	// Load file .env (optional for local dev)
	_ = godotenv.Load()

	// Lấy DATABASE_URL từ environment hoặc dùng default
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Default Railway PostgreSQL
		dsn = "postgresql://postgres:uWLeEYyJrtghKNqabNlBQodJpGpVBnyt@shinkansen.proxy.rlwy.net:18641/railway"
	}

	log.Println("Connecting to database...")

	// Kết nối database
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully!")
	DB = database
}

// GetDB trả về instance của database
func GetDB() *gorm.DB {
	return DB
}
}
