-- Add session_code column
ALTER TABLE operator_sessions ADD COLUMN IF NOT EXISTS session_code VARCHAR(50);

-- Generate sequential session codes for existing records
WITH ranked_sessions AS (
    SELECT id, start_time,
           ROW_NUMBER() OVER (PARTITION BY start_time::date ORDER BY start_time ASC) as rn
    FROM operator_sessions
)
UPDATE operator_sessions os
SET session_code = 'SESS-' || to_char(os.start_time, 'DD-MM-YYYY') || '-' || lpad(ranked_sessions.rn::text, 2, '0')
FROM ranked_sessions
WHERE os.id = ranked_sessions.id AND os.session_code IS NULL;

-- Make column NOT NULL
ALTER TABLE operator_sessions ALTER COLUMN session_code SET NOT NULL;

-- Drop and recreate unique constraint
ALTER TABLE operator_sessions DROP CONSTRAINT IF EXISTS operator_sessions_session_code_key;
ALTER TABLE operator_sessions ADD CONSTRAINT operator_sessions_session_code_key UNIQUE (session_code);
