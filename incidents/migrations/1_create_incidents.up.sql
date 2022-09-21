CREATE TABLE incidents
(
    id               BIGSERIAL PRIMARY KEY,
    assigned_user_id INTEGER,
    body             TEXT      NOT NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    acknowledged_at  TIMESTAMP
);
