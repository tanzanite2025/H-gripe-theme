DELETE FROM settings
WHERE key = 'refund_cancellation_policy'
  AND "group" = 'refund_cancellation';
