CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    email text UNIQUE NOT NULL,
    hashed_password bytea NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    is_activated bool NOT NULL,
    version integer NOT NULL DEFAULT 1
);