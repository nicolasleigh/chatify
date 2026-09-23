-- A delivery row is the durable intent to notify one recipient through one
-- browser subscription. RabbitMQ may redeliver the source event, so the
-- unique key below is the idempotency boundary for notification creation.
CREATE TABLE IF NOT EXISTS notification_deliveries (
  id bigserial NOT NULL PRIMARY KEY,
  event_id bigint NOT NULL,
  message_id bigint NOT NULL,
  recipient_id bigint NOT NULL,
  subscription_id bigint NOT NULL,
  channel varchar(50) NOT NULL DEFAULT 'web_push',
  status varchar(20) NOT NULL DEFAULT 'pending',
  attempts integer NOT NULL DEFAULT 0,
  available_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  locked_at timestamp(0) with time zone,
  sent_at timestamp(0) with time zone,
  last_error text,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  CONSTRAINT notification_deliveries_event_fk
    FOREIGN KEY (event_id) REFERENCES outbox_events (id) ON DELETE CASCADE,
  CONSTRAINT notification_deliveries_message_fk
    FOREIGN KEY (message_id) REFERENCES messages (id) ON DELETE CASCADE,
  CONSTRAINT notification_deliveries_recipient_fk
    FOREIGN KEY (recipient_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT notification_deliveries_subscription_fk
    FOREIGN KEY (subscription_id) REFERENCES push_subscriptions (id) ON DELETE CASCADE,
  CONSTRAINT notification_deliveries_channel_check
    CHECK (channel IN ('web_push')),
  CONSTRAINT notification_deliveries_status_check
    CHECK (status IN ('pending', 'processing', 'sent', 'failed', 'skipped')),
  CONSTRAINT notification_deliveries_attempts_check
    CHECK (attempts >= 0)
);

-- The unique key makes RabbitMQ redelivery harmless and permits multiple
-- browser subscriptions for the same recipient to receive the same message.
CREATE UNIQUE INDEX IF NOT EXISTS idx_notification_deliveries_idempotency
ON notification_deliveries (event_id, recipient_id, subscription_id, channel);

-- Delivery workers poll by status and availability, while operations often
-- inspect a user's notification history separately.
CREATE INDEX IF NOT EXISTS idx_notification_deliveries_delivery
ON notification_deliveries (status, available_at, id);

CREATE INDEX IF NOT EXISTS idx_notification_deliveries_recipient
ON notification_deliveries (recipient_id, created_at DESC, id DESC);
