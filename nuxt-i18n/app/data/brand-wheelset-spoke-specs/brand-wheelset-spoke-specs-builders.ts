import type {
  BrandWheelPosition,
  BrandWheelsetVerificationStatus,
  BrandWheelsetLifecycleStatus,
  BrandWheelsetSpokeRecord,
  BrandWheelSideSpokeSpec,
  BrandWheelPositionSpokeSpec,
} from './brand-wheelset-spoke-specs-types'

export type SideInput = Pick<BrandWheelSideSpokeSpec, 'spokeModel' | 'headType' | 'lengthMm'>
export type WheelInput = {
  spokeCount: number
  lacingPattern: string
  left: SideInput
  right: SideInput
}
export type WheelsetInput = {
  slug: string
  model: string
  lifecycleStatus?: BrandWheelsetLifecycleStatus
  verificationStatus?: BrandWheelsetVerificationStatus
  rim: BrandWheelsetSpokeRecord['rim']
  front: WheelInput
  rear: WheelInput
  nippleModel: string
  nippleLengthMm: number | null
  sourceUrl: string
}

export const straight = (spokeModel: string, lengthMm: number | null = null): SideInput => ({
  spokeModel,
  headType: 'straight-pull',
  lengthMm,
})

export const jBend = (spokeModel: string, lengthMm: number | null = null): SideInput => ({
  spokeModel,
  headType: 'j-bend',
  lengthMm,
})

export const wheel = (position: BrandWheelPosition, input: WheelInput): BrandWheelPositionSpokeSpec => ({
  position,
  spokeCount: input.spokeCount,
  lacingPattern: input.lacingPattern,
  sides: [
    { side: 'left', ...input.left },
    { side: 'right', ...input.right },
  ],
})

export const wheelset = (input: WheelsetInput): BrandWheelsetSpokeRecord => ({
  slug: input.slug,
  model: input.model,
  lifecycleStatus: input.lifecycleStatus ?? 'current',
  verificationStatus: input.verificationStatus ?? 'verified',
  rim: input.rim,
  nippleModel: input.nippleModel,
  nippleLengthMm: input.nippleLengthMm,
  sourceUrl: input.sourceUrl,
  wheels: [
    wheel('front', input.front),
    wheel('rear', input.rear),
  ],
})

export const pair = (
  spokeModel: string,
  spokeCount = 24,
  lacingPattern = '2X',
  lengthMm: number | null = null,
): WheelInput => ({
  spokeCount,
  lacingPattern,
  left: straight(spokeModel, lengthMm),
  right: straight(spokeModel, lengthMm),
})
