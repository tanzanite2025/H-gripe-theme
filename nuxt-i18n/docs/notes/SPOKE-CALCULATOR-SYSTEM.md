# Spoke Calculator System / 辐条计算器系统手册

Last updated: 2026-10-01

Status: Active reference. Re-audit when the Go spoke API contract, the manual
calculator input model, or the recorded-result projection changes.

## 1. Non-negotiable boundary / 不可混淆的边界

The spoke page contains two independent systems. They share the page and the
catalog API, but they do not share calculation state.

### 1.1 Manual calculator / 手工计算系统

This is the upper wizard and calculator card. It accepts the dimensions and
build parameters entered by the user:

- wizard order: spoke head type, front/rear ERD, front/rear rim offsets and
  alternating drilling offsets, front/rear hub PCD/WL/WR, then flange-hole
  diameter for the shared length-measurement correction;
- spoke head type (J-bend or straight-pull);
- front and rear ERD;
- front and rear rim offset (asymmetric rim offset) and alternating drilling
  offset;
- front and rear PCD, WL, and WR;
- front and rear flange-hole diameter for J-bend and straight-pull length
  measurement;
- spoke count, crossing pattern, nipple settings, and physical correction
  inputs.

The manual calculator must calculate from those values. Catalog selection must
never fill, overwrite, or satisfy the manual geometry fields.

The browser payload created by `app/utils/spokeCalculatorPayload.ts` sends empty
`rimId` and `hubId` values and sends the measured geometry fields instead.
`hasSpokeCalculationGeometry()` is the gate for a calculation request. A
change in the lower catalog/search area must not clear or replace the manual
draft.

Relevant files:

- `app/components/SpokeCalculatorBlueprint.vue`
- `app/components/SpokeCalculatorWheelPanel.vue`
- `app/components/SpokeCalculatorBuildSettings.vue`
- `app/components/SpokeHeadTypeStep.vue`
- `app/components/SpokeERDStep.vue`
- `app/components/SpokePCDStep.vue`
- `app/components/SpokeAlternatingDrillingStep.vue`
- `app/components/SpokeHoleEngagementStep.vue`
- `app/composables/useSpokeCalculatorWizard.ts`
- `app/composables/useSpokeCalculatorManualOptions.ts`
- `app/composables/useSpokeCalculatorRun.ts`
- `app/utils/spokeCalculatorPayload.ts`

### 1.2 Recorded catalog/search system / 已录入结果目录系统

This is the independent card below the calculator. It is for finding records
entered and verified by the backend:

- `SpokeCalculatorCatalogPanel.vue` owns the front/rear catalog selections;
- `SpokeSmartSearch.vue` combines those selections with the keyword query to
  filter recorded build metadata/results;
- `useSpokeCalculatorCatalogSelection.ts` stores only catalog filter state;
- `useSpokeCalculatorWheelCatalog.ts` supplies catalog brand/model options;
- `go-backend/web/admin/src/views/SpokeCatalog.vue` and
  `/api/admin/spoke-catalog` own catalog entry and import workflows.

Selecting a rim or hub in this card only filters the recorded-result system. It
must not write to the wizard draft, change the manual calculation payload, or
replace a calculated result with a catalog value.

### 1.3 Wheelset lacing topology reference page / 轮组编法拓扑参考页

Separate from those two calculator systems, `/resources/轮组编法` is an
independent display and topology-reference surface. It is not a calculator
step and it does not own any calculator state.

- The page may define and display pure topology facts such as hole mapping,
  cross-count support, `21H G3 2:1 (14/7)`, and uniform `24H 2:1 (16/8)`.
- The calculator must not call the page, read its UI state, consume its
  geometry-projection telemetry, or receive an automatic result back from it.
- A future change may reuse a separately tested pure type, hole-mapping
  function, or topology-selection rule. That reuse must not share page
  components, DOM state, telemetry, interference labels, or calculation
  results.
- `2:1` is a topology distribution label, not a tension ratio and not a
  substitute for the calculator's measured geometry inputs.

The topology page's SVG display radii are screen coordinates. They are never
valid ERD, PCD, flange-spacing, or spoke-length inputs for this calculator.

## 2. API and data boundaries / API 与数据边界

The Go service remains the authoritative source for CAD geometry and backend
catalog records. The browser uses two separate API purposes:

- `POST /api/v1/spoke/calc`: manual calculation request. The public Nuxt
  calculator sends measured ERD/flange/PCD values and empty catalog IDs.
