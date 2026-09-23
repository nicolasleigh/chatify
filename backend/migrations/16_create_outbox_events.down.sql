-- Outbox events are retained for audit and retry purposes. Dropping this table
-- is therefore intentionally an explicit rollback operation, not something
-- the application performs during normal startup.
DROP TABLE IF EXISTS outbox_events;
