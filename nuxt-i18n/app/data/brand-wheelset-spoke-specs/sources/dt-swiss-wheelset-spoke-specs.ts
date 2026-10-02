import type { BrandWheelsetSpokeCatalog, BrandWheelsetSpokeRecord } from '../brand-wheelset-spoke-specs-types.js'
import { jBend, pair, straight, wheelset } from '../brand-wheelset-spoke-specs-builders.js'

const sources = {
  arc1100: 'https://www.dtswiss.com/en/wheels/wheels-road/aero/arc-1100-dicut-db',
  arc1400: 'https://www.dtswiss.com/en/wheels/wheels-road/aero/arc-1400-dicut-db',
  erc1100: 'https://www.dtswiss.com/en/wheels/wheels-road/endurance/erc-1100-dicut',
  erc1400: 'https://www.dtswiss.com/en/wheels/wheels-road/endurance/erc-1400-dicut',
  grc1100: 'https://www.dtswiss.com/en/wheels/wheels-road/gravel/grc-1100-dicut',
  grc1400: 'https://www.dtswiss.com/en/wheels/wheels-road/gravel/grc-1400-dicut',
  prc1400: 'https://www.dtswiss.com/en/wheels/wheels-road/performance/prc-1400-spline',
  xrc1200: 'https://www.dtswiss.com/en/wheels/wheels-mtb/cross-country/xrc-1200-spline',
  exc1200: 'https://www.dtswiss.com/en/wheels/wheels-mtb/enduro/exc-1200-classic',
  productSupport: 'https://www.dtswiss.com/en/support/product-support',
}

const makeArc = (tier: '1100' | '1400', depthMm: 38 | 55 | 65 | 85): BrandWheelsetSpokeRecord => {
  const is1100 = tier === '1100'
  const rim = depthMm === 38
    ? { depthMm, innerWidthMm: 20, outerWidthMm: 26 }
    : depthMm === 55
      ? { depthMm, innerWidthMm: 22, outerWidthMm: 28 }
      : depthMm === 65
        ? { depthMm, innerWidthMm: 22, outerWidthMm: 29 }
        : { depthMm, innerWidthMm: 22, outerWidthMm: 32 }
  const defaultModel = is1100 ? 'DT Aerolite II T-head' : 'DT Aero Comp II T-head'
  const knownFrontLengths: Record<number, [number | null, number | null]> = is1100
    ? { 38: [285, 285], 55: [269, 271], 65: [259, 261], 85: [239, 241] }
    : { 38: [null, null], 55: [269, 271], 65: [259, 261], 85: [239, 241] }
  const knownRearLengths: Record<number, [number | null, number | null]> = is1100
    ? { 38: [null, null], 55: [266, 263], 65: [255, 253], 85: [236, 233] }
    : { 38: [null, null], 55: [266, 263], 65: [256, 253], 85: [236, 233] }
  const frontSpokeCount = depthMm === 38 ? 24 : 20
  const frontSpokes = knownFrontLengths[depthMm]!
  const rearSpokes = knownRearLengths[depthMm]!

  return wheelset({
    slug: `arc-${tier}-dicut-db-${depthMm}`,
    model: `ARC ${tier} DICUT DB ${depthMm}`,
    rim,
    front: {
      spokeCount: frontSpokeCount,
      lacingPattern: '2X',
      left: straight(defaultModel, frontSpokes[0]),
      right: straight(defaultModel, frontSpokes[1]),
    },
    rear: {
      spokeCount: 24,
      lacingPattern: '2X',
      left: straight(is1100 && depthMm === 38 ? 'DT Aerolite II T-head' : defaultModel, rearSpokes[0]),
      right: straight(is1100 && depthMm === 38 ? 'DT Aero Comp II T-head' : defaultModel, rearSpokes[1]),
    },
    nippleModel: 'DT Pro Lock Hidden Aluminum',
    nippleLengthMm: 12,
    sourceUrl: is1100 ? sources.arc1100 : sources.arc1400,
  })
}

