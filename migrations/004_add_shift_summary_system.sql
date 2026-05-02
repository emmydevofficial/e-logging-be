-- Migration: 004_add_shift_summary_system.sql
-- Adds shift summary system with generation station tracking

-- Add station_type column to stations table
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'stations' AND column_name = 'station_type') THEN
        ALTER TABLE stations ADD COLUMN station_type VARCHAR(50) CHECK (station_type IN ('Generation', 'Transmission', 'Distribution'));
    END IF;
END $$;

-- Create shift_summary table
CREATE TABLE IF NOT EXISTS shift_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES operator_sessions(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id),
    summary_date DATE NOT NULL,
    summary_time TIME NOT NULL,
    shift_note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create generation_summary table (for each generating station in the shift summary)
CREATE TABLE IF NOT EXISTS generation_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shift_summary_id UUID NOT NULL REFERENCES shift_summary(id) ON DELETE CASCADE,
    station_id UUID NOT NULL REFERENCES stations(id),
    running_units INTEGER NOT NULL DEFAULT 0,
    reserve_energy_mw DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(shift_summary_id, station_id)
);

-- Create note_summary table (stores shift notes separately for better management)
CREATE TABLE IF NOT EXISTS note_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shift_summary_id UUID NOT NULL REFERENCES shift_summary(id) ON DELETE CASCADE,
    note_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add columns to logs table for shift summary tracking
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'logs' AND column_name = 'is_summary') THEN
        ALTER TABLE logs ADD COLUMN is_summary BOOLEAN DEFAULT FALSE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'logs' AND column_name = 'shift_summary_id') THEN
        ALTER TABLE logs ADD COLUMN shift_summary_id UUID REFERENCES shift_summary(id);
    END IF;
END $$;

-- Add unique constraint for only one active session per shift_lead
CREATE TABLE IF NOT EXISTS active_sessions_constraint (
    shift_date DATE NOT NULL,
    session_id UUID NOT NULL REFERENCES operator_sessions(id),
    CONSTRAINT one_active_session_per_day UNIQUE (shift_date, session_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_shift_summary_session_id ON shift_summary(session_id);
CREATE INDEX IF NOT EXISTS idx_shift_summary_created_by ON shift_summary(created_by);
CREATE INDEX IF NOT EXISTS idx_generation_summary_shift_id ON generation_summary(shift_summary_id);
CREATE INDEX IF NOT EXISTS idx_generation_summary_station_id ON generation_summary(station_id);
CREATE INDEX IF NOT EXISTS idx_note_summary_shift_id ON note_summary(shift_summary_id);
CREATE INDEX IF NOT EXISTS idx_logs_is_summary ON logs(is_summary);
CREATE INDEX IF NOT EXISTS idx_logs_shift_summary_id ON logs(shift_summary_id);

-- Update triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_shift_summary_updated_at ON shift_summary;
CREATE TRIGGER update_shift_summary_updated_at BEFORE UPDATE ON shift_summary FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_generation_summary_updated_at ON generation_summary;
CREATE TRIGGER update_generation_summary_updated_at BEFORE UPDATE ON generation_summary FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_note_summary_updated_at ON note_summary;
CREATE TRIGGER update_note_summary_updated_at BEFORE UPDATE ON note_summary FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
