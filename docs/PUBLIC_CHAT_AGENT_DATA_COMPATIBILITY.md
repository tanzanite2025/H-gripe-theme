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

Go currently projects customer-service agents from `users`:

- `users.status = 'active'` maps to PHP `wp_tz_cs_agents.status = 'active'`.
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
5. Go still does not have PHP's dedicated agent profile fields (`whatsapp`, `avatar`) or persistent `online_status`. Those need a later schema/data migration if exact PHP parity is required.

## Minimal compatibility fix applied

- Normalize WordPress roles to Go roles in auth, profile responses, JWT claims, and the import script.
- Treat `admin`, `manager`, and `support` as customer-service staff; keep legacy `agent` as an alias.
- Return `is_agent`, `agent_id`, `roles`, and `display_name` in `users` API responses.
- Return `wp_user_id = users.id` for Go-projected public agents, so the current-user filter works.

## Remaining preflight before M2.9/M3

Run against production/read-only data before cutting product/cart or auto-reply:

```sql
-- PHP source of truth: active agents.
SELECT agent_id, wp_user_id, name, email, status, online_status
FROM wp_tz_cs_agents
WHERE status = 'active'
ORDER BY created_at ASC;

-- Go projection candidates.
SELECT id, username, email, role, status
FROM users
WHERE status = 'active'
  AND (
    role IN ('admin', 'manager', 'support', 'agent', 'administrator', 'shop_manager', 'customer_service', 'customer_support')
    OR LOWER(role) LIKE '%administrator%'
    OR LOWER(role) LIKE '%shop_manager%'
    OR LOWER(role) LIKE '%customer_service%'
    OR LOWER(role) LIKE '%customer_support%'
    OR LOWER(role) LIKE '%support%'
    OR LOWER(role) LIKE '%agent%'
  )
ORDER BY role ASC, created_at ASC;

-- Linked PHP agents that do not have a matching Go user.
SELECT a.agent_id, a.wp_user_id, a.name, a.email, a.status, a.online_status
FROM wp_tz_cs_agents a
LEFT JOIN users u ON u.id = a.wp_user_id
WHERE a.status = 'active'
  AND (a.wp_user_id IS NULL OR u.id IS NULL);
```

Proceed to M2.9/M3 only after every active `wp_tz_cs_agents` row is either represented by a Go `users` row with a compatible role, or intentionally documented as not migrating.
