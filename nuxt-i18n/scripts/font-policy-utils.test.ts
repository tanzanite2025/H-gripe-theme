import assert from 'node:assert/strict'
import { fontPolicyLineViolations, fontPolicySourceViolations } from './font-policy-utils.ts'

assert.deepEqual(
  fontPolicyLineViolations("font-family: 'StorefrontSystemLatin', 'StorefrontSystem';"),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations('font-family: StorefrontSystemLatin, StorefrontSystem;'),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations('font-family: var(--tz-font-system);'),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations('font-family: var(--storefront-font-system);'),
  ['unapproved storefront font variable "--storefront-font-system"'],
)
assert.deepEqual(
  fontPolicyLineViolations('font-family: var(--tz-font-system, Inter);'),
  ['unapproved storefront font family "Inter"'],
)
assert.deepEqual(
  fontPolicyLineViolations('fontFamily: family'),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations("fontFamily: 'latin-accent'"),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations("font-family: 'Inter';"),
  ['unapproved storefront font family "Inter"'],
)
assert.deepEqual(
  fontPolicyLineViolations('font-family: StorefrontSystemLatin, Inter;'),
  ['unapproved storefront font family "Inter"'],
)
assert.deepEqual(
  fontPolicyLineViolations("font-family: 'StorefrontSystemLatin', Arial, sans-serif;"),
  ['system or generic font fallback'],
)
assert.deepEqual(
  fontPolicyLineViolations("font: 700 1rem/1.2 'Inter', sans-serif;"),
  [
    'system or generic font fallback',
    'unapproved storefront font family "Inter"',
  ],
)
assert.deepEqual(
  fontPolicyLineViolations('font: inherit;'),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations('font: menu;'),
  ['system or generic font fallback'],
)
assert.deepEqual(
  fontPolicyLineViolations("[font-family:StorefrontSystemLatin,Inter]"),
  ['unapproved storefront font family "Inter"'],
)
assert.deepEqual(
  fontPolicyLineViolations("@font-face { src: local('Inter'), url('/fonts/StorefrontSystem-Latin.woff2') format('woff2'); }"),
  ['local font source'],
)
assert.deepEqual(
  fontPolicyLineViolations("@import url('https://fonts.googleapis.com/css2?family=Inter');"),
  ['external font source'],
)
assert.deepEqual(
  fontPolicyLineViolations("@font-face { src: url('https://cdn.example.test/fonts/inter.woff2') format('woff2'); }"),
  ['external font source'],
)
assert.deepEqual(
  fontPolicySourceViolations(`
.sample {
  font-family:
    StorefrontSystemLatin,
    Inter,
    sans-serif;
}
`).map(finding => finding.violation),
  [
    'system or generic font fallback',
    'unapproved storefront font family "Inter"',
  ],
)
assert.deepEqual(
  fontPolicySourceViolations(`
:root {
  --tz-font-system: StorefrontSystemLatin, Inter;
}
`).map(finding => finding.violation),
  ['unapproved storefront font family "Inter"'],
)
assert.deepEqual(
  fontPolicySourceViolations(`
export const storefrontFontFamily = 'StorefrontSystemLatin, Inter'
export const storefrontArabicFontFamily = \`StorefrontSystemArabic, \${storefrontFontFamily}\`
`).map(finding => finding.violation),
  ['unapproved storefront font family "Inter"'],
)
assert.deepEqual(
  fontPolicySourceViolations(`
export const storefrontFontFamily = 'StorefrontSystemLatin, StorefrontSystem'
export const storefrontThaiFontFamily = \`StorefrontSystemThai, \${storefrontFontFamily}\`
`).map(finding => finding.violation),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations("[font-family:'AerialFaster',sans-serif]"),
  [
    'obsolete AerialFaster reference',
    'system or generic font fallback',
    'unapproved storefront font family "AerialFaster"',
  ],
)

console.log('Font policy edge-case contract passed.')
