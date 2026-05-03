package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"e-logging-app/internal/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUsers(ctx context.Context) ([]*models.User, error)
}

type StationRepository interface {
	CreateStation(ctx context.Context, station *models.Station) error
	GetStations(ctx context.Context) ([]*models.Station, error)
}

type DeviceRepository interface {
	CreateDevice(ctx context.Context, device *models.Device) error
	GetDevices(ctx context.Context) ([]*models.Device, error)
	GetDeviceByFingerprint(ctx context.Context, fingerprint string) (*models.Device, error)
	GetDeviceByID(ctx context.Context, id uuid.UUID) (*models.Device, error)
	UpdateDevice(ctx context.Context, id uuid.UUID, update *models.DeviceUpdate) (*models.Device, error)
	DeactivateDevice(ctx context.Context, id uuid.UUID) error
}

type LogRepository interface {
	CreateLog(ctx context.Context, log *models.Log) error
	GetLogs(ctx context.Context, filters map[string]interface{}, sortBy string, order string, limit int, offset int) ([]*models.Log, error)
	UpdateLog(ctx context.Context, id uuid.UUID, log *models.Log) error
	GetLogByID(ctx context.Context, id uuid.UUID) (*models.Log, error)
	GetDashboardStats(ctx context.Context) (*models.DashboardStats, error)
}

type OperatorSessionRepository interface {
	CreateSession(ctx context.Context, session *models.OperatorSession) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*models.OperatorSession, error)
	GetActiveSessions(ctx context.Context) ([]*models.OperatorSession, error)
	EndSession(ctx context.Context, id uuid.UUID) error
	SignInOperator(ctx context.Context, signIn *models.OperatorSignIn) error
	SignOutOperator(ctx context.Context, sessionID, operatorID uuid.UUID) error
	GetSignedInOperators(ctx context.Context, sessionID uuid.UUID) ([]*models.SignedInOperator, error)
	GetOperatorCurrentSession(ctx context.Context, operatorID uuid.UUID) (*models.OperatorSession, error)
	GetActiveSessionCount(ctx context.Context, date time.Time) (int, error)
}

type ShiftSummaryRepository interface {
	CreateShiftSummary(ctx context.Context, summary *models.ShiftSummary) error
	GetShiftSummaryByID(ctx context.Context, id uuid.UUID) (*models.ShiftSummary, error)
	GetShiftSummaryBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.ShiftSummary, error)
	AddGenerationSummary(ctx context.Context, genSummary *models.GenerationSummary) error
	GetGenerationSummaries(ctx context.Context, summaryID uuid.UUID) ([]*models.GenerationSummary, error)
	AddNoteSummary(ctx context.Context, noteSummary *models.NoteSummary) error
	GetNoteSummary(ctx context.Context, summaryID uuid.UUID) (*models.NoteSummary, error)
}

type userRepository struct {
	db *Database
}

type stationRepository struct {
	db *Database
}

type deviceRepository struct {
	db *Database
}

type logRepository struct {
	db *Database
}

type operatorSessionRepository struct {
	db *Database
}

type shiftSummaryRepository struct {
	db *Database
}

func NewUserRepository(db *Database) UserRepository {
	return &userRepository{db: db}
}

func NewStationRepository(db *Database) StationRepository {
	return &stationRepository{db: db}
}

func NewDeviceRepository(db *Database) DeviceRepository {
	return &deviceRepository{db: db}
}

func NewLogRepository(db *Database) LogRepository {
	return &logRepository{db: db}
}

func NewOperatorSessionRepository(db *Database) OperatorSessionRepository {
	return &operatorSessionRepository{db: db}
}

