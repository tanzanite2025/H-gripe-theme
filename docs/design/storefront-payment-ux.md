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

## Checkout Total Confirmation

- The shared checkout must refresh the backend quote immediately before creating a local order.
- The create-order request must include that quote's exact `total_minor` as `expected_total_minor`.
- The backend recomputes the quote inside the order-creation transaction and compares minor units exactly.
- Any mismatch returns HTTP `409` with code `order_total_changed`; it must not create the order, deduct stock, spend points, or start provider payment.
- The storefront refreshes the quote and asks the customer to review the updated total before retrying.

## Checkout Cart Consumption

- The API reads the authenticated customer's cart and passes its explicit cart ID into order creation.
- The order-creation transaction locks that cart and its item rows, rebuilds the quote from the locked cart snapshot, reserves stock, creates the order, and deletes exactly those cart rows before commit.
- This applies to card, PayPal, Alipay, WeChat Pay, and other asynchronous payment methods. Browser return pages are never the source of truth for cart consumption.
- A second concurrent checkout against the same cart receives HTTP `409` with code `checkout_cart_already_consumed`; it must not create another order or deduct stock.
- If an unpaid order is cancelled or expires before payment, its order items, points, and coupon usage are restored in the same rollback transaction. A paid order does not restore the cart.
- After payment success, the storefront reloads the backend cart instead of issuing a broad client-side clear, so products added while an asynchronous payment was pending are preserved.

## Visual Assets

- Use existing SVG payment assets from `nuxt-i18n/public/icons/payment/`.
- Brand methods such as PayPal, Alipay, and WeChat Pay should use their brand SVGs.
- Card checkout should show card network SVGs, not a Stripe button.
- Keep accessible labels in code even when the visual asset avoids extra translation.

## Shared Code

The canonical storefront mapping lives in:

- `nuxt-i18n/app/utils/paymentPresentation.ts`

Product detail pages and checkout UI should use this helper instead of duplicating provider-to-method presentation logic.
