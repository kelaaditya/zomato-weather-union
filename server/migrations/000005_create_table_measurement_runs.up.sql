CREATE TABLE IF NOT EXISTS measurement_runs(
    run_id UUID PRIMARY KEY NOT NULL,
    time_stamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);