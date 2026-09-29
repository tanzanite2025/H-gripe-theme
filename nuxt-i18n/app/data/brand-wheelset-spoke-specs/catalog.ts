export type BrandWheelsetPublicationStatus = 'draft' | 'published'
export type BrandWheelsetVerificationStatus = 'pending' | 'verified'
export type BrandWheelPosition = 'front' | 'rear'
export type BrandWheelSpokeSide = 'left' | 'right' | 'drive' | 'nonDrive'

export interface BrandWheelSideSpokeLength {
  side: BrandWheelSpokeSide
  lengthMm: number
}

export interface BrandWheelPositionSpokeSpec {
  position: BrandWheelPosition
  spokeCount: number
  lacingPattern: string
  spokeType: string
  sides: BrandWheelSideSpokeLength[]
}

export interface BrandWheelsetSpokeRecord {
  slug: string
  model: string
  modelYears: string
  verificationStatus: BrandWheelsetVerificationStatus
  wheels: BrandWheelPositionSpokeSpec[]
}

export interface BrandWheelsetSpokeCatalog {
  brandSlug: string
  brandName: string
  publicationStatus: BrandWheelsetPublicationStatus
  sourceCheckedAt: string | null
  wheelsets: BrandWheelsetSpokeRecord[]
}

interface DraftBrandWheelsetInput {
  slug: string
  model: string
  modelYears: string
  spokeCount: number
  frontPattern: string
  frontLeft: number
  frontRight: number
  frontSpokeType: string
  rearPattern: string
  rearDrive: number
  rearNonDrive: number
  rearSpokeType: string
}

const draftBrandWheelset = (input: DraftBrandWheelsetInput): BrandWheelsetSpokeRecord => ({
  slug: input.slug,
  model: input.model,
  modelYears: input.modelYears,
  verificationStatus: 'pending',
  wheels: [
    {
      position: 'front',
      spokeCount: input.spokeCount,
      lacingPattern: input.frontPattern,
      spokeType: input.frontSpokeType,
      sides: [
        { side: 'left', lengthMm: input.frontLeft },
        { side: 'right', lengthMm: input.frontRight },
      ],
    },
    {
      position: 'rear',
      spokeCount: input.spokeCount,
      lacingPattern: input.rearPattern,
      spokeType: input.rearSpokeType,
      sides: [
        { side: 'drive', lengthMm: input.rearDrive },
        { side: 'nonDrive', lengthMm: input.rearNonDrive },
      ],
    },
  ],
})

