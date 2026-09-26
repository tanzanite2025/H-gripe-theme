import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const appRoot = process.cwd()
const repoRoot = path.resolve(appRoot, '..')
const read = (root: string, relativePath: string) => fs.readFileSync(path.resolve(root, relativePath), 'utf8')

const cart = read(appRoot, 'app/composables/useCart.ts')
assert.match(cart, /useStorefrontContext\(\)/)
assert.match(cart, /const \{ countryCode, baseCurrency, displayCurrency \}/)
assert.doesNotMatch(cart, /displayCurrencyCookie/)

const quote = read(appRoot, 'app/composables/useShippingQuote.ts')
for (const field of ['tax_minor', 'member_discount_minor', 'coupon_discount_minor', 'discount_minor', 'total_minor']) {
  assert.match(quote, new RegExp(field))
}

const paymentCompletion = read(appRoot, 'app/composables/usePaymentReturnCompletion.ts')
assert.match(paymentCompletion, /clearCart\(\)/)
assert.match(paymentCompletion, /reloadCartFromBackend\(\)/)
assert.match(paymentCompletion, /checkout\/success/)

const orderContract = read(repoRoot, 'go-backend/internal/api/v1/order/order_contract.go')
assert.match(orderContract, /CouponCode\s+string\s+`json:"coupon_code"`/)
assert.match(orderContract, /Notes\s+string\s+`json:"notes"`/)

const tirePressureCalculator = read(appRoot, 'app/components/tireguides/TirePressureCalculator.vue')
assert.match(tirePressureCalculator, /@submit\.prevent="solve"/)
assert.match(tirePressureCalculator, /JSON\.stringify\(\{/)
assert.match(tirePressureCalculator, /\/engineering\/tire-pressure\/solve/)
assert.match(tirePressureCalculator, /effective_limit_bar/)
assert.match(tirePressureCalculator, /noRecommendation/)

const tirePressureRoutes = read(repoRoot, 'go-backend/internal/api/v1/tirepressure/handler.go')
assert.match(tirePressureRoutes, /group\.POST\("\/solve", h\.Solve\)/)
assert.match(tirePressureRoutes, /MODEL_NOT_CALIBRATED/)
assert.match(tirePressureRoutes, /pressure_recommendation/)

console.log('Commerce architecture contract checks passed.')
