-- name: ClaimOutboxEvents :many
-- Claim a bounded batch with row locks. SKIP LOCKED lets multiple API
-- instances run workers concurrently without waiting on one another, while
-- the lease condition recovers rows from a worker that crashed mid-publish.
WITH candidates AS (
    SELECT id
    FROM outbox_events
    WHERE (
        status = 'pending'
        AND available_at <= NOW()
    ) OR (
        status = 'processing'
        AND locked_at < NOW() - INTERVAL '5 minutes'
    )
    ORDER BY id
    LIMIT $1
    FOR UPDATE SKIP LOCKED
)
UPDATE outbox_events AS event
SET
    status = 'processing',
    attempts = event.attempts + 1,
    locked_at = NOW()
FROM candidates
WHERE event.id = candidates.id
RETURNING
    event.id,
    event.event_type,
    event.event_version,
    event.aggregate_type,
    event.aggregate_id,
    event.payload,
    event.attempts,
    event.created_at;

-- name: MarkOutboxEventPublished :exec
-- A successful publisher confirmation is the only point at which an event
-- becomes published. The row remains for auditability and deduplication.
UPDATE outbox_events
SET
    status = 'published',
    published_at = NOW(),
    locked_at = NULL,
    last_error = NULL
WHERE id = $1;

-- name: MarkOutboxEventFailed :exec
-- The worker chooses the next status and retry time. Keeping this update
-- generic allows transient failures to return to pending while exhausted
-- events can be retained as failed records for operational inspection.
UPDATE outbox_events
SET
    status = $2,
    available_at = NOW() + ($3::integer * INTERVAL '1 second'),
    locked_at = NULL,
    last_error = $4
WHERE id = $1;
