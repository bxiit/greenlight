CREATE TABLE IF NOT EXISTS user_info_tokens
(
    hash    bytea PRIMARY KEY,
    user_info_id bigint                      NOT NULL REFERENCES user_info ON DELETE CASCADE,
    expiry  timestamp(0) with time zone NOT NULL,
    scope   text                        NOT NULL
);
