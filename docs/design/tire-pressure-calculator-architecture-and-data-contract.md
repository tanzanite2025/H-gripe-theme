# Tire Pressure Calculator Architecture and Data Contract

Status: current implementation baseline, updated 2026-10-04.

This document is the implementation owner for the tire-pressure demonstration page. Older dynamic tire-pressure and impedance notes remain historical design records; when they conflict with this document or the code, this document and the tests are authoritative.

## Scope and routes

The feature has one primary user-facing guide page and one compatibility route:

| Route | Responsibility | Current state |
| --- | --- | --- |
| `/guides/tireguides/tire-pressure` | Interactive force and contact-area demonstration with the explanation tab. It owns the calculator inputs, selected product pressure display, and dynamics output. | This is the primary calculation entry point. |
| `/guides/tireguides/tire-pressure-calculator` | Compatibility route that keeps the original direct calculator URL and its page-level FAQ/feedback metadata. | It reuses the same calculator component; it is not a second physics implementation. |

The primary guide page keeps the calculator and its explanation together so users do not need to move between scattered guide pages. The product selector and pressure range are read from the same local Schwalbe catalog projection wherever the shared calculator component is mounted. A future sourced product-standard table can still be added to the guide without introducing another calculation implementation.

## Frontend ownership

| File | Responsibility | Boundary |
| --- | --- | --- |
| `app/pages/guides/tire-pressure.vue` | Route shell and SEO metadata for the combined pressure/force guide. | Calculation state remains in the mounted calculator component. |
| `app/pages/guides/tire-pressure-calculator.vue` | Route shell, back link, SEO/JSON-LD metadata and page-level FAQ/feedback metadata. | Does not implement physics or catalog parsing. |
| `app/components/tireguides/TirePressureGuide.vue` | Loads the page message shard and mounts the combined pressure/force guide. | Does not fetch the product catalog. |
| `app/components/TirePressureSection.vue` | Calculator/details tab state and explanatory content. | It mounts the shared calculator component but does not implement physics. |
| `app/components/tireguides/tirepressure/TirePressureProductModelSelector.vue` | Search-as-you-type pressure-reference product selection and selected catalog row. | It emits the narrow pressure-reference item; it does not calculate pressure or force. |
| `app/components/tireguides/tirepressure/TirePressureControls.vue` | Renders user controls and consumes the calculator model provided by the parent. | It does not own API requests or formula logic. |
| `app/components/tireguides/tirepressure/TirePressureCalculator.vue` | Composes selector, controls and result metrics; derives display-only labels from backend data. | It must not gain a second physics implementation. |
| `app/composables/useTirePressureCalculatorDynamicsRequest.ts` | Owns the dynamics request body, normalized input key, debounce, bounded cache, stale-response protection, retry and unmount cleanup. | It returns backend data and request state; it does not render UI or translate labels. |
| `app/data/tireguides/schwalbeCatalog.ts` | Validates general catalog responses and the narrower pressure-reference response shape. | The calculator uses `SchwalbeTirePressureReferenceCatalogItem`; it does not depend on the general product attribute type. |
| `app/i18n/page-messages/guidesTirePressure/*.json` | All visible page, calculator and selector UI strings. | Do not add calculator copy to component source; FAQ answers are database migration content. |

The next optional frontend split is a result-metrics component if the template grows again. It should receive already-normalized backend values and formatting callbacks; it should not own request state. The pressure calculator already consumes a dedicated `SchwalbeTirePressureReferenceCatalogItem`; general Schwalbe catalog consumers continue to use `SchwalbeTireCatalogItem`.

## Backend ownership

| File/package | Responsibility |
| --- | --- |
| `internal/api/v1/tirepressure/tire_pressure_reference_http_handler.go` | Read-only calculator DTO from local Schwalbe product rows: model identity, ETRTO/inch label, pressure range and source date. It does not expose the general product attribute set. |
| `internal/api/v1/tirepressure/tire_pressure_engine_http_handler.go` | JSON validation, stable error mapping, rate-limited dynamics and metadata endpoints, response envelope and warnings. |
| `internal/domain/tirepressure/tire_pressure_dynamic_force_and_contact_area_engine.go` | Request validation, deterministic load split, ground-frame force decomposition, contact-area estimate, pressure comparison, vertical-deformation reference estimate and fixed friction baseline. |
| `internal/domain/tirepressure/tire_pressure_wet_pressure_demonstration_model.go` | Isolated wet-pressure demonstration proxy, equivalent-pressure reduction and wet reference-grip output. It depends on the engine's wheel result but owns wet-model constants and formulas. |
| `internal/domain/tirepressure/tire_pressure_vertical_deformation_reference_data.go` and `*_v1.json` | Versioned local reference dataset and validation for the generic vertical-deformation estimate. |
| `internal/repository/schwalbe_tire_catalog_repository.go` | Reads the persisted local product database. It is the only catalog storage dependency of the reference handler. |
| `internal/api/v1/router.go` | Registers the two tire-pressure route groups. |

The catalog handler and dynamics handler do not call each other. The frontend reads the catalog once, derives the selected product pressure midpoint, then sends the calculation input to the dynamics endpoint. This keeps product data failures separate from calculation failures.

## Data flow

```text
local Schwalbe database
  -> GET /api/v1/engineering/tire-pressure/reference-data/catalog
  -> schwalbeCatalog adapter validates {product, pressure, source_checked_at}
  -> ProductModelSelector emits selected row
  -> midpoint(min_pressure, max_pressure) becomes shared front/rear input
  -> POST /api/v1/engineering/tire-pressure/dynamics
  -> Go domain engine
  -> force, load, contact area, deformation reference, friction status
  -> calculator result metrics
```

