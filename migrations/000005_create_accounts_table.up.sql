CREATE TABLE IF NOT EXISTS accounts
(
    id         bigserial PRIMARY KEY,
    user_id    bigint                      NOT NULL REFERENCES users ON DELETE CASCADE,
    name       text                        NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    version    integer                     NOT NULL DEFAULT 1
);