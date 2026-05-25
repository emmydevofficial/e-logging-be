-- Migration: 008_add_priority_level.sql
-- Adds priority_level column (levels 1-6) to logs table

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'logs' AND column_name = 'priority_level') THEN
        ALTER TABLE logs ADD COLUMN priority_level INTEGER CHECK (priority_level >= 1 AND priority_level <= 6);
    END IF;
END $$;
