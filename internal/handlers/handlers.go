package handlers

import (
	"encoding/csv"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"e-logging-app/internal/auth"
	"e-logging-app/internal/db"
	"e-logging-app/internal/models"
)

type AuthHandler struct {
	userRepo db.UserRepository
}

type LogHandler struct {
	logRepo     db.LogRepository
	deviceRepo  db.DeviceRepository
}

type StationHandler struct {
	stationRepo db.StationRepository
}

type DeviceHandler struct {
	deviceRepo db.DeviceRepository
}

type UserHandler struct {
	userRepo db.UserRepository
}

type STTHandler struct{}

func NewAuthHandler(userRepo db.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

func NewLogHandler(logRepo db.LogRepository, deviceRepo db.DeviceRepository) *LogHandler {
	return &LogHandler{logRepo: logRepo, deviceRepo: deviceRepo}
}

func NewStationHandler(stationRepo db.StationRepository) *StationHandler {
	return &StationHandler{stationRepo: stationRepo}
}

func NewDeviceHandler(deviceRepo db.DeviceRepository) *DeviceHandler {
	return &DeviceHandler{deviceRepo: deviceRepo}
}

func NewUserHandler(userRepo db.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func NewSTTHandler() *STTHandler {
	return &STTHandler{}
}

// Login authenticates a user with email and password
// @Summary User login
// @Description Authenticates a user with email and password and returns JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param login body map[string]string true "Login credentials" example({"email":"user@example.com","password":"password123"})
// @Success 200 {object} map[string]interface{} "Access and refresh tokens"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	user, err := h.userRepo.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid credentials",
		})
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid credentials",
		})
	}

	accessToken, err := auth.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to generate access token",
		})
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to generate refresh token",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}

// Refresh generates a new access token using a refresh token
// @Summary Refresh access token
// @Description Generates a new access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body map[string]string true "Refresh token" example({"refresh_token":"token_value"})
// @Success 200 {object} map[string]interface{} "New access token"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Invalid refresh token"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	userID, err := auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid refresh token",
		})
	}

	user, err := h.userRepo.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get user",
		})
	}

	accessToken, err := auth.GenerateJWT(userID, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to generate access token",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"access_token": accessToken,
		},
	})
}

// GetLogs retrieves logs with filtering, sorting, and pagination
// @Summary Get logs
// @Description Retrieves logs with optional filtering, sorting, and pagination
// @Tags logs
// @Security BearerAuth
// @Produce json
// @Param sort_by query string false "Sort by field (created_at, log_date)" default(created_at)
// @Param order query string false "Sort order (asc, desc)" default(desc)
// @Param station_id query string false "Filter by station ID (UUID)"
// @Param date_from query string false "Filter logs from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter logs until date (YYYY-MM-DD)"
// @Param limit query int false "Limit number of results" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{} "List of logs"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /logs [get]
func (h *LogHandler) GetLogs(c *fiber.Ctx) error {
	sortBy := c.Query("sort_by", "created_at")
	order := c.Query("order", "desc")
	stationIDStr := c.Query("station_id")
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")
	limitStr := c.Query("limit", "10")
	offsetStr := c.Query("offset", "0")

	filters := make(map[string]interface{})
	if stationIDStr != "" {
		stationID, err := uuid.Parse(stationIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid station_id",
			})
		}
		filters["station_id"] = stationID
	}
	if dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid date_from",
			})
		}
		filters["date_from"] = dateFrom
	}
	if dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid date_to",
			})
		}
		filters["date_to"] = dateTo
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid limit",
			})
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid offset",
		})
	}

	logs, err := h.logRepo.GetLogs(c.Context(), filters, sortBy, order, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get logs",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    logs,
	})
}

// CreateLog creates a new log entry
// @Summary Create log
// @Description Creates a new log entry (requires device fingerprint middleware)
// @Tags logs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param log body models.Log true "Log data"
// @Success 201 {object} map[string]interface{} "Created log"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /logs [post]
func (h *LogHandler) CreateLog(c *fiber.Ctx) error {
	var log models.Log
	if err := c.BodyParser(&log); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)
	deviceID := c.Locals("device_id").(uuid.UUID)

	log.CreatedBy = userID
	log.DeviceID = deviceID

	if err := h.logRepo.CreateLog(c.Context(), &log); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create log",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    log,
	})
}

