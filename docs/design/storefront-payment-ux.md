# Storefront Payment UX

The storefront must separate backend payment providers from customer-facing payment methods.

## Product Detail Flow

- Product detail pages may show a compact payment method selector.
- Product detail pages must show a quantity selector; the selected quantity applies to both `Add to cart` and `Buy now`.
- The only purchase actions beside product options are `Add to cart` and `Buy now`.
- `Buy now` opens the shared checkout modal with the selected payment method preselected.
- Product detail pages must not render one direct action button per provider, such as separate `Stripe`, `PayPal`, `Alipay`, and `WeChat Pay` buy buttons.

## Customer-Facing Method Names

| Backend provider | Storefront method | Customer label |
| --- | --- | --- |
| `stripe` | `card` | `Credit / Debit cards` |
| `paypal` | `paypal` | `PayPal` |
| `alipay` | `alipay` | `Alipay` |
| `wechat` | `wechat` | `WeChat Pay` |

Stripe is infrastructure for card checkout. Customers should not be asked to "pay with Stripe" as if it were a wallet.

## Descriptions

- Card: `Secure card checkout powered by Stripe.`
- PayPal: `Pay with a PayPal account or supported wallet.`
- Alipay: `Pay through Alipay.`
- WeChat Pay: `Scan a WeChat Pay QR code to complete payment.`

## Availability Language

Frontend customer UI must not show backend configuration states such as `Not configured` or `Configuration error`.

Use customer-safe language instead:

- `Temporarily unavailable`
- `This payment method is temporarily unavailable.`

Admin tools may still show exact gateway configuration states.

## Visual Assets

- Use existing SVG payment assets from `nuxt-i18n/public/icons/payment/`.
- Brand methods such as PayPal, Alipay, and WeChat Pay should use their brand SVGs.
- Card checkout should show card network SVGs, not a Stripe button.
- Keep accessible labels in code even when the visual asset avoids extra translation.

## Shared Code

The canonical storefront mapping lives in:

- `nuxt-i18n/app/utils/paymentPresentation.ts`

Product detail pages and checkout UI should use this helper instead of duplicating provider-to-method presentation logic.
