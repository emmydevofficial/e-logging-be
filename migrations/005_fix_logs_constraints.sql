-- Migration: 005_fix_logs_constraints.sql
-- Relax constraints on logs table to support shift summary logs and operator sign-in logs

-- Make station_id nullable
ALTER TABLE logs ALTER COLUMN station_id DROP NOT NULL;

-- Make device_id nullable
ALTER TABLE logs ALTER COLUMN device_id DROP NOT NULL;

-- Drop the action check constraint as we need new actions (like create_summary, Operator Sign-In, Operator Sign-Out)
ALTER TABLE logs DROP CONSTRAINT IF EXISTS logs_action_check;