// UpdateLog updates an existing log entry (only within 24 hours of creation)
// @Summary Update log
// @Description Updates a log entry. Only the creator can update, and only within 24 hours of creation
// @Tags logs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Log ID (UUID)"
// @Param log body models.Log true "Updated log data"
// @Success 200 {object} map[string]interface{} "Updated log"
// @Failure 400 {object} map[string]interface{} "Invalid request body or ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden (not creator or after 24 hours)"
// @Failure 404 {object} map[string]interface{} "Log not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /logs/{id} [put]
func (h *LogHandler) UpdateLog(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid log ID",
		})
	}

	var log models.Log
	if err := c.BodyParser(&log); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	existingLog, err := h.logRepo.GetLogByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Log not found",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)
	if existingLog.CreatedBy != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "You can only edit your own logs",
		})
	}

	if time.Since(existingLog.CreatedAt) > 24*time.Hour {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Cannot edit log after 24 hours",
		})
	}

	if err := h.logRepo.UpdateLog(c.Context(), id, &log); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update log",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    log,
	})
}

// ExportLogs exports logs to CSV format
// @Summary Export logs to CSV
// @Description Exports filtered logs to CSV format
// @Tags logs
// @Security BearerAuth
// @Produce text/csv
// @Param sort_by query string false "Sort by field" default(created_at)
// @Param order query string false "Sort order (asc, desc)" default(desc)
// @Param station_id query string false "Filter by station ID (UUID)"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter until date (YYYY-MM-DD)"
// @Success 200 {file} file "CSV file with logs"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /logs/export [get]
func (h *LogHandler) ExportLogs(c *fiber.Ctx) error {
	// Similar to GetLogs but export to CSV or PDF
	// For simplicity, implement CSV export
	sortBy := c.Query("sort_by", "created_at")
	order := c.Query("order", "desc")
	stationIDStr := c.Query("station_id")
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")

	filters := make(map[string]interface{})
	if stationIDStr != "" {
		stationID, err := uuid.Parse(stationIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid station_id",
			})
		}
		filters["station_id"] = stationID
	}
	if dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid date_from",
			})
		}
		filters["date_from"] = dateFrom
	}
	if dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid date_to",
			})
		}
		filters["date_to"] = dateTo
	}

	logs, err := h.logRepo.GetLogs(c.Context(), filters, sortBy, order, 0, 0) // No limit for export
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get logs",
		})
	}

	// Generate CSV
	var csvData strings.Builder
	writer := csv.NewWriter(&csvData)
	writer.Write([]string{"ID", "Date", "Time", "Station ID", "Operator", "Action", "Event", "Created By", "Created At"})
	for _, log := range logs {
		writer.Write([]string{
			log.ID.String(),
			log.LogDate.Format("2006-01-02"),
			log.LogTime,
			log.StationID.String(),
			log.OperatorName,
			log.Action,
			log.Event,
			log.CreatedBy.String(),
			log.CreatedAt.Format(time.RFC3339),
		})
	}
	writer.Flush()

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=logs.csv")
	return c.SendString(csvData.String())
}

// GetStations retrieves all stations
// @Summary Get all stations
// @Description Retrieves a list of all stations
// @Tags stations
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "List of stations"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /stations [get]
func (h *StationHandler) GetStations(c *fiber.Ctx) error {
	stations, err := h.stationRepo.GetStations(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get stations",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stations,
	})
}

// CreateStation creates a new station
// @Summary Create station
// @Description Creates a new station (admin only)
// @Tags stations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param station body models.Station true "Station data"
// @Success 201 {object} map[string]interface{} "Created station"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden (admin only)"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /stations [post]
func (h *StationHandler) CreateStation(c *fiber.Ctx) error {
	var station models.Station
	if err := c.BodyParser(&station); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if err := h.stationRepo.CreateStation(c.Context(), &station); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create station",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    station,
	})
}

