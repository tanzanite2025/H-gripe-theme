# Nuxt storefront i18n current status

Last updated: 2026-07-23

This document is the current source of truth for the Nuxt storefront language-file workflow. Older completion reports and archived notes are historical context only. Update this file after each i18n block so future work does not drift.

## Current reality

- The storefront declares 34 locales in `app/i18n/locales.manifest.js`.
- Nuxt reads aggregated locale files from `app/i18n/locales/*.json`.
- The maintainable source files are split by module under `app/i18n/messages/<locale>/*.json`.
- `npm run i18n:build` generates the aggregated `app/i18n/locales/*.json` files from the split `messages` files.
- The i18n scripts resolve `app/i18n` first, falling back to root `i18n` only if needed.
- `npm run i18n:check -- --limit=40` currently reports:
  - `Used keys missing in en: 0`
  - `Duplicate JSON keys: 0`
  - extra locale keys only as cleanup warnings
  - latest checked baseline: `Base keys (en): 898`, `Static translation refs: 619`

## Translation scope rules

Internationalization should be added by bounded product area, not by bulk replacing every English string.

Good candidates:

- stable UI labels, buttons, table headings, empty-state text
- stable accessibility labels and UI alt text
- static guide shell text after the component/page structure is stable
- policy/support page shell copy when the backend source of truth is not expected to own it

Do not blindly translate:

- product names, SKU option names, product dimensions, weights, prices, stock and technical specs
- admin-entered product descriptions, FAQ bodies, policy bodies, or marketing content that should eventually come from backend-localized content
- dynamic category names, collection names, product type names, or search behavior parameters
- brand names, model names, standards, units, and technical abbreviations

If a file mixes static UI and large content blocks, split the component first, then internationalize the stable child component.

## File ownership

Use one module file per bounded area:

```text
app/i18n/messages/<locale>/
├── common.json
├── products.json
├── support.json
├── guidesTireguides.json
├── productInformationTabs.json
├── shopCategoryMenu.json
├── shopPage.json
├── quickBuy.json
├── dockMenu.json
├── checkout.json
├── cookieConsent.json
├── authModal.json
├── chatModal.json
└── ...
```

Naming rule:

- page or feature module: `guidesTireguides`, `productInformationTabs`, `shopCategoryMenu`, `quickBuy`, `dockMenu`
- broad site areas: `products`, `support`, `company`, `member`
- avoid dumping page-specific copy into `common.json`

After changing any `messages` file, run:

```powershell
cd nuxt-i18n
npm run i18n:build
npm run i18n:check -- --limit=40
npm run build
```

## Current known i18n debt

- Static key coverage is clean: no currently referenced static key is missing from `en` or from any configured locale.
- Some non-English locale files still contain historical extra keys. They are cleanup warnings only and should be handled separately from active translation work.
- Dynamic key access such as `t(\`quickBuy.hints.${stepKey}\`)` is not fully discoverable by the current static scanner, so any dynamic i18n family must keep all variants in every locale file.
- Native-language review is still needed for long-tail locales. Short UI translations have been supplied for the latest completed blocks, but product/business-sensitive content should be reviewed before production marketing use.

## Recommended next stages

1. Continue account/profile UI only where the copy is static and not backend-owned.
2. Continue individual guide child components only after the page layout and component boundaries are stable.
3. Review long-tail locale wording after the English/Chinese UI source is stable.
4. Keep dynamic product/catalog/backend-owned content out of static language files until backend localization fields exist.

## Current completed block log

### 2026-07-23 — Tire Guides / Inner Tube

- Component: `app/components/tireguides/InnerTubeGuide.vue`
- Source module: `app/i18n/messages/<locale>/guidesTireguides.json`
- Aggregated into: `app/i18n/locales/*.json`
- Reviewed languages: `en`, `zh_cn`
- Fallback languages: all other configured locales use practical fallback copy until native review
- Intentionally not translated:
  - behavior parameters such as product search keyword presets
  - technical abbreviations such as AV, DV, SV
  - numeric measurements and units such as `40mm`, `32/40mm`, `≤25mm`

