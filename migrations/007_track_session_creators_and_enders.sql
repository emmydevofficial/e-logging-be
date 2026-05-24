-- Migration: 007_track_session_creators_and_enders.sql
-- Adds ended_by_id tracking to operator_sessions

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'operator_sessions' AND column_name = 'ended_by_id') THEN
        ALTER TABLE operator_sessions ADD COLUMN ended_by_id UUID REFERENCES users(id);
    END IF;
END $$;