// GetDevices retrieves all devices
// @Summary Get all devices
// @Description Retrieves a list of all registered devices (admin only)
// @Tags devices
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "List of devices"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden (admin only)"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /devices [get]
func (h *DeviceHandler) GetDevices(c *fiber.Ctx) error {
	devices, err := h.deviceRepo.GetDevices(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get devices",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    devices,
	})
}

// CreateDevice registers a new device
// @Summary Create device
// @Description Registers a new device with device fingerprint (admin only)
// @Tags devices
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param device body models.Device true "Device data"
// @Success 201 {object} map[string]interface{} "Created device"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden (admin only)"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /devices [post]
func (h *DeviceHandler) CreateDevice(c *fiber.Ctx) error {
	var device models.Device
	if err := c.BodyParser(&device); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)
	device.RegisteredBy = userID

	if err := h.deviceRepo.CreateDevice(c.Context(), &device); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create device",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    device,
	})
}

// DeactivateDevice deactivates a device
// @Summary Deactivate device
// @Description Deactivates a device by ID (admin only)
// @Tags devices
// @Security BearerAuth
// @Produce json
// @Param id path string true "Device ID (UUID)"
// @Success 200 {object} map[string]interface{} "Device deactivated"
// @Failure 400 {object} map[string]interface{} "Invalid device ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden (admin only)"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /devices/{id} [delete]
func (h *DeviceHandler) DeactivateDevice(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid device ID",
		})
	}

	if err := h.deviceRepo.DeactivateDevice(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to deactivate device",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    nil,
	})
}

// GetUsers retrieves all users
// @Summary Get all users
// @Description Retrieves a list of all users (admin only)
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "List of users"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden (admin only)"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users [get]
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	users, err := h.userRepo.GetUsers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get users",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    users,
	})
}

// CreateUser creates a new user
// @Summary Create user
// @Description Creates a new user with email and password (admin only)
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user body map[string]string true "User data" example({"name":"John Doe","email":"john@example.com","password":"password123","role":"operator"})
// @Success 201 {object} map[string]interface{} "Created user"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden (admin only)"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	type CreateUserRequest struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to hash password",
		})
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         req.Role,
	}

	if err := h.userRepo.CreateUser(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

// CreateAdminUser creates the initial admin user (SETUP ONLY - COMMENT OUT AFTER FIRST ADMIN IS CREATED)
// @Summary Create admin user (SETUP ONLY)
// @Description Creates the initial admin user without authentication. IMPORTANT: Comment out this endpoint after creating your first admin user for security reasons.
// @Tags users
// @Accept json
// @Produce json
// @Param admin body map[string]string true "Admin user data" example({"name":"Admin User","email":"admin@example.com","password":"securepassword123"})
// @Success 201 {object} map[string]interface{} "Created admin user"
// @Failure 400 {object} map[string]interface{} "Invalid request body or user already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/create-admin [post]
func (h *UserHandler) CreateAdminUser(c *fiber.Ctx) error {
	type CreateAdminRequest struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req CreateAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Name, email, and password are required",
		})
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to hash password",
		})
	}

	// Create admin user with role "admin"
	admin := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         "admin",
	}

	if err := h.userRepo.CreateUser(c.Context(), admin); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create admin user (user may already exist)",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Admin user created successfully. IMPORTANT: Comment out this endpoint in main.go after creation!",
		"data": fiber.Map{
			"id":    admin.ID,
			"name":  admin.Name,
			"email": admin.Email,
			"role":  admin.Role,
		},
	})
}

// Transcribe converts audio to text using OpenAI Whisper API
// @Summary Transcribe audio to text
// @Description Converts audio files to text using OpenAI's Whisper API
// @Tags stt
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param audio formData file true "Audio file (MP3, WAV, M4A, etc.)"
// @Success 200 {object} map[string]interface{} "Transcription result"
// @Failure 400 {object} map[string]interface{} "Missing audio file"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Failed to transcribe audio"
// @Router /stt [post]
func (h *STTHandler) Transcribe(c *fiber.Ctx) error {
	file, err := c.FormFile("audio")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Missing audio file",
		})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to open file",
		})
	}
	defer src.Close()

	transcription, err := transcribeAudio(src, file.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to transcribe audio",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"transcription": transcription,
		},
	})
}