const makeErc = (tier: '1100' | '1400', depthMm: 35 | 45): BrandWheelsetSpokeRecord => {
  const is1100 = tier === '1100'
  const model = is1100 ? 'DT Aerolite II T-head' : 'DT Aero Comp T-head'
  const lengths: Record<number, { front: [number, number], rear: [number, number] }> = is1100
    ? { 35: { front: [285, 288], rear: [287, 285] }, 45: { front: [275, 278], rear: [277, 275] } }
    : { 35: { front: [285, 288], rear: [287, 285] }, 45: { front: [275, 278], rear: [277, 275] } }
  const size = lengths[depthMm]!

  return wheelset({
    slug: `erc-${tier}-dicut-${depthMm}`,
    model: `ERC ${tier} DICUT DB ${depthMm}`,
    rim: { depthMm, innerWidthMm: 22, outerWidthMm: 28.5 },
    front: {
      spokeCount: 24,
      lacingPattern: '2X',
      left: straight(model, size.front[0]),
      right: straight(model, size.front[1]),
    },
    rear: is1100
      ? {
          spokeCount: 24,
          lacingPattern: '2X',
          left: straight('DT Aerolite II T-head', size.rear[0]),
          right: straight('DT Aero Comp II T-head', size.rear[1]),
        }
      : {
          spokeCount: 24,
          lacingPattern: '2X',
          left: straight('DT Aero Comp T-head', size.rear[0]),
          right: straight('DT Aero Comp T-head', size.rear[1]),
        },
    nippleModel: 'DT Pro Lock Hidden Aluminum',
    nippleLengthMm: 12,
    sourceUrl: is1100 ? sources.erc1100 : sources.erc1400,
  })
}

const makeGrc = (
  tier: '1100' | '1400',
  depthMm: 30 | 50,
  diameter: '650B' | '700C',
): BrandWheelsetSpokeRecord => {
  const is1100 = tier === '1100'
  const spokeModel = is1100 ? 'DT Aerolite II T-head' : 'DT Aero Comp II T-head'
  const rim = depthMm === 30
    ? { depthMm, innerWidthMm: 24, outerWidthMm: 31 }
    : { depthMm, innerWidthMm: 24, outerWidthMm: 36.5 }
  const pairLengths: Record<string, { front: [number | null, number | null], rear: [number | null, number | null] }> = {
    '1100-30-650B': { front: [271, 271], rear: [273, 271] },
    '1100-30-700C': { front: [null, null], rear: [null, null] },
    '1100-50-700C': { front: [272, 275], rear: [null, 271] },
    '1400-30-650B': { front: [271, 271], rear: [270, 270] },
    '1400-30-700C': { front: [null, null], rear: [null, null] },
    '1400-50-700C': { front: [270, 275], rear: [274, 270] },
  }
  const lengths = pairLengths[`${tier}-${depthMm}-${diameter}`]!
  const rearLeftModel = is1100 ? 'DT Aerolite II T-head' : spokeModel
  const rearRightModel = is1100 ? 'DT Aero Comp II T-head' : spokeModel
  const suffix = depthMm === 30 ? `-${diameter.toLowerCase()}` : ''

  return wheelset({
    slug: `grc-${tier}-dicut-${depthMm}${suffix}`,
    model: `GRC ${tier} DICUT DB ${depthMm} (${diameter})`,
    rim,
    front: {
      spokeCount: 24,
      lacingPattern: '2X',
      left: straight(spokeModel, lengths.front[0]),
      right: straight(spokeModel, lengths.front[1]),
    },
    rear: {
      spokeCount: 24,
      lacingPattern: '2X',
      left: straight(rearLeftModel, lengths.rear[0]),
      right: straight(rearRightModel, lengths.rear[1]),
    },
    nippleModel: 'DT Pro Lock Hidden Aluminum',
    nippleLengthMm: 12,
    sourceUrl: is1100 ? sources.grc1100 : sources.grc1400,
  })
}

