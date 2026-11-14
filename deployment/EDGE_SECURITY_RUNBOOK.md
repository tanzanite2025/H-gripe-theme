# Edge Security Runbook

This runbook defines the production boundary for CDN, WAF, client IP handling,
and upload protection. Nginx and the Go API are origin controls; they are not
a replacement for a CDN or a network-layer DDoS service.

## Required Topology

```text
Client
  -> CDN/WAF
  -> edge reverse proxy
  -> Nginx web container
  -> Go API / Nuxt services
```

The origin host must not publish the web container, API container, or admin
container directly to the public internet. Only the edge reverse proxy should
be reachable from the host ingress.

## CDN/WAF Configuration

Configure the selected CDN/WAF provider to:

- Proxy the storefront and admin hostnames.
- Enable managed WAF rules and network-layer DDoS protection.
- Reject requests that target the origin IP with a host header outside the
  expected storefront and admin hostnames.
- Do not cache `/api/`, `/health`, admin pages, upload endpoints, or responses
  that contain account/session data.
- Cache immutable static assets and public media only when the cache key does
  not include cookies or authorization headers.
- Cache successful `/_ipx/*` image derivatives with a cache key that includes
  the complete path and query string, while respecting the origin cache
  headers. Do not cache `429`, `4xx`, or `5xx` responses for this path. This
  keeps repeat image sizes at the CDN and prevents an image view from
  repeatedly invoking Sharp at the origin.
- Alert on spikes in `413`, `429`, `5xx`, origin connection count, and upload
  volume.

### Mobile in-app browser compatibility

Apply the following protocol policy to every enabled storefront and alias
hostname generated from `ops_domain_bindings`:

- The edge generator publishes an explicit HTTP-to-HTTPS `301` for GET/HEAD
  navigation and `308` for other methods so request bodies survive. It keeps
  the configured host, original encoded path and query. The deployment
  preflight rejects a generated route that omits this policy. Check the
  response from both the public hostname and the origin's Caddy listener:
  during this investigation the origin returned Caddy `308`, and Cloudflare
  relayed it. That exact legacy response is a plausible cause of the supplied
  Android WebView failure. Some old embedded browsers fail before following
  `308`; a page-level notice cannot help before HTML loads. Confirm that GET
  requests return `301` after the new generated route is deployed. Do not put
  a blanket `301` on non-GET methods, because clients may turn POST into GET.
  Nuxt does not redirect private HTTP origins or trust arbitrary forwarded
  headers to decide the public protocol.
- Keep the provider SSL/TLS mode at Full (strict) and keep certificates active
  for every alias. Do not rely on a page-level redirect that exists only for
  the current canonical hostname.
- Inspect the CDN security event for a failed request (Ray ID, timestamp,
  hostname, path, country and action). A `403` with `Cf-Mitigated: challenge`
  means the CDN stopped the request before Nuxt rendered any page. Review the
  rule that actually triggered, and use the provider's supported compatibility
  controls or a narrowly scoped exception only when the event justifies it.
  Do not blanket-allow user-agent strings or disable storefront bot protection:
  user agents are spoofable, and a curl challenge alone does not prove that a
  real customer's WeChat request received the same challenge.

The storefront displays one centralized in-app browser notice after the page
loads. It explains the app-menu action and offers a selectable current HTTPS
URL with clipboard and manual-copy fallbacks. It derives the URL from the
browser location, updates after navigation and never claims that a new tab can
force the app to launch the system browser. Detection is client-side and does
not change the SSR/cache response. Some apps expose the same user agent as an
ordinary browser; absence of a notice does not indicate that sign-in or payment
will work. The notice cannot repair DNS, TLS, app-level domain blocking or a
CDN rejection that prevents the page from loading.

When replacing a domain, update the domain bindings, DNS/CDN zone, active TLS
certificates, deployment public URL (`NUXT_PUBLIC_SITE_URL`/`NUXT_SITE_URL`),
and any payment/OAuth callback allowlists and provider domain verification.
The storefront image workflow keeps `STOREFRONT_SITE_URL` synchronized: for a
planned canonical-domain change, manually dispatch `Publish Production Images`
and enter the new HTTPS origin in the `storefront_site_url` input. The run
builds all images with that origin and, only after every image publishes
successfully, a separate job with only Actions-variable write permission saves
it to the repository variable. Leave the input blank for routine publishes.
This planned-migration path skips only the live-origin comparison, because the
new host may not be serving the storefront yet; it still validates the value
as an HTTPS origin and runs the normal release checks. It does not change Ops
domain bindings, deployment environment, DNS, CDN or certificates. Complete
those deployment steps after publishing; routine publishing stays blocked
until the live canonical origin matches the new repository variable. Routine
publishes fetch the current homepage and compare its canonical origin with the
repository variable; a stale or unavailable live canonical check blocks
publication with a diagnostic. This catches a missed variable update after an
out-of-band domain change. The Nuxt build
automatically adds the canonical host to the optimized image allowlist;
`STOREFRONT_IMAGE_DOMAINS` remains for additional image hosts. `deploy.sh`
checks deployment URLs against the generated bindings. Keep old DNS/TLS active
while old links remain in circulation.
An origin-only redirect target (for example `https://new.example` or
`https://new.example/`) preserves the original request URI in both generated
Caddy configuration and Nginx preview. A target containing an explicit path,
query or fragment remains an exact destination. Fragments are never sent to
the gateway, so verify browser anchor behavior separately. Both the initial
HTTP upgrade and the following old-domain migration return `301` to GET/HEAD
requests; non-navigation methods use `308` at both steps to preserve bodies.
Do not place gateway placeholders in the stored URL.

