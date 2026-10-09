import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { allCountries } from 'country-region-data'
import { COUNTRIES } from '../app/data/countries'
import {
  getCheckoutStateProvinceOptions,
  normalizeCheckoutStateProvinceCode,
} from '../app/data/checkoutStateProvinceCatalog'

const unitedStatesOptions = getCheckoutStateProvinceOptions('US')
assert.ok(unitedStatesOptions.some(option => option.code === 'CA' && option.name === 'California'))
assert.equal(normalizeCheckoutStateProvinceCode('US', 'CA'), 'CA')
assert.equal(normalizeCheckoutStateProvinceCode('US', 'California'), 'CA')
assert.equal(normalizeCheckoutStateProvinceCode('US', ' california '), 'CA')
assert.equal(normalizeCheckoutStateProvinceCode('CA', 'Ontario'), 'ON')
assert.equal(normalizeCheckoutStateProvinceCode('US', 'Unknown Region'), '')
assert.equal(normalizeCheckoutStateProvinceCode('ZZ', 'Unknown Region'), '')
assert.equal(normalizeCheckoutStateProvinceCode('US', ''), '')

for (const country of COUNTRIES) {
  const sourceCountry = allCountries.find(([, countryCode]) => countryCode === country.code)
  const expectedOptions = (sourceCountry?.[2] || [])
    .filter(([, regionCode]) => regionCode.trim() !== '')
    .map(([name, code]) => ({ code, name }))
  assert.deepEqual(
    getCheckoutStateProvinceOptions(country.code),
    expectedOptions,
    `${country.code} must expose the same regions as the source country catalog`,
  )
}

const storefrontRoot = process.cwd()
const readStorefrontFile = (filePath: string): string => (
  fs.readFileSync(path.resolve(storefrontRoot, filePath), 'utf8')
)

const checkoutModalSource = readStorefrontFile('app/components/CheckoutModal.vue')
assert.match(checkoutModalSource, /getCheckoutStateProvinceOptions\(form\.value\.country\)/)
assert.match(checkoutModalSource, /state: normalizeCheckoutStateProvinceCode\(addressForm\.country, addressForm\.state\)/)
assert.match(checkoutModalSource, /form\.value\.state, form\.value\.zip/)
assert.doesNotMatch(checkoutModalSource, /StateProvinceLegacyValue/)
assert.match(checkoutModalSource, /v-else[\s\S]*?disabled[\s\S]*?placeholder/)

const savedAddressSource = readStorefrontFile('app/components/account/AccountAddressesTab.vue')
assert.match(savedAddressSource, /normalizeCheckoutStateProvinceCode\(countryCode, source\?\.state \|\| ''\)/)
assert.doesNotMatch(savedAddressSource, /savedStateProvinceValueOutsideCatalog/)
assert.doesNotMatch(savedAddressSource, /savedCountryValueOutsideCatalog/)
assert.match(savedAddressSource, /v-else[\s\S]*?disabled[\s\S]*?placeholder/)

const expressCheckoutSource = readStorefrontFile('app/composables/useStripeExpressCheckoutOrder.ts')
assert.match(expressCheckoutSource, /if \(originalValue && !normalizedValue\)/)
assert.match(expressCheckoutSource, /normalizeStripeExpressCheckoutStateProvince\(countryCode, addressDetails\.address\.state\)/)
assert.match(expressCheckoutSource, /normalizeStripeExpressCheckoutStateProvince\(countryCode, shippingEvent\.address\.state\)/)

console.log('Checkout state and province codes resolve consistently.')
