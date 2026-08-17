# Site Quality Engine

## Terminal Scheduling State

The Site Quality control plane has one durable execution path:
`target -> job -> run -> evaluation -> finding -> event`. Operator checks,
scheduled checks, and finding rechecks all enter through `POST /jobs` semantics
and return a durable `job_id`; the worker claims jobs asynchronously and the
admin client polls job state.

Route-catalog changes synchronize target identity through an outbox event. A
low-frequency full reconciliation remains as repair coverage. Reconciliation
never performs provider work. Planning only creates jobs for already-enabled,
already-due targets, and dispatch only claims jobs that already exist.

Local development defaults the storefront target to `9199`, matching
`start-dev.ps1`. When an already-running storefront uses another port, set both
`STOREFRONT_BASE_URL` for the API and `SITE_QUALITY_ALLOWED_ORIGIN` for the
runner to the same `http://host.docker.internal:<port>` origin before enabling
`WORKER_SITE_QUALITY_ENABLED`.

## Purpose

Site Quality is a durable page-rendering quality engine in the pre-launch
checks domain. It turns repeated Lighthouse runner observations into
statistically confirmed engineering work.
It is not a URL availability, redirect, canonical, or sitemap checker; those
belong to URL Management.

It is also not the image-dimension engine. Stored image assets, their intrinsic
width/height, and generated display-size derivatives are checked and repaired
from the dedicated `上线前检查 / 图片尺寸` admin tab. Site Quality may still
report page-level delivery opportunities such as cache policy, compression,
and image formats, but it must not create work items for image dimensions.

The engine must answer four separate questions:

1. Which canonical storefront pages should be inspected?
2. Did a provider observation describe a root cause that an engineer can act on?
3. Is that root cause persistent enough to create or reopen work?
4. Has a claimed fix remained healthy long enough to verify it?

## Domain Boundary

URL Management owns route discovery and route correctness. Site Quality reads
the canonical, non-alias entries from that catalog as a candidate source, then
creates its own targets and never writes URL-management issue state.

| Record | Responsibility |
| --- | --- |
| Quality target | Canonical page identity, locale, source type, sampling tier |
| Quality job | Leased, idempotent unit of provider work |
| Lighthouse run | One immutable runner sample and its raw payload |
| Evaluation | The statistical decision over a job's samples |
| Finding | Current actionable root-cause work item |
| Finding event | Immutable lifecycle decision history |

The page identity is a normalized canonical URL. The source URL, provider
`finalUrl`, redirect result, locale, and release identifier are provenance, not
finding keys.

## Page Inventory And Sampling

The route catalog synchronizes active, canonical, checkable entries into Site
Quality targets. Each target records its source (`route_catalog` or `operator`)
and ledger sync marker. `route_entry_id` is the first identity key; canonical
URL is the migration key when a route changes URL. An invalidated ledger entry
disables its target without deleting historical runs, evaluations, findings, or
finding events.

Targets are assigned a sampling tier for reporting and future prioritization:

| Tier | Typical pages |
| --- | --- |
| critical | storefront, checkout, account, support and static conversion pages |
| standard | active product and category pages |
| background | articles and remaining indexable content |

An operator chooses a URL and strategy when a check is needed. A finding
recheck creates a priority job bound to exactly one `finding_id`; it cannot
fan out to other findings for the same target and strategy. Scheduled jobs
advance target cadence independently of provider completion.

Mobile and desktop remain separate because their Lighthouse execution and
remediation may differ. A manual recheck does not bypass the sampling or
confirmation policy.

## Audit Taxonomy

Raw Lighthouse audits are retained, but only rules in the audit registry can
create findings.

| Kind | Examples | May create a finding |
| --- | --- | --- |
| metric observation | FCP, LCP, INP, CLS, TBT, Speed Index | no |
| actionable opportunity | render blocking resources, unused CSS/JS, cache TTL, image delivery, compression | yes |
| diagnostic | third-party impact and informational audits | no by default |