The catalog endpoint is a local database read. It is not an external Schwalbe request and it does not reach the dynamics endpoint. The dynamics endpoint accepts the selected values as input and has no dependency on the selector component.

## Input rules and model boundaries

### Product pressure midpoint

When a selected product has usable minimum and maximum pressure values, the calculator uses:

```text
reference_pressure_psi = round((minimum_pressure_psi + maximum_pressure_psi) / 2, 1)
```

The midpoint is a transparent demonstration input shared by front and rear wheels. It is not a pressure recommendation. A missing or incomplete range leaves the calculation pending and disables pressure comparison controls.

### Load and force

The backend resolves front/rear load from measured wheel loads when supplied; otherwise it applies the selected riding-position split to rider plus bike mass. The response marks the load source. Lean and speed affect the ground-frame lateral demand and equivalent turn radius; the gravity load remains a vertical component.

### Contact area

The current comparison is the first-order static estimate:

```text
A_contact ~= Fz / P
```

The backend returns `estimated_static_contact_area_cm2`. Width-limited equivalent circular dimensions are display geometry for comparing inputs, not measured tire footprints. The API does not return an SVG shape or claim an ellipse measurement.

### Tire body and road

The calculation uses one generic rubber baseline with `tire_body_normalization_factor = 1` and one `FLAT_ROAD` surface. Tread, compound, casing construction, TPI, weather and road-texture choices are not hidden inputs and must not be added as fake multipliers. Nominal tire width constrains the displayed contact dimensions; it is not a measured inflated width unless the request explicitly supplies one.

### Pressure and friction

The API returns a `pressure_friction_coefficient` object with:

- `nominal_coefficient = 0.76` for the fixed flat-road demonstration baseline;
- `estimated_coefficient` equal to the nominal value;
- `relative_coefficient_index = 1`;
- `pressure_effect_applied = false`;
- `data_status = INSUFFICIENT_MEASUREMENT_SUPPORT`.

This is deliberate. A pressure-dependent friction curve needs matched measurements for the same tire, load, surface and speed. Contact-area change alone cannot be promoted to a friction coefficient change. Until that dataset exists, the UI must describe fixed μ as a relative demonstration baseline and must not claim a pressure-derived grip gain.

### Vertical deformation

The vertical-deformation output is a versioned generic reference-tire estimate. The local data contains a measured reference condition and the application uses a first-order pressure scaling assumption. It must remain labelled as a reference model and cannot be presented as a selected Schwalbe model measurement.

### Wet demonstration

The optional wet control is a bounded demonstration proxy using a fixed 1 mm water-film baseline and the selected product's minimum pressure for an equivalent-area comparison. The request must include both front/rear operating pressures and both front/rear minimum pressures; validation rejects a wet request without that complete pair. It returns `pressure_reduction_psi` and `pressure_reduction_pct`, calculated from reference pressure minus the final equivalent pressure after the minimum-pressure floor is applied. These are model-equivalent reductions, not a universal minus-5 or minus-7 PSI rule, a wet-grip guarantee or a calibrated tire-road law. Its warning text and FAQ must remain aligned with this boundary.

## API contracts

### Catalog snapshot

`GET /api/v1/engineering/tire-pressure/reference-data/catalog` returns `{ code: 0, data: { schema_version, items } }`. Each item contains a product projection, a pressure projection and `source_checked_at`. The response is built from the local repository and is cacheable for a short public TTL.

### Dynamics

`POST /api/v1/engineering/tire-pressure/dynamics` accepts a single JSON object. The current page sends rider weight, bike weight, nominal tire width, speed, riding position, `FLAT_ROAD`, lean angle and the selected shared front/rear pressure. Optional pressure-comparison and wet-demonstration fields are sent only when enabled. Wet demonstration requests must also send the selected minimum pressure for both wheels.

The success envelope is `{ code: 0, data: { model_version, load_source, front_load_kg, rear_load_kg, dynamics, warnings } }`. The response is an estimate and never a production pressure recommendation. Handler validation rejects unknown fields, invalid numbers and unsupported enum values.

`POST /api/v1/engineering/tire-pressure/solve` remains a metadata validation endpoint. It can return `MODEL_NOT_CALIBRATED` and a null recommendation; it is not used by the calculator page's interactive dynamics request.

## FAQ, i18n and migration ownership

- FAQ route/page content is owned by migration `372_add_tire_pressure_calculator_faq_content.up.sql` and its rollback. The products layout places the FAQ slot above feedback; the page metadata supplies the calculator feedback thread key.
- Calculator and guide copy is owned by `app/i18n/page-messages/guidesTirePressure/en.json` and `zh_cn.json`; other locale shards follow the page-message loading convention.
- Route discovery, footer, breadcrumb and navigation contracts must keep both the primary guide route and the compatibility calculator route valid. Update their matching tests when changing a path.
- The old HTML prototypes and design-only route names are reference artifacts. They are not runtime dependencies and must not be copied into the calculator.

## Follow-up split order

1. Keep the current request composable as the only owner of dynamics transport, cache and stale-response handling.
2. If result markup grows, extract a display-only metrics component from `TirePressureCalculator.vue`.
3. Split the pressure-catalog adapter from the general Schwalbe catalog adapter after shared types and tests are moved together.
4. Add a pressure-to-friction dataset only after matched measurements, provenance, fitting method and regression tests are available. Until then, preserve the explicit fixed-μ status.

## Claims that must not be made

The calculator must not turn the generic model into a Schwalbe product test, infer tread or compound behavior, claim a universal wet-pressure reduction, state that contact-area increase equals grip increase, or output a safety/compatibility conclusion. Manufacturer pressure limits remain product, rim, wheel and bicycle documentation concerns and are not generated by the dynamics model.
