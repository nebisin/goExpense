ALTER TABLE transactions
    ADD account_id bigint,
    ADD FOREIGN KEY (account_id) REFERENCES accounts ON DELETE CASCADE;