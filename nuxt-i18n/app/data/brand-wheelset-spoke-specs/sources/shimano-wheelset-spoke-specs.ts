import type { BrandWheelsetSpokeCatalog, BrandWheelsetSpokeRecord } from '../brand-wheelset-spoke-specs-types'
import { jBend, straight, wheelset, type SideInput } from '../brand-wheelset-spoke-specs-builders'

const sources = {
  r9270C50: 'https://si.shimano.com/en/pdfs/ev/WH-R9270-C50-TL-R-4831C/EV-WH-R9270-C50-TL-R-4831C.pdf',
  r9270C36: 'https://si.shimano.com/en/pdfs/ev/WH-R9270-C36-TL-R-4829C/EV-WH-R9270-C36-TL-R-4829C.pdf',
  r9270C60Hr: 'https://si.shimano.com/en/pdfs/ev/WH-R9270-C60-HR-TL-R-4833C/EV-WH-R9270-C60-HR-TL-R-4833C.pdf',
  r8170C50: 'https://si.shimano.com/en/pdfs/ev/WH-R8170-C50-TL-R-4870B/EV-WH-R8170-C50-TL-R-4870B.pdf',
  r8170C36: 'https://si.shimano.com/en/pdfs/ev/WH-R8170-C36-TL-R-4868B/EV-WH-R8170-C36-TL-R-4868B.pdf',
  r8170C60: 'https://si.shimano.com/en/pdfs/ev/WH-R8170-C60-TL-R-4872B/EV-WH-R8170-C60-TL-R-4872B.pdf',
  rs710C46: 'https://si.shimano.com/en/pdfs/ev/WH-RS710-C46-TL-R-4991A/EV-WH-RS710-C46-TL-R-4991A.pdf',
  rs710C32: 'https://si.shimano.com/en/pdfs/ev/WH-RS710-C32-TL-R-4989A/EV-WH-RS710-C32-TL-R-4989A.pdf',
  rx880: 'https://si.shimano.com/en/pdfs/ev/WH-RX880-TL-R12-5118A/EV-WH-RX880-TL-R12-5118A.pdf',
  m8100: 'https://si.shimano.com/en/pdfs/ev/WH-M8100-TL-R12-B-4555B/EV-WH-M8100-TL-R12-B-4555B.pdf',
}

const nipple = 'Shimano 14G aluminum nipple'
const xtNipple = 'Shimano 14G aluminum nipple with spherical washer'

type HeadFactory = (spokeModel: string, lengthMm: number | null) => SideInput

const makeRoadWheelset = (input: {
  slug: string
  model: string
  rim: BrandWheelsetSpokeRecord['rim']
  spokeModel: string
  head: HeadFactory
  front: [number, number]
  rear: [number, number]
  rearLacingPattern?: string
  nippleModel?: string
  sourceUrl: string
}): BrandWheelsetSpokeRecord => wheelset({
  slug: input.slug,
  model: input.model,
  rim: input.rim,
  front: {
    spokeCount: 24,
    lacingPattern: '2X',
    left: input.head(input.spokeModel, input.front[0]),
    right: input.head(input.spokeModel, input.front[1]),
  },
  rear: {
    spokeCount: 24,
    lacingPattern: input.rearLacingPattern ?? '2X',
    left: input.head(input.spokeModel, input.rear[0]),
    right: input.head(input.spokeModel, input.rear[1]),
  },
  nippleModel: input.nippleModel ?? nipple,
  nippleLengthMm: null,
  verificationStatus: 'verified',
  sourceUrl: input.sourceUrl,
})

