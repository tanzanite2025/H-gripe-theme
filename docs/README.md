# Project Documentation

This directory is a project-level documentation hub. It is not the source of truth for runtime behavior; the current code and root `README.md` win when documents disagree.

## Current Entry Points

- Project overview: `../README.md`
- Backend guide: `../go-backend/README.md`
- Backend API notes: `../go-backend/API.md`
- Backend quick start: `../go-backend/QUICK_START.md`
- Backend deployment notes: `../go-backend/DEPLOYMENT.md`
- Current production readiness status: `ops/production-readiness-status.md`
- CDN/WAF and responsive image cache runbook: `../deployment/EDGE_SECURITY_RUNBOOK.md`
- Cross-border email delivery and domain reputation guide: `ops/cross-border-email-delivery-and-domain-reputation-guide.md`
- Backend module notes: `../go-backend/docs/`
- Backend security follow-ups: `../go-backend/docs/SECURITY_FOLLOW_UPS.md`
- Form Honeypot anti-spam architecture: `security/form-honeypot-anti-spam-architecture.md`
- Storefront Content Security Policy: `security/content-security-policy.md`
- Admin console guide: `../go-backend/web/admin/README.md`
- Distributed locks, task state machine, and idempotency architecture: `design/distributed-lock-task-state-machine-idempotency-architecture.md`
- Payment channel and risk architecture: `design/payment-channel-domain-architecture.md`
- Internal payment-provider cost accounting and profit snapshots: `design/payment-provider-cost-accounting-architecture.md`
- Payment gateway onboarding and multi-currency settlement guide: `design/payment-gateway-onboarding-and-settlement-guide.md`
- Storefront hub: `../nuxt-i18n/docs/README.md`
- Storefront active notes: `../nuxt-i18n/docs/notes/`
- Storefront archive: `../nuxt-i18n/docs/archive/`
- Storefront i18n current status: `../nuxt-i18n/docs/notes/I18N-CURRENT-STATUS.md`
- Storefront payment UX rules: `design/storefront-payment-ux.md`
- QUICK selection flow architecture: `design/quick-buy-configuration-architecture.md`
- Order evidence package architecture: `design/order-evidence-package-architecture.md`
- Wheelset fit questionnaire specification: `design/wheelset-fit-questionnaire-specification.md`
- Technical algorithm and simulation standards: `design/technical-algorithm-engineering-standards.md`
- Spoke stress relief materials science whitepaper: `design/spoke-stress-relief-materials-science.md`
- Storefront recommendation UX and algorithm contract: `design/storefront-recommendation-ux.md`
- Storefront Maple UI font performance strategy: `design/storefront-font-performance-strategy.md`
- Storefront light-theme color tokens: `design/emerald-light-theme-palette.md`
- Ops control plane and workflow engine design: `design/ops-control-plane-workflow-engine.md`
- Product supplier-cost and profitability isolation design: `design/product-supplier-cost-profitability-isolation-architecture.md`
- Product template and configurable options architecture: `design/product-template-and-options-architecture.md`
- Referral and loyalty reward system architecture: `design/referral-reward-system-longterm-architecture.md`
- Shipping quote, route plan, and price-lock architecture: `design/shipping-quote-plan-architecture.md`
- Multi-carrier logistics three-tier short-link architecture and carrier SPI: `design/cross-border-logistics-three-tier-architecture.md`
- Yanwen small-packet cross-border logistics implementation specification: `design/yanwen-logistics-hub-integration-architecture.md`
- 4PX bulky cross-border logistics and overseas warehouse implementation specification: `design/4px-logistics-hub-integration-architecture.md`
- Admin dashboard command center and operational intelligence architecture: `design/admin-dashboard-command-center-architecture.md`
- Money, pricing pipeline, display/read model, and Outbox implementation status: `design/money-pricing-display-outbox-status.md`
- SEO architecture: `seo/SEO_SYSTEM_ARCHITECTURE.md`
- E-commerce URL and SEO target architecture: `seo/ECOMMERCE_URL_ARCHITECTURE.md`
- SEO documentation index: `seo/README.md`
- URL management domain architecture, diagnostic fixes, and performance optimization spec: `design/url-management-domain-architecture-and-fix-guide.md`
- Customer service workbench deep interaction bugs and UX/DX remediation spec: `design/customer-service-workbench-deep-bugs-and-interaction-spec.md`
- Customer-service conversation lifecycle and inbox archive architecture: `design/customer-service-conversation-lifecycle-and-inbox-architecture.md`
- Customer-service retention operations runbook: `ops/customer-service-retention-runbook.md`
- Transactional email, template engine, and after-sales notification architecture: `design/transactional-email-and-notification-architecture.md`
- Order status, domain-event, and email-template contract: `design/order-status-event-template-contract.md`
- Transactional email and after-sales implementation status: `design/transactional-email-notification-status.md`

## Archive

- `archive/project/` contains old project status, completion, deployment-ready, payment, storage, monitoring, and fix reports.
- `archive/ops/` contains completed domain cutover and brand-renaming operation records.
- `archive/backend/` contains old backend security, quality, frontend-page, and completion reports.
- `archive/audit/` contains old audit and optimization reports.
- `archive/refactoring/` contains old handler refactoring completion reports.
- `archive/bugfix/` contains old bugfix reports.
- `archive/splitting/` contains old file-splitting reports.
- `archive/optimization/` contains old optimization and cleanup reports.

Archived files are historical context only. They should not be used to claim production readiness, feature completeness, benchmark numbers, or current architecture.

## Maintenance Rules

- Keep active docs short, factual, and tied to current code.
- Move one-off completion reports into `archive/` after the work is done.
- Avoid claims like "production ready", "100% complete", or exact performance gains unless they are backed by current tests or measurements.
- Prefer one source of truth for each area: backend docs under `go-backend/`, storefront notes under `nuxt-i18n/`, project-level docs under `docs/`.
- Remove legacy WordPress compatibility docs unless they describe an explicit migration-only tool.

Last updated: 2026-09-21.
