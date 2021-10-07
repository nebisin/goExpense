CREATE TABLE IF NOT EXISTS statistics (
    account_id bigint NOT NULL REFERENCES accounts ON DELETE CASCADE,
    date date NOT NULL,
    earning real NOT NULL,
    spending real NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1,
    PRIMARY KEY (account_id, date)
);