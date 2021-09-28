CREATE TABLE IF NOT EXISTS accounts
(
    id          bigserial PRIMARY KEY,
    owner_id    bigint                      NOT NULL REFERENCES users ON DELETE CASCADE,
    title       text                        NOT NULL,
    description text,
    created_at  timestamp(0) with time zone NOT NULL DEFAULT now(),
    version     integer                     NOT NULL DEFAULT 1
);