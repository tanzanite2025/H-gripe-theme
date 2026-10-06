-- Keep trade-remedy sensitivity on the independent customs profile. This is
-- operational guidance attached to a classification profile, not a product
-- specification-template concern.
ALTER TABLE customs_classification_profiles
    ADD COLUMN IF NOT EXISTS trade_remedy_risk_level VARCHAR(16) NOT NULL DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS trade_remedy_risk_tags_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS trade_remedy_declaration_advice TEXT NOT NULL DEFAULT '';

-- The wheelset profile is the known high-sensitivity case. Keep the advice
-- factual and review-oriented: destination measures, origin and the actual
-- presentation of the goods must be checked at the time of shipment. Do not
-- hard-code a duty rate that can change with the current tariff or an
-- exclusion decision.
UPDATE customs_classification_profiles
SET trade_remedy_risk_level = 'high',
    trade_remedy_risk_tags_json = '[
        "eu_anti_dumping_attention",
        "eu_complete_wheelset_anti_circumvention",
        "us_section_301_list_3_review"
    ]'::jsonb,
    trade_remedy_declaration_advice = '欧盟出货前必须按当前 TARIC、目的成员国、原产国和实际货物状态复核反倾销、反规避措施及 EC 88/97 相关适用条件。只有货物确实是独立散件且实物、发票、装箱单和运单一致时，才可分别申报；不得为规避贸易救济措施虚假拆单、改名或拆分申报。无法确认时暂停出单并交关务复核。美国线路另行核对当前 HTS 9903.88.03、Section 301 List 3、排除/豁免与原产地规则。'
WHERE slug = 'bicycle-wheelset'
  AND source = 'built_in'
  AND BTRIM(COALESCE(trade_remedy_declaration_advice, '')) = ''
  AND COALESCE(trade_remedy_risk_tags_json, '[]'::jsonb) = '[]'::jsonb;
