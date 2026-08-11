-- Remove legacy brand prefixes from public asset URLs and public spoke presets.
-- Operational identifiers and historical migration files remain unchanged.

UPDATE settings
SET value = replace(
                replace(
                    replace(
                        replace(
                            replace(value, '/tanzanite-', '/'),
                            '/TANZANITE-', '/'
                        ),
                        '/tanzantie-', '/'
                    ),
                    '/tanznaite-', '/'
                ),
                '/tananite-', '/'
            ),
    updated_at = NOW()
WHERE "group" = 'website_profile'
  AND key = 'factory_image_url'
  AND (
      value LIKE '%/tanzanite-%'
      OR value LIKE '%/TANZANITE-%'
      OR value LIKE '%/tanzantie-%'
      OR value LIKE '%/tanznaite-%'
      OR value LIKE '%/tananite-%'
  );

UPDATE product_media
SET url = replace(
              replace(
                  replace(
                      replace(
                          replace(url, '/tanzanite-', '/'),
                          '/TANZANITE-', '/'
                      ),
                      '/tanzantie-', '/'
                  ),
                  '/tanznaite-', '/'
              ),
              '/tananite-', '/'
          ),
    thumbnail_url = replace(
                        replace(
                            replace(
                                replace(
                                    replace(thumbnail_url, '/tanzanite-', '/'),
                                    '/TANZANITE-', '/'
                                ),
                                '/tanzantie-', '/'
                            ),
                            '/tanznaite-', '/'
                        ),
                        '/tananite-', '/'
                    ),
    poster_url = replace(
                     replace(
                         replace(
                             replace(
                                 replace(poster_url, '/tanzanite-', '/'),
                                 '/TANZANITE-', '/'
                             ),
                             '/tanzantie-', '/'
                         ),
                         '/tanznaite-', '/'
                     ),
                     '/tananite-', '/'
                 ),
    updated_at = NOW()
WHERE url LIKE '%/tanzanite-%'
   OR url LIKE '%/TANZANITE-%'
   OR url LIKE '%/tanzantie-%'
   OR url LIKE '%/tanznaite-%'
   OR url LIKE '%/tananite-%'
   OR thumbnail_url LIKE '%/tanzanite-%'
   OR thumbnail_url LIKE '%/TANZANITE-%'
   OR thumbnail_url LIKE '%/tanzantie-%'
   OR thumbnail_url LIKE '%/tanznaite-%'
   OR thumbnail_url LIKE '%/tananite-%'
   OR poster_url LIKE '%/tanzanite-%'
   OR poster_url LIKE '%/TANZANITE-%'
   OR poster_url LIKE '%/tanzantie-%'
   OR poster_url LIKE '%/tanznaite-%'
   OR poster_url LIKE '%/tananite-%';

UPDATE media_assets
SET url = replace(
              replace(
                  replace(
                      replace(
                          replace(url, '/tanzanite-', '/'),
                          '/TANZANITE-', '/'
                      ),
                      '/tanzantie-', '/'
                  ),
                  '/tanznaite-', '/'
              ),
              '/tananite-', '/'
          ),
    filename = regexp_replace(filename, '^(tanzanite|tanzantie|tanznaite|tananite)-', '', 1, 0, 'i'),
    original_filename = regexp_replace(original_filename, '^(tanzanite|tanzantie|tanznaite|tananite)-', '', 1, 0, 'i'),
    updated_at = NOW()
WHERE url LIKE '%/tanzanite-%'
   OR url LIKE '%/TANZANITE-%'
   OR url LIKE '%/tanzantie-%'
   OR url LIKE '%/tanznaite-%'
   OR url LIKE '%/tananite-%'
   OR filename ~* '^(tanzanite|tanzantie|tanznaite|tananite)-'
   OR original_filename ~* '^(tanzanite|tanzantie|tanznaite|tananite)-';

UPDATE spoke_build_presets
SET name = regexp_replace(name, '^(Tanzanite|H-GRIPE)[[:space:]]+', '', 1, 0, 'i'),
    updated_at = NOW()
WHERE name ~* '^(Tanzanite|H-GRIPE)[[:space:]]+';