### 2026-07-23 — Product information tabs

- Component: `app/components/ProductInformationTabs.vue`
- Source module: `app/i18n/messages/<locale>/productInformationTabs.json`
- Aggregated into: `app/i18n/locales/*.json`
- Coverage: all 34 configured locales
- Intentionally not translated:
  - product description HTML from backend
  - after-sales/packaging/shipping HTML from backend

### 2026-07-23 — Shop page and vertical category menu

- Components:
  - `app/pages/shop/index.vue`
  - `app/components/shop/ShopCategoryVerticalMenu.vue`
- Source modules:
  - `app/i18n/messages/<locale>/shopPage.json`
  - `app/i18n/messages/<locale>/shopCategoryMenu.json`
- Aggregated into: `app/i18n/locales/*.json`
- Coverage: all 34 configured locales
- Intentionally not translated:
  - product names, product thumbnails, prices, product URLs
  - dynamic category names from backend
  - popular search keywords used as search behavior inputs

### 2026-07-23 — Quick Buy modal and bottom Dock menu

- Components:
  - `app/components/QuickBuy.vue`
  - `app/components/GradientDockMenu.vue`
- Source modules:
  - `app/i18n/messages/<locale>/quickBuy.json`
  - `app/i18n/messages/<locale>/dockMenu.json`
- Aggregated into: `app/i18n/locales/*.json`
- Coverage: all 34 configured locales
- Validation:
  - `npm run i18n:check -- --limit=40`
  - `npm run build`
- Intentionally not translated:
  - product titles, thumbnails, prices and API error messages
  - backend-configured Quick Buy step names when present
  - cart totals and formatted prices

### 2026-07-23 — Cart drawer and Wishlist drawer

- Components:
  - `app/components/CartDrawer.vue`
  - `app/components/WishlistDrawer.vue`
- Source modules:
  - `app/i18n/messages/<locale>/cartDrawer.json`
  - `app/i18n/messages/<locale>/wishlistDrawer.json`
- Aggregated into: `app/i18n/locales/*.json`
- Coverage: all 34 configured locales
- Validation:
  - `npm run i18n:check -- --limit=40`
  - `npm run build`
- Intentionally not translated:
  - product titles, thumbnails, prices and dynamic wishlist/cart API errors
  - real SKU values; only the fixed `SKU` label is centralized
  - cart totals and formatted prices

### 2026-07-23 — Checkout modal and checkout stepper

- Components:
  - `app/components/CheckoutModal.vue`
  - `app/components/CheckoutStepper.vue`
- Source module: `app/i18n/messages/<locale>/checkout.json`
- Aggregated into: `app/i18n/locales/*.json`
- Coverage: all 34 configured locales
- Reviewed languages: `en`, `zh_cn`
- Fallback languages: other configured locales use English fallback copy until native UI review
- Validation:
  - `npm run i18n:fill-missing`
  - `npm run i18n:build`
  - `npm run i18n:check -- --limit=80`
  - `npm run build`
- Data-source cleanup:
  - `CheckoutStepper` no longer owns fake payment subtitle pricing such as hardcoded `≈ $11.00`
  - payment option display copy is passed from `CheckoutModal` when a real quote/price breakdown is available
- Intentionally not translated:
  - product titles, thumbnails, prices, order totals and formatted amounts
  - shipping quote labels returned by backend
  - country names from the shipping-country source
  - backend API error payloads
  - payment brand names such as PayPal, Stripe, WorldFirst, Visa and Mastercard

### 2026-07-23 — Cookie consent banner and preferences modal

- Component: `app/components/CookieConsent.vue`
- Source module: `app/i18n/messages/<locale>/cookieConsent.json`
- Aggregated into: `app/i18n/locales/*.json`
- Coverage: all 34 configured locales
- Reviewed languages: `en`, `zh_cn`
- Fallback languages: other configured locales use English fallback copy until native UI review
- Validation:
  - `npm run i18n:fill-missing`
  - `npm run i18n:build`
  - `npm run i18n:check -- --limit=60`
