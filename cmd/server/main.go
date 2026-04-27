// Package main E-Logging REST API
//
//	@title			E-Logging API
//	@version		1.0
//	@description	A REST API for E-Logging with authentication, device management, and log tracking.
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io
//
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//
//	@host		localhost:3000
//	@BasePath	/api
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	_ "e-logging-app/docs" // ← ADD THIS
	"e-logging-app/internal/config"
	"e-logging-app/internal/db"
	"e-logging-app/internal/handlers"
	"e-logging-app/internal/middleware"
)

func main() {
	// Load configuration from .env file
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	database, err := db.NewDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Initialize repositories
	userRepo := db.NewUserRepository(database)
	stationRepo := db.NewStationRepository(database)
	deviceRepo := db.NewDeviceRepository(database)
	logRepo := db.NewLogRepository(database)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, deviceRepo)
	logHandler := handlers.NewLogHandler(logRepo, deviceRepo)
	stationHandler := handlers.NewStationHandler(stationRepo)
	deviceHandler := handlers.NewDeviceHandler(deviceRepo)
	userHandler := handlers.NewUserHandler(userRepo)
	sttHandler := handlers.NewSTTHandler()
	dashboardHandler := handlers.NewDashboardHandler(logRepo)

	// Initialize Fiber app
	app := fiber.New()

	// Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173, http://localhost:3000, http://127.0.0.1:5173",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization,X-Device-ID,X-Device-Name",
		MaxAge:       3600,
	}))

	// Auth routes
	authGroup := app.Group("/api/auth")
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)

	// ========== ADMIN SETUP ENDPOINT - COMMENT OUT AFTER FIRST ADMIN IS CREATED ==========
	// IMPORTANT: Only uncomment this for initial setup. Disable after creating the first admin.
	// To use: Make a POST to http://localhost:3000/api/admin/create-admin with JSON:
	// {"name":"Your Name","email":"admin@example.com","password":"securepassword"}
	// See ADMIN_SETUP.md for detailed instructions
	adminGroup := app.Group("/api/admin")
	adminGroup.Post("/create-admin", userHandler.CreateAdminUser)
	// =====================================================================================

	// Protected routes
	api := app.Group("/api", middleware.JWTMiddleware())

	// Logs
	logsGroup := api.Group("/logs")
	logsGroup.Get("/", logHandler.GetLogs)
	logsGroup.Post("/", middleware.DeviceFingerprintMiddleware(deviceRepo), middleware.RoleMiddleware("operator", "admin"), logHandler.CreateLog)
	logsGroup.Put("/:id", logHandler.UpdateLog)
	logsGroup.Get("/export", middleware.RoleMiddleware("downloader", "admin"), logHandler.ExportLogs)

	// Stations
	stationsGroup := api.Group("/stations")
	stationsGroup.Get("/", stationHandler.GetStations)
	stationsGroup.Post("/", middleware.RoleMiddleware("admin"), stationHandler.CreateStation)

	// Devices
	devicesGroup := api.Group("/devices")
	devicesGroup.Get("/", middleware.RoleMiddleware("admin"), deviceHandler.GetDevices)
	devicesGroup.Post("/", middleware.RoleMiddleware("admin"), deviceHandler.CreateDevice)
	devicesGroup.Put("/:id", middleware.RoleMiddleware("admin"), deviceHandler.UpdateDevice)
	devicesGroup.Delete("/:id", middleware.RoleMiddleware("admin"), deviceHandler.DeactivateDevice)

	// Users
	usersGroup := api.Group("/users")
	usersGroup.Get("/", middleware.RoleMiddleware("admin"), userHandler.GetUsers)
	usersGroup.Post("/", middleware.RoleMiddleware("admin"), userHandler.CreateUser)

	// STT
	api.Post("/stt", sttHandler.Transcribe)

	// Dashboard
	dashboardGroup := api.Group("/dashboard")
	dashboardGroup.Get("/stats", dashboardHandler.GetDashboardStats)

	// Swagger documentation
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	log.Printf("Starting server on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