An actionable rule has an explicit version and a minimum provider saving or
rule-specific threshold. This prevents score-only metric observations from
becoming duplicate work items. Unknown future Lighthouse audit IDs are retained
as raw evidence but are not auto-promoted.

## Statistical Decision Protocol

A job captures `sample_count` independent provider samples for one target and
strategy. The default is three samples with two confirmations required.

```text
observed samples -> median / confidence evaluation -> finding projection
```

- An actionable audit is confirmed only when it appears in at least two of
  three successful samples.
- Severity uses the confirmed rule and median savings/score, not a single
  outlier.
- A resolved finding reopens only after a confirmed recurrence.
- A resolved finding receives one clean evaluation only when the audit is
  absent from every successful sample in that evaluation.
- It becomes `verified` only after two consecutive clean evaluations.
- Open and acknowledged findings never disappear automatically.
- Provider failures never count as clean evidence.

Every decision stores the sample count, confirmations, rule version, confidence
and representative evidence. This makes the operator-visible state auditable.

## Execution And Reliability

Jobs are persisted before execution. Workers claim them with database row locks
and leases, so multiple API replicas cannot execute the same job concurrently.

- job idempotency keys prevent duplicate enqueueing;
- lease generation, expiry, and heartbeat are compare-and-set guards for every
  run, evaluation, finding, and job completion write;
- stale leases can be reclaimed after timeout;
- failures use bounded exponential backoff with jitter;
- `429`, network, and `5xx` failures consume attempts but never mutate
  findings;
- a configurable concurrency and request interval bound internal runner load;
- the dedicated Node runner pins the Chromium and Lighthouse versions;
- runner authentication uses an internal deployment token, never an
  operator-managed or third-party API key;
- a job and its resulting finding projection commit atomically;
- operational health exposes queue age, failed jobs, lease recovery, provider
  errors, and last successful evaluation.

The admin page must surface a live operational summary, not just historical
runs. At minimum, it should show:

- runner readiness and configured storefront origin;
- target inventory;
- claimable jobs, processing jobs, dead letters, and stale leases;
- provider slot capacity and next-available time;
- latest successful sample across scheduled and operator jobs, plus
  warning messages whenever current retry, dead-letter, stale-lease, or latest
  sample-failure state degrades the summary.

The worker therefore runs three ordered stages: low-frequency full
reconciliation, due-target planning, and ready-job processing. It is not
allowed to create work from an HTTP request or to execute a direct `/run`
operation.

## Provenance And Field Data

Each raw sample retains the runner's normalized Lighthouse report plus
normalized provenance:
canonical target, provider final URL, strategy, Lighthouse version/environment
when supplied, sampled timestamp, and storefront release identifier. Lab data
and CrUX field data are separate signals. Field data may later be attached to
the same canonical target for impact prioritization, but it must never be
presented as a substitute for an individual Lighthouse sample.

## Lifecycle

```text
open -> acknowledged -> resolved -> verified
  ^                      |            |
  +----------------------+------------+
        confirmed recurrence only
```

Resolving still requires an operator note. Verification is an evidence-backed
decision, not the absence of an audit from one transient provider response.

## Internal Runner Boundary

The Go API owns planning, database leases, retries, evaluation, and finding
lifecycle. It never launches a browser. A separate internal-only Node
`site-quality-runner` container owns Chromium and Lighthouse execution.

```text
URL Management catalog
  -> Site Quality target/job/evaluation
  -> internal Lighthouse runner
  -> normalized Lighthouse report
  -> durable run evidence and findings
```

The runner has no public port in production. Its run endpoint requires a
deployment token, accepts only the configured storefront origin, rejects
credential-bearing URLs and private DNS answers in production, and is capped
by timeout, process, CPU, memory, and concurrency limits. The Go API applies
the same storefront-origin constraint before scheduling or capturing a run.