const makeExc = (diameter: '27.5' | '29'): BrandWheelsetSpokeRecord => wheelset({
  slug: `exc-1200-classic-${diameter.replace('.', '-')}`,
  model: `EXC 1200 CLASSIC ${diameter}\u2033 30`,
  rim: { depthMm: 22, innerWidthMm: 30, outerWidthMm: 37.2 },
  front: {
    spokeCount: 32,
    lacingPattern: '3X',
    left: jBend('DT Revolite', diameter === '29' ? 296 : null),
    right: jBend('DT Revolite', diameter === '29' ? 297 : null),
  },
  rear: {
    spokeCount: 32,
    lacingPattern: '3X',
    left: jBend('DT Revolite'),
    right: jBend('DT Revolite'),
  },
  nippleModel: 'DT Pro Lock Flat Hexagonal Aluminum',
  nippleLengthMm: 13,
  sourceUrl: sources.exc1200,
})

const makeLegacyArc = (
  tier: '1100' | '1400',
  depthMm: 50 | 62 | 80,
): BrandWheelsetSpokeRecord => {
  const spokeModel = tier === '1100' ? 'DT Aerolite II T-head' : 'DT Aero Comp T-head'
  const lengths: Record<number, {
    front: [number, number]
    rear: [number, number]
  }> = {
    50: { front: [272, 273], rear: [273, 270] },
    62: { front: [258, 261], rear: [260, 256] },
    80: { front: [238, 238], rear: [238, 233] },
  }
  const size = lengths[depthMm]!

  return wheelset({
    slug: `arc-${tier}-dicut-db-${depthMm}-legacy`,
    model: `ARC ${tier} DICUT DB ${depthMm}`,
    lifecycleStatus: 'legacy',
    rim: {
      depthMm,
      innerWidthMm: 20,
      outerWidthMm: depthMm === 50 ? 26.5 : depthMm === 62 ? 27 : 32,
    },
    front: {
      spokeCount: 24,
      lacingPattern: '2X',
      left: straight(spokeModel, size.front[0]),
      right: straight(spokeModel, size.front[1]),
    },
    rear: {
      spokeCount: 24,
      lacingPattern: '2X',
      left: straight(spokeModel, size.rear[0]),
      right: straight(tier === '1100' ? 'DT Aero Comp II T-head' : spokeModel, size.rear[1]),
    },
    nippleModel: 'DT Pro Lock Hidden Aluminum',
    nippleLengthMm: 15,
    sourceUrl: sources.productSupport,
  })
}

const legacyWheelsets: BrandWheelsetSpokeRecord[] = [
  ...([50, 62, 80] as const).map(depth => makeLegacyArc('1100', depth)),
  ...([50, 62] as const).map(depth => makeLegacyArc('1400', depth)),
  wheelset({
    slug: 'prc-1100-mon-chasseral-35-legacy',
    model: 'PRC 1100 DICUT Mon Chasseral 35 (Rim Brake)',
    lifecycleStatus: 'legacy',
    rim: { depthMm: 35, innerWidthMm: 18, outerWidthMm: 25 },
    front: {
      spokeCount: 16,
      lacingPattern: '0X',
      left: straight('DT Aerolite Straightpull', 283),
      right: straight('DT Aerolite Straightpull', 283),
    },
    rear: {
      spokeCount: 21,
      lacingPattern: '2:1',
      left: straight('DT Aerolite T-head', 286),
      right: straight('DT Aero Comp Straightpull', 284),
    },
    nippleModel: 'DT Pro Lock Hidden Aluminum',
    nippleLengthMm: 15,
    sourceUrl: sources.productSupport,
  }),
  wheelset({
    slug: 'prc-1100-mon-chasseral-24-db-legacy',
    model: 'PRC 1100 DICUT 24 DB Mon Chasseral',
    lifecycleStatus: 'legacy',
    rim: { depthMm: 24, innerWidthMm: 18, outerWidthMm: 24 },
    front: {
      spokeCount: 24,
      lacingPattern: '2X',
      left: straight('DT Aerolite II T-head', 294),
      right: straight('DT Aerolite II T-head', 296),
    },
    rear: {
      spokeCount: 24,
      lacingPattern: '2X',
      left: straight('DT Aerolite II T-head', 296),
      right: straight('DT Aero Comp II T-head', 294),
    },
    nippleModel: 'DT Pro Lock Hidden Aluminum',
    nippleLengthMm: 15,
    sourceUrl: sources.productSupport,
  }),
  wheelset({
    slug: 'cr-1400-dicut-db-25-legacy',
    model: 'CR 1400 DICUT DB 25',
    lifecycleStatus: 'legacy',
    rim: { depthMm: 25, innerWidthMm: 24, outerWidthMm: 28 },
    front: {
      spokeCount: 24,
      lacingPattern: '2X',
      left: straight('DT Aerolite Straightpull', 292),
      right: straight('DT Aerolite Straightpull', 294),
    },
    rear: {
      spokeCount: 24,
      lacingPattern: '2X',
      left: straight('DT Aerolite Straightpull', 294),
      right: straight('DT Aero Comp Straightpull', 290),
    },
    nippleModel: 'DT Pro Lock Squorx Pro Head Aluminum',
    nippleLengthMm: 15,
    sourceUrl: sources.productSupport,
  }),
  wheelset({
    slug: 'pr-1600-dicut-21-rim-brake-legacy',
    model: 'PR 1600 DICUT 21（圈刹）',
    lifecycleStatus: 'legacy',
    rim: { depthMm: 21, innerWidthMm: 18, outerWidthMm: 21.5 },
    front: {
      spokeCount: 16,
      lacingPattern: '0X',
      left: straight('DT Aero Comp Straightpull', 282),
      right: straight('DT Aero Comp Straightpull', 282),
    },
    rear: {
      spokeCount: 24,
      lacingPattern: '2X',
      left: straight('DT Aero Comp Straightpull', 288),
      right: straight('DT Aero Comp Straightpull', 292),
    },
    nippleModel: 'DT Pro Lock Aluminum',
    nippleLengthMm: 12,
    sourceUrl: sources.productSupport,
  }),
]

