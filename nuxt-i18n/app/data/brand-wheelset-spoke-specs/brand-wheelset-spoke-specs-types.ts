export type BrandWheelsetPublicationStatus = 'draft' | 'published'
export type BrandWheelsetVerificationStatus = 'pending' | 'verified'
export type BrandWheelsetLifecycleStatus = 'current' | 'legacy'
export type BrandWheelPosition = 'front' | 'rear'
export type BrandWheelSpokeSide = 'left' | 'right'
export type BrandSpokeHeadType = 'straight-pull' | 'j-bend' | 'unknown'

export interface BrandWheelSideSpokeSpec {
  side: BrandWheelSpokeSide
  lengthMm: number | null
  spokeModel: string
  headType: BrandSpokeHeadType
}

export interface BrandWheelPositionSpokeSpec {
  position: BrandWheelPosition
  spokeCount: number
  lacingPattern: string
  sides: BrandWheelSideSpokeSpec[]
}

export interface BrandWheelsetSpokeRecord {
  slug: string
  model: string
  lifecycleStatus: BrandWheelsetLifecycleStatus
  verificationStatus: BrandWheelsetVerificationStatus
  rim: {
    depthMm: number
    innerWidthMm: number
    outerWidthMm: number
  }
  nippleModel: string
  nippleLengthMm: number | null
  sourceUrl: string
  wheels: BrandWheelPositionSpokeSpec[]
}

export interface BrandWheelsetSpokeCatalog {
  brandSlug: string
  brandName: string
  publicationStatus: BrandWheelsetPublicationStatus
  sourceCheckedAt: string | null
  wheelsets: BrandWheelsetSpokeRecord[]
}
