CREATE INDEX IF NOT EXISTS transactions_title_idx ON transactions USING gin (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS transactions_tags_idx ON posts USING gin (tags);