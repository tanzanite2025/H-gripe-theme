-- NULL distinguishes historical uncaptured annotations from collected empty data.
ALTER TABLE tax_rate_source_snapshots
    ADD COLUMN country_rate_comments JSONB;
ALTER TABLE tax_rate_source_snapshots
    ADD CONSTRAINT tax_rate_source_snapshot_country_rate_comments_object
    CHECK (country_rate_comments IS NULL OR jsonb_typeof(country_rate_comments) = 'object');
