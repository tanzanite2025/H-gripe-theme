DELETE FROM settings
WHERE "group" = 'currency'
  AND key = 'currency_primary_currency'
  AND locale = 'en';