- Intentionally not translated:
  - cookie consent storage key and localStorage payload shape
  - policy route `/policies/cookie`

### 2026-07-23 — Auth modal

- Component: `app/components/AuthModal.vue`
- Source module: `app/i18n/messages/<locale>/authModal.json`
- Aggregated into: `app/i18n/locales/*.json`
- Coverage: all 34 configured locales
- Reviewed languages: `en`, `zh_cn`
- Fallback languages: other configured locales use English fallback copy until native UI review
- Validation:
  - `npm run i18n:fill-missing`
  - `npm run i18n:build`
  - `npm run i18n:check -- --limit=60`
  - `npm run build`
- Scope cleanup:
  - modal-specific UI labels, accessibility labels, validation messages and local fallback errors moved out of the component and into `authModal.json`
  - legacy `auth.json` is left in place because other account/login surfaces may still reference it
- Intentionally not translated:
  - backend/API error payloads when an actual error message is returned
  - Google brand name
  - local form state names and auth composable data shape

### 2026-07-23 — Chat modal shell

- Components:
  - `app/components/WhatsAppChatModal.vue`
  - `app/components/ChatWelcomeAgentSelector.vue`
  - `app/components/whatsapp/ChatWelcomePanel.vue`
  - `app/components/whatsapp/ChatTransferModal.vue`
  - `app/components/whatsapp/AgentChatPanel.vue`
  - `app/components/whatsapp/UserChatBody.vue`
  - `app/composables/chat/useWhatsAppState.ts`
- Source module: `app/i18n/messages/<locale>/chatModal.json`
- Aggregated into: `app/i18n/locales/*.json`
- Coverage: all 34 configured locales
- Reviewed languages: `en`, `zh_cn`
- Fallback languages: other configured locales use English fallback copy until native UI review
- Validation:
  - `npm run i18n:fill-missing`
  - `npm run i18n:build`
  - `npm run i18n:check -- --limit=80`
  - `npm run build`
- Scope cleanup:
  - fixed shell labels, tab labels, button titles, accessibility labels, agent status labels and transfer alerts are centralized
  - agent names, emails, avatars, conversation messages, product/order context and backend-provided error messages remain dynamic
  - chat date formatting now follows the active locale instead of forcing `en-US`
- Intentionally not translated:
  - product names, order data, message bodies and customer/agent directory data
  - backend response error payloads
  - WhatsApp and other brand names

### 2026-07-23 — Feedback thread and product-feedback form

- Components:
  - `app/components/UserFeedbackThread.vue`
  - `app/components/feedback/SuggestionForm.vue`
  - `app/pages/support/product-feedback.vue`
- Source modules:
  - `app/i18n/messages/<locale>/feedback.json`
  - `app/i18n/messages/<locale>/feedbackForm.json`
- Aggregated into: `app/i18n/locales/*.json`
- Coverage: all 34 configured locales have the complete key shape
- Reviewed language: `en`, `zh_cn`
- Fallback languages: other configured locales use English fallback copy until native review
- Scope cleanup:
  - fixed feedback thread labels, empty states, search copy and submit states remain in `feedback.json`
  - product-feedback page/form copy remains in `feedbackForm.json`
  - the `Register / Login` CTA is now `feedbackForm.actions.authCta`
  - category and request-type option labels are reactive to locale changes
  - country labels reuse the local `app/data/countries.ts` fact source instead of duplicating country data in language files
  - `getCountryName` now recognizes both hyphenated and underscored Chinese locale codes such as `zh-CN` and `zh_cn`
- Validation:
  - `npm run i18n:fill-missing`
  - `npm run i18n:build`
  - `npm run i18n:check -- --limit=80`
  - `npm run build`
- Intentionally not translated:
  - user-submitted feedback content, names, emails and uploaded file names
  - backend response messages and API error payloads when the server provides them
  - country codes and other submission payload values
