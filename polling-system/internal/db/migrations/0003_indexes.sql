CREATE INDEX IF NOT EXISTS idx_polls_status ON polls(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_options_poll_text ON options(poll_id, text);
