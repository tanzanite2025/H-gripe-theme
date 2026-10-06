// Shared types for the spoke length calculator UI and API.

export interface RimGeometry {
  id: string
  sku?: string
  brand?: string
  model?: string

  // Geometry used for spoke length calculations
  erd: number
  spokeHoles: number
  offsetMm?: number

  // Optional, for display / filtering only
  diameterLabel?: string
  internalWidthMm?: number
  externalWidthMm?: number
  nippleSeatType?: string
  holeType?: 'eyelet' | 'non-eyelet' | string
  material?: 'alloy' | 'carbon' | string
}

export interface HubGeometry {
  id: string
  sku?: string
  brand?: string
  model?: string

  spokeHoles: number
  type?: 'front' | 'rear' | 'front-rear-compatible'
  brakeType?: 'disc' | 'rim' | 'centerlock' | string
  axleWidthMm?: number

  // Geometry used for spoke length calculations
  leftFlangePcdMm: number
  rightFlangePcdMm: number
  spokeHoleDiameterMm?: number | null
  leftFlangeToCenterMm: number
  rightFlangeToCenterMm: number
}

// Request payload coming from the Nuxt page to the spoke-calc API.
export interface SpokeCalcInput {
  rimId: string
  hubId: string
  wheelPosition: 'front' | 'rear'
  topologyId: string
  spokeCount: number
  crossing: number
  g3RimHoleSpacingAToBDegrees?: number
  g3RimHoleSpacingBToADegrees?: number
  g3RimHoleSpacingAToNextGroupADegrees?: number
  erdMm?: number | null
  leftFlangeMm?: number | null
  rightFlangeMm?: number | null
  leftFlangePcdMm?: number | null
  rightFlangePcdMm?: number | null
  rimOffsetMm?: number
  nippleType?: 'standard' | 'hidden' | string
  nippleLengthMm?: number | null
  spokeHeadType?: 'j_bend' | 'straight_pull'
  spokeHoleDiameterMm?: number | null
  straightPullTangentOffsetMm?: number
  spokeProfile?: 'round_2_0' | 'round_1_8' | 'bladed_0_9x2_2'
  targetTensionN?: number
  alternatingDrillingOffsetMm?: number
  interlacing?: boolean
  interlaceCompensationMm?: number | null
  spokeElongationCompensationMm?: number | null
}

export interface SpokeTensionRatio {
  leftToRight: number
  rightToLeft: number
  lowerToHigher: number
  lowerSide: 'left' | 'right' | 'balanced'
  leftBracingAngleDeg: number
  rightBracingAngleDeg: number
}

// Minimal response the UI needs from the API.
export interface SpokeCalcResult {
  leftLengthMm: number
  rightLengthMm: number
  tensionRatio: SpokeTensionRatio | null
  topologyId?: string
  distribution?: 'symmetric_1to1' | 'uniform_2to1' | 'g3_2to1'
  spokeLengths?: SpokeCalcSpokeLengthResult[]
}

export interface SpokeCalcSpokeLengthResult {
  id: number
  side: 'A' | 'B'
  physicalSide: 'left' | 'right'
  type: 'leading' | 'trailing' | 'nondrive'
  hubHoleId: number
  rimHoleId: number
  lengthMm: number
}
