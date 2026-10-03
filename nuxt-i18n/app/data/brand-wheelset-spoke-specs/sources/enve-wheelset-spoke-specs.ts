import type {
  BrandSpokeHeadType,
  BrandWheelsetVerificationStatus,
  BrandWheelsetSpokeCatalog,
  BrandWheelsetSpokeRecord,
} from '../brand-wheelset-spoke-specs-types.js'
import { jBend, straight, wheelset, type SideInput } from '../brand-wheelset-spoke-specs-builders.js'

const sourceUrl = 'https://support.enve.com/hc/en-us/articles/360058866734-Spoke-Chart-and-Tension'
type EnveHeadType = Exclude<BrandSpokeHeadType, 'unknown'>

type EnveWheelInput = {
  spokeModel: string
  headType: EnveHeadType
  lengths: [number | null, number | null]
  spokeCount?: number
  lacingPattern?: string
}

const makeSide = (input: EnveWheelInput, lengthMm: number | null): SideInput => (
  input.headType === 'j-bend'
    ? jBend(input.spokeModel, lengthMm)
    : straight(input.spokeModel, lengthMm)
)

const makeEnveWheelset = (input: {
  slug: string
  model: string
  depthFrontMm: number
  depthRearMm: number
  innerWidthMm: number
  outerWidthMm: number
  front: EnveWheelInput
  rear: EnveWheelInput
  nippleModel: string
  verificationStatus?: BrandWheelsetVerificationStatus
}): BrandWheelsetSpokeRecord => wheelset({
  slug: input.slug,
  model: input.model,
  rim: {
    depthMm: input.depthFrontMm,
    depthFrontMm: input.depthFrontMm,
    depthRearMm: input.depthRearMm,
    innerWidthMm: input.innerWidthMm,
    outerWidthMm: input.outerWidthMm,
  },
  front: {
    spokeCount: input.front.spokeCount ?? 24,
    lacingPattern: input.front.lacingPattern ?? '2X',
    left: makeSide(input.front, input.front.lengths[0]),
    right: makeSide(input.front, input.front.lengths[1]),
  },
  rear: {
    spokeCount: input.rear.spokeCount ?? 24,
    lacingPattern: input.rear.lacingPattern ?? '2X',
    left: makeSide(input.rear, input.rear.lengths[0]),
    right: makeSide(input.rear, input.rear.lengths[1]),
  },
  nippleModel: input.nippleModel,
  nippleLengthMm: null,
  verificationStatus: input.verificationStatus ?? 'verified',
  sourceUrl,
})

const straightCxRay = (lengths: [number | null, number | null]): EnveWheelInput => ({
  spokeModel: 'Sapim CX-Ray TCS OH bladed straight-pull',
  headType: 'straight-pull',
  lengths,
})

const straightCxRayWide = (lengths: [number | null, number | null]): EnveWheelInput => ({
  spokeModel: 'Sapim CX-Ray bladed straight-pull',
  headType: 'straight-pull',
  lengths,
})

const jBendCxSprint = (lengths: [number | null, number | null]): EnveWheelInput => ({
  spokeModel: 'Sapim CX-Sprint bladed J-bend',
  headType: 'j-bend',
  lengths,
})

