DELETE FROM email_templates
WHERE locale = 'en'
  AND version = 1
  AND code IN (
      'order_confirmation',
      'order_payment_expired',
      'order_cancelled',
      'order_shipping_notification',
      'order_delivered',
      'order_completed',
      'order_refunded',
      'after_sales_requested',
      'after_sales_approved',
      'after_sales_awaiting_return',
      'after_sales_return_in_transit',
      'after_sales_received',
      'after_sales_resolving',
      'after_sales_completed',
      'after_sales_rejected'
  );
