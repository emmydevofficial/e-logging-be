package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Station struct {
	ID   uuid.UUID `json:"id" db:"id"`
	Name string    `json:"name" db:"name"`
}

type Device struct {
	ID           uuid.UUID `json:"id" db:"id"`
	DeviceName   string    `json:"device_name" db:"device_name"`
	Fingerprint  string    `json:"fingerprint" db:"fingerprint"`
	RegisteredBy uuid.UUID `json:"registered_by" db:"registered_by"`
	RegisteredAt time.Time `json:"registered_at" db:"registered_at"`
	IsActive     bool      `json:"is_active" db:"is_active"`
}

type Log struct {
	ID           uuid.UUID `json:"id" db:"id"`
	LogDate      time.Time `json:"log_date" db:"log_date"`
	LogTime      string    `json:"log_time" db:"log_time"`
	StationID    uuid.UUID `json:"station_id" db:"station_id"`
	OperatorName string    `json:"operator_name" db:"operator_name"`
	Action       string    `json:"action" db:"action"`
	Event        string    `json:"event" db:"event"`
	CreatedBy    uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	DeviceID     uuid.UUID `json:"device_id" db:"device_id"`
}