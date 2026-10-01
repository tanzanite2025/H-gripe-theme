import type {
  BrandWheelsetLifecycleStatus,
  BrandWheelsetSpokeCatalog,
  BrandWheelsetSpokeRecord,
} from '../brand-wheelset-spoke-specs-types'
import { jBend, wheelset } from '../brand-wheelset-spoke-specs-builders'

const sources = {
  firecrest303: 'https://www.sram.com/en/service/models/wh-303-ftld-a1',
  firecrest404: 'https://www.sram.com/en/service/models/wh-404-ftld-b1',
  firecrest808: 'https://www.sram.com/en/service/models/wh-808-ftld-b1',
  nsw353: 'https://www.sram.com/en/service/models/wh-353-ntld-a1',
  nsw454: 'https://www.sram.com/en/service/models/wh-454-ntld-b1',
  nsw858: 'https://www.sram.com/en/service/models/wh-858-ntld-b1',
  xplr303Sw: 'https://www.sram.com/en/service/models/wh-303-xpsw-a1',
  xplr303S: 'https://www.sram.com/en/service/models/wh-303-xps-a1',
  xplr101: 'https://www.sram.com/en/service/models/wh-101-xplr-a1',
  series303S: 'https://www.sram.com/en/service/models/wh-303-stld-a1',
  legacy303: 'https://www.sram.com/en/service',
}

type ZippPositionInput = {
  spokeModel: string
  lengths: [number | null, number | null]
  spokeCount: number
  lacingPattern?: string
}

const makePosition = (input: ZippPositionInput) => ({
  spokeCount: input.spokeCount,
  lacingPattern: input.lacingPattern ?? '2X',
  left: jBend(input.spokeModel, input.lengths[0]),
  right: jBend(input.spokeModel, input.lengths[1]),
})

const makeZippWheelset = (input: {
  slug: string
  model: string
  lifecycleStatus?: BrandWheelsetLifecycleStatus
  depthMm: number
  innerWidthMm: number
  outerWidthMm: number
  front: ZippPositionInput
  rear: ZippPositionInput
  nippleModel: string
  nippleLengthMm: number | null
  sourceUrl: string
}): BrandWheelsetSpokeRecord => wheelset({
  slug: input.slug,
  model: input.model,
  lifecycleStatus: input.lifecycleStatus,
  rim: {
    depthMm: input.depthMm,
    innerWidthMm: input.innerWidthMm,
    outerWidthMm: input.outerWidthMm,
  },
  front: makePosition(input.front),
  rear: makePosition(input.rear),
  nippleModel: input.nippleModel,
  nippleLengthMm: input.nippleLengthMm,
  sourceUrl: input.sourceUrl,
})

