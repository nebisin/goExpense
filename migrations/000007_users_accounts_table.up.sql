CREATE TABLE IF NOT EXISTS users_accounts (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    account_id bigint NOT NULL REFERENCES accounts ON DELETE CASCADE INITIALLY DEFERRED,
    PRIMARY KEY (user_id, account_id)
);