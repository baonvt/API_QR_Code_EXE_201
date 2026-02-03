package routes

import (
	"go-api/handlers"
	"go-api/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes thiết lập các routes cho API
func SetupRoutes(router *gin.Engine) {
	// CORS middleware
	router.Use(middleware.CORSMiddleware())

	// API versioning
	api := router.Group("/api/v1")

	// API Key middleware (tùy chọn - kiểm tra nếu có)
	api.Use(middleware.OptionalAPIKeyMiddleware())

	{
		// ================================
		// Health check
		// ================================
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"success": true,
				"message": "API is running",
				"version": "1.0.0",
			})
		})

		// ================================
		// AUTH - Public routes
		// ================================
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.Login)
			auth.POST("/register", handlers.Register)
			auth.POST("/check-email", handlers.CheckEmail)

			// Protected auth routes
			authProtected := auth.Group("")
			authProtected.Use(middleware.AuthMiddleware())
			{
				authProtected.POST("/logout", handlers.Logout)
				authProtected.GET("/me", handlers.GetMe)
				authProtected.POST("/refresh", handlers.RefreshToken)
			}
		}

		// ================================
		// PACKAGES - Public
		// ================================
		api.GET("/packages", handlers.GetPackages)

		// ================================
		// WEBHOOKS - Public (SePay callback)
		// ================================
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/sepay", handlers.HandleSepayWebhook)
		}

		// ================================
		// PAYMENT - Public (Đăng ký + Thanh toán)
		// ================================
		payment := api.Group("/payment")
		{
			// Đăng ký gói mới + nhận QR thanh toán
			payment.POST("/subscribe", handlers.CreateSubscription)
			// Kiểm tra trạng thái đăng ký
			payment.GET("/subscribe/:code/status", handlers.GetSubscriptionStatus)
			// Lấy lại QR code
			payment.GET("/subscribe/:code/qr", handlers.GetSubscriptionQR)
			// Tạo QR thanh toán đơn hàng
			payment.POST("/orders/:id/qr", handlers.CreateOrderPaymentQR)
			// Kiểm tra trạng thái thanh toán đơn hàng
			payment.GET("/orders/:id/status", handlers.GetOrderPaymentStatus)
		}

		// ================================
		// PUBLIC - Customer routes (by slug)
		// ================================
		public := api.Group("/public")
		{
			// Xem nhà hàng theo slug
			public.GET("/restaurants/:slug", handlers.GetRestaurantBySlug)
			// Xem menu theo slug
			public.GET("/restaurants/:slug/menu", handlers.GetMenuBySlug)
			// Xem bàn theo slug + số bàn (cho khách quét QR)
			public.GET("/restaurants/:slug/tables/:tableNumber", handlers.GetTableBySlugAndNumber)
			// Customer tạo đơn hàng
			public.POST("/restaurants/:slug/orders", handlers.CreateOrder)
		}

		// ================================
		// RESTAURANTS - By ID (Protected)
		// ================================
		restaurants := api.Group("/restaurants")
		{
			// Public: Xem categories
			restaurants.GET("/:id/categories", handlers.GetCategories)

			// Public: Xem menu
			restaurants.GET("/:id/menu", handlers.GetMenu)

			// Protected routes
			restaurantsProtected := restaurants.Group("")
			restaurantsProtected.Use(middleware.AuthMiddleware())
			restaurantsProtected.Use(middleware.RestaurantOrAdmin())
			{
				// Restaurant info
				restaurantsProtected.GET("/me", handlers.GetMyRestaurant)
				restaurantsProtected.PUT("/:id", handlers.UpdateRestaurant)

				// Tables
				restaurantsProtected.GET("/:id/tables", handlers.GetTables)
				restaurantsProtected.POST("/:id/tables", handlers.CreateTable)

				// Categories
				restaurantsProtected.POST("/:id/categories", handlers.CreateCategory)

				// Menu
				restaurantsProtected.POST("/:id/menu", handlers.CreateMenuItem)

				// Orders
				restaurantsProtected.GET("/:id/orders", handlers.GetOrders)

				// Payment Settings
				restaurantsProtected.GET("/:id/payment-settings", handlers.GetPaymentSettings)
				restaurantsProtected.PUT("/:id/payment-settings", handlers.UpdatePaymentSettings)

				// SePay Linking (Restaurant nhận tiền từ khách)
				restaurantsProtected.POST("/:id/sepay/link", handlers.LinkSepayAccount)
				restaurantsProtected.GET("/:id/sepay/link/check", handlers.CheckSepayLinkingSession)
				restaurantsProtected.GET("/:id/sepay/status", handlers.GetSepayStatus)
				restaurantsProtected.DELETE("/:id/sepay/unlink", handlers.UnlinkSepayAccount)

				// Statistics
				restaurantsProtected.GET("/:id/stats/overview", handlers.GetStatsOverview)
				restaurantsProtected.GET("/:id/stats/revenue", handlers.GetStatsRevenue)
				restaurantsProtected.GET("/:id/stats/menu", handlers.GetStatsMenu)
			}
		}

		// ================================
		// TABLES - Protected
		// ================================
		tables := api.Group("/tables")
		tables.Use(middleware.AuthMiddleware())
		tables.Use(middleware.RestaurantOrAdmin())
		{
			tables.GET("/:id/detail", handlers.GetTableDetail)
			tables.PUT("/:id", handlers.UpdateTable)
			tables.DELETE("/:id", handlers.DeleteTable)
		}

		// ================================
		// CATEGORIES - Mixed
		// ================================
		categories := api.Group("/categories")
		{
			// Public: Xem món theo danh mục
			categories.GET("/:id/items", handlers.GetCategoryItems)

			// Protected
			categoriesProtected := categories.Group("")
			categoriesProtected.Use(middleware.AuthMiddleware())
			categoriesProtected.Use(middleware.RestaurantOrAdmin())
			{
				categoriesProtected.PUT("/:id", handlers.UpdateCategory)
				categoriesProtected.DELETE("/:id", handlers.DeleteCategory)
			}
		}

		// ================================
		// MENU - Protected
		// ================================
		menu := api.Group("/menu")
		menu.Use(middleware.AuthMiddleware())
		menu.Use(middleware.RestaurantOrAdmin())
		{
			menu.PUT("/:id", handlers.UpdateMenuItem)
			menu.DELETE("/:id", handlers.DeleteMenuItem)
		}

		// ================================
		// ORDERS - Mixed
		// ================================
		orders := api.Group("/orders")
		{
			// Public: Xem chi tiết đơn hàng (cho tracking)
			orders.GET("/:id", handlers.GetOrder)

			// Public: Thêm món vào đơn hàng
			orders.POST("/:id/items", handlers.AddOrderItems)

			// Protected
			ordersProtected := orders.Group("")
			ordersProtected.Use(middleware.AuthMiddleware())
			ordersProtected.Use(middleware.RestaurantOrAdmin())
			{
				ordersProtected.PUT("/:id/status", handlers.UpdateOrderStatus)
				ordersProtected.PUT("/:id/pay", handlers.PayOrder)
				ordersProtected.GET("/:id/bill", handlers.GetOrderBill)
			}
		}

		// ================================
		// ADMIN - Admin only
		// ================================
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.AdminOnly())
		{
			// Restaurants management
			admin.GET("/restaurants", handlers.GetAllRestaurants)
			admin.PUT("/restaurants/:id/status", handlers.UpdateRestaurantStatus)

			// Packages management
			admin.POST("/packages", handlers.CreatePackage)
			admin.PUT("/packages/:id", handlers.UpdatePackage)

			// Stats
			admin.GET("/stats", handlers.GetAdminStats)
		}

		// ================================
		// UPLOAD - Protected
		// ================================
		upload := api.Group("/upload")
		upload.Use(middleware.AuthMiddleware())
		{
			upload.POST("/image", handlers.UploadImage)
			upload.POST("/images", handlers.UploadMultipleImages)
			upload.POST("/url", handlers.UploadImageFromURL)
			upload.DELETE("/image", handlers.DeleteImageHandler)
		}
	}
}
