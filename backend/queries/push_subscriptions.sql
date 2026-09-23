-- name: UpsertPushSubscription :one
-- Registering the same browser subscription repeatedly is safe. If a browser
-- logs out and another user later signs in on it, ownership is transferred
-- only through this authenticated endpoint and the old disabled state is
-- cleared for the new delivery owner.
INSERT INTO push_subscriptions (
    user_id,
    endpoint,
    p256dh,
    auth,
    user_agent,
    device_label,
    enabled,
    disabled_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, true, NULL, NOW())
ON CONFLICT (endpoint) DO UPDATE SET
    user_id = EXCLUDED.user_id,
    p256dh = EXCLUDED.p256dh,
    auth = EXCLUDED.auth,
    user_agent = EXCLUDED.user_agent,
    device_label = EXCLUDED.device_label,
    enabled = true,
    disabled_at = NULL,
    updated_at = NOW()
RETURNING
    id,
    user_id,
    endpoint,
    p256dh,
    auth,
    user_agent,
    device_label,
    enabled,
    last_used_at,
    disabled_at,
    created_at,
    updated_at;

-- name: GetEnabledPushSubscriptions :many
-- The notification consumer only needs currently enabled endpoints. Disabled
-- rows remain available for auditing and are excluded at the database layer.
SELECT
    id,
    user_id,
    endpoint,
    p256dh,
    auth,
    user_agent,
    device_label,
    enabled,
    last_used_at,
    disabled_at,
    created_at,
    updated_at
FROM push_subscriptions
WHERE user_id = $1
  AND enabled = true
ORDER BY id;

-- name: DisablePushSubscription :exec
-- Ownership is part of the predicate so one user cannot disable another
-- user's browser endpoint even if an ID is guessed.
UPDATE push_subscriptions
SET
    enabled = false,
    disabled_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND user_id = $2;

-- name: DisablePushSubscriptionByEndpoint :exec
-- Push providers commonly identify an invalid subscription by endpoint. This
-- operation is intentionally scoped to the endpoint and can be called by the
-- trusted delivery worker after a permanent provider response.
UPDATE push_subscriptions
SET
    enabled = false,
    disabled_at = NOW(),
    updated_at = NOW()
WHERE endpoint = $1;

-- name: MarkPushSubscriptionUsed :exec
UPDATE push_subscriptions
SET
    last_used_at = NOW(),
    updated_at = NOW()
WHERE id = $1;
