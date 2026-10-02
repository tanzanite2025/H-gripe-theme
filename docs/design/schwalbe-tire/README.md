# Schwalbe tire docs and implementation map

This directory contains the design and implementation references for the Schwalbe tire catalog and selector. Use the owner document below for each contract; other documents should link to it instead of copying its full rules.

## Canonical document owners

| Document | Owns | Does not own |
| --- | --- | --- |
| [Phase 1 implementation guide](./phase1-schwalbe-tire-system-template-implementation-guide.md) | Admin product workflow, catalog-to-Product handoff, validation and migration/deployment operations | The canonical field matrix or selector-page interaction contract |
| [Catalog specification matrix](./schwalbe-master-catalog-specification-matrix.md) | The 19 template fields, `is_filterable`, storage constraints, imported seed snapshot facts and verified enums | Page controls, URL state, SSR and SEO behavior |
| [Phase 2 page implementation guide](./phase2-standalone-page-implementation-guide.md) | Selector page behavior, filters, URL contract, SSR pagination, component boundaries and current implementation status | Product-template schema or SEO indexing policy |
| [Wheel-size page tabs design](./schwalbe-wheel-size-page-tabs-design.md) | Implemented page-level wheel diameter + BSD navigation and card size projection | Product-template schema or SEO/indexing policy |
| [Selector SEO/GEO specification](./schwalbe-tire-selector-and-geo-specification.md) | Indexing/canonical policy, structured data and public content claims | Catalog fields, filters and page implementation details |
| [FAQ content guide](./schwalbe-faq-knowledge-base-input-guide.md) | FAQ sources, answer boundaries and editorial workflow | FAQ runtime integration or Product data |

## Reference artifacts

- `preview-schwalbe-tire-selector.html` is a historical visual prototype. Its mock products, compatibility output, counts, source links and promotional claims are not implementation or product facts.
- `preview-schwalbe-tread-profile-guide.html` is an unreferenced historical concept. Treat its model counts and product claims as stale unless separately verified against current official sources.
- `schwalbe-tire-rim-width-matching-design.md` owns the implemented rim-system and inner-width reference contract: the selector uses reviewed Hooked/Hookless facts and cards show the narrow DT Swiss reference derived from ETRTO nominal width. Broad Schwalbe/ETRTO combination ranges are not part of the selector contract.
- `schwalbe-wheel-size-page-tabs-design.md` records the implemented move of wheel size + BSD from the filter drawer to page-level size navigation, plus wheel diameter/BSD on each card. Keep the Phase 2 guide as the single runtime status summary.
- Runtime behavior is owned by the Nuxt page/components, filter/query modules and backend API. Migration and seed files are the source for imported catalog facts; the matrix section 8 is the human-readable baseline for the checked-in seed snapshot and its observed enums. Environment-specific database row counts are not a stable cross-environment fact and must be verified during deployment.

## Data exposure rule

`source_url` remains in the candidate database for provenance. The public catalog response and storefront hydration payload must not include it. `source_checked_at` may be returned to the selector for the visible checked-date label.

The full-list `GET /products/schwalbe-tire-catalog` remains the Admin Products/Customs source until both entry points move to remote model search and Article No. loading. The storefront selector uses `GET /products/schwalbe-tire-catalog/selector` for its page, result count and search-scoped filter options; do not add a default limit to the Admin endpoint.

## Change routing

- A template field, type, unit, required flag or `is_filterable` change updates the matrix first, then the Phase 1 workflow if needed.
- A selector control, facet rule, URL key, sort, pagination or empty/error behavior updates the Phase 2 guide and the matching code/tests.
- A change to where users choose wheel diameter + BSD updates the wheel-size tabs design first, then the Phase 2 guide and matching code/tests.
- SEO, canonical, robots or structured-data behavior updates the SEO/GEO specification and page metadata.
- FAQ answer sourcing updates the FAQ content guide; FAQ query/rendering changes update the Phase 2 runtime section.
- Keep one selector implementation-state summary in the Phase 2 guide. Other documents should link to it instead of maintaining copied page-status lists.
