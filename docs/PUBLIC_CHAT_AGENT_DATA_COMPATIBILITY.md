# Public chat agent data compatibility check

Context: before M2.9/M3, public customer-service chat must not lose active agents when the frontend reads `/api/v1/customer-service/agents` from Go instead of PHP.

## Source data in PHP

PHP exposes active public agents from `wp_tz_cs_agents`, not directly from WordPress roles:

```sql
SELECT agent_id, wp_user_id, name, email, avatar, whatsapp, online_status
FROM wp_tz_cs_agents
WHERE status = 'active'
ORDER BY created_at ASC;
```

Important fields:

- `agent_id`: legacy string id used by PHP conversations/messages.
- `wp_user_id`: optional WordPress user link used to hide the current user from their own agent list.
- `status`: account lifecycle; only `active` is public.
- `online_status`: runtime presence (`online`, `busy`, `away`, `offline`).
- `whatsapp`/`avatar`: public agent profile fields.

## Go mapping

Go stores customer-service profile data in `customer_service_agent_profiles` and links each public agent to `users.id`:

- `customer_service_agent_profiles.status = 'active'` maps to PHP `wp_tz_cs_agents.status = 'active'`.
- `customer_service_agent_profiles.online_status`, `whatsapp`, `avatar`, and `agent_id` are migrated from PHP instead of being inferred at runtime.
- `customer_service_agent_profiles.user_id` maps to the Go user id used by ticket assignment and the frontend `wp_user_id` self-filter.
- `users.status = 'active'` is still required for an agent to be publicly visible.
- customer-service roles are canonicalized to `admin`, `manager`, or `support`.
- WordPress legacy roles are normalized:
  - `administrator` → `admin`
  - `shop_manager` → `manager`
  - `agent`, `customer_service`, `customer_support` → `support`

## Findings

1. The M2.8 query included `agent` but omitted the canonical Go `manager` role.
2. The WordPress export/import path previously copied raw role strings (`administrator`, `shop_manager`, comma-separated roles) into `users.role`; those rows would not match `role IN ('admin', 'agent', 'support')`.
3. Go auth/profile responses did not expose `is_agent`, `agent_id`, `roles`, or `display_name`, while the Nuxt chat component uses those fields to enter agent mode and filter the current user.
4. Go `/customer-service/agents` hard-coded `wp_user_id: 0`, so a linked agent user would not be filtered from their own public agent list.
5. Go now has a dedicated agent profile table for PHP's public profile fields (`whatsapp`, `avatar`, `agent_id`, `online_status`).

## Minimal compatibility fix applied

- Normalize WordPress roles to Go roles in auth, profile responses, JWT claims, and the import script.
- Treat `admin`, `manager`, and `support` as customer-service staff; keep legacy `agent` as an alias.
- Return `is_agent`, `agent_id`, `roles`, and `display_name` in `users` API responses.
- Return `wp_user_id = users.id` for Go-projected public agents, so the current-user filter works.
- Import `wp_tz_cs_agents` into `customer_service_agent_profiles`, creating or promoting linked Go support users when needed.

## Go-only verification before M2.9/M3

Migration flow:

```bash
# From the WordPress root, export legacy agent profile rows.
php go-backend/scripts/wordpress-export/export-customer-service-agents.php

# Copy/keep the JSON at go-backend/scripts/export/customer-service-agents.json,
# then import into Go's customer_service_agent_profiles table.
cd go-backend
go run scripts/import-data.go
```

After running the agent profile import, the admin settings panel and public chat read Go data only. Verify the migrated Go tables before cutting product/cart or auto-reply:

```sql
-- Go agent profiles used by public chat.
SELECT p.agent_id, p.user_id AS wp_user_id, p.name, p.email, p.avatar, p.whatsapp,
       p.status, p.online_status, u.role, u.status AS user_status
FROM customer_service_agent_profiles p
JOIN users u ON u.id = p.user_id
WHERE p.status = 'active'
  AND u.status = 'active'
ORDER BY p.created_at ASC, p.id ASC;

-- Go agent profiles that cannot be exposed because user linkage is missing.
SELECT p.agent_id, p.user_id, p.name, p.email, p.status, p.online_status
FROM customer_service_agent_profiles p
LEFT JOIN users u ON u.id = p.user_id
WHERE p.status = 'active'
  AND (p.user_id IS NULL OR u.id IS NULL);
```

Proceed to M2.9/M3 only after active customer-service agents are represented by `customer_service_agent_profiles` rows linked to active Go `users` rows with compatible roles.
