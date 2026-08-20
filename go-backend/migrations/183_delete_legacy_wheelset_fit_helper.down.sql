UPDATE selection_assistant_flows
SET
    is_enabled = FALSE,
    updated_at = NOW()
WHERE slug = 'wheelset-fit-helper';
