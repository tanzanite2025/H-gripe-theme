import assert from 'node:assert/strict'
import { storefrontFontArtifactViolations } from './font-artifact-contract.ts'

const messagesFor = (css: string): string[] => (
  storefrontFontArtifactViolations(css).map(finding => finding.violation)
)

assert.deepEqual(
  messagesFor(`
@font-face {
  font-family: StorefrontSystem;
  src: url(../fonts/StorefrontSystem-CJK.f8ce6d72e8cb.woff2) format("woff2");
  font-display: swap;
  unicode-range: U+4E00-9FFF;
}
`),
  [],
)

assert.deepEqual(
  messagesFor(`
:host,html{font-family:StorefrontSystem}
@font-face{font-family:StorefrontSystem;src:url(../fonts/StorefrontSystem-CN-Latin.woff2) format("woff2");font-weight:100 900;font-style:normal;font-display:block}
`),
  [
    'storefront font file reference must be content-addressed: StorefrontSystem-CN-Latin.woff2',
    'StorefrontSystem must use StorefrontSystem-CJK.f8ce6d72e8cb.woff2, not StorefrontSystem-CN-Latin.woff2',
    'StorefrontSystem must use font-display: swap',
    'StorefrontSystem must declare unicode-range in built CSS',
  ],
)

assert.deepEqual(
  messagesFor(`
@font-face {
  font-family: StorefrontSystemLatin;
  src: url(../fonts/StorefrontSystem-Latin.00af3fec5b34.woff2) format("woff2");
  font-display: swap;
  unicode-range: U+0000-00FF;
}
`),
  ['StorefrontSystemLatin must not declare unicode-range in built CSS'],
)

assert.deepEqual(
  messagesFor(`
@font-face {
  font-family: StorefrontSystemLegacy;
  src: url(../fonts/StorefrontSystem-Legacy.0123456789ab.woff2) format("woff2");
  font-display: swap;
}
`),
  ['unapproved storefront @font-face family "StorefrontSystemLegacy"'],
)

console.log('Font artifact contract edge-case checks passed.')
