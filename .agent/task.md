# Safe Style Refactoring Task List

- [x] **Phase 1: Preparation (Safe Mode)**
  - [x] Create `app/assets/css/components/nav.css` with new component classes
  - [x] Register new CSS file in `nuxt.config.ts`
  - [x] Verify build references the new file

- [x] **Phase 2: Incremental Adoption (One by One)**
  - [x] **ProductsTopNav.vue**: Add new classes, keep old `<style>` temporarily
  - [x] **SupportTopNav.vue**: Add new classes, keep old `<style>` temporarily
  - [x] **MembershipAndPointsTabs.vue**: Add new classes, keep old `<style>` temporarily
  - [x] **PoliciesTabs.vue**: Add new classes, keep old `<style>` temporarily

- [x] **Phase 3: Verification**
  - [x] Verify visual consistency for each component
  - [x] Check for style conflicts (specificity issues)

- [x] **Phase 4: Cleanup**
  - [x] Remove redundant styles from `ProductsTopNav.vue`
  - [x] Remove redundant styles from `SupportTopNav.vue`
  - [x] Remove redundant styles from `MembershipAndPointsTabs.vue`
  - [x] Remove redundant styles from `PoliciesTabs.vue`

- [x] **Phase 5: Investigation & Unification**
  - [x] Locate the component rendering "TIRE SIZE" tabs in `/guides/tireguides`
  - [x] Analyze the gradient active state style
  - [x] **Refactor `sizecharts.vue` to use global `nav.css` (Option A)**
  - [x] Remove redundant local styles from `sizecharts.vue`

- [x] **Phase 6: Technical Guide Unification**
  - [x] **Refactor `technical.vue` to use global `nav.css`**
  - [x] Remove redundant local styles from `technical.vue`

- [x] **Phase 7: Final Sweep (Additional Pages)**
  - [x] **Refactor `wheelset-buyers.vue`** (`/guides/wheelset-buyers`)
  - [x] **Refactor `spoke-calculator.vue`** (`/spoke-calculator`)
  - [x] **Refactor `test-report.vue`** (`/support/test-report`)

- [x] **Phase 8: Company About Page (User Request)**
  - [x] **Refactor `about.vue`** (`/company/about`) to use `nav.css`
  - [x] Remove redundant local styles from `about.vue`
