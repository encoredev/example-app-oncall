CREATE TABLE schedules
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    INTEGER   NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time   TIMESTAMP NOT NULL
);

CREATE INDEX schedules_range_index ON schedules (start_time, end_time);