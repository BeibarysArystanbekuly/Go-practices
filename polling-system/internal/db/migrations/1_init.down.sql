DROP INDEX IF EXISTS idx_votes_option_id;
DROP INDEX IF EXISTS idx_votes_user_id;
DROP INDEX IF EXISTS idx_votes_poll_id;

DROP TABLE IF EXISTS aggregated_results;
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS options;
DROP TABLE IF EXISTS polls;
DROP TABLE IF EXISTS users;
