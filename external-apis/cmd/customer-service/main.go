package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"external-apis/internal/customer/handler"
	"external-apis/internal/customer/repository"
	"external-apis/internal/customer/service"
	"external-apis/internal/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	initLogger()

	// Get port from environment or use default
	port := getEnv("PORT", "3002")

	logrus.WithField("port", port).Info("Starting Customer Service")

	// Initialize dependencies
	customerRepo := repository.NewMemoryCustomerRepository()
	customerService := service.NewCustomerService(customerRepo)
	customerHandler := handler.NewCustomerHandler(customerService)

	// Setup Gin router
	router := setupRouter(customerHandler)

	// Setup graceful shutdown
	setupGracefulShutdown()

	logrus.Info("✅ Customer Service started successfully")
	logrus.WithField("url", fmt.Sprintf("http://localhost:%s", port)).Info("Service is available")

	// Start server
	if err := router.Run(":" + port); err != nil {
		logrus.WithError(err).Fatal("Failed to start server")
	}
}

// initLogger configures the logger
func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	level := getEnv("LOG_LEVEL", "info")
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.WithError(err).Warn("Invalid log level, using info")
		logLevel = logrus.InfoLevel
	}

	logrus.SetLevel(logLevel)
	logrus.Info("Logger initialized")
}

// setupRouter configures the Gin router with middleware and routes
func setupRouter(customerHandler *handler.CustomerHandler) *gin.Engine {
	// Set Gin mode
	if getEnv("GIN_MODE", "debug") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "customer-service",
			"version": "1.0.0",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		customerHandler.RegisterRoutes(api)
	}

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Customer Service API",
			"version": "1.0.0",
			"endpoints": gin.H{
				"health":    "/health",
				"customers": "/api/customers",
			},
		})
	})

	return router
}

// setupGracefulShutdown sets up graceful shutdown handling
func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logrus.Info("Received shutdown signal, shutting down gracefully...")
		// Here you would close database connections, etc.
		logrus.Info("Customer Service shutdown complete")
		os.Exit(0)
	}()
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
