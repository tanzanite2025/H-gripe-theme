-- Review summaries are a materialized public read model for approved reviews.
-- Keep the table creation here as a forward migration for existing databases;
-- the initial schema contains the same contract for clean installs.
CREATE TABLE IF NOT EXISTS review_summaries (
    product_id BIGINT PRIMARY KEY,
    total_reviews INTEGER NOT NULL DEFAULT 0,
    average_rating NUMERIC(3,2) NOT NULL DEFAULT 0,
    rating_1_count INTEGER NOT NULL DEFAULT 0,
    rating_2_count INTEGER NOT NULL DEFAULT 0,
    rating_3_count INTEGER NOT NULL DEFAULT 0,
    rating_4_count INTEGER NOT NULL DEFAULT 0,
    rating_5_count INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT fk_review_summaries_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

INSERT INTO review_summaries (
    product_id,
    total_reviews,
    average_rating,
    rating_1_count,
    rating_2_count,
    rating_3_count,
    rating_4_count,
    rating_5_count
)
SELECT
    product_id,
    COUNT(*)::INTEGER,
    COALESCE(AVG(rating), 0),
    COUNT(*) FILTER (WHERE rating = 1)::INTEGER,
    COUNT(*) FILTER (WHERE rating = 2)::INTEGER,
    COUNT(*) FILTER (WHERE rating = 3)::INTEGER,
    COUNT(*) FILTER (WHERE rating = 4)::INTEGER,
    COUNT(*) FILTER (WHERE rating = 5)::INTEGER
FROM reviews
WHERE status = 'approved'
GROUP BY product_id
ON CONFLICT (product_id) DO UPDATE SET
    total_reviews = EXCLUDED.total_reviews,
    average_rating = EXCLUDED.average_rating,
    rating_1_count = EXCLUDED.rating_1_count,
    rating_2_count = EXCLUDED.rating_2_count,
    rating_3_count = EXCLUDED.rating_3_count,
    rating_4_count = EXCLUDED.rating_4_count,
    rating_5_count = EXCLUDED.rating_5_count;