func NewShiftSummaryRepository(db *Database) ShiftSummaryRepository {
	return &shiftSummaryRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return r.db.Pool.QueryRow(ctx, query, user.Name, user.Email, user.PasswordHash, user.Role).Scan(&user.ID, &user.CreatedAt)
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = $1`
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, name, email, password_hash, role, created_at FROM users WHERE id = $1`
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetUsers(ctx context.Context) ([]*models.User, error) {
	query := `SELECT id, name, email, role, created_at FROM users`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *stationRepository) CreateStation(ctx context.Context, station *models.Station) error {
	query := `INSERT INTO stations (name) VALUES ($1) RETURNING id`
	return r.db.Pool.QueryRow(ctx, query, station.Name).Scan(&station.ID)
}

func (r *stationRepository) GetStations(ctx context.Context) ([]*models.Station, error) {
	query := `SELECT id, name, station_type FROM stations`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stations []*models.Station
	for rows.Next() {
		station := &models.Station{}
		err := rows.Scan(&station.ID, &station.Name, &station.StationType)
		if err != nil {
			return nil, err
		}
		stations = append(stations, station)
	}
	return stations, nil
}

func (r *deviceRepository) CreateDevice(ctx context.Context, device *models.Device) error {
	query := `INSERT INTO devices (device_name, fingerprint, registered_by, can_log, can_download, is_active) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, registered_at, can_log, can_download, is_active`
	return r.db.Pool.QueryRow(ctx, query, device.DeviceName, device.Fingerprint, device.RegisteredBy, device.CanLog, device.CanDownload, device.IsActive).Scan(&device.ID, &device.RegisteredAt, &device.CanLog, &device.CanDownload, &device.IsActive)
}

func (r *deviceRepository) GetDevices(ctx context.Context) ([]*models.Device, error) {
	query := `SELECT id, device_name, fingerprint, registered_by, registered_at, can_log, can_download, is_active FROM devices`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		device := &models.Device{}
		err := rows.Scan(&device.ID, &device.DeviceName, &device.Fingerprint, &device.RegisteredBy, &device.RegisteredAt, &device.CanLog, &device.CanDownload, &device.IsActive)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	return devices, nil
}

func (r *deviceRepository) GetDeviceByFingerprint(ctx context.Context, fingerprint string) (*models.Device, error) {
	device := &models.Device{}
	query := `SELECT id, device_name, fingerprint, registered_by, registered_at, can_log, can_download, is_active FROM devices WHERE fingerprint = $1`
	err := r.db.Pool.QueryRow(ctx, query, fingerprint).Scan(&device.ID, &device.DeviceName, &device.Fingerprint, &device.RegisteredBy, &device.RegisteredAt, &device.CanLog, &device.CanDownload, &device.IsActive)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (r *deviceRepository) GetDeviceByID(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	device := &models.Device{}
	query := `SELECT id, device_name, fingerprint, registered_by, registered_at, can_log, can_download, is_active FROM devices WHERE id = $1`
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&device.ID, &device.DeviceName, &device.Fingerprint, &device.RegisteredBy, &device.RegisteredAt, &device.CanLog, &device.CanDownload, &device.IsActive)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (r *deviceRepository) UpdateDevice(ctx context.Context, id uuid.UUID, update *models.DeviceUpdate) (*models.Device, error) {
	if update == nil {
		return r.GetDeviceByID(ctx, id)
	}

	fields := []string{}
	args := []interface{}{}
	argPos := 1

	if update.DeviceName != nil {
		fields = append(fields, fmt.Sprintf("device_name = $%d", argPos))
		args = append(args, *update.DeviceName)
		argPos++
	}
	if update.CanLog != nil {
		fields = append(fields, fmt.Sprintf("can_log = $%d", argPos))
		args = append(args, *update.CanLog)
		argPos++
	}
	if update.CanDownload != nil {
		fields = append(fields, fmt.Sprintf("can_download = $%d", argPos))
		args = append(args, *update.CanDownload)
		argPos++
	}
	if update.IsActive != nil {
		fields = append(fields, fmt.Sprintf("is_active = $%d", argPos))
		args = append(args, *update.IsActive)
		argPos++
	}

	if len(fields) == 0 {
		return r.GetDeviceByID(ctx, id)
	}

	query := fmt.Sprintf("UPDATE devices SET %s WHERE id = $%d RETURNING id, device_name, fingerprint, registered_by, registered_at, can_log, can_download, is_active", strings.Join(fields, ", "), argPos)
	args = append(args, id)

	device := &models.Device{}
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&device.ID, &device.DeviceName, &device.Fingerprint, &device.RegisteredBy, &device.RegisteredAt, &device.CanLog, &device.CanDownload, &device.IsActive)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (r *deviceRepository) DeactivateDevice(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE devices SET is_active = false WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *logRepository) CreateLog(ctx context.Context, log *models.Log) error {
	var stationID interface{} = log.StationID
	if log.StationID == uuid.Nil {
		stationID = nil
	}
	var deviceID interface{} = log.DeviceID
	if log.DeviceID == uuid.Nil {
		deviceID = nil
	}
	
	query := `INSERT INTO logs (log_date, log_time, station_id, operator_name, action, event, created_by, device_id, event_type, session_id, is_summary, shift_summary_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id, created_at, updated_at`
	return r.db.Pool.QueryRow(ctx, query, log.LogDate, log.LogTime, stationID, log.OperatorName, log.Action, log.Event, log.CreatedBy, deviceID, log.EventType, log.SessionID, log.IsSummary, log.ShiftSummaryID).Scan(&log.ID, &log.CreatedAt, &log.UpdatedAt)
}

func (r *logRepository) GetLogs(ctx context.Context, filters map[string]interface{}, sortBy string, order string, limit int, offset int) ([]*models.Log, error) {
	query := `SELECT l.id, l.log_date, l.log_time, COALESCE(l.station_id, '00000000-0000-0000-0000-000000000000'::uuid), l.operator_name, l.action, l.event, l.created_by, l.created_at, l.updated_at, COALESCE(l.device_id, '00000000-0000-0000-0000-000000000000'::uuid), COALESCE(s.name, ''), COALESCE(u.name, ''), l.event_type, l.session_id, COALESCE(l.is_summary, false), l.shift_summary_id
	          FROM logs l
	          LEFT JOIN stations s ON l.station_id = s.id
	          LEFT JOIN users u ON l.created_by = u.id
	          WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if stationID, ok := filters["station_id"]; ok {
		argCount++
		query += fmt.Sprintf(" AND l.station_id = $%d", argCount)
		args = append(args, stationID)
	}
	if dateFrom, ok := filters["date_from"]; ok {
		argCount++
		query += fmt.Sprintf(" AND l.log_date >= $%d", argCount)
		args = append(args, dateFrom)
	}
	if dateTo, ok := filters["date_to"]; ok {
		argCount++
		query += fmt.Sprintf(" AND l.log_date <= $%d", argCount)
		args = append(args, dateTo)
	}
	if hour, ok := filters["hour"]; ok {
		argCount++
		query += fmt.Sprintf(" AND EXTRACT(HOUR FROM l.created_at) = $%d", argCount)
		args = append(args, hour)
	}
	if timeFrom, ok := filters["time_from"]; ok {
		argCount++
		query += fmt.Sprintf(" AND EXTRACT(HOUR FROM l.created_at) >= $%d", argCount)
		args = append(args, timeFrom)
	}
	if timeTo, ok := filters["time_to"]; ok {
		argCount++
		query += fmt.Sprintf(" AND EXTRACT(HOUR FROM l.created_at) <= $%d", argCount)
		args = append(args, timeTo)
	}

	if sortBy == "" {
		sortBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}
	query += fmt.Sprintf(" ORDER BY l.%s %s", sortBy, order)

	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}
	if offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.Log
	for rows.Next() {
		log := &models.Log{}
		err := rows.Scan(&log.ID, &log.LogDate, &log.LogTime, &log.StationID, &log.OperatorName, &log.Action, &log.Event, &log.CreatedBy, &log.CreatedAt, &log.UpdatedAt, &log.DeviceID, &log.StationName, &log.UserName, &log.EventType, &log.SessionID, &log.IsSummary, &log.ShiftSummaryID)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *logRepository) UpdateLog(ctx context.Context, id uuid.UUID, log *models.Log) error {
	query := `UPDATE logs SET log_date = $1, log_time = $2, station_id = $3, operator_name = $4, action = $5, event = $6, updated_at = NOW() WHERE id = $7`
	_, err := r.db.Pool.Exec(ctx, query, log.LogDate, log.LogTime, log.StationID, log.OperatorName, log.Action, log.Event, id)
	return err
}

func (r *logRepository) GetLogByID(ctx context.Context, id uuid.UUID) (*models.Log, error) {
	log := &models.Log{}
	query := `SELECT id, log_date, log_time, station_id, operator_name, action, event, created_by, created_at, updated_at, device_id, event_type, session_id FROM logs WHERE id = $1`
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&log.ID, &log.LogDate, &log.LogTime, &log.StationID, &log.OperatorName, &log.Action, &log.Event, &log.CreatedBy, &log.CreatedAt, &log.UpdatedAt, &log.DeviceID, &log.EventType, &log.SessionID)
	if err != nil {
		return nil, err
	}
	return log, nil
}

func (r *logRepository) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	stats := &models.DashboardStats{
		ActionBreakdown: make(map[string]int),
		StationActivity: []models.StationActivity{},
	}

	// Get today's date
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// Total logs today
	err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM logs WHERE DATE(created_at) = $1", today).Scan(&stats.TotalLogsToday)
	if err != nil {
		return nil, fmt.Errorf("failed to get total logs today: %w", err)
	}

	// Logs yesterday
	err = r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM logs WHERE DATE(created_at) = $1", yesterday).Scan(&stats.LogsYesterday)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs yesterday: %w", err)
	}

	// Total stations
	err = r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM stations").Scan(&stats.TotalStations)
	if err != nil {
		return nil, fmt.Errorf("failed to get total stations: %w", err)
	}

	// Active stations (stations with logs today)
	err = r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT station_id)
		FROM logs
		WHERE DATE(created_at) = $1
	`, today).Scan(&stats.ActiveStations)
	if err != nil {
		return nil, fmt.Errorf("failed to get active stations: %w", err)
	}

	// Operators on duty (unique operators with logs today)
	err = r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT created_by)
		FROM logs
		WHERE DATE(created_at) = $1
	`, today).Scan(&stats.OperatorsOnDuty)
	if err != nil {
		return nil, fmt.Errorf("failed to get operators on duty: %w", err)
	}

	// Last entry info
	var lastEntryTime time.Time
	var lastEntryStationID *uuid.UUID
	var lastEntryOperator string
	err = r.db.Pool.QueryRow(ctx, `
		SELECT created_at, station_id, operator_name
		FROM logs
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(&lastEntryTime, &lastEntryStationID, &lastEntryOperator)
	if err == nil {
		stats.LastEntryTime = lastEntryTime.Format("15:04")
		stats.LastEntryOperator = lastEntryOperator

		// Get station name
		if lastEntryStationID != nil {
			var stationName string
			err = r.db.Pool.QueryRow(ctx, "SELECT name FROM stations WHERE id = $1", *lastEntryStationID).Scan(&stationName)
			if err == nil {
				stats.LastEntryStation = stationName
			}
		}
	}

	// Action breakdown
	actionRows, err := r.db.Pool.Query(ctx, `
		SELECT action, COUNT(*) as count
		FROM logs
		WHERE DATE(created_at) = $1
		GROUP BY action
	`, today)
	if err != nil {
		return nil, fmt.Errorf("failed to get action breakdown: %w", err)
	}
	defer actionRows.Close()

	for actionRows.Next() {
		var action string
		var count int
		err := actionRows.Scan(&action, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action breakdown: %w", err)
		}
		stats.ActionBreakdown[action] = count
	}

	// Station activity
	stationRows, err := r.db.Pool.Query(ctx, `
		SELECT s.name, COUNT(l.id) as log_count
		FROM stations s
		LEFT JOIN logs l ON s.id = l.station_id AND DATE(l.created_at) = $1
		GROUP BY s.id, s.name
		ORDER BY log_count DESC
	`, today)
	if err != nil {
		return nil, fmt.Errorf("failed to get station activity: %w", err)
	}
	defer stationRows.Close()

	for stationRows.Next() {
		var stationName string
		var logCount int
		err := stationRows.Scan(&stationName, &logCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan station activity: %w", err)
		}

		status := "low"
		if logCount >= 10 {
			status = "high"
		} else if logCount >= 5 {
			status = "medium"
		}

		stats.StationActivity = append(stats.StationActivity, models.StationActivity{
			StationName: stationName,
			LogCount:    logCount,
			Status:      status,
		})
	}

	// Recent logs (last 10)
	recentRows, err := r.db.Pool.Query(ctx, `
		SELECT l.id, l.log_date, l.log_time, COALESCE(l.station_id, '00000000-0000-0000-0000-000000000000'::uuid), l.operator_name, l.action, l.event, l.created_by, l.created_at, l.updated_at, COALESCE(l.device_id, '00000000-0000-0000-0000-000000000000'::uuid), COALESCE(s.name, ''), COALESCE(u.name, ''), l.event_type, l.session_id
		FROM logs l
		LEFT JOIN stations s ON l.station_id = s.id
		LEFT JOIN users u ON l.created_by = u.id
		ORDER BY l.created_at DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent logs: %w", err)
	}
	defer recentRows.Close()

	var recentLogs []*models.Log
	for recentRows.Next() {
		log := &models.Log{}
		err := recentRows.Scan(&log.ID, &log.LogDate, &log.LogTime, &log.StationID, &log.OperatorName, &log.Action, &log.Event, &log.CreatedBy, &log.CreatedAt, &log.UpdatedAt, &log.DeviceID, &log.StationName, &log.UserName, &log.EventType, &log.SessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recent log: %w", err)
		}
		recentLogs = append(recentLogs, log)
	}
	stats.RecentLogs = recentLogs

	return stats, nil
}

// OperatorSessionRepository implementation
func (r *operatorSessionRepository) CreateSession(ctx context.Context, session *models.OperatorSession) error {
	query := `INSERT INTO operator_sessions (shift_lead_id, start_time, is_active, max_sign_ins) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	return r.db.Pool.QueryRow(ctx, query, session.ShiftLeadID, session.StartTime, session.IsActive, session.MaxSignIns).Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)
}

