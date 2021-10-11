CREATE TABLE IF NOT EXISTS accounts (
    id bigserial PRIMARY KEY,
    owner_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    title text NOT NULL,
    description text,
    total_income real NOT NULL DEFAULT 0,
    total_expense real NOT NULL DEFAULT 0,
    currency text NOT NULL DEFAULT 'USD',
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1
);