Before promotion, verify HTTP and HTTPS on the canonical host, every alias,
the admin host, and the old redirect host. Check a deep locale path with an
encoded query, GET/HEAD navigation and a POST redirect that retains its body.
Test the resulting login and checkout flows in real iOS/Android WeChat,
Facebook and WhatsApp on relevant customer networks; desktop user-agent
emulation does not certify embedded WebView behavior. Run the generated Caddy
HTTP/TLS regression with Caddy installed:

```sh
cd go-backend
go test ./internal/service -run 'TestGeneratedCaddyPreservesMobileNavigationAndPostBodiesAcrossDomainReplacement|TestDomainMigrationRedirect' -count=1 -v
```

Without Caddy on PATH, the HTTP/TLS regression explicitly skips; a skipped run
is not gateway validation. The production image workflow installs a pinned
Caddy release and makes a missing binary fail the backend verification job.

At the origin firewall, allow HTTP/HTTPS only from the CDN provider's current
published egress ranges. If another load balancer is used, allow only that
load balancer's egress ranges instead. Do not copy a historical provider IP
list without a maintenance process.

## Edge Rate Rules

Create provider-side rate rules before traffic reaches the origin:

| Path | Suggested key | Initial policy |
| --- | --- | --- |
| `/api/admin/media/assets` | authenticated account and client IP | 3 requests/minute, small burst |
| `/api/v1/registration/warranty/claim` | client IP | 1 request/minute |
| `/api/v1/customer-service/attachments` | client IP | 6 requests/minute |
| `/api/v1/suggestion-feedback/upload` | client IP | 6 requests/minute |
| `/api/v1/showcase/upload` | client IP | 6 requests/minute |
| login, verification, and password-reset paths | client IP plus account/email where supported | challenge or block on repeated failures |

The edge rule is an early rejection layer. The matching application and
Nginx limits must remain enabled because requests may arrive through another
trusted internal path.

## Client IP Chain

The chain must be configured as follows:

1. The CDN validates its own incoming forwarding headers and sends the client
   address in its documented forwarding header.
2. The edge reverse proxy trusts only the CDN's published ranges and rebuilds
   the forwarded chain.
3. Nginx trusts only the private network of the edge reverse proxy and uses
   `X-Forwarded-For` recursively.
4. Gin trusts only the private network or load balancer addresses that can
   reach the API. The production environment must set `TRUSTED_PROXIES`.

Never configure the application to trust arbitrary client-supplied
`X-Forwarded-For`, `X-Real-IP`, or provider-specific headers. After deployment,
verify the Go access log `ip` field with two separate public networks and
confirm that rate limits are independent for those addresses.

The bundled Caddy configuration contains the currently published Cloudflare
proxy ranges as a starting point. Review and update that list whenever the
provider publishes a change. For a different CDN, replace the list with that
provider's current ranges and keep the Nginx private-proxy trust boundary.

## Upload Safety

Keep all of the following aligned:

- CDN/WAF request body limit.
- Nginx `client_max_body_size`.
- Go `MaxBytesReader` limit.
- Per-file and combined-file validation.
- Per-account cumulative storage quota.
- Per-IP and per-account rate limits.

Do not increase only one layer. A larger edge limit than the application limit
still consumes bandwidth and temporary storage before the request is rejected.

## Deployment Verification

### Public SEO/GEO crawler contract

The storefront origin must return `200` for `/robots.txt`, `/sitemap.xml`, and
public guide routes when requested without cookies or an allow-listed browser
session. Do not apply the admin/commercial-intelligence bot deny rule to the
storefront server: it would block legitimate indexing and answer engines before
they can read SSR HTML. Bot restrictions may remain on the admin hostname.

Run these probes through the real CDN hostname after every edge or WAF change:

```text
GET /robots.txt       User-Agent: Googlebot
GET /robots.txt       User-Agent: GPTBot
GET /sitemap.xml      User-Agent: Bingbot
GET /tools/freehub-cassette-fitment-engine
                      User-Agent: PerplexityBot
```

Require `200`, a `Sitemap:` line pointing at the canonical production host,
and guide HTML containing `TechArticle`, `HowTo`, `Dataset`, canonical, and
hreflang markup. A `403` from any of these public probes is a release blocker.

The storefront robots policy may still disallow commercial intelligence
crawlers such as AhrefsBot or SemrushBot in `robots.txt`; that advisory policy
must not be implemented as a blanket Nginx `return 403` on the storefront.

Before opening public DNS:

- Confirm the origin address is not reachable from an unrelated public network.
- Send requests through the CDN and verify the Go log records the real client
  address, not the CDN or edge container address.
- Send a forged forwarding header through the CDN and confirm it does not
  replace the logged client address.
- Confirm upload paths return `413` at the intended body limit and `429` after
  repeated requests.
- Confirm static media can be cached while authenticated API responses are not.
- Confirm `/_internal/image-health` succeeds in the storefront container and
  `/_internal/image-health?source=internal` succeeds from the web container.
  The second check verifies the private `web:8080` upload origin, the upload
  mount, IPX, and Sharp together without using public DNS, CDN, or TLS.
- Configure the CDN rule for `/_ipx/*`, request one known derivative twice,
  and confirm the second response is served from the edge cache. A cache miss
  after expiry is expected; a persistent miss indicates the CDN rule or cache
  key is wrong.
- Confirm CDN/WAF alerts are connected to the on-call channel.
