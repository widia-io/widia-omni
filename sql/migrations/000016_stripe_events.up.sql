SET search_path TO widia_omni;

CREATE TABLE stripe_events_processed (
    event_id        TEXT PRIMARY KEY,
    event_type      TEXT NOT NULL,
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
