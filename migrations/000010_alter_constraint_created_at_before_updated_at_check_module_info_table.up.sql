ALTER TABLE module_info
    DROP CONSTRAINT IF EXISTS created_at_before_updated_at_check;

ALTER TABLE module_info
    ADD CONSTRAINT created_at_before_updated_at_check
        CHECK (created_at <= updated_at);