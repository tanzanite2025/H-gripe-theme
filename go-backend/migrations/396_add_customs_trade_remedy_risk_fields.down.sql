ALTER TABLE customs_classification_profiles
    DROP COLUMN IF EXISTS trade_remedy_risk_level,
    DROP COLUMN IF EXISTS trade_remedy_risk_tags_json,
    DROP COLUMN IF EXISTS trade_remedy_declaration_advice;
