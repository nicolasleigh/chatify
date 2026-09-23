-- Delivery history is intentionally retained during normal operation. This
-- rollback removes the queue state only when the migration is explicitly
-- reversed by an operator.
DROP TABLE IF EXISTS notification_deliveries;
