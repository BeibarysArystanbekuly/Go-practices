ALTER TABLE options
    ADD CONSTRAINT options_id_poll_id_unique UNIQUE (id, poll_id);

ALTER TABLE votes
    DROP CONSTRAINT IF EXISTS votes_option_id_fkey,
    ADD CONSTRAINT votes_option_poll_fkey
        FOREIGN KEY (option_id, poll_id) REFERENCES options (id, poll_id) ON DELETE CASCADE;

ALTER TABLE aggregated_results
    DROP CONSTRAINT IF EXISTS aggregated_results_option_id_fkey,
    ADD CONSTRAINT aggregated_results_option_poll_fkey
        FOREIGN KEY (option_id, poll_id) REFERENCES options (id, poll_id) ON DELETE CASCADE;
