-- Drop the check constraint if it exists
ALTER TABLE logs DROP CONSTRAINT IF EXISTS logs_action_check;

-- Update existing records in logs table to use the new action names
UPDATE logs SET action = 'reported' WHERE action = 'reported_that';
UPDATE logs SET action = 'was_instructed' WHERE action = 'was_instructed_to';
UPDATE logs SET action = 'was_granted_approval' WHERE action = 'was_given_approval';

-- Add the updated constraint
ALTER TABLE logs ADD CONSTRAINT logs_action_check CHECK (action IN ('reported', 'was_instructed', 'was_granted_approval', 'was_not_granted_approval', 'miscellaneous', 'create_summary', 'Operator Sign-In', 'Operator Sign-Out'));
