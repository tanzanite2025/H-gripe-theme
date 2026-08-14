UPDATE settings
SET value = '/company/about/factory',
    updated_at = NOW()
WHERE key = 'factory_link'
  AND locale = 'global'
  AND "group" = 'website_profile';
