-- A browser PushSubscription belongs to one Chatify user and may exist on
-- several browsers or devices at the same time. The endpoint is globally
-- unique so repeated registration from the same browser is idempotent.
CREATE TABLE IF NOT EXISTS push_subscriptions (
  id bigserial NOT NULL PRIMARY KEY,
  user_id bigint NOT NULL,
  endpoint text NOT NULL,
  p256dh text NOT NULL,
  auth text NOT NULL,
  user_agent text,
  device_label varchar(200),
  enabled boolean NOT NULL DEFAULT true,
  last_used_at timestamp(0) with time zone,
  disabled_at timestamp(0) with time zone,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  CONSTRAINT push_subscriptions_user_fk
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT push_subscriptions_endpoint_check
    CHECK (char_length(trim(endpoint)) > 0),
  CONSTRAINT push_subscriptions_p256dh_check
    CHECK (char_length(trim(p256dh)) > 0),
  CONSTRAINT push_subscriptions_auth_check
    CHECK (char_length(trim(auth)) > 0)
);

-- Notification consumers query active subscriptions by recipient. Keeping
-- disabled rows allows invalid browser endpoints to be audited without
-- repeatedly attempting delivery to them.
CREATE INDEX IF NOT EXISTS idx_push_subscriptions_user_enabled
ON push_subscriptions (user_id, enabled, id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_push_subscriptions_endpoint
ON push_subscriptions (endpoint);
