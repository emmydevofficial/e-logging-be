-- Migration: 003_add_operator_signin_system.sql
-- Adds operator sign-in system with sessions and sign-ins tracking

-- Add columns to users table (without foreign key constraint yet)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'is_operator') THEN
        ALTER TABLE users ADD COLUMN is_operator BOOLEAN DEFAULT FALSE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'current_session_id') THEN
        ALTER TABLE users ADD COLUMN current_session_id UUID;
    END IF;
END $$;

-- Create operator_sessions table
CREATE TABLE IF NOT EXISTS operator_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shift_lead_id UUID NOT NULL REFERENCES users(id),
    start_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    end_time TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    max_sign_ins INTEGER DEFAULT 5,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create operator_sign_ins table
CREATE TABLE IF NOT EXISTS operator_sign_ins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES operator_sessions(id) ON DELETE CASCADE,
    operator_id UUID NOT NULL REFERENCES users(id),
    signed_by_id UUID NOT NULL REFERENCES users(id),
    signed_in_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    signed_out_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(session_id, operator_id)
);

-- Add foreign key constraint to users table if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE constraint_name = 'fk_users_current_session_id') THEN
        ALTER TABLE users ADD CONSTRAINT fk_users_current_session_id FOREIGN KEY (current_session_id) REFERENCES operator_sessions(id);
    END IF;
END $$;

-- Add columns to logs table
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'logs' AND column_name = 'event_type') THEN
        ALTER TABLE logs ADD COLUMN event_type VARCHAR(50);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'logs' AND column_name = 'session_id') THEN
        ALTER TABLE logs ADD COLUMN session_id UUID REFERENCES operator_sessions(id);
    END IF;
END $$;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_operator_sessions_is_active ON operator_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_operator_sign_ins_session_id_active ON operator_sign_ins(session_id, is_active);
CREATE INDEX IF NOT EXISTS idx_users_current_session_id ON users(current_session_id);
CREATE INDEX IF NOT EXISTS idx_logs_event_type ON logs(event_type);
CREATE INDEX IF NOT EXISTS idx_logs_session_id ON logs(session_id);

-- Update trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_operator_sessions_updated_at ON operator_sessions;
CREATE TRIGGER update_operator_sessions_updated_at BEFORE UPDATE ON operator_sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_operator_sign_ins_updated_at ON operator_sign_ins;
CREATE TRIGGER update_operator_sign_ins_updated_at BEFORE UPDATE ON operator_sign_ins FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();