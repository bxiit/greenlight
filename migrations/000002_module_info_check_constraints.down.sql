ALTER TABLE module_info
    DROP CONSTRAINT IF EXISTS created_at_before_updated_at_check;

ALTER TABLE module_info
    DROP CONSTRAINT IF EXISTS module_duration_check;