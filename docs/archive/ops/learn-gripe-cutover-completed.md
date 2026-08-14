# 【已完成】learn.gripe 域名切换记录

> 本次域名切换已完成。本文保留为历史执行记录，不再作为当前操作手册。

This runbook prepares `learn.gripe` to replace `legacy.example` without
changing the Hostinger VPS, Docker project, database, Redis, or uploads
boundary.

## Current Snapshot

Snapshot date: 2026-08-11.

- `learn.gripe` is currently delegated to Hostinger nameservers:
  `nova.dns-parking.com` and `cosmos.dns-parking.com`.
- Public DNS currently returns the Hostinger parking address `2.57.91.91`.
- `www.learn.gripe` currently aliases the root domain.
- No `admin.learn.gripe` record exists yet.
- The repository runbook identifies the expected production VPS as VM
  `1834903` with IPv4 `2.25.85.201`. Confirm this through Hostinger before
  changing DNS.
- The current `legacy.example` zone is already delegated to Cloudflare.

The DNS snapshot is informational. Before changing nameservers, inspect the
Hostinger DNS zone for mail, verification, and third-party service records and
recreate any required records in Cloudflare.

## Prepared In The Repository

The deployment assets now accept both domains during the transition:

- Caddy routes `learn.gripe`, `www.learn.gripe`, and `admin.learn.gripe` to
  `theme-web`.
- Internal Nginx accepts the same new hostnames.
- Nuxt receives `NUXT_PUBLIC_SITE_URL` from the production environment.
- The existing `legacy.example` routes remain in place for rollback and
  verification.

Deploy these image and gateway changes before switching public DNS.

## Cloudflare Setup

1. Add `learn.gripe` to the existing Cloudflare account as a new zone.
2. Review the imported records. Do not delete mail, SPF, DKIM, DMARC, OAuth,
   or verification records.
3. Create these web records using the expected VPS address:

   ```text
   A      @       2.25.85.201
   CNAME  www     learn.gripe
   A      admin   2.25.85.201
   ```

4. Keep the web records DNS-only while the origin certificate is being
   issued, if Caddy requires direct HTTP/HTTPS reachability.
5. Copy the two Cloudflare-assigned nameservers into the registrar's
   nameserver settings. Turn off DNSSEC at the registrar first if it is
   enabled.
6. Wait until Cloudflare reports the zone as active, then verify the
   nameservers from multiple public resolvers.
7. Do not add an AAAA record until direct IPv6 connectivity and TLS have been
   tested on the VPS.

## Application Cutover

After the new hostname serves the healthy application, update the untracked
`deployment/production.env` values as one change set:

```text
SERVER_BASE_URL=https://learn.gripe
STOREFRONT_BASE_URL=https://learn.gripe
NUXT_PUBLIC_SITE_URL=https://learn.gripe
CORS_ORIGINS=https://learn.gripe,https://www.learn.gripe,https://admin.learn.gripe
GOOGLE_MERCHANT_REDIRECT_URL=https://learn.gripe/api/admin/google-merchant/oauth/callback
GOOGLE_MERCHANT_POST_CONNECT_URL=https://learn.gripe/google-merchant
```

Keep `COOKIE_DOMAIN` empty. Host-only cookies make rollback and the
old/new-domain transition easier to reason about.

Before enabling customer traffic, also update:

- Cloudflare Turnstile allowed hostnames for `learn.gripe`,
  `www.learn.gripe`, and `admin.learn.gripe`.
- Google OAuth authorized origins and redirect URIs.
- Google Merchant Center redirect URI and post-connect URL.
- Payment provider webhook URLs and any provider allowlists.
- SMTP sender verification, SPF, DKIM, and DMARC before changing
  `SMTP_FROM` to `@learn.gripe`.
- Search Console, analytics, sitemap, canonical URL, and third-party
  integrations.

## Verification Order

1. Confirm all six Docker services are healthy.
2. Confirm the shared Caddy gateway contains the new hostnames.
3. Confirm `https://learn.gripe/healthz` returns `200`.
4. Confirm the storefront, `/api/v1/settings/site`, admin login, uploads,
   WebSocket upgrade, customer login, and checkout smoke tests.
5. Enable Cloudflare proxying and set SSL/TLS to `Full (strict)` after the
   origin certificate is valid.
6. Re-run the same smoke tests through the proxied hostname.
7. Change the production env to the new canonical domain and update the
   existing `commerce-platform` project. Do not create a second application
   project.

## Old Domain Redirect

Keep `legacy.example` live until the new domain is stable. Then change the
old storefront routes to permanent redirects:

```text
legacy.example/*       -> https://learn.gripe/$1
www.legacy.example/*   -> https://www.learn.gripe/$1
admin.legacy.example/* -> https://admin.learn.gripe/$1
```

Do not redirect `erp.legacy.example`; it belongs to the separate ERP
application boundary.

## Rollback

To roll back, keep the old Caddy routes and old production URLs, restore the
old Cloudflare DNS records, and update the existing `commerce-platform` project.
Do not delete the new Cloudflare zone or the new DNS records during the
observation window.
