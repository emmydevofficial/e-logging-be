package handlers

import (
	"context"
	"time"

	"e-logging-app/internal/db"
	"e-logging-app/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ShiftSummaryHandler struct {
	shiftSummaryRepo db.ShiftSummaryRepository
	logRepo          db.LogRepository
	stationRepo      db.StationRepository
	userRepo         db.UserRepository
}

func NewShiftSummaryHandler(shiftSummaryRepo db.ShiftSummaryRepository, logRepo db.LogRepository, stationRepo db.StationRepository, userRepo db.UserRepository) *ShiftSummaryHandler {
	return &ShiftSummaryHandler{
		shiftSummaryRepo: shiftSummaryRepo,
		logRepo:          logRepo,
		stationRepo:      stationRepo,
		userRepo:         userRepo,
	}
}

type CreateShiftSummaryRequest struct {
	SessionID           string                     `json:"session_id"`
	SummaryDate         string                     `json:"summary_date"`
	SummaryTime         string                     `json:"summary_time"`
	ShiftNote           string                     `json:"shift_note"`
	GenerationSummaries []GenerationSummaryRequest `json:"generation_summaries"`
}

type GenerationSummaryRequest struct {
	StationID       string  `json:"station_id"`
	RunningUnits    int     `json:"running_units"`
	ReserveEnergyMW float64 `json:"reserve_energy_mw"`
}

type ShiftSummaryResponse struct {
	ID                 string                      `json:"id"`
	SessionID          string                      `json:"session_id"`
	CreatedBy          string                      `json:"created_by"`
	CreatedByName      string                      `json:"created_by_name"`
	SummaryDate        string                      `json:"summary_date"`
	SummaryTime        string                      `json:"summary_time"`
	ShiftNote          string                      `json:"shift_note"`
	GenerationStations []GenerationStationResponse `json:"generation_stations"`
	CreatedAt          time.Time                   `json:"created_at"`
	UpdatedAt          time.Time                   `json:"updated_at"`
}

type GenerationStationResponse struct {
	StationID       string  `json:"station_id"`
	StationName     string  `json:"station_name"`
	RunningUnits    int     `json:"running_units"`
	ReserveEnergyMW float64 `json:"reserve_energy_mw"`
}

type GenerationStationListResponse struct {
	StationID   string `json:"station_id"`
	StationName string `json:"station_name"`
}

