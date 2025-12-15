ALTER TABLE aggregated_results
    DROP CONSTRAINT IF EXISTS aggregated_results_option_poll_fkey,
    ADD CONSTRAINT aggregated_results_option_id_fkey
        FOREIGN KEY (option_id) REFERENCES options (id) ON DELETE CASCADE;

ALTER TABLE votes
    DROP CONSTRAINT IF EXISTS votes_option_poll_fkey,
    ADD CONSTRAINT votes_option_id_fkey
        FOREIGN KEY (option_id) REFERENCES options (id) ON DELETE CASCADE;

ALTER TABLE options
    DROP CONSTRAINT IF EXISTS options_id_poll_id_unique;