const shimanoWheelsets: BrandWheelsetSpokeRecord[] = [
  makeRoadWheelset({
    slug: 'wh-r9270-c50-tl',
    model: 'DURA-ACE WH-R9270-C50-TL',
    rim: { depthMm: 50, innerWidthMm: 21, outerWidthMm: 28 },
    spokeModel: 'Shimano DURA-ACE bladed 2.0-1.5-2.0 mm',
    head: straight,
    front: [267, 269],
    rear: [251.5, 267.5],
    rearLacingPattern: '2X / 2:1',
    sourceUrl: sources.r9270C50,
  }),
  makeRoadWheelset({
    slug: 'wh-r9270-c36-tl',
    model: 'DURA-ACE WH-R9270-C36-TL',
    rim: { depthMm: 36, innerWidthMm: 21, outerWidthMm: 28 },
    spokeModel: 'Shimano DURA-ACE bladed 2.0-1.5-2.0 mm',
    head: straight,
    front: [280.5, 282],
    rear: [264, 279],
    rearLacingPattern: '2X / 2:1',
    sourceUrl: sources.r9270C36,
  }),
  makeRoadWheelset({
    slug: 'wh-r9270-c60-hr-tl',
    model: 'DURA-ACE WH-R9270-C60-HR-TL',
    rim: { depthMm: 60, innerWidthMm: 21, outerWidthMm: 28 },
    spokeModel: 'Shimano DURA-ACE HR bladed 2.0-1.8-2.0 mm',
    head: straight,
    front: [248, 248],
    rear: [241, 257.5],
    rearLacingPattern: '2X / 2:1',
    sourceUrl: sources.r9270C60Hr,
  }),
  makeRoadWheelset({
    slug: 'wh-r8170-c50-tl',
    model: 'ULTEGRA WH-R8170-C50-TL',
    rim: { depthMm: 50, innerWidthMm: 21, outerWidthMm: 28 },
    spokeModel: 'Shimano ULTEGRA bladed 2.0-1.6-2.0 mm',
    head: straight,
    front: [272, 272],
    rear: [253, 272],
    rearLacingPattern: '2X / 2:1',
    sourceUrl: sources.r8170C50,
  }),
  makeRoadWheelset({
    slug: 'wh-r8170-c36-tl',
    model: 'ULTEGRA WH-R8170-C36-TL',
    rim: { depthMm: 36, innerWidthMm: 21, outerWidthMm: 28 },
    spokeModel: 'Shimano ULTEGRA bladed 2.0-1.6-2.0 mm',
    head: straight,
    front: [286, 286],
    rear: [265, 284],
    rearLacingPattern: '2X / 2:1',
    sourceUrl: sources.r8170C36,
  }),
  makeRoadWheelset({
    slug: 'wh-r8170-c60-tl',
    model: 'ULTEGRA WH-R8170-C60-TL',
    rim: { depthMm: 60, innerWidthMm: 21, outerWidthMm: 28 },
    spokeModel: 'Shimano ULTEGRA bladed 2.0-1.6-2.0 mm',
    head: straight,
    front: [262, 262],
    rear: [244, 262],
    rearLacingPattern: '2X / 2:1',
    sourceUrl: sources.r8170C60,
  }),
  makeRoadWheelset({
    slug: 'wh-rs710-c46-tl',
    model: '105 WH-RS710-C46-TL',
    rim: { depthMm: 46, innerWidthMm: 21, outerWidthMm: 28 },
    spokeModel: 'Shimano 105 bladed 2.0-1.6-2.0 mm',
    head: jBend,
    front: [265.5, 267.5],
    rear: [267.5, 265.5],
    sourceUrl: sources.rs710C46,
  }),
  makeRoadWheelset({
    slug: 'wh-rs710-c32-tl',
    model: '105 WH-RS710-C32-TL',
    rim: { depthMm: 32, innerWidthMm: 21, outerWidthMm: 28 },
    spokeModel: 'Shimano 105 bladed 2.0-1.6-2.0 mm',
    head: jBend,
    front: [279, 281.5],
    rear: [279, 281.5],
    sourceUrl: sources.rs710C32,
  }),
  makeRoadWheelset({
    slug: 'wh-rx880-tl',
    model: 'GRX WH-RX880-TL',
    rim: { depthMm: 32, innerWidthMm: 25, outerWidthMm: 30.7 },
    spokeModel: 'Shimano GRX bladed 2.0-1.6-2.0 mm',
    head: jBend,
    front: [281.5, 283],
    rear: [280, 279],
    sourceUrl: sources.rx880,
  }),
  wheelset({
    slug: 'wh-m8100-tl-29',
    model: 'DEORE XT WH-M8100-TL-29 (Boost)',
    rim: { depthMm: 18.8, innerWidthMm: 24, outerWidthMm: 27.9 },
    front: {
      spokeCount: 28,
      lacingPattern: '3X',
      left: straight('Shimano XT butted 2.0-1.5-2.0 mm', 301.5),
      right: straight('Shimano XT butted 2.0-1.5-2.0 mm', 301.5),
    },
    rear: {
      spokeCount: 28,
      lacingPattern: '3X',
      left: straight('Shimano XT butted 2.0-1.5-2.0 mm', 301.5),
      right: straight('Shimano XT butted 2.0-1.5-2.0 mm', 298),
    },
    nippleModel: xtNipple,
    nippleLengthMm: null,
    verificationStatus: 'verified',
    sourceUrl: sources.m8100,
  }),
]

// The source HTML maps to the repair-kit fields only. Hub, axle, OEM part,
// complete wheel-kit, and EV document identifiers stay out of the page data.
// Spoke lengths come from the supplied repair-kit master directory and are
// mapped to the corresponding Shimano EV records. Nipple length is null where
// Shimano does not publish a usable length for the replacement nipple.
export const shimanoWheelsetSpokeCatalog: BrandWheelsetSpokeCatalog = {
  brandSlug: 'shimano',
  brandName: 'Shimano',
  publicationStatus: 'published',
  sourceCheckedAt: '2026-10-01',
  wheelsets: shimanoWheelsets,
}
