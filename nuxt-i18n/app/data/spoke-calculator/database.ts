/**
 * Public spoke calculator contracts.
 *
 * CAD dimensions and verified build measurements are proprietary server data.
 * Keep this module free of catalog records so they cannot enter the browser
 * bundle or source maps. The API returns identifier/label projections only.
 */
export interface HubGeometry {
  leftFlange: number | null
  rightFlange: number | null
  leftFlangePcd: number | null
  rightFlangePcd: number | null
  spokeHoleDiameter?: number | null
}

export interface HubModel {
  id: string
  name: string
  front?: HubGeometry
  rear?: HubGeometry
}

export interface RimModel {
  id: string
  name: string
  erd: number | null
  weight?: number
}

export interface Brand<T> {
  id: string
  name: string
  items: T[]
}

export interface IntOption {
  value: number
  label: string
}

export interface StringOption {
  value: string
  label: string
}

export interface SpokeCatalogOptions {
  spokeCounts: IntOption[]
  crossings: IntOption[]
  nippleTypes: StringOption[]
  wheelPositions: StringOption[]
}

export interface SpokeCatalog {
  options: SpokeCatalogOptions
  rims: Brand<RimModel>[]
  hubs: Brand<HubModel>[]
  presets: WheelBuildPreset[]
}

export interface WheelBuildActualLengths {
  frontLeft: number | null
  frontRight: number | null
  rearLeft: number | null
  rearRight: number | null
  notes?: string
}

export interface WheelBuildPreset {
  id: string
  name: string
  keywords: string[]
  description?: string
  rimBrandId: string
  rimModelId: string
  hubBrandId: string
  hubModelId: string
  spokeCount: number
  crossing: number
  nippleType: 'standard' | 'hidden'
  nippleLength: number | null
  wheelPosition?: 'auto' | 'front' | 'rear'
  actualLengths?: WheelBuildActualLengths | null
}

export const SPOKE_CALCULATOR_OPTIONS: SpokeCatalogOptions = {
  spokeCounts: [16, 18, 20, 24, 28, 32, 36].map(value => ({ label: String(value), value })),
  crossings: [
    { label: '0-cross (Radial)', value: 0 },
    { label: '1-cross', value: 1 },
    { label: '2-cross', value: 2 },
    { label: '3-cross', value: 3 },
    { label: '4-cross', value: 4 },
  ],
  nippleTypes: [
    { label: 'Standard external', value: 'standard' },
    { label: 'Hidden / aero', value: 'hidden' },
  ],
  wheelPositions: [
    { label: 'Auto', value: 'auto' },
    { label: 'Front', value: 'front' },
    { label: 'Rear', value: 'rear' },
  ],
}

// Compatibility exports intentionally contain no proprietary records.
export const RIM_DATABASE: Brand<RimModel>[] = []
export const HUB_DATABASE: Brand<HubModel>[] = []
// Search metadata is not sensitive and is kept for instant SSR rendering.
// Dimensions and verified lengths remain server-only.
export const PRESET_BUILDS: WheelBuildPreset[] = [
  {
    id: 'tz_ar45_dt350_fr',
    name: 'AR 45 Disc + DT Swiss 350',
    description: 'Popular all-rounder build.',
    keywords: ['350', 'dt swiss', '45mm', 'road'],
    rimBrandId: 'dt_swiss',
    rimModelId: 'rr411_db',
    hubBrandId: 'dt_swiss',
    hubModelId: '350_road_db_cl',
    spokeCount: 24,
    crossing: 2,
    nippleType: 'standard',
    nippleLength: null,
  },
  {
    id: 'tz_ar50_dt240_fr',
    name: 'AR 50 Disc + DT Swiss 240 EXP',
    description: 'Lightweight racing build.',
    keywords: ['240', 'dt swiss', '50mm', 'racing'],
    rimBrandId: 'dt_swiss',
    rimModelId: 'rr511_db',
    hubBrandId: 'dt_swiss',
    hubModelId: '240_road_db_cl',
    spokeCount: 24,
    crossing: 2,
    nippleType: 'hidden',
    nippleLength: null,
  },
  {
    id: 'mavic_open_dt350',
    name: 'Mavic Open Pro UST + DT Swiss 350',
    description: 'Classic training wheelset.',
    keywords: ['mavic', 'open pro', '350', 'training'],
    rimBrandId: 'mavic',
    rimModelId: 'open_pro_ust_disc',
    hubBrandId: 'dt_swiss',
    hubModelId: '350_road_db_cl',
    spokeCount: 28,
    crossing: 3,
    nippleType: 'standard',
    nippleLength: null,
  },
]

export const DEFAULT_SPOKE_CATALOG: SpokeCatalog = {
  options: SPOKE_CALCULATOR_OPTIONS,
  rims: RIM_DATABASE,
  hubs: HUB_DATABASE,
  presets: PRESET_BUILDS,
}
