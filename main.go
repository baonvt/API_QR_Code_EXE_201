package main

import (
	"log"
	"os"

	"go-api/config"
	_ "go-api/docs" // Swagger docs
	"go-api/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Restaurant API
// @version 1.0
// @description API quản lý nhà hàng - Khách order và thanh toán trực tiếp

// @host apiqrcodeexe201-production.up.railway.app
// @BasePath /api/v1
// @schemes https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Nhập token: Bearer {token}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.ReleaseMode
	}
	gin.SetMode(mode)

	// Kết nối database
	config.ConnectDatabase()

	// Load SePay config
	config.LoadSepayConfig()

	// Chạy migrations
	if err := config.RunMigrations(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Chạy seed data
	if err := config.RunSeeds(); err != nil {
		log.Fatal("Failed to run seeds:", err)
	}

	// Khởi tạo Gin router
	router := gin.Default()

	// Serve static files (ảnh upload)
	router.Static("/assets", "./assets")

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup routes
	routes.SetupRoutes(router)

	// Lấy port từ environment variable
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Chạy server
	log.Printf("Server is running on http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