// This catalog covers market-available branded complete wheelsets. It stays
// separate from custom-build calculator inputs and user-entered records.
// Add one catalog per brand to the shared reference directory.
export const brandWheelsetSpokeCatalogs: BrandWheelsetSpokeCatalog[] = [
  {
    brandSlug: 'dt-swiss',
    brandName: 'DT Swiss',
    publicationStatus: 'draft',
    sourceCheckedAt: null,
    wheelsets: [
    draftBrandWheelset({
      slug: 'arc-1100-dicut-db-50',
      model: 'ARC 1100 DICUT DB 50',
      modelYears: '2023–2025',
      spokeCount: 24,
      frontPattern: '2X / 2X',
      frontLeft: 268,
      frontRight: 270,
      frontSpokeType: 'DT Aerolite II / Aero Comp II T-head',
      rearPattern: '2X / 2X',
      rearDrive: 272,
      rearNonDrive: 266,
      rearSpokeType: 'DT Aero Comp II / Aerolite II T-head',
    }),
    draftBrandWheelset({
      slug: 'arc-1100-dicut-db-62',
      model: 'ARC 1100 DICUT DB 62',
      modelYears: '2023–2025',
      spokeCount: 24,
      frontPattern: '2X / 2X',
      frontLeft: 256,
      frontRight: 258,
      frontSpokeType: 'DT Aerolite II / Aero Comp II T-head',
      rearPattern: '2X / 2X',
      rearDrive: 260,
      rearNonDrive: 254,
      rearSpokeType: 'DT Aero Comp II / Aerolite II T-head',
    }),
    draftBrandWheelset({
      slug: 'arc-1100-dicut-db-80',
      model: 'ARC 1100 DICUT DB 80',
      modelYears: '2023–2025',
      spokeCount: 24,
      frontPattern: '2X / 2X',
      frontLeft: 238,
      frontRight: 240,
      frontSpokeType: 'DT Aerolite II T-head',
      rearPattern: '2X / 2X',
      rearDrive: 242,
      rearNonDrive: 236,
      rearSpokeType: 'DT Aero Comp II T-head',
    }),
    draftBrandWheelset({
      slug: 'arc-1400-dicut-db-50',
      model: 'ARC 1400 DICUT DB 50',
      modelYears: '2023–2025',
      spokeCount: 24,
      frontPattern: '2X / 2X',
      frontLeft: 268,
      frontRight: 270,
      frontSpokeType: 'DT Aero Comp T-head',
      rearPattern: '2X / 2X',
      rearDrive: 272,
      rearNonDrive: 266,
      rearSpokeType: 'DT Aero Comp T-head',
    }),
    draftBrandWheelset({
      slug: 'arc-1400-dicut-db-62',
      model: 'ARC 1400 DICUT DB 62',
      modelYears: '2023–2025',
      spokeCount: 24,
      frontPattern: '2X / 2X',
      frontLeft: 256,
      frontRight: 258,
      frontSpokeType: 'DT Aero Comp T-head',
      rearPattern: '2X / 2X',
      rearDrive: 260,
      rearNonDrive: 254,
      rearSpokeType: 'DT Aero Comp T-head',
    }),
    draftBrandWheelset({
      slug: 'prc-1100-mon-chasseral-db-24',
      model: 'PRC 1100 Mon Chasseral DB 24',
      modelYears: '2022–2024',
      spokeCount: 24,
      frontPattern: '2X / 2X',
      frontLeft: 292,
      frontRight: 294,
      frontSpokeType: 'DT Aerolite II T-head',
      rearPattern: '2X / 2X',
      rearDrive: 296,
      rearNonDrive: 290,
      rearSpokeType: 'DT Aero Comp II T-head',
    }),
    draftBrandWheelset({
      slug: 'erc-1100-dicut-db-35',
      model: 'ERC 1100 DICUT DB 35',
      modelYears: '2023–2025',
      spokeCount: 24,
      frontPattern: '2X / 2X',
      frontLeft: 282,
      frontRight: 284,
      frontSpokeType: 'DT Aerolite II T-head',
      rearPattern: '2X / 2X',
      rearDrive: 286,
      rearNonDrive: 280,
      rearSpokeType: 'DT Aero Comp II T-head',
    }),
    draftBrandWheelset({
      slug: 'erc-1100-dicut-db-45',
      model: 'ERC 1100 DICUT DB 45',
      modelYears: '2023–2025',
      spokeCount: 24,
      frontPattern: '2X / 2X',
      frontLeft: 272,
      frontRight: 274,
      frontSpokeType: 'DT Aerolite II T-head',
      rearPattern: '2X / 2X',
      rearDrive: 276,
      rearNonDrive: 270,
      rearSpokeType: 'DT Aero Comp II T-head',
    }),
    draftBrandWheelset({
      slug: 'grc-1400-dicut-db-42',
      model: 'GRC 1400 DICUT DB 42',
      modelYears: '2022–2025',
      spokeCount: 24,
      frontPattern: '2X / 2X',
      frontLeft: 276,
      frontRight: 278,
      frontSpokeType: 'DT Aero Comp T-head',
      rearPattern: '2X / 2X',
      rearDrive: 280,
      rearNonDrive: 274,
      rearSpokeType: 'DT Aero Comp T-head',
    }),
    draftBrandWheelset({
      slug: 'xrc-1200-spline-30-29',
      model: 'XRC 1200 SPLINE 30 29″',
      modelYears: '2023–2025',
      spokeCount: 28,
      frontPattern: '3X / 3X',
      frontLeft: 302,
      frontRight: 304,
      frontSpokeType: 'DT Revolite Straightpull',
      rearPattern: '3X / 3X',
      rearDrive: 304,
      rearNonDrive: 300,
      rearSpokeType: 'DT Revolite Straightpull',
    }),
    draftBrandWheelset({
      slug: 'exc-1200-spline-30-29',
      model: 'EXC 1200 SPLINE 30 29″',
      modelYears: '2023–2025',
      spokeCount: 28,
      frontPattern: '3X / 3X',
      frontLeft: 300,
      frontRight: 302,
      frontSpokeType: 'DT Competition Race Straightpull',
      rearPattern: '3X / 3X',
      rearDrive: 302,
      rearNonDrive: 298,
      rearSpokeType: 'DT Competition Race Straightpull',
    }),
    ],
  },
]