const primaryWheelsets: BrandWheelsetSpokeRecord[] = [
  makeEnveWheelset({
    slug: 'enve-ses-2-3-gen4',
    model: 'ENVE SES 2.3 (Gen 4)',
    depthFrontMm: 28,
    depthRearMm: 32,
    innerWidthMm: 21,
    outerWidthMm: 25,
    front: straightCxRay([300, 300]),
    rear: straightCxRay([296, 296]),
    nippleModel: 'ENVE 7075-T6 inverted internal alloy nipple',
  }),
  makeEnveWheelset({
    slug: 'enve-ses-3-4-gen4',
    model: 'ENVE SES 3.4 (Gen 4)',
    depthFrontMm: 39,
    depthRearMm: 43,
    innerWidthMm: 25,
    outerWidthMm: 32,
    front: {
      ...straightCxRay([287, 287]),
      spokeModel: 'Sapim CX-Ray TCS OH bladed 2.0-0.9 x 2.2-2.0 mm straight-pull',
    },
    rear: {
      ...straightCxRay([283, 283]),
      spokeModel: 'Sapim CX-Ray TCS OH bladed 2.0-0.9 x 2.2-2.0 mm straight-pull',
    },
    nippleModel: 'ENVE molded inverted internal alloy nipple',
  }),
  makeEnveWheelset({
    slug: 'enve-ses-4-5-gen4',
    model: 'ENVE SES 4.5 (Gen 4)',
    depthFrontMm: 50,
    depthRearMm: 56,
    innerWidthMm: 25,
    outerWidthMm: 32,
    front: straightCxRay([276, 276]),
    rear: straightCxRay([271, 271]),
    nippleModel: 'ENVE inverted internal alloy nipple',
  }),
  makeEnveWheelset({
    slug: 'enve-ses-4-5-pro',
    model: 'ENVE SES 4.5 PRO (Team UAE Emirates Edition)',
    depthFrontMm: 50,
    depthRearMm: 56,
    innerWidthMm: 23.5,
    outerWidthMm: 31,
    front: {
      spokeModel: 'Alpina Ultralite Aero R5 bladed 2.0-1.5-2.0 mm straight-pull',
      headType: 'straight-pull',
      lengths: [276, 278],
    },
    rear: {
      spokeModel: 'Alpina Ultralite Aero R5 bladed 2.0-1.5-2.0 mm straight-pull',
      headType: 'straight-pull',
      lengths: [270, 273],
    },
    nippleModel: 'Alpina 7075-T6 inverted internal alloy nipple',
    verificationStatus: 'pending',
  }),
  makeEnveWheelset({
    slug: 'enve-ses-6-7-gen4',
    model: 'ENVE SES 6.7 (Gen 4)',
    depthFrontMm: 60,
    depthRearMm: 67,
    innerWidthMm: 23,
    outerWidthMm: 30,
    front: straightCxRay([267, 267]),
    rear: straightCxRay([261, 261]),
    nippleModel: 'ENVE inverted internal alloy nipple',
  }),
  makeEnveWheelset({
    slug: 'enve-g23-gravel',
    model: 'ENVE G23 (700c Gravel)',
    depthFrontMm: 25,
    depthRearMm: 25,
    innerWidthMm: 23,
    outerWidthMm: 31.5,
    front: straightCxRayWide([301, 301]),
    rear: straightCxRayWide([301, 301]),
    nippleModel: 'ENVE inverted internal alloy nipple 14G',
  }),
  makeEnveWheelset({
    slug: 'enve-g27-gravel',
    model: 'ENVE G27 (650b Gravel)',
    depthFrontMm: 25,
    depthRearMm: 25,
    innerWidthMm: 27,
    outerWidthMm: 35.5,
    // The supplied chart lists Alloy CL lengths in the four position fields
    // and identifies 282 mm for the primary INNERDRIVE configuration. This
    // record follows the primary configuration used by the current wheelset.
    front: straightCxRayWide([282, 282]),
    rear: straightCxRayWide([282, 282]),
    nippleModel: 'ENVE inverted internal alloy nipple 14G',
    verificationStatus: 'pending',
  }),
  makeEnveWheelset({
    slug: 'enve-foundation-45',
    model: 'ENVE 45 (Foundation Road)',
    depthFrontMm: 45,
    depthRearMm: 45,
    innerWidthMm: 21,
    outerWidthMm: 28,
    front: jBendCxSprint([274, 276]),
    rear: jBendCxSprint([276, 270]),
    nippleModel: 'Sapim Double Square external alloy/brass nipple 14G',
  }),
  makeEnveWheelset({
    slug: 'enve-foundation-65',
    model: 'ENVE 65 (Foundation Road)',
    depthFrontMm: 65,
    depthRearMm: 65,
    innerWidthMm: 21,
    outerWidthMm: 28,
    front: jBendCxSprint([254, 256]),
    rear: jBendCxSprint([256, 250]),
    nippleModel: 'Sapim Double Square external alloy/brass nipple 14G',
  }),
  makeEnveWheelset({
    slug: 'enve-foundation-ag25',
    model: 'ENVE AG25 (Foundation Gravel 700c)',
    depthFrontMm: 21,
    depthRearMm: 21,
    innerWidthMm: 25,
    outerWidthMm: 32,
    front: jBendCxSprint([290, 294]),
    rear: jBendCxSprint([292, 288]),
    nippleModel: 'Sapim Double Square external brass nipple 14G',
  }),
  makeEnveWheelset({
    slug: 'enve-foundation-ag28',
    model: 'ENVE AG28 (Foundation Gravel 650b)',
    depthFrontMm: 21,
    depthRearMm: 21,
    innerWidthMm: 28,
    outerWidthMm: 34,
    front: jBendCxSprint([272, 276]),
    rear: jBendCxSprint([274, 270]),
    nippleModel: 'Sapim Double Square external brass nipple 14G',
  }),
  makeEnveWheelset({
    slug: 'enve-foundation-am30-29',
    model: 'ENVE AM30 (Foundation MTB 29" Boost)',
    depthFrontMm: 20,
    depthRearMm: 20,
    innerWidthMm: 30,
    outerWidthMm: 39,
    front: {
      spokeCount: 28,
      ...jBendCxSprint([288, 290]),
      spokeModel: 'Sapim Race butted 2.0-1.8-2.0 mm J-bend',
    },
    rear: {
      spokeCount: 28,
      ...jBendCxSprint([290, 286]),
      spokeModel: 'Sapim Race butted 2.0-1.8-2.0 mm J-bend',
    },
    nippleModel: 'Alpina Nylock external brass nipple 14G',
  }),
]

// The supplied ENVE chart includes alternate Alloy, DT, Industry Nine, and
// Classified hub configurations with different spoke lengths. The directory
// stores the chart's primary configuration for the verified records; G27 and
// SES 4.5 PRO stay in the backend audit queue until their variant lengths are
// checked separately.
export const enveWheelsetSpokeCatalog: BrandWheelsetSpokeCatalog = {
  brandSlug: 'enve',
  brandName: 'ENVE',
  publicationStatus: 'published',
  sourceCheckedAt: '2026-10-01',
  wheelsets: primaryWheelsets,
}
