-- Migration: 006_add_supervisor_role.sql
-- Alter users table check constraint to support the new "supervisor" role

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'supervisor', 'operator', 'downloader', 'viewer'));
