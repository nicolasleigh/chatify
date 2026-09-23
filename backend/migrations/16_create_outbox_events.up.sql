-- The outbox stores the event in the same database transaction as the
-- business change that produced it. This prevents a successful message write
-- from being lost merely because RabbitMQ is temporarily unavailable.
CREATE TABLE IF NOT EXISTS outbox_events (
  id bigserial NOT NULL PRIMARY KEY,
  event_type varchar(200) NOT NULL,
  event_version integer NOT NULL DEFAULT 1,
  aggregate_type varchar(200) NOT NULL,
  aggregate_id bigint NOT NULL,
  payload jsonb NOT NULL,
  status varchar(20) NOT NULL DEFAULT 'pending',
  attempts integer NOT NULL DEFAULT 0,
  available_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  locked_at timestamp(0) with time zone,
  published_at timestamp(0) with time zone,
  last_error text,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  CONSTRAINT outbox_events_status_check
    CHECK (status IN ('pending', 'processing', 'published', 'failed')),
  CONSTRAINT outbox_events_attempts_check
    CHECK (attempts >= 0)
);

-- The worker selects events that are ready for delivery in id order. The
-- status/available_at prefix keeps polling efficient as the table grows.
CREATE INDEX IF NOT EXISTS idx_outbox_events_delivery
ON outbox_events (status, available_at, id);

-- This index supports operational queries that inspect events belonging to a
-- particular aggregate without making the delivery index wider.
CREATE INDEX IF NOT EXISTS idx_outbox_events_aggregate
ON outbox_events (aggregate_type, aggregate_id, id);
