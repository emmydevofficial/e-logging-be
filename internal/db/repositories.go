package db

import (
	"context"
	"fmt"

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
	DeactivateDevice(ctx context.Context, id uuid.UUID) error
}

type LogRepository interface {
	CreateLog(ctx context.Context, log *models.Log) error
	GetLogs(ctx context.Context, filters map[string]interface{}, sortBy string, order string, limit int, offset int) ([]*models.Log, error)
	UpdateLog(ctx context.Context, id uuid.UUID, log *models.Log) error
	GetLogByID(ctx context.Context, id uuid.UUID) (*models.Log, error)
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
	query := `INSERT INTO devices (device_name, fingerprint, registered_by) VALUES ($1, $2, $3) RETURNING id, registered_at`
	return r.db.Pool.QueryRow(ctx, query, device.DeviceName, device.Fingerprint, device.RegisteredBy).Scan(&device.ID, &device.RegisteredAt)
}

func (r *deviceRepository) GetDevices(ctx context.Context) ([]*models.Device, error) {
	query := `SELECT id, device_name, fingerprint, registered_by, registered_at, is_active FROM devices`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		device := &models.Device{}
		err := rows.Scan(&device.ID, &device.DeviceName, &device.Fingerprint, &device.RegisteredBy, &device.RegisteredAt, &device.IsActive)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	return devices, nil
}

func (r *deviceRepository) GetDeviceByFingerprint(ctx context.Context, fingerprint string) (*models.Device, error) {
	device := &models.Device{}
	query := `SELECT id, device_name, fingerprint, registered_by, registered_at, is_active FROM devices WHERE fingerprint = $1 AND is_active = true`
	err := r.db.Pool.QueryRow(ctx, query, fingerprint).Scan(&device.ID, &device.DeviceName, &device.Fingerprint, &device.RegisteredBy, &device.RegisteredAt, &device.IsActive)
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