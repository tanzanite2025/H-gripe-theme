import assert from 'node:assert/strict'
import { storefrontFontArtifactViolations } from './font-artifact-contract.ts'
import { storefrontCjkCharacterShardManifest } from '../storefront-font-character-shard-contract.ts'

const messagesFor = (css: string): string[] => (
  storefrontFontArtifactViolations(css).map(finding => finding.violation)
)

assert.deepEqual(
  messagesFor(`
@font-face {
  font-family: MapleUICJK;
  src: url(../fonts/${storefrontCjkCharacterShardManifest.shards[0].filename}) format("woff2");
  font-display: block;
  unicode-range: ${storefrontCjkCharacterShardManifest.shards[0].unicodeRange};
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
    'MapleUICJK must use an approved character-shard inventory file, not MapleUI-CN-Latin.woff2',
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

const firstCharacterShard = storefrontCjkCharacterShardManifest.shards[0]
const minifiedCharacterShardCss = `@font-face{font-family:MapleUICJK;src:url(../fonts/${firstCharacterShard.filename}) format("woff2");font-display:block;unicode-range:${firstCharacterShard.unicodeRange.toLowerCase().replaceAll(' ', '')}}`
assert.deepEqual(messagesFor(minifiedCharacterShardCss), [])
assert.deepEqual(
  messagesFor(minifiedCharacterShardCss.replace(firstCharacterShard.unicodeRange.toLowerCase().replaceAll(' ', ''), 'U+0000-FFFF')),
  [`${firstCharacterShard.filename} changed its approved character coverage in built CSS`],
)

console.log('Font artifact contract edge-case checks passed.')
