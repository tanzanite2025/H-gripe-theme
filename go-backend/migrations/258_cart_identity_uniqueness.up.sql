-- Cart identity is unique only for active carts. Deleted carts remain
-- historical records and must not block a new cart for the same identity.

UPDATE carts
   SET session_id = btrim(session_id)
 WHERE session_id IS NOT NULL
   AND session_id <> btrim(session_id);

DO $$
DECLARE
    duplicate_cart RECORD;
    cart_item RECORD;
    canonical_item_id BIGINT;
BEGIN
    -- Keep the oldest active cart for each authenticated user and merge the
    -- duplicate cart items before removing the redundant cart row.
    FOR duplicate_cart IN
        SELECT duplicate.id AS duplicate_id,
               canonical.id AS canonical_id
          FROM carts AS duplicate
          JOIN LATERAL (
              SELECT candidate.id
                FROM carts AS candidate
               WHERE candidate.deleted_at IS NULL
                 AND candidate.user_id = duplicate.user_id
               ORDER BY candidate.id
               LIMIT 1
          ) AS canonical ON TRUE
         WHERE duplicate.deleted_at IS NULL
           AND duplicate.user_id IS NOT NULL
           AND duplicate.id <> canonical.id
         ORDER BY duplicate.id
    LOOP
        FOR cart_item IN
            SELECT id, product_id, variant_id, quantity, price, currency
              FROM cart_items
             WHERE cart_id = duplicate_cart.duplicate_id
             ORDER BY id
        LOOP
            canonical_item_id := NULL;
            SELECT existing.id
              INTO canonical_item_id
              FROM cart_items AS existing
             WHERE existing.cart_id = duplicate_cart.canonical_id
               AND existing.product_id = cart_item.product_id
               AND existing.variant_id = cart_item.variant_id
             ORDER BY existing.id
             LIMIT 1;

            IF canonical_item_id IS NULL THEN
                UPDATE cart_items
                   SET cart_id = duplicate_cart.canonical_id
                 WHERE id = cart_item.id;
            ELSE
                UPDATE cart_items
                   SET quantity = quantity + cart_item.quantity,
                       price = cart_item.price,
                       currency = cart_item.currency,
                       updated_at = CURRENT_TIMESTAMP
                 WHERE id = canonical_item_id;
                DELETE FROM cart_items WHERE id = cart_item.id;
            END IF;
        END LOOP;

        UPDATE carts
           SET deleted_at = CURRENT_TIMESTAMP,
               updated_at = CURRENT_TIMESTAMP
         WHERE id = duplicate_cart.duplicate_id;
    END LOOP;

    -- Keep the oldest active anonymous cart for each non-empty session.
    FOR duplicate_cart IN
        SELECT duplicate.id AS duplicate_id,
               canonical.id AS canonical_id
          FROM carts AS duplicate
          JOIN LATERAL (
              SELECT candidate.id
                FROM carts AS candidate
               WHERE candidate.deleted_at IS NULL
                 AND candidate.user_id IS NULL
                 AND candidate.session_id = duplicate.session_id
               ORDER BY candidate.id
               LIMIT 1
          ) AS canonical ON TRUE
         WHERE duplicate.deleted_at IS NULL
           AND duplicate.user_id IS NULL
           AND duplicate.session_id <> ''
           AND duplicate.id <> canonical.id
         ORDER BY duplicate.id
    LOOP
        FOR cart_item IN
            SELECT id, product_id, variant_id, quantity, price, currency
              FROM cart_items
             WHERE cart_id = duplicate_cart.duplicate_id
             ORDER BY id
        LOOP
            canonical_item_id := NULL;
            SELECT existing.id
              INTO canonical_item_id
              FROM cart_items AS existing
             WHERE existing.cart_id = duplicate_cart.canonical_id
               AND existing.product_id = cart_item.product_id
               AND existing.variant_id = cart_item.variant_id
             ORDER BY existing.id
             LIMIT 1;

            IF canonical_item_id IS NULL THEN
                UPDATE cart_items
                   SET cart_id = duplicate_cart.canonical_id
                 WHERE id = cart_item.id;
            ELSE
                UPDATE cart_items
                   SET quantity = quantity + cart_item.quantity,
                       price = cart_item.price,
                       currency = cart_item.currency,
                       updated_at = CURRENT_TIMESTAMP
                 WHERE id = canonical_item_id;
                DELETE FROM cart_items WHERE id = cart_item.id;
            END IF;
        END LOOP;

        UPDATE carts
           SET deleted_at = CURRENT_TIMESTAMP,
               updated_at = CURRENT_TIMESTAMP
         WHERE id = duplicate_cart.duplicate_id;
    END LOOP;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_carts_active_user_id
    ON carts(user_id)
    WHERE user_id IS NOT NULL
      AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_carts_active_session_id
    ON carts(session_id)
    WHERE user_id IS NULL
      AND session_id <> ''
      AND deleted_at IS NULL;
