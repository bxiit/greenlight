CREATE TABLE IF NOT EXISTS user_info_permissions
(
    user_info_id       bigint NOT NULL REFERENCES user_info ON DELETE CASCADE,
    permission_id bigint NOT NULL REFERENCES permissions ON DELETE CASCADE,
    PRIMARY KEY (user_info_id, permission_id)
);