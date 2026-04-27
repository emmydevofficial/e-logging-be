package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"e-logging-app/internal/models"
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
	query := `SELECT id, name FROM stations`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stations []*models.Station
	for rows.Next() {
		station := &models.Station{}
		err := rows.Scan(&station.ID, &station.Name)
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
	query := `INSERT INTO logs (log_date, log_time, station_id, operator_name, action, event, created_by, device_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	return r.db.Pool.QueryRow(ctx, query, log.LogDate, log.LogTime, log.StationID, log.OperatorName, log.Action, log.Event, log.CreatedBy, log.DeviceID).Scan(&log.ID, &log.CreatedAt, &log.UpdatedAt)
}

func (r *logRepository) GetLogs(ctx context.Context, filters map[string]interface{}, sortBy string, order string, limit int, offset int) ([]*models.Log, error) {
	query := `SELECT id, log_date, log_time, station_id, operator_name, action, event, created_by, created_at, updated_at, device_id FROM logs WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if stationID, ok := filters["station_id"]; ok {
		argCount++
		query += fmt.Sprintf(" AND station_id = $%d", argCount)
		args = append(args, stationID)
	}
	if dateFrom, ok := filters["date_from"]; ok {
		argCount++
		query += fmt.Sprintf(" AND log_date >= $%d", argCount)
		args = append(args, dateFrom)
	}
	if dateTo, ok := filters["date_to"]; ok {
		argCount++
		query += fmt.Sprintf(" AND log_date <= $%d", argCount)
		args = append(args, dateTo)
	}

	if sortBy == "" {
		sortBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, order)

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
		err := rows.Scan(&log.ID, &log.LogDate, &log.LogTime, &log.StationID, &log.OperatorName, &log.Action, &log.Event, &log.CreatedBy, &log.CreatedAt, &log.UpdatedAt, &log.DeviceID)
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
	query := `SELECT id, log_date, log_time, station_id, operator_name, action, event, created_by, created_at, updated_at, device_id FROM logs WHERE id = $1`
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&log.ID, &log.LogDate, &log.LogTime, &log.StationID, &log.OperatorName, &log.Action, &log.Event, &log.CreatedBy, &log.CreatedAt, &log.UpdatedAt, &log.DeviceID)
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
	var lastEntryStationID uuid.UUID
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
		var stationName string
		err = r.db.Pool.QueryRow(ctx, "SELECT name FROM stations WHERE id = $1", lastEntryStationID).Scan(&stationName)
		if err == nil {
			stats.LastEntryStation = stationName
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
		SELECT l.id, l.log_date, l.log_time, l.station_id, l.operator_name, l.action, l.event, l.created_by, l.created_at, l.updated_at, l.device_id, s.name as station_name
		FROM logs l
		JOIN stations s ON l.station_id = s.id
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
		err := recentRows.Scan(&log.ID, &log.LogDate, &log.LogTime, &log.StationID, &log.OperatorName, &log.Action, &log.Event, &log.CreatedBy, &log.CreatedAt, &log.UpdatedAt, &log.DeviceID, &log.StationName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recent log: %w", err)
		}
		recentLogs = append(recentLogs, log)
	}
	stats.RecentLogs = recentLogs

	return stats, nil
}