- `GET /api/v1/spoke/catalog/export` (and `/spoke/export`): browser-facing
  catalog projection used for labels, identifiers, and catalog filtering.
- `GET /api/v1/spoke/catalog/results`: separate browser-facing projection for
  backend-recorded spoke lengths. It returns only preset search metadata and
  the four measured length fields; it does not return geometry, nipple length,
  or internal import notes.
- `GET /api/admin/spoke-catalog`: authenticated full catalog projection for
  backend maintenance, including geometry and recorded build measurements.

### 2.1 Dimension units / 尺寸单位

`ERD` and `PCD` are always diameters in `mm` in the calculator contract. Other
physical lengths (flange spacing, rim offset, drilling offset, and hole
diameter) also use `mm`; angles use `°`. If a formula needs a radius, it must
derive it explicitly as `ERD / 2` or `PCD / 2` and use a radius-named variable.
The topology reference page's display radius constants have no physical unit
and must never be sent in this payload.

The Go calculation service still accepts `rimId`/`hubId` for controlled legacy
or integration callers. That compatibility path does not authorize the Nuxt
calculator to use catalog selection as an automatic geometry source. Any change
to that API contract requires a separate review of the manual/catalog boundary.

### 2.2 Public result projection / 公共结果投影

The current public export deliberately removes CAD geometry and
`actualLengths`; it remains a safe identifier/label projection. The lower
search card reads the separate `/spoke/catalog/results` projection through
`useSpokeCalculatorRecordedResults.ts`. This keeps verified cut lengths
available to the search card without allowing catalog selection to fill or
replace manual calculator geometry. Do not work around this boundary by
sending catalog IDs to `/spoke/calc`, and do not put proprietary geometry into
the Nuxt bundle.

## 3. Data management and sync / 数据管理与同步

Catalog management lives in the Go admin API. The admin workflow may import
verified build lengths and catalog geometry, but those records are not manual
calculator state.

- RIM/HUB geometry stays in the backend database.
- Preset names, keywords, and stable IDs may be projected for browser search.
- Verified cut lengths reach the browser only through the narrow recorded-result
  projection; they never enter the manual calculator catalog state.
- The browser never receives proprietary CAD geometry through the public
  catalog export.

The brand wheelset repair-kit directory is a separate surface from the
calculator catalog described above. Its access boundary and update procedure
are documented in
[`BRAND-WHEELSET-SPOKE-SPECS.md`](./BRAND-WHEELSET-SPOKE-SPECS.md). Do not
reuse the calculator's public export for exact wheelset repair-kit lengths or
nipple data.

## 4. Calculation and tension rules / 计算与张力规则

J-bend and straight-pull geometry are separate backend calculators. Physical
corrections (hole-radius deduction, straight-pull tangent offset, elastic
stretch estimate, alternating drilling offset, and interlacing compensation)
are calculation inputs. The frontend manual flow chooses and submits these
inputs; it does not infer them from a catalog selection.

Interlacing compensation is an explicit optional length input for each wheel.
The backend applies only the submitted value when interlacing is enabled; a
missing value does not trigger a crossing-count-based default. This keeps the
correction under the user's control because spoke section/butting and hub exit
geometry vary, and many current hubs are built without interlacing.

The tension ratio is a derived result of the manual geometry calculation. It is
not the 2:1 or 1:1 spoke-hole topology ratio. See
`docs/design/spoke-tension-ratio-architecture.md` for the sign convention and
ratio verification rules.

## 5. Related file index / 相关文件索引

- **Manual calculator data/contracts**:
  `app/types/spokeCalculator.ts`, `app/utils/spokeCalculatorPayload.ts`
- **Catalog normalization**: `app/utils/spokeCatalogNormalizer.ts`
- **Frontend catalog state**: `app/composables/useSpokeCalculatorCatalog.ts`,
  `app/composables/useSpokeCalculatorRecordedResults.ts`,
  `app/composables/useSpokeCalculatorCatalogSelection.ts`,
  `app/composables/useSpokeCalculatorWheelCatalog.ts`
- **Backend calculation**: `go-backend/internal/service/spoke_service.go`
- **Backend geometry calculators**:
  `go-backend/internal/service/spoke_geometry_j_bend.go`,
  `go-backend/internal/service/spoke_geometry_straight_pull.go`
- **Public API handler**: `go-backend/internal/api/v1/spoke/handler.go`
- **Admin catalog handler**: `go-backend/internal/api/admin/spoke_catalog_handler.go`
