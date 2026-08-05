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
- Alert on spikes in `413`, `429`, `5xx`, origin connection count, and upload
  volume.

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

Before opening public DNS:

- Confirm the origin address is not reachable from an unrelated public network.
- Send requests through the CDN and verify the Go log records the real client
  address, not the CDN or edge container address.
- Send a forged forwarding header through the CDN and confirm it does not
  replace the logged client address.
- Confirm upload paths return `413` at the intended body limit and `429` after
  repeated requests.
- Confirm static media can be cached while authenticated API responses are not.
- Confirm CDN/WAF alerts are connected to the on-call channel.
