-- name: InsertNotificationDelivery :exec
-- RabbitMQ provides at-least-once delivery. A conflict means this exact
-- event/recipient/subscription combination has already been scheduled and is
-- therefore a successful idempotent operation.
INSERT INTO notification_deliveries (
    event_id,
    message_id,
    recipient_id,
    subscription_id,
    channel
)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (event_id, recipient_id, subscription_id, channel) DO NOTHING;

-- name: GetConversationNotificationRecipients :many
-- Recipients are derived from the authoritative membership table instead of
-- trusting a client-provided list. The sender is excluded because users do
-- not need an offline notification for their own message.
SELECT member_id
FROM conversation_members
WHERE conversation_id = $1
  AND member_id <> $2
ORDER BY member_id;

-- name: ClaimNotificationDeliveries :many
-- Lease a bounded batch so multiple delivery workers can operate concurrently
-- without sending the same row at the same time. A stale processing lease is
-- eligible again after five minutes when a worker crashes.
WITH candidates AS (
    SELECT id
    FROM notification_deliveries
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
UPDATE notification_deliveries AS delivery
SET
    status = 'processing',
    attempts = delivery.attempts + 1,
    locked_at = NOW(),
    updated_at = NOW()
FROM candidates
WHERE delivery.id = candidates.id
RETURNING
    delivery.id,
    delivery.event_id,
    delivery.message_id,
    delivery.recipient_id,
    delivery.subscription_id,
    delivery.channel,
    delivery.attempts,
    delivery.created_at;

-- name: GetNotificationDelivery :one
-- The subscription is read after claiming the delivery. Joining here keeps
-- endpoint credentials out of the queue event and lets disabled endpoints be
-- skipped without contacting a push provider.
SELECT
    delivery.id,
    delivery.event_id,
    delivery.message_id,
    delivery.recipient_id,
    delivery.subscription_id,
    delivery.channel,
    delivery.attempts,
    delivery.created_at,
    subscription.endpoint,
    subscription.p256dh,
    subscription.auth,
    subscription.enabled
FROM notification_deliveries AS delivery
LEFT JOIN push_subscriptions AS subscription
  ON subscription.id = delivery.subscription_id
WHERE delivery.id = $1;

-- name: MarkNotificationDeliverySent :exec
UPDATE notification_deliveries
SET
    status = 'sent',
    sent_at = NOW(),
    locked_at = NULL,
    last_error = NULL,
    updated_at = NOW()
WHERE id = $1;

-- name: MarkNotificationDeliveryFailed :exec
-- The worker supplies the next status and retry delay. Permanent failures use
-- skipped or failed with a zero delay and remain inspectable.
UPDATE notification_deliveries
SET
    status = $2,
    available_at = NOW() + ($3::integer * INTERVAL '1 second'),
    locked_at = NULL,
    last_error = $4,
    updated_at = NOW()
WHERE id = $1;
