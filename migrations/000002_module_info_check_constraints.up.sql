ALTER TABLE module_info
ADD CONSTRAINT created_at_before_updated_at_check
CHECK ( created_at < updated_at );

ALTER TABLE module_info
ADD CONSTRAINT module_duration_check
CHECK ( module_duration > 5 AND module_duration < 15);