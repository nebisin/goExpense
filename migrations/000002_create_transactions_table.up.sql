CREATE TABLE IF NOT EXISTS transactions
(
    id          bigserial PRIMARY KEY,
    user_id     bigint                      NOT NULL REFERENCES users ON DELETE CASCADE,
    ts_type     text                        NOT NULL,
    title       text                        NOT NULL,
    description text                        NOT NULL,
    amount      float8                      NOT NULL,
    payday      timestamp(0) with time zone NOT NULL,
    created_at  timestamp(0) with time zone NOT NULL DEFAULT now(),
    version     integer                     NOT NULL DEFAULT 1
);