func (r *operatorSessionRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.OperatorSession, error) {
	session := &models.OperatorSession{}
	query := `SELECT id, shift_lead_id, start_time, end_time, is_active, max_sign_ins, created_at, updated_at FROM operator_sessions WHERE id = $1`
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&session.ID, &session.ShiftLeadID, &session.StartTime, &session.EndTime, &session.IsActive, &session.MaxSignIns, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *operatorSessionRepository) GetActiveSessions(ctx context.Context) ([]*models.OperatorSession, error) {
	query := `SELECT id, shift_lead_id, start_time, end_time, is_active, max_sign_ins, created_at, updated_at FROM operator_sessions WHERE is_active = true ORDER BY start_time DESC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.OperatorSession
	for rows.Next() {
		session := &models.OperatorSession{}
		err := rows.Scan(&session.ID, &session.ShiftLeadID, &session.StartTime, &session.EndTime, &session.IsActive, &session.MaxSignIns, &session.CreatedAt, &session.UpdatedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (r *operatorSessionRepository) EndSession(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE operator_sessions SET end_time = NOW(), is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *operatorSessionRepository) SignInOperator(ctx context.Context, signIn *models.OperatorSignIn) error {
	query := `INSERT INTO operator_sign_ins (session_id, operator_id, signed_by_id, signed_in_at, is_active) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`
	return r.db.Pool.QueryRow(ctx, query, signIn.SessionID, signIn.OperatorID, signIn.SignedByID, signIn.SignedInAt, signIn.IsActive).Scan(&signIn.ID, &signIn.CreatedAt, &signIn.UpdatedAt)
}

func (r *operatorSessionRepository) SignOutOperator(ctx context.Context, sessionID, operatorID uuid.UUID) error {
	query := `UPDATE operator_sign_ins SET signed_out_at = NOW(), is_active = false, updated_at = NOW() WHERE session_id = $1 AND operator_id = $2 AND is_active = true`
	_, err := r.db.Pool.Exec(ctx, query, sessionID, operatorID)
	return err
}

func (r *operatorSessionRepository) GetSignedInOperators(ctx context.Context, sessionID uuid.UUID) ([]*models.SignedInOperator, error) {
	query := `SELECT u.id, u.name, si.signed_in_at
	          FROM operator_sign_ins si
	          JOIN users u ON si.operator_id = u.id
	          WHERE si.session_id = $1 AND si.is_active = true
	          ORDER BY si.signed_in_at ASC`
	rows, err := r.db.Pool.Query(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var operators []*models.SignedInOperator
	for rows.Next() {
		operator := &models.SignedInOperator{}
		err := rows.Scan(&operator.ID, &operator.Name, &operator.SignedInAt)
		if err != nil {
			return nil, err
		}
		operators = append(operators, operator)
	}
	return operators, nil
}

func (r *operatorSessionRepository) GetOperatorCurrentSession(ctx context.Context, operatorID uuid.UUID) (*models.OperatorSession, error) {
	session := &models.OperatorSession{}
	query := `SELECT os.id, os.shift_lead_id, os.start_time, os.end_time, os.is_active, os.max_sign_ins, os.created_at, os.updated_at
	          FROM operator_sessions os
	          JOIN operator_sign_ins si ON os.id = si.session_id
	          WHERE si.operator_id = $1 AND si.is_active = true AND os.is_active = true
	          LIMIT 1`
	err := r.db.Pool.QueryRow(ctx, query, operatorID).Scan(&session.ID, &session.ShiftLeadID, &session.StartTime, &session.EndTime, &session.IsActive, &session.MaxSignIns, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *operatorSessionRepository) GetActiveSessionCount(ctx context.Context, date time.Time) (int, error) {
	count := 0
	query := `SELECT COUNT(*) FROM operator_sessions WHERE DATE(start_time) = $1 AND is_active = true`
	err := r.db.Pool.QueryRow(ctx, query, date.Format("2006-01-02")).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ShiftSummaryRepository implementation
func (r *shiftSummaryRepository) CreateShiftSummary(ctx context.Context, summary *models.ShiftSummary) error {
	query := `INSERT INTO shift_summary (session_id, created_by, summary_date, summary_time, shift_note) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`
	return r.db.Pool.QueryRow(ctx, query, summary.SessionID, summary.CreatedBy, summary.SummaryDate, summary.SummaryTime, summary.ShiftNote).Scan(&summary.ID, &summary.CreatedAt, &summary.UpdatedAt)
}

func (r *shiftSummaryRepository) GetShiftSummaryByID(ctx context.Context, id uuid.UUID) (*models.ShiftSummary, error) {
	summary := &models.ShiftSummary{}
	query := `SELECT id, session_id, created_by, summary_date::text, summary_time::text, COALESCE(shift_note, ''), created_at, updated_at FROM shift_summary WHERE id = $1`
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&summary.ID, &summary.SessionID, &summary.CreatedBy, &summary.SummaryDate, &summary.SummaryTime, &summary.ShiftNote, &summary.CreatedAt, &summary.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Get associated generation summaries
	genSummaries, err := r.GetGenerationSummaries(ctx, id)
	if err == nil {
		summary.GenerationStations = genSummaries
	}

	return summary, nil
}

func (r *shiftSummaryRepository) GetShiftSummaryBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.ShiftSummary, error) {
	summary := &models.ShiftSummary{}
	query := `SELECT id, session_id, created_by, summary_date::text, summary_time::text, COALESCE(shift_note, ''), created_at, updated_at FROM shift_summary WHERE session_id = $1 LIMIT 1`
	err := r.db.Pool.QueryRow(ctx, query, sessionID).Scan(&summary.ID, &summary.SessionID, &summary.CreatedBy, &summary.SummaryDate, &summary.SummaryTime, &summary.ShiftNote, &summary.CreatedAt, &summary.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Get associated generation summaries
	genSummaries, err := r.GetGenerationSummaries(ctx, summary.ID)
	if err == nil {
		summary.GenerationStations = genSummaries
	}

	return summary, nil
}

func (r *shiftSummaryRepository) AddGenerationSummary(ctx context.Context, genSummary *models.GenerationSummary) error {
	query := `INSERT INTO generation_summary (shift_summary_id, station_id, running_units, reserve_energy_mw) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	return r.db.Pool.QueryRow(ctx, query, genSummary.ShiftSummaryID, genSummary.StationID, genSummary.RunningUnits, genSummary.ReserveEnergyMW).Scan(&genSummary.ID, &genSummary.CreatedAt, &genSummary.UpdatedAt)
}

