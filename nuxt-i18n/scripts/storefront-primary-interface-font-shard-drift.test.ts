import assert from 'node:assert/strict'
import {
  findPrimaryStorefrontInterfaceFontShardDrift,
  hasUnsupportedStorefrontFontUnicodeVariationSequences,
} from './storefront-font-character-shard-contract.ts'

const originalFontCodePoints = new Set([0x0041, 0x4E00, 0x4E01])
const latinFontCodePoints = new Set([0x0041])
const primaryMessageCodePoints = new Set([0x0041, 0x4E00, 0x4E01, 0x20000])

assert.deepEqual(
  findPrimaryStorefrontInterfaceFontShardDrift(
    primaryMessageCodePoints,
    originalFontCodePoints,
    latinFontCodePoints,
    new Set([0x4E00]),
  ),
  [0x4E01],
  'A new primary-interface character from the original CJK face must require shard regeneration.',
)

assert.deepEqual(
  findPrimaryStorefrontInterfaceFontShardDrift(
    primaryMessageCodePoints,
    originalFontCodePoints,
    latinFontCodePoints,
    new Set([0x4E00, 0x4E01]),
  ),
  [],
  'Latin characters and characters outside the original Maple UI face must not trigger CJK shard regeneration.',
)

assert.equal(
  hasUnsupportedStorefrontFontUnicodeVariationSequences([
    { platformID: 0, encodingID: 4, table: { version: 12 } },
    { platformID: 3, encodingID: 10, table: { version: 12 } },
  ]),
  false,
  'Standard Unicode cmap subtables must remain supported.',
)

assert.equal(
  hasUnsupportedStorefrontFontUnicodeVariationSequences([
    { platformID: 0, encodingID: 5, table: { version: 14 } },
  ]),
  true,
  'Unicode variation sequences must block generation until per-sequence coverage can be verified.',
)

console.log('Primary interface font shard drift regression checks passed.')