const zippWheelsets: BrandWheelsetSpokeRecord[] = [
  makeZippWheelset({
    slug: 'zipp-303-firecrest-b1',
    model: 'ZIPP 303 Firecrest [B1 世代 · 2021–2026]',
    depthMm: 40,
    innerWidthMm: 25,
    outerWidthMm: 30,
    front: {
      spokeModel: 'Sapim CX-Sprint',
      lengths: [270, 272],
      spokeCount: 24,
    },
    rear: {
      spokeModel: 'Sapim CX-Sprint',
      lengths: [270, 266],
      spokeCount: 24,
    },
    nippleModel: 'Sapim Secure Lock external black alloy nipple, 2.0 mm',
    nippleLengthMm: 14,
    sourceUrl: sources.firecrest303,
  }),
  makeZippWheelset({
    slug: 'zipp-404-firecrest-b1',
    model: 'ZIPP 404 Firecrest [B1 世代 · 2021–2026]',
    depthMm: 58,
    innerWidthMm: 23,
    outerWidthMm: 28,
    front: {
      spokeModel: 'Sapim CX-Sprint',
      lengths: [254, 256],
      spokeCount: 24,
    },
    rear: {
      spokeModel: 'Sapim CX-Sprint',
      lengths: [254, 250],
      spokeCount: 24,
    },
    nippleModel: 'Sapim Secure Lock external alloy nipple, 2.0 mm',
    nippleLengthMm: 14,
    sourceUrl: sources.firecrest404,
  }),
  makeZippWheelset({
    slug: 'zipp-808-firecrest-b1',
    model: 'ZIPP 808 Firecrest [B1 世代 · 2022–2026]',
    depthMm: 80,
    innerWidthMm: 23,
    outerWidthMm: 27.5,
    front: {
      spokeModel: 'Sapim CX-Sprint',
      lengths: [230, 226],
      spokeCount: 20,
    },
    rear: {
      spokeModel: 'Sapim CX-Sprint',
      lengths: [230, 224],
      spokeCount: 20,
    },
    nippleModel: 'Sapim Secure Lock external alloy nipple',
    nippleLengthMm: 14,
    sourceUrl: sources.firecrest808,
  }),
  makeZippWheelset({
    slug: 'zipp-353-nsw-a1',
    model: 'ZIPP 353 NSW [A1 世代 · Cognition V2]',
    depthMm: 45,
    innerWidthMm: 25,
    outerWidthMm: 30,
    front: {
      spokeModel: 'Sapim CX-Ray',
      lengths: [264, 266],
      spokeCount: 24,
    },
    rear: {
      spokeModel: 'Sapim CX-Ray',
      lengths: [266, 260],
      spokeCount: 24,
    },
    nippleModel: 'Sapim Secure Lock external alloy nipple',
    nippleLengthMm: 14,
    sourceUrl: sources.nsw353,
  }),
  makeZippWheelset({
    slug: 'zipp-454-nsw-b1-c1',
    model: 'ZIPP 454 NSW [B1/C1 世代 · Cognition V2]',
    depthMm: 58,
    innerWidthMm: 23,
    outerWidthMm: 28,
    front: {
      spokeModel: 'Sapim CX-Ray',
      lengths: [256, 252],
      spokeCount: 24,
    },
    rear: {
      spokeModel: 'Sapim CX-Ray',
      lengths: [256, 252],
      spokeCount: 24,
    },
    nippleModel: 'Sapim Secure Lock external alloy nipple',
    nippleLengthMm: 14,
    sourceUrl: sources.nsw454,
  }),
  makeZippWheelset({
    slug: 'zipp-858-nsw-b1-d1',
    model: 'ZIPP 858 NSW [B1/D1 世代 · Cognition V2]',
    depthMm: 85,
    innerWidthMm: 23,
    outerWidthMm: 27.5,
    front: {
      spokeModel: 'Sapim CX-Ray',
      lengths: [232, 226],
      spokeCount: 20,
    },
    rear: {
      spokeModel: 'Sapim CX-Ray',
      lengths: [230, 232],
      spokeCount: 20,
    },
    nippleModel: 'Sapim Secure Lock external alloy nipple',
    nippleLengthMm: null,
    sourceUrl: sources.nsw858,
  }),
  makeZippWheelset({
    slug: 'zipp-303-xplr-sw-a1',
    model: 'ZIPP 303 XPLR SW [A1 世代 · 2024–2026]',
    depthMm: 54,
    innerWidthMm: 32,
    outerWidthMm: 40,
    front: {
      spokeModel: 'Sapim CX-Ray',
      lengths: [256, 258],
      spokeCount: 24,
    },
    rear: {
      spokeModel: 'Sapim CX-Ray',
      lengths: [258, 260],
      spokeCount: 24,
    },
    nippleModel: 'Sapim Secure Lock external alloy nipple',
    nippleLengthMm: 14,
    sourceUrl: sources.xplr303Sw,
  }),
  makeZippWheelset({
    slug: 'zipp-303-xplr-s-a1',
    model: 'ZIPP 303 XPLR S [A1 世代 · 2024–2026]',
    depthMm: 54,
    innerWidthMm: 32,
    outerWidthMm: 40,
    front: {
      spokeModel: 'Sapim CX-Sprint',
      lengths: [258, 260],
      spokeCount: 24,
    },
    rear: {
      spokeModel: 'Sapim CX-Sprint',
      lengths: [260, 258],
      spokeCount: 24,
    },
    nippleModel: 'Sapim Secure Lock external brass/alloy nipple',
    nippleLengthMm: null,
    sourceUrl: sources.xplr303S,
  }),
  makeZippWheelset({
    slug: 'zipp-101-xplr-a1-700c',
    model: 'ZIPP 101 XPLR 700c [A1 世代 · MOTO]',
    depthMm: 15,
    innerWidthMm: 27,
    outerWidthMm: 35,
    front: {
      spokeModel: 'Sapim CX-Sprint',
      lengths: [304, 302],
      spokeCount: 28,
      lacingPattern: '3X',
    },
    rear: {
      spokeModel: 'Sapim CX-Sprint',
      lengths: [304, 302],
      spokeCount: 28,
      lacingPattern: '3X',
    },
    nippleModel: 'Sapim Secure Lock external alloy nipple',
    nippleLengthMm: null,
    sourceUrl: sources.xplr101,
  }),
  makeZippWheelset({
    slug: 'zipp-303-s-a1',
    model: 'ZIPP 303 S [A1 世代]',
    depthMm: 45,
    innerWidthMm: 23,
    outerWidthMm: 27,
    front: {
      spokeModel: 'Sapim CX54 / CX-Sprint',
      lengths: [266, 268],
      spokeCount: 24,
    },
    rear: {
      spokeModel: 'Sapim CX54 / CX-Sprint',
      lengths: [266, 264],
      spokeCount: 24,
    },
    nippleModel: 'Sapim external brass/alloy nipple',
    nippleLengthMm: null,
    sourceUrl: sources.series303S,
  }),
  makeZippWheelset({
    slug: 'zipp-303-firecrest-a1-legacy-77-177',
    model: 'ZIPP 303 Firecrest Carbon Clincher Disc [A1 世代 · MY16–MY19 · 77/177D]',
    lifecycleStatus: 'legacy',
    depthMm: 45,
    innerWidthMm: 19,
    outerWidthMm: 29.9,
    front: {
      spokeModel: 'Sapim CX-Ray',
      lengths: [272, 274],
      spokeCount: 24,
    },
    rear: {
      spokeModel: 'Sapim CX-Ray',
      lengths: [274, 272],
      spokeCount: 24,
    },
    nippleModel: 'Sapim Secure Lock external nipple, 2.0 mm',
    nippleLengthMm: 14,
    sourceUrl: sources.legacy303,
  }),
]

// ZIPP has several overlapping model generations and hub revisions. Keep the
// generation code in every public model title so a repair-kit lookup cannot
// collapse an A1, B1, C1, or D1 record into an ambiguous family name. The
// supplied chart's primary lengths are recorded here; alternate batch notes
// stay out of the customer-facing repair fields until they become separate
// compatibility records.
export const zippSramWheelsetSpokeCatalog: BrandWheelsetSpokeCatalog = {
  brandSlug: 'zipp',
  brandName: 'ZIPP',
  publicationStatus: 'published',
  sourceCheckedAt: '2026-10-01',
  wheelsets: zippWheelsets,
}