func (r *shiftSummaryRepository) GetGenerationSummaries(ctx context.Context, summaryID uuid.UUID) ([]*models.GenerationSummary, error) {
	query := `SELECT gs.id, gs.shift_summary_id, gs.station_id, s.name, gs.running_units, gs.reserve_energy_mw, gs.created_at, gs.updated_at
	          FROM generation_summary gs
	          JOIN stations s ON gs.station_id = s.id
	          WHERE gs.shift_summary_id = $1
	          ORDER BY s.name ASC`
	rows, err := r.db.Pool.Query(ctx, query, summaryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*models.GenerationSummary
	for rows.Next() {
		genSummary := &models.GenerationSummary{}
		err := rows.Scan(&genSummary.ID, &genSummary.ShiftSummaryID, &genSummary.StationID, &genSummary.StationName, &genSummary.RunningUnits, &genSummary.ReserveEnergyMW, &genSummary.CreatedAt, &genSummary.UpdatedAt)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, genSummary)
	}
	return summaries, nil
}

func (r *shiftSummaryRepository) AddNoteSummary(ctx context.Context, noteSummary *models.NoteSummary) error {
	query := `INSERT INTO note_summary (shift_summary_id, note_text) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	return r.db.Pool.QueryRow(ctx, query, noteSummary.ShiftSummaryID, noteSummary.NoteText).Scan(&noteSummary.ID, &noteSummary.CreatedAt, &noteSummary.UpdatedAt)
}

func (r *shiftSummaryRepository) GetNoteSummary(ctx context.Context, summaryID uuid.UUID) (*models.NoteSummary, error) {
	noteSummary := &models.NoteSummary{}
	query := `SELECT id, shift_summary_id, note_text, created_at, updated_at FROM note_summary WHERE shift_summary_id = $1 LIMIT 1`
	err := r.db.Pool.QueryRow(ctx, query, summaryID).Scan(&noteSummary.ID, &noteSummary.ShiftSummaryID, &noteSummary.NoteText, &noteSummary.CreatedAt, &noteSummary.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return noteSummary, nil
}
