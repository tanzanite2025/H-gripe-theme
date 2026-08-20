import assert from 'node:assert/strict'
import { fontPolicyLineViolations, fontPolicySourceViolations } from './font-policy-utils.ts'

assert.deepEqual(
  fontPolicyLineViolations("font-family: 'MapleUILatin', 'MapleUICJK';"),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations('font-family: MapleUILatin, MapleUICJK;'),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations('font-family: var(--tz-font-ui);'),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations('font-family: var(--storefront-font-system);'),
  ['unapproved storefront font variable "--storefront-font-system"'],
)
assert.deepEqual(
  fontPolicyLineViolations('font-family: var(--tz-font-ui, Inter);'),
  ['unapproved Maple UI font family "Inter"'],
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
  ['unapproved Maple UI font family "Inter"'],
)
assert.deepEqual(
  fontPolicyLineViolations('font-family: MapleUILatin, Inter;'),
  ['unapproved Maple UI font family "Inter"'],
)
assert.deepEqual(
  fontPolicyLineViolations("font-family: 'MapleUILatin', Arial, sans-serif;"),
  ['forbidden system or generic font family'],
)
assert.deepEqual(
  fontPolicyLineViolations("font: 700 1rem/1.2 'Inter', sans-serif;"),
  [
    'forbidden system or generic font family',
    'unapproved Maple UI font family "Inter"',
  ],
)
assert.deepEqual(
  fontPolicyLineViolations('font: inherit;'),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations('font: menu;'),
  ['forbidden system or generic font family'],
)
assert.deepEqual(
  fontPolicyLineViolations("[font-family:MapleUILatin,Inter]"),
  ['unapproved Maple UI font family "Inter"'],
)
assert.deepEqual(
  fontPolicyLineViolations("@font-face { src: local('Inter'), url('/fonts/MapleUI-Latin.woff2') format('woff2'); }"),
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
    MapleUILatin,
    Inter,
    sans-serif;
}
`).map(finding => finding.violation),
  [
    'forbidden system or generic font family',
    'unapproved Maple UI font family "Inter"',
  ],
)
assert.deepEqual(
  fontPolicySourceViolations(`
:root {
  --tz-font-ui: MapleUILatin, Inter;
}
`).map(finding => finding.violation),
  ['unapproved Maple UI font family "Inter"'],
)
assert.deepEqual(
  fontPolicySourceViolations(`
<div class="font-serif">Address</div>
<text font-family="Georgia">Address</text>
`).map(finding => finding.violation),
  [
    'forbidden Tailwind font family utility "font-serif"',
    'forbidden system or generic font family',
  ],
)
assert.deepEqual(
  fontPolicySourceViolations(`
export const storefrontFontFamily = 'MapleUILatin, Inter'
export const storefrontArabicFontFamily = \`MapleUICoverageNotoSansArabic, \${storefrontFontFamily}\`
`).map(finding => finding.violation),
  ['unapproved Maple UI font family "Inter"'],
)
assert.deepEqual(
  fontPolicySourceViolations(`
export const storefrontFontFamily = 'MapleUILatin, MapleUICJK'
export const storefrontThaiFontFamily = \`MapleUICoverageNotoSansThai, \${storefrontFontFamily}\`
`).map(finding => finding.violation),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations("[font-family:'AerialFaster',sans-serif]"),
  [
    'obsolete or secondary font authority reference',
    'forbidden system or generic font family',
    'unapproved Maple UI font family "AerialFaster"',
  ],
)
assert.deepEqual(
  fontPolicyLineViolations("@nuxt/fonts"),
  ['obsolete or secondary font authority reference'],
)
assert.deepEqual(
  fontPolicyLineViolations("src: url('/fonts/StorefrontSystem-Latin.00af3fec5b34.woff2')"),
  ['obsolete or secondary font authority reference'],
)
assert.deepEqual(
  fontPolicyLineViolations('<text font-family="MapleUILatin">SSL</text>'),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations('<text font-family="Georgia">SSL</text>'),
  ['forbidden system or generic font family'],
)
assert.deepEqual(
  fontPolicyLineViolations('<div class="text-sm font-serif">Address</div>'),
  ['forbidden Tailwind font family utility "font-serif"'],
)
assert.deepEqual(
  fontPolicyLineViolations('<div class="md:font-serif">Address</div>'),
  ['forbidden Tailwind font family utility "md:font-serif"'],
)
assert.deepEqual(
  fontPolicyLineViolations('<div class="font-[family-name:var(--tz-font-ui)]">Address</div>'),
  [],
)
assert.deepEqual(
  fontPolicyLineViolations('<div class="font-[family-name:\'Inter\']">Address</div>'),
  ['unapproved Maple UI font family "Inter"'],
)
assert.deepEqual(
  fontPolicyLineViolations('<div class="font-[Inter]">Address</div>'),
  ['unapproved Maple UI font family "Inter"'],
)

console.log('Font policy edge-case contract passed.')
