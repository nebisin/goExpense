ALTER TABLE transactions
    DROP CONSTRAINT transactions_account_id_fkey,
    DROP COLUMN account_id;