CREATE TABLE IF NOT EXISTS tokens (
    hash bytea PRIMARY KEY,
    USER_ID BIGINT NOT NULL REFERENCES USERS ON DELETE CASCADE,
    expiry timestamp(0) with time zone NOT NULL,
    scope text NOT NULL
)