// Current model names and component types come from DT Swiss product pages.
// Legacy repair-kit records are retained for owners of discontinued wheelsets.
// Spoke lengths are filled only when the matching spare-part or kit record lists them.
export const dtSwissWheelsetSpokeCatalog: BrandWheelsetSpokeCatalog = {
  brandSlug: 'dt-swiss',
  brandName: 'DT Swiss',
  publicationStatus: 'published',
  sourceCheckedAt: '2026-10-01',
  wheelsets: [
    ...([38, 55, 65, 85] as const).map(depth => makeArc('1100', depth)),
    ...([38, 55, 65, 85] as const).map(depth => makeArc('1400', depth)),
    ...([35, 45] as const).flatMap(depth => [makeErc('1100', depth), makeErc('1400', depth)]),
    ...([30, 50] as const).flatMap(depth => [
      makeGrc('1100', depth, '700C'),
      makeGrc('1400', depth, '700C'),
    ]),
    makeGrc('1100', 30, '650B'),
    makeGrc('1400', 30, '650B'),
    wheelset({
      slug: 'prc-1400-spline-35',
      model: 'PRC 1400 SPLINE 35',
      rim: { depthMm: 35, innerWidthMm: 18, outerWidthMm: 25 },
      front: pair('DT Aerolite T-head', 20, '0X', 282),
      rear: {
        spokeCount: 24,
        lacingPattern: '2X',
        left: straight('DT Aero Comp Straight-pull', 295),
        right: straight('DT Aero Comp Straight-pull', 290),
      },
      nippleModel: 'DT Pro Lock Hidden Aluminum',
      nippleLengthMm: 12,
      sourceUrl: sources.prc1400,
    }),
    wheelset({
      slug: 'xrc-1200-spline-29-30',
      model: 'XRC 1200 SPLINE 29″ 30',
      rim: { depthMm: 20, innerWidthMm: 30, outerWidthMm: 36 },
      front: {
        spokeCount: 24,
        lacingPattern: '2X',
        left: straight('DT Revolite T-head', 299),
        right: straight('DT Revolite T-head', 300),
      },
      rear: pair('DT Revolite T-head', 24, '2X', 299),
      nippleModel: 'DT ProLock Squorx ProHead Aluminum',
      nippleLengthMm: 15,
      sourceUrl: sources.xrc1200,
    }),
    makeExc('27.5'),
    makeExc('29'),
    ...legacyWheelsets,
  ],
}
