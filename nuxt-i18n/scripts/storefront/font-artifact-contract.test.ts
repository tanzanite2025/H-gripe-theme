import assert from 'node:assert/strict'
import { storefrontFontArtifactViolations } from './font-artifact-contract.ts'

const messagesFor = (css: string): string[] => (
  storefrontFontArtifactViolations(css).map(finding => finding.violation)
)

assert.deepEqual(
  messagesFor(`
@font-face {
  font-family: MapleUICJK;
  src: url(../fonts/MapleUI-CJK.f8ce6d72e8cb.woff2) format("woff2");
  font-display: block;
  unicode-range: U+4E00-9FFF;
}
`),
  [],
)

assert.deepEqual(
  messagesFor(`
:host,html{font-family:MapleUICJK}
@font-face{font-family:MapleUICJK;src:url(../fonts/MapleUI-CN-Latin.woff2) format("woff2");font-weight:100 900;font-style:normal;font-display:block}
`),
  [
    'Maple UI font file reference must be content-addressed: MapleUI-CN-Latin.woff2',
    'MapleUICJK must use MapleUI-CJK.f8ce6d72e8cb.woff2, not MapleUI-CN-Latin.woff2',
    'MapleUICJK must declare unicode-range in built CSS',
  ],
)

assert.deepEqual(
  messagesFor(`
@font-face {
  font-family: MapleUILatin;
  src: url(../fonts/MapleUI-Latin.00af3fec5b34.woff2) format("woff2");
  font-display: block;
  unicode-range: U+0000-00FF;
}
`),
  ['MapleUILatin must not declare unicode-range in built CSS'],
)

assert.deepEqual(
  messagesFor(`
@font-face {
  font-family: MapleUILegacy;
  src: url(../fonts/MapleUI-Legacy.0123456789ab.woff2) format("woff2");
  font-display: block;
}
`),
  ['unapproved Maple UI @font-face family "MapleUILegacy"'],
)

console.log('Font artifact contract edge-case checks passed.')
