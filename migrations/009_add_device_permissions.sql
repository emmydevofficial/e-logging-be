-- Migration: 009_add_device_permissions.sql
-- Adds can_log and can_download columns to devices table

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'can_log') THEN
        ALTER TABLE devices ADD COLUMN can_log BOOLEAN NOT NULL DEFAULT FALSE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'devices' AND column_name = 'can_download') THEN
        ALTER TABLE devices ADD COLUMN can_download BOOLEAN NOT NULL DEFAULT FALSE;
    END IF;
END $$;
