-- Removing subscriptions revokes browser notification delivery. This is an
-- explicit migration rollback and is never performed by normal application
-- cleanup code.
DROP TABLE IF EXISTS push_subscriptions;
