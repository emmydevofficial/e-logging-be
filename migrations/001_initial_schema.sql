-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'operator', 'downloader', 'viewer')),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create stations table
CREATE TABLE IF NOT EXISTS stations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL
);

-- Create devices table
CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_name TEXT NOT NULL,
    fingerprint TEXT UNIQUE NOT NULL,
    registered_by UUID NOT NULL REFERENCES users(id),
    registered_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

-- Create logs table
CREATE TABLE IF NOT EXISTS logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    log_date DATE NOT NULL,
    log_time TIME NOT NULL,
    station_id UUID NOT NULL REFERENCES stations(id),
    operator_name TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('reported_that', 'was_instructed_to', 'was_given_approval', 'miscellaneous')),
    event TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    device_id UUID NOT NULL REFERENCES devices(id)
);