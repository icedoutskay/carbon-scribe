package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"carbon-scribe/project-portal/project-portal-backend/internal/config"
	"carbon-scribe/project-portal/project-portal-backend/internal/geospatial"
	"carbon-scribe/project-portal/project-portal-backend/pkg/postgis"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env file (ignore error if not found - will use system env vars)
	_ = godotenv.Load()

	// Initialize logger
	zapLogger, err := initLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	zapLogger.Info("Starting CarbonScribe Project Portal API")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		zapLogger.Fatal("Failed to load configuration", zap.Error(err))
	}

	zapLogger.Info("Configuration loaded",
		zap.String("mode", cfg.Server.Mode),
		zap.Int("port", cfg.Server.Port))

	// Initialize database
	dbClient, err := initDatabase(cfg, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer dbClient.Close()

	zapLogger.Info("Database connection established")

	// Initialize services
	geospatialService := geospatial.NewService(dbClient.DB, zapLogger)

	// Initialize handlers
	geospatialHandler := geospatial.NewHandler(geospatialService, zapLogger)

	// Setup Gin
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check endpoint
	router.GET("/health", healthHandler(dbClient))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		geospatialHandler.RegisterRoutes(v1)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		zapLogger.Info("Server starting",
			zap.String("address", srv.Addr))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Error("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited")
}

// initLogger initializes the zap logger
func initLogger() (*zap.Logger, error) {
	env := os.Getenv("SERVER_MODE")
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

// initDatabase initializes the database connection
func initDatabase(cfg *config.Config, logger *zap.Logger) (*postgis.Client, error) {
	logLevel := logger.Info
	if cfg.Logging.Level == "debug" {
		logLevel = logger.Debug
	}

	gormLogLevel := logger.Default
	if cfg.Server.Mode == "production" {
		gormLogLevel = logger.Error
	}

	dbConfig := &postgis.Config{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		DBName:       cfg.Database.DBName,
		SSLMode:      cfg.Database.SSLMode,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
		LogLevel:     gormLogLevel,
	}

	client, err := postgis.NewClient(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := client.Health(); err != nil {
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	logLevel("Database connection successful")

	return client, nil
}

// healthHandler returns a health check handler
func healthHandler(dbClient *postgis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database
		if err := dbClient.Health(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "unhealthy",
				"database": "disconnected",
				"error":    err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"database":  "connected",
			"timestamp": time.Now().Unix(),
			"service":   "carbonscribe-project-portal",
		})
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "*"
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
	v1 "carbon-scribe/project-portal/project-portal-backend/api/v1"
	"carbon-scribe/project-portal/project-portal-backend/internal/monitoring"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	config := loadConfig()

	// Connect to database
	db, err := sqlx.Connect("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	fmt.Println("✅ Connected to database")

	// Initialize repository
	repo := monitoring.NewPostgresRepository(db)

	// Setup monitoring dependencies
	handler, satelliteIngestion, iotIngestion, alertEngine, err := v1.SetupDependencies(repo)
	if err != nil {
		log.Fatalf("Failed to setup monitoring dependencies: %v", err)
	}
	fmt.Println("✅ Monitoring dependencies initialized")

	// Create Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Register API routes
	api := router.Group("/api/v1")
	v1.RegisterMonitoringRoutes(api, handler)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "carbon-scribe-monitoring",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Start HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.Port),
		Handler: router,
	}

	// Channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		fmt.Printf("🚀 Server starting on port %s\n", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	fmt.Println("✅ Monitoring API server started")
	fmt.Printf("📡 Listening on http://localhost:%s\n", config.Port)
	fmt.Println("📊 Health check: http://localhost:" + config.Port + "/health")

	// Wait for interrupt signal
	<-quit
	fmt.Println("\n🛑 Shutdown signal received...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("✅ Server exited gracefully")
}

// Config holds application configuration
type Config struct {
	Port        string
	DatabaseURL string
	Debug       bool
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/carbon_scribe?sslmode=disable"
	}

	debug := os.Getenv("DEBUG") == "true"

	return &Config{
		Port:        port,
		DatabaseURL: databaseURL,
		Debug:       debug,
	}
}
