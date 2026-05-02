package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	Name             string     `json:"name" db:"name"`
	Email            string     `json:"email" db:"email"`
	PasswordHash     string     `json:"-" db:"password_hash"`
	Role             string     `json:"role" db:"role"`
	IsOperator       bool       `json:"is_operator" db:"is_operator"`
	CurrentSessionID *uuid.UUID `json:"current_session_id" db:"current_session_id"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

type Station struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	StationType string    `json:"station_type" db:"station_type"`
}

type Device struct {
	ID           uuid.UUID `json:"id" db:"id"`
	DeviceName   string    `json:"device_name" db:"device_name"`
	Fingerprint  string    `json:"fingerprint" db:"fingerprint"`
	RegisteredBy uuid.UUID `json:"registered_by" db:"registered_by"`
	RegisteredAt time.Time `json:"registered_at" db:"registered_at"`
	CanLog       bool      `json:"can_log" db:"can_log"`
	CanDownload  bool      `json:"can_download" db:"can_download"`
	IsActive     bool      `json:"is_active" db:"is_active"`
}

type DeviceUpdate struct {
	DeviceName  *string `json:"device_name"`
	CanLog      *bool   `json:"can_log"`
	CanDownload *bool   `json:"can_download"`
	IsActive    *bool   `json:"is_active"`
}

type Log struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	LogDate        time.Time  `json:"log_date" db:"log_date"`
	LogTime        string     `json:"log_time" db:"log_time"`
	StationID      uuid.UUID  `json:"station_id" db:"station_id"`
	OperatorName   string     `json:"operator_name" db:"operator_name"`
	Action         string     `json:"action" db:"action"`
	Event          string     `json:"event" db:"event"`
	CreatedBy      uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	DeviceID       uuid.UUID  `json:"device_id" db:"device_id"`
	StationName    string     `json:"station_name"`
	UserName       string     `json:"user_name"`
	EventType      string     `json:"event_type" db:"event_type"`
	SessionID      *uuid.UUID `json:"session_id" db:"session_id"`
	IsSummary      bool       `json:"is_summary" db:"is_summary"`
	ShiftSummaryID *uuid.UUID `json:"shift_summary_id" db:"shift_summary_id"`
}

type DashboardStats struct {
	TotalLogsToday    int               `json:"total_logs_today"`
	LogsYesterday     int               `json:"logs_yesterday"`
	ActiveStations    int               `json:"active_stations"`
	TotalStations     int               `json:"total_stations"`
	OperatorsOnDuty   int               `json:"operators_on_duty"`
	LastEntryTime     string            `json:"last_entry_time"`
	LastEntryStation  string            `json:"last_entry_station"`
	LastEntryOperator string            `json:"last_entry_operator"`
	ActionBreakdown   map[string]int    `json:"action_breakdown"`
	StationActivity   []StationActivity `json:"station_activity"`
	RecentLogs        []*Log            `json:"recent_logs"`
}

type StationActivity struct {
	StationName string `json:"station_name"`
	LogCount    int    `json:"log_count"`
	Status      string `json:"status"` // "high", "medium", "low" based on activity
}

type OperatorSession struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ShiftLeadID uuid.UUID  `json:"shift_lead_id" db:"shift_lead_id"`
	StartTime   time.Time  `json:"start_time" db:"start_time"`
	EndTime     *time.Time `json:"end_time" db:"end_time"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	MaxSignIns  int        `json:"max_sign_ins" db:"max_sign_ins"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type OperatorSignIn struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	SessionID   uuid.UUID  `json:"session_id" db:"session_id"`
	OperatorID  uuid.UUID  `json:"operator_id" db:"operator_id"`
	SignedByID  uuid.UUID  `json:"signed_by_id" db:"signed_by_id"`
	SignedInAt  time.Time  `json:"signed_in_at" db:"signed_in_at"`
	SignedOutAt *time.Time `json:"signed_out_at" db:"signed_out_at"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type SignedInOperator struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	SignedInAt time.Time `json:"signed_in_at"`
}

type ShiftSummary struct {
	ID                 uuid.UUID            `json:"id" db:"id"`
	SessionID          uuid.UUID            `json:"session_id" db:"session_id"`
	CreatedBy          uuid.UUID            `json:"created_by" db:"created_by"`
	SummaryDate        string               `json:"summary_date" db:"summary_date"`
	SummaryTime        string               `json:"summary_time" db:"summary_time"`
	ShiftNote          string               `json:"shift_note" db:"shift_note"`
	GenerationStations []*GenerationSummary `json:"generation_stations"`
	CreatedAt          time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at" db:"updated_at"`
}

type GenerationSummary struct {
	ID              uuid.UUID `json:"id" db:"id"`
	ShiftSummaryID  uuid.UUID `json:"shift_summary_id" db:"shift_summary_id"`
	StationID       uuid.UUID `json:"station_id" db:"station_id"`
	StationName     string    `json:"station_name"`
	RunningUnits    int       `json:"running_units" db:"running_units"`
	ReserveEnergyMW float64   `json:"reserve_energy_mw" db:"reserve_energy_mw"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type NoteSummary struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ShiftSummaryID uuid.UUID `json:"shift_summary_id" db:"shift_summary_id"`
	NoteText       string    `json:"note_text" db:"note_text"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type GenerationStationSummary struct {
	StationID       uuid.UUID `json:"station_id"`
	StationName     string    `json:"station_name"`
	RunningUnits    int       `json:"running_units"`
	ReserveEnergyMW float64   `json:"reserve_energy_mw"`
}