// CreateShiftSummary creates a new shift summary with generation station details
// @Summary Create shift summary
// @Description Create a shift summary with generation station running units and reserve energy
// @Tags Shift Summary
// @Accept json
// @Produce json
// @Param request body CreateShiftSummaryRequest true "Shift summary details"
// @Success 201 {object} ShiftSummaryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /shift-summary [post]
func (h *ShiftSummaryHandler) CreateShiftSummary(c *fiber.Ctx) error {
	req := &CreateShiftSummaryRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Parse session ID
	sessionID, err := uuid.Parse(req.SessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session_id format"})
	}

	// Get user ID from JWT
	userID := c.Locals("user_id").(uuid.UUID)

	// Create shift summary
	summary := &models.ShiftSummary{
		ID:          uuid.New(),
		SessionID:   sessionID,
		CreatedBy:   userID,
		SummaryDate: req.SummaryDate,
		SummaryTime: req.SummaryTime,
		ShiftNote:   req.ShiftNote,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	ctx := context.Background()

	// Create summary in database
	if err := h.shiftSummaryRepo.CreateShiftSummary(ctx, summary); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create shift summary"})
	}

	// Add generation summaries
	for _, genReq := range req.GenerationSummaries {
		stationID, err := uuid.Parse(genReq.StationID)
		if err != nil {
			continue
		}

		genSummary := &models.GenerationSummary{
			ID:              uuid.New(),
			ShiftSummaryID:  summary.ID,
			StationID:       stationID,
			RunningUnits:    genReq.RunningUnits,
			ReserveEnergyMW: genReq.ReserveEnergyMW,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if err := h.shiftSummaryRepo.AddGenerationSummary(ctx, genSummary); err != nil {
			continue
		}
	}

	logDate, _ := time.Parse("2006-01-02", req.SummaryDate)
	if logDate.IsZero() {
		logDate = time.Now()
	}

	// Create audit log entry
	auditLog := &models.Log{
		ID:             uuid.New(),
		LogDate:        logDate,
		LogTime:        req.SummaryTime,
		StationID:      uuid.Nil,
		OperatorName:   "Shift Summary",
		Action:         "create_summary",
		Event:          "Shift summary created",
		CreatedBy:      userID,
		DeviceID:       uuid.Nil,
		EventType:      "shift_summary",
		SessionID:      &sessionID,
		IsSummary:      true,
		ShiftSummaryID: &summary.ID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	h.logRepo.CreateLog(ctx, auditLog)

	// Get creator's name
	creator, _ := h.userRepo.GetUserByID(ctx, userID)
	creatorName := ""
	if creator != nil {
		creatorName = creator.Name
	}

	// Get generation summaries for response
	genStations, _ := h.shiftSummaryRepo.GetGenerationSummaries(ctx, summary.ID)

	allStations, _ := h.stationRepo.GetStations(ctx)
	savedGenMap := make(map[string]*models.GenerationSummary)
	if genStations != nil {
		for _, gen := range genStations {
			savedGenMap[gen.StationID.String()] = gen
		}
	}

	genResponse := make([]GenerationStationResponse, 0)
	if allStations != nil {
		for _, station := range allStations {
			if station.StationType == "Generation" {
				savedGen, exists := savedGenMap[station.ID.String()]
				if exists {
					genResponse = append(genResponse, GenerationStationResponse{
						StationID:       station.ID.String(),
						StationName:     station.Name,
						RunningUnits:    savedGen.RunningUnits,
						ReserveEnergyMW: savedGen.ReserveEnergyMW,
					})
				} else {
					genResponse = append(genResponse, GenerationStationResponse{
						StationID:       station.ID.String(),
						StationName:     station.Name,
						RunningUnits:    0,
						ReserveEnergyMW: 0.0,
					})
				}
			}
		}
	}

	response := &ShiftSummaryResponse{
		ID:                 summary.ID.String(),
		SessionID:          summary.SessionID.String(),
		CreatedBy:          summary.CreatedBy.String(),
		CreatedByName:      creatorName,
		SummaryDate:        summary.SummaryDate,
		SummaryTime:        summary.SummaryTime,
		ShiftNote:          summary.ShiftNote,
		GenerationStations: genResponse,
		CreatedAt:          summary.CreatedAt,
		UpdatedAt:          summary.UpdatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// GetShiftSummary retrieves a shift summary by ID
// @Summary Get shift summary
// @Description Get a shift summary with all generation station details
// @Tags Shift Summary
// @Accept json
// @Produce json
// @Param id path string true "Summary ID"
// @Success 200 {object} ShiftSummaryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /shift-summary/{id} [get]
func (h *ShiftSummaryHandler) GetShiftSummary(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	ctx := context.Background()
	summary, err := h.shiftSummaryRepo.GetShiftSummaryByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shift summary not found"})
	}

	// Get creator's name
	creator, _ := h.userRepo.GetUserByID(ctx, summary.CreatedBy)
	creatorName := ""
	if creator != nil {
		creatorName = creator.Name
	}

	// Get generation summaries for response
	genStations, _ := h.shiftSummaryRepo.GetGenerationSummaries(ctx, summary.ID)

	allStations, _ := h.stationRepo.GetStations(ctx)
	savedGenMap := make(map[string]*models.GenerationSummary)
	if genStations != nil {
		for _, gen := range genStations {
			savedGenMap[gen.StationID.String()] = gen
		}
	}

	genResponse := make([]GenerationStationResponse, 0)
	if allStations != nil {
		for _, station := range allStations {
			if station.StationType == "Generation" {
				savedGen, exists := savedGenMap[station.ID.String()]
				if exists {
					genResponse = append(genResponse, GenerationStationResponse{
						StationID:       station.ID.String(),
						StationName:     station.Name,
						RunningUnits:    savedGen.RunningUnits,
						ReserveEnergyMW: savedGen.ReserveEnergyMW,
					})
				} else {
					genResponse = append(genResponse, GenerationStationResponse{
						StationID:       station.ID.String(),
						StationName:     station.Name,
						RunningUnits:    0,
						ReserveEnergyMW: 0.0,
					})
				}
			}
		}
	}

	response := &ShiftSummaryResponse{
		ID:                 summary.ID.String(),
		SessionID:          summary.SessionID.String(),
		CreatedBy:          summary.CreatedBy.String(),
		CreatedByName:      creatorName,
		SummaryDate:        summary.SummaryDate,
		SummaryTime:        summary.SummaryTime,
		ShiftNote:          summary.ShiftNote,
		GenerationStations: genResponse,
		CreatedAt:          summary.CreatedAt,
		UpdatedAt:          summary.UpdatedAt,
	}

	return c.JSON(response)
}

// GetSessionShiftSummary retrieves the shift summary for a given session
// @Summary Get session shift summary
// @Description Get the shift summary associated with a specific operator session
// @Tags Shift Summary
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} ShiftSummaryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /shift-summary/session/{sessionId} [get]
func (h *ShiftSummaryHandler) GetSessionShiftSummary(c *fiber.Ctx) error {
	sessionIDStr := c.Params("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid session_id format"})
	}

	ctx := context.Background()
	summary, err := h.shiftSummaryRepo.GetShiftSummaryBySessionID(ctx, sessionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shift summary not found for this session"})
	}

	// Get creator's name
	creator, _ := h.userRepo.GetUserByID(ctx, summary.CreatedBy)
	creatorName := ""
	if creator != nil {
		creatorName = creator.Name
	}

	// Get generation summaries for response
	genStations, _ := h.shiftSummaryRepo.GetGenerationSummaries(ctx, summary.ID)

	allStations, _ := h.stationRepo.GetStations(ctx)
	savedGenMap := make(map[string]*models.GenerationSummary)
	if genStations != nil {
		for _, gen := range genStations {
			savedGenMap[gen.StationID.String()] = gen
		}
	}

	genResponse := make([]GenerationStationResponse, 0)
	if allStations != nil {
		for _, station := range allStations {
			if station.StationType == "Generation" {
				savedGen, exists := savedGenMap[station.ID.String()]
				if exists {
					genResponse = append(genResponse, GenerationStationResponse{
						StationID:       station.ID.String(),
						StationName:     station.Name,
						RunningUnits:    savedGen.RunningUnits,
						ReserveEnergyMW: savedGen.ReserveEnergyMW,
					})
				} else {
					genResponse = append(genResponse, GenerationStationResponse{
						StationID:       station.ID.String(),
						StationName:     station.Name,
						RunningUnits:    0,
						ReserveEnergyMW: 0.0,
					})
				}
			}
		}
	}

	response := &ShiftSummaryResponse{
		ID:                 summary.ID.String(),
		SessionID:          summary.SessionID.String(),
		CreatedBy:          summary.CreatedBy.String(),
		CreatedByName:      creatorName,
		SummaryDate:        summary.SummaryDate,
		SummaryTime:        summary.SummaryTime,
		ShiftNote:          summary.ShiftNote,
		GenerationStations: genResponse,
		CreatedAt:          summary.CreatedAt,
		UpdatedAt:          summary.UpdatedAt,
	}

	return c.JSON(response)
}

// GetGenerationStations retrieves all generation stations for the shift summary form
// @Summary Get generation stations
// @Description Get all stations with type "Generation" for shift summary creation
// @Tags Shift Summary
// @Accept json
// @Produce json
// @Success 200 {array} GenerationStationListResponse
// @Failure 500 {object} ErrorResponse
// @Router /shift-summary/stations/generation [get]
func (h *ShiftSummaryHandler) GetGenerationStations(c *fiber.Ctx) error {
	ctx := context.Background()
	stations, err := h.stationRepo.GetStations(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve stations"})
	}

	response := make([]GenerationStationListResponse, 0)
	for _, station := range stations {
		// Only include Generation type stations
		if station.StationType == "Generation" {
			response = append(response, GenerationStationListResponse{
				StationID:   station.ID.String(),
				StationName: station.Name,
			})
		}
	}

	return c.JSON(response)
}
