package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/yashjain/konnect/docs"

	"github.com/yashjain/konnect/internal/config"
	"github.com/yashjain/konnect/internal/database"
	"github.com/yashjain/konnect/internal/handlers"
)

// @title Services API
// @version 1.0
// @description A REST API for managing services and their versions
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	if err := database.Init(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Setup router
	router := setupRouter(cfg)

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Printf("Server failed to start: %v", err)
	}
}

// setupRouter configures the Gin router with all routes
func setupRouter(cfg *config.Config) *gin.Engine {
	// Set Gin mode based on configuration
	if cfg.LogLevel == "info" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	r.GET("/health", handlers.HealthCheck)

	// API routes
	setupAPIRoutes(r)

	return r
}

// setupAPIRoutes configures all API routes
func setupAPIRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		// Service routes
		api.GET("/services", handlers.GetServices)
		api.GET("/services/search", handlers.SearchServices)
		api.POST("/services", handlers.CreateService)
		api.GET("/services/:id", handlers.GetService)
		api.PUT("/services/:id", handlers.UpdateService)
		api.DELETE("/services/:id", handlers.DeleteService)

		// Version routes
		api.GET("/services/:id/versions", handlers.GetVersions)
		api.POST("/services/:id/versions", handlers.CreateVersion)
	